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

type ReadingChoiceOneOptionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewReadingChoiceOneOptionRepository(db *gorm.DB, logger *logger.PrettyLogger) *ReadingChoiceOneOptionRepository {
	return &ReadingChoiceOneOptionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ReadingChoiceOneOptionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ReadingChoiceOneOptionRepository) Create(ctx context.Context, option *reading.ReadingChoiceOneOption) error {
	now := time.Now()
	option.CreatedAt = now
	option.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(option).Error
	if err != nil {
		r.logger.Error("reading_choice_one_option_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create option")
		return err
	}
	return nil
}

func (r *ReadingChoiceOneOptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*reading.ReadingChoiceOneOption, error) {
	var option reading.ReadingChoiceOneOption
	err := r.db.WithContext(ctx).First(&option, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("reading_choice_one_option_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get option")
		return nil, err
	}
	return &option, nil
}

func (r *ReadingChoiceOneOptionRepository) Update(ctx context.Context, option *reading.ReadingChoiceOneOption) error {
	option.UpdatedAt = time.Now()

	// Get the existing option to determine which fields have changed
	var existingOption reading.ReadingChoiceOneOption
	if err := r.db.WithContext(ctx).First(&existingOption, "id = ?", option.ID).Error; err != nil {
		return err
	}

	// Create updates map with only changed fields
	updates := map[string]interface{}{
		"updated_at": option.UpdatedAt,
	}

	if option.Options != existingOption.Options {
		updates["options"] = option.Options
	}
	if option.IsCorrect != existingOption.IsCorrect {
		updates["is_correct"] = option.IsCorrect
	}

	result := r.db.WithContext(ctx).Model(&reading.ReadingChoiceOneOption{}).
		Where("id = ?", option.ID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("reading_choice_one_option_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update option")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ReadingChoiceOneOptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&reading.ReadingChoiceOneOption{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("reading_choice_one_option_repository.delete", map[string]interface{}{
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

func (r *ReadingChoiceOneOptionRepository) GetByReadingChoiceOneQuestionID(ctx context.Context, questionID uuid.UUID) ([]*reading.ReadingChoiceOneOption, error) {
	var options []*reading.ReadingChoiceOneOption
	err := r.db.WithContext(ctx).
		Where("reading_choice_one_question_id = ?", questionID).
		Find(&options).Error

	if err != nil {
		r.logger.Error("reading_choice_one_option_repository.get_by_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    questionID,
		}, "Failed to get options by question ID")
		return nil, err
	}

	return options, nil
}

func (r *ReadingChoiceOneOptionRepository) GetCorrectOption(ctx context.Context, questionID uuid.UUID) (*reading.ReadingChoiceOneOption, error) {
	var option reading.ReadingChoiceOneOption
	err := r.db.WithContext(ctx).
		Where("reading_choice_one_question_id = ? AND is_correct = true", questionID).
		First(&option).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("reading_choice_one_option_repository.get_correct_option", map[string]interface{}{
			"error":       err.Error(),
			"question_id": questionID,
		}, "Failed to get correct option")
		return nil, err
	}

	return &option, nil
}
