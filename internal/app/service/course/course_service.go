package course

import (
	"context"
	"errors"
	courseHelper "fluencybe/internal/app/helper/course"
	"fluencybe/internal/app/model/course"
	redisClient "fluencybe/internal/app/redis"
	courseRepo "fluencybe/internal/app/repository/course"
	courseValidator "fluencybe/internal/app/validator"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"
	"time"

	courseDTO "fluencybe/internal/app/dto"
	searchClient "fluencybe/internal/app/opensearch"

	"github.com/google/uuid"
)

var (
	ErrCourseNotFound = errors.New("course not found")
	ErrInvalidInput   = errors.New("invalid input")
)

type CourseService struct {
	repo                  *courseRepo.CourseRepository
	logger                *logger.PrettyLogger
	redis                 *redisClient.CourseRedis
	search                *searchClient.CourseSearch
	updater               *courseHelper.CourseFieldUpdater
	lessonService         *LessonService
	lessonQuestionService *LessonQuestionService
	courseOtherService    *CourseOtherService
	courseBookService     *CourseBookService
	courseUpdator         *courseHelper.CourseUpdator
}

func NewCourseService(
	repo *courseRepo.CourseRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	search *searchClient.CourseSearch,
	lessonService *LessonService,
	lessonQuestionService *LessonQuestionService,
	courseOtherService *CourseOtherService,
	courseBookService *CourseBookService,
	courseUpdator *courseHelper.CourseUpdator,
) *CourseService {
	return &CourseService{
		repo:    repo,
		logger:  logger,
		redis:   redisClient.NewCourseRedis(cache, logger),
		search:  search,
		updater: courseHelper.NewCourseFieldUpdater(logger),

		lessonService:         lessonService,
		lessonQuestionService: lessonQuestionService,
		courseBookService:     courseBookService,
		courseOtherService:    courseOtherService,
		courseUpdator:         courseUpdator,
	}
}

func (s *CourseService) Create(ctx context.Context, course *course.Course) error {
	if err := courseValidator.ValidateCourse(course); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if err := s.repo.Create(ctx, course); err != nil {
		s.logger.Error("course_service.create", map[string]interface{}{
			"error": err.Error(),
			"type":  course.Type,
		}, "Failed to create course")
		return err
	}

	courseDetail := courseDTO.CourseDetail{
		CourseResponse: courseDTO.CourseResponse{
			ID:        course.ID,
			Type:      course.Type,
			Title:     course.Title,
			Overview:  course.Overview,
			Skills:    course.Skills,
			Band:      course.Band,
			ImageURLs: course.ImageURLs,
			CreatedAt: course.CreatedAt,
			UpdatedAt: course.UpdatedAt,
		},
	}

	// Cache the course
	s.redis.SetCacheCourseDetail(ctx, &courseDetail)

	// Index to OpenSearch
	if err := s.search.UpsertCourse(ctx, &courseDetail); err != nil {
		s.logger.Error("course_service.create.opensearch", map[string]interface{}{
			"error": err.Error(),
			"id":    course.ID,
		}, "Failed to index course in OpenSearch")
	}

	return nil
}

func (s *CourseService) GetByID(ctx context.Context, id uuid.UUID) (*courseDTO.CourseDetail, error) {
	// Try to get from Redis first
	cachedCourse, err := s.redis.GetCacheCourseDetail(ctx, id)
	if err == nil && cachedCourse != nil {
		return cachedCourse, nil
	}

	// If not in cache or error getting from cache, get from database
	course, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("course_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get course")
		return nil, err
	}

	response, err := s.BuildCourseDetail(ctx, course)
	if err != nil {
		s.logger.Error("course_service.get.detail", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get course details")
		return nil, err
	}

	// Cache the course for future requests
	if err := s.redis.SetCacheCourseDetail(ctx, response); err != nil {
		s.logger.Error("course_service.get.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to cache course detail")
		// Continue even if caching fails
	}

	return response, nil
}

func (s *CourseService) Update(ctx context.Context, id uuid.UUID, update courseDTO.UpdateCourseFieldRequest) error {
	// First get the base course from database
	baseCourse, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, courseRepo.ErrCourseNotFound) {
			return ErrCourseNotFound
		}
		return fmt.Errorf("failed to get course: %w", err)
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

	// Update the base course fields
	if err := s.updater.UpdateField(baseCourse, update); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update field: %w", err)
	}

	// Update timestamps
	baseCourse.UpdatedAt = time.Now()

	// Update in database
	if err := s.repo.Update(ctx, baseCourse); err != nil {
		tx.Rollback()
		s.logger.Error("course_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    baseCourse.ID,
		}, "Failed to update course in database")
		return fmt.Errorf("failed to update course in database: %w", err)
	}

	// First invalidate existing cache
	if err := s.redis.RemoveCourseCacheEntries(ctx, baseCourse.ID); err != nil {
		s.logger.Error("course_service.update.remove_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    baseCourse.ID,
		}, "Failed to remove old cache entries")
	}

	// Build course detail for cache and search update
	courseDetail := &courseDTO.CourseDetail{
		CourseResponse: courseDTO.CourseResponse{
			ID:        baseCourse.ID,
			Type:      baseCourse.Type,
			Title:     baseCourse.Title,
			Overview:  baseCourse.Overview,
			Skills:    baseCourse.Skills,
			Band:      baseCourse.Band,
			ImageURLs: baseCourse.ImageURLs,
			CreatedAt: baseCourse.CreatedAt,
			UpdatedAt: baseCourse.UpdatedAt,
		},
	}

	// Try to load additional data but don't fail if not found
	if baseCourse.Type == "BOOK" {
		courseBook, err := s.courseBookService.GetByCourseID(ctx, baseCourse.ID)
		if err == nil && courseBook != nil {
			courseDetail.CourseBook = &courseDTO.CourseBookResponse{
				ID:              courseBook.ID,
				CourseID:        courseBook.CourseID,
				Publishers:      courseBook.Publishers,
				Authors:         courseBook.Authors,
				PublicationYear: courseBook.PublicationYear,
			}
		}
	} else if baseCourse.Type == "OTHER" {
		courseOther, err := s.courseOtherService.GetByCourseID(ctx, baseCourse.ID)
		if err == nil && courseOther != nil {
			courseDetail.CourseOther = &courseDTO.CourseOtherResponse{
				ID:       courseOther.ID,
				CourseID: courseOther.CourseID,
			}
		}
	}

	// Load lessons if available
	lessons, err := s.lessonService.GetByCourseID(ctx, baseCourse.ID)
	if err == nil && len(lessons) > 0 {
		courseDetail.Lessons = make([]courseDTO.LessonResponse, len(lessons))
		for i, lesson := range lessons {
			courseDetail.Lessons[i] = courseDTO.LessonResponse{
				ID:       lesson.ID,
				CourseID: lesson.CourseID,
				Sequence: lesson.Sequence,
				Title:    lesson.Title,
				Overview: lesson.Overview,
			}

			// Try to load questions
			questions, err := s.lessonQuestionService.GetByLessonID(ctx, lesson.ID)
			if err == nil && len(questions) > 0 {
				courseDetail.Lessons[i].Questions = make([]courseDTO.LessonQuestionResponse, len(questions))
				for j, question := range questions {
					courseDetail.Lessons[i].Questions[j] = courseDTO.LessonQuestionResponse{
						ID:           question.ID,
						LessonID:     question.LessonID,
						Sequence:     question.Sequence,
						QuestionID:   question.QuestionID,
						QuestionType: question.QuestionType,
					}
				}
			}
		}
	}

	// Update cache with new data
	if err := s.redis.SetCacheCourseDetail(ctx, courseDetail); err != nil {
		s.logger.Error("course_service.update.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    baseCourse.ID,
		}, "Failed to update cache")
	}

	// Update search index
	if err := s.search.UpsertCourse(ctx, courseDetail); err != nil {
		s.logger.Error("course_service.update.search", map[string]interface{}{
			"error": err.Error(),
			"id":    baseCourse.ID,
		}, "Failed to update search index")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *CourseService) Delete(ctx context.Context, id uuid.UUID) error {
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
		s.logger.Error("course_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete course")
		return err
	}

	// Invalidate cache
	if err := s.redis.RemoveCourseCacheEntries(ctx, id); err != nil {
		tx.Rollback()
		s.logger.Error("course_service.delete.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to remove cache entries")
		return err
	}

	// Delete from OpenSearch
	if err := s.search.DeleteCourseFromIndex(ctx, id); err != nil {
		tx.Rollback()
		s.logger.Error("course_service.delete.opensearch", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete from OpenSearch")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *CourseService) SearchWithFilter(ctx context.Context, filter courseDTO.CourseSearchRequest) (*courseDTO.ListCoursePagination, error) {
	// Add debug logging
	s.logger.Debug("course_service.search.start", map[string]interface{}{
		"filter": filter,
	}, "Starting course search")

	result, err := s.search.SearchCourses(ctx, filter)
	if err != nil {
		s.logger.Error("course_service.search", map[string]interface{}{
			"error":  err.Error(),
			"filter": filter,
		}, "Failed to search courses")
		return nil, fmt.Errorf("failed to search courses: %w", err)
	}

	// Add debug logging for results
	s.logger.Debug("course_service.search.complete", map[string]interface{}{
		"total_results": result.Total,
		"page":          result.Page,
		"page_size":     result.PageSize,
	}, "Search completed")

	return result, nil
}

func (s *CourseService) BuildCourseDetail(ctx context.Context, course *course.Course) (*courseDTO.CourseDetail, error) {
	response := &courseDTO.CourseDetail{
		CourseResponse: courseDTO.CourseResponse{
			ID:        course.ID,
			Type:      course.Type,
			Title:     course.Title,
			Overview:  course.Overview,
			Skills:    course.Skills,
			Band:      course.Band,
			ImageURLs: course.ImageURLs,
			CreatedAt: course.CreatedAt,
			UpdatedAt: course.UpdatedAt,
		},
	}

	var err error
	switch course.Type {
	case "BOOK":
		err = s.loadCourseBookData(ctx, course.ID, response)
	case "OTHER":
		err = s.loadCourseOtherData(ctx, course.ID, response)
	default:
		s.logger.Error("buildCourseDetail.unknown_type", map[string]interface{}{
			"type": course.Type,
		}, "Unknown course type")
		return nil, fmt.Errorf("unknown course type: %s", course.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load %s data: %w", course.Type, err)
	}

	// Load lessons for all course types
	if err := s.loadLessonsData(ctx, course.ID, response); err != nil {
		return nil, fmt.Errorf("failed to load lessons data: %w", err)
	}

	return response, nil
}

func (s *CourseService) loadCourseBookData(ctx context.Context, courseID uuid.UUID, response *courseDTO.CourseDetail) error {
	courseBook, err := s.courseBookService.GetByCourseID(ctx, courseID)
	if err != nil {
		// Only return error if it's not a "not found" error
		if !errors.Is(err, courseRepo.ErrCourseNotFound) {
			return fmt.Errorf("failed to get course book: %w", err)
		}
		// If course book not found, just return nil - this is valid for new courses
		return nil
	}

	if courseBook != nil {
		response.CourseBook = &courseDTO.CourseBookResponse{
			ID:              courseBook.ID,
			CourseID:        courseBook.CourseID,
			Publishers:      courseBook.Publishers,
			Authors:         courseBook.Authors,
			PublicationYear: courseBook.PublicationYear,
		}
	}
	return nil
}

func (s *CourseService) loadCourseOtherData(ctx context.Context, courseID uuid.UUID, response *courseDTO.CourseDetail) error {
	courseOther, err := s.courseOtherService.GetByCourseID(ctx, courseID)
	if err != nil {
		// Only return error if it's not a "not found" error
		if !errors.Is(err, courseRepo.ErrCourseNotFound) {
			return fmt.Errorf("failed to get course other: %w", err)
		}
		// If course other not found, just return nil
		return nil
	}

	if courseOther != nil {
		response.CourseOther = &courseDTO.CourseOtherResponse{
			ID:       courseOther.ID,
			CourseID: courseOther.CourseID,
		}
	}
	return nil
}

func (s *CourseService) loadLessonsData(ctx context.Context, courseID uuid.UUID, response *courseDTO.CourseDetail) error {
	lessons, err := s.lessonService.GetByCourseID(ctx, courseID)
	if err != nil {
		return fmt.Errorf("failed to get lessons: %w", err)
	}

	response.Lessons = make([]courseDTO.LessonResponse, len(lessons))
	for i, lesson := range lessons {
		response.Lessons[i] = courseDTO.LessonResponse{
			ID:       lesson.ID,
			CourseID: lesson.CourseID,
			Sequence: lesson.Sequence,
			Title:    lesson.Title,
			Overview: lesson.Overview,
		}

		// Load questions for each lesson
		questions, err := s.lessonQuestionService.GetByLessonID(ctx, lesson.ID)
		if err != nil {
			return fmt.Errorf("failed to get lesson questions: %w", err)
		}

		response.Lessons[i].Questions = make([]courseDTO.LessonQuestionResponse, len(questions))
		for j, question := range questions {
			response.Lessons[i].Questions[j] = courseDTO.LessonQuestionResponse{
				ID:           question.ID,
				LessonID:     question.LessonID,
				Sequence:     question.Sequence,
				QuestionID:   question.QuestionID,
				QuestionType: question.QuestionType,
			}
		}
	}
	return nil
}

func (s *CourseService) DeleteAllCourseData(ctx context.Context) error {
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

	// Delete all records from lesson_questions
	if err := tx.Exec("DELETE FROM lesson_questions").Error; err != nil {
		tx.Rollback()
		s.logger.Error("course_service.delete_all.lesson_questions", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete lesson questions")
		return err
	}

	// Delete all records from lessons
	if err := tx.Exec("DELETE FROM lessons").Error; err != nil {
		tx.Rollback()
		s.logger.Error("course_service.delete_all.lessons", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete lessons")
		return err
	}

	// Delete all records from course_books
	if err := tx.Exec("DELETE FROM course_books").Error; err != nil {
		tx.Rollback()
		s.logger.Error("course_service.delete_all.course_books", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete course books")
		return err
	}

	// Delete all records from course_others
	if err := tx.Exec("DELETE FROM course_others").Error; err != nil {
		tx.Rollback()
		s.logger.Error("course_service.delete_all.course_others", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete course others")
		return err
	}

	// Delete all records from courses
	if err := tx.Exec("DELETE FROM courses").Error; err != nil {
		tx.Rollback()
		s.logger.Error("course_service.delete_all.courses", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete courses")
		return err
	}

	// Delete all Redis cache with pattern course:*
	if err := s.redis.GetCache().DeletePattern(ctx, "course:*"); err != nil {
		tx.Rollback()
		s.logger.Error("course_service.delete_all.cache", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete Redis cache")
		return err
	}

	// Delete OpenSearch index
	if err := s.search.RemoveCoursesIndex(ctx); err != nil {
		tx.Rollback()
		s.logger.Error("course_service.delete_all.search", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete OpenSearch index")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
