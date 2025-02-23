package reading

import (
	"context"
	"database/sql"
	"errors"
	readingHelper "fluencybe/internal/app/helper/reading"
	"fluencybe/internal/app/model/reading"
	ReadingRepository "fluencybe/internal/app/repository/reading"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReadingFillInTheBlankAnswerService struct {
	repo                *ReadingRepository.ReadingFillInTheBlankAnswerRepository
	questionRepo        *ReadingRepository.ReadingFillInTheBlankQuestionRepository
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository
	logger              *logger.PrettyLogger
	cache               cache.Cache
	questionUpdator     *readingHelper.ReadingQuestionUpdator
}

func NewReadingFillInTheBlankAnswerService(
	repo *ReadingRepository.ReadingFillInTheBlankAnswerRepository,
	questionRepo *ReadingRepository.ReadingFillInTheBlankQuestionRepository,
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *readingHelper.ReadingQuestionUpdator,
) *ReadingFillInTheBlankAnswerService {
	return &ReadingFillInTheBlankAnswerService{
		repo:                repo,
		questionRepo:        questionRepo,
		readingQuestionRepo: readingQuestionRepo,
		logger:              logger,
		cache:               cache,
		questionUpdator:     questionUpdator,
	}
}

func (s *ReadingFillInTheBlankAnswerService) SetQuestionUpdator(updator *readingHelper.ReadingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ReadingFillInTheBlankAnswerService) validateAnswer(answer *reading.ReadingFillInTheBlankAnswer) error {
	if answer == nil {
		return errors.New("invalid input")
	}
	if answer.ReadingFillInTheBlankQuestionID == uuid.Nil {
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

func (s *ReadingFillInTheBlankAnswerService) CreateAnswer(ctx context.Context, answer *reading.ReadingFillInTheBlankAnswer) error {
	if err := s.validateAnswer(answer); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Get the parent fill in blank question
	fillInBlankQuestion, err := s.questionRepo.GetByID(ctx, answer.ReadingFillInTheBlankQuestionID)
	if err != nil {
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Create answer in database
	if err := s.repo.Create(ctx, answer); err != nil {
		s.logger.Error("reading_fill_in_the_blank_answer_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create answer")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, fillInBlankQuestion.ReadingQuestionID)
	if err != nil {
		s.logger.Error("reading_fill_in_the_blank_answer_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    fillInBlankQuestion.ReadingQuestionID,
		}, "Failed to get parent reading question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		s.logger.Error("reading_fill_in_the_blank_answer_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
	}

	return nil
}

func (s *ReadingFillInTheBlankAnswerService) GetAnswer(ctx context.Context, id uuid.UUID) (*reading.ReadingFillInTheBlankAnswer, error) {
	answer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		s.logger.Error("reading_fill_in_the_blank_answer_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get answer")
		return nil, err
	}
	return answer, nil
}

func (s *ReadingFillInTheBlankAnswerService) UpdateAnswer(ctx context.Context, answer *reading.ReadingFillInTheBlankAnswer) error {
	if err := s.validateAnswer(answer); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Get the parent fill in blank question
	fillInBlankQuestion, err := s.questionRepo.GetByID(ctx, answer.ReadingFillInTheBlankQuestionID)
	if err != nil {
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Update answer in database
	if err := s.repo.Update(ctx, answer); err != nil {
		s.logger.Error("reading_fill_in_the_blank_answer_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    answer.ID,
		}, "Failed to update answer")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, fillInBlankQuestion.ReadingQuestionID)
	if err != nil {
		s.logger.Error("reading_fill_in_the_blank_answer_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    fillInBlankQuestion.ReadingQuestionID,
		}, "Failed to get parent reading question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		s.logger.Error("reading_fill_in_the_blank_answer_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
	}

	return nil
}

func (s *ReadingFillInTheBlankAnswerService) DeleteAnswer(ctx context.Context, id uuid.UUID) error {
	// Get answer before deletion to get parent ID
	answer, err := s.GetAnswer(ctx, id)
	if err != nil {
		return err
	}

	// Get the parent fill in blank question
	fillInBlankQuestion, err := s.questionRepo.GetByID(ctx, answer.ReadingFillInTheBlankQuestionID)
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
		s.logger.Error("reading_fill_in_the_blank_answer_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete answer")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, fillInBlankQuestion.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_fill_in_the_blank_answer_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    fillInBlankQuestion.ReadingQuestionID,
		}, "Failed to get parent reading question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_fill_in_the_blank_answer_service.update_cache", map[string]interface{}{
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

func (s *ReadingFillInTheBlankAnswerService) GetAnswersByReadingFillInTheBlankQuestionID(ctx context.Context, questionID uuid.UUID) ([]*reading.ReadingFillInTheBlankAnswer, error) {
	answers, err := s.repo.GetByReadingFillInTheBlankQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("reading_fill_in_the_blank_answer_service.get_by_question", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get answers")
		return nil, err
	}
	return answers, nil
}
