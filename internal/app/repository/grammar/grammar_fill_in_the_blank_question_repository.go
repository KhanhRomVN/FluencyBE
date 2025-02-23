package grammar

import (
	"context"
	"errors"
	"fluencybe/internal/app/model/grammar"
	"fluencybe/pkg/logger"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GrammarFillInTheBlankQuestionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewGrammarFillInTheBlankQuestionRepository(db *gorm.DB, logger *logger.PrettyLogger) *GrammarFillInTheBlankQuestionRepository {
	return &GrammarFillInTheBlankQuestionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *GrammarFillInTheBlankQuestionRepository) Create(ctx context.Context, question *grammar.GrammarFillInTheBlankQuestion) error {
	now := time.Now()
	question.CreatedAt = now
	question.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(question).Error
	if err != nil {
		if strings.Contains(err.Error(), "unique_grammar_fill_in_the_blank_question") {
			return ErrDuplicateQuestion
		}
		r.logger.Error("grammar_fill_in_the_blank_question_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")
		return err
	}

	return nil
}

func (r *GrammarFillInTheBlankQuestionRepository) GetByID(ctx context.Context, id uuid.UUID) (*grammar.GrammarFillInTheBlankQuestion, error) {
	var question grammar.GrammarFillInTheBlankQuestion
	err := r.db.WithContext(ctx).First(&question, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("grammar_fill_in_the_blank_question_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get question")
		return nil, err
	}

	return &question, nil
}

func (r *GrammarFillInTheBlankQuestionRepository) Update(ctx context.Context, question *grammar.GrammarFillInTheBlankQuestion) error {
	question.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&grammar.GrammarFillInTheBlankQuestion{}).
		Where("id = ?", question.ID).
		Updates(map[string]interface{}{
			"question":   question.Question,
			"updated_at": question.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("grammar_fill_in_the_blank_question_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GrammarFillInTheBlankQuestionRepository) GetByGrammarQuestionID(ctx context.Context, grammarQuestionID uuid.UUID) ([]*grammar.GrammarFillInTheBlankQuestion, error) {
	var questions []*grammar.GrammarFillInTheBlankQuestion

	r.logger.Debug("GetByGrammarQuestionID", map[string]interface{}{
		"grammarQuestionID": grammarQuestionID,
	}, "Getting fill in blank questions")

	err := r.db.WithContext(ctx).
		Where("grammar_question_id = ?", grammarQuestionID).
		Order("created_at ASC").
		Find(&questions).Error

	if err != nil {
		r.logger.Error("GetByGrammarQuestionID", map[string]interface{}{
			"error":             err.Error(),
			"grammarQuestionID": grammarQuestionID,
		}, "Failed to get questions")
		return nil, err
	}

	r.logger.Debug("GetByGrammarQuestionID", map[string]interface{}{
		"count": len(questions),
	}, "Got questions")

	return questions, nil
}

func (r *GrammarFillInTheBlankQuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&grammar.GrammarFillInTheBlankQuestion{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("grammar_fill_in_the_blank_question_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to delete question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GrammarFillInTheBlankQuestionRepository) GetDB() *gorm.DB {
	return r.db
}
