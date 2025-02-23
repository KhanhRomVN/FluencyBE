package listening

import (
	"context"
	"database/sql"
	"errors"
	listeningHelper "fluencybe/internal/app/helper/listening"
	"fluencybe/internal/app/model/listening"
	ListeningRepository "fluencybe/internal/app/repository/listening"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListeningFillInTheBlankAnswerService struct {
	repo                  *ListeningRepository.ListeningFillInTheBlankAnswerRepository
	questionRepo          *ListeningRepository.ListeningFillInTheBlankQuestionRepository
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository
	logger                *logger.PrettyLogger
	cache                 cache.Cache
	questionUpdator       *listeningHelper.ListeningQuestionUpdator
}

func NewListeningFillInTheBlankAnswerService(
	repo *ListeningRepository.ListeningFillInTheBlankAnswerRepository,
	questionRepo *ListeningRepository.ListeningFillInTheBlankQuestionRepository,
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *listeningHelper.ListeningQuestionUpdator,
) *ListeningFillInTheBlankAnswerService {
	return &ListeningFillInTheBlankAnswerService{
		repo:                  repo,
		questionRepo:          questionRepo,
		listeningQuestionRepo: listeningQuestionRepo,
		logger:                logger,
		cache:                 cache,
		questionUpdator:       questionUpdator,
	}
}

func (s *ListeningFillInTheBlankAnswerService) SetQuestionUpdator(updator *listeningHelper.ListeningQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ListeningFillInTheBlankAnswerService) validateAnswer(answer *listening.ListeningFillInTheBlankAnswer) error {
	if answer == nil {
		return errors.New("invalid input")
	}
	if answer.ListeningFillInTheBlankQuestionID == uuid.Nil {
		return errors.New("question ID is required")
	}
	if answer.Answer == "" {
		return errors.New("answer text is required")
	}
	if answer.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *ListeningFillInTheBlankAnswerService) CreateAnswer(ctx context.Context, answer *listening.ListeningFillInTheBlankAnswer) error {
	if err := s.validateAnswer(answer); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Get the parent fill in blank question
	fillInBlankQuestion, err := s.questionRepo.GetByID(ctx, answer.ListeningFillInTheBlankQuestionID)
	if err != nil {
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Create answer in database
	if err := s.repo.Create(ctx, answer); err != nil {
		s.logger.Error("listening_fill_in_the_blank_answer_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create answer")
		return err
	}

	// Get parent listening question for cache/search update
	parentQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, fillInBlankQuestion.ListeningQuestionID)
	if err != nil {
		s.logger.Error("listening_fill_in_the_blank_answer_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    fillInBlankQuestion.ListeningQuestionID,
		}, "Failed to get parent listening question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		s.logger.Error("listening_fill_in_the_blank_answer_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
	}

	return nil
}

func (s *ListeningFillInTheBlankAnswerService) GetAnswer(ctx context.Context, id uuid.UUID) (*listening.ListeningFillInTheBlankAnswer, error) {
	answer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		s.logger.Error("listening_fill_in_the_blank_answer_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get answer")
		return nil, err
	}
	return answer, nil
}

func (s *ListeningFillInTheBlankAnswerService) UpdateAnswer(ctx context.Context, answer *listening.ListeningFillInTheBlankAnswer) error {
	if err := s.validateAnswer(answer); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Get the parent fill in blank question
	fillInBlankQuestion, err := s.questionRepo.GetByID(ctx, answer.ListeningFillInTheBlankQuestionID)
	if err != nil {
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Update answer in database
	if err := s.repo.Update(ctx, answer); err != nil {
		s.logger.Error("listening_fill_in_the_blank_answer_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    answer.ID,
		}, "Failed to update answer")
		return err
	}

	// Get parent listening question for cache/search update
	parentQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, fillInBlankQuestion.ListeningQuestionID)
	if err != nil {
		s.logger.Error("listening_fill_in_the_blank_answer_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    fillInBlankQuestion.ListeningQuestionID,
		}, "Failed to get parent listening question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		s.logger.Error("listening_fill_in_the_blank_answer_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
	}

	return nil
}

func (s *ListeningFillInTheBlankAnswerService) DeleteAnswer(ctx context.Context, id uuid.UUID) error {
	// Get answer before deletion to get parent question ID
	answer, err := s.GetAnswer(ctx, id)
	if err != nil {
		return err
	}

	// Get the parent fill in blank question
	fillInBlankQuestion, err := s.questionRepo.GetByID(ctx, answer.ListeningFillInTheBlankQuestionID)
	if err != nil {
		return fmt.Errorf("failed to get parent question: %w", err)
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

	// Delete from database
	if err := s.repo.Delete(ctx, id); err != nil {
		tx.Rollback()
		s.logger.Error("listening_fill_in_the_blank_answer_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete answer")
		return err
	}

	// Get parent listening question for cache/search update
	// Important: Get fresh copy after deletion to get updated version
	parentQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, fillInBlankQuestion.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("listening_fill_in_the_blank_answer_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    fillInBlankQuestion.ListeningQuestionID,
		}, "Failed to get parent listening question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_fill_in_the_blank_answer_service.update_cache", map[string]interface{}{
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

func (s *ListeningFillInTheBlankAnswerService) GetAnswersByListeningFillInTheBlankQuestionID(ctx context.Context, questionID uuid.UUID) ([]*listening.ListeningFillInTheBlankAnswer, error) {
	answers, err := s.repo.GetByListeningFillInTheBlankQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("listening_fill_in_the_blank_answer_service.get_by_question", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get answers")
		return nil, err
	}
	return answers, nil
}
