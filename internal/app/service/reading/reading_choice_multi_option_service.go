package reading

import (
	"context"
	"database/sql"
	"errors"
	readingHelper "fluencybe/internal/app/helper/reading"
	"fluencybe/internal/app/model/reading"
	ReadingRepository "fluencybe/internal/app/repository/reading"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReadingChoiceMultiOptionService struct {
	repo                *ReadingRepository.ReadingChoiceMultiOptionRepository
	questionRepo        *ReadingRepository.ReadingChoiceMultiQuestionRepository
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository
	logger              *logger.PrettyLogger
	cache               cache.Cache
	questionUpdator     *readingHelper.ReadingQuestionUpdator
}

func NewReadingChoiceMultiOptionService(
	repo *ReadingRepository.ReadingChoiceMultiOptionRepository,
	questionRepo *ReadingRepository.ReadingChoiceMultiQuestionRepository,
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *readingHelper.ReadingQuestionUpdator,
) *ReadingChoiceMultiOptionService {
	return &ReadingChoiceMultiOptionService{
		repo:                repo,
		questionRepo:        questionRepo,
		readingQuestionRepo: readingQuestionRepo,
		logger:              logger,
		cache:               cache,
		questionUpdator:     questionUpdator,
	}
}

func (s *ReadingChoiceMultiOptionService) SetQuestionUpdator(updator *readingHelper.ReadingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ReadingChoiceMultiOptionService) CreateOption(ctx context.Context, option *reading.ReadingChoiceMultiOption) error {
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
		s.logger.Error("reading_choice_multi_option_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.ReadingChoiceMultiQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Get reading question for cache/search update
	readingQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, parentQuestion.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get reading question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, readingQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_choice_multi_option_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    readingQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ReadingChoiceMultiOptionService) GetOption(ctx context.Context, id uuid.UUID) (*reading.ReadingChoiceMultiOption, error) {
	option, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		s.logger.Error("reading_choice_multi_option_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get option")
		return nil, err
	}

	return option, nil
}

func (s *ReadingChoiceMultiOptionService) UpdateOption(ctx context.Context, option *reading.ReadingChoiceMultiOption) error {
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
		s.logger.Error("reading_choice_multi_option_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    option.ID,
		}, "Failed to update option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.ReadingChoiceMultiQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Get reading question for cache/search update
	readingQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, parentQuestion.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get reading question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, readingQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_choice_multi_option_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    readingQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ReadingChoiceMultiOptionService) DeleteOption(ctx context.Context, id uuid.UUID) error {
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
		s.logger.Error("reading_choice_multi_option_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.ReadingChoiceMultiQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Get reading question for cache/search update
	readingQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, parentQuestion.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get reading question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, readingQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_choice_multi_option_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    readingQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ReadingChoiceMultiOptionService) GetOptionsByQuestionID(ctx context.Context, questionID uuid.UUID) ([]*reading.ReadingChoiceMultiOption, error) {
	options, err := s.repo.GetByReadingChoiceMultiQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("reading_choice_multi_option_service.get_by_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": questionID,
		}, "Failed to get options by question ID")
		return nil, err
	}
	return options, nil
}
