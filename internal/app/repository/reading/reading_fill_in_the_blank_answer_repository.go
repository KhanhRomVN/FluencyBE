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

type ReadingFillInTheBlankAnswerRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewReadingFillInTheBlankAnswerRepository(db *gorm.DB, logger *logger.PrettyLogger) *ReadingFillInTheBlankAnswerRepository {
	return &ReadingFillInTheBlankAnswerRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ReadingFillInTheBlankAnswerRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ReadingFillInTheBlankAnswerRepository) Create(ctx context.Context, answer *reading.ReadingFillInTheBlankAnswer) error {
	now := time.Now()
	answer.CreatedAt = now
	answer.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(answer).Error
	if err != nil {
		r.logger.Error("reading_fill_in_the_blank_answer_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create answer")
		return err
	}
	return nil
}

func (r *ReadingFillInTheBlankAnswerRepository) GetByID(ctx context.Context, id uuid.UUID) (*reading.ReadingFillInTheBlankAnswer, error) {
	var answer reading.ReadingFillInTheBlankAnswer
	err := r.db.WithContext(ctx).First(&answer, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("reading_fill_in_the_blank_answer_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get answer")
		return nil, err
	}
	return &answer, nil
}

func (r *ReadingFillInTheBlankAnswerRepository) Update(ctx context.Context, answer *reading.ReadingFillInTheBlankAnswer) error {
	answer.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&reading.ReadingFillInTheBlankAnswer{}).
		Where("id = ?", answer.ID).
		Updates(map[string]interface{}{
			"answer":     answer.Answer,
			"explain":    answer.Explain,
			"updated_at": answer.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("reading_fill_in_the_blank_answer_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update answer")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingFillInTheBlankAnswerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&reading.ReadingFillInTheBlankAnswer{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("reading_fill_in_the_blank_answer_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete answer")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingFillInTheBlankAnswerRepository) GetByReadingFillInTheBlankQuestionID(ctx context.Context, questionID uuid.UUID) ([]*reading.ReadingFillInTheBlankAnswer, error) {
	var answers []*reading.ReadingFillInTheBlankAnswer
	err := r.db.WithContext(ctx).
		Where("reading_fill_in_the_blank_question_id = ?", questionID).
		Find(&answers).Error

	if err != nil {
		r.logger.Error("reading_fill_in_the_blank_answer_repository.get_by_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    questionID,
		}, "Failed to get answers by question ID")
		return nil, err
	}

	return answers, nil
}
