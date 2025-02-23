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

type SpeakingConversationalOpenRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewSpeakingConversationalOpenRepository(db *gorm.DB, logger *logger.PrettyLogger) *SpeakingConversationalOpenRepository {
	return &SpeakingConversationalOpenRepository{
		db:     db,
		logger: logger,
	}
}

func (r *SpeakingConversationalOpenRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *SpeakingConversationalOpenRepository) Create(ctx context.Context, open *speaking.SpeakingConversationalOpen) error {
	now := time.Now()
	open.CreatedAt = now
	open.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(open).Error
	if err != nil {
		r.logger.Error("speaking_conversational_open_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create conversational open")
		return err
	}
	return nil
}

func (r *SpeakingConversationalOpenRepository) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingConversationalOpen, error) {
	var open speaking.SpeakingConversationalOpen
	err := r.db.WithContext(ctx).First(&open, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("speaking_conversational_open_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get conversational open")
		return nil, err
	}
	return &open, nil
}

func (r *SpeakingConversationalOpenRepository) GetBySpeakingQuestionID(ctx context.Context, speakingQuestionID uuid.UUID) ([]*speaking.SpeakingConversationalOpen, error) {
	var opens []*speaking.SpeakingConversationalOpen
	err := r.db.WithContext(ctx).
		Where("speaking_question_id = ?", speakingQuestionID).
		Find(&opens).Error

	if err != nil {
		r.logger.Error("speaking_conversational_open_repository.get_by_speaking_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": speakingQuestionID,
		}, "Failed to get conversational opens")
		return nil, err
	}

	return opens, nil
}

func (r *SpeakingConversationalOpenRepository) Update(ctx context.Context, open *speaking.SpeakingConversationalOpen) error {
	open.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&speaking.SpeakingConversationalOpen{}).
		Where("id = ?", open.ID).
		Updates(map[string]interface{}{
			"title":                open.Title,
			"overview":             open.Overview,
			"example_conversation": open.ExampleConversation,
			"updated_at":           open.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("speaking_conversational_open_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    open.ID,
		}, "Failed to update conversational open")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SpeakingConversationalOpenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&speaking.SpeakingConversationalOpen{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("speaking_conversational_open_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete conversational open")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
