package listening

import (
	"context"
	"errors"
	"fluencybe/internal/app/model/listening"
	"fluencybe/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListeningChoiceMultiOptionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewListeningChoiceMultiOptionRepository(db *gorm.DB, logger *logger.PrettyLogger) *ListeningChoiceMultiOptionRepository {
	return &ListeningChoiceMultiOptionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ListeningChoiceMultiOptionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ListeningChoiceMultiOptionRepository) Create(ctx context.Context, option *listening.ListeningChoiceMultiOption) error {
	now := time.Now()
	option.CreatedAt = now
	option.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(option).Error
	if err != nil {
		r.logger.Error("listening_choice_multi_option_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create option")
		return err
	}
	return nil
}

func (r *ListeningChoiceMultiOptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*listening.ListeningChoiceMultiOption, error) {
	var option listening.ListeningChoiceMultiOption
	err := r.db.WithContext(ctx).First(&option, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("listening_choice_multi_option_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get option")
		return nil, err
	}
	return &option, nil
}

func (r *ListeningChoiceMultiOptionRepository) GetByListeningChoiceMultiQuestionID(ctx context.Context, questionID uuid.UUID) ([]*listening.ListeningChoiceMultiOption, error) {
	var options []*listening.ListeningChoiceMultiOption
	err := r.db.WithContext(ctx).
		Where("listening_choice_multi_question_id = ?", questionID).
		Find(&options).Error

	if err != nil {
		r.logger.Error("listening_choice_multi_option_repository.get_by_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": questionID,
		}, "Failed to get options by question ID")
		return nil, err
	}

	return options, nil
}

func (r *ListeningChoiceMultiOptionRepository) Update(ctx context.Context, option *listening.ListeningChoiceMultiOption) error {
	option.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&listening.ListeningChoiceMultiOption{}).
		Where("id = ?", option.ID).
		Updates(map[string]interface{}{
			"options":    option.Options,
			"is_correct": option.IsCorrect,
			"updated_at": option.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("listening_choice_multi_option_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    option.ID,
		}, "Failed to update option")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ListeningChoiceMultiOptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&listening.ListeningChoiceMultiOption{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("listening_choice_multi_option_repository.delete", map[string]interface{}{
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
