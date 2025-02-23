package writing

import (
	"context"
	"errors"
	writingHelper "fluencybe/internal/app/helper/writing"
	"fluencybe/internal/app/model/writing"
	writingRepository "fluencybe/internal/app/repository/writing"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
)

type WritingEssayService struct {
	repo                *writingRepository.WritingEssayRepository
	writingQuestionRepo *writingRepository.WritingQuestionRepository
	logger              *logger.PrettyLogger
	cache               cache.Cache
	questionUpdator     *writingHelper.WritingQuestionUpdator
}

func NewWritingEssayService(
	repo *writingRepository.WritingEssayRepository,
	writingQuestionRepo *writingRepository.WritingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *writingHelper.WritingQuestionUpdator,
) *WritingEssayService {
	return &WritingEssayService{
		repo:                repo,
		writingQuestionRepo: writingQuestionRepo,
		logger:              logger,
		cache:               cache,
		questionUpdator:     questionUpdator,
	}
}

func (s *WritingEssayService) SetQuestionUpdator(updator *writingHelper.WritingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *WritingEssayService) validateEssay(essay *writing.WritingEssay) error {
	if essay == nil {
		return errors.New("invalid input")
	}
	if essay.WritingQuestionID == uuid.Nil {
		return errors.New("writing question ID is required")
	}
	if essay.EssayType == "" {
		return errors.New("essay type is required")
	}
	if len(essay.RequiredPoints) == 0 {
		return errors.New("at least one required point is needed")
	}
	if essay.MinWords < 1 {
		return errors.New("minimum words must be at least 1")
	}
	if essay.MaxWords < essay.MinWords {
		return errors.New("maximum words must be greater than minimum words")
	}
	if essay.SampleEssay == "" {
		return errors.New("sample essay is required")
	}
	if essay.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *WritingEssayService) Create(ctx context.Context, essay *writing.WritingEssay) error {
	if err := s.validateEssay(essay); err != nil {
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

	if err := s.repo.Create(ctx, essay); err != nil {
		tx.Rollback()
		s.logger.Error("writing_essay_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create essay")
		return err
	}

	// Get parent writing question for cache/search update
	parentQuestion, err := s.writingQuestionRepo.GetWritingQuestionByID(ctx, essay.WritingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get writing question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("writing_essay_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *WritingEssayService) GetByID(ctx context.Context, id uuid.UUID) (*writing.WritingEssay, error) {
	essay, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("writing_essay_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get essay by ID")
		return nil, err
	}
	return essay, nil
}

func (s *WritingEssayService) GetByWritingQuestionID(ctx context.Context, writingQuestionID uuid.UUID) ([]*writing.WritingEssay, error) {
	essays, err := s.repo.GetByWritingQuestionID(ctx, writingQuestionID)
	if err != nil {
		s.logger.Error("writing_essay_service.get_by_writing_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    writingQuestionID,
		}, "Failed to get essays")
		return nil, err
	}
	return essays, nil
}

func (s *WritingEssayService) Update(ctx context.Context, essay *writing.WritingEssay) error {
	if err := s.validateEssay(essay); err != nil {
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

	if err := s.repo.Update(ctx, essay); err != nil {
		tx.Rollback()
		s.logger.Error("writing_essay_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    essay.ID,
		}, "Failed to update essay")
		return err
	}

	// Get parent writing question for cache/search update
	parentQuestion, err := s.writingQuestionRepo.GetWritingQuestionByID(ctx, essay.WritingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get writing question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("writing_essay_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *WritingEssayService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get essay first to get parent ID
	essay, err := s.GetByID(ctx, id)
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
		s.logger.Error("writing_essay_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete essay")
		return err
	}

	// Get parent writing question for cache/search update
	parentQuestion, err := s.writingQuestionRepo.GetWritingQuestionByID(ctx, essay.WritingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get writing question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("writing_essay_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
