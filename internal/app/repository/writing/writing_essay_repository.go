package writing

import (
	"context"
	"errors"
	"fluencybe/internal/app/model/writing"
	"fluencybe/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WritingEssayRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewWritingEssayRepository(db *gorm.DB, logger *logger.PrettyLogger) *WritingEssayRepository {
	return &WritingEssayRepository{
		db:     db,
		logger: logger,
	}
}

func (r *WritingEssayRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *WritingEssayRepository) Create(ctx context.Context, essay *writing.WritingEssay) error {
	now := time.Now()
	essay.CreatedAt = now
	essay.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(essay).Error
	if err != nil {
		r.logger.Error("writing_essay_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create essay")
		return err
	}
	return nil
}

func (r *WritingEssayRepository) GetByID(ctx context.Context, id uuid.UUID) (*writing.WritingEssay, error) {
	var essay writing.WritingEssay
	err := r.db.WithContext(ctx).First(&essay, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("writing_essay_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get essay")
		return nil, err
	}
	return &essay, nil
}

func (r *WritingEssayRepository) GetByWritingQuestionID(ctx context.Context, writingQuestionID uuid.UUID) ([]*writing.WritingEssay, error) {
	var essays []*writing.WritingEssay
	err := r.db.WithContext(ctx).
		Where("writing_question_id = ?", writingQuestionID).
		Find(&essays).Error

	if err != nil {
		r.logger.Error("writing_essay_repository.get_by_writing_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": writingQuestionID,
		}, "Failed to get essays")
		return nil, err
	}

	return essays, nil
}

func (r *WritingEssayRepository) Update(ctx context.Context, essay *writing.WritingEssay) error {
	essay.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&writing.WritingEssay{}).
		Where("id = ?", essay.ID).
		Updates(map[string]interface{}{
			"essay_type":      essay.EssayType,
			"required_points": essay.RequiredPoints,
			"min_words":       essay.MinWords,
			"max_words":       essay.MaxWords,
			"sample_essay":    essay.SampleEssay,
			"explain":         essay.Explain,
			"updated_at":      essay.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("writing_essay_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    essay.ID,
		}, "Failed to update essay")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *WritingEssayRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&writing.WritingEssay{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("writing_essay_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete essay")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
