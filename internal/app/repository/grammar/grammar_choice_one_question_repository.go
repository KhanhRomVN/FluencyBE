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

type GrammarChoiceOneQuestionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewGrammarChoiceOneQuestionRepository(db *gorm.DB, logger *logger.PrettyLogger) *GrammarChoiceOneQuestionRepository {
	return &GrammarChoiceOneQuestionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *GrammarChoiceOneQuestionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *GrammarChoiceOneQuestionRepository) Create(ctx context.Context, question *grammar.GrammarChoiceOneQuestion) error {
	now := time.Now()
	question.CreatedAt = now
	question.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(question).Error
	if err != nil {
		r.logger.Error("grammar_choice_one_question_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")
		return err
	}
	return nil
}

func (r *GrammarChoiceOneQuestionRepository) GetByID(ctx context.Context, id uuid.UUID) (*grammar.GrammarChoiceOneQuestion, error) {
	var question grammar.GrammarChoiceOneQuestion
	err := r.db.WithContext(ctx).First(&question, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("grammar_choice_one_question_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get question")
		return nil, err
	}
	return &question, nil
}

func (r *GrammarChoiceOneQuestionRepository) Update(ctx context.Context, question *grammar.GrammarChoiceOneQuestion) error {
	question.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&grammar.GrammarChoiceOneQuestion{}).
		Where("id = ?", question.ID).
		Updates(map[string]interface{}{
			"question":   question.Question,
			"explain":    question.Explain,
			"updated_at": question.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("grammar_choice_one_question_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GrammarChoiceOneQuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&grammar.GrammarChoiceOneQuestion{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("grammar_choice_one_question_repository.delete", map[string]interface{}{
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

func (r *GrammarChoiceOneQuestionRepository) GetByGrammarQuestionID(ctx context.Context, grammarQuestionID uuid.UUID) ([]*grammar.GrammarChoiceOneQuestion, error) {
	var questions []*grammar.GrammarChoiceOneQuestion
	err := r.db.WithContext(ctx).
		Where("grammar_question_id = ?", grammarQuestionID).
		Order("created_at ASC").
		Find(&questions).Error

	if err != nil {
		r.logger.Error("grammar_choice_one_question_repository.get_by_grammar_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    grammarQuestionID,
		}, "Failed to get questions")
		return nil, err
	}

	return questions, nil
}
