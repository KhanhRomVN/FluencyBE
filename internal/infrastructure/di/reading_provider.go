package di

import (
	readingHandler "fluencybe/internal/app/handler/reading"
	readingHelper "fluencybe/internal/app/helper/reading"
	searchClient "fluencybe/internal/app/opensearch"
	readingRepo "fluencybe/internal/app/repository/reading"
	readingSer "fluencybe/internal/app/service/reading"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"

	"github.com/opensearch-project/opensearch-go/v2"
	"gorm.io/gorm"
)

type ReadingModule struct {
	QuestionHandler               *readingHandler.ReadingQuestionHandler
	FillInTheBlankQuestionHandler *readingHandler.ReadingFillInTheBlankQuestionHandler
	FillInTheBlankAnswerHandler   *readingHandler.ReadingFillInTheBlankAnswerHandler
	ChoiceOneQuestionHandler      *readingHandler.ReadingChoiceOneQuestionHandler
	ChoiceOneOptionHandler        *readingHandler.ReadingChoiceOneOptionHandler
	ChoiceMultiQuestionHandler    *readingHandler.ReadingChoiceMultiQuestionHandler
	ChoiceMultiOptionHandler      *readingHandler.ReadingChoiceMultiOptionHandler
	TrueFalseHandler              *readingHandler.ReadingTrueFalseHandler
	MatchingHandler               *readingHandler.ReadingMatchingHandler
}

func ProvideReadingModule(
	gormDB *gorm.DB,
	redisClient *cache.RedisClient,
	openSearchClient *opensearch.Client,
	log *logger.PrettyLogger,
) *ReadingModule {
	// Repositories
	questionRepo := readingRepo.NewReadingQuestionRepository(gormDB, log)
	fillInBlankQuestionRepo := readingRepo.NewReadingFillInTheBlankQuestionRepository(gormDB, log)
	fillInBlankAnswerRepo := readingRepo.NewReadingFillInTheBlankAnswerRepository(gormDB, log)
	choiceOneQuestionRepo := readingRepo.NewReadingChoiceOneQuestionRepository(gormDB, log)
	choiceOneOptionRepo := readingRepo.NewReadingChoiceOneOptionRepository(gormDB, log)
	choiceMultiQuestionRepo := readingRepo.NewReadingChoiceMultiQuestionRepository(gormDB, log)
	choiceMultiOptionRepo := readingRepo.NewReadingChoiceMultiOptionRepository(gormDB, log)
	trueFalseRepo := readingRepo.NewReadingTrueFalseRepository(gormDB, log)
	matchingRepo := readingRepo.NewReadingMatchingRepository(gormDB, log)

	// Search
	questionSearch := searchClient.NewReadingQuestionSearch(openSearchClient, log)

	// Services
	fillInBlankQuestionService := readingSer.NewReadingFillInTheBlankQuestionService(
		fillInBlankQuestionRepo,
		questionRepo,
		log,
		redisClient,
		questionSearch,
		nil,
	)

	fillInBlankAnswerService := readingSer.NewReadingFillInTheBlankAnswerService(
		fillInBlankAnswerRepo,
		fillInBlankQuestionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	choiceOneQuestionService := readingSer.NewReadingChoiceOneQuestionService(
		choiceOneQuestionRepo,
		questionRepo,
		log,
		nil,
	)

	choiceOneOptionService := readingSer.NewReadingChoiceOneOptionService(
		choiceOneOptionRepo,
		choiceOneQuestionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	choiceMultiQuestionService := readingSer.NewReadingChoiceMultiQuestionService(
		choiceMultiQuestionRepo,
		questionRepo,
		log,
		nil,
	)

	choiceMultiOptionService := readingSer.NewReadingChoiceMultiOptionService(
		choiceMultiOptionRepo,
		choiceMultiQuestionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	trueFalseService := readingSer.NewReadingTrueFalseService(
		trueFalseRepo,
		questionRepo,
		log,
		nil,
	)

	matchingService := readingSer.NewReadingMatchingService(
		matchingRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	// Question Updator
	questionUpdator := readingHelper.NewReadingQuestionUpdator(
		log,
		redisClient,
		openSearchClient,
		fillInBlankQuestionService,
		fillInBlankAnswerService,
		choiceOneQuestionService,
		choiceOneOptionService,
		choiceMultiQuestionService,
		choiceMultiOptionService,
		trueFalseService,
		matchingService,
	)

	// Set updator for all services
	fillInBlankQuestionService.SetQuestionUpdator(questionUpdator)
	fillInBlankAnswerService.SetQuestionUpdator(questionUpdator)
	choiceOneQuestionService.SetQuestionUpdator(questionUpdator)
	choiceOneOptionService.SetQuestionUpdator(questionUpdator)
	choiceMultiQuestionService.SetQuestionUpdator(questionUpdator)
	choiceMultiOptionService.SetQuestionUpdator(questionUpdator)
	matchingService.SetQuestionUpdator(questionUpdator)
	trueFalseService.SetQuestionUpdator(questionUpdator)

	// Main question service
	questionService := readingSer.NewReadingQuestionService(
		questionRepo,
		log,
		redisClient,
		openSearchClient,
		fillInBlankQuestionService,
		fillInBlankAnswerService,
		choiceOneQuestionService,
		choiceOneOptionService,
		choiceMultiQuestionService,
		choiceMultiOptionService,
		trueFalseService,
		matchingService,
		questionUpdator,
	)

	// Handlers
	questionHandler := readingHandler.NewReadingQuestionHandler(
		questionService,
		fillInBlankQuestionService,
		fillInBlankAnswerService,
		choiceOneQuestionService,
		choiceOneOptionService,
		choiceMultiQuestionService,
		choiceMultiOptionService,
		trueFalseService,
		matchingService,
		log,
	)

	fillInBlankQuestionHandler := readingHandler.NewReadingFillInTheBlankQuestionHandler(
		fillInBlankQuestionService,
		questionRepo,
		log,
	)

	fillInBlankAnswerHandler := readingHandler.NewReadingFillInTheBlankAnswerHandler(
		fillInBlankAnswerService,
		log,
	)

	choiceOneQuestionHandler := readingHandler.NewReadingChoiceOneQuestionHandler(
		choiceOneQuestionService,
		choiceOneOptionService,
		log,
	)

	choiceOneOptionHandler := readingHandler.NewReadingChoiceOneOptionHandler(
		choiceOneOptionService,
		log,
	)

	choiceMultiQuestionHandler := readingHandler.NewReadingChoiceMultiQuestionHandler(
		choiceMultiQuestionService,
		choiceMultiOptionService,
		log,
	)

	choiceMultiOptionHandler := readingHandler.NewReadingChoiceMultiOptionHandler(
		choiceMultiOptionService,
		log,
	)

	matchingHandler := readingHandler.NewReadingMatchingHandler(
		matchingService,
		log,
	)

	trueFalseHandler := readingHandler.NewReadingTrueFalseHandler(
		trueFalseService,
		log,
	)

	return &ReadingModule{
		QuestionHandler:               questionHandler,
		FillInTheBlankQuestionHandler: fillInBlankQuestionHandler,
		FillInTheBlankAnswerHandler:   fillInBlankAnswerHandler,
		ChoiceOneQuestionHandler:      choiceOneQuestionHandler,
		ChoiceOneOptionHandler:        choiceOneOptionHandler,
		ChoiceMultiQuestionHandler:    choiceMultiQuestionHandler,
		ChoiceMultiOptionHandler:      choiceMultiOptionHandler,
		TrueFalseHandler:              trueFalseHandler,
		MatchingHandler:               matchingHandler,
	}
}
