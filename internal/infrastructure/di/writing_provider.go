package di

import (
	writingHandler "fluencybe/internal/app/handler/writing"
	writingHelper "fluencybe/internal/app/helper/writing"
	searchClient "fluencybe/internal/app/opensearch"
	writingRepo "fluencybe/internal/app/repository/writing"
	writingSer "fluencybe/internal/app/service/writing"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"

	"github.com/opensearch-project/opensearch-go/v2"
	"gorm.io/gorm"
)

type WritingModule struct {
	QuestionHandler           *writingHandler.WritingQuestionHandler
	SentenceCompletionHandler *writingHandler.WritingSentenceCompletionHandler
	EssayHandler              *writingHandler.WritingEssayHandler
}

func ProvideWritingModule(
	gormDB *gorm.DB,
	redisClient *cache.RedisClient,
	openSearchClient *opensearch.Client,
	log *logger.PrettyLogger,
) *WritingModule {
	// Repositories
	questionRepo := writingRepo.NewWritingQuestionRepository(gormDB, log)
	essayRepo := writingRepo.NewWritingEssayRepository(gormDB, log)
	sentenceCompletionRepo := writingRepo.NewWritingSentenceCompletionRepository(gormDB, log)

	// Search
	questionSearch := searchClient.NewWritingQuestionSearch(openSearchClient, log)

	// Services
	essayService := writingSer.NewWritingEssayService(
		essayRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	sentenceCompletionService := writingSer.NewWritingSentenceCompletionService(
		sentenceCompletionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	// Question Updator
	questionUpdator := writingHelper.NewWritingQuestionUpdator(
		log,
		redisClient,
		openSearchClient,
		sentenceCompletionService,
		essayService,
	)

	// Set updator for all services
	essayService.SetQuestionUpdator(questionUpdator)
	sentenceCompletionService.SetQuestionUpdator(questionUpdator)

	// Main question service
	questionService := writingSer.NewWritingQuestionService(
		questionRepo,
		log,
		redisClient,
		questionSearch,
		sentenceCompletionService,
		essayService,
		questionUpdator,
	)

	// Handlers
	questionHandler := writingHandler.NewWritingQuestionHandler(
		questionService,
		sentenceCompletionService,
		essayService,
		log,
	)

	sentenceCompletionHandler := writingHandler.NewWritingSentenceCompletionHandler(
		sentenceCompletionService,
		log,
	)

	essayHandler := writingHandler.NewWritingEssayHandler(
		essayService,
		log,
	)

	return &WritingModule{
		QuestionHandler:           questionHandler,
		SentenceCompletionHandler: sentenceCompletionHandler,
		EssayHandler:              essayHandler,
	}
}
