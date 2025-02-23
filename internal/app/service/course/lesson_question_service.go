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

type LessonQuestionService struct {
	repo          *courseRepo.LessonQuestionRepository
	lessonRepo    *courseRepo.LessonRepository
	courseRepo    *courseRepo.CourseRepository
	logger        *logger.PrettyLogger
	cache         cache.Cache
	courseUpdator *courseHelper.CourseUpdator
}

func NewLessonQuestionService(
	repo *courseRepo.LessonQuestionRepository,
	lessonRepo *courseRepo.LessonRepository,
	courseRepo *courseRepo.CourseRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	courseUpdator *courseHelper.CourseUpdator,
) *LessonQuestionService {
	return &LessonQuestionService{
		repo:          repo,
		lessonRepo:    lessonRepo,
		courseRepo:    courseRepo,
		logger:        logger,
		cache:         cache,
		courseUpdator: courseUpdator,
	}
}

func (s *LessonQuestionService) SetCourseUpdator(updator *courseHelper.CourseUpdator) {
	s.courseUpdator = updator
}

func (s *LessonQuestionService) validateLessonQuestion(question *course.LessonQuestion) error {
	if question == nil {
		return errors.New("invalid input")
	}
	if question.LessonID == uuid.Nil {
		return errors.New("lesson ID is required")
	}
	if question.QuestionID == uuid.Nil {
		return errors.New("question ID is required")
	}
	if question.QuestionType == "" {
		return errors.New("question type is required")
	}
	return nil
}

func (s *LessonQuestionService) Create(ctx context.Context, question *course.LessonQuestion) error {
	if err := s.validateLessonQuestion(question); err != nil {
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

	if err := s.repo.Create(ctx, question); err != nil {
		tx.Rollback()
		s.logger.Error("lesson_question_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create lesson question")
		return err
	}

	// Get parent lesson
	lesson, err := s.lessonRepo.GetByID(ctx, question.LessonID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get lesson: %w", err)
	}

	// Get course for cache/search update
	course, err := s.courseRepo.GetByID(ctx, lesson.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Update cache and search
	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, course); err != nil {
		tx.Rollback()
		s.logger.Error("lesson_question_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    course.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *LessonQuestionService) GetByID(ctx context.Context, id uuid.UUID) (*course.LessonQuestion, error) {
	question, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("lesson_question_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get lesson question by ID")
		return nil, err
	}
	return question, nil
}

func (s *LessonQuestionService) GetByLessonID(ctx context.Context, lessonID uuid.UUID) ([]*course.LessonQuestion, error) {
	questions, err := s.repo.GetByLessonID(ctx, lessonID)
	if err != nil {
		s.logger.Error("lesson_question_service.get_by_lesson_id", map[string]interface{}{
			"error": err.Error(),
			"id":    lessonID,
		}, "Failed to get lesson questions")
		return nil, err
	}
	return questions, nil
}

func (s *LessonQuestionService) Update(ctx context.Context, question *course.LessonQuestion) error {
	if err := s.validateLessonQuestion(question); err != nil {
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

	if err := s.repo.Update(ctx, question); err != nil {
		tx.Rollback()
		s.logger.Error("lesson_question_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update lesson question")
		return err
	}

	// Get parent lesson
	lesson, err := s.lessonRepo.GetByID(ctx, question.LessonID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get lesson: %w", err)
	}

	// Get course for cache/search update
	course, err := s.courseRepo.GetByID(ctx, lesson.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Update cache and search
	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, course); err != nil {
		tx.Rollback()
		s.logger.Error("lesson_question_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    course.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *LessonQuestionService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get question first to get parent IDs
	question, err := s.GetByID(ctx, id)
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

	// Delete the question (resequencing is handled by trigger)
	if err := s.repo.Delete(ctx, id); err != nil {
		tx.Rollback()
		return err
	}

	// Get parent lesson
	lesson, err := s.lessonRepo.GetByID(ctx, question.LessonID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get lesson: %w", err)
	}

	// Get course for cache/search update
	course, err := s.courseRepo.GetByID(ctx, lesson.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Update cache and search
	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, course); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *LessonQuestionService) SwapSequence(ctx context.Context, id1, id2 uuid.UUID) error {
	// Get both questions
	q1, err := s.GetByID(ctx, id1)
	if err != nil {
		return err
	}

	q2, err := s.GetByID(ctx, id2)
	if err != nil {
		return err
	}

	// Verify questions belong to same lesson
	if q1.LessonID != q2.LessonID {
		return errors.New("questions must belong to the same lesson")
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
	if err := s.repo.SwapSequence(ctx, q1, q2); err != nil {
		tx.Rollback()
		return err
	}

	// Get parent lesson
	lesson, err := s.lessonRepo.GetByID(ctx, q1.LessonID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get lesson: %w", err)
	}

	// Get course for cache/search update
	course, err := s.courseRepo.GetByID(ctx, lesson.CourseID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Update cache and search
	if err := s.courseUpdator.UpdateCacheAndSearch(ctx, course); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
