package listening

import (
	"context"
	"database/sql"
	"errors"
	listeningHelper "fluencybe/internal/app/helper/listening"
	"fluencybe/internal/app/model/listening"
	ListeningRepository "fluencybe/internal/app/repository/listening"
	"fluencybe/pkg/logger"

	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListeningChoiceMultiQuestionService struct {
	repo                  *ListeningRepository.ListeningChoiceMultiQuestionRepository
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository
	logger                *logger.PrettyLogger
	questionUpdator       *listeningHelper.ListeningQuestionUpdator
}

func NewListeningChoiceMultiQuestionService(
	repo *ListeningRepository.ListeningChoiceMultiQuestionRepository,
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository,
	logger *logger.PrettyLogger,
	questionUpdator *listeningHelper.ListeningQuestionUpdator,
) *ListeningChoiceMultiQuestionService {
	return &ListeningChoiceMultiQuestionService{
		repo:                  repo,
		listeningQuestionRepo: listeningQuestionRepo,
		logger:                logger,
		questionUpdator:       questionUpdator,
	}
}

func (s *ListeningChoiceMultiQuestionService) SetQuestionUpdator(updator *listeningHelper.ListeningQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ListeningChoiceMultiQuestionService) validateQuestion(question *listening.ListeningChoiceMultiQuestion) error {

	if question == nil {
		return errors.New("invalid input")
	}
	if question.ListeningQuestionID == uuid.Nil {
		return errors.New("listening question ID is required")
	}
	if question.Question == "" {
		return errors.New("question text is required")
	}
	if question.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *ListeningChoiceMultiQuestionService) CreateQuestion(ctx context.Context, question *listening.ListeningChoiceMultiQuestion) error {
	if err := s.validateQuestion(question); err != nil {
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

	if err := s.repo.Create(ctx, question); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_question_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")
		return err
	}

	// Get parent listening question for cache/search update
	parentQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, question.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ListeningQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_question_service.update_cache", map[string]interface{}{
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

func (s *ListeningChoiceMultiQuestionService) GetQuestion(ctx context.Context, id uuid.UUID) (*listening.ListeningChoiceMultiQuestion, error) {
	question, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		s.logger.Error("listening_choice_multi_question_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get question")
		return nil, err
	}
	return question, nil
}

func (s *ListeningChoiceMultiQuestionService) UpdateQuestion(ctx context.Context, question *listening.ListeningChoiceMultiQuestion) error {
	if err := s.validateQuestion(question); err != nil {
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

	if err := s.repo.Update(ctx, question); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_question_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update question")
		return err
	}

	// Get parent listening question for cache/search update
	parentQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, question.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ListeningQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_question_service.update_cache", map[string]interface{}{
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

func (s *ListeningChoiceMultiQuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	// Get question before deletion to get parent ID
	question, err := s.GetQuestion(ctx, id)
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
		s.logger.Error("listening_choice_multi_question_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question")
		return err
	}

	// Get parent listening question for cache/search update
	parentQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, question.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ListeningQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_multi_question_service.update_cache", map[string]interface{}{
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

func (s *ListeningChoiceMultiQuestionService) GetQuestionsByListeningQuestionID(ctx context.Context, listeningQuestionID uuid.UUID) ([]*listening.ListeningChoiceMultiQuestion, error) {
	questions, err := s.repo.GetByListeningQuestionID(ctx, listeningQuestionID)
	if err != nil {
		s.logger.Error("listening_choice_multi_question_service.get_by_listening_id", map[string]interface{}{
			"error":                 err.Error(),
			"listening_question_id": listeningQuestionID,
		}, "Failed to get questions by listening question ID")
		return nil, err
	}
	return questions, nil
}
