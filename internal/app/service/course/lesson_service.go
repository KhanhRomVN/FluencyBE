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

type LessonService struct {
	repo          *courseRepo.LessonRepository
	courseRepo    *courseRepo.CourseRepository
	logger        *logger.PrettyLogger
	cache         cache.Cache
	courseUpdator *courseHelper.CourseUpdator
}

func NewLessonService(
	repo *courseRepo.LessonRepository,
	courseRepo *courseRepo.CourseRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	courseUpdator *courseHelper.CourseUpdator,
) *LessonService {
	return &LessonService{
		repo:          repo,
		courseRepo:    courseRepo,
		logger:        logger,
		cache:         cache,
		courseUpdator: courseUpdator,
	}
}

func (s *LessonService) SetCourseUpdator(updator *courseHelper.CourseUpdator) {
	s.courseUpdator = updator
}

func (s *LessonService) validateLesson(lesson *course.Lesson) error {
	if lesson == nil {
		return errors.New("invalid input")
	}
	if lesson.CourseID == uuid.Nil {
		return errors.New("course ID is required")
	}
	if lesson.Title == "" {
		return errors.New("title is required")
	}
	if lesson.Overview == "" {
		return errors.New("overview is required")
	}
	return nil
}

func (s *LessonService) Create(ctx context.Context, lesson *course.Lesson) error {
	if err := s.validateLesson(lesson); err != nil {
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

	if err := s.repo.Create(ctx, lesson); err != nil {
		tx.Rollback()
		s.logger.Error("lesson_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create lesson")
		return err
	}

	// Get parent course for cache/search update
	parentCourse, err := s.courseRepo.GetByID(ctx, lesson.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Update cache and search
	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, parentCourse); err != nil {
		tx.Rollback()
		s.logger.Error("lesson_service.update_cache", map[string]interface{}{
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

func (s *LessonService) GetByID(ctx context.Context, id uuid.UUID) (*course.Lesson, error) {
	lesson, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("lesson_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get lesson by ID")
		return nil, err
	}
	return lesson, nil
}

func (s *LessonService) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]*course.Lesson, error) {
	lessons, err := s.repo.GetByCourseID(ctx, courseID)
	if err != nil {
		s.logger.Error("lesson_service.get_by_course_id", map[string]interface{}{
			"error": err.Error(),
			"id":    courseID,
		}, "Failed to get lessons")
		return nil, err
	}
	return lessons, nil
}

func (s *LessonService) Update(ctx context.Context, lesson *course.Lesson) error {
	if err := s.validateLesson(lesson); err != nil {
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

	if err := s.repo.Update(ctx, lesson); err != nil {
		tx.Rollback()
		s.logger.Error("lesson_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    lesson.ID,
		}, "Failed to update lesson")
		return err
	}

	// Get parent course for cache/search update
	parentCourse, err := s.courseRepo.GetByID(ctx, lesson.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Update cache and search
	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, parentCourse); err != nil {
		tx.Rollback()
		s.logger.Error("lesson_service.update_cache", map[string]interface{}{
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

func (s *LessonService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get lesson first to get course ID for cache update
	lesson, err := s.GetByID(ctx, id)
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

	// Delete the lesson (resequencing is handled by trigger)
	if err := s.repo.Delete(ctx, id); err != nil {
		tx.Rollback()
		return err
	}

	// Get parent course for cache/search update
	parentCourse, err := s.courseRepo.GetByID(ctx, lesson.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Update cache and search
	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, parentCourse); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *LessonService) SwapSequence(ctx context.Context, id1, id2 uuid.UUID) error {
	// Get both lessons
	lesson1, err := s.GetByID(ctx, id1)
	if err != nil {
		return err
	}

	lesson2, err := s.GetByID(ctx, id2)
	if err != nil {
		return err
	}

	// Verify lessons belong to same course
	if lesson1.CourseID != lesson2.CourseID {
		return errors.New("lessons must belong to the same course")
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

	// Swap sequences using database function
	if err := s.repo.SwapSequence(ctx, lesson1, lesson2); err != nil {
		tx.Rollback()
		return err
	}

	// Update cache and search
	parentCourse, err := s.courseRepo.GetByID(ctx, lesson1.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, parentCourse); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
