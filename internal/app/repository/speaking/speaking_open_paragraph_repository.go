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

type SpeakingOpenParagraphRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewSpeakingOpenParagraphRepository(db *gorm.DB, logger *logger.PrettyLogger) *SpeakingOpenParagraphRepository {
	return &SpeakingOpenParagraphRepository{
		db:     db,
		logger: logger,
	}
}

func (r *SpeakingOpenParagraphRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *SpeakingOpenParagraphRepository) Create(ctx context.Context, paragraph *speaking.SpeakingOpenParagraph) error {
	now := time.Now()
	paragraph.CreatedAt = now
	paragraph.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(paragraph).Error
	if err != nil {
		r.logger.Error("speaking_open_paragraph_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create open paragraph")
		return err
	}
	return nil
}

func (r *SpeakingOpenParagraphRepository) GetByID(ctx context.Context, id uuid.UUID) (*speaking.SpeakingOpenParagraph, error) {
	var paragraph speaking.SpeakingOpenParagraph
	err := r.db.WithContext(ctx).First(&paragraph, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("speaking_open_paragraph_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get open paragraph")
		return nil, err
	}
	return &paragraph, nil
}

func (r *SpeakingOpenParagraphRepository) Update(ctx context.Context, paragraph *speaking.SpeakingOpenParagraph) error {
	paragraph.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&speaking.SpeakingOpenParagraph{}).
		Where("id = ?", paragraph.ID).
		Updates(map[string]interface{}{
			"question":                paragraph.Question,
			"example_passage":         paragraph.ExamplePassage,
			"mean_of_example_passage": paragraph.MeanOfExamplePassage,
			"updated_at":              paragraph.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("speaking_open_paragraph_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update open paragraph")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SpeakingOpenParagraphRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&speaking.SpeakingOpenParagraph{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("speaking_open_paragraph_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete open paragraph")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SpeakingOpenParagraphRepository) GetBySpeakingQuestionID(ctx context.Context, speakingQuestionID uuid.UUID) ([]*speaking.SpeakingOpenParagraph, error) {
	var paragraphs []*speaking.SpeakingOpenParagraph
	err := r.db.WithContext(ctx).
		Where("speaking_question_id = ?", speakingQuestionID).
		Find(&paragraphs).Error

	if err != nil {
		r.logger.Error("speaking_open_paragraph_repository.get_by_speaking_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": speakingQuestionID,
		}, "Failed to get open paragraphs")
		return nil, err
	}

	return paragraphs, nil
}
