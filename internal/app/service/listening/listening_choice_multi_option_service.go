package listening

import (
	"context"
	"database/sql"
	"errors"
	listeningHelper "fluencybe/internal/app/helper/listening"
	"fluencybe/internal/app/model/listening"
	ListeningRepository "fluencybe/internal/app/repository/listening"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListeningChoiceMultiOptionService struct {
	repo                  *ListeningRepository.ListeningChoiceMultiOptionRepository
	questionRepo          *ListeningRepository.ListeningChoiceMultiQuestionRepository
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository
	logger                *logger.PrettyLogger
	cache                 cache.Cache
	questionUpdator       *listeningHelper.ListeningQuestionUpdator
}

func NewListeningChoiceMultiOptionService(
	repo *ListeningRepository.ListeningChoiceMultiOptionRepository,
	questionRepo *ListeningRepository.ListeningChoiceMultiQuestionRepository,
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *listeningHelper.ListeningQuestionUpdator,
) *ListeningChoiceMultiOptionService {
	return &ListeningChoiceMultiOptionService{
		repo:                  repo,
		questionRepo:          questionRepo,
		listeningQuestionRepo: listeningQuestionRepo,
		logger:                logger,
		cache:                 cache,
		questionUpdator:       questionUpdator,
	}
}

func (s *ListeningChoiceMultiOptionService) SetQuestionUpdator(updator *listeningHelper.ListeningQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ListeningChoiceMultiOptionService) validateOption(option *listening.ListeningChoiceMultiOption) error {
	if option == nil {
		return errors.New("invalid input")
	}
	if option.ListeningChoiceMultiQuestionID == uuid.Nil {
		return errors.New("question ID is required")
	}
	if option.Options == "" {
		return errors.New("option text is required")
	}
	return nil
}

func (s *ListeningChoiceMultiOptionService) CreateOption(ctx context.Context, option *listening.ListeningChoiceMultiOption) error {
	if err := s.validateOption(option); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Start transaction
	tx := s.repo.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.repo.Create(ctx, option); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_option_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.ListeningChoiceMultiQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Get listening question for cache/search update
	listeningQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, parentQuestion.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get listening question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, listeningQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_option_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    listeningQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ListeningChoiceMultiOptionService) GetOption(ctx context.Context, id uuid.UUID) (*listening.ListeningChoiceMultiOption, error) {
	option, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		s.logger.Error("listening_choice_multi_option_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get option")
		return nil, err
	}

	return option, nil
}

func (s *ListeningChoiceMultiOptionService) UpdateOption(ctx context.Context, option *listening.ListeningChoiceMultiOption) error {
	if err := s.validateOption(option); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Start transaction
	tx := s.repo.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.repo.Update(ctx, option); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_option_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    option.ID,
		}, "Failed to update option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.ListeningChoiceMultiQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Get listening question for cache/search update
	listeningQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, parentQuestion.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get listening question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, listeningQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_option_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    listeningQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ListeningChoiceMultiOptionService) DeleteOption(ctx context.Context, id uuid.UUID) error {
	// Get the option first to get the question ID for cache invalidation
	option, err := s.GetOption(ctx, id)
	if err != nil {
		return err
	}

	// Start transaction
	tx := s.repo.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.repo.Delete(ctx, id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sql.ErrNoRows
		}
		s.logger.Error("listening_choice_multi_option_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.ListeningChoiceMultiQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Get listening question for cache/search update
	listeningQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, parentQuestion.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get listening question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, listeningQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_option_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    listeningQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ListeningChoiceMultiOptionService) GetOptionsByQuestionID(ctx context.Context, questionID uuid.UUID) ([]*listening.ListeningChoiceMultiOption, error) {
	options, err := s.repo.GetByListeningChoiceMultiQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("listening_choice_multi_option_service.get_by_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": questionID,
		}, "Failed to get options by question ID")
		return nil, err
	}
	return options, nil
}
