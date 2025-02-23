package course

import (
	"context"
	"errors"
	courseDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/course"
	searchClient "fluencybe/internal/app/opensearch"
	redisClient "fluencybe/internal/app/redis"
	courseRepo "fluencybe/internal/app/repository/course"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
)

type CourseUpdator struct {
	logger                *logger.PrettyLogger
	redis                 *redisClient.CourseRedis
	search                *searchClient.CourseSearch
	lessonService         LessonService
	lessonQuestionService LessonQuestionService
	courseBookService     CourseBookService
	courseOtherService    CourseOtherService
}

func NewCourseUpdator(
	log *logger.PrettyLogger,
	cache cache.Cache,
	openSearch *opensearch.Client,
	lessonService LessonService,
	lessonQuestionService LessonQuestionService,
	courseBookService CourseBookService,
	courseOtherService CourseOtherService,
) *CourseUpdator {
	return &CourseUpdator{
		logger:                log,
		redis:                 redisClient.NewCourseRedis(cache, log),
		search:                searchClient.NewCourseSearch(openSearch, log),
		lessonService:         lessonService,
		lessonQuestionService: lessonQuestionService,
		courseBookService:     courseBookService,
		courseOtherService:    courseOtherService,
	}
}

func (u *CourseUpdator) UpdateCacheAndSearch(ctx context.Context, course *course.Course) error {
	courseDetail, err := u.buildCourseDetail(ctx, course)
	if err != nil {
		return fmt.Errorf("failed to build course detail: %w", err)
	}

	if err := u.redis.UpdateCachedCourse(ctx, courseDetail); err != nil {
		u.logger.Error("course_updator.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    course.ID,
		}, "Failed to update cache")
	}

	if err := u.search.UpsertCourse(ctx, courseDetail); err != nil {
		u.logger.Error("course_updator.update_search", map[string]interface{}{
			"error": err.Error(),
			"id":    course.ID,
		}, "Failed to update course in OpenSearch")
	}

	return nil
}

func (u *CourseUpdator) buildCourseDetail(ctx context.Context, course *course.Course) (*courseDTO.CourseDetail, error) {
	response := &courseDTO.CourseDetail{
		CourseResponse: courseDTO.CourseResponse{
			ID:        course.ID,
			Type:      course.Type,
			Title:     course.Title,
			Overview:  course.Overview,
			Skills:    course.Skills,
			Band:      course.Band,
			ImageURLs: course.ImageURLs,
		},
	}

	var err error
	switch course.Type {
	case "BOOK":
		err = u.loadCourseBookData(ctx, course.ID, response)
	case "OTHER":
		err = u.loadCourseOtherData(ctx, course.ID, response)
	default:
		u.logger.Error("buildCourseDetail.unknown_type", map[string]interface{}{
			"type": course.Type,
		}, "Unknown course type")
		return nil, fmt.Errorf("unknown course type: %s", course.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load %s data: %w", course.Type, err)
	}

	// Load lessons for all course types
	if err := u.loadLessonsData(ctx, course.ID, response); err != nil {
		return nil, fmt.Errorf("failed to load lessons data: %w", err)
	}

	return response, nil
}

func (u *CourseUpdator) loadCourseBookData(ctx context.Context, courseID uuid.UUID, response *courseDTO.CourseDetail) error {
	courseBook, err := u.courseBookService.GetByCourseID(ctx, courseID)
	if err != nil {
		// Nếu không tìm thấy course book, coi như hợp lệ và return nil
		if errors.Is(err, courseRepo.ErrCourseNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get course book: %w", err)
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

func (u *CourseUpdator) loadCourseOtherData(ctx context.Context, courseID uuid.UUID, response *courseDTO.CourseDetail) error {
	courseOther, err := u.courseOtherService.GetByCourseID(ctx, courseID)
	if err != nil {
		// Nếu không tìm thấy course other, coi như hợp lệ và return nil
		if errors.Is(err, courseRepo.ErrCourseNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get course other: %w", err)
	}

	if courseOther != nil {
		response.CourseOther = &courseDTO.CourseOtherResponse{
			ID:       courseOther.ID,
			CourseID: courseOther.CourseID,
		}
	}
	return nil
}

func (u *CourseUpdator) loadLessonsData(ctx context.Context, courseID uuid.UUID, response *courseDTO.CourseDetail) error {
	lessons, err := u.lessonService.GetByCourseID(ctx, courseID)
	if err != nil {
		// Nếu không tìm thấy lessons, coi như hợp lệ và return empty array
		if errors.Is(err, courseRepo.ErrLessonNotFound) {
			response.Lessons = []courseDTO.LessonResponse{}
			return nil
		}
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
		questions, err := u.lessonQuestionService.GetByLessonID(ctx, lesson.ID)
		if err != nil {
			// Nếu không tìm thấy questions, coi như hợp lệ và return empty array
			if errors.Is(err, courseRepo.ErrLessonQuestionNotFound) {
				response.Lessons[i].Questions = []courseDTO.LessonQuestionResponse{}
				continue
			}
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
