package bootstrap

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"

	constants "fluencybe/internal/core/constants"

	accountHa "fluencybe/internal/app/handler/account"
	accountRepo "fluencybe/internal/app/repository/account"
	accountSer "fluencybe/internal/app/service/account"

	listeningHa "fluencybe/internal/app/handler/listening"
	listeningHelper "fluencybe/internal/app/helper/listening"
	listeningModel "fluencybe/internal/app/model/listening"
	listeningRepo "fluencybe/internal/app/repository/listening"
	listeningSer "fluencybe/internal/app/service/listening"

	readingHa "fluencybe/internal/app/handler/reading"
	readingHelper "fluencybe/internal/app/helper/reading"
	readingModel "fluencybe/internal/app/model/reading"
	readingRepo "fluencybe/internal/app/repository/reading"
	readingSer "fluencybe/internal/app/service/reading"

	speakingHa "fluencybe/internal/app/handler/speaking"
	speakingHelper "fluencybe/internal/app/helper/speaking"
	speakingModel "fluencybe/internal/app/model/speaking"
	speakingRepo "fluencybe/internal/app/repository/speaking"
	speakingSer "fluencybe/internal/app/service/speaking"

	grammarHa "fluencybe/internal/app/handler/grammar"
	grammarHelper "fluencybe/internal/app/helper/grammar"
	grammarModel "fluencybe/internal/app/model/grammar"
	grammarRepo "fluencybe/internal/app/repository/grammar"
	grammarSer "fluencybe/internal/app/service/grammar"

	writingHa "fluencybe/internal/app/handler/writing"
	writingHelper "fluencybe/internal/app/helper/writing"
	writingRepo "fluencybe/internal/app/repository/writing"
	writingSer "fluencybe/internal/app/service/writing"

	courseHa "fluencybe/internal/app/handler/course"
	courseHelper "fluencybe/internal/app/helper/course"
	courseModel "fluencybe/internal/app/model/course"
	courseRepo "fluencybe/internal/app/repository/course"
	courseSer "fluencybe/internal/app/service/course"

	searchClient "fluencybe/internal/app/opensearch"
	redis "fluencybe/pkg/cache"

	"fluencybe/internal/core/config"
	"fluencybe/internal/core/status"
	"fluencybe/internal/infrastructure/discord"
	"fluencybe/internal/infrastructure/metrics"
	"fluencybe/internal/infrastructure/router"
	"fluencybe/pkg/db"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/search"

	"github.com/gin-gonic/gin"
	"github.com/opensearch-project/opensearch-go/v2"
	"gorm.io/gorm"
)

func maskString(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "****" + value[len(value)-2:]
}

type Application struct {
	Config     *config.Config
	Logger     *logger.PrettyLogger
	Router     *gin.Engine
	Server     *http.Server
	DBConn     *sql.DB
	GormDB     *gorm.DB
	Redis      *redis.RedisClient
	OpenSearch *opensearch.Client
	DiscordBot *discord.Bot
	Metrics    *metrics.Metrics
}

func NewApplication(cfg *config.Config) (*Application, error) {
	// ! ------------------------------------------------------------------------------
	// ! - Logger
	// ! ------------------------------------------------------------------------------
	logLevels := strings.ToUpper(os.Getenv("LOG_LEVELS"))
	enabledLevels := make(map[logger.LogLevel]bool)

	if logLevels == "" {
		logLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))
		if logLevel != "" {
			switch logLevel {
			case "DEBUG":
				enabledLevels[logger.LevelDebug] = true
			case "INFO":
				enabledLevels[logger.LevelInfo] = true
			case "WARNING":
				enabledLevels[logger.LevelWarning] = true
			case "ERROR":
				enabledLevels[logger.LevelError] = true
			case "CRITICAL":
				enabledLevels[logger.LevelCritical] = true
			default:
				enabledLevels[logger.LevelInfo] = true
			}
		} else {
			enabledLevels[logger.LevelInfo] = true
		}
	} else {
		levels := strings.Split(logLevels, ",")
		for _, level := range levels {
			level = strings.TrimSpace(level)
			switch level {
			case "DEBUG":
				enabledLevels[logger.LevelDebug] = true
			case "INFO":
				enabledLevels[logger.LevelInfo] = true
			case "WARNING":
				enabledLevels[logger.LevelWarning] = true
			case "ERROR":
				enabledLevels[logger.LevelError] = true
			case "CRITICAL":
				enabledLevels[logger.LevelCritical] = true
			}
		}

		if len(enabledLevels) == 0 {
			enabledLevels[logger.LevelInfo] = true
		}
	}

	log := logger.InitGlobalLogger(logger.GlobalLoggerOptions{
		Level:    logger.LevelInfo,
		Service:  "FluencyBE",
		Colorful: true,
	})

	// ! ------------------------------------------------------------------------------
	// ! - Database
	// ! ------------------------------------------------------------------------------
	dbConn, err := db.NewDBConnection(cfg.DBConfig, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	gormDB, err := db.NewGormDBConnection(cfg.DBConfig, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	if err := gormDB.AutoMigrate(
		// Grammar
		&grammarModel.GrammarQuestion{},
		&grammarModel.GrammarFillInTheBlankQuestion{},
		&grammarModel.GrammarFillInTheBlankAnswer{},
		&grammarModel.GrammarChoiceOneQuestion{},
		&grammarModel.GrammarChoiceOneOption{},
		&grammarModel.GrammarErrorIdentification{},
		&grammarModel.GrammarSentenceTransformation{},
		// Listening
		&listeningModel.ListeningQuestion{},
		&listeningModel.ListeningFillInTheBlankQuestion{},
		&listeningModel.ListeningFillInTheBlankAnswer{},
		&listeningModel.ListeningChoiceOneQuestion{},
		&listeningModel.ListeningChoiceOneOption{},
		&listeningModel.ListeningChoiceMultiQuestion{},
		&listeningModel.ListeningChoiceMultiOption{},
		&listeningModel.ListeningMapLabelling{},
		&listeningModel.ListeningMatching{},
		// Reading
		&readingModel.ReadingQuestion{},
		&readingModel.ReadingFillInTheBlankQuestion{},
		&readingModel.ReadingFillInTheBlankAnswer{},
		&readingModel.ReadingChoiceOneQuestion{},
		&readingModel.ReadingChoiceOneOption{},
		&readingModel.ReadingChoiceMultiQuestion{},
		&readingModel.ReadingChoiceMultiOption{},
		&readingModel.ReadingMatching{},
		&readingModel.ReadingTrueFalse{},
		// Speaking
		&speakingModel.SpeakingQuestion{},
		&speakingModel.SpeakingWordRepetition{},
		&speakingModel.SpeakingPhraseRepetition{},
		&speakingModel.SpeakingParagraphRepetition{},
		&speakingModel.SpeakingOpenParagraph{},
		&speakingModel.SpeakingConversationalRepetition{},
		&speakingModel.SpeakingConversationalRepetitionQA{},
		&speakingModel.SpeakingConversationalOpen{},
		// Course
		&courseModel.Course{},
		&courseModel.CourseBook{},
		&courseModel.CourseOther{},
		&courseModel.Lesson{},
		&courseModel.LessonQuestion{},
	); err != nil {
		log.Critical("GORM_AUTOMIGRATE", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to auto-migrate database schema")
	}

	// ! ------------------------------------------------------------------------------
	// ! - Redis - Opensearch - Discord - Metric
	// ! ------------------------------------------------------------------------------
	redisClient, err := redis.NewRedisClient(cfg.RedisConfig, log)
	if err != nil {
		maskedHost := maskString(cfg.RedisConfig.Host)
		maskedPort := maskString(cfg.RedisConfig.Port)

		log.Critical("REDIS_INIT_CRITICAL", map[string]interface{}{
			"error": err.Error(),
			"host":  maskedHost,
			"port":  maskedPort,
		}, "Failed to initialize Redis, continuing without caching")
	}

	openSearchClient, err := search.NewOpenSearchClient(cfg.OpenSearchConfig, log)
	if err != nil {
		log.Critical("OPENSEARCH_INIT_CRITICAL", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to initialize OpenSearch, continuing without search functionality")
	}

	discordBot, err := discord.NewBot(log)
	if err != nil {
		log.Critical("DISCORD_BOT_INIT", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to initialize Discord bot")
	} else {
		if err := discordBot.Start(); err != nil {
			log.Critical("DISCORD_BOT_START", map[string]interface{}{
				"error": err.Error(),
			}, "Failed to start Discord bot")
		}
	}

	metricsCollector := metrics.NewMetrics(log, redisClient)
	metricsCollector.StartMetricsCollection()

	status.InitConnectionStatus(log)

	status.StartHealthCheck(redisClient, openSearchClient)

	// ! ------------------------------------------------------------------------------
	// ! - Repository
	// ! ------------------------------------------------------------------------------
	// ? ------------------------------------------------------------------------------
	// ? - Repository - Account
	// ? ------------------------------------------------------------------------------
	userRepo := accountRepo.NewUserRepository(dbConn, redisClient, log)
	developerRepo := accountRepo.NewDeveloperRepository(dbConn, redisClient)
	// ? ------------------------------------------------------------------------------
	// ? - Repository - Grammar
	// ? ------------------------------------------------------------------------------
	grammarQuestionRepo := grammarRepo.NewGrammarQuestionRepository(gormDB, log)
	grammarFillInTheBlankQuestionRepo := grammarRepo.NewGrammarFillInTheBlankQuestionRepository(gormDB, log)
	grammarFillInTheBlankAnswerRepo := grammarRepo.NewGrammarFillInTheBlankAnswerRepository(gormDB, log)
	grammarChoiceOneQuestionRepo := grammarRepo.NewGrammarChoiceOneQuestionRepository(gormDB, log)
	grammarChoiceOneOptionRepo := grammarRepo.NewGrammarChoiceOneOptionRepository(gormDB, log)
	grammarErrorIdentificationRepo := grammarRepo.NewGrammarErrorIdentificationRepository(gormDB, log)
	grammarSentenceTransformationRepo := grammarRepo.NewGrammarSentenceTransformationRepository(gormDB, log)
	grammarQuestionSearch := searchClient.NewGrammarQuestionSearch(openSearchClient, log)
	// ? ------------------------------------------------------------------------------
	// ? - Repository - Listening
	// ? ------------------------------------------------------------------------------
	listeningQuestionRepo := listeningRepo.NewListeningQuestionRepository(gormDB, log)
	listeningFillInTheBlankQuestionRepo := listeningRepo.NewListeningFillInTheBlankQuestionRepository(gormDB, log)
	listeningFillInTheBlankAnswerRepo := listeningRepo.NewListeningFillInTheBlankAnswerRepository(gormDB, log)
	listeningChoiceChoiceOneQuestionRepo := listeningRepo.NewListeningChoiceOneQuestionRepository(gormDB, log)
	listeningChoiceChoiceOneOptionRepo := listeningRepo.NewListeningChoiceOneOptionRepository(gormDB, log)
	listeningChoiceMultiQuestionRepo := listeningRepo.NewListeningChoiceMultiQuestionRepository(gormDB, log)
	listeningChoiceMultiOptionRepo := listeningRepo.NewListeningChoiceMultiOptionRepository(gormDB, log)
	listeningMapLabellingRepo := listeningRepo.NewListeningMapLabellingRepository(gormDB, log)
	listeningMatchingRepo := listeningRepo.NewListeningMatchingRepository(gormDB, log)
	listeningQuestionSearch := searchClient.NewListeningQuestionSearch(openSearchClient, log)
	// ? ------------------------------------------------------------------------------
	// ? - Repository - Reading
	// ? ------------------------------------------------------------------------------
	readingQuestionRepo := readingRepo.NewReadingQuestionRepository(gormDB, log)
	readingFillInTheBlankQuestionRepo := readingRepo.NewReadingFillInTheBlankQuestionRepository(gormDB, log)
	readingFillInTheBlankAnswerRepo := readingRepo.NewReadingFillInTheBlankAnswerRepository(gormDB, log)
	readingChoiceOneQuestionRepo := readingRepo.NewReadingChoiceOneQuestionRepository(gormDB, log)
	readingChoiceOneOptionRepo := readingRepo.NewReadingChoiceOneOptionRepository(gormDB, log)
	readingChoiceMultiQuestionRepo := readingRepo.NewReadingChoiceMultiQuestionRepository(gormDB, log)
	readingChoiceMultiOptionRepo := readingRepo.NewReadingChoiceMultiOptionRepository(gormDB, log)
	readingMatchingRepo := readingRepo.NewReadingMatchingRepository(gormDB, log)
	readingTrueFalseRepo := readingRepo.NewReadingTrueFalseRepository(gormDB, log)
	readingQuestionSearch := searchClient.NewReadingQuestionSearch(openSearchClient, log)
	// ? ------------------------------------------------------------------------------
	// ? - Repository - Speaking
	// ? ------------------------------------------------------------------------------
	speakingQuestionRepo := speakingRepo.NewSpeakingQuestionRepository(gormDB, log)
	speakingWordRepetitionRepo := speakingRepo.NewSpeakingWordRepetitionRepository(gormDB, log)
	speakingPhraseRepetitionRepo := speakingRepo.NewSpeakingPhraseRepetitionRepository(gormDB, log)
	speakingParagraphRepetitionRepo := speakingRepo.NewSpeakingParagraphRepetitionRepository(gormDB, log)
	speakingOpenParagraphRepo := speakingRepo.NewSpeakingOpenParagraphRepository(gormDB, log)
	speakingConversationalRepetitionRepo := speakingRepo.NewSpeakingConversationalRepetitionRepository(gormDB, log)
	speakingConversationalRepetitionQARepo := speakingRepo.NewSpeakingConversationalRepetitionQARepository(gormDB, log)
	speakingConversationalOpenRepo := speakingRepo.NewSpeakingConversationalOpenRepository(gormDB, log)
	speakingQuestionSearch := searchClient.NewSpeakingQuestionSearch(openSearchClient, log)
	// ? ------------------------------------------------------------------------------
	// ? - Repository - Writing
	// ? ------------------------------------------------------------------------------
	writingQuestionRepo := writingRepo.NewWritingQuestionRepository(gormDB, log)
	writingEssayRepo := writingRepo.NewWritingEssayRepository(gormDB, log)
	writingSentenceCompletionRepo := writingRepo.NewWritingSentenceCompletionRepository(gormDB, log)
	// ? ------------------------------------------------------------------------------
	// ? - Repository - Course
	// ? ------------------------------------------------------------------------------
	courseRepository := courseRepo.NewCourseRepository(gormDB, log)
	courseBookRepo := courseRepo.NewCourseBookRepository(gormDB, log)
	courseOtherRepo := courseRepo.NewCourseOtherRepository(gormDB, log)
	lessonRepo := courseRepo.NewLessonRepository(gormDB, log)
	lessonQuestionRepo := courseRepo.NewLessonQuestionRepository(gormDB, log)
	courseSearch := searchClient.NewCourseSearch(openSearchClient, log)

	// ! ------------------------------------------------------------------------------
	// ! - Service
	// ! ------------------------------------------------------------------------------
	// ? ------------------------------------------------------------------------------
	// ? - Service - Account
	// ? ------------------------------------------------------------------------------

	userService := accountSer.NewUserService(userRepo)
	developerService := accountSer.NewDeveloperService(developerRepo)

	// ? ------------------------------------------------------------------------------
	// ? - Service - Grammar
	// ? ------------------------------------------------------------------------------
	grammarFillInTheBlankQuestionService := grammarSer.NewGrammarFillInTheBlankQuestionService(
		grammarFillInTheBlankQuestionRepo,
		grammarQuestionRepo,
		log,
		redisClient,
		grammarQuestionSearch,
		nil,
	)

	grammarFillInTheBlankAnswerService := grammarSer.NewGrammarFillInTheBlankAnswerService(
		grammarFillInTheBlankAnswerRepo,
		grammarFillInTheBlankQuestionRepo,
		grammarQuestionRepo,
		log,
		redisClient,
		nil,
	)

	grammarChoiceOneQuestionService := grammarSer.NewGrammarChoiceOneQuestionService(
		grammarChoiceOneQuestionRepo,
		grammarQuestionRepo,
		log,
		nil,
	)

	grammarChoiceOneOptionService := grammarSer.NewGrammarChoiceOneOptionService(
		grammarChoiceOneOptionRepo,
		grammarChoiceOneQuestionRepo,
		grammarQuestionRepo,
		log,
		redisClient,
		nil,
	)

	grammarErrorIdentificationService := grammarSer.NewGrammarErrorIdentificationService(
		grammarErrorIdentificationRepo,
		grammarQuestionRepo,
		log,
		redisClient,
		nil,
	)

	grammarSentenceTransformationService := grammarSer.NewGrammarSentenceTransformationService(
		grammarSentenceTransformationRepo,
		grammarQuestionRepo,
		log,
		redisClient,
		nil,
	)

	grammarQuestionUpdator := grammarHelper.NewGrammarQuestionUpdator(
		log,
		redisClient,
		openSearchClient,
		grammarFillInTheBlankQuestionService,
		grammarFillInTheBlankAnswerService,
		grammarChoiceOneQuestionService,
		grammarChoiceOneOptionService,
		grammarErrorIdentificationService,
		grammarSentenceTransformationService,
	)

	grammarFillInTheBlankQuestionService.SetQuestionUpdator(grammarQuestionUpdator)
	grammarFillInTheBlankAnswerService.SetQuestionUpdator(grammarQuestionUpdator)
	grammarChoiceOneQuestionService.SetQuestionUpdator(grammarQuestionUpdator)
	grammarChoiceOneOptionService.SetQuestionUpdator(grammarQuestionUpdator)
	grammarErrorIdentificationService.SetQuestionUpdator(grammarQuestionUpdator)
	grammarSentenceTransformationService.SetQuestionUpdator(grammarQuestionUpdator)

	grammarQuestionService := grammarSer.NewGrammarQuestionService(
		grammarQuestionRepo,
		log,
		redisClient,
		openSearchClient,
		grammarFillInTheBlankQuestionService,
		grammarFillInTheBlankAnswerService,
		grammarChoiceOneQuestionService,
		grammarChoiceOneOptionService,
		grammarErrorIdentificationService,
		grammarSentenceTransformationService,
		grammarQuestionUpdator,
	)

	// ? ------------------------------------------------------------------------------
	// ? - Service - Listening
	// ? ------------------------------------------------------------------------------
	listeningFillInTheBlankQuestionService := listeningSer.NewListeningFillInTheBlankQuestionService(
		listeningFillInTheBlankQuestionRepo,
		listeningQuestionRepo,
		log,
		redisClient,
		listeningQuestionSearch,
		nil,
	)

	listeningFillInTheBlankAnswerService := listeningSer.NewListeningFillInTheBlankAnswerService(
		listeningFillInTheBlankAnswerRepo,
		listeningFillInTheBlankQuestionRepo,
		listeningQuestionRepo,
		log,
		redisClient,
		nil,
	)

	listeningChoiceOneQuestionService := listeningSer.NewListeningChoiceOneQuestionService(
		listeningChoiceChoiceOneQuestionRepo,
		listeningQuestionRepo,
		log,
		redisClient,
		nil,
	)

	listeningChoiceOneOptionService := listeningSer.NewListeningChoiceOneOptionService(
		listeningChoiceChoiceOneOptionRepo,
		listeningChoiceChoiceOneQuestionRepo,
		listeningQuestionRepo,
		log,
		redisClient,
		nil,
	)

	listeningChoiceMultiQuestionService := listeningSer.NewListeningChoiceMultiQuestionService(
		listeningChoiceMultiQuestionRepo,
		listeningQuestionRepo,
		log,
		nil,
	)

	listeningChoiceMultiOptionService := listeningSer.NewListeningChoiceMultiOptionService(
		listeningChoiceMultiOptionRepo,
		listeningChoiceMultiQuestionRepo,
		listeningQuestionRepo,
		log,
		redisClient,
		nil,
	)

	listeningMapLabellingService := listeningSer.NewListeningMapLabellingService(
		listeningMapLabellingRepo,
		listeningQuestionRepo,
		log,
		redisClient,
		nil,
	)

	listeningMatchingService := listeningSer.NewListeningMatchingService(
		listeningMatchingRepo,
		listeningQuestionRepo,
		log,
		redisClient,
		nil,
	)

	questionUpdator := listeningHelper.NewListeningQuestionUpdator(
		log,
		redisClient,
		openSearchClient,
		listeningFillInTheBlankQuestionService,
		listeningFillInTheBlankAnswerService,
		listeningChoiceOneQuestionService,
		listeningChoiceOneOptionService,
		listeningChoiceMultiQuestionService,
		listeningChoiceMultiOptionService,
		listeningMapLabellingService,
		listeningMatchingService,
	)

	listeningFillInTheBlankQuestionService.SetQuestionUpdator(questionUpdator)
	listeningFillInTheBlankAnswerService.SetQuestionUpdator(questionUpdator)
	listeningChoiceOneQuestionService.SetQuestionUpdator(questionUpdator)
	listeningChoiceOneOptionService.SetQuestionUpdator(questionUpdator)
	listeningChoiceMultiQuestionService.SetQuestionUpdator(questionUpdator)
	listeningChoiceMultiOptionService.SetQuestionUpdator(questionUpdator)
	listeningMapLabellingService.SetQuestionUpdator(questionUpdator)
	listeningMatchingService.SetQuestionUpdator(questionUpdator)

	listeningQuestionService := listeningSer.NewListeningQuestionService(
		listeningQuestionRepo,
		log,
		redisClient,
		openSearchClient,
		listeningFillInTheBlankQuestionService,
		listeningFillInTheBlankAnswerService,
		listeningChoiceOneQuestionService,
		listeningChoiceOneOptionService,
		listeningChoiceMultiQuestionService,
		listeningChoiceMultiOptionService,
		listeningMapLabellingService,
		listeningMatchingService,
		questionUpdator,
	)

	// ? ------------------------------------------------------------------------------
	// ? - Service - Reading
	// ? ------------------------------------------------------------------------------
	readingFillInTheBlankQuestionService := readingSer.NewReadingFillInTheBlankQuestionService(
		readingFillInTheBlankQuestionRepo,
		readingQuestionRepo,
		log,
		redisClient,
		readingQuestionSearch,
		nil,
	)

	readingFillInTheBlankAnswerService := readingSer.NewReadingFillInTheBlankAnswerService(
		readingFillInTheBlankAnswerRepo,
		readingFillInTheBlankQuestionRepo,
		readingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	readingChoiceOneQuestionService := readingSer.NewReadingChoiceOneQuestionService(
		readingChoiceOneQuestionRepo,
		readingQuestionRepo,
		log,
		nil,
	)

	readingChoiceOneOptionService := readingSer.NewReadingChoiceOneOptionService(
		readingChoiceOneOptionRepo,
		readingChoiceOneQuestionRepo,
		readingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	readingChoiceMultiQuestionService := readingSer.NewReadingChoiceMultiQuestionService(
		readingChoiceMultiQuestionRepo,
		readingQuestionRepo,
		log,
		nil,
	)

	readingChoiceMultiOptionService := readingSer.NewReadingChoiceMultiOptionService(
		readingChoiceMultiOptionRepo,
		readingChoiceMultiQuestionRepo,
		readingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	readingMatchingService := readingSer.NewReadingMatchingService(
		readingMatchingRepo,
		readingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	readingTrueFalseService := readingSer.NewReadingTrueFalseService(
		readingTrueFalseRepo,
		readingQuestionRepo,
		log,
		// redisClient,
		nil,
	)

	readingQuestionUpdator := readingHelper.NewReadingQuestionUpdator(
		log,
		redisClient,
		openSearchClient,
		readingFillInTheBlankQuestionService,
		readingFillInTheBlankAnswerService,
		readingChoiceOneQuestionService,
		readingChoiceOneOptionService,
		readingChoiceMultiQuestionService,
		readingChoiceMultiOptionService,
		readingTrueFalseService,
		readingMatchingService,
	)

	readingFillInTheBlankQuestionService.SetQuestionUpdator(readingQuestionUpdator)
	readingFillInTheBlankAnswerService.SetQuestionUpdator(readingQuestionUpdator)
	readingChoiceOneQuestionService.SetQuestionUpdator(readingQuestionUpdator)
	readingChoiceOneOptionService.SetQuestionUpdator(readingQuestionUpdator)
	readingChoiceMultiQuestionService.SetQuestionUpdator(readingQuestionUpdator)
	readingChoiceMultiOptionService.SetQuestionUpdator(readingQuestionUpdator)
	readingMatchingService.SetQuestionUpdator(readingQuestionUpdator)
	readingTrueFalseService.SetQuestionUpdator(readingQuestionUpdator)

	readingQuestionService := readingSer.NewReadingQuestionService(
		readingQuestionRepo,
		log,
		redisClient,
		openSearchClient,
		readingFillInTheBlankQuestionService,
		readingFillInTheBlankAnswerService,
		readingChoiceOneQuestionService,
		readingChoiceOneOptionService,
		readingChoiceMultiQuestionService,
		readingChoiceMultiOptionService,
		readingTrueFalseService,
		readingMatchingService,
		readingQuestionUpdator,
	)

	// ? ------------------------------------------------------------------------------
	// ? - Service - Speaking
	// ? ------------------------------------------------------------------------------
	speakingWordRepetitionService := speakingSer.NewSpeakingWordRepetitionService(
		speakingWordRepetitionRepo,
		speakingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	speakingPhraseRepetitionService := speakingSer.NewSpeakingPhraseRepetitionService(
		speakingPhraseRepetitionRepo,
		speakingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	speakingParagraphRepetitionService := speakingSer.NewSpeakingParagraphRepetitionService(
		speakingParagraphRepetitionRepo,
		speakingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	speakingOpenParagraphService := speakingSer.NewSpeakingOpenParagraphService(
		speakingOpenParagraphRepo,
		speakingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	speakingConversationalRepetitionService := speakingSer.NewSpeakingConversationalRepetitionService(
		speakingConversationalRepetitionRepo,
		speakingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	speakingConversationalRepetitionQAService := speakingSer.NewSpeakingConversationalRepetitionQAService(
		speakingConversationalRepetitionQARepo,
		speakingConversationalRepetitionRepo,
		speakingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	speakingConversationalOpenService := speakingSer.NewSpeakingConversationalOpenService(
		speakingConversationalOpenRepo,
		speakingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	speakingQuestionUpdator := speakingHelper.NewSpeakingQuestionUpdator(
		log,
		redisClient,
		openSearchClient,
		speakingWordRepetitionService,
		speakingPhraseRepetitionService,
		speakingParagraphRepetitionService,
		speakingOpenParagraphService,
		speakingConversationalRepetitionService,
		speakingConversationalRepetitionQAService,
		speakingConversationalOpenService,
	)

	speakingWordRepetitionService.SetQuestionUpdator(speakingQuestionUpdator)
	speakingPhraseRepetitionService.SetQuestionUpdator(speakingQuestionUpdator)
	speakingParagraphRepetitionService.SetQuestionUpdator(speakingQuestionUpdator)
	speakingOpenParagraphService.SetQuestionUpdator(speakingQuestionUpdator)
	speakingConversationalRepetitionService.SetQuestionUpdator(speakingQuestionUpdator)
	speakingConversationalRepetitionQAService.SetQuestionUpdator(speakingQuestionUpdator)
	speakingConversationalOpenService.SetQuestionUpdator(speakingQuestionUpdator)

	speakingQuestionService := speakingSer.NewSpeakingQuestionService(
		speakingQuestionRepo,
		log,
		redisClient,
		speakingQuestionSearch,
		speakingWordRepetitionService,
		speakingPhraseRepetitionService,
		speakingParagraphRepetitionService,
		speakingOpenParagraphService,
		speakingConversationalRepetitionService,
		speakingConversationalRepetitionQAService,
		speakingConversationalOpenService,
		speakingQuestionUpdator,
	)

	// ? ------------------------------------------------------------------------------
	// ? - Service - Writing
	// ? ------------------------------------------------------------------------------
	writingEssayService := writingSer.NewWritingEssayService(
		writingEssayRepo,
		writingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	writingSentenceCompletionService := writingSer.NewWritingSentenceCompletionService(
		writingSentenceCompletionRepo,
		writingQuestionRepo,
		log,
		redisClient,
		nil,
	)

	writingQuestionSearch := searchClient.NewWritingQuestionSearch(openSearchClient, log)

	writingQuestionUpdator := writingHelper.NewWritingQuestionUpdator(
		log,
		redisClient,
		openSearchClient,
		writingSentenceCompletionService,
		writingEssayService,
	)

	writingEssayService.SetQuestionUpdator(writingQuestionUpdator)
	writingSentenceCompletionService.SetQuestionUpdator(writingQuestionUpdator)

	writingQuestionService := writingSer.NewWritingQuestionService(
		writingQuestionRepo,
		log,
		redisClient,
		writingQuestionSearch,
		writingSentenceCompletionService,
		writingEssayService,
		writingQuestionUpdator,
	)
	// ? ------------------------------------------------------------------------------
	// ? - Service - Course
	// ? ------------------------------------------------------------------------------
	lessonQuestionService := courseSer.NewLessonQuestionService(
		lessonQuestionRepo,
		lessonRepo,
		courseRepository,
		log,
		redisClient,
		nil,
	)

	lessonService := courseSer.NewLessonService(
		lessonRepo,
		courseRepository,
		log,
		redisClient,
		nil,
	)

	courseOtherService := courseSer.NewCourseOtherService(
		courseOtherRepo,
		courseRepository,
		log,
		redisClient,
		nil,
	)

	courseBookService := courseSer.NewCourseBookService(
		courseBookRepo,
		courseRepository,
		log,
		redisClient,
		nil,
	)

	courseUpdator := courseHelper.NewCourseUpdator(
		log,
		redisClient,
		openSearchClient,
		lessonService,
		lessonQuestionService,
		courseBookService,
		courseOtherService,
	)

	lessonQuestionService.SetCourseUpdator(courseUpdator)
	lessonService.SetCourseUpdator(courseUpdator)
	courseOtherService.SetCourseUpdator(courseUpdator)
	courseBookService.SetCourseUpdator(courseUpdator)

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

	// ! ------------------------------------------------------------------------------
	// ! - Handler
	// ! ------------------------------------------------------------------------------
	// ? ------------------------------------------------------------------------------
	// ? - Handler - Account
	// ? ------------------------------------------------------------------------------
	userHandler := accountHa.NewUserHandler(userService)
	developerHandler := accountHa.NewDeveloperHandler(developerService)
	// ? ------------------------------------------------------------------------------
	// ? - Handler - Grammar
	// ? ------------------------------------------------------------------------------
	grammarQuestionHandler := grammarHa.NewGrammarQuestionHandler(
		grammarQuestionService,
		grammarFillInTheBlankQuestionService,
		grammarFillInTheBlankAnswerService,
		grammarChoiceOneQuestionService,
		grammarChoiceOneOptionService,
		grammarErrorIdentificationService,
		grammarSentenceTransformationService,
		log,
	)

	grammarFillInTheBlankQuestionHandler := grammarHa.NewGrammarFillInTheBlankQuestionHandler(
		grammarFillInTheBlankQuestionService,
		grammarQuestionRepo,
		log,
	)

	grammarFillInTheBlankAnswerHandler := grammarHa.NewGrammarFillInTheBlankAnswerHandler(
		grammarFillInTheBlankAnswerService,
		log,
	)

	grammarChoiceOneQuestionHandler := grammarHa.NewGrammarChoiceOneQuestionHandler(
		grammarChoiceOneQuestionService,
		grammarChoiceOneOptionService,
		log,
	)

	grammarChoiceOneOptionHandler := grammarHa.NewGrammarChoiceOneOptionHandler(
		grammarChoiceOneOptionService,
		log,
	)

	grammarErrorIdentificationHandler := grammarHa.NewGrammarErrorIdentificationHandler(
		grammarErrorIdentificationService,
		log,
	)

	grammarSentenceTransformationHandler := grammarHa.NewGrammarSentenceTransformationHandler(
		grammarSentenceTransformationService,
		log,
	)

	// ? ------------------------------------------------------------------------------
	// ? - Handler - Listening
	// ? ------------------------------------------------------------------------------
	listeningQuestionHandler := listeningHa.NewListeningQuestionHandler(
		listeningQuestionService,
		listeningFillInTheBlankQuestionService,
		listeningFillInTheBlankAnswerService,
		listeningChoiceOneQuestionService,
		listeningChoiceOneOptionService,
		listeningChoiceMultiQuestionService,
		listeningChoiceMultiOptionService,
		listeningMapLabellingService,
		listeningMatchingService,
		log,
	)

	listeningFillInTheBlankQuestionHandler := listeningHa.NewListeningFillInTheBlankQuestionHandler(
		listeningFillInTheBlankQuestionService,
		listeningQuestionRepo,
		log,
	)

	listeningFillInTheBlankAnswerHandler := listeningHa.NewListeningFillInTheBlankAnswerHandler(
		listeningFillInTheBlankAnswerService,
		log,
	)

	listeningChoiceOneQuestionHandler := listeningHa.NewListeningChoiceOneQuestionHandler(
		listeningChoiceOneQuestionService,
		listeningChoiceOneOptionService,
		log,
	)

	listeningChoiceOneOptionHandler := listeningHa.NewListeningChoiceOneOptionHandler(
		listeningChoiceOneOptionService,
		log,
	)

	listeningChoiceMultiQuestionHandler := listeningHa.NewListeningChoiceMultiQuestionHandler(
		listeningChoiceMultiQuestionService,
		listeningChoiceMultiOptionService,
		log,
	)

	listeningChoiceMultiOptionHandler := listeningHa.NewListeningChoiceMultiOptionHandler(
		listeningChoiceMultiOptionService,
		log,
	)

	listeningMapLabellingHandler := listeningHa.NewListeningMapLabellingHandler(
		listeningMapLabellingService,
		log,
	)

	listeningMatchingHandler := listeningHa.NewListeningMatchingHandler(
		listeningMatchingService,
		log,
	)

	// ? ------------------------------------------------------------------------------
	// ? - Handler - Reading
	// ? ------------------------------------------------------------------------------
	readingQuestionHandler := readingHa.NewReadingQuestionHandler(
		readingQuestionService,
		readingFillInTheBlankQuestionService,
		readingFillInTheBlankAnswerService,
		readingChoiceOneQuestionService,
		readingChoiceOneOptionService,
		readingChoiceMultiQuestionService,
		readingChoiceMultiOptionService,
		readingTrueFalseService,
		readingMatchingService,
		log,
	)

	readingFillInTheBlankQuestionHandler := readingHa.NewReadingFillInTheBlankQuestionHandler(
		readingFillInTheBlankQuestionService,
		readingQuestionRepo,
		log,
	)

	readingFillInTheBlankAnswerHandler := readingHa.NewReadingFillInTheBlankAnswerHandler(
		readingFillInTheBlankAnswerService,
		log,
	)

	readingChoiceOneQuestionHandler := readingHa.NewReadingChoiceOneQuestionHandler(
		readingChoiceOneQuestionService,
		readingChoiceOneOptionService,
		log,
	)

	readingChoiceOneOptionHandler := readingHa.NewReadingChoiceOneOptionHandler(
		readingChoiceOneOptionService,
		log,
	)

	readingChoiceMultiQuestionHandler := readingHa.NewReadingChoiceMultiQuestionHandler(
		readingChoiceMultiQuestionService,
		readingChoiceMultiOptionService,
		log,
	)

	readingChoiceMultiOptionHandler := readingHa.NewReadingChoiceMultiOptionHandler(
		readingChoiceMultiOptionService,
		log,
	)

	readingMatchingHandler := readingHa.NewReadingMatchingHandler(
		readingMatchingService,
		log,
	)

	readingTrueFalseHandler := readingHa.NewReadingTrueFalseHandler(
		readingTrueFalseService,
		log,
	)

	// ? ------------------------------------------------------------------------------
	// ? - Handler - Speaking
	// ? ------------------------------------------------------------------------------
	speakingQuestionHandler := speakingHa.NewSpeakingQuestionHandler(
		speakingQuestionService,
		speakingWordRepetitionService,
		speakingPhraseRepetitionService,
		speakingParagraphRepetitionService,
		speakingOpenParagraphService,
		speakingConversationalRepetitionService,
		speakingConversationalRepetitionQAService,
		speakingConversationalOpenService,
		log,
	)

	speakingWordRepetitionHandler := speakingHa.NewSpeakingWordRepetitionHandler(
		speakingWordRepetitionService,
		log,
	)

	speakingPhraseRepetitionHandler := speakingHa.NewSpeakingPhraseRepetitionHandler(
		speakingPhraseRepetitionService,
		log,
	)

	speakingParagraphRepetitionHandler := speakingHa.NewSpeakingParagraphRepetitionHandler(
		speakingParagraphRepetitionService,
		log,
	)

	speakingOpenParagraphHandler := speakingHa.NewSpeakingOpenParagraphHandler(
		speakingOpenParagraphService,
		log,
	)

	speakingConversationalRepetitionHandler := speakingHa.NewSpeakingConversationalRepetitionHandler(
		speakingConversationalRepetitionService,
		log,
	)

	speakingConversationalRepetitionQAHandler := speakingHa.NewSpeakingConversationalRepetitionQAHandler(
		speakingConversationalRepetitionQAService,
		log,
	)

	speakingConversationalOpenHandler := speakingHa.NewSpeakingConversationalOpenHandler(
		speakingConversationalOpenService,
		log,
	)

	// ? ------------------------------------------------------------------------------
	// ? - Handler - Writing
	// ? ------------------------------------------------------------------------------
	writingQuestionHandler := writingHa.NewWritingQuestionHandler(
		writingQuestionService,
		writingSentenceCompletionService,
		writingEssayService,
		log,
	)

	writingSentenceCompletionHandler := writingHa.NewWritingSentenceCompletionHandler(
		writingSentenceCompletionService,
		log,
	)

	writingEssayHandler := writingHa.NewWritingEssayHandler(
		writingEssayService,
		log,
	)

	// ? ------------------------------------------------------------------------------
	// ? - Handler - Course
	// ? ------------------------------------------------------------------------------
	courseHandler := courseHa.NewCourseHandler(
		courseService,
		courseBookService,
		courseOtherService,
		lessonService,
		lessonQuestionService,
		log,
	)

	courseBookHandler := courseHa.NewCourseBookHandler(
		courseBookService,
		log,
	)

	courseOtherHandler := courseHa.NewCourseOtherHandler(
		courseOtherService,
		log,
	)

	lessonHandler := courseHa.NewLessonHandler(
		lessonService,
		log,
	)

	lessonQuestionHandler := courseHa.NewLessonQuestionHandler(
		lessonQuestionService,
		log,
	)

	// ! ------------------------------------------------------------------------------
	// ! - Routers
	// ! ------------------------------------------------------------------------------
	r := router.NewRouter(dbConn)
	r.SetupRoutes(
		userHandler,
		developerHandler,
		grammarQuestionHandler,
		grammarFillInTheBlankQuestionHandler,
		grammarFillInTheBlankAnswerHandler,
		grammarChoiceOneQuestionHandler,
		grammarChoiceOneOptionHandler,
		grammarErrorIdentificationHandler,
		grammarSentenceTransformationHandler,
		listeningQuestionHandler,
		listeningFillInTheBlankQuestionHandler,
		listeningFillInTheBlankAnswerHandler,
		listeningChoiceOneQuestionHandler,
		listeningChoiceOneOptionHandler,
		listeningChoiceMultiQuestionHandler,
		listeningChoiceMultiOptionHandler,
		listeningMapLabellingHandler,
		listeningMatchingHandler,
		readingQuestionHandler,
		readingFillInTheBlankQuestionHandler,
		readingFillInTheBlankAnswerHandler,
		readingChoiceOneQuestionHandler,
		readingChoiceOneOptionHandler,
		readingChoiceMultiQuestionHandler,
		readingChoiceMultiOptionHandler,
		readingTrueFalseHandler,
		readingMatchingHandler,
		speakingQuestionHandler,
		speakingWordRepetitionHandler,
		speakingPhraseRepetitionHandler,
		speakingParagraphRepetitionHandler,
		speakingOpenParagraphHandler,
		speakingConversationalRepetitionHandler,
		speakingConversationalRepetitionQAHandler,
		speakingConversationalOpenHandler,
		writingQuestionHandler,
		writingSentenceCompletionHandler,
		writingEssayHandler,
		courseHandler,
		courseBookHandler,
		courseOtherHandler,
		lessonHandler,
		lessonQuestionHandler,
	)

	ginEngine := r.Engine

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      ginEngine,
		WriteTimeout: constants.HTTPWriteTimeout,
		ReadTimeout:  constants.HTTPReadTimeout,
		IdleTimeout:  constants.HTTPIdleTimeout,
	}

	return &Application{
		Config:     cfg,
		Logger:     log,
		Router:     ginEngine,
		Server:     server,
		DBConn:     dbConn,
		GormDB:     gormDB,
		Redis:      redisClient,
		OpenSearch: openSearchClient,
		DiscordBot: discordBot,
		Metrics:    metricsCollector,
	}, nil
}
