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

type SpeakingPhraseRepetitionService struct {
	repo                 *speakingRepository.SpeakingPhraseRepetitionRepository
	speakingQuestionRepo *speakingRepository.SpeakingQuestionRepository
	logger               *logger.PrettyLogger
	cache                cache.Cache
	questionUpdator      *speakingHelper.SpeakingQuestionUpdator
}

func NewSpeakingPhraseRepetitionService(
	repo *speakingRepository.SpeakingPhraseRepetitionRepository,
	speakingQuestionRepo *speakingRepository.SpeakingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *speakingHelper.SpeakingQuestionUpdator,
) *SpeakingPhraseRepetitionService {
	return &SpeakingPhraseRepetitionService{
		repo:                 repo,
		speakingQuestionRepo: speakingQuestionRepo,
		logger:               logger,
		cache:                cache,
		questionUpdator:      questionUpdator,
	}
}

func (s *SpeakingPhraseRepetitionService) SetQuestionUpdator(updator *speakingHelper.SpeakingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *SpeakingPhraseRepetitionService) validatePhraseRepetition(phrase *speaking.SpeakingPhraseRepetition) error {
	if phrase == nil {
		return errors.New("invalid input")
	}
	if phrase.SpeakingQuestionID == uuid.Nil {
		return errors.New("speaking question ID is required")
	}
	if phrase.Phrase == "" {
		return errors.New("phrase text is required")
	}
	if phrase.Mean == "" {
		return errors.New("meaning is required")
	}
	return nil
}

func (s *SpeakingPhraseRepetitionService) Create(ctx context.Context, phrase *speaking.SpeakingPhraseRepetition) error {
	if err := s.validatePhraseRepetition(phrase); err != nil {
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

	if err := s.repo.Create(ctx, phrase); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_phrase_repetition_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create phrase repetition")
		return err
	}

	// Get parent speaking question for cache/search update
	parentQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, phrase.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_phrase_repetition_service.update_cache", map[string]interface{}{
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

func (s *SpeakingPhraseRepetitionService) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingPhraseRepetition, error) {
	phrase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("speaking_phrase_repetition_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get phrase repetition by ID")
		return nil, err
	}
	return phrase, nil
}

func (s *SpeakingPhraseRepetitionService) GetBySpeakingQuestionID(ctx context.Context, speakingQuestionID uuid.UUID) ([]*speaking.SpeakingPhraseRepetition, error) {
	phrases, err := s.repo.GetBySpeakingQuestionID(ctx, speakingQuestionID)
	if err != nil {
		s.logger.Error("speaking_phrase_repetition_service.get_by_speaking_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    speakingQuestionID,
		}, "Failed to get phrase repetitions")
		return nil, err
	}
	return phrases, nil
}

func (s *SpeakingPhraseRepetitionService) Update(ctx context.Context, phrase *speaking.SpeakingPhraseRepetition) error {
	if err := s.validatePhraseRepetition(phrase); err != nil {
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

	if err := s.repo.Update(ctx, phrase); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_phrase_repetition_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    phrase.ID,
		}, "Failed to update phrase repetition")
		return err
	}

	// Get parent speaking question for cache/search update
	parentQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, phrase.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_phrase_repetition_service.update_cache", map[string]interface{}{
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

func (s *SpeakingPhraseRepetitionService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get phrase before deletion to get parent ID
	phrase, err := s.GetByID(ctx, id)
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
		s.logger.Error("speaking_phrase_repetition_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete phrase repetition")
		return err
	}

	// Get parent speaking question for cache/search update
	parentQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, phrase.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_phrase_repetition_service.update_cache", map[string]interface{}{
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
