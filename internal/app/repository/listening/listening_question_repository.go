package listening

import (
	"context"
	"errors"
	"fluencybe/internal/app/model/listening"
	"fluencybe/pkg/logger"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrQuestionNotFound   = errors.New("listening question not found")
	ErrInvalidInput       = errors.New("invalid input data")
	ErrDuplicateQuestion  = errors.New("duplicate question")
	ErrTransactionFailed  = errors.New("transaction failed")
	ErrInvalidQueryParams = errors.New("invalid query parameters")
)

type ListeningQuestionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewListeningQuestionRepository(db *gorm.DB, logger *logger.PrettyLogger) *ListeningQuestionRepository {
	return &ListeningQuestionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ListeningQuestionRepository) CreateListeningQuestion(ctx context.Context, question *listening.ListeningQuestion) error {
	now := time.Now().UTC()
	question.CreatedAt = now
	question.UpdatedAt = now
	question.Version = 1

	// Create record in database
	result := r.db.WithContext(ctx).Create(question)
	if result.Error != nil {
		r.logger.Error("listening_question_repository.create", map[string]interface{}{
			"error": result.Error.Error(),
		}, "Failed to create listening question")
		return result.Error
	}

	return nil
}

func (r *ListeningQuestionRepository) GetListeningQuestionByID(ctx context.Context, id uuid.UUID) (*listening.ListeningQuestion, error) {
	var question listening.ListeningQuestion
	err := r.db.WithContext(ctx).First(&question, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuestionNotFound
		}
		r.logger.Error("listening_question_repository.get_by_id", map[string]interface{}{"error": err.Error()}, "Failed to get listening question")
		return nil, err
	}

	return &question, nil
}

func (r *ListeningQuestionRepository) UpdateListeningQuestion(ctx context.Context, question *listening.ListeningQuestion) error {
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
	result := tx.Model(&listening.ListeningQuestion{}).
		Where("id = ?", question.ID).
		Updates(map[string]interface{}{
			"type":        question.Type,
			"topic":       question.Topic,
			"instruction": question.Instruction,
			"audio_urls":  question.AudioURLs,
			"image_urls":  question.ImageURLs,
			"transcript":  question.Transcript,
			"max_time":    question.MaxTime,
			"version":     question.Version,
			"updated_at":  question.UpdatedAt,
		})

	if result.Error != nil {
		tx.Rollback()
		r.logger.Error("listening_question_repository.update", map[string]interface{}{"error": result.Error.Error()}, "Failed to update listening question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return ErrQuestionNotFound
	}

	// Get the updated question with the new version
	if err := tx.First(question, "id = ?", question.ID).Error; err != nil {
		tx.Rollback()
		r.logger.Error("listening_question_repository.update.get_updated", map[string]interface{}{"error": err.Error()}, "Failed to get updated question")
		return err
	}

	return tx.Commit().Error
}

func (r *ListeningQuestionRepository) DeleteListeningQuestion(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&listening.ListeningQuestion{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("listening_question_repository.delete", map[string]interface{}{"error": result.Error.Error()}, "Failed to delete listening question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrQuestionNotFound
	}

	return nil
}

func (r *ListeningQuestionRepository) GetAllCompleteQuestions(ctx context.Context) ([]*listening.ListeningQuestion, error) {
	var questions []*listening.ListeningQuestion

	err := r.db.WithContext(ctx).
		Preload("FillInTheBlankQuestion").
		Preload("FillInTheBlankAnswers").
		Order("created_at DESC").
		Find(&questions).Error

	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *ListeningQuestionRepository) SearchQuestions(ctx context.Context, searchPattern string, pageSize int, offset int) (int64, []*listening.ListeningQuestion, error) {
	var total int64
	var questions []*listening.ListeningQuestion

	// Count total matches
	err := r.db.WithContext(ctx).Model(&listening.ListeningQuestion{}).
		Where("instruction ILIKE ? OR transcript ILIKE ?", "%"+searchPattern+"%", "%"+searchPattern+"%").
		Count(&total).Error
	if err != nil {
		return 0, nil, err
	}

	// Get paginated results
	err = r.db.WithContext(ctx).
		Where("instruction ILIKE ? OR transcript ILIKE ?", "%"+searchPattern+"%", "%"+searchPattern+"%").
		Limit(pageSize).
		Offset(offset).
		Find(&questions).Error
	if err != nil {
		return 0, nil, err
	}

	return total, questions, nil
}

func (r *ListeningQuestionRepository) GetDB() *gorm.DB {
	return r.db
}

const (
	baseSelectQuery = `
		SELECT /*+ INDEX(lq listening_questions_type_idx) */
			id, type, topic, instruction, audio_urls, image_urls, 
			transcript, max_time, created_at, updated_at
		FROM listening_questions lq
	`

	baseCountQuery = `
		SELECT COUNT(*) 
		FROM listening_questions lq
	`
)

func (r *ListeningQuestionRepository) ExecuteCountQuery(ctx context.Context, query string, args ...interface{}) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&total).Error
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}
	return total, nil
}

func (r *ListeningQuestionRepository) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*gorm.DB, error) {
	result := r.db.WithContext(ctx).Raw(query, args...)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to execute query: %w", result.Error)
	}
	return result, nil
}

func (r *ListeningQuestionRepository) GetNewUpdatedQuestions(ctx context.Context, versionChecks []struct {
	ID      uuid.UUID
	Version int
}) ([]*listening.ListeningQuestion, error) {
	if len(versionChecks) == 0 {
		return nil, nil
	}

	// Build query conditions
	conditions := make([]string, len(versionChecks))
	values := make([]interface{}, len(versionChecks)*2)
	for i, check := range versionChecks {
		conditions[i] = "(id = ? AND version > ?)"
		values[i*2] = check.ID
		values[i*2+1] = check.Version
	}

	var questions []*listening.ListeningQuestion
	query := r.db.WithContext(ctx).
		Where(strings.Join(conditions, " OR "), values...).
		Order("created_at DESC")

	err := query.Find(&questions).Error
	if err != nil {
		r.logger.Error("listening_question_repository.get_new_updated", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get updated questions")
		return nil, err
	}

	return questions, nil
}
