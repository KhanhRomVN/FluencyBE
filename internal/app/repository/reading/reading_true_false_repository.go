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

type ReadingTrueFalseRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewReadingTrueFalseRepository(db *gorm.DB, logger *logger.PrettyLogger) *ReadingTrueFalseRepository {
	return &ReadingTrueFalseRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ReadingTrueFalseRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ReadingTrueFalseRepository) Create(ctx context.Context, trueFalse *reading.ReadingTrueFalse) error {
	now := time.Now()
	trueFalse.CreatedAt = now
	trueFalse.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(trueFalse).Error
	if err != nil {
		r.logger.Error("reading_true_false_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create true/false question")
		return err
	}
	return nil
}

func (r *ReadingTrueFalseRepository) GetByID(ctx context.Context, id uuid.UUID) (*reading.ReadingTrueFalse, error) {
	var trueFalse reading.ReadingTrueFalse
	err := r.db.WithContext(ctx).First(&trueFalse, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("reading_true_false_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get true/false question")
		return nil, err
	}
	return &trueFalse, nil
}

func (r *ReadingTrueFalseRepository) Update(ctx context.Context, trueFalse *reading.ReadingTrueFalse) error {
	trueFalse.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&reading.ReadingTrueFalse{}).
		Where("id = ?", trueFalse.ID).
		Updates(map[string]interface{}{
			"answer":     trueFalse.Answer,
			"explain":    trueFalse.Explain,
			"updated_at": trueFalse.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("reading_true_false_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update true/false question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingTrueFalseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&reading.ReadingTrueFalse{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("reading_true_false_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete true/false question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingTrueFalseRepository) GetByReadingQuestionID(ctx context.Context, readingQuestionID uuid.UUID) ([]*reading.ReadingTrueFalse, error) {
	var trueFalses []*reading.ReadingTrueFalse
	err := r.db.WithContext(ctx).
		Where("reading_question_id = ?", readingQuestionID).
		Find(&trueFalses).Error

	if err != nil {
		r.logger.Error("reading_true_false_repository.get_by_reading_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    readingQuestionID,
		}, "Failed to get true/false questions by reading question ID")
		return nil, err
	}

	return trueFalses, nil
}
