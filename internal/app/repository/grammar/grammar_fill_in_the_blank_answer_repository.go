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

type GrammarFillInTheBlankAnswerRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewGrammarFillInTheBlankAnswerRepository(db *gorm.DB, logger *logger.PrettyLogger) *GrammarFillInTheBlankAnswerRepository {
	return &GrammarFillInTheBlankAnswerRepository{
		db:     db,
		logger: logger,
	}
}

func (r *GrammarFillInTheBlankAnswerRepository) Create(ctx context.Context, answer *grammar.GrammarFillInTheBlankAnswer) error {
	now := time.Now()
	answer.CreatedAt = now
	answer.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(answer).Error
	if err != nil {
		r.logger.Error("grammar_fill_in_the_blank_answer_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create answer")
		return err
	}

	return nil
}

func (r *GrammarFillInTheBlankAnswerRepository) GetByGrammarFillInTheBlankQuestionID(ctx context.Context, grammarFillInTheBlankQuestionID uuid.UUID) ([]*grammar.GrammarFillInTheBlankAnswer, error) {
	var answers []*grammar.GrammarFillInTheBlankAnswer

	err := r.db.WithContext(ctx).
		Where("grammar_fill_in_the_blank_question_id = ?", grammarFillInTheBlankQuestionID).
		Order("created_at ASC").
		Find(&answers).Error

	if err != nil {
		r.logger.Error("grammar_fill_in_the_blank_answer_repository.get_by_question_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get answers")
		return nil, err
	}

	return answers, nil
}

func (r *GrammarFillInTheBlankAnswerRepository) GetByID(ctx context.Context, id uuid.UUID) (*grammar.GrammarFillInTheBlankAnswer, error) {
	var answer grammar.GrammarFillInTheBlankAnswer

	err := r.db.WithContext(ctx).First(&answer, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("grammar_fill_in_the_blank_answer_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get fill in the blank answer")
		return nil, err
	}

	return &answer, nil
}

func (r *GrammarFillInTheBlankAnswerRepository) Update(ctx context.Context, answer *grammar.GrammarFillInTheBlankAnswer) error {
	answer.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&grammar.GrammarFillInTheBlankAnswer{}).
		Where("id = ?", answer.ID).
		Updates(map[string]interface{}{
			"answer":     answer.Answer,
			"explain":    answer.Explain,
			"updated_at": answer.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("grammar_fill_in_the_blank_answer_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update answer")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GrammarFillInTheBlankAnswerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&grammar.GrammarFillInTheBlankAnswer{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("grammar_fill_in_the_blank_answer_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete answer")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GrammarFillInTheBlankAnswerRepository) GetDB() *gorm.DB {
	return r.db
}
