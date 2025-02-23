package listening

import (
	"context"
	"fluencybe/internal/app/model/listening"
	"fluencybe/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListeningMapLabellingRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewListeningMapLabellingRepository(db *gorm.DB, logger *logger.PrettyLogger) *ListeningMapLabellingRepository {
	return &ListeningMapLabellingRepository{db: db, logger: logger}
}

func (r *ListeningMapLabellingRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ListeningMapLabellingRepository) Create(ctx context.Context, qa *listening.ListeningMapLabelling) error {
	now := time.Now()
	qa.CreatedAt = now
	qa.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(qa).Error; err != nil {
		r.logger.Error("listening_map_labelling_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create map labelling")
		return err
	}
	return nil
}

func (r *ListeningMapLabellingRepository) GetByID(ctx context.Context, id uuid.UUID) (*listening.ListeningMapLabelling, error) {
	var qa listening.ListeningMapLabelling
	if err := r.db.WithContext(ctx).First(&qa, "id = ?", id).Error; err != nil {
		r.logger.Error("listening_map_labelling_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get map labelling")
		return nil, err
	}
	return &qa, nil
}

func (r *ListeningMapLabellingRepository) GetByListeningQuestionID(ctx context.Context, listeningQuestionID uuid.UUID) ([]*listening.ListeningMapLabelling, error) {
	r.logger.Debug("GetByListeningQuestionID.start", map[string]interface{}{
		"listeningQuestionID": listeningQuestionID,
	}, "Starting to get map labellings")

	var labellings []*listening.ListeningMapLabelling
	result := r.db.WithContext(ctx).
		Where("listening_question_id = ?", listeningQuestionID).
		Find(&labellings)

	if result.Error != nil {
		r.logger.Error("listening_map_labelling_repository.get_by_listening_question_id", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    listeningQuestionID,
		}, "Failed to get map labellings")
		return nil, result.Error
	}

	return labellings, nil
}

func (r *ListeningMapLabellingRepository) Update(ctx context.Context, qa *listening.ListeningMapLabelling) error {
	qa.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&listening.ListeningMapLabelling{}).
		Where("id = ?", qa.ID).
		Updates(map[string]interface{}{
			"question":   qa.Question,
			"answer":     qa.Answer,
			"explain":    qa.Explain,
			"updated_at": qa.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("listening_map_labelling_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    qa.ID,
		}, "Failed to update map labelling")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ListeningMapLabellingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&listening.ListeningMapLabelling{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("listening_map_labelling_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete map labelling")
		return result.Error
	}
	return nil
}
