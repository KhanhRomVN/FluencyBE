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
	ErrLessonNotFound = errors.New("lesson not found")
)

type LessonRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewLessonRepository(db *gorm.DB, logger *logger.PrettyLogger) *LessonRepository {
	return &LessonRepository{
		db:     db,
		logger: logger,
	}
}

func (r *LessonRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *LessonRepository) Create(ctx context.Context, lesson *course.Lesson) error {
	now := time.Now()
	lesson.CreatedAt = now
	lesson.UpdatedAt = now

	// Lấy sequence tự động nếu chưa được set
	if lesson.Sequence <= 0 {
		var nextSeq int
		err := r.db.WithContext(ctx).Raw(
			"SELECT get_next_lesson_sequence($1)",
			lesson.CourseID,
		).Scan(&nextSeq).Error
		if err != nil {
			r.logger.Error("lesson_repository.get_next_sequence", map[string]interface{}{
				"error":    err.Error(),
				"courseID": lesson.CourseID,
			}, "Failed to get next sequence")
			return err
		}
		lesson.Sequence = nextSeq
	}

	err := r.db.WithContext(ctx).Create(lesson).Error
	if err != nil {
		r.logger.Error("lesson_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create lesson")
		return err
	}
	return nil
}

func (r *LessonRepository) GetByID(ctx context.Context, id uuid.UUID) (*course.Lesson, error) {
	var result course.Lesson
	err := r.db.WithContext(ctx).First(&result, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLessonNotFound
		}
		r.logger.Error("lesson_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get lesson")
		return nil, err
	}
	return &result, nil
}

func (r *LessonRepository) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]*course.Lesson, error) {
	var lessons []*course.Lesson
	err := r.db.WithContext(ctx).
		Where("course_id = ?", courseID).
		Order("sequence").
		Find(&lessons).Error

	if err != nil {
		r.logger.Error("lesson_repository.get_by_course_id", map[string]interface{}{
			"error":    err.Error(),
			"courseID": courseID,
		}, "Failed to get lessons")
		return nil, err
	}
	return lessons, nil
}

func (r *LessonRepository) Update(ctx context.Context, lesson *course.Lesson) error {
	lesson.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&course.Lesson{}).
		Where("id = ?", lesson.ID).
		Updates(map[string]interface{}{
			"sequence":   lesson.Sequence,
			"title":      lesson.Title,
			"overview":   lesson.Overview,
			"updated_at": lesson.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("lesson_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    lesson.ID,
		}, "Failed to update lesson")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrLessonNotFound
	}

	return nil
}

func (r *LessonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&course.Lesson{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("lesson_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete lesson")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrLessonNotFound
	}

	return nil
}

func (r *LessonRepository) SwapSequence(ctx context.Context, lesson1, lesson2 *course.Lesson) error {
	result := r.db.WithContext(ctx).Exec(
		"SELECT swap_lesson_sequence($1, $2)",
		lesson1.ID,
		lesson2.ID,
	)

	if result.Error != nil {
		r.logger.Error("lesson_repository.swap_sequence", map[string]interface{}{
			"error": result.Error.Error(),
			"id1":   lesson1.ID,
			"id2":   lesson2.ID,
		}, "Failed to swap lesson sequences")
		return result.Error
	}

	// Cập nhật các giá trị trong struct
	tempSeq := lesson1.Sequence
	lesson1.Sequence = lesson2.Sequence
	lesson2.Sequence = tempSeq

	return nil
}
