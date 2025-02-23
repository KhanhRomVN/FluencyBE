package reading

import (
	"context"
	"database/sql"
	"errors"
	readingHelper "fluencybe/internal/app/helper/reading"
	"fluencybe/internal/app/model/reading"
	ReadingRepository "fluencybe/internal/app/repository/reading"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReadingChoiceMultiQuestionService struct {
	repo                *ReadingRepository.ReadingChoiceMultiQuestionRepository
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository
	logger              *logger.PrettyLogger
	questionUpdator     *readingHelper.ReadingQuestionUpdator
}

func NewReadingChoiceMultiQuestionService(
	repo *ReadingRepository.ReadingChoiceMultiQuestionRepository,
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository,
	logger *logger.PrettyLogger,
	questionUpdator *readingHelper.ReadingQuestionUpdator,
) *ReadingChoiceMultiQuestionService {
	return &ReadingChoiceMultiQuestionService{
		repo:                repo,
		readingQuestionRepo: readingQuestionRepo,
		logger:              logger,
		questionUpdator:     questionUpdator,
	}
}

func (s *ReadingChoiceMultiQuestionService) SetQuestionUpdator(updator *readingHelper.ReadingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ReadingChoiceMultiQuestionService) validateQuestion(question *reading.ReadingChoiceMultiQuestion) error {
	if question == nil {
		return errors.New("invalid input")
	}
	if question.ReadingQuestionID == uuid.Nil {
		return errors.New("reading question ID is required")
	}
	if question.Question == "" {
		return errors.New("question text is required")
	}
	if question.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *ReadingChoiceMultiQuestionService) CreateQuestion(ctx context.Context, question *reading.ReadingChoiceMultiQuestion) error {
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
		s.logger.Error("reading_choice_multi_question_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, question.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_choice_multi_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ReadingQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_choice_multi_question_service.update_cache", map[string]interface{}{
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

func (s *ReadingChoiceMultiQuestionService) GetQuestion(ctx context.Context, id uuid.UUID) (*reading.ReadingChoiceMultiQuestion, error) {
	question, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		s.logger.Error("reading_choice_multi_question_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get question")
		return nil, err
	}
	return question, nil
}

func (s *ReadingChoiceMultiQuestionService) UpdateQuestion(ctx context.Context, question *reading.ReadingChoiceMultiQuestion) error {
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
		s.logger.Error("reading_choice_multi_question_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update question")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, question.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_choice_multi_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ReadingQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_choice_multi_question_service.update_cache", map[string]interface{}{
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

func (s *ReadingChoiceMultiQuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
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
		s.logger.Error("reading_choice_multi_question_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, question.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_choice_multi_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ReadingQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_choice_multi_question_service.update_cache", map[string]interface{}{
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

func (s *ReadingChoiceMultiQuestionService) GetQuestionsByReadingQuestionID(ctx context.Context, readingQuestionID uuid.UUID) ([]*reading.ReadingChoiceMultiQuestion, error) {
	questions, err := s.repo.GetByReadingQuestionID(ctx, readingQuestionID)
	if err != nil {
		s.logger.Error("reading_choice_multi_question_service.get_by_reading_id", map[string]interface{}{
			"error":               err.Error(),
			"reading_question_id": readingQuestionID,
		}, "Failed to get questions by reading question ID")
		return nil, err
	}
	return questions, nil
}
