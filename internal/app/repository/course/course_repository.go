package course

import (
	"context"
	"errors"
	"fluencybe/pkg/logger"
	"time"

	courseModel "fluencybe/internal/app/model/course"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrCourseNotFound  = errors.New("course not found")
	ErrInvalidInput    = errors.New("invalid input data")
	ErrDuplicateCourse = errors.New("duplicate course")
)

type CourseRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewCourseRepository(db *gorm.DB, logger *logger.PrettyLogger) *CourseRepository {
	return &CourseRepository{
		db:     db,
		logger: logger,
	}
}

func (r *CourseRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *CourseRepository) Create(ctx context.Context, course *courseModel.Course) error {
	now := time.Now()
	course.CreatedAt = now
	course.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(course).Error
	if err != nil {
		r.logger.Error("course_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create course")
		return err
	}
	return nil
}

func (r *CourseRepository) GetByID(ctx context.Context, id uuid.UUID) (*courseModel.Course, error) {
	var result courseModel.Course
	err := r.db.WithContext(ctx).First(&result, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCourseNotFound
		}
		r.logger.Error("course_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get course")
		return nil, err
	}
	return &result, nil
}

func (r *CourseRepository) Update(ctx context.Context, course *courseModel.Course) error {
	course.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&courseModel.Course{}).
		Where("id = ?", course.ID).
		Updates(map[string]interface{}{
			"type":       course.Type,
			"title":      course.Title,
			"overview":   course.Overview,
			"skills":     course.Skills,
			"band":       course.Band,
			"image_urls": course.ImageURLs,
			"updated_at": course.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("course_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    course.ID,
		}, "Failed to update course")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrCourseNotFound
	}

	return nil
}

func (r *CourseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&courseModel.Course{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("course_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete course")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrCourseNotFound
	}

	return nil
}

func (r *CourseRepository) List(ctx context.Context, page, pageSize int) ([]*courseModel.Course, int64, error) {
	var total int64
	var courses []*courseModel.Course

	if err := r.db.WithContext(ctx).Model(&courseModel.Course{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(pageSize).
		Find(&courses).Error; err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}

func (r *CourseRepository) SearchByTitle(ctx context.Context, title string, page, pageSize int) ([]*courseModel.Course, int64, error) {
	var total int64
	var courses []*courseModel.Course

	query := r.db.WithContext(ctx).Model(&courseModel.Course{}).
		Where("title ILIKE ?", "%"+title+"%")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Find(&courses).Error; err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}

func (r *CourseRepository) GetBySkills(ctx context.Context, skills []string, page, pageSize int) ([]*courseModel.Course, int64, error) {
	var total int64
	var courses []*courseModel.Course

	query := r.db.WithContext(ctx).Model(&courseModel.Course{}).
		Where("skills && ?", skills)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Find(&courses).Error; err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}

func (r *CourseRepository) GetByBand(ctx context.Context, band string, page, pageSize int) ([]*courseModel.Course, int64, error) {
	var total int64
	var courses []*courseModel.Course

	query := r.db.WithContext(ctx).Model(&courseModel.Course{}).
		Where("band = ?", band)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Find(&courses).Error; err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}
