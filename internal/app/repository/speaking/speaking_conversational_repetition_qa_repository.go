package speaking

import (
	"context"
	"errors"
	"fluencybe/internal/app/model/speaking"
	"fluencybe/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SpeakingConversationalRepetitionQARepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewSpeakingConversationalRepetitionQARepository(db *gorm.DB, logger *logger.PrettyLogger) *SpeakingConversationalRepetitionQARepository {
	return &SpeakingConversationalRepetitionQARepository{
		db:     db,
		logger: logger,
	}
}

func (r *SpeakingConversationalRepetitionQARepository) GetDB() *gorm.DB {
	return r.db
}

func (r *SpeakingConversationalRepetitionQARepository) Create(ctx context.Context, qa *speaking.SpeakingConversationalRepetitionQA) error {
	now := time.Now()
	qa.CreatedAt = now
	qa.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(qa).Error
	if err != nil {
		r.logger.Error("speaking_conversational_repetition_qa_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create QA")
		return err
	}
	return nil
}

func (r *SpeakingConversationalRepetitionQARepository) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingConversationalRepetitionQA, error) {
	var qa speaking.SpeakingConversationalRepetitionQA
	err := r.db.WithContext(ctx).First(&qa, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("speaking_conversational_repetition_qa_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get QA")
		return nil, err
	}
	return &qa, nil
}

func (r *SpeakingConversationalRepetitionQARepository) GetBySpeakingConversationalRepetitionID(ctx context.Context, repetitionID uuid.UUID) ([]*speaking.SpeakingConversationalRepetitionQA, error) {
	var qas []*speaking.SpeakingConversationalRepetitionQA
	err := r.db.WithContext(ctx).
		Where("speaking_conversational_repetition_id = ?", repetitionID).
		Find(&qas).Error

	if err != nil {
		r.logger.Error("speaking_conversational_repetition_qa_repository.get_by_repetition_id", map[string]interface{}{
			"error":        err.Error(),
			"repetitionID": repetitionID,
		}, "Failed to get QAs")
		return nil, err
	}

	return qas, nil
}

func (r *SpeakingConversationalRepetitionQARepository) Update(ctx context.Context, qa *speaking.SpeakingConversationalRepetitionQA) error {
	qa.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&speaking.SpeakingConversationalRepetitionQA{}).
		Where("id = ?", qa.ID).
		Updates(map[string]interface{}{
			"question":         qa.Question,
			"answer":           qa.Answer,
			"mean_of_question": qa.MeanOfQuestion,
			"mean_of_answer":   qa.MeanOfAnswer,
			"explain":          qa.Explain,
			"updated_at":       qa.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("speaking_conversational_repetition_qa_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    qa.ID,
		}, "Failed to update QA")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SpeakingConversationalRepetitionQARepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&speaking.SpeakingConversationalRepetitionQA{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("speaking_conversational_repetition_qa_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete QA")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
