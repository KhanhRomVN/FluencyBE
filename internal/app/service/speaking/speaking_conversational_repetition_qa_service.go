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

type SpeakingConversationalRepetitionQAService struct {
	repo                         *speakingRepository.SpeakingConversationalRepetitionQARepository
	conversationalRepetitionRepo *speakingRepository.SpeakingConversationalRepetitionRepository
	speakingQuestionRepo         *speakingRepository.SpeakingQuestionRepository
	logger                       *logger.PrettyLogger
	cache                        cache.Cache
	questionUpdator              *speakingHelper.SpeakingQuestionUpdator
}

func NewSpeakingConversationalRepetitionQAService(
	repo *speakingRepository.SpeakingConversationalRepetitionQARepository,
	conversationalRepetitionRepo *speakingRepository.SpeakingConversationalRepetitionRepository,
	speakingQuestionRepo *speakingRepository.SpeakingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *speakingHelper.SpeakingQuestionUpdator,
) *SpeakingConversationalRepetitionQAService {
	return &SpeakingConversationalRepetitionQAService{
		repo:                         repo,
		conversationalRepetitionRepo: conversationalRepetitionRepo,
		speakingQuestionRepo:         speakingQuestionRepo,
		logger:                       logger,
		cache:                        cache,
		questionUpdator:              questionUpdator,
	}
}

func (s *SpeakingConversationalRepetitionQAService) SetQuestionUpdator(updator *speakingHelper.SpeakingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *SpeakingConversationalRepetitionQAService) validateQA(qa *speaking.SpeakingConversationalRepetitionQA) error {
	if qa == nil {
		return errors.New("invalid input")
	}
	if qa.SpeakingConversationalRepetitionID == uuid.Nil {
		return errors.New("conversational repetition ID is required")
	}
	if qa.Question == "" {
		return errors.New("question text is required")
	}
	if qa.Answer == "" {
		return errors.New("answer text is required")
	}
	if qa.MeanOfQuestion == "" {
		return errors.New("meaning of question is required")
	}
	if qa.MeanOfAnswer == "" {
		return errors.New("meaning of answer is required")
	}
	if qa.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *SpeakingConversationalRepetitionQAService) Create(ctx context.Context, qa *speaking.SpeakingConversationalRepetitionQA) error {
	if err := s.validateQA(qa); err != nil {
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

	if err := s.repo.Create(ctx, qa); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_repetition_qa_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create QA")
		return err
	}

	// Get parent conversational repetition
	parentRepetition, err := s.conversationalRepetitionRepo.GetByID(ctx, qa.SpeakingConversationalRepetitionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent repetition: %w", err)
	}

	// Get speaking question for cache/search update
	speakingQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, parentRepetition.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, speakingQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_repetition_qa_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    speakingQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SpeakingConversationalRepetitionQAService) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingConversationalRepetitionQA, error) {
	qa, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("speaking_conversational_repetition_qa_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get QA by ID")
		return nil, err
	}
	return qa, nil
}

func (s *SpeakingConversationalRepetitionQAService) GetBySpeakingConversationalRepetitionID(ctx context.Context, repetitionID uuid.UUID) ([]*speaking.SpeakingConversationalRepetitionQA, error) {
	qas, err := s.repo.GetBySpeakingConversationalRepetitionID(ctx, repetitionID)
	if err != nil {
		s.logger.Error("speaking_conversational_repetition_qa_service.get_by_repetition_id", map[string]interface{}{
			"error": err.Error(),
			"id":    repetitionID,
		}, "Failed to get QAs")
		return nil, err
	}
	return qas, nil
}

func (s *SpeakingConversationalRepetitionQAService) Update(ctx context.Context, qa *speaking.SpeakingConversationalRepetitionQA) error {
	if err := s.validateQA(qa); err != nil {
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

	if err := s.repo.Update(ctx, qa); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_repetition_qa_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    qa.ID,
		}, "Failed to update QA")
		return err
	}

	// Get parent conversational repetition
	parentRepetition, err := s.conversationalRepetitionRepo.GetByID(ctx, qa.SpeakingConversationalRepetitionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent repetition: %w", err)
	}

	// Get speaking question for cache/search update
	speakingQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, parentRepetition.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, speakingQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_repetition_qa_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    speakingQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SpeakingConversationalRepetitionQAService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get QA first to get parent IDs
	qa, err := s.GetByID(ctx, id)
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
		s.logger.Error("speaking_conversational_repetition_qa_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete QA")
		return err
	}

	// Get parent conversational repetition
	parentRepetition, err := s.conversationalRepetitionRepo.GetByID(ctx, qa.SpeakingConversationalRepetitionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent repetition: %w", err)
	}

	// Get speaking question for cache/search update
	speakingQuestion, err := s.speakingQuestionRepo.GetSpeakingQuestionByID(ctx, parentRepetition.SpeakingQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get speaking question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, speakingQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_conversational_repetition_qa_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    speakingQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
