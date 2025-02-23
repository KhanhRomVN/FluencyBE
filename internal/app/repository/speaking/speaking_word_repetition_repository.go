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

type SpeakingWordRepetitionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewSpeakingWordRepetitionRepository(db *gorm.DB, logger *logger.PrettyLogger) *SpeakingWordRepetitionRepository {
	return &SpeakingWordRepetitionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *SpeakingWordRepetitionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *SpeakingWordRepetitionRepository) Create(ctx context.Context, word *speaking.SpeakingWordRepetition) error {
	now := time.Now()
	word.CreatedAt = now
	word.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(word).Error
	if err != nil {
		r.logger.Error("speaking_word_repetition_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create word repetition")
		return err
	}
	return nil
}

func (r *SpeakingWordRepetitionRepository) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingWordRepetition, error) {
	var word speaking.SpeakingWordRepetition
	err := r.db.WithContext(ctx).First(&word, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("speaking_word_repetition_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get word repetition")
		return nil, err
	}
	return &word, nil
}

func (r *SpeakingWordRepetitionRepository) Update(ctx context.Context, word *speaking.SpeakingWordRepetition) error {
	word.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&speaking.SpeakingWordRepetition{}).
		Where("id = ?", word.ID).
		Updates(map[string]interface{}{
			"word":       word.Word,
			"mean":       word.Mean,
			"updated_at": word.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("speaking_word_repetition_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update word repetition")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SpeakingWordRepetitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&speaking.SpeakingWordRepetition{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("speaking_word_repetition_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete word repetition")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SpeakingWordRepetitionRepository) GetBySpeakingQuestionID(ctx context.Context, speakingQuestionID uuid.UUID) ([]*speaking.SpeakingWordRepetition, error) {
	var words []*speaking.SpeakingWordRepetition
	err := r.db.WithContext(ctx).
		Where("speaking_question_id = ?", speakingQuestionID).
		Find(&words).Error

	if err != nil {
		r.logger.Error("speaking_word_repetition_repository.get_by_speaking_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": speakingQuestionID,
		}, "Failed to get word repetitions")
		return nil, err
	}

	return words, nil
}
