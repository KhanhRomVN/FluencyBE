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

type GrammarErrorIdentificationRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewGrammarErrorIdentificationRepository(db *gorm.DB, logger *logger.PrettyLogger) *GrammarErrorIdentificationRepository {
	return &GrammarErrorIdentificationRepository{
		db:     db,
		logger: logger,
	}
}

func (r *GrammarErrorIdentificationRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *GrammarErrorIdentificationRepository) Create(ctx context.Context, identification *grammar.GrammarErrorIdentification) error {
	now := time.Now()
	identification.CreatedAt = now
	identification.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(identification).Error
	if err != nil {
		r.logger.Error("grammar_error_identification_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create error identification")
		return err
	}
	return nil
}

func (r *GrammarErrorIdentificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*grammar.GrammarErrorIdentification, error) {
	var identification grammar.GrammarErrorIdentification
	err := r.db.WithContext(ctx).First(&identification, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("grammar_error_identification_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get error identification")
		return nil, err
	}
	return &identification, nil
}

func (r *GrammarErrorIdentificationRepository) Update(ctx context.Context, identification *grammar.GrammarErrorIdentification) error {
	identification.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&grammar.GrammarErrorIdentification{}).
		Where("id = ?", identification.ID).
		Updates(map[string]interface{}{
			"error_sentence": identification.ErrorSentence,
			"error_word":     identification.ErrorWord,
			"correct_word":   identification.CorrectWord,
			"explain":        identification.Explain,
			"updated_at":     identification.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("grammar_error_identification_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    identification.ID,
		}, "Failed to update error identification")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GrammarErrorIdentificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&grammar.GrammarErrorIdentification{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("grammar_error_identification_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete error identification")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GrammarErrorIdentificationRepository) GetByGrammarQuestionID(ctx context.Context, grammarQuestionID uuid.UUID) ([]*grammar.GrammarErrorIdentification, error) {
	var identifications []*grammar.GrammarErrorIdentification
	err := r.db.WithContext(ctx).
		Where("grammar_question_id = ?", grammarQuestionID).
		Find(&identifications).Error

	if err != nil {
		r.logger.Error("grammar_error_identification_repository.get_by_grammar_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    grammarQuestionID,
		}, "Failed to get error identifications")
		return nil, err
	}

	return identifications, nil
}
