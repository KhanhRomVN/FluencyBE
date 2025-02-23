package reading

import (
	"context"
	"errors"
	readingHelper "fluencybe/internal/app/helper/reading"
	"fluencybe/internal/app/model/reading"
	ReadingRepository "fluencybe/internal/app/repository/reading"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
)

type ReadingMatchingService struct {
	repo                *ReadingRepository.ReadingMatchingRepository
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository
	logger              *logger.PrettyLogger
	cache               cache.Cache
	questionUpdator     *readingHelper.ReadingQuestionUpdator
}

func NewReadingMatchingService(
	repo *ReadingRepository.ReadingMatchingRepository,
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *readingHelper.ReadingQuestionUpdator,
) *ReadingMatchingService {
	return &ReadingMatchingService{
		repo:                repo,
		readingQuestionRepo: readingQuestionRepo,
		logger:              logger,
		cache:               cache,
		questionUpdator:     questionUpdator,
	}
}

func (s *ReadingMatchingService) SetQuestionUpdator(updator *readingHelper.ReadingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ReadingMatchingService) validateMatching(matching *reading.ReadingMatching) error {
	if matching == nil {
		return errors.New("invalid input")
	}
	if matching.ReadingQuestionID == uuid.Nil {
		return errors.New("reading question ID is required")
	}
	if matching.Question == "" {
		return errors.New("question text is required")
	}
	if matching.Answer == "" {
		return errors.New("answer text is required")
	}
	if matching.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *ReadingMatchingService) Create(ctx context.Context, matching *reading.ReadingMatching) error {
	if err := s.validateMatching(matching); err != nil {
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

	if err := s.repo.Create(ctx, matching); err != nil {
		tx.Rollback()
		s.logger.Error("reading_matching_service.create", map[string]interface{}{
			"error": err.Error(),
			"id":    matching.ID,
		}, "Failed to create matching")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, matching.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_matching_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    matching.ReadingQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_matching_service.update_cache", map[string]interface{}{
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

func (s *ReadingMatchingService) GetByID(ctx context.Context, id uuid.UUID) (*reading.ReadingMatching, error) {
	matching, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("reading_matching_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get matching by ID")
		return nil, err
	}
	return matching, nil
}

func (s *ReadingMatchingService) Update(ctx context.Context, matching *reading.ReadingMatching) error {
	if err := s.validateMatching(matching); err != nil {
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

	if err := s.repo.Update(ctx, matching); err != nil {
		tx.Rollback()
		s.logger.Error("reading_matching_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    matching.ID,
		}, "Failed to update matching")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, matching.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_matching_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    matching.ReadingQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_matching_service.update_cache", map[string]interface{}{
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

func (s *ReadingMatchingService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get matching first to get parent ID
	matching, err := s.GetByID(ctx, id)
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
		s.logger.Error("reading_matching_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete matching")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, matching.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_matching_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    matching.ReadingQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_matching_service.update_cache", map[string]interface{}{
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

func (s *ReadingMatchingService) GetByReadingQuestionID(ctx context.Context, readingQuestionID uuid.UUID) ([]*reading.ReadingMatching, error) {
	matchings, err := s.repo.GetByReadingQuestionID(ctx, readingQuestionID)
	if err != nil {
		s.logger.Error("reading_matching_service.get_by_reading_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    readingQuestionID,
		}, "Failed to get matchings")
		return nil, err
	}
	return matchings, nil
}
