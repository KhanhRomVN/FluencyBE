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

type SpeakingParagraphRepetitionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewSpeakingParagraphRepetitionRepository(db *gorm.DB, logger *logger.PrettyLogger) *SpeakingParagraphRepetitionRepository {
	return &SpeakingParagraphRepetitionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *SpeakingParagraphRepetitionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *SpeakingParagraphRepetitionRepository) Create(ctx context.Context, paragraph *speaking.SpeakingParagraphRepetition) error {
	now := time.Now()
	paragraph.CreatedAt = now
	paragraph.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(paragraph).Error
	if err != nil {
		r.logger.Error("speaking_paragraph_repetition_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create paragraph repetition")
		return err
	}
	return nil
}

func (r *SpeakingParagraphRepetitionRepository) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingParagraphRepetition, error) {
	var paragraph speaking.SpeakingParagraphRepetition
	err := r.db.WithContext(ctx).First(&paragraph, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("speaking_paragraph_repetition_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get paragraph repetition")
		return nil, err
	}
	return &paragraph, nil
}

func (r *SpeakingParagraphRepetitionRepository) GetBySpeakingQuestionID(ctx context.Context, speakingQuestionID uuid.UUID) ([]*speaking.SpeakingParagraphRepetition, error) {
	var paragraphs []*speaking.SpeakingParagraphRepetition
	err := r.db.WithContext(ctx).
		Where("speaking_question_id = ?", speakingQuestionID).
		Find(&paragraphs).Error

	if err != nil {
		r.logger.Error("speaking_paragraph_repetition_repository.get_by_speaking_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": speakingQuestionID,
		}, "Failed to get paragraph repetitions")
		return nil, err
	}

	return paragraphs, nil
}

func (r *SpeakingParagraphRepetitionRepository) Update(ctx context.Context, paragraph *speaking.SpeakingParagraphRepetition) error {
	paragraph.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&speaking.SpeakingParagraphRepetition{}).
		Where("id = ?", paragraph.ID).
		Updates(map[string]interface{}{
			"paragraph":  paragraph.Paragraph,
			"mean":       paragraph.Mean,
			"updated_at": paragraph.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("speaking_paragraph_repetition_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    paragraph.ID,
		}, "Failed to update paragraph repetition")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SpeakingParagraphRepetitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&speaking.SpeakingParagraphRepetition{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("speaking_paragraph_repetition_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete paragraph repetition")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
