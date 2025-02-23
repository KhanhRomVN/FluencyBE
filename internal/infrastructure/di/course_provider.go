package di

import (
	courseHandler "fluencybe/internal/app/handler/course"
	courseHelper "fluencybe/internal/app/helper/course"
	searchClient "fluencybe/internal/app/opensearch"
	courseRepo "fluencybe/internal/app/repository/course"
	courseSer "fluencybe/internal/app/service/course"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"

	"github.com/opensearch-project/opensearch-go/v2"
	"gorm.io/gorm"
)

type CourseModule struct {
	CourseHandler         *courseHandler.CourseHandler
	CourseBookHandler     *courseHandler.CourseBookHandler
	CourseOtherHandler    *courseHandler.CourseOtherHandler
	LessonHandler         *courseHandler.LessonHandler
	LessonQuestionHandler *courseHandler.LessonQuestionHandler
}

func ProvideCourseModule(
	gormDB *gorm.DB,
	redisClient *cache.RedisClient,
	openSearchClient *opensearch.Client,
	log *logger.PrettyLogger,
) *CourseModule {
	// Repositories
	courseRepository := courseRepo.NewCourseRepository(gormDB, log)
	courseBookRepository := courseRepo.NewCourseBookRepository(gormDB, log)
	courseOtherRepository := courseRepo.NewCourseOtherRepository(gormDB, log)
	lessonRepository := courseRepo.NewLessonRepository(gormDB, log)
	lessonQuestionRepository := courseRepo.NewLessonQuestionRepository(gormDB, log)

	// Search
	courseSearch := searchClient.NewCourseSearch(openSearchClient, log)

	// Services
	lessonQuestionService := courseSer.NewLessonQuestionService(
		lessonQuestionRepository,
		lessonRepository,
		courseRepository,
		log,
		redisClient,
		nil,
	)

	lessonService := courseSer.NewLessonService(
		lessonRepository,
		courseRepository,
		log,
		redisClient,
		nil,
	)

	courseOtherService := courseSer.NewCourseOtherService(
		courseOtherRepository,
		courseRepository,
		log,
		redisClient,
		nil,
	)

	courseBookService := courseSer.NewCourseBookService(
		courseBookRepository,
		courseRepository,
		log,
		redisClient,
		nil,
	)

	// Course Updator
	courseUpdator := courseHelper.NewCourseUpdator(
		log,
		redisClient,
		openSearchClient,
		lessonService,
		lessonQuestionService,
		courseBookService,
		courseOtherService,
	)

	// Set updator for all services
	lessonQuestionService.SetCourseUpdator(courseUpdator)
	lessonService.SetCourseUpdator(courseUpdator)
	courseOtherService.SetCourseUpdator(courseUpdator)
	courseBookService.SetCourseUpdator(courseUpdator)

	// Main course service
	courseService := courseSer.NewCourseService(
		courseRepository,
		log,
		redisClient,
		courseSearch,
		lessonService,
		lessonQuestionService,
		courseOtherService,
		courseBookService,
		courseUpdator,
	)

	// Handlers
	mainCourseHandler := courseHandler.NewCourseHandler(
		courseService,
		courseBookService,
		courseOtherService,
		lessonService,
		lessonQuestionService,
		log,
	)

	bookHandler := courseHandler.NewCourseBookHandler(
		courseBookService,
		log,
	)

	otherHandler := courseHandler.NewCourseOtherHandler(
		courseOtherService,
		log,
	)

	lesHandler := courseHandler.NewLessonHandler(
		lessonService,
		log,
	)

	lesQuestionHandler := courseHandler.NewLessonQuestionHandler(
		lessonQuestionService,
		log,
	)

	return &CourseModule{
		CourseHandler:         mainCourseHandler,
		CourseBookHandler:     bookHandler,
		CourseOtherHandler:    otherHandler,
		LessonHandler:         lesHandler,
		LessonQuestionHandler: lesQuestionHandler,
	}
}
