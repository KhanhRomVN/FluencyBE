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

type SpeakingConversationalOpenService struct {
	repo                 *speakingRepository.SpeakingConversationalOpenRepository
	speakingQuestionRepo *speakingRepository.SpeakingQuestionRepository
	logger               *logger.PrettyLogger
	cache                cache.Cache
	questionUpdator      *speakingHelper.SpeakingQuestionUpdator
}

func NewSpeakingConversationalOpenService(
	repo *speakingRepository.SpeakingConversationalOpenRepository,
	speakingQuestionRepo *speakingRepository.SpeakingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *speakingHelper.SpeakingQuestionUpdator,
) *SpeakingConversationalOpenService {
	return &SpeakingConversationalOpenService{
		repo:                 repo,
		speakingQuestionRepo: speakingQuestionRepo,
		logger:               logger,
		cache:                cache,
		questionUpdator:      questionUpdator,
	}
}

func (s *SpeakingConversationalOpenService) SetQuestionUpdator(updator *speakingHelper.SpeakingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *SpeakingConversationalOpenService) validateConversationalOpen(conversationalOpen *speaking.SpeakingConversationalOpen) error {
	if conversationalOpen == nil {
		return errors.New("invalid input")
	}
	if conversationalOpen.SpeakingQuestionID == uuid.Nil {
		return errors.New("speaking question ID is required")
	}
	if conversationalOpen.Title == "" {
		return errors.New("title is required")
	}
	if conversationalOpen.Overview == "" {
		return errors.New("overview is required")
	}
	if conversationalOpen.ExampleConversation == "" {
		return errors.New("example conversation is required")
	}
	return nil
}

func (s *SpeakingConversationalOpenService) Create(ctx context.Context, conversationalOpen *speaking.SpeakingConversationalOpen) error {
	if err := s.validateConversationalOpen(conversationalOpen); err != nil {
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

	if err := s.repo.Create(ctx, conversationalOpen); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_open_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create conversational open")
		return err
	}

	// Get parent speaking question for cache/search update
	parentQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, conversationalOpen.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_open_service.update_cache", map[string]interface{}{
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

func (s *SpeakingConversationalOpenService) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingConversationalOpen, error) {
	conversationalOpen, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("speaking_conversational_open_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get conversational open by ID")
		return nil, err
	}
	return conversationalOpen, nil
}

func (s *SpeakingConversationalOpenService) GetBySpeakingQuestionID(ctx context.Context, speakingQuestionID uuid.UUID) ([]*speaking.SpeakingConversationalOpen, error) {
	conversationalOpens, err := s.repo.GetBySpeakingQuestionID(ctx, speakingQuestionID)
	if err != nil {
		s.logger.Error("speaking_conversational_open_service.get_by_speaking_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    speakingQuestionID,
		}, "Failed to get conversational opens")
		return nil, err
	}
	return conversationalOpens, nil
}

func (s *SpeakingConversationalOpenService) Update(ctx context.Context, conversationalOpen *speaking.SpeakingConversationalOpen) error {
	if err := s.validateConversationalOpen(conversationalOpen); err != nil {
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

	if err := s.repo.Update(ctx, conversationalOpen); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_open_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    conversationalOpen.ID,
		}, "Failed to update conversational open")
		return err
	}

	// Get parent speaking question for cache/search update
	parentQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, conversationalOpen.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_open_service.update_cache", map[string]interface{}{
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

func (s *SpeakingConversationalOpenService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get conversational open first to get parent ID
	conversationalOpen, err := s.GetByID(ctx, id)
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
		s.logger.Error("speaking_conversational_open_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete conversational open")
		return err
	}

	// Get parent speaking question for cache/search update
	parentQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, conversationalOpen.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_open_service.update_cache", map[string]interface{}{
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
