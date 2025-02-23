package grammar

import (
	"context"
	"errors"
	"fluencybe/internal/app/model/grammar"
	"fluencybe/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GrammarSentenceTransformationRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewGrammarSentenceTransformationRepository(db *gorm.DB, logger *logger.PrettyLogger) *GrammarSentenceTransformationRepository {
	return &GrammarSentenceTransformationRepository{
		db:     db,
		logger: logger,
	}
}

func (r *GrammarSentenceTransformationRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *GrammarSentenceTransformationRepository) Create(ctx context.Context, transformation *grammar.GrammarSentenceTransformation) error {
	now := time.Now()
	transformation.CreatedAt = now
	transformation.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(transformation).Error
	if err != nil {
		r.logger.Error("grammar_sentence_transformation_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create sentence transformation")
		return err
	}
	return nil
}

func (r *GrammarSentenceTransformationRepository) GetByID(ctx context.Context, id uuid.UUID) (*grammar.GrammarSentenceTransformation, error) {
	var transformation grammar.GrammarSentenceTransformation
	err := r.db.WithContext(ctx).First(&transformation, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("grammar_sentence_transformation_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get sentence transformation")
		return nil, err
	}
	return &transformation, nil
}

func (r *GrammarSentenceTransformationRepository) Update(ctx context.Context, transformation *grammar.GrammarSentenceTransformation) error {
	transformation.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&grammar.GrammarSentenceTransformation{}).
		Where("id = ?", transformation.ID).
		Updates(map[string]interface{}{
			"original_sentence":        transformation.OriginalSentence,
			"beginning_word":           transformation.BeginningWord,
			"example_correct_sentence": transformation.ExampleCorrectSentence,
			"explain":                  transformation.Explain,
			"updated_at":               transformation.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("grammar_sentence_transformation_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    transformation.ID,
		}, "Failed to update sentence transformation")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GrammarSentenceTransformationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&grammar.GrammarSentenceTransformation{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("grammar_sentence_transformation_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete sentence transformation")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GrammarSentenceTransformationRepository) GetByGrammarQuestionID(ctx context.Context, grammarQuestionID uuid.UUID) ([]*grammar.GrammarSentenceTransformation, error) {
	var transformations []*grammar.GrammarSentenceTransformation
	err := r.db.WithContext(ctx).
		Where("grammar_question_id = ?", grammarQuestionID).
		Find(&transformations).Error

	if err != nil {
		r.logger.Error("grammar_sentence_transformation_repository.get_by_grammar_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    grammarQuestionID,
		}, "Failed to get sentence transformations")
		return nil, err
	}

	return transformations, nil
}
