package di

import (
	"database/sql"
	"fluencybe/internal/core/config"
	"fluencybe/internal/infrastructure/discord"
	"fluencybe/internal/infrastructure/metrics"
	"fluencybe/internal/infrastructure/router"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/opensearch-project/opensearch-go/v2"
	"gorm.io/gorm"
)

// Container holds all the dependencies for the application
type Container struct {
	Config     *config.Config
	Logger     *logger.PrettyLogger
	Router     *gin.Engine
	DBConn     *sql.DB
	GormDB     *gorm.DB
	Redis      *cache.RedisClient
	OpenSearch *opensearch.Client
	DiscordBot *discord.Bot
	Metrics    *metrics.Metrics

	// Feature Modules
	Account   *AccountModule
	Grammar   *GrammarModule
	Listening *ListeningModule
	Reading   *ReadingModule
	Speaking  *SpeakingModule
	Writing   *WritingModule
	Course    *CourseModule
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config) (*Container, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Initialize logger first
	log := logger.InitGlobalLogger(logger.GlobalLoggerOptions{
		Level:    logger.LevelInfo,
		Output:   os.Stdout,
		Service:  "FluencyBE",
		Colorful: true,
	})

	container := &Container{
		Config: cfg,
		Logger: log,
	}

	var err error

	// Initialize infrastructure components
	container.DBConn, err = initializeDB(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	container.GormDB, err = initializeGormDB(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	container.Redis, err = initializeRedis(cfg, log)
	if err != nil {
		log.Critical("REDIS_INIT_CRITICAL", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to initialize Redis, continuing without caching")
	}

	container.OpenSearch, err = initializeOpenSearch(cfg, log)
	if err != nil {
		log.Critical("OPENSEARCH_INIT_CRITICAL", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to initialize OpenSearch, continuing without search functionality")
	}

	container.DiscordBot, err = initializeDiscordBot(log)
	if err != nil {
		log.Critical("DISCORD_BOT_INIT", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to initialize Discord bot")
	}

	container.Metrics = initializeMetrics(log, container.Redis)

	// Initialize feature modules
	container.Account = ProvideAccountModule(container.DBConn, container.Redis, log)
	container.Grammar = ProvideGrammarModule(container.GormDB, container.Redis, container.OpenSearch, log)
	container.Listening = ProvideListeningModule(container.GormDB, container.Redis, container.OpenSearch, log)
	container.Reading = ProvideReadingModule(container.GormDB, container.Redis, container.OpenSearch, log)
	container.Speaking = ProvideSpeakingModule(container.GormDB, container.Redis, container.OpenSearch, log)
	container.Writing = ProvideWritingModule(container.GormDB, container.Redis, container.OpenSearch, log)
	container.Course = ProvideCourseModule(container.GormDB, container.Redis, container.OpenSearch, log)

	// Initialize router with all handlers
	r := router.NewRouter(container.DBConn)
	r.SetupRoutes(
		// Account handlers
		container.Account.UserHandler,
		container.Account.DeveloperHandler,

		// Grammar handlers
		container.Grammar.QuestionHandler,
		container.Grammar.FillInTheBlankQuestionHandler,
		container.Grammar.FillInTheBlankAnswerHandler,
		container.Grammar.ChoiceOneQuestionHandler,
		container.Grammar.ChoiceOneOptionHandler,
		container.Grammar.ErrorIdentificationHandler,
		container.Grammar.SentenceTransformationHandler,

		// Listening handlers
		container.Listening.QuestionHandler,
		container.Listening.FillInTheBlankQuestionHandler,
		container.Listening.FillInTheBlankAnswerHandler,
		container.Listening.ChoiceOneQuestionHandler,
		container.Listening.ChoiceOneOptionHandler,
		container.Listening.ChoiceMultiQuestionHandler,
		container.Listening.ChoiceMultiOptionHandler,
		container.Listening.MapLabellingHandler,
		container.Listening.MatchingHandler,

		// Reading handlers
		container.Reading.QuestionHandler,
		container.Reading.FillInTheBlankQuestionHandler,
		container.Reading.FillInTheBlankAnswerHandler,
		container.Reading.ChoiceOneQuestionHandler,
		container.Reading.ChoiceOneOptionHandler,
		container.Reading.ChoiceMultiQuestionHandler,
		container.Reading.ChoiceMultiOptionHandler,
		container.Reading.TrueFalseHandler,
		container.Reading.MatchingHandler,

		// Speaking handlers
		container.Speaking.QuestionHandler,
		container.Speaking.WordRepetitionHandler,
		container.Speaking.PhraseRepetitionHandler,
		container.Speaking.ParagraphRepetitionHandler,
		container.Speaking.OpenParagraphHandler,
		container.Speaking.ConversationalRepetitionHandler,
		container.Speaking.ConversationalRepetitionQAHandler,
		container.Speaking.ConversationalOpenHandler,

		// Writing handlers
		container.Writing.QuestionHandler,
		container.Writing.SentenceCompletionHandler,
		container.Writing.EssayHandler,

		// Course handlers
		container.Course.CourseHandler,
		container.Course.CourseBookHandler,
		container.Course.CourseOtherHandler,
		container.Course.LessonHandler,
		container.Course.LessonQuestionHandler,
	)

	container.Router = r.Engine

	// Initialize health check
	initializeHealthCheck(container.Redis, container.OpenSearch, log)

	return container, nil
}
