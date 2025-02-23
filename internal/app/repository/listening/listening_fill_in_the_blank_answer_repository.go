package listening

import (
	"context"
	"errors"
	"fluencybe/internal/app/model/listening"
	"fluencybe/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListeningFillInTheBlankAnswerRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewListeningFillInTheBlankAnswerRepository(db *gorm.DB, logger *logger.PrettyLogger) *ListeningFillInTheBlankAnswerRepository {
	return &ListeningFillInTheBlankAnswerRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ListeningFillInTheBlankAnswerRepository) Create(ctx context.Context, answer *listening.ListeningFillInTheBlankAnswer) error {
	now := time.Now()
	answer.CreatedAt = now
	answer.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(answer).Error
	if err != nil {
		r.logger.Error("listening_fill_in_the_blank_answer_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create answer")
		return err
	}

	return nil
}

func (r *ListeningFillInTheBlankAnswerRepository) GetByListeningFillInTheBlankQuestionID(ctx context.Context, listeningFillInTheBlankQuestionID uuid.UUID) ([]*listening.ListeningFillInTheBlankAnswer, error) {
	var answers []*listening.ListeningFillInTheBlankAnswer

	err := r.db.WithContext(ctx).
		Where("listening_fill_in_the_blank_question_id = ?", listeningFillInTheBlankQuestionID).
		Order("created_at ASC").
		Find(&answers).Error

	if err != nil {
		r.logger.Error("listening_fill_in_the_blank_answer_repository.get_by_question_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get answers")
		return nil, err
	}

	return answers, nil
}

func (r *ListeningFillInTheBlankAnswerRepository) GetByID(ctx context.Context, id uuid.UUID) (*listening.ListeningFillInTheBlankAnswer, error) {
	var answer listening.ListeningFillInTheBlankAnswer

	err := r.db.WithContext(ctx).First(&answer, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("listening_fill_in_the_blank_answer_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get fill in the blank answer")
		return nil, err
	}

	return &answer, nil
}

func (r *ListeningFillInTheBlankAnswerRepository) Update(ctx context.Context, answer *listening.ListeningFillInTheBlankAnswer) error {
	answer.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&listening.ListeningFillInTheBlankAnswer{}).
		Where("id = ?", answer.ID).
		Updates(map[string]interface{}{
			"answer":     answer.Answer,
			"explain":    answer.Explain,
			"updated_at": answer.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("listening_fill_in_the_blank_answer_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update answer")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ListeningFillInTheBlankAnswerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&listening.ListeningFillInTheBlankAnswer{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("listening_fill_in_the_blank_answer_repository.delete", map[string]interface{}{
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

func (r *ListeningFillInTheBlankAnswerRepository) GetDB() *gorm.DB {
	return r.db
}
