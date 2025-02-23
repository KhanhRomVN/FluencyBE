package di

import (
	grammarHandler "fluencybe/internal/app/handler/grammar"
	grammarHelper "fluencybe/internal/app/helper/grammar"
	searchClient "fluencybe/internal/app/opensearch"
	grammarRepo "fluencybe/internal/app/repository/grammar"
	grammarSer "fluencybe/internal/app/service/grammar"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"

	"github.com/opensearch-project/opensearch-go/v2"
	"gorm.io/gorm"
)

type GrammarModule struct {
	QuestionHandler               *grammarHandler.GrammarQuestionHandler
	FillInTheBlankQuestionHandler *grammarHandler.GrammarFillInTheBlankQuestionHandler
	FillInTheBlankAnswerHandler   *grammarHandler.GrammarFillInTheBlankAnswerHandler
	ChoiceOneQuestionHandler      *grammarHandler.GrammarChoiceOneQuestionHandler
	ChoiceOneOptionHandler        *grammarHandler.GrammarChoiceOneOptionHandler
	ErrorIdentificationHandler    *grammarHandler.GrammarErrorIdentificationHandler
	SentenceTransformationHandler *grammarHandler.GrammarSentenceTransformationHandler
}

func ProvideGrammarModule(
	gormDB *gorm.DB,
	redisClient *cache.RedisClient,
	openSearchClient *opensearch.Client,
	log *logger.PrettyLogger,
) *GrammarModule {
	// Repositories
	questionRepo := grammarRepo.NewGrammarQuestionRepository(gormDB, log)
	fillInBlankQuestionRepo := grammarRepo.NewGrammarFillInTheBlankQuestionRepository(gormDB, log)
	fillInBlankAnswerRepo := grammarRepo.NewGrammarFillInTheBlankAnswerRepository(gormDB, log)
	choiceOneQuestionRepo := grammarRepo.NewGrammarChoiceOneQuestionRepository(gormDB, log)
	choiceOneOptionRepo := grammarRepo.NewGrammarChoiceOneOptionRepository(gormDB, log)
	errorIdentificationRepo := grammarRepo.NewGrammarErrorIdentificationRepository(gormDB, log)
	sentenceTransformationRepo := grammarRepo.NewGrammarSentenceTransformationRepository(gormDB, log)

	// Search
	questionSearch := searchClient.NewGrammarQuestionSearch(openSearchClient, log)

	// Services
	fillInBlankQuestionService := grammarSer.NewGrammarFillInTheBlankQuestionService(
		fillInBlankQuestionRepo,
		questionRepo,
		log,
		redisClient,
		questionSearch,
		nil,
	)

	fillInBlankAnswerService := grammarSer.NewGrammarFillInTheBlankAnswerService(
		fillInBlankAnswerRepo,
		fillInBlankQuestionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	choiceOneQuestionService := grammarSer.NewGrammarChoiceOneQuestionService(
		choiceOneQuestionRepo,
		questionRepo,
		log,
		nil,
	)

	choiceOneOptionService := grammarSer.NewGrammarChoiceOneOptionService(
		choiceOneOptionRepo,
		choiceOneQuestionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	errorIdentificationService := grammarSer.NewGrammarErrorIdentificationService(
		errorIdentificationRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	sentenceTransformationService := grammarSer.NewGrammarSentenceTransformationService(
		sentenceTransformationRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	// Question Updator
	questionUpdator := grammarHelper.NewGrammarQuestionUpdator(
		log,
		redisClient,
		openSearchClient,
		fillInBlankQuestionService,
		fillInBlankAnswerService,
		choiceOneQuestionService,
		choiceOneOptionService,
		errorIdentificationService,
		sentenceTransformationService,
	)

	// Set updator for all services
	fillInBlankQuestionService.SetQuestionUpdator(questionUpdator)
	fillInBlankAnswerService.SetQuestionUpdator(questionUpdator)
	choiceOneQuestionService.SetQuestionUpdator(questionUpdator)
	choiceOneOptionService.SetQuestionUpdator(questionUpdator)
	errorIdentificationService.SetQuestionUpdator(questionUpdator)
	sentenceTransformationService.SetQuestionUpdator(questionUpdator)

	// Main question service
	questionService := grammarSer.NewGrammarQuestionService(
		questionRepo,
		log,
		redisClient,
		openSearchClient,
		fillInBlankQuestionService,
		fillInBlankAnswerService,
		choiceOneQuestionService,
		choiceOneOptionService,
		errorIdentificationService,
		sentenceTransformationService,
		questionUpdator,
	)

	// Handlers
	questionHandler := grammarHandler.NewGrammarQuestionHandler(
		questionService,
		fillInBlankQuestionService,
		fillInBlankAnswerService,
		choiceOneQuestionService,
		choiceOneOptionService,
		errorIdentificationService,
		sentenceTransformationService,
		log,
	)

	fillInBlankQuestionHandler := grammarHandler.NewGrammarFillInTheBlankQuestionHandler(
		fillInBlankQuestionService,
		questionRepo,
		log,
	)

	fillInBlankAnswerHandler := grammarHandler.NewGrammarFillInTheBlankAnswerHandler(
		fillInBlankAnswerService,
		log,
	)

	choiceOneQuestionHandler := grammarHandler.NewGrammarChoiceOneQuestionHandler(
		choiceOneQuestionService,
		choiceOneOptionService,
		log,
	)

	choiceOneOptionHandler := grammarHandler.NewGrammarChoiceOneOptionHandler(
		choiceOneOptionService,
		log,
	)

	errorIdentificationHandler := grammarHandler.NewGrammarErrorIdentificationHandler(
		errorIdentificationService,
		log,
	)

	sentenceTransformationHandler := grammarHandler.NewGrammarSentenceTransformationHandler(
		sentenceTransformationService,
		log,
	)

	return &GrammarModule{
		QuestionHandler:               questionHandler,
		FillInTheBlankQuestionHandler: fillInBlankQuestionHandler,
		FillInTheBlankAnswerHandler:   fillInBlankAnswerHandler,
		ChoiceOneQuestionHandler:      choiceOneQuestionHandler,
		ChoiceOneOptionHandler:        choiceOneOptionHandler,
		ErrorIdentificationHandler:    errorIdentificationHandler,
		SentenceTransformationHandler: sentenceTransformationHandler,
	}
}
