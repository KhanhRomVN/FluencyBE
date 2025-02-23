package writing

import (
	"context"
	"errors"
	"fluencybe/internal/app/model/writing"
	"fluencybe/pkg/logger"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrQuestionNotFound   = errors.New("writing question not found")
	ErrInvalidInput       = errors.New("invalid input data")
	ErrDuplicateQuestion  = errors.New("duplicate question")
	ErrTransactionFailed  = errors.New("transaction failed")
	ErrInvalidQueryParams = errors.New("invalid query parameters")
)

type WritingQuestionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewWritingQuestionRepository(db *gorm.DB, logger *logger.PrettyLogger) *WritingQuestionRepository {
	return &WritingQuestionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *WritingQuestionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *WritingQuestionRepository) CreateWritingQuestion(ctx context.Context, question *writing.WritingQuestion) error {
	now := time.Now().UTC()
	question.CreatedAt = now
	question.UpdatedAt = now
	question.Version = 1

	result := r.db.WithContext(ctx).Create(question)
	if result.Error != nil {
		r.logger.Error("writing_question_repository.create", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to create writing question")
		return result.Error
	}

	return nil
}

func (r *WritingQuestionRepository) GetWritingQuestionByID(ctx context.Context, id uuid.UUID) (*writing.WritingQuestion, error) {
	var question writing.WritingQuestion
	err := r.db.WithContext(ctx).First(&question, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuestionNotFound
		}
		r.logger.Error("writing_question_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get writing question")
		return nil, err
	}

	return &question, nil
}

func (r *WritingQuestionRepository) UpdateWritingQuestion(ctx context.Context, question *writing.WritingQuestion) error {
	question.UpdatedAt = time.Now().UTC()

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Model(&writing.WritingQuestion{}).
		Where("id = ?", question.ID).
		Updates(map[string]interface{}{
			"type":        question.Type,
			"topic":       question.Topic,
			"instruction": question.Instruction,
			"image_urls":  question.ImageURLs,
			"max_time":    question.MaxTime,
			"version":     question.Version,
			"updated_at":  question.UpdatedAt,
		})

	if result.Error != nil {
		tx.Rollback()
		r.logger.Error("writing_question_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to update writing question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return ErrQuestionNotFound
	}

	if err := tx.First(question, "id = ?", question.ID).Error; err != nil {
		tx.Rollback()
		r.logger.Error("writing_question_repository.update.get_updated", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get updated question")
		return err
	}

	return tx.Commit().Error
}

func (r *WritingQuestionRepository) DeleteWritingQuestion(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&writing.WritingQuestion{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("writing_question_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete writing question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrQuestionNotFound
	}

	return nil
}

func (r *WritingQuestionRepository) GetNewUpdatedQuestions(ctx context.Context, versionChecks []struct {
	ID      uuid.UUID
	Version int
}) ([]*writing.WritingQuestion, error) {
	if len(versionChecks) == 0 {
		return nil, nil
	}

	conditions := make([]string, len(versionChecks))
	values := make([]interface{}, len(versionChecks)*2)
	for i, check := range versionChecks {
		conditions[i] = "(id = ? AND version > ?)"
		values[i*2] = check.ID
		values[i*2+1] = check.Version
	}

	var questions []*writing.WritingQuestion
	query := r.db.WithContext(ctx).
		Where(strings.Join(conditions, " OR "), values...).
		Order("created_at DESC")

	err := query.Find(&questions).Error
	if err != nil {
		r.logger.Error("writing_question_repository.get_new_updated", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get updated questions")
		return nil, err
	}

	return questions, nil
}

const (
	baseSelectQuery = `
		SELECT /*+ INDEX(wq writing_questions_type_idx) */
			id, type, topic, instruction, image_urls, 
			max_time, created_at, updated_at
		FROM writing_questions wq
	`

	baseCountQuery = `
		SELECT COUNT(*) 
		FROM writing_questions wq
	`
)

func (r *WritingQuestionRepository) ExecuteCountQuery(ctx context.Context, query string, args ...interface{}) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&total).Error
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}
	return total, nil
}

func (r *WritingQuestionRepository) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*gorm.DB, error) {
	result := r.db.WithContext(ctx).Raw(query, args...)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to execute query: %w", result.Error)
	}
	return result, nil
}
