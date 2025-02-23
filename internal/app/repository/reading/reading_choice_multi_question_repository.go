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

type ReadingChoiceMultiQuestionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewReadingChoiceMultiQuestionRepository(db *gorm.DB, logger *logger.PrettyLogger) *ReadingChoiceMultiQuestionRepository {
	return &ReadingChoiceMultiQuestionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ReadingChoiceMultiQuestionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ReadingChoiceMultiQuestionRepository) Create(ctx context.Context, question *reading.ReadingChoiceMultiQuestion) error {
	now := time.Now()
	question.CreatedAt = now
	question.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(question).Error
	if err != nil {
		r.logger.Error("reading_choice_multi_question_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")
		return err
	}
	return nil
}

func (r *ReadingChoiceMultiQuestionRepository) GetByID(ctx context.Context, id uuid.UUID) (*reading.ReadingChoiceMultiQuestion, error) {
	var question reading.ReadingChoiceMultiQuestion
	err := r.db.WithContext(ctx).First(&question, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("reading_choice_multi_question_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get question")
		return nil, err
	}
	return &question, nil
}

func (r *ReadingChoiceMultiQuestionRepository) Update(ctx context.Context, question *reading.ReadingChoiceMultiQuestion) error {
	question.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&reading.ReadingChoiceMultiQuestion{}).
		Where("id = ?", question.ID).
		Updates(map[string]interface{}{
			"question":   question.Question,
			"explain":    question.Explain,
			"updated_at": question.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("reading_choice_multi_question_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingChoiceMultiQuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&reading.ReadingChoiceMultiQuestion{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("reading_choice_multi_question_repository.delete", map[string]interface{}{
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

func (r *ReadingChoiceMultiQuestionRepository) GetByReadingQuestionID(ctx context.Context, readingQuestionID uuid.UUID) ([]*reading.ReadingChoiceMultiQuestion, error) {
	var questions []*reading.ReadingChoiceMultiQuestion
	err := r.db.WithContext(ctx).
		Where("reading_question_id = ?", readingQuestionID).
		Find(&questions).Error

	if err != nil {
		r.logger.Error("reading_choice_multi_question_repository.get_by_reading_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    readingQuestionID,
		}, "Failed to get questions by reading question ID")
		return nil, err
	}

	return questions, nil
}
