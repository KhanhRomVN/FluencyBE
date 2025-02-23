package course

import (
	"context"
	"errors"
	courseHelper "fluencybe/internal/app/helper/course"
	"fluencybe/internal/app/model/course"
	courseRepo "fluencybe/internal/app/repository/course"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
)

type CourseBookService struct {
	repo          *courseRepo.CourseBookRepository
	courseRepo    *courseRepo.CourseRepository
	logger        *logger.PrettyLogger
	cache         cache.Cache
	courseUpdator *courseHelper.CourseUpdator
}

func NewCourseBookService(
	repo *courseRepo.CourseBookRepository,
	courseRepo *courseRepo.CourseRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	courseUpdator *courseHelper.CourseUpdator,
) *CourseBookService {
	return &CourseBookService{
		repo:          repo,
		courseRepo:    courseRepo,
		logger:        logger,
		cache:         cache,
		courseUpdator: courseUpdator,
	}
}

func (s *CourseBookService) SetCourseUpdator(updator *courseHelper.CourseUpdator) {
	s.courseUpdator = updator
}

func (s *CourseBookService) validateCourseBook(courseBook *course.CourseBook) error {
	if courseBook == nil {
		return errors.New("invalid input")
	}
	if courseBook.CourseID == uuid.Nil {
		return errors.New("course ID is required")
	}
	if len(courseBook.Publishers) == 0 {
		return errors.New("at least one publisher is required")
	}
	if len(courseBook.Authors) == 0 {
		return errors.New("at least one author is required")
	}
	if courseBook.PublicationYear < 1900 {
		return errors.New("publication year must be 1900 or later")
	}
	return nil
}

func (s *CourseBookService) Create(ctx context.Context, courseBook *course.CourseBook) error {
	if err := s.validateCourseBook(courseBook); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Start transaction
	tx := s.repo.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.repo.Create(ctx, courseBook); err != nil {
		tx.Rollback()
		s.logger.Error("course_book_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create course book")
		return err
	}

	// Get parent course for cache/search update
	parentCourse, err := s.courseRepo.GetByID(ctx, courseBook.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Update cache and search
	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, parentCourse); err != nil {
		tx.Rollback()
		s.logger.Error("course_book_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentCourse.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *CourseBookService) GetByID(ctx context.Context, id uuid.UUID) (*course.CourseBook, error) {
	courseBook, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("course_book_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get course book by ID")
		return nil, err
	}
	return courseBook, nil
}

func (s *CourseBookService) GetByCourseID(ctx context.Context, courseID uuid.UUID) (*course.CourseBook, error) {
	courseBook, err := s.repo.GetByCourseID(ctx, courseID)
	if err != nil {
		s.logger.Error("course_book_service.get_by_course_id", map[string]interface{}{
			"error": err.Error(),
			"id":    courseID,
		}, "Failed to get course book by course ID")
		return nil, err
	}
	return courseBook, nil
}

func (s *CourseBookService) Update(ctx context.Context, courseBook *course.CourseBook) error {
	if err := s.validateCourseBook(courseBook); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Start transaction
	tx := s.repo.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.repo.Update(ctx, courseBook); err != nil {
		tx.Rollback()
		s.logger.Error("course_book_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    courseBook.ID,
		}, "Failed to update course book")
		return err
	}

	// Get parent course for cache/search update
	parentCourse, err := s.courseRepo.GetByID(ctx, courseBook.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Update cache and search
	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, parentCourse); err != nil {
		tx.Rollback()
		s.logger.Error("course_book_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentCourse.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *CourseBookService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get course book first to get parent ID
	courseBook, err := s.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, courseRepo.ErrCourseNotFound) {
			return courseRepo.ErrCourseNotFound
		}
		return err
	}

	// Start transaction
	tx := s.repo.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.repo.Delete(ctx, id); err != nil {
		tx.Rollback()
		s.logger.Error("course_book_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete course book")
		return err
	}

	// Get parent course for cache/search update
	parentCourse, err := s.courseRepo.GetByID(ctx, courseBook.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Update cache and search
	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, parentCourse); err != nil {
		tx.Rollback()
		s.logger.Error("course_book_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentCourse.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
