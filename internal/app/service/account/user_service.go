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

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo   *accountRepository.UserRepository
	logger *logger.PrettyLogger
}

func NewUserService(repo *accountRepository.UserRepository) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger.GetGlobalLogger(),
	}
}

func (s *UserService) hashPassword(password string) (string, error) {
	s.logger.Debug("hash_password_start", nil, "Starting password hashing")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("hash_password_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to hash password")
		return "", errors.New("failed to hash password")
	}
	s.logger.Debug("hash_password_success", nil, "Password hashed successfully")
	return string(hashedPassword), nil
}

func (s *UserService) verifyPassword(hashedPassword, password string) error {
	s.logger.Debug("verify_password_start", nil, "Starting password verification")
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		s.logger.Error("verify_password_failed", nil, "Invalid credentials")
		return errors.New("invalid credentials")
	}
	s.logger.Debug("verify_password_success", nil, "Password verified successfully")
	return nil
}

func (s *UserService) Register(ctx context.Context, user *accountModel.User) error {
	s.logger.Info("register_start", map[string]interface{}{
		"email": user.Email,
	}, "Starting user registration")

	if user.Email == "" || user.Password == "" {
		s.logger.Error("register_validation_failed", nil, "Email and password are required")
		return errors.New("email and password are required")
	}

	hashedPassword, err := s.hashPassword(user.Password)
	if err != nil {
		s.logger.Error("register_hash_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to hash password during registration")
		return err
	}

	user.Password = hashedPassword
	if err := s.repo.Create(ctx, user); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			s.logger.Error("register_duplicate", map[string]interface{}{
				"email": user.Email,
			}, "Email already exists")
			return errors.New("email already exists")
		}
		s.logger.Error("register_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create user in database")
		return err
	}

	s.logger.Info("register_success", map[string]interface{}{
		"user_id": user.ID.String(),
		"email":   user.Email,
	}, "User registered successfully")
	return nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*accountModel.User, string, error) {
	s.logger.Info("login_start", map[string]interface{}{
		"email": email,
	}, "Starting user login")

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("login_user_not_found", map[string]interface{}{
			"email": email,
			"error": err.Error(),
		}, "User not found during login")
		return nil, "", errors.New("invalid credentials")
	}

	if err := s.verifyPassword(user.Password, password); err != nil {
		s.logger.Error("login_invalid_password", map[string]interface{}{
			"email": email,
		}, "Invalid password during login")
		return nil, "", err
	}

	token, err := utils.GenerateJWT(user.ID.String(), "user")
	if err != nil {
		s.logger.Error("login_token_generation_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to generate JWT token")
		return nil, "", fmt.Errorf("failed to generate token: %v", err)
	}

	s.logger.Info("login_success", map[string]interface{}{
		"user_id": user.ID.String(),
		"email":   user.Email,
	}, "User logged in successfully")
	return user, token, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*accountModel.User, error) {
	s.logger.Info("get_user_start", map[string]interface{}{
		"user_id": id,
	}, "Starting to get user")

	userID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("get_user_invalid_id", map[string]interface{}{
			"user_id": id,
			"error":   err.Error(),
		}, "Invalid user ID format")
		return nil, errors.New("invalid user ID")
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("get_user_failed", map[string]interface{}{
			"user_id": id,
			"error":   err.Error(),
		}, "Failed to get user")
		return nil, err
	}

	s.logger.Info("get_user_success", map[string]interface{}{
		"user_id": id,
	}, "User retrieved successfully")
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, updates map[string]interface{}) error {
	s.logger.Info("update_user_start", map[string]interface{}{
		"user_id": id,
		"updates": updates,
	}, "Starting to update user")

	userID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("update_user_invalid_id", map[string]interface{}{
			"user_id": id,
			"error":   err.Error(),
		}, "Invalid user ID format")
		return errors.New("invalid user ID")
	}

	// If password is being updated, hash it
	if password, ok := updates["password"].(string); ok {
		hashedPassword, err := s.hashPassword(password)
		if err != nil {
			s.logger.Error("update_user_password_hash_failed", map[string]interface{}{
				"error": err.Error(),
			}, "Failed to hash new password")
			return err
		}
		updates["password"] = hashedPassword
	}

	if err := s.repo.Update(ctx, userID, updates); err != nil {
		s.logger.Error("update_user_failed", map[string]interface{}{
			"user_id": id,
			"error":   err.Error(),
		}, "Failed to update user")
		return err
	}

	s.logger.Info("update_user_success", map[string]interface{}{
		"user_id": id,
	}, "User updated successfully")
	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	s.logger.Info("delete_user_start", map[string]interface{}{
		"user_id": id,
	}, "Starting to delete user")

	userID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("delete_user_invalid_id", map[string]interface{}{
			"user_id": id,
			"error":   err.Error(),
		}, "Invalid user ID format")
		return errors.New("invalid user ID")
	}

	if err := s.repo.Delete(ctx, userID); err != nil {
		s.logger.Error("delete_user_failed", map[string]interface{}{
			"user_id": id,
			"error":   err.Error(),
		}, "Failed to delete user")
		return err
	}

	s.logger.Info("delete_user_success", map[string]interface{}{
		"user_id": id,
	}, "User deleted successfully")
	return nil
}

func (s *UserService) GetUserList(ctx context.Context, page string, limit string) ([]*accountModel.User, int64, error) {
	s.logger.Info("get_user_list_start", map[string]interface{}{
		"page":  page,
		"limit": limit,
	}, "Starting to get list of users")

	pageNum, err := strconv.Atoi(page)
	if err != nil {
		s.logger.Error("get_user_list_invalid_page", map[string]interface{}{
			"page":  page,
			"error": err.Error(),
		}, "Invalid page number")
		return nil, 0, errors.New("invalid page number")
	}

	limitNum, err := strconv.Atoi(limit)
	if err != nil {
		s.logger.Error("get_user_list_invalid_limit", map[string]interface{}{
			"limit": limit,
			"error": err.Error(),
		}, "Invalid limit number")
		return nil, 0, errors.New("invalid limit number")
	}

	offset := (pageNum - 1) * limitNum
	users, err := s.repo.GetList(ctx, limitNum, offset)
	if err != nil {
		s.logger.Error("get_user_list_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get list of users")
		return nil, 0, err
	}

	total, err := s.repo.GetTotalCount(ctx)
	if err != nil {
		s.logger.Error("get_user_list_count_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get total count")
		return nil, 0, err
	}

	s.logger.Info("get_user_list_success", map[string]interface{}{
		"total": total,
		"count": len(users),
	}, "User list retrieved successfully")
	return users, total, nil
}
