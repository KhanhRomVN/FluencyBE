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

type SpeakingParagraphRepetitionService struct {
	repo                 *speakingRepository.SpeakingParagraphRepetitionRepository
	speakingQuestionRepo *speakingRepository.SpeakingQuestionRepository
	logger               *logger.PrettyLogger
	cache                cache.Cache
	questionUpdator      *speakingHelper.SpeakingQuestionUpdator
}

func NewSpeakingParagraphRepetitionService(
	repo *speakingRepository.SpeakingParagraphRepetitionRepository,
	speakingQuestionRepo *speakingRepository.SpeakingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *speakingHelper.SpeakingQuestionUpdator,
) *SpeakingParagraphRepetitionService {
	return &SpeakingParagraphRepetitionService{
		repo:                 repo,
		speakingQuestionRepo: speakingQuestionRepo,
		logger:               logger,
		cache:                cache,
		questionUpdator:      questionUpdator,
	}
}

func (s *SpeakingParagraphRepetitionService) SetQuestionUpdator(updator *speakingHelper.SpeakingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *SpeakingParagraphRepetitionService) validateParagraphRepetition(paragraph *speaking.SpeakingParagraphRepetition) error {
	if paragraph == nil {
		return errors.New("invalid input")
	}
	if paragraph.SpeakingQuestionID == uuid.Nil {
		return errors.New("speaking question ID is required")
	}
	if paragraph.Paragraph == "" {
		return errors.New("paragraph text is required")
	}
	if paragraph.Mean == "" {
		return errors.New("meaning is required")
	}
	return nil
}

func (s *SpeakingParagraphRepetitionService) Create(ctx context.Context, paragraph *speaking.SpeakingParagraphRepetition) error {
	if err := s.validateParagraphRepetition(paragraph); err != nil {
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
		s.logger.Error("speaking_paragraph_repetition_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create paragraph repetition")
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
		s.logger.Error("speaking_paragraph_repetition_service.update_cache", map[string]interface{}{
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

func (s *SpeakingParagraphRepetitionService) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingParagraphRepetition, error) {
	paragraph, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("speaking_paragraph_repetition_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get paragraph repetition by ID")
		return nil, err
	}
	return paragraph, nil
}

func (s *SpeakingParagraphRepetitionService) GetBySpeakingQuestionID(ctx context.Context, speakingQuestionID uuid.UUID) ([]*speaking.SpeakingParagraphRepetition, error) {
	paragraphs, err := s.repo.GetBySpeakingQuestionID(ctx, speakingQuestionID)
	if err != nil {
		s.logger.Error("speaking_paragraph_repetition_service.get_by_speaking_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    speakingQuestionID,
		}, "Failed to get paragraph repetitions")
		return nil, err
	}
	return paragraphs, nil
}

func (s *SpeakingParagraphRepetitionService) Update(ctx context.Context, paragraph *speaking.SpeakingParagraphRepetition) error {
	if err := s.validateParagraphRepetition(paragraph); err != nil {
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
		s.logger.Error("speaking_paragraph_repetition_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    paragraph.ID,
		}, "Failed to update paragraph repetition")
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
		s.logger.Error("speaking_paragraph_repetition_service.update_cache", map[string]interface{}{
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

func (s *SpeakingParagraphRepetitionService) Delete(ctx context.Context, id uuid.UUID) error {
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
		s.logger.Error("speaking_paragraph_repetition_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete paragraph repetition")
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
		s.logger.Error("speaking_paragraph_repetition_service.update_cache", map[string]interface{}{
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
