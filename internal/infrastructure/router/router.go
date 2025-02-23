package router

import (
	"context"
	"database/sql"
	accountHandler "fluencybe/internal/app/handler/account"
	courseHa "fluencybe/internal/app/handler/course"
	grammarHandler "fluencybe/internal/app/handler/grammar"
	listeningHandler "fluencybe/internal/app/handler/listening"
	readingHandler "fluencybe/internal/app/handler/reading"
	speakingHandler "fluencybe/internal/app/handler/speaking"
	writingHandler "fluencybe/internal/app/handler/writing"
	constants "fluencybe/internal/core/constants"
	"net/http"

	CORS "fluencybe/pkg/cors"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/middleware"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type contextKey string

const (
	GinContextKey contextKey = "GinContextKey"
)

type Router struct {
	*gin.Engine
	logger *logger.PrettyLogger
	db     *sql.DB
}

func (r *Router) wrapHandler(handler func(context.Context, http.ResponseWriter, *http.Request)) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		handler(ctx, c.Writer, c.Request)
	}
}

func NewRouter(db *sql.DB) *Router {
	// Set Gin mode based on GIN_DEBUG_LOG environment variable
	if os.Getenv("GIN_DEBUG_LOG") == "TRUE" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	return &Router{
		Engine: gin.New(),
		logger: logger.GetGlobalLogger(),
		db:     db,
	}
}

type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (r *Router) SetupRoutes(
	//* Account
	userHandler *accountHandler.UserHandler,
	developerHandler *accountHandler.DeveloperHandler,
	//* Grammar
	grammarQuestionHandler *grammarHandler.GrammarQuestionHandler,
	grammarFillInTheBlankQuestionHandler *grammarHandler.GrammarFillInTheBlankQuestionHandler,
	grammarFillInTheBlankAnswerHandler *grammarHandler.GrammarFillInTheBlankAnswerHandler,
	grammarChoiceOneQuestionHandler *grammarHandler.GrammarChoiceOneQuestionHandler,
	grammarChoiceOneOptionHandler *grammarHandler.GrammarChoiceOneOptionHandler,
	grammarErrorIdentificationHandler *grammarHandler.GrammarErrorIdentificationHandler,
	grammarSentenceTransformationHandler *grammarHandler.GrammarSentenceTransformationHandler,
	//* Listening
	listeningQuestionHandler *listeningHandler.ListeningQuestionHandler,
	listeningFillInTheBlankQuestionHandler *listeningHandler.ListeningFillInTheBlankQuestionHandler,
	listeningFillInTheBlankAnswerHandler *listeningHandler.ListeningFillInTheBlankAnswerHandler,
	listeningChoiceOneQuestionHandler *listeningHandler.ListeningChoiceOneQuestionHandler,
	listeningChoiceOneOptionHandler *listeningHandler.ListeningChoiceOneOptionHandler,
	listeningChoiceMultiQuestionHandler *listeningHandler.ListeningChoiceMultiQuestionHandler,
	listeningChoiceMultiOptionHandler *listeningHandler.ListeningChoiceMultiOptionHandler,
	listeningMapLabellingHandler *listeningHandler.ListeningMapLabellingHandler,
	listeningMatchingHandler *listeningHandler.ListeningMatchingHandler,
	//* Reading
	readingQuestionHandler *readingHandler.ReadingQuestionHandler,
	readingFillInTheBlankQuestionHandler *readingHandler.ReadingFillInTheBlankQuestionHandler,
	readingFillInTheBlankAnswerHandler *readingHandler.ReadingFillInTheBlankAnswerHandler,
	readingChoiceOneQuestionHandler *readingHandler.ReadingChoiceOneQuestionHandler,
	readingChoiceOneOptionHandler *readingHandler.ReadingChoiceOneOptionHandler,
	readingChoiceMultiQuestionHandler *readingHandler.ReadingChoiceMultiQuestionHandler,
	readingChoiceMultiOptionHandler *readingHandler.ReadingChoiceMultiOptionHandler,
	readingTrueFalseHandler *readingHandler.ReadingTrueFalseHandler,
	readingMatchingHandler *readingHandler.ReadingMatchingHandler,
	//* Speaking
	speakingQuestionHandler *speakingHandler.SpeakingQuestionHandler,
	speakingWordRepetitionHandler *speakingHandler.SpeakingWordRepetitionHandler,
	speakingPhraseRepetitionHandler *speakingHandler.SpeakingPhraseRepetitionHandler,
	speakingParagraphRepetitionHandler *speakingHandler.SpeakingParagraphRepetitionHandler,
	speakingOpenParagraphHandler *speakingHandler.SpeakingOpenParagraphHandler,
	speakingConversationalRepetitionHandler *speakingHandler.SpeakingConversationalRepetitionHandler,
	speakingConversationalRepetitionQAHandler *speakingHandler.SpeakingConversationalRepetitionQAHandler,
	speakingConversationalOpenHandler *speakingHandler.SpeakingConversationalOpenHandler,
	//* Writing
	writingQuestionHandler *writingHandler.WritingQuestionHandler,
	writingSentenceCompletionHandler *writingHandler.WritingSentenceCompletionHandler,
	writingEssayHandler *writingHandler.WritingEssayHandler,
	//* Course
	courseHandler *courseHa.CourseHandler,
	courseBookHandler *courseHa.CourseBookHandler,
	courseOtherHandler *courseHa.CourseOtherHandler,
	lessonHandler *courseHa.LessonHandler,
	lessonQuestionHandler *courseHa.LessonQuestionHandler,
) {

	gin.ForceConsoleColor()

	// Create rate limiter
	limiter := rate.NewLimiter(rate.Every(time.Minute), 100)

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(429, StandardResponse{
				Success: false,
				Error:   "Too many requests",
			})
			c.Abort()
			return
		}
		c.Next()
	})

	CORS.SetupCORS(r.Engine)

	// API v1 group
	api := r.Group("/v1")

	// ! ------------------------------------------------------------------------------
	// ! - Account
	// ! ------------------------------------------------------------------------------
	// ? ------------------------------------------------------------------------------
	// ? - Account - User
	// ? ------------------------------------------------------------------------------
	user := api.Group("/user")
	{
		user.POST("/register", gin.HandlerFunc(func(c *gin.Context) {
			userHandler.Register(c.Request.Context(), c.Writer, c.Request)
		}))
		user.POST("/login", gin.HandlerFunc(func(c *gin.Context) {
			userHandler.Login(c.Request.Context(), c.Writer, c.Request)
		}))
		user.GET("", middleware.UserAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.GetMyUser(c.Request.Context(), c.Writer, c.Request)
		}))
		user.GET("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.GetUser(c.Request.Context(), c.Writer, c.Request)
		}))
		user.PUT("", middleware.UserAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.UpdateMyUser(c.Request.Context(), c.Writer, c.Request)
		}))
		user.PUT("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.UpdateUser(c.Request.Context(), c.Writer, c.Request)
		}))
		user.DELETE("", middleware.UserAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.DeleteMyUser(c.Request.Context(), c.Writer, c.Request)
		}))
		user.DELETE("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.DeleteUser(c.Request.Context(), c.Writer, c.Request)
		}))
		user.GET("/list", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.GetListUserWithPagination(c.Request.Context(), c.Writer, c.Request)
		}))
	}
	// ? ------------------------------------------------------------------------------
	// ? - Account - Developer
	// ? ------------------------------------------------------------------------------
	developer := api.Group("/developer")
	{
		developer.POST("/register", gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.Register(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.POST("/login", gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.Login(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.GET("", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.GetMyDeveloper(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.GET("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.GetDeveloper(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.PUT("", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.UpdateMyDeveloper(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.PUT("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.UpdateDeveloper(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.DELETE("", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.DeleteMyDeveloper(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.DELETE("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.DeleteDeveloper(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.GET("/list", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.GetListDeveloperWithPagination(c.Request.Context(), c.Writer, c.Request)
		}))
	}
	// ! ------------------------------------------------------------------------------
	// ! - Listening
	// ! ------------------------------------------------------------------------------
	// ? ------------------------------------------------------------------------------
	// ? - Listening - ListeningQuestion
	// ? ------------------------------------------------------------------------------
	listeningQuestion := api.Group("/listening/question")
	listeningQuestion.POST("", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		listeningQuestionHandler.CreateListeningQuestion(ctx, c.Writer, c.Request)
	}))

	listeningQuestion.GET("/:id", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		listeningQuestionHandler.GetListeningQuestionDetail(ctx, c.Writer, c.Request)
	}))

	listeningQuestion.PUT("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		listeningQuestionHandler.UpdateListeningQuestion(ctx, c.Writer, c.Request)
	}))

	listeningQuestion.DELETE("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		listeningQuestionHandler.DeleteListeningQuestion(ctx, c.Writer, c.Request)
	}))

	listeningQuestion.POST("/get-new-updates", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		listeningQuestionHandler.GetListNewListeningQuestionByListVersionAndID(ctx, c.Writer, c.Request)
	}))

	listeningQuestion.POST("/get-by-list-id", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		listeningQuestionHandler.GetListListeningByListID(ctx, c.Writer, c.Request)
	}))

	listeningQuestion.GET("/search", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		listeningQuestionHandler.GetListListeningQuestiondetailPaganationWithFilter(ctx, c.Writer, c.Request)
	}))

	listeningQuestion.DELETE("/delete-all", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		listeningQuestionHandler.DeleteAllListeningData(ctx, c.Writer, c.Request)
	}))

	// ? ------------------------------------------------------------------------------
	// ? - Listening - Listening Fill In The Blank
	// ? ------------------------------------------------------------------------------
	listeningFillInTheBlankQuestion := api.Group("/listening/fill-in-the-blank-question")
	listeningFillInTheBlankQuestion.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		listeningFillInTheBlankQuestion.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningFillInTheBlankQuestionHandler.CreateQuestion(ctx, c.Writer, c.Request)
		}))
		listeningFillInTheBlankQuestion.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningFillInTheBlankQuestionHandler.UpdateQuestion(ctx, c.Writer, c.Request)
		}))
		listeningFillInTheBlankQuestion.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningFillInTheBlankQuestionHandler.DeleteQuestion(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Listening - Listening Fill In The Blank Answer
	// ? ------------------------------------------------------------------------------
	listeningFillInTheBlankAnswer := api.Group("/listening/fill-in-the-blank-answer")
	listeningFillInTheBlankAnswer.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		listeningFillInTheBlankAnswer.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningFillInTheBlankAnswerHandler.CreateAnswer(ctx, c.Writer, c.Request)
		}))
		listeningFillInTheBlankAnswer.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningFillInTheBlankAnswerHandler.UpdateAnswer(ctx, c.Writer, c.Request)
		}))
		listeningFillInTheBlankAnswer.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningFillInTheBlankAnswerHandler.DeleteAnswer(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Listening - Listening Choice One Question
	// ? ------------------------------------------------------------------------------
	listeningChoiceOneQuestion := api.Group("/listening/choice-one-question")
	listeningChoiceOneQuestion.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		listeningChoiceOneQuestion.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningChoiceOneQuestionHandler.CreateQuestion(ctx, c.Writer, c.Request)
		}))
		listeningChoiceOneQuestion.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningChoiceOneQuestionHandler.UpdateQuestion(ctx, c.Writer, c.Request)
		}))
		listeningChoiceOneQuestion.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningChoiceOneQuestionHandler.DeleteQuestion(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Listening - Listening Choice One Option
	// ? ------------------------------------------------------------------------------
	listeningChoiceOneOption := api.Group("/listening/choice-one-option")
	listeningChoiceOneOption.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		listeningChoiceOneOption.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningChoiceOneOptionHandler.CreateOption(ctx, c.Writer, c.Request)
		}))
		listeningChoiceOneOption.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningChoiceOneOptionHandler.UpdateOption(ctx, c.Writer, c.Request)
		}))
		listeningChoiceOneOption.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningChoiceOneOptionHandler.DeleteOption(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Listening - Listening Choice Multi Question
	// ? ------------------------------------------------------------------------------
	listeningChoiceMultiQuestion := api.Group("/listening/choice-multi-question")
	listeningChoiceMultiQuestion.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		listeningChoiceMultiQuestion.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningChoiceMultiQuestionHandler.CreateQuestion(ctx, c.Writer, c.Request)
		}))
		listeningChoiceMultiQuestion.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningChoiceMultiQuestionHandler.UpdateQuestion(ctx, c.Writer, c.Request)
		}))
		listeningChoiceMultiQuestion.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningChoiceMultiQuestionHandler.DeleteQuestion(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Listening - Listening Choice Multi Option
	// ? ------------------------------------------------------------------------------
	listeningChoiceMultiOption := api.Group("/listening/choice-multi-option")
	listeningChoiceMultiOption.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		listeningChoiceMultiOption.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningChoiceMultiOptionHandler.CreateOption(ctx, c.Writer, c.Request)
		}))
		listeningChoiceMultiOption.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningChoiceMultiOptionHandler.UpdateOption(ctx, c.Writer, c.Request)
		}))
		listeningChoiceMultiOption.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningChoiceMultiOptionHandler.DeleteOption(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Listening - Listening Map Labelling
	// ? ------------------------------------------------------------------------------
	listeningMapLabelling := api.Group("/listening/map-labelling")
	listeningMapLabelling.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		listeningMapLabelling.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningMapLabellingHandler.Create(ctx, c.Writer, c.Request)
		}))
		listeningMapLabelling.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningMapLabellingHandler.Update(ctx, c.Writer, c.Request)
		}))
		listeningMapLabelling.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningMapLabellingHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Listening - Listening Matching
	// ? ------------------------------------------------------------------------------
	listeningMatching := api.Group("/listening/matching")
	listeningMatching.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		listeningMatching.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningMatchingHandler.Create(ctx, c.Writer, c.Request)
		}))
		listeningMatching.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningMatchingHandler.Update(ctx, c.Writer, c.Request)
		}))
		listeningMatching.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			listeningMatchingHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ! ------------------------------------------------------------------------------
	// ! - Grammar
	// ! ------------------------------------------------------------------------------
	// ? ------------------------------------------------------------------------------
	// ? - Grammar - GrammarQuestion
	// ? ------------------------------------------------------------------------------
	grammarQuestion := api.Group("/grammar/question")
	grammarQuestion.POST("", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		grammarQuestionHandler.CreateGrammarQuestion(ctx, c.Writer, c.Request)
	}))

	grammarQuestion.GET("/:id", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		grammarQuestionHandler.GetGrammarQuestionDetail(ctx, c.Writer, c.Request)
	}))

	grammarQuestion.PUT("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		grammarQuestionHandler.UpdateGrammarQuestion(ctx, c.Writer, c.Request)
	}))

	grammarQuestion.DELETE("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		grammarQuestionHandler.DeleteGrammarQuestion(ctx, c.Writer, c.Request)
	}))

	grammarQuestion.POST("/get-new-updates", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		grammarQuestionHandler.GetListNewGrammarQuestionByListVersionAndID(ctx, c.Writer, c.Request)
	}))

	grammarQuestion.POST("/get-by-list-id", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		grammarQuestionHandler.GetListGrammarByListID(ctx, c.Writer, c.Request)
	}))

	grammarQuestion.GET("/search", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		grammarQuestionHandler.GetListGrammarQuestiondetailPaginationWithFilter(ctx, c.Writer, c.Request)
	}))

	grammarQuestion.DELETE("/delete-all", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		grammarQuestionHandler.DeleteAllGrammarData(ctx, c.Writer, c.Request)
	}))

	// ? ------------------------------------------------------------------------------
	// ? - grammar - grammar Fill In The Blank
	// ? ------------------------------------------------------------------------------
	grammarFillInTheBlankQuestion := api.Group("/grammar/fill-in-the-blank-question")
	grammarFillInTheBlankQuestion.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		grammarFillInTheBlankQuestion.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarFillInTheBlankQuestionHandler.CreateQuestion(ctx, c.Writer, c.Request)
		}))
		grammarFillInTheBlankQuestion.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarFillInTheBlankQuestionHandler.UpdateQuestion(ctx, c.Writer, c.Request)
		}))
		grammarFillInTheBlankQuestion.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarFillInTheBlankQuestionHandler.DeleteQuestion(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - grammar - grammar Fill In The Blank Answer
	// ? ------------------------------------------------------------------------------
	grammarFillInTheBlankAnswer := api.Group("/grammar/fill-in-the-blank-answer")
	grammarFillInTheBlankAnswer.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		grammarFillInTheBlankAnswer.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarFillInTheBlankAnswerHandler.CreateAnswer(ctx, c.Writer, c.Request)
		}))
		grammarFillInTheBlankAnswer.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarFillInTheBlankAnswerHandler.UpdateAnswer(ctx, c.Writer, c.Request)
		}))
		grammarFillInTheBlankAnswer.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarFillInTheBlankAnswerHandler.DeleteAnswer(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - grammar - grammar Choice One Question
	// ? ------------------------------------------------------------------------------
	grammarChoiceOneQuestion := api.Group("/grammar/choice-one-question")
	grammarChoiceOneQuestion.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		grammarChoiceOneQuestion.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarChoiceOneQuestionHandler.CreateQuestion(ctx, c.Writer, c.Request)
		}))
		grammarChoiceOneQuestion.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarChoiceOneQuestionHandler.UpdateQuestion(ctx, c.Writer, c.Request)
		}))
		grammarChoiceOneQuestion.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarChoiceOneQuestionHandler.DeleteQuestion(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - grammar - grammar Choice One Option
	// ? ------------------------------------------------------------------------------
	grammarChoiceOneOption := api.Group("/grammar/choice-one-option")
	grammarChoiceOneOption.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		grammarChoiceOneOption.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarChoiceOneOptionHandler.CreateOption(ctx, c.Writer, c.Request)
		}))
		grammarChoiceOneOption.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarChoiceOneOptionHandler.UpdateOption(ctx, c.Writer, c.Request)
		}))
		grammarChoiceOneOption.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarChoiceOneOptionHandler.DeleteOption(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Grammar - Error Identification
	// ? ------------------------------------------------------------------------------
	grammarErrorIdentification := api.Group("/grammar/error-identification")
	grammarErrorIdentification.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		grammarErrorIdentification.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarErrorIdentificationHandler.Create(ctx, c.Writer, c.Request)
		}))
		grammarErrorIdentification.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarErrorIdentificationHandler.Update(ctx, c.Writer, c.Request)
		}))
		grammarErrorIdentification.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarErrorIdentificationHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Grammar - Sentence Transformation
	// ? ------------------------------------------------------------------------------
	grammarSentenceTransformation := api.Group("/grammar/sentence-transformation")
	grammarSentenceTransformation.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		grammarSentenceTransformation.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarSentenceTransformationHandler.Create(ctx, c.Writer, c.Request)
		}))
		grammarSentenceTransformation.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarSentenceTransformationHandler.Update(ctx, c.Writer, c.Request)
		}))
		grammarSentenceTransformation.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			grammarSentenceTransformationHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ! ------------------------------------------------------------------------------
	// ! - Reading
	// ! ------------------------------------------------------------------------------
	// ? ------------------------------------------------------------------------------
	// ? - Reading - Reading Question
	// ? ------------------------------------------------------------------------------
	readingQuestion := api.Group("/reading/question")
	readingQuestion.POST("", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		readingQuestionHandler.CreateReadingQuestion(ctx, c.Writer, c.Request)
	}))

	readingQuestion.GET("/:id", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		readingQuestionHandler.GetReadingQuestionDetail(ctx, c.Writer, c.Request)
	}))

	readingQuestion.PUT("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		readingQuestionHandler.UpdateReadingQuestion(ctx, c.Writer, c.Request)
	}))

	readingQuestion.DELETE("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		readingQuestionHandler.DeleteReadingQuestion(ctx, c.Writer, c.Request)
	}))

	readingQuestion.POST("/get-new-updates", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		readingQuestionHandler.GetListNewReadingQuestionByListVersionAndID(ctx, c.Writer, c.Request)
	}))

	readingQuestion.POST("/get-by-list-id", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		readingQuestionHandler.GetListReadingByListID(ctx, c.Writer, c.Request)
	}))

	readingQuestion.GET("/search", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		readingQuestionHandler.GetListReadingQuestiondetailPaganationWithFilter(ctx, c.Writer, c.Request)
	}))

	readingQuestion.DELETE("/delete-all", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		readingQuestionHandler.DeleteAllReadingData(ctx, c.Writer, c.Request)
	}))

	// ? ------------------------------------------------------------------------------
	// ? - Reading - Reading Fill In The Blank Question
	// ? ------------------------------------------------------------------------------
	readingFillInTheBlankQuestion := api.Group("/reading/fill-in-the-blank-question")
	readingFillInTheBlankQuestion.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		readingFillInTheBlankQuestion.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingFillInTheBlankQuestionHandler.CreateQuestion(ctx, c.Writer, c.Request)
		}))
		readingFillInTheBlankQuestion.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingFillInTheBlankQuestionHandler.UpdateQuestion(ctx, c.Writer, c.Request)
		}))
		readingFillInTheBlankQuestion.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingFillInTheBlankQuestionHandler.DeleteQuestion(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Reading - Reading Fill In The Blank Answer
	// ? ------------------------------------------------------------------------------
	readingFillInTheBlankAnswer := api.Group("/reading/fill-in-the-blank-answer")
	readingFillInTheBlankAnswer.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		readingFillInTheBlankAnswer.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingFillInTheBlankAnswerHandler.CreateAnswer(ctx, c.Writer, c.Request)
		}))
		readingFillInTheBlankAnswer.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingFillInTheBlankAnswerHandler.UpdateAnswer(ctx, c.Writer, c.Request)
		}))
		readingFillInTheBlankAnswer.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingFillInTheBlankAnswerHandler.DeleteAnswer(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Reading - Reading Choice One Question
	// ? ------------------------------------------------------------------------------
	readingChoiceOneQuestion := api.Group("/reading/choice-one-question")
	readingChoiceOneQuestion.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		readingChoiceOneQuestion.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingChoiceOneQuestionHandler.CreateQuestion(ctx, c.Writer, c.Request)
		}))
		readingChoiceOneQuestion.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingChoiceOneQuestionHandler.UpdateQuestion(ctx, c.Writer, c.Request)
		}))
		readingChoiceOneQuestion.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingChoiceOneQuestionHandler.DeleteQuestion(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Reading - Reading Choice One Option
	// ? ------------------------------------------------------------------------------
	readingChoiceOneOption := api.Group("/reading/choice-one-option")
	readingChoiceOneOption.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		readingChoiceOneOption.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingChoiceOneOptionHandler.CreateOption(ctx, c.Writer, c.Request)
		}))
		readingChoiceOneOption.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingChoiceOneOptionHandler.UpdateOption(ctx, c.Writer, c.Request)
		}))
		readingChoiceOneOption.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingChoiceOneOptionHandler.DeleteOption(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Reading - Reading Choice Multi Question
	// ? ------------------------------------------------------------------------------
	readingChoiceMultiQuestion := api.Group("/reading/choice-multi-question")
	readingChoiceMultiQuestion.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		readingChoiceMultiQuestion.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingChoiceMultiQuestionHandler.CreateQuestion(ctx, c.Writer, c.Request)
		}))
		readingChoiceMultiQuestion.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingChoiceMultiQuestionHandler.UpdateQuestion(ctx, c.Writer, c.Request)
		}))
		readingChoiceMultiQuestion.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingChoiceMultiQuestionHandler.DeleteQuestion(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Reading - Reading Choice Multi Option
	// ? ------------------------------------------------------------------------------
	readingChoiceMultiOption := api.Group("/reading/choice-multi-option")
	readingChoiceMultiOption.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		readingChoiceMultiOption.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingChoiceMultiOptionHandler.CreateOption(ctx, c.Writer, c.Request)
		}))
		readingChoiceMultiOption.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingChoiceMultiOptionHandler.UpdateOption(ctx, c.Writer, c.Request)
		}))
		readingChoiceMultiOption.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingChoiceMultiOptionHandler.DeleteOption(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Reading - Reading Matching
	// ? ------------------------------------------------------------------------------
	readingMatching := api.Group("/reading/matching")
	readingMatching.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		readingMatching.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingMatchingHandler.Create(ctx, c.Writer, c.Request)
		}))
		readingMatching.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingMatchingHandler.Update(ctx, c.Writer, c.Request)
		}))
		readingMatching.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingMatchingHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Reading - Reading True/False
	// ? ------------------------------------------------------------------------------
	readingTrueFalse := api.Group("/reading/true-false")
	readingTrueFalse.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		readingTrueFalse.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingTrueFalseHandler.Create(ctx, c.Writer, c.Request)
		}))
		readingTrueFalse.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingTrueFalseHandler.Update(ctx, c.Writer, c.Request)
		}))
		readingTrueFalse.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			readingTrueFalseHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ! ------------------------------------------------------------------------------
	// ! - Speaking
	// ! ------------------------------------------------------------------------------
	// ? ------------------------------------------------------------------------------
	// ? - Speaking - SpeakingQuestion
	// ? ------------------------------------------------------------------------------
	speakingQuestion := api.Group("/speaking/question")
	speakingQuestion.POST("", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		speakingQuestionHandler.CreateSpeakingQuestion(ctx, c.Writer, c.Request)
	}))

	speakingQuestion.GET("/:id", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		speakingQuestionHandler.GetSpeakingQuestionDetail(ctx, c.Writer, c.Request)
	}))

	speakingQuestion.PUT("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		speakingQuestionHandler.UpdateSpeakingQuestion(ctx, c.Writer, c.Request)
	}))

	speakingQuestion.DELETE("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		speakingQuestionHandler.DeleteSpeakingQuestion(ctx, c.Writer, c.Request)
	}))

	speakingQuestion.POST("/get-new-updates", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		speakingQuestionHandler.GetListNewSpeakingQuestionByListVersionAndID(ctx, c.Writer, c.Request)
	}))

	speakingQuestion.POST("/get-by-list-id", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		speakingQuestionHandler.GetListSpeakingByListID(ctx, c.Writer, c.Request)
	}))

	speakingQuestion.GET("/search", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		speakingQuestionHandler.GetListSpeakingQuestiondetailPaganationWithFilter(ctx, c.Writer, c.Request)
	}))

	speakingQuestion.DELETE("/delete-all", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		speakingQuestionHandler.DeleteAllSpeakingData(ctx, c.Writer, c.Request)
	}))

	// ? ------------------------------------------------------------------------------
	// ? - Speaking - Word Repetition
	// ? ------------------------------------------------------------------------------
	speakingWordRepetition := api.Group("/speaking/word-repetition")
	speakingWordRepetition.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		speakingWordRepetition.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingWordRepetitionHandler.Create(ctx, c.Writer, c.Request)
		}))
		speakingWordRepetition.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingWordRepetitionHandler.Update(ctx, c.Writer, c.Request)
		}))
		speakingWordRepetition.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingWordRepetitionHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Speaking - Phrase Repetition
	// ? ------------------------------------------------------------------------------
	speakingPhraseRepetition := api.Group("/speaking/phrase-repetition")
	speakingPhraseRepetition.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		speakingPhraseRepetition.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingPhraseRepetitionHandler.Create(ctx, c.Writer, c.Request)
		}))
		speakingPhraseRepetition.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingPhraseRepetitionHandler.Update(ctx, c.Writer, c.Request)
		}))
		speakingPhraseRepetition.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingPhraseRepetitionHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Speaking - Paragraph Repetition
	// ? ------------------------------------------------------------------------------
	speakingParagraphRepetition := api.Group("/speaking/paragraph-repetition")
	speakingParagraphRepetition.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		speakingParagraphRepetition.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingParagraphRepetitionHandler.Create(ctx, c.Writer, c.Request)
		}))
		speakingParagraphRepetition.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingParagraphRepetitionHandler.Update(ctx, c.Writer, c.Request)
		}))
		speakingParagraphRepetition.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingParagraphRepetitionHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Speaking - Open Paragraph
	// ? ------------------------------------------------------------------------------
	speakingOpenParagraph := api.Group("/speaking/open-paragraph")
	speakingOpenParagraph.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		speakingOpenParagraph.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingOpenParagraphHandler.Create(ctx, c.Writer, c.Request)
		}))
		speakingOpenParagraph.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingOpenParagraphHandler.Update(ctx, c.Writer, c.Request)
		}))
		speakingOpenParagraph.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingOpenParagraphHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Speaking - Conversational Repetition
	// ? ------------------------------------------------------------------------------
	speakingConversationalRepetition := api.Group("/speaking/conversational-repetition")
	speakingConversationalRepetition.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		speakingConversationalRepetition.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingConversationalRepetitionHandler.Create(ctx, c.Writer, c.Request)
		}))
		speakingConversationalRepetition.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingConversationalRepetitionHandler.Update(ctx, c.Writer, c.Request)
		}))
		speakingConversationalRepetition.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingConversationalRepetitionHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Speaking - Conversational Repetition QA
	// ? ------------------------------------------------------------------------------
	speakingConversationalRepetitionQA := api.Group("/speaking/conversational-repetition-qa")
	speakingConversationalRepetitionQA.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		speakingConversationalRepetitionQA.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingConversationalRepetitionQAHandler.Create(ctx, c.Writer, c.Request)
		}))
		speakingConversationalRepetitionQA.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingConversationalRepetitionQAHandler.Update(ctx, c.Writer, c.Request)
		}))
		speakingConversationalRepetitionQA.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingConversationalRepetitionQAHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Speaking - Conversational Open
	// ? ------------------------------------------------------------------------------
	speakingConversationalOpen := api.Group("/speaking/conversational-open")
	speakingConversationalOpen.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		speakingConversationalOpen.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingConversationalOpenHandler.Create(ctx, c.Writer, c.Request)
		}))
		speakingConversationalOpen.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingConversationalOpenHandler.Update(ctx, c.Writer, c.Request)
		}))
		speakingConversationalOpen.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			speakingConversationalOpenHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ! ------------------------------------------------------------------------------
	// ! - Writing
	// ! ------------------------------------------------------------------------------
	// ? ------------------------------------------------------------------------------
	// ? - Writing - WritingQuestion
	// ? ------------------------------------------------------------------------------
	writingQuestion := api.Group("/writing/question")
	writingQuestion.POST("", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		writingQuestionHandler.CreateWritingQuestion(ctx, c.Writer, c.Request)
	}))

	writingQuestion.GET("/:id", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		writingQuestionHandler.GetWritingQuestionDetail(ctx, c.Writer, c.Request)
	}))

	writingQuestion.PUT("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		writingQuestionHandler.UpdateWritingQuestion(ctx, c.Writer, c.Request)
	}))

	writingQuestion.DELETE("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		writingQuestionHandler.DeleteWritingQuestion(ctx, c.Writer, c.Request)
	}))

	writingQuestion.POST("/get-new-updates", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		writingQuestionHandler.GetListNewWritingQuestionByListVersionAndID(ctx, c.Writer, c.Request)
	}))

	writingQuestion.POST("/get-by-list-id", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		writingQuestionHandler.GetListWritingByListID(ctx, c.Writer, c.Request)
	}))

	writingQuestion.GET("/search", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		writingQuestionHandler.GetListWritingQuestiondetailPaganationWithFilter(ctx, c.Writer, c.Request)
	}))

	writingQuestion.DELETE("/delete-all", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		writingQuestionHandler.DeleteAllWritingData(ctx, c.Writer, c.Request)
	}))

	// ? ------------------------------------------------------------------------------
	// ? - Writing - Sentence Completion
	// ? ------------------------------------------------------------------------------
	writingSentenceCompletion := api.Group("/writing/sentence-completion")
	writingSentenceCompletion.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		writingSentenceCompletion.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			writingSentenceCompletionHandler.Create(ctx, c.Writer, c.Request)
		}))
		writingSentenceCompletion.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			writingSentenceCompletionHandler.Update(ctx, c.Writer, c.Request)
		}))
		writingSentenceCompletion.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			writingSentenceCompletionHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Writing - Essay
	// ? ------------------------------------------------------------------------------
	writingEssay := api.Group("/writing/essay")
	writingEssay.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		writingEssay.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			writingEssayHandler.Create(ctx, c.Writer, c.Request)
		}))
		writingEssay.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			writingEssayHandler.Update(ctx, c.Writer, c.Request)
		}))
		writingEssay.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			writingEssayHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ! ------------------------------------------------------------------------------
	// ! - Course
	// ! ------------------------------------------------------------------------------
	// ? ------------------------------------------------------------------------------
	// ? - Course - Course
	// ? ------------------------------------------------------------------------------
	course := api.Group("/course")
	course.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		course.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			courseHandler.Create(ctx, c.Writer, c.Request)
		}))
		course.GET("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			courseHandler.GetByID(ctx, c.Writer, c.Request)
		}))
		course.GET("/search", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			courseHandler.Search(ctx, c.Writer, c.Request)
		}))
		course.PUT("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			courseHandler.Update(ctx, c.Writer, c.Request)
		}))
		course.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			courseHandler.Delete(ctx, c.Writer, c.Request)
		}))

	}

	// ? ------------------------------------------------------------------------------
	// ? - Course - Course Book
	// ? ------------------------------------------------------------------------------
	courseBook := api.Group("/course-book")
	courseBook.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		courseBook.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			courseBookHandler.Create(ctx, c.Writer, c.Request)
		}))
		courseBook.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			courseBookHandler.Update(ctx, c.Writer, c.Request)
		}))
		courseBook.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			courseBookHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Course - Course Book
	// ? ------------------------------------------------------------------------------
	courseOther := api.Group("/course-other")
	courseOther.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		courseOther.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			courseOtherHandler.Create(ctx, c.Writer, c.Request)
		}))
		courseOther.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			courseOtherHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Course - Lesson
	// ? ------------------------------------------------------------------------------
	lesson := api.Group("/lesson")
	lesson.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		lesson.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			lessonHandler.Create(ctx, c.Writer, c.Request)
		}))
		lesson.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			lessonHandler.Update(ctx, c.Writer, c.Request)
		}))
		lesson.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			lessonHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// ? ------------------------------------------------------------------------------
	// ? - Course - LessonQuestion
	// ? ------------------------------------------------------------------------------
	lessonQuestion := api.Group("/lesson-question")
	lessonQuestion.Use(middleware.DeveloperAuthMiddleware(r.db))
	{
		lessonQuestion.POST("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			lessonQuestionHandler.Create(ctx, c.Writer, c.Request)
		}))
		lessonQuestion.PUT("", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			lessonQuestionHandler.Update(ctx, c.Writer, c.Request)
		}))
		lessonQuestion.DELETE("/:id", gin.HandlerFunc(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
			lessonQuestionHandler.Delete(ctx, c.Writer, c.Request)
		}))
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})

	// 404 handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, StandardResponse{
			Success: false,
			Error:   "404 - Not Found",
		})
	})

	// 405 handler
	r.NoMethod(func(c *gin.Context) {
		c.JSON(405, StandardResponse{
			Success: false,
			Error:   "405 - Method Not Allowed",
		})
	})

	// Print all routes in debug mode
	pwd, _ := os.Getwd()
	routes := r.Engine.Routes()
	r.logger.Info("ROUTES_REGISTERED", map[string]interface{}{
		"working_directory": pwd,
		"total_routes":      len(routes),
	}, "Registered routes in debug mode")
}
