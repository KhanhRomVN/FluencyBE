package di

import (
	listeningHandler "fluencybe/internal/app/handler/listening"
	listeningHelper "fluencybe/internal/app/helper/listening"
	searchClient "fluencybe/internal/app/opensearch"
	listeningRepo "fluencybe/internal/app/repository/listening"
	listeningSer "fluencybe/internal/app/service/listening"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"

	"github.com/opensearch-project/opensearch-go/v2"
	"gorm.io/gorm"
)

type ListeningModule struct {
	QuestionHandler               *listeningHandler.ListeningQuestionHandler
	FillInTheBlankQuestionHandler *listeningHandler.ListeningFillInTheBlankQuestionHandler
	FillInTheBlankAnswerHandler   *listeningHandler.ListeningFillInTheBlankAnswerHandler
	ChoiceOneQuestionHandler      *listeningHandler.ListeningChoiceOneQuestionHandler
	ChoiceOneOptionHandler        *listeningHandler.ListeningChoiceOneOptionHandler
	ChoiceMultiQuestionHandler    *listeningHandler.ListeningChoiceMultiQuestionHandler
	ChoiceMultiOptionHandler      *listeningHandler.ListeningChoiceMultiOptionHandler
	MapLabellingHandler           *listeningHandler.ListeningMapLabellingHandler
	MatchingHandler               *listeningHandler.ListeningMatchingHandler
}

func ProvideListeningModule(
	gormDB *gorm.DB,
	redisClient *cache.RedisClient,
	openSearchClient *opensearch.Client,
	log *logger.PrettyLogger,
) *ListeningModule {
	// Repositories
	questionRepo := listeningRepo.NewListeningQuestionRepository(gormDB, log)
	fillInBlankQuestionRepo := listeningRepo.NewListeningFillInTheBlankQuestionRepository(gormDB, log)
	fillInBlankAnswerRepo := listeningRepo.NewListeningFillInTheBlankAnswerRepository(gormDB, log)
	choiceOneQuestionRepo := listeningRepo.NewListeningChoiceOneQuestionRepository(gormDB, log)
	choiceOneOptionRepo := listeningRepo.NewListeningChoiceOneOptionRepository(gormDB, log)
	choiceMultiQuestionRepo := listeningRepo.NewListeningChoiceMultiQuestionRepository(gormDB, log)
	choiceMultiOptionRepo := listeningRepo.NewListeningChoiceMultiOptionRepository(gormDB, log)
	mapLabellingRepo := listeningRepo.NewListeningMapLabellingRepository(gormDB, log)
	matchingRepo := listeningRepo.NewListeningMatchingRepository(gormDB, log)

	// Search
	questionSearch := searchClient.NewListeningQuestionSearch(openSearchClient, log)

	// Services
	fillInBlankQuestionService := listeningSer.NewListeningFillInTheBlankQuestionService(
		fillInBlankQuestionRepo,
		questionRepo,
		log,
		redisClient,
		questionSearch,
		nil,
	)

	fillInBlankAnswerService := listeningSer.NewListeningFillInTheBlankAnswerService(
		fillInBlankAnswerRepo,
		fillInBlankQuestionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	choiceOneQuestionService := listeningSer.NewListeningChoiceOneQuestionService(
		choiceOneQuestionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	choiceOneOptionService := listeningSer.NewListeningChoiceOneOptionService(
		choiceOneOptionRepo,
		choiceOneQuestionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	choiceMultiQuestionService := listeningSer.NewListeningChoiceMultiQuestionService(
		choiceMultiQuestionRepo,
		questionRepo,
		log,
		nil,
	)

	choiceMultiOptionService := listeningSer.NewListeningChoiceMultiOptionService(
		choiceMultiOptionRepo,
		choiceMultiQuestionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	mapLabellingService := listeningSer.NewListeningMapLabellingService(
		mapLabellingRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	matchingService := listeningSer.NewListeningMatchingService(
		matchingRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	// Question Updator
	questionUpdator := listeningHelper.NewListeningQuestionUpdator(
		log,
		redisClient,
		openSearchClient,
		fillInBlankQuestionService,
		fillInBlankAnswerService,
		choiceOneQuestionService,
		choiceOneOptionService,
		choiceMultiQuestionService,
		choiceMultiOptionService,
		mapLabellingService,
		matchingService,
	)

	// Set updator for all services
	fillInBlankQuestionService.SetQuestionUpdator(questionUpdator)
	fillInBlankAnswerService.SetQuestionUpdator(questionUpdator)
	choiceOneQuestionService.SetQuestionUpdator(questionUpdator)
	choiceOneOptionService.SetQuestionUpdator(questionUpdator)
	choiceMultiQuestionService.SetQuestionUpdator(questionUpdator)
	choiceMultiOptionService.SetQuestionUpdator(questionUpdator)
	mapLabellingService.SetQuestionUpdator(questionUpdator)
	matchingService.SetQuestionUpdator(questionUpdator)

	// Main question service
	questionService := listeningSer.NewListeningQuestionService(
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
		mapLabellingService,
		matchingService,
		questionUpdator,
	)

	// Handlers
	questionHandler := listeningHandler.NewListeningQuestionHandler(
		questionService,
		fillInBlankQuestionService,
		fillInBlankAnswerService,
		choiceOneQuestionService,
		choiceOneOptionService,
		choiceMultiQuestionService,
		choiceMultiOptionService,
		mapLabellingService,
		matchingService,
		log,
	)

	fillInBlankQuestionHandler := listeningHandler.NewListeningFillInTheBlankQuestionHandler(
		fillInBlankQuestionService,
		questionRepo,
		log,
	)

	fillInBlankAnswerHandler := listeningHandler.NewListeningFillInTheBlankAnswerHandler(
		fillInBlankAnswerService,
		log,
	)

	choiceOneQuestionHandler := listeningHandler.NewListeningChoiceOneQuestionHandler(
		choiceOneQuestionService,
		choiceOneOptionService,
		log,
	)

	choiceOneOptionHandler := listeningHandler.NewListeningChoiceOneOptionHandler(
		choiceOneOptionService,
		log,
	)

	choiceMultiQuestionHandler := listeningHandler.NewListeningChoiceMultiQuestionHandler(
		choiceMultiQuestionService,
		choiceMultiOptionService,
		log,
	)

	choiceMultiOptionHandler := listeningHandler.NewListeningChoiceMultiOptionHandler(
		choiceMultiOptionService,
		log,
	)

	mapLabellingHandler := listeningHandler.NewListeningMapLabellingHandler(
		mapLabellingService,
		log,
	)

	matchingHandler := listeningHandler.NewListeningMatchingHandler(
		matchingService,
		log,
	)

	return &ListeningModule{
		QuestionHandler:               questionHandler,
		FillInTheBlankQuestionHandler: fillInBlankQuestionHandler,
		FillInTheBlankAnswerHandler:   fillInBlankAnswerHandler,
		ChoiceOneQuestionHandler:      choiceOneQuestionHandler,
		ChoiceOneOptionHandler:        choiceOneOptionHandler,
		ChoiceMultiQuestionHandler:    choiceMultiQuestionHandler,
		ChoiceMultiOptionHandler:      choiceMultiOptionHandler,
		MapLabellingHandler:           mapLabellingHandler,
		MatchingHandler:               matchingHandler,
	}
}
