package di

import (
	speakingHandler "fluencybe/internal/app/handler/speaking"
	speakingHelper "fluencybe/internal/app/helper/speaking"
	searchClient "fluencybe/internal/app/opensearch"
	speakingRepo "fluencybe/internal/app/repository/speaking"
	speakingSer "fluencybe/internal/app/service/speaking"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"

	"github.com/opensearch-project/opensearch-go/v2"
	"gorm.io/gorm"
)

type SpeakingModule struct {
	QuestionHandler                   *speakingHandler.SpeakingQuestionHandler
	WordRepetitionHandler             *speakingHandler.SpeakingWordRepetitionHandler
	PhraseRepetitionHandler           *speakingHandler.SpeakingPhraseRepetitionHandler
	ParagraphRepetitionHandler        *speakingHandler.SpeakingParagraphRepetitionHandler
	OpenParagraphHandler              *speakingHandler.SpeakingOpenParagraphHandler
	ConversationalRepetitionHandler   *speakingHandler.SpeakingConversationalRepetitionHandler
	ConversationalRepetitionQAHandler *speakingHandler.SpeakingConversationalRepetitionQAHandler
	ConversationalOpenHandler         *speakingHandler.SpeakingConversationalOpenHandler
}

func ProvideSpeakingModule(
	gormDB *gorm.DB,
	redisClient *cache.RedisClient,
	openSearchClient *opensearch.Client,
	log *logger.PrettyLogger,
) *SpeakingModule {
	// Repositories
	questionRepo := speakingRepo.NewSpeakingQuestionRepository(gormDB, log)
	wordRepetitionRepo := speakingRepo.NewSpeakingWordRepetitionRepository(gormDB, log)
	phraseRepetitionRepo := speakingRepo.NewSpeakingPhraseRepetitionRepository(gormDB, log)
	paragraphRepetitionRepo := speakingRepo.NewSpeakingParagraphRepetitionRepository(gormDB, log)
	openParagraphRepo := speakingRepo.NewSpeakingOpenParagraphRepository(gormDB, log)
	conversationalRepetitionRepo := speakingRepo.NewSpeakingConversationalRepetitionRepository(gormDB, log)
	conversationalRepetitionQARepo := speakingRepo.NewSpeakingConversationalRepetitionQARepository(gormDB, log)
	conversationalOpenRepo := speakingRepo.NewSpeakingConversationalOpenRepository(gormDB, log)

	// Search
	questionSearch := searchClient.NewSpeakingQuestionSearch(openSearchClient, log)

	// Services
	wordRepetitionService := speakingSer.NewSpeakingWordRepetitionService(
		wordRepetitionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	phraseRepetitionService := speakingSer.NewSpeakingPhraseRepetitionService(
		phraseRepetitionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	paragraphRepetitionService := speakingSer.NewSpeakingParagraphRepetitionService(
		paragraphRepetitionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	openParagraphService := speakingSer.NewSpeakingOpenParagraphService(
		openParagraphRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	conversationalRepetitionService := speakingSer.NewSpeakingConversationalRepetitionService(
		conversationalRepetitionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	conversationalRepetitionQAService := speakingSer.NewSpeakingConversationalRepetitionQAService(
		conversationalRepetitionQARepo,
		conversationalRepetitionRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	conversationalOpenService := speakingSer.NewSpeakingConversationalOpenService(
		conversationalOpenRepo,
		questionRepo,
		log,
		redisClient,
		nil,
	)

	// Question Updator
	questionUpdator := speakingHelper.NewSpeakingQuestionUpdator(
		log,
		redisClient,
		openSearchClient,
		wordRepetitionService,
		phraseRepetitionService,
		paragraphRepetitionService,
		openParagraphService,
		conversationalRepetitionService,
		conversationalRepetitionQAService,
		conversationalOpenService,
	)

	// Set updator for all services
	wordRepetitionService.SetQuestionUpdator(questionUpdator)
	phraseRepetitionService.SetQuestionUpdator(questionUpdator)
	paragraphRepetitionService.SetQuestionUpdator(questionUpdator)
	openParagraphService.SetQuestionUpdator(questionUpdator)
	conversationalRepetitionService.SetQuestionUpdator(questionUpdator)
	conversationalRepetitionQAService.SetQuestionUpdator(questionUpdator)
	conversationalOpenService.SetQuestionUpdator(questionUpdator)

	// Main question service
	questionService := speakingSer.NewSpeakingQuestionService(
		questionRepo,
		log,
		redisClient,
		questionSearch,
		wordRepetitionService,
		phraseRepetitionService,
		paragraphRepetitionService,
		openParagraphService,
		conversationalRepetitionService,
		conversationalRepetitionQAService,
		conversationalOpenService,
		questionUpdator,
	)

	// Handlers
	questionHandler := speakingHandler.NewSpeakingQuestionHandler(
		questionService,
		wordRepetitionService,
		phraseRepetitionService,
		paragraphRepetitionService,
		openParagraphService,
		conversationalRepetitionService,
		conversationalRepetitionQAService,
		conversationalOpenService,
		log,
	)

	wordRepetitionHandler := speakingHandler.NewSpeakingWordRepetitionHandler(
		wordRepetitionService,
		log,
	)

	phraseRepetitionHandler := speakingHandler.NewSpeakingPhraseRepetitionHandler(
		phraseRepetitionService,
		log,
	)

	paragraphRepetitionHandler := speakingHandler.NewSpeakingParagraphRepetitionHandler(
		paragraphRepetitionService,
		log,
	)

	openParagraphHandler := speakingHandler.NewSpeakingOpenParagraphHandler(
		openParagraphService,
		log,
	)

	conversationalRepetitionHandler := speakingHandler.NewSpeakingConversationalRepetitionHandler(
		conversationalRepetitionService,
		log,
	)

	conversationalRepetitionQAHandler := speakingHandler.NewSpeakingConversationalRepetitionQAHandler(
		conversationalRepetitionQAService,
		log,
	)

	conversationalOpenHandler := speakingHandler.NewSpeakingConversationalOpenHandler(
		conversationalOpenService,
		log,
	)

	return &SpeakingModule{
		QuestionHandler:                   questionHandler,
		WordRepetitionHandler:             wordRepetitionHandler,
		PhraseRepetitionHandler:           phraseRepetitionHandler,
		ParagraphRepetitionHandler:        paragraphRepetitionHandler,
		OpenParagraphHandler:              openParagraphHandler,
		ConversationalRepetitionHandler:   conversationalRepetitionHandler,
		ConversationalRepetitionQAHandler: conversationalRepetitionQAHandler,
		ConversationalOpenHandler:         conversationalOpenHandler,
	}
}
