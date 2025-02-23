package reading

import (
	"context"
	"errors"
	readingHelper "fluencybe/internal/app/helper/reading"
	"fluencybe/internal/app/model/reading"
	ReadingRepository "fluencybe/internal/app/repository/reading"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReadingTrueFalseService struct {
	repo                *ReadingRepository.ReadingTrueFalseRepository
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository
	logger              *logger.PrettyLogger
	questionUpdator     *readingHelper.ReadingQuestionUpdator
}

func NewReadingTrueFalseService(
	repo *ReadingRepository.ReadingTrueFalseRepository,
	readingQuestionRepo *ReadingRepository.ReadingQuestionRepository,
	logger *logger.PrettyLogger,
	questionUpdator *readingHelper.ReadingQuestionUpdator,
) *ReadingTrueFalseService {
	return &ReadingTrueFalseService{
		repo:                repo,
		readingQuestionRepo: readingQuestionRepo,
		logger:              logger,
		questionUpdator:     questionUpdator,
	}
}

func (s *ReadingTrueFalseService) SetQuestionUpdator(updator *readingHelper.ReadingQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ReadingTrueFalseService) validateTrueFalse(tf *reading.ReadingTrueFalse) error {
	if tf == nil {
		return errors.New("invalid input")
	}
	if tf.ReadingQuestionID == uuid.Nil {
		return errors.New("reading question ID is required")
	}
	if tf.Question == "" {
		return errors.New("question text is required")
	}
	if tf.Answer == "" {
		return errors.New("answer is required")
	}
	if tf.Answer != "TRUE" && tf.Answer != "FALSE" && tf.Answer != "NOT GIVEN" {
		return errors.New("answer must be TRUE, FALSE, or NOT GIVEN")
	}
	if tf.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *ReadingTrueFalseService) Create(ctx context.Context, trueFalse *reading.ReadingTrueFalse) error {
	if err := s.validateTrueFalse(trueFalse); err != nil {
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

	if err := s.repo.Create(ctx, trueFalse); err != nil {
		tx.Rollback()
		s.logger.Error("reading_true_false_service.create", map[string]interface{}{
			"error": err.Error(),
			"id":    trueFalse.ID,
		}, "Failed to create true/false question")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, trueFalse.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_true_false_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    trueFalse.ReadingQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_true_false_service.update_cache", map[string]interface{}{
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

func (s *ReadingTrueFalseService) GetByID(ctx context.Context, id uuid.UUID) (*reading.ReadingTrueFalse, error) {
	trueFalse, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("reading_true_false_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get true/false question")
		return nil, err
	}
	return trueFalse, nil
}

func (s *ReadingTrueFalseService) Update(ctx context.Context, trueFalse *reading.ReadingTrueFalse) error {
	if err := s.validateTrueFalse(trueFalse); err != nil {
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

	if err := s.repo.Update(ctx, trueFalse); err != nil {
		tx.Rollback()
		s.logger.Error("reading_true_false_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    trueFalse.ID,
		}, "Failed to update true/false question")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, trueFalse.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_true_false_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    trueFalse.ReadingQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_true_false_service.update_cache", map[string]interface{}{
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

func (s *ReadingTrueFalseService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get true/false question before deletion to get parent ID
	trueFalse, err := s.GetByID(ctx, id)
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
		s.logger.Error("reading_true_false_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete true/false question")
		return err
	}

	// Get parent reading question for cache/search update
	parentQuestion, err := s.readingQuestionRepo.GetReadingQuestionByID(ctx, trueFalse.ReadingQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("reading_true_false_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    trueFalse.ReadingQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("reading_true_false_service.update_cache", map[string]interface{}{
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

func (s *ReadingTrueFalseService) GetByReadingQuestionID(ctx context.Context, readingQuestionID uuid.UUID) ([]*reading.ReadingTrueFalse, error) {
	trueFalses, err := s.repo.GetByReadingQuestionID(ctx, readingQuestionID)
	if err != nil {
		s.logger.Error("reading_true_false_service.get_by_reading_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    readingQuestionID,
		}, "Failed to get true/false questions")
		return nil, err
	}
	return trueFalses, nil
}

func (s *ReadingTrueFalseService) GetDB() *gorm.DB {
	return s.repo.GetDB()
}
