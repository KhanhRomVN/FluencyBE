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

type GrammarChoiceOneOptionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewGrammarChoiceOneOptionRepository(db *gorm.DB, logger *logger.PrettyLogger) *GrammarChoiceOneOptionRepository {
	return &GrammarChoiceOneOptionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *GrammarChoiceOneOptionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *GrammarChoiceOneOptionRepository) Create(ctx context.Context, option *grammar.GrammarChoiceOneOption) error {
	now := time.Now()
	option.CreatedAt = now
	option.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(option).Error
	if err != nil {
		r.logger.Error("grammar_choice_one_option_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create option")
		return err
	}
	return nil
}

func (r *GrammarChoiceOneOptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*grammar.GrammarChoiceOneOption, error) {
	var option grammar.GrammarChoiceOneOption
	err := r.db.WithContext(ctx).First(&option, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("grammar_choice_one_option_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get option")
		return nil, err
	}
	return &option, nil
}

func (r *GrammarChoiceOneOptionRepository) Update(ctx context.Context, option *grammar.GrammarChoiceOneOption) error {
	option.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&grammar.GrammarChoiceOneOption{}).
		Where("id = ?", option.ID).
		Updates(map[string]interface{}{
			"options":    option.Options,
			"is_correct": option.IsCorrect,
			"updated_at": option.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("grammar_choice_one_option_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    option.ID,
		}, "Failed to update option")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GrammarChoiceOneOptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&grammar.GrammarChoiceOneOption{}, "id = ?", id)

	if result.Error != nil {
		r.logger.Error("grammar_choice_one_option_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete option")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GrammarChoiceOneOptionRepository) GetByQuestionID(ctx context.Context, questionID uuid.UUID) ([]*grammar.GrammarChoiceOneOption, error) {
	var options []*grammar.GrammarChoiceOneOption
	err := r.db.WithContext(ctx).
		Where("grammar_choice_one_question_id = ?", questionID).
		Find(&options).Error

	if err != nil {
		r.logger.Error("grammar_choice_one_option_repository.get_by_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": questionID,
		}, "Failed to get options by question ID")
		return nil, err
	}

	return options, nil
}

func (r *GrammarChoiceOneOptionRepository) GetCorrectOption(ctx context.Context, questionID uuid.UUID) (*grammar.GrammarChoiceOneOption, error) {
	var option grammar.GrammarChoiceOneOption
	err := r.db.WithContext(ctx).
		Where("grammar_choice_one_question_id = ? AND is_correct = true", questionID).
		First(&option).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("grammar_choice_one_option_repository.get_correct_option", map[string]interface{}{
			"error":       err.Error(),
			"question_id": questionID,
		}, "Failed to get correct option")
		return nil, err
	}

	return &option, nil
}
