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

type SpeakingConversationalRepetitionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewSpeakingConversationalRepetitionRepository(db *gorm.DB, logger *logger.PrettyLogger) *SpeakingConversationalRepetitionRepository {
	return &SpeakingConversationalRepetitionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *SpeakingConversationalRepetitionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *SpeakingConversationalRepetitionRepository) Create(ctx context.Context, repetition *speaking.SpeakingConversationalRepetition) error {
	now := time.Now()
	repetition.CreatedAt = now
	repetition.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(repetition).Error
	if err != nil {
		r.logger.Error("speaking_conversational_repetition_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create conversational repetition")
		return err
	}
	return nil
}

func (r *SpeakingConversationalRepetitionRepository) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingConversationalRepetition, error) {
	var repetition speaking.SpeakingConversationalRepetition
	err := r.db.WithContext(ctx).First(&repetition, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("speaking_conversational_repetition_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get conversational repetition")
		return nil, err
	}
	return &repetition, nil
}

func (r *SpeakingConversationalRepetitionRepository) GetBySpeakingQuestionID(ctx context.Context, speakingQuestionID uuid.UUID) ([]*speaking.SpeakingConversationalRepetition, error) {
	var repetitions []*speaking.SpeakingConversationalRepetition
	err := r.db.WithContext(ctx).
		Where("speaking_question_id = ?", speakingQuestionID).
		Find(&repetitions).Error

	if err != nil {
		r.logger.Error("speaking_conversational_repetition_repository.get_by_speaking_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": speakingQuestionID,
		}, "Failed to get conversational repetitions")
		return nil, err
	}

	return repetitions, nil
}

func (r *SpeakingConversationalRepetitionRepository) Update(ctx context.Context, repetition *speaking.SpeakingConversationalRepetition) error {
	repetition.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&speaking.SpeakingConversationalRepetition{}).
		Where("id = ?", repetition.ID).
		Updates(map[string]interface{}{
			"title":      repetition.Title,
			"overview":   repetition.Overview,
			"updated_at": repetition.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("speaking_conversational_repetition_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    repetition.ID,
		}, "Failed to update conversational repetition")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SpeakingConversationalRepetitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&speaking.SpeakingConversationalRepetition{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("speaking_conversational_repetition_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete conversational repetition")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
