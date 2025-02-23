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

type ReadingChoiceMultiOptionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewReadingChoiceMultiOptionRepository(db *gorm.DB, logger *logger.PrettyLogger) *ReadingChoiceMultiOptionRepository {
	return &ReadingChoiceMultiOptionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ReadingChoiceMultiOptionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ReadingChoiceMultiOptionRepository) Create(ctx context.Context, option *reading.ReadingChoiceMultiOption) error {
	now := time.Now()
	option.CreatedAt = now
	option.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(option).Error
	if err != nil {
		r.logger.Error("reading_choice_multi_option_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create option")
		return err
	}
	return nil
}

func (r *ReadingChoiceMultiOptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*reading.ReadingChoiceMultiOption, error) {
	var option reading.ReadingChoiceMultiOption
	err := r.db.WithContext(ctx).First(&option, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("reading_choice_multi_option_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get option")
		return nil, err
	}
	return &option, nil
}

func (r *ReadingChoiceMultiOptionRepository) Update(ctx context.Context, option *reading.ReadingChoiceMultiOption) error {
	option.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&reading.ReadingChoiceMultiOption{}).
		Where("id = ?", option.ID).
		Updates(map[string]interface{}{
			"is_correct": option.IsCorrect,
			"updated_at": option.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("reading_choice_multi_option_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update option")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingChoiceMultiOptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&reading.ReadingChoiceMultiOption{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("reading_choice_multi_option_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete option")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingChoiceMultiOptionRepository) GetByReadingChoiceMultiQuestionID(ctx context.Context, questionID uuid.UUID) ([]*reading.ReadingChoiceMultiOption, error) {
	var options []*reading.ReadingChoiceMultiOption
	err := r.db.WithContext(ctx).
		Where("reading_choice_multi_question_id = ?", questionID).
		Find(&options).Error

	if err != nil {
		r.logger.Error("reading_choice_multi_option_repository.get_by_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    questionID,
		}, "Failed to get options by question ID")
		return nil, err
	}

	return options, nil
}
