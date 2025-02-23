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

type WritingSentenceCompletionService struct {
	repo                *writingRepository.WritingSentenceCompletionRepository
	writingQuestionRepo *writingRepository.WritingQuestionRepository
	logger              *logger.PrettyLogger
	cache               cache.Cache
	questionUpdator     *writingHelper.WritingQuestionUpdator
}

func NewWritingSentenceCompletionService(
	repo *writingRepository.WritingSentenceCompletionRepository,
	writingQuestionRepo *writingRepository.WritingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *writingHelper.WritingQuestionUpdator,
) *WritingSentenceCompletionService {
	return &WritingSentenceCompletionService{
		repo:                repo,
		writingQuestionRepo: writingQuestionRepo,
		logger:              logger,
		cache:               cache,
		questionUpdator:     questionUpdator,
	}
}

func (s *WritingSentenceCompletionService) SetQuestionUpdator(updator *writingHelper.WritingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *WritingSentenceCompletionService) validateSentenceCompletion(sentence *writing.WritingSentenceCompletion) error {
	if sentence == nil {
		return errors.New("invalid input")
	}
	if sentence.WritingQuestionID == uuid.Nil {
		return errors.New("writing question ID is required")
	}
	if sentence.ExampleSentence == "" {
		return errors.New("example sentence is required")
	}
	if sentence.GivenPartSentence == "" {
		return errors.New("given part sentence is required")
	}
	if sentence.Position != "start" && sentence.Position != "end" {
		return errors.New("position must be either 'start' or 'end'")
	}
	if len(sentence.RequiredWords) == 0 {
		return errors.New("at least one required word is needed")
	}
	if sentence.Explain == "" {
		return errors.New("explanation is required")
	}
	if sentence.MinWords < 1 {
		return errors.New("minimum words must be at least 1")
	}
	if sentence.MaxWords < sentence.MinWords {
		return errors.New("maximum words must be greater than minimum words")
	}
	return nil
}

func (s *WritingSentenceCompletionService) Create(ctx context.Context, sentence *writing.WritingSentenceCompletion) error {
	if err := s.validateSentenceCompletion(sentence); err != nil {
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

	if err := s.repo.Create(ctx, sentence); err != nil {
		tx.Rollback()
		s.logger.Error("writing_sentence_completion_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create sentence completion")
		return err
	}

	// Get parent writing question for cache/search update
	parentQuestion, err := s.writingQuestionRepo.GetWritingQuestionByID(ctx, sentence.WritingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get writing question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("writing_sentence_completion_service.update_cache", map[string]interface{}{
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

func (s *WritingSentenceCompletionService) GetByID(ctx context.Context, id uuid.UUID) (*writing.WritingSentenceCompletion, error) {
	sentence, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("writing_sentence_completion_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get sentence completion by ID")
		return nil, err
	}
	return sentence, nil
}

func (s *WritingSentenceCompletionService) GetByWritingQuestionID(ctx context.Context, writingQuestionID uuid.UUID) ([]*writing.WritingSentenceCompletion, error) {
	sentences, err := s.repo.GetByWritingQuestionID(ctx, writingQuestionID)
	if err != nil {
		s.logger.Error("writing_sentence_completion_service.get_by_writing_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    writingQuestionID,
		}, "Failed to get sentence completions")
		return nil, err
	}
	return sentences, nil
}

func (s *WritingSentenceCompletionService) Update(ctx context.Context, sentence *writing.WritingSentenceCompletion) error {
	if err := s.validateSentenceCompletion(sentence); err != nil {
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

	if err := s.repo.Update(ctx, sentence); err != nil {
		tx.Rollback()
		s.logger.Error("writing_sentence_completion_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    sentence.ID,
		}, "Failed to update sentence completion")
		return err
	}

	// Get parent writing question for cache/search update
	parentQuestion, err := s.writingQuestionRepo.GetWritingQuestionByID(ctx, sentence.WritingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get writing question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("writing_sentence_completion_service.update_cache", map[string]interface{}{
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

func (s *WritingSentenceCompletionService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get sentence first to get parent ID
	sentence, err := s.GetByID(ctx, id)
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
		s.logger.Error("writing_sentence_completion_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete sentence completion")
		return err
	}

	// Get parent writing question for cache/search update
	parentQuestion, err := s.writingQuestionRepo.GetWritingQuestionByID(ctx, sentence.WritingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get writing question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("writing_sentence_completion_service.update_cache", map[string]interface{}{
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
