package speaking

import (
	"context"
	"errors"
	speakingHelper "fluencybe/internal/app/helper/speaking"
	"fluencybe/internal/app/model/speaking"
	speakingRepository "fluencybe/internal/app/repository/speaking"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
)

type SpeakingOpenParagraphService struct {
	repo                 *speakingRepository.SpeakingOpenParagraphRepository
	speakingQuestionRepo *speakingRepository.SpeakingQuestionRepository
	logger               *logger.PrettyLogger
	cache                cache.Cache
	questionUpdator      *speakingHelper.SpeakingQuestionUpdator
}

func NewSpeakingOpenParagraphService(
	repo *speakingRepository.SpeakingOpenParagraphRepository,
	speakingQuestionRepo *speakingRepository.SpeakingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *speakingHelper.SpeakingQuestionUpdator,
) *SpeakingOpenParagraphService {
	return &SpeakingOpenParagraphService{
		repo:                 repo,
		speakingQuestionRepo: speakingQuestionRepo,
		logger:               logger,
		cache:                cache,
		questionUpdator:      questionUpdator,
	}
}

func (s *SpeakingOpenParagraphService) SetQuestionUpdator(updator *speakingHelper.SpeakingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *SpeakingOpenParagraphService) validateOpenParagraph(paragraph *speaking.SpeakingOpenParagraph) error {
	if paragraph == nil {
		return errors.New("invalid input")
	}
	if paragraph.SpeakingQuestionID == uuid.Nil {
		return errors.New("speaking question ID is required")
	}
	if paragraph.Question == "" {
		return errors.New("question text is required")
	}
	if paragraph.ExamplePassage == "" {
		return errors.New("example passage is required")
	}
	if paragraph.MeanOfExamplePassage == "" {
		return errors.New("meaning of example passage is required")
	}
	return nil
}

func (s *SpeakingOpenParagraphService) Create(ctx context.Context, paragraph *speaking.SpeakingOpenParagraph) error {
	if err := s.validateOpenParagraph(paragraph); err != nil {
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

	if err := s.repo.Create(ctx, paragraph); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_open_paragraph_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create open paragraph")
		return err
	}

	// Get parent speaking question for cache/search update
	parentQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, paragraph.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_open_paragraph_service.update_cache", map[string]interface{}{
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

func (s *SpeakingOpenParagraphService) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingOpenParagraph, error) {
	paragraph, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("speaking_open_paragraph_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get open paragraph by ID")
		return nil, err
	}
	return paragraph, nil
}

func (s *SpeakingOpenParagraphService) GetBySpeakingQuestionID(ctx context.Context, speakingQuestionID uuid.UUID) ([]*speaking.SpeakingOpenParagraph, error) {
	paragraphs, err := s.repo.GetBySpeakingQuestionID(ctx, speakingQuestionID)
	if err != nil {
		s.logger.Error("speaking_open_paragraph_service.get_by_speaking_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    speakingQuestionID,
		}, "Failed to get open paragraphs")
		return nil, err
	}
	return paragraphs, nil
}

func (s *SpeakingOpenParagraphService) Update(ctx context.Context, paragraph *speaking.SpeakingOpenParagraph) error {
	if err := s.validateOpenParagraph(paragraph); err != nil {
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

	if err := s.repo.Update(ctx, paragraph); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_open_paragraph_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    paragraph.ID,
		}, "Failed to update open paragraph")
		return err
	}

	// Get parent speaking question for cache/search update
	parentQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, paragraph.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_open_paragraph_service.update_cache", map[string]interface{}{
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

func (s *SpeakingOpenParagraphService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get paragraph first to get parent ID
	paragraph, err := s.GetByID(ctx, id)
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
		s.logger.Error("speaking_open_paragraph_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete open paragraph")
		return err
	}

	// Get parent speaking question for cache/search update
	parentQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, paragraph.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_open_paragraph_service.update_cache", map[string]interface{}{
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
