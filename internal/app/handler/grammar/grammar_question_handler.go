package grammar

import (
	"context"
	"encoding/json"
	"errors"

	grammarDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/grammar"
	grammarService "fluencybe/internal/app/service/grammar"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GrammarQuestionHandler struct {
	service                       *grammarService.GrammarQuestionService
	fillInBlankQuestionService    *grammarService.GrammarFillInTheBlankQuestionService
	fillInBlankAnswerService      *grammarService.GrammarFillInTheBlankAnswerService
	choiceOneQuestionService      *grammarService.GrammarChoiceOneQuestionService
	choiceOneOptionService        *grammarService.GrammarChoiceOneOptionService
	errorIdentificationService    *grammarService.GrammarErrorIdentificationService
	sentenceTransformationService *grammarService.GrammarSentenceTransformationService
	logger                        *logger.PrettyLogger
}

func NewGrammarQuestionHandler(
	service *grammarService.GrammarQuestionService,
	fillInBlankQuestionService *grammarService.GrammarFillInTheBlankQuestionService,
	fillInBlankAnswerService *grammarService.GrammarFillInTheBlankAnswerService,
	choiceOneQuestionService *grammarService.GrammarChoiceOneQuestionService,
	choiceOneOptionService *grammarService.GrammarChoiceOneOptionService,
	errorIdentificationService *grammarService.GrammarErrorIdentificationService,
	sentenceTransformationService *grammarService.GrammarSentenceTransformationService,
	logger *logger.PrettyLogger,
) *GrammarQuestionHandler {
	return &GrammarQuestionHandler{
		service:                       service,
		fillInBlankQuestionService:    fillInBlankQuestionService,
		fillInBlankAnswerService:      fillInBlankAnswerService,
		choiceOneQuestionService:      choiceOneQuestionService,
		choiceOneOptionService:        choiceOneOptionService,
		errorIdentificationService:    errorIdentificationService,
		sentenceTransformationService: sentenceTransformationService,
		logger:                        logger,
	}
}

func (h *GrammarQuestionHandler) CreateGrammarQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req grammarDTO.CreateGrammarQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("grammar_question_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Invalid request format")
		response.WriteError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	question := &grammar.GrammarQuestion{
		ID:          uuid.New(),
		Type:        grammar.GrammarQuestionType(req.Type),
		Topic:       req.Topic,
		Instruction: req.Instruction,
		ImageURLs:   req.ImageURLs,
		MaxTime:     req.MaxTime,
		Version:     1,
	}

	if err := h.service.CreateQuestion(ctx, question); err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, grammarService.ErrInvalidInput) {
			code = http.StatusBadRequest
		}
		h.logger.Error("grammar_question_handler.create", map[string]interface{}{
			"error":         err.Error(),
			"question_type": question.Type,
		}, "Failed to create grammar question")
		response.WriteError(w, code, "Failed to create grammar question")
		return
	}

	responseData := grammarDTO.GrammarQuestionResponse{
		ID:          question.ID,
		Type:        string(question.Type),
		Topic:       question.Topic,
		Instruction: question.Instruction,
		ImageURLs:   question.ImageURLs,
		MaxTime:     question.MaxTime,
		Version:     question.Version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseData)
}

func (h *GrammarQuestionHandler) GetGrammarQuestionDetail(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtxValue := ctx.Value(constants.GinContextKey)
	if ginCtxValue == nil {
		h.logger.Error("grammar_question_handler.get", nil, "GinContextKey not found in context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	ginCtx, ok := ginCtxValue.(*gin.Context)
	if !ok {
		h.logger.Error("grammar_question_handler.get", nil, "Failed to convert context value to *gin.Context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("grammar_question_handler.get.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID format")
		return
	}

	responseData, err := h.service.GetGrammarQuestionDetail(ctx, id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, grammarService.ErrQuestionNotFound) {
			statusCode = http.StatusNotFound
		}
		h.logger.Error("grammar_question_handler.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get grammar question")
		response.WriteError(w, statusCode, "Failed to get grammar question")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

func (h *GrammarQuestionHandler) UpdateGrammarQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtxValue := ctx.Value(constants.GinContextKey)
	if ginCtxValue == nil {
		h.logger.Error("grammar_question_handler.update", nil, "GinContextKey not found in context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	ginCtx, ok := ginCtxValue.(*gin.Context)
	if !ok {
		h.logger.Error("grammar_question_handler.update", nil, "Failed to convert context value to *gin.Context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("grammar_question_handler.update.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID format")
		return
	}

	var req grammarDTO.UpdateGrammarQuestionFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("grammar_question_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.UpdateQuestion(ctx, id, req); err != nil {
		h.logger.Error("grammar_question_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update grammar question")

		statusCode := http.StatusInternalServerError
		if errors.Is(err, grammarService.ErrQuestionNotFound) {
			statusCode = http.StatusNotFound
		}
		response.WriteError(w, statusCode, "Failed to update grammar question")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GrammarQuestionHandler) DeleteGrammarQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtxValue := ctx.Value(constants.GinContextKey)
	if ginCtxValue == nil {
		h.logger.Error("grammar_question_handler.delete", nil, "GinContextKey not found in context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	ginCtx, ok := ginCtxValue.(*gin.Context)
	if !ok {
		h.logger.Error("grammar_question_handler.delete", nil, "Failed to convert context value to *gin.Context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("grammar_question_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID format")
		return
	}

	if err := h.service.DeleteQuestion(ctx, id); err != nil {
		h.logger.Error("grammar_question_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete grammar question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete grammar question")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GrammarQuestionHandler) GetListNewGrammarQuestionByListVersionAndID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req grammarDTO.GetNewUpdatesGrammarQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("grammar_question_handler.get_new_updates.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert request to internal format
	versionChecks := make([]struct {
		ID      uuid.UUID
		Version int
	}, len(req.Questions))
	for i, q := range req.Questions {
		versionChecks[i].ID = q.GrammarQuestionID
		versionChecks[i].Version = q.Version
	}

	questions, err := h.service.GetNewUpdatedQuestions(ctx, versionChecks)
	if err != nil {
		h.logger.Error("grammar_question_handler.get_new_grammar_questions", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get updated questions")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get updated questions")
		return
	}

	// Convert to response format
	responseData := make([]grammarDTO.GrammarQuestionDetail, len(questions))
	for i, q := range questions {
		responseData[i] = grammarDTO.GrammarQuestionDetail{
			GrammarQuestionResponse: grammarDTO.GrammarQuestionResponse{
				ID:          q.ID,
				Type:        string(q.Type),
				Topic:       q.Topic,
				Instruction: q.Instruction,
				ImageURLs:   q.ImageURLs,
				MaxTime:     q.MaxTime,
			},
		}
	}

	response.WriteJSON(w, http.StatusOK, gin.H{
		"success": true,
		"data":    responseData,
	})
}

func (h *GrammarQuestionHandler) GetService() *grammarService.GrammarQuestionService {
	return h.service
}

type GetListGrammarByListIDRequest struct {
	QuestionIDs []string `json:"question_ids" validate:"required,min=1"`
}

func (h *GrammarQuestionHandler) GetListGrammarByListID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req GetListGrammarByListIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("grammar_question_handler.get_list_by_ids.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert string IDs to UUIDs
	questionIDs := make([]uuid.UUID, 0, len(req.QuestionIDs))
	for _, idStr := range req.QuestionIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.logger.Error("grammar_question_handler.get_list_by_ids.parse_id", map[string]interface{}{
				"error": err.Error(),
				"id":    idStr,
			}, "Invalid question ID format")
			response.WriteError(w, http.StatusBadRequest, "Invalid question ID format")
			return
		}
		questionIDs = append(questionIDs, id)
	}

	// Get questions with details
	questions, err := h.service.GetGrammarByListID(ctx, questionIDs)
	if err != nil {
		h.logger.Error("grammar_question_handler.get_list_by_ids", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get questions")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get questions")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{
		"success": true,
		"data":    questions,
	})
}

func (h *GrammarQuestionHandler) GetListGrammarQuestiondetailPaginationWithFilter(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("grammar_question_handler.search.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var filter grammarDTO.GrammarQuestionSearchFilter
	if err := ginCtx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("grammar_question_handler.search.bind", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to bind query parameters")
		response.WriteError(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	questions, err := h.service.SearchQuestionsWithFilter(ctx, filter)
	if err != nil {
		h.logger.Error("grammar_question_handler.search", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to search questions")
		response.WriteError(w, http.StatusInternalServerError, "Failed to search questions")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{
		"success": true,
		"data":    questions,
	})
}

func (h *GrammarQuestionHandler) DeleteAllGrammarData(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteAllQuestions(ctx); err != nil {
		h.logger.Error("grammar_question_handler.delete_all", map[string]interface{}{"error": err.Error()}, "Failed to delete all grammar data")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete all grammar data")
		return
	}
	response.WriteJSON(w, http.StatusOK, gin.H{"message": "All grammar data deleted successfully"})
}
