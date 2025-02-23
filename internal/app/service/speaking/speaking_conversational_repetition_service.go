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

type SpeakingConversationalRepetitionService struct {
	repo                 *speakingRepository.SpeakingConversationalRepetitionRepository
	speakingQuestionRepo *speakingRepository.SpeakingQuestionRepository
	logger               *logger.PrettyLogger
	cache                cache.Cache
	questionUpdator      *speakingHelper.SpeakingQuestionUpdator
}

func NewSpeakingConversationalRepetitionService(
	repo *speakingRepository.SpeakingConversationalRepetitionRepository,
	speakingQuestionRepo *speakingRepository.SpeakingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *speakingHelper.SpeakingQuestionUpdator,
) *SpeakingConversationalRepetitionService {
	return &SpeakingConversationalRepetitionService{
		repo:                 repo,
		speakingQuestionRepo: speakingQuestionRepo,
		logger:               logger,
		cache:                cache,
		questionUpdator:      questionUpdator,
	}
}

func (s *SpeakingConversationalRepetitionService) SetQuestionUpdator(updator *speakingHelper.SpeakingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *SpeakingConversationalRepetitionService) validateConversationalRepetition(conversation *speaking.SpeakingConversationalRepetition) error {
	if conversation == nil {
		return errors.New("invalid input")
	}
	if conversation.SpeakingQuestionID == uuid.Nil {
		return errors.New("speaking question ID is required")
	}
	if conversation.Title == "" {
		return errors.New("title is required")
	}
	if conversation.Overview == "" {
		return errors.New("overview is required")
	}
	return nil
}

func (s *SpeakingConversationalRepetitionService) Create(ctx context.Context, conversation *speaking.SpeakingConversationalRepetition) error {
	if err := s.validateConversationalRepetition(conversation); err != nil {
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

	if err := s.repo.Create(ctx, conversation); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_repetition_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create conversational repetition")
		return err
	}

	// Get parent speaking question for cache/search update
	parentQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, conversation.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_repetition_service.update_cache", map[string]interface{}{
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

func (s *SpeakingConversationalRepetitionService) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingConversationalRepetition, error) {
	conversation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("speaking_conversational_repetition_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get conversational repetition by ID")
		return nil, err
	}
	return conversation, nil
}

func (s *SpeakingConversationalRepetitionService) GetBySpeakingQuestionID(ctx context.Context, speakingQuestionID uuid.UUID) ([]*speaking.SpeakingConversationalRepetition, error) {
	conversations, err := s.repo.GetBySpeakingQuestionID(ctx, speakingQuestionID)
	if err != nil {
		s.logger.Error("speaking_conversational_repetition_service.get_by_speaking_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    speakingQuestionID,
		}, "Failed to get conversational repetitions")
		return nil, err
	}
	return conversations, nil
}

func (s *SpeakingConversationalRepetitionService) Update(ctx context.Context, conversation *speaking.SpeakingConversationalRepetition) error {
	if err := s.validateConversationalRepetition(conversation); err != nil {
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

	if err := s.repo.Update(ctx, conversation); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_repetition_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    conversation.ID,
		}, "Failed to update conversational repetition")
		return err
	}

	// Get parent speaking question for cache/search update
	parentQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, conversation.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_repetition_service.update_cache", map[string]interface{}{
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

func (s *SpeakingConversationalRepetitionService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get conversation first to get parent ID
	conversation, err := s.GetByID(ctx, id)
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
		s.logger.Error("speaking_conversational_repetition_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete conversational repetition")
		return err
	}

	// Get parent speaking question for cache/search update
	parentQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, conversation.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_repetition_service.update_cache", map[string]interface{}{
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
