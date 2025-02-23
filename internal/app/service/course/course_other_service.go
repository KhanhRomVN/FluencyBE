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

type CourseOtherService struct {
	repo          *courseRepo.CourseOtherRepository
	courseRepo    *courseRepo.CourseRepository
	logger        *logger.PrettyLogger
	cache         cache.Cache
	courseUpdator *courseHelper.CourseUpdator
}

func NewCourseOtherService(
	repo *courseRepo.CourseOtherRepository,
	courseRepo *courseRepo.CourseRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	courseUpdator *courseHelper.CourseUpdator,
) *CourseOtherService {
	return &CourseOtherService{
		repo:          repo,
		courseRepo:    courseRepo,
		logger:        logger,
		cache:         cache,
		courseUpdator: courseUpdator,
	}
}

func (s *CourseOtherService) SetCourseUpdator(updator *courseHelper.CourseUpdator) {
	s.courseUpdator = updator
}

func (s *CourseOtherService) validateCourseOther(courseOther *course.CourseOther) error {
	if courseOther == nil {
		return errors.New("invalid input")
	}
	if courseOther.CourseID == uuid.Nil {
		return errors.New("course ID is required")
	}
	return nil
}

func (s *CourseOtherService) Create(ctx context.Context, courseOther *course.CourseOther) error {
	if err := s.validateCourseOther(courseOther); err != nil {
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

	if err := s.repo.Create(ctx, courseOther); err != nil {
		tx.Rollback()
		s.logger.Error("course_other_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create course other")
		return err
	}

	// Get parent course for cache/search update
	parentCourse, err := s.courseRepo.GetByID(ctx, courseOther.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Update cache and search
	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, parentCourse); err != nil {
		tx.Rollback()
		s.logger.Error("course_other_service.update_cache", map[string]interface{}{
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

func (s *CourseOtherService) GetByID(ctx context.Context, id uuid.UUID) (*course.CourseOther, error) {
	courseOther, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("course_other_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get course other by ID")
		return nil, err
	}
	return courseOther, nil
}

func (s *CourseOtherService) GetByCourseID(ctx context.Context, courseID uuid.UUID) (*course.CourseOther, error) {
	courseOther, err := s.repo.GetByCourseID(ctx, courseID)
	if err != nil {
		s.logger.Error("course_other_service.get_by_course_id", map[string]interface{}{
			"error": err.Error(),
			"id":    courseID,
		}, "Failed to get course other by course ID")
		return nil, err
	}
	return courseOther, nil
}

func (s *CourseOtherService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get course other first to get parent ID
	courseOther, err := s.GetByID(ctx, id)
	if err != nil {
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
		s.logger.Error("course_other_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete course other")
		return err
	}

	// Get parent course for cache/search update
	parentCourse, err := s.courseRepo.GetByID(ctx, courseOther.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Update cache and search
	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, parentCourse); err != nil {
		tx.Rollback()
		s.logger.Error("course_other_service.update_cache", map[string]interface{}{
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
