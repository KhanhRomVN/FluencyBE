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

type ListeningChoiceOneQuestionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewListeningChoiceOneQuestionRepository(db *gorm.DB, logger *logger.PrettyLogger) *ListeningChoiceOneQuestionRepository {
	return &ListeningChoiceOneQuestionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ListeningChoiceOneQuestionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ListeningChoiceOneQuestionRepository) Create(ctx context.Context, question *listening.ListeningChoiceOneQuestion) error {
	now := time.Now()
	question.CreatedAt = now
	question.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(question).Error
	if err != nil {
		r.logger.Error("listening_choice_one_question_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")
		return err
	}
	return nil
}

func (r *ListeningChoiceOneQuestionRepository) GetByID(ctx context.Context, id uuid.UUID) (*listening.ListeningChoiceOneQuestion, error) {
	var question listening.ListeningChoiceOneQuestion
	err := r.db.WithContext(ctx).First(&question, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("listening_choice_one_question_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get question")
		return nil, err
	}
	return &question, nil
}

func (r *ListeningChoiceOneQuestionRepository) Update(ctx context.Context, question *listening.ListeningChoiceOneQuestion) error {
	question.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&listening.ListeningChoiceOneQuestion{}).
		Where("id = ?", question.ID).
		Updates(map[string]interface{}{
			"question":   question.Question,
			"explain":    question.Explain,
			"updated_at": question.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("listening_choice_one_question_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ListeningChoiceOneQuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&listening.ListeningChoiceOneQuestion{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("listening_choice_one_question_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ListeningChoiceOneQuestionRepository) GetByListeningQuestionID(ctx context.Context, listeningQuestionID uuid.UUID) ([]*listening.ListeningChoiceOneQuestion, error) {
	var questions []*listening.ListeningChoiceOneQuestion
	err := r.db.WithContext(ctx).
		Where("listening_question_id = ?", listeningQuestionID).
		Order("created_at ASC").
		Find(&questions).Error

	if err != nil {
		r.logger.Error("listening_choice_one_question_repository.get_by_listening_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    listeningQuestionID,
		}, "Failed to get questions")
		return nil, err
	}

	return questions, nil
}
