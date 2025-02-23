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

type WritingSentenceCompletionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewWritingSentenceCompletionRepository(db *gorm.DB, logger *logger.PrettyLogger) *WritingSentenceCompletionRepository {
	return &WritingSentenceCompletionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *WritingSentenceCompletionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *WritingSentenceCompletionRepository) Create(ctx context.Context, sentence *writing.WritingSentenceCompletion) error {
	now := time.Now()
	sentence.CreatedAt = now
	sentence.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(sentence).Error
	if err != nil {
		r.logger.Error("writing_sentence_completion_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create sentence completion")
		return err
	}
	return nil
}

func (r *WritingSentenceCompletionRepository) GetByID(ctx context.Context, id uuid.UUID) (*writing.WritingSentenceCompletion, error) {
	var sentence writing.WritingSentenceCompletion
	err := r.db.WithContext(ctx).First(&sentence, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("writing_sentence_completion_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get sentence completion")
		return nil, err
	}
	return &sentence, nil
}

func (r *WritingSentenceCompletionRepository) GetByWritingQuestionID(ctx context.Context, writingQuestionID uuid.UUID) ([]*writing.WritingSentenceCompletion, error) {
	var sentences []*writing.WritingSentenceCompletion
	err := r.db.WithContext(ctx).
		Where("writing_question_id = ?", writingQuestionID).
		Find(&sentences).Error

	if err != nil {
		r.logger.Error("writing_sentence_completion_repository.get_by_writing_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": writingQuestionID,
		}, "Failed to get sentence completions")
		return nil, err
	}

	return sentences, nil
}

func (r *WritingSentenceCompletionRepository) Update(ctx context.Context, sentence *writing.WritingSentenceCompletion) error {
	sentence.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&writing.WritingSentenceCompletion{}).
		Where("id = ?", sentence.ID).
		Updates(map[string]interface{}{
			"example_sentence":    sentence.ExampleSentence,
			"given_part_sentence": sentence.GivenPartSentence,
			"position":            sentence.Position,
			"required_words":      sentence.RequiredWords,
			"explain":             sentence.Explain,
			"min_words":           sentence.MinWords,
			"max_words":           sentence.MaxWords,
			"updated_at":          sentence.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("writing_sentence_completion_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    sentence.ID,
		}, "Failed to update sentence completion")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *WritingSentenceCompletionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&writing.WritingSentenceCompletion{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("writing_sentence_completion_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete sentence completion")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
