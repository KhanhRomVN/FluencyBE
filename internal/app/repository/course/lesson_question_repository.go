package course

import (
	"context"
	"errors"
	"fluencybe/internal/app/model/course"
	"fluencybe/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrLessonQuestionNotFound = errors.New("lesson question not found")
)

type LessonQuestionRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewLessonQuestionRepository(db *gorm.DB, logger *logger.PrettyLogger) *LessonQuestionRepository {
	return &LessonQuestionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *LessonQuestionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *LessonQuestionRepository) Create(ctx context.Context, question *course.LessonQuestion) error {
	now := time.Now()
	question.CreatedAt = now
	question.UpdatedAt = now

	// Lấy sequence tự động nếu chưa được set
	if question.Sequence <= 0 {
		var nextSeq int
		err := r.db.WithContext(ctx).Raw(
			"SELECT get_next_lesson_question_sequence($1)",
			question.LessonID,
		).Scan(&nextSeq).Error
		if err != nil {
			r.logger.Error("lesson_question_repository.get_next_sequence", map[string]interface{}{
				"error":    err.Error(),
				"lessonID": question.LessonID,
			}, "Failed to get next sequence")
			return err
		}
		question.Sequence = nextSeq
	}

	err := r.db.WithContext(ctx).Create(question).Error
	if err != nil {
		r.logger.Error("lesson_question_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create lesson question")
		return err
	}
	return nil
}

func (r *LessonQuestionRepository) GetByID(ctx context.Context, id uuid.UUID) (*course.LessonQuestion, error) {
	var result course.LessonQuestion
	err := r.db.WithContext(ctx).First(&result, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLessonQuestionNotFound
		}
		r.logger.Error("lesson_question_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get lesson question")
		return nil, err
	}
	return &result, nil
}

func (r *LessonQuestionRepository) GetByLessonID(ctx context.Context, lessonID uuid.UUID) ([]*course.LessonQuestion, error) {
	var questions []*course.LessonQuestion
	err := r.db.WithContext(ctx).
		Where("lesson_id = ?", lessonID).
		Order("sequence").
		Find(&questions).Error

	if err != nil {
		r.logger.Error("lesson_question_repository.get_by_lesson_id", map[string]interface{}{
			"error":    err.Error(),
			"lessonID": lessonID,
		}, "Failed to get lesson questions")
		return nil, err
	}
	return questions, nil
}

func (r *LessonQuestionRepository) Update(ctx context.Context, question *course.LessonQuestion) error {
	question.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&course.LessonQuestion{}).
		Where("id = ?", question.ID).
		Updates(map[string]interface{}{
			"sequence":      question.Sequence,
			"question_id":   question.QuestionID,
			"question_type": question.QuestionType,
			"updated_at":    question.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("lesson_question_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    question.ID,
		}, "Failed to update lesson question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrLessonQuestionNotFound
	}

	return nil
}

func (r *LessonQuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&course.LessonQuestion{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("lesson_question_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete lesson question")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrLessonQuestionNotFound
	}

	return nil
}

func (r *LessonQuestionRepository) SwapSequence(ctx context.Context, q1, q2 *course.LessonQuestion) error {
	result := r.db.WithContext(ctx).Exec(
		"SELECT swap_lesson_question_sequence($1, $2)",
		q1.ID,
		q2.ID,
	)

	if result.Error != nil {
		r.logger.Error("lesson_question_repository.swap_sequence", map[string]interface{}{
			"error": result.Error.Error(),
			"id1":   q1.ID,
			"id2":   q2.ID,
		}, "Failed to swap question sequences")
		return result.Error
	}

	// Cập nhật các giá trị trong struct
	tempSeq := q1.Sequence
	q1.Sequence = q2.Sequence
	q2.Sequence = tempSeq

	return nil
}
