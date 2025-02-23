package reading

import (
	"context"
	"errors"
	"fluencybe/internal/app/model/reading"
	"fluencybe/pkg/logger"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrQuestionNotFound   = errors.New("reading question not found")
	ErrInvalidInput       = errors.New("invalid input data")
	ErrDuplicateQuestion  = errors.New("duplicate question")
	ErrTransactionFailed  = errors.New("transaction failed")
	ErrInvalidQueryParams = errors.New("invalid query parameters")
)

type ReadingQuestionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewReadingQuestionRepository(db *gorm.DB, logger *logger.PrettyLogger) *ReadingQuestionRepository {
	return &ReadingQuestionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ReadingQuestionRepository) CreateReadingQuestion(ctx context.Context, question *reading.ReadingQuestion) error {
	now := time.Now().UTC()
	question.CreatedAt = now
	question.UpdatedAt = now
	question.Version = 1

	// Create record in database
	result := r.db.WithContext(ctx).Create(question)
	if result.Error != nil {
		r.logger.Error("reading_question_repository.create", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to create reading question")
		return result.Error
	}

	return nil
}

func (r *ReadingQuestionRepository) GetReadingQuestionByID(ctx context.Context, id uuid.UUID) (*reading.ReadingQuestion, error) {
	var question reading.ReadingQuestion
	err := r.db.WithContext(ctx).First(&question, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuestionNotFound
		}
		r.logger.Error("reading_question_repository.get_by_id", map[string]interface{}{"error": err.Error()}, "Failed to get reading question")
		return nil, err
	}

	return &question, nil
}

func (r *ReadingQuestionRepository) UpdateReadingQuestion(ctx context.Context, question *reading.ReadingQuestion) error {
	question.UpdatedAt = time.Now().UTC()

	// Start a transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Perform the update
	result := tx.Model(&reading.ReadingQuestion{}).
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
		r.logger.Error("reading_question_repository.update", map[string]interface{}{"error": result.Error.Error()}, "Failed to update reading question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return ErrQuestionNotFound
	}

	return tx.Commit().Error
}

func (r *ReadingQuestionRepository) DeleteReadingQuestion(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&reading.ReadingQuestion{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("reading_question_repository.delete", map[string]interface{}{"error": result.Error.Error()}, "Failed to delete reading question")
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrQuestionNotFound
	}
	return nil
}

func (r *ReadingQuestionRepository) GetAllCompleteQuestions(ctx context.Context) ([]*reading.ReadingQuestion, error) {
	var questions []*reading.ReadingQuestion
	err := r.db.WithContext(ctx).Find(&questions).Error
	if err != nil {
		r.logger.Error("reading_question_repository.get_all", map[string]interface{}{"error": err.Error()}, "Failed to get all reading questions")
		return nil, err
	}
	return questions, nil
}

func (r *ReadingQuestionRepository) SearchQuestions(ctx context.Context, searchPattern string, pageSize int, offset int) (int64, []*reading.ReadingQuestion, error) {
	var questions []*reading.ReadingQuestion
	var total int64

	query := r.db.WithContext(ctx).Model(&reading.ReadingQuestion{})

	if searchPattern != "" {
		pattern := "%" + strings.ToLower(searchPattern) + "%"
		query = query.Where("LOWER(topic) LIKE ? OR LOWER(content) LIKE ?", pattern, pattern)
	}

	err := query.Count(&total).Error
	if err != nil {
		r.logger.Error("reading_question_repository.search_count", map[string]interface{}{"error": err.Error()}, "Failed to count reading questions")
		return 0, nil, err
	}

	err = query.Limit(pageSize).Offset(offset).Find(&questions).Error
	if err != nil {
		r.logger.Error("reading_question_repository.search", map[string]interface{}{"error": err.Error()}, "Failed to search reading questions")
		return 0, nil, err
	}

	return total, questions, nil
}

func (r *ReadingQuestionRepository) GetDB() *gorm.DB {
	return r.db
}

const (
	baseSelectQuery = `
		SELECT /*+ INDEX(rq reading_questions_type_idx) */
			id, type, topic, instruction, content, image_urls, 
			max_time, version, created_at, updated_at
		FROM reading_questions rq
	`
)

func (r *ReadingQuestionRepository) ExecuteCountQuery(ctx context.Context, query string, args ...interface{}) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}
	return count, nil
}

func (r *ReadingQuestionRepository) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*gorm.DB, error) {
	result := r.db.WithContext(ctx).Raw(query, args...)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to execute query: %w", result.Error)
	}
	return result, nil
}

func (r *ReadingQuestionRepository) GetNewUpdatedQuestions(ctx context.Context, versionChecks []struct {
	ID      uuid.UUID
	Version int
}) ([]*reading.ReadingQuestion, error) {
	if len(versionChecks) == 0 {
		return nil, nil
	}

	var conditions []string
	var args []interface{}

	for _, check := range versionChecks {
		conditions = append(conditions, "(id = ? AND version > ?)")
		args = append(args, check.ID, check.Version)
	}

	var questions []*reading.ReadingQuestion
	query := r.db.WithContext(ctx).
		Where(strings.Join(conditions, " OR "), args...).
		Find(&questions)

	if query.Error != nil {
		r.logger.Error("reading_question_repository.get_new_updated", map[string]interface{}{
			"error": query.Error.Error(),
		}, "Failed to get new/updated questions")
		return nil, query.Error
	}

	return questions, nil
}
