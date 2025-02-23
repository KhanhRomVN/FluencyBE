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

type CourseOtherRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewCourseOtherRepository(db *gorm.DB, logger *logger.PrettyLogger) *CourseOtherRepository {
	return &CourseOtherRepository{
		db:     db,
		logger: logger,
	}
}

func (r *CourseOtherRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *CourseOtherRepository) Create(ctx context.Context, courseOther *course.CourseOther) error {
	now := time.Now()
	courseOther.CreatedAt = now
	courseOther.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(courseOther).Error
	if err != nil {
		r.logger.Error("course_other_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create course other")
		return err
	}
	return nil
}

func (r *CourseOtherRepository) GetByID(ctx context.Context, id uuid.UUID) (*course.CourseOther, error) {
	var result course.CourseOther
	err := r.db.WithContext(ctx).First(&result, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCourseNotFound
		}
		r.logger.Error("course_other_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get course other")
		return nil, err
	}
	return &result, nil
}

func (r *CourseOtherRepository) GetByCourseID(ctx context.Context, courseID uuid.UUID) (*course.CourseOther, error) {
	var result course.CourseOther
	err := r.db.WithContext(ctx).First(&result, "course_id = ?", courseID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCourseNotFound
		}
		r.logger.Error("course_other_repository.get_by_course_id", map[string]interface{}{
			"error":    err.Error(),
			"courseID": courseID,
		}, "Failed to get course other")
		return nil, err
	}
	return &result, nil
}

func (r *CourseOtherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&course.CourseOther{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("course_other_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete course other")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrCourseNotFound
	}

	return nil
}
