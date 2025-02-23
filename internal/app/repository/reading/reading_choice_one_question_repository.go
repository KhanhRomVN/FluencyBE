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

type ReadingChoiceOneQuestionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewReadingChoiceOneQuestionRepository(db *gorm.DB, logger *logger.PrettyLogger) *ReadingChoiceOneQuestionRepository {
	return &ReadingChoiceOneQuestionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ReadingChoiceOneQuestionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ReadingChoiceOneQuestionRepository) Create(ctx context.Context, question *reading.ReadingChoiceOneQuestion) error {
	now := time.Now()
	question.CreatedAt = now
	question.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(question).Error
	if err != nil {
		r.logger.Error("reading_choice_one_question_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")
		return err
	}
	return nil
}

func (r *ReadingChoiceOneQuestionRepository) GetByID(ctx context.Context, id uuid.UUID) (*reading.ReadingChoiceOneQuestion, error) {
	var question reading.ReadingChoiceOneQuestion
	err := r.db.WithContext(ctx).First(&question, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("reading_choice_one_question_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get question")
		return nil, err
	}
	return &question, nil
}

func (r *ReadingChoiceOneQuestionRepository) Update(ctx context.Context, question *reading.ReadingChoiceOneQuestion) error {
	question.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&reading.ReadingChoiceOneQuestion{}).
		Where("id = ?", question.ID).
		Updates(map[string]interface{}{
			"question":   question.Question,
			"explain":    question.Explain,
			"updated_at": question.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("reading_choice_one_question_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingChoiceOneQuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&reading.ReadingChoiceOneQuestion{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("reading_choice_one_question_repository.delete", map[string]interface{}{
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

func (r *ReadingChoiceOneQuestionRepository) GetByReadingQuestionID(ctx context.Context, readingQuestionID uuid.UUID) ([]*reading.ReadingChoiceOneQuestion, error) {
	var questions []*reading.ReadingChoiceOneQuestion
	err := r.db.WithContext(ctx).
		Where("reading_question_id = ?", readingQuestionID).
		Find(&questions).Error

	if err != nil {
		r.logger.Error("reading_choice_one_question_repository.get_by_reading_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    readingQuestionID,
		}, "Failed to get questions by reading question ID")
		return nil, err
	}

	return questions, nil
}
