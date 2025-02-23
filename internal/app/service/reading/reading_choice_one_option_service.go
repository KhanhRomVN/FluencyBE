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

type ReadingChoiceOneOptionService struct {
	repo                *ReadingRepository.ReadingChoiceOneOptionRepository
	questionRepo        *ReadingRepository.ReadingChoiceOneQuestionRepository
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository
	logger              *logger.PrettyLogger
	cache               cache.Cache
	questionUpdator     *readingHelper.ReadingQuestionUpdator
}

func NewReadingChoiceOneOptionService(
	repo *ReadingRepository.ReadingChoiceOneOptionRepository,
	questionRepo *ReadingRepository.ReadingChoiceOneQuestionRepository,
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *readingHelper.ReadingQuestionUpdator,
) *ReadingChoiceOneOptionService {
	return &ReadingChoiceOneOptionService{
		repo:                repo,
		questionRepo:        questionRepo,
		readingQuestionRepo: readingQuestionRepo,
		logger:              logger,
		cache:               cache,
		questionUpdator:     questionUpdator,
	}
}

func (s *ReadingChoiceOneOptionService) SetQuestionUpdator(updator *readingHelper.ReadingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ReadingChoiceOneOptionService) CreateOption(ctx context.Context, option *reading.ReadingChoiceOneOption) error {
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

	if option.IsCorrect {
		// If this option is marked as correct, ensure no other option for this question is marked as correct
		existingCorrectOption, err := s.repo.GetCorrectOption(ctx, option.ReadingChoiceOneQuestionID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			s.logger.Error("reading_choice_one_option_service.create", map[string]interface{}{
				"error": err.Error(),
			}, "Failed to check existing correct option")
			return err
		}

		if existingCorrectOption != nil {
			// Found an existing correct option, update it to be incorrect
			existingCorrectOption.IsCorrect = false
			if err := s.repo.Update(ctx, existingCorrectOption); err != nil {
				tx.Rollback()
				s.logger.Error("reading_choice_one_option_service.create", map[string]interface{}{
					"error": err.Error(),
				}, "Failed to update existing correct option")
				return err
			}
		}
	}

	if err := s.repo.Create(ctx, option); err != nil {
		tx.Rollback()
		s.logger.Error("reading_choice_one_option_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.ReadingChoiceOneQuestionID)
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
		s.logger.Error("reading_choice_one_option_service.update_cache", map[string]interface{}{
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

func (s *ReadingChoiceOneOptionService) GetOption(ctx context.Context, id uuid.UUID) (*reading.ReadingChoiceOneOption, error) {
	option, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		s.logger.Error("reading_choice_one_option_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get option")
		return nil, err
	}

	return option, nil
}

func (s *ReadingChoiceOneOptionService) UpdateOption(ctx context.Context, option *reading.ReadingChoiceOneOption) error {
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

	if option.IsCorrect {
		// If this option is being marked as correct, ensure no other option is marked as correct
		existingCorrectOption, err := s.repo.GetCorrectOption(ctx, option.ReadingChoiceOneQuestionID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			s.logger.Error("reading_choice_one_option_service.update", map[string]interface{}{
				"error": err.Error(),
			}, "Failed to check existing correct option")
			return err
		}

		if existingCorrectOption != nil && existingCorrectOption.ID != option.ID {
			// Found an existing correct option, update it to be incorrect
			existingCorrectOption.IsCorrect = false
			if err := s.repo.Update(ctx, existingCorrectOption); err != nil {
				tx.Rollback()
				s.logger.Error("reading_choice_one_option_service.update", map[string]interface{}{
					"error": err.Error(),
				}, "Failed to update existing correct option")
				return err
			}
		}
	}

	if err := s.repo.Update(ctx, option); err != nil {
		tx.Rollback()
		s.logger.Error("reading_choice_one_option_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    option.ID,
		}, "Failed to update option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.ReadingChoiceOneQuestionID)
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
		s.logger.Error("reading_choice_one_option_service.update_cache", map[string]interface{}{
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

func (s *ReadingChoiceOneOptionService) DeleteOption(ctx context.Context, id uuid.UUID) error {
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
		s.logger.Error("reading_choice_one_option_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.ReadingChoiceOneQuestionID)
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
		s.logger.Error("reading_choice_one_option_service.update_cache", map[string]interface{}{
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

func (s *ReadingChoiceOneOptionService) GetOptionsByQuestionID(ctx context.Context, questionID uuid.UUID) ([]*reading.ReadingChoiceOneOption, error) {
	options, err := s.repo.GetByReadingChoiceOneQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("reading_choice_one_option_service.get_by_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": questionID,
		}, "Failed to get options by question ID")
		return nil, err
	}
	return options, nil
}
