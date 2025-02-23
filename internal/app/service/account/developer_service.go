package account

import (
	"context"
	"errors"
	accountModel "fluencybe/internal/app/model/account"
	accountRepository "fluencybe/internal/app/repository/account"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/utils"
	"fmt"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrNotFound           = errors.New("developer not found")
)

type DeveloperService struct {
	repo   *accountRepository.DeveloperRepository
	logger *logger.PrettyLogger
}

func NewDeveloperService(repo *accountRepository.DeveloperRepository) *DeveloperService {
	return &DeveloperService{
		repo:   repo,
		logger: logger.GetGlobalLogger(),
	}
}

func (s *DeveloperService) withRetry(operation func() error) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 10 * time.Second

	return backoff.Retry(operation, b)
}

func (s *DeveloperService) Register(ctx context.Context, dev *accountModel.Developer) error {
	if dev == nil {
		return ErrInvalidInput
	}

	if err := dev.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dev.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("developer_service.register.hash_password", map[string]interface{}{"error": err.Error()}, "Failed to hash password")
		return fmt.Errorf("failed to hash password: %w", err)
	}
	dev.Password = string(hashedPassword)

	err = s.withRetry(func() error {
		err := s.repo.Create(ctx, dev)
		if err != nil {
			if errors.Is(err, accountRepository.ErrDeveloperDuplicateEmail) {
				return backoff.Permanent(ErrEmailExists)
			}
			return err
		}
		return nil
	})

	if err != nil {
		s.logger.Error("developer_service.register", map[string]interface{}{"error": err.Error()}, "Failed to register developer")
		return err
	}

	return nil
}

func (s *DeveloperService) Login(ctx context.Context, email, password string) (*accountModel.Developer, string, error) {
	if email == "" || password == "" {
		return nil, "", ErrInvalidInput
	}

	var dev *accountModel.Developer
	var err error

	err = s.withRetry(func() error {
		dev, err = s.repo.GetByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, accountRepository.ErrDeveloperNotFound) {
				return backoff.Permanent(ErrNotFound)
			}
			return err
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		s.logger.Error("developer_service.login.get_developer", map[string]interface{}{"error": err.Error()}, "Failed to get developer")
		return nil, "", fmt.Errorf("failed to retrieve developer: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dev.Password), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Generate JWT token with expiration
	token, err := utils.GenerateJWT(dev.ID.String(), "developer")
	if err != nil {
		s.logger.Error("developer_service.login.generate_token", map[string]interface{}{"error": err.Error()}, "Failed to generate token")
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return dev, token, nil
}

func (s *DeveloperService) GetDeveloper(ctx context.Context, id string) (*accountModel.Developer, error) {
	devID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidInput
	}

	var dev *accountModel.Developer

	err = s.withRetry(func() error {
		var err error
		dev, err = s.repo.GetByID(ctx, devID)
		if err != nil {
			if errors.Is(err, accountRepository.ErrDeveloperNotFound) {
				return backoff.Permanent(ErrNotFound)
			}
			return err
		}
		return nil
	})

	if err != nil {
		s.logger.Error("developer_service.get_developer", map[string]interface{}{"error": err.Error()}, "Failed to get developer")
		return nil, err
	}

	return dev, nil
}

func (s *DeveloperService) UpdateDeveloper(ctx context.Context, id string, updates map[string]interface{}) error {
	devID, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidInput
	}

	if len(updates) == 0 {
		return ErrInvalidInput
	}

	if password, ok := updates["password"].(string); ok {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("developer_service.update_developer.hash_password", map[string]interface{}{"error": err.Error()}, "Failed to hash password")
			return fmt.Errorf("failed to hash password: %w", err)
		}
		updates["password"] = string(hashedPassword)
	}

	err = s.withRetry(func() error {
		err := s.repo.Update(ctx, devID, updates)
		if err != nil {
			if errors.Is(err, accountRepository.ErrDeveloperNotFound) {
				return backoff.Permanent(ErrNotFound)
			} else if errors.Is(err, accountRepository.ErrDeveloperDuplicateEmail) {
				return backoff.Permanent(ErrEmailExists)
			}
			return err
		}
		return nil
	})

	if err != nil {
		s.logger.Error("developer_service.update_developer", map[string]interface{}{"error": err.Error()}, "Failed to update developer")
		return err
	}

	return nil
}

func (s *DeveloperService) DeleteDeveloper(ctx context.Context, id string) error {
	devID, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidInput
	}

	err = s.withRetry(func() error {
		err := s.repo.Delete(ctx, devID)
		if err != nil {
			if errors.Is(err, accountRepository.ErrDeveloperNotFound) {
				return backoff.Permanent(ErrNotFound)
			}
			return err
		}
		return nil
	})

	if err != nil {
		s.logger.Error("developer_service.delete_developer", map[string]interface{}{"error": err.Error()}, "Failed to delete developer")
		return err
	}

	return nil
}

func (s *DeveloperService) GetDeveloperList(ctx context.Context, page string, limit string) ([]*accountModel.Developer, int64, error) {
	// Convert page and limit to integers
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}
	
	limitNum, err := strconv.Atoi(limit)
	if err != nil || limitNum < 1 {
		limitNum = 10
	}

	// Calculate offset
	offset := (pageNum - 1) * limitNum

	var developers []*accountModel.Developer
	var total int64

	// Get developers with pagination
	err = s.withRetry(func() error {
		var err error
		developers, err = s.repo.GetList(ctx, limitNum, offset)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		s.logger.Error("developer_service.get_developer_list", map[string]interface{}{"error": err.Error()}, "Failed to get developer list")
		return nil, 0, fmt.Errorf("failed to get developer list: %w", err)
	}

	// Get total count
	err = s.withRetry(func() error {
		var err error
		total, err = s.repo.GetTotalCount(ctx)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		s.logger.Error("developer_service.get_total_count", map[string]interface{}{"error": err.Error()}, "Failed to get total count")
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return developers, total, nil
}
