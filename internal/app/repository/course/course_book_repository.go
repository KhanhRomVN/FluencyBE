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

type CourseBookRepository struct {
	db     *gorm.DB
	logger *logger.PrettyLogger
}

func NewCourseBookRepository(db *gorm.DB, logger *logger.PrettyLogger) *CourseBookRepository {
	return &CourseBookRepository{
		db:     db,
		logger: logger,
	}
}

func (r *CourseBookRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *CourseBookRepository) Create(ctx context.Context, courseBook *course.CourseBook) error {
	now := time.Now()
	courseBook.CreatedAt = now
	courseBook.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(courseBook).Error
	if err != nil {
		r.logger.Error("course_book_repository.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create course book")
		return err
	}
	return nil
}

func (r *CourseBookRepository) GetByID(ctx context.Context, id uuid.UUID) (*course.CourseBook, error) {
	var result course.CourseBook
	err := r.db.WithContext(ctx).First(&result, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCourseNotFound
		}
		r.logger.Error("course_book_repository.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get course book")
		return nil, err
	}
	return &result, nil
}

func (r *CourseBookRepository) GetByCourseID(ctx context.Context, courseID uuid.UUID) (*course.CourseBook, error) {
	var result course.CourseBook
	err := r.db.WithContext(ctx).First(&result, "course_id = ?", courseID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCourseNotFound
		}
		r.logger.Error("course_book_repository.get_by_course_id", map[string]interface{}{
			"error":    err.Error(),
			"courseID": courseID,
		}, "Failed to get course book")
		return nil, err
	}
	return &result, nil
}

func (r *CourseBookRepository) Update(ctx context.Context, courseBook *course.CourseBook) error {
	courseBook.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&course.CourseBook{}).
		Where("id = ?", courseBook.ID).
		Updates(map[string]interface{}{
			"publishers":       courseBook.Publishers,
			"authors":          courseBook.Authors,
			"publication_year": courseBook.PublicationYear,
			"updated_at":       courseBook.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("course_book_repository.update", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    courseBook.ID,
		}, "Failed to update course book")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrCourseNotFound
	}

	return nil
}

func (r *CourseBookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&course.CourseBook{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("course_book_repository.delete", map[string]interface{}{
			"error": result.Error.Error(),
			"id":    id,
		}, "Failed to delete course book")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrCourseNotFound
	}

	return nil
}
