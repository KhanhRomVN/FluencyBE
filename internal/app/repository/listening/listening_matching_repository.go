package listening

import (
	"context"
	"fluencybe/internal/app/model/listening"
	"fluencybe/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListeningMatchingRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewListeningMatchingRepository(db *gorm.DB, logger *logger.PrettyLogger) *ListeningMatchingRepository {
	return &ListeningMatchingRepository{db: db, logger: logger}
}

func (r *ListeningMatchingRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ListeningMatchingRepository) Create(ctx context.Context, matching *listening.ListeningMatching) error {
	now := time.Now()
	matching.CreatedAt = now
	matching.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(matching).Error; err != nil {
		r.logger.Error("listening_matching_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create matching")
		return err
	}
	return nil
}

func (r *ListeningMatchingRepository) GetByID(ctx context.Context, id uuid.UUID) (*listening.ListeningMatching, error) {
	var matching listening.ListeningMatching
	if err := r.db.WithContext(ctx).First(&matching, "id = ?", id).Error; err != nil {
		r.logger.Error("listening_matching_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get matching")
		return nil, err
	}
	return &matching, nil
}

func (r *ListeningMatchingRepository) GetByListeningQuestionID(ctx context.Context, listeningQuestionID uuid.UUID) ([]*listening.ListeningMatching, error) {
	var matchings []*listening.ListeningMatching
	if err := r.db.WithContext(ctx).
		Where("listening_question_id = ?", listeningQuestionID).
		Find(&matchings).Error; err != nil {
		r.logger.Error("listening_matching_repository.get_by_listening_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    listeningQuestionID,
		}, "Failed to get matchings")
		return nil, err
	}
	return matchings, nil
}

func (r *ListeningMatchingRepository) Update(ctx context.Context, matching *listening.ListeningMatching) error {
	matching.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&listening.ListeningMatching{}).
		Where("id = ?", matching.ID).
		Updates(map[string]interface{}{
			"question":   matching.Question,
			"answer":     matching.Answer,
			"explain":    matching.Explain,
			"updated_at": matching.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("listening_matching_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    matching.ID,
		}, "Failed to update matching")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ListeningMatchingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&listening.ListeningMatching{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("listening_matching_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete matching")
		return result.Error
	}
	return nil
}
