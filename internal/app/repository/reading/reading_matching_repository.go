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

type ReadingMatchingRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewReadingMatchingRepository(db *gorm.DB, logger *logger.PrettyLogger) *ReadingMatchingRepository {
	return &ReadingMatchingRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ReadingMatchingRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ReadingMatchingRepository) Create(ctx context.Context, matching *reading.ReadingMatching) error {
	now := time.Now()
	matching.CreatedAt = now
	matching.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(matching).Error
	if err != nil {
		r.logger.Error("reading_matching_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create matching")
		return err
	}
	return nil
}

func (r *ReadingMatchingRepository) GetByID(ctx context.Context, id uuid.UUID) (*reading.ReadingMatching, error) {
	var matching reading.ReadingMatching
	err := r.db.WithContext(ctx).First(&matching, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("reading_matching_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get matching")
		return nil, err
	}
	return &matching, nil
}

func (r *ReadingMatchingRepository) Update(ctx context.Context, matching *reading.ReadingMatching) error {
	matching.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&reading.ReadingMatching{}).
		Where("id = ?", matching.ID).
		Updates(map[string]interface{}{
			"question":   matching.Question,
			"answer":     matching.Answer,
			"updated_at": matching.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("reading_matching_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update matching")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingMatchingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&reading.ReadingMatching{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("reading_matching_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete matching")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingMatchingRepository) GetByReadingQuestionID(ctx context.Context, readingQuestionID uuid.UUID) ([]*reading.ReadingMatching, error) {
	var matchings []*reading.ReadingMatching
	err := r.db.WithContext(ctx).
		Where("reading_question_id = ?", readingQuestionID).
		Find(&matchings).Error

	if err != nil {
		r.logger.Error("reading_matching_repository.get_by_reading_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    readingQuestionID,
		}, "Failed to get matchings by reading question ID")
		return nil, err
	}

	return matchings, nil
}
