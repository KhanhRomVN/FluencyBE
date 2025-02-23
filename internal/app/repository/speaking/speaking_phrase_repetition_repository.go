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

type SpeakingPhraseRepetitionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewSpeakingPhraseRepetitionRepository(db *gorm.DB, logger *logger.PrettyLogger) *SpeakingPhraseRepetitionRepository {
	return &SpeakingPhraseRepetitionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *SpeakingPhraseRepetitionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *SpeakingPhraseRepetitionRepository) Create(ctx context.Context, phrase *speaking.SpeakingPhraseRepetition) error {
	now := time.Now()
	phrase.CreatedAt = now
	phrase.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(phrase).Error
	if err != nil {
		r.logger.Error("speaking_phrase_repetition_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create phrase repetition")
		return err
	}
	return nil
}

func (r *SpeakingPhraseRepetitionRepository) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingPhraseRepetition, error) {
	var phrase speaking.SpeakingPhraseRepetition
	err := r.db.WithContext(ctx).First(&phrase, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("speaking_phrase_repetition_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get phrase repetition")
		return nil, err
	}
	return &phrase, nil
}

func (r *SpeakingPhraseRepetitionRepository) Update(ctx context.Context, phrase *speaking.SpeakingPhraseRepetition) error {
	phrase.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&speaking.SpeakingPhraseRepetition{}).
		Where("id = ?", phrase.ID).
		Updates(map[string]interface{}{
			"phrase":     phrase.Phrase,
			"mean":       phrase.Mean,
			"updated_at": phrase.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("speaking_phrase_repetition_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update phrase repetition")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SpeakingPhraseRepetitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&speaking.SpeakingPhraseRepetition{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("speaking_phrase_repetition_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete phrase repetition")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SpeakingPhraseRepetitionRepository) GetBySpeakingQuestionID(ctx context.Context, speakingQuestionID uuid.UUID) ([]*speaking.SpeakingPhraseRepetition, error) {
	var phrases []*speaking.SpeakingPhraseRepetition
	err := r.db.WithContext(ctx).
		Where("speaking_question_id = ?", speakingQuestionID).
		Find(&phrases).Error

	if err != nil {
		r.logger.Error("speaking_phrase_repetition_repository.get_by_speaking_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": speakingQuestionID,
		}, "Failed to get phrase repetitions")
		return nil, err
	}

	return phrases, nil
}
