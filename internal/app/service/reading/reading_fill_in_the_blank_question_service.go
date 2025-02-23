package reading

import (
	"context"
	"errors"
	readingHelper "fluencybe/internal/app/helper/reading"
	"fluencybe/internal/app/model/reading"
	searchClient "fluencybe/internal/app/opensearch"
	ReadingRepository "fluencybe/internal/app/repository/reading"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrDuplicateQuestion = errors.New("duplicate fill in the blank question")
)

type ReadingFillInTheBlankQuestionService struct {
	repo                *ReadingRepository.ReadingFillInTheBlankQuestionRepository
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository
	logger              *logger.PrettyLogger
	cache               cache.Cache
	search              *searchClient.ReadingQuestionSearch
	questionUpdator     *readingHelper.ReadingQuestionUpdator
}

func NewReadingFillInTheBlankQuestionService(
	repo *ReadingRepository.ReadingFillInTheBlankQuestionRepository,
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	search *searchClient.ReadingQuestionSearch,
	questionUpdator *readingHelper.ReadingQuestionUpdator,
) *ReadingFillInTheBlankQuestionService {
	return &ReadingFillInTheBlankQuestionService{
		repo:                repo,
		readingQuestionRepo: readingQuestionRepo,
		logger:              logger,
		cache:               cache,
		search:              search,
		questionUpdator:     questionUpdator,
	}
}

func (s *ReadingFillInTheBlankQuestionService) SetQuestionUpdator(updator *readingHelper.ReadingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ReadingFillInTheBlankQuestionService) validateQuestion(question *reading.ReadingFillInTheBlankQuestion) error {
	if question == nil {
		return ErrInvalidInput
	}
	if question.ReadingQuestionID == uuid.Nil {
		return errors.New("reading question ID is required")
	}
	if question.Question == "" {
		return errors.New("question text is required")
	}
	return nil
}

func (s *ReadingFillInTheBlankQuestionService) CreateQuestion(ctx context.Context, question *reading.ReadingFillInTheBlankQuestion) error {
	if err := s.validateQuestion(question); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Start a transaction
	tx := s.repo.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create question in database
	if err := s.repo.Create(ctx, question); err != nil {
		tx.Rollback()
		if errors.Is(err, ErrDuplicateQuestion) {
			return ErrDuplicateQuestion
		}
		s.logger.Error("reading_fill_in_the_blank_question_service.create", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to create question")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, question.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_fill_in_the_blank_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ReadingQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_fill_in_the_blank_question_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ReadingFillInTheBlankQuestionService) GetQuestion(ctx context.Context, id uuid.UUID) (*reading.ReadingFillInTheBlankQuestion, error) {
	question, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("reading_fill_in_the_blank_question_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get question")
		return nil, err
	}
	return question, nil
}

func (s *ReadingFillInTheBlankQuestionService) UpdateQuestion(ctx context.Context, question *reading.ReadingFillInTheBlankQuestion) error {
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

	// Update in database
	if err := s.repo.Update(ctx, question); err != nil {
		tx.Rollback()
		s.logger.Error("reading_fill_in_the_blank_question_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update question")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, question.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_fill_in_the_blank_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ReadingQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_fill_in_the_blank_question_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ReadingFillInTheBlankQuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
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

	// Delete from database
	if err := s.repo.Delete(ctx, id); err != nil {
		tx.Rollback()
		s.logger.Error("reading_fill_in_the_blank_question_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, question.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_fill_in_the_blank_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ReadingQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_fill_in_the_blank_question_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ReadingFillInTheBlankQuestionService) GetQuestionsByReadingQuestionID(ctx context.Context, readingQuestionID uuid.UUID) ([]*reading.ReadingFillInTheBlankQuestion, error) {
	questions, err := s.repo.GetByReadingQuestionID(ctx, readingQuestionID)
	if err != nil {
		s.logger.Error("reading_fill_in_the_blank_question_service.get_by_reading_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    readingQuestionID,
		}, "Failed to get questions")
		return nil, err
	}
	return questions, nil
}

func (s *ReadingFillInTheBlankQuestionService) GetDB() *gorm.DB {
	return s.repo.GetDB()
}
