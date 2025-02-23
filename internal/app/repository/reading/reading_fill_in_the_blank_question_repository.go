package reading

import (
	"context"
	"errors"
	"fluencybe/internal/app/model/reading"
	"fluencybe/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReadingFillInTheBlankQuestionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewReadingFillInTheBlankQuestionRepository(db *gorm.DB, logger *logger.PrettyLogger) *ReadingFillInTheBlankQuestionRepository {
	return &ReadingFillInTheBlankQuestionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ReadingFillInTheBlankQuestionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ReadingFillInTheBlankQuestionRepository) Create(ctx context.Context, question *reading.ReadingFillInTheBlankQuestion) error {
	now := time.Now()
	question.CreatedAt = now
	question.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(question).Error
	if err != nil {
		r.logger.Error("reading_fill_in_the_blank_question_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")
		return err
	}
	return nil
}

func (r *ReadingFillInTheBlankQuestionRepository) GetByID(ctx context.Context, id uuid.UUID) (*reading.ReadingFillInTheBlankQuestion, error) {
	var question reading.ReadingFillInTheBlankQuestion
	err := r.db.WithContext(ctx).First(&question, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("reading_fill_in_the_blank_question_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get question")
		return nil, err
	}
	return &question, nil
}

func (r *ReadingFillInTheBlankQuestionRepository) Update(ctx context.Context, question *reading.ReadingFillInTheBlankQuestion) error {
	question.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&reading.ReadingFillInTheBlankQuestion{}).
		Where("id = ?", question.ID).
		Updates(map[string]interface{}{
			"question":   question.Question,
			"updated_at": question.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("reading_fill_in_the_blank_question_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingFillInTheBlankQuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&reading.ReadingFillInTheBlankQuestion{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("reading_fill_in_the_blank_question_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingFillInTheBlankQuestionRepository) GetByReadingQuestionID(ctx context.Context, readingQuestionID uuid.UUID) ([]*reading.ReadingFillInTheBlankQuestion, error) {
	var questions []*reading.ReadingFillInTheBlankQuestion
	err := r.db.WithContext(ctx).
		Where("reading_question_id = ?", readingQuestionID).
		Find(&questions).Error

	if err != nil {
		r.logger.Error("reading_fill_in_the_blank_question_repository.get_by_reading_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    readingQuestionID,
		}, "Failed to get questions by reading question ID")
		return nil, err
	}

	return questions, nil
}
