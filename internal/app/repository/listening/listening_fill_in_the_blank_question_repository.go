package listening

import (
	"context"
	"errors"
	"fluencybe/internal/app/model/listening"
	"fluencybe/pkg/logger"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListeningFillInTheBlankQuestionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewListeningFillInTheBlankQuestionRepository(db *gorm.DB, logger *logger.PrettyLogger) *ListeningFillInTheBlankQuestionRepository {
	return &ListeningFillInTheBlankQuestionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ListeningFillInTheBlankQuestionRepository) Create(ctx context.Context, question *listening.ListeningFillInTheBlankQuestion) error {
	now := time.Now()
	question.CreatedAt = now
	question.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(question).Error
	if err != nil {
		if strings.Contains(err.Error(), "unique_listening_fill_in_the_blank_question") {
			return ErrDuplicateQuestion
		}
		r.logger.Error("listening_fill_in_the_blank_question_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")
		return err
	}

	return nil
}

func (r *ListeningFillInTheBlankQuestionRepository) GetByID(ctx context.Context, id uuid.UUID) (*listening.ListeningFillInTheBlankQuestion, error) {
	var question listening.ListeningFillInTheBlankQuestion
	err := r.db.WithContext(ctx).First(&question, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("listening_fill_in_the_blank_question_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get question")
		return nil, err
	}

	return &question, nil
}

func (r *ListeningFillInTheBlankQuestionRepository) Update(ctx context.Context, question *listening.ListeningFillInTheBlankQuestion) error {
	question.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&listening.ListeningFillInTheBlankQuestion{}).
		Where("id = ?", question.ID).
		Updates(map[string]interface{}{
			"question":   question.Question,
			"updated_at": question.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("listening_fill_in_the_blank_question_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ListeningFillInTheBlankQuestionRepository) GetByListeningQuestionID(ctx context.Context, listeningQuestionID uuid.UUID) ([]*listening.ListeningFillInTheBlankQuestion, error) {
	var questions []*listening.ListeningFillInTheBlankQuestion

	// Add debug logging
	r.logger.Debug("GetByListeningQuestionID", map[string]interface{}{
		"listeningQuestionID": listeningQuestionID,
	}, "Getting fill in blank questions")

	err := r.db.WithContext(ctx).
		Where("listening_question_id = ?", listeningQuestionID).
		Order("created_at ASC").
		Find(&questions).Error

	if err != nil {
		r.logger.Error("GetByListeningQuestionID", map[string]interface{}{
			"error":               err.Error(),
			"listeningQuestionID": listeningQuestionID,
		}, "Failed to get questions")
		return nil, err
	}

	r.logger.Debug("GetByListeningQuestionID", map[string]interface{}{
		"count": len(questions),
	}, "Got questions")

	return questions, nil
}

func (r *ListeningFillInTheBlankQuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&listening.ListeningFillInTheBlankQuestion{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("listening_fill_in_the_blank_question_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to delete question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ListeningFillInTheBlankQuestionRepository) GetDB() *gorm.DB {
	return r.db
}
