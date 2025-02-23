package listening

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	listeningDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/listening"
	listeningService "fluencybe/internal/app/service/listening"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ListeningQuestionHandler struct {
	service                    *listeningService.ListeningQuestionService
	fillInBlankQuestionService *listeningService.ListeningFillInTheBlankQuestionService
	fillInBlankAnswerService   *listeningService.ListeningFillInTheBlankAnswerService
	choiceOneQuestionService   *listeningService.ListeningChoiceOneQuestionService
	choiceOneOptionService     *listeningService.ListeningChoiceOneOptionService
	choiceMultiQuestionService *listeningService.ListeningChoiceMultiQuestionService
	choiceMultiOptionService   *listeningService.ListeningChoiceMultiOptionService
	mapLabellingService        *listeningService.ListeningMapLabellingService
	matchingService            *listeningService.ListeningMatchingService
	logger                     *logger.PrettyLogger
}

func NewListeningQuestionHandler(
	service *listeningService.ListeningQuestionService,
	fillInBlankQuestionService *listeningService.ListeningFillInTheBlankQuestionService,
	fillInBlankAnswerService *listeningService.ListeningFillInTheBlankAnswerService,
	choiceOneQuestionService *listeningService.ListeningChoiceOneQuestionService,
	choiceOneOptionService *listeningService.ListeningChoiceOneOptionService,
	choiceMultiQuestionService *listeningService.ListeningChoiceMultiQuestionService,
	choiceMultiOptionService *listeningService.ListeningChoiceMultiOptionService,
	mapLabellingService *listeningService.ListeningMapLabellingService,
	matchingService *listeningService.ListeningMatchingService,
	logger *logger.PrettyLogger,
) *ListeningQuestionHandler {
	return &ListeningQuestionHandler{
		service:                    service,
		fillInBlankQuestionService: fillInBlankQuestionService,
		fillInBlankAnswerService:   fillInBlankAnswerService,
		choiceOneQuestionService:   choiceOneQuestionService,
		choiceOneOptionService:     choiceOneOptionService,
		choiceMultiQuestionService: choiceMultiQuestionService,
		choiceMultiOptionService:   choiceMultiOptionService,
		mapLabellingService:        mapLabellingService,
		matchingService:            matchingService,
		logger:                     logger,
	}
}

func (h *ListeningQuestionHandler) CreateListeningQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req listeningDTO.CreateListeningQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_question_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Invalid request format")
		response.WriteError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	question := &listening.ListeningQuestion{
		ID:          uuid.New(),
		Type:        req.Type,
		Topic:       req.Topic,
		Instruction: req.Instruction,
		AudioURLs:   req.AudioURLs,
		ImageURLs:   req.ImageURLs,
		Transcript:  req.Transcript,
		MaxTime:     req.MaxTime,
		Version:     1,
	}

	if err := h.service.CreateQuestion(ctx, question); err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, listeningService.ErrInvalidInput) {
			code = http.StatusBadRequest
		}
		h.logger.Error("listening_question_handler.create", map[string]interface{}{
			"error":         err.Error(),
			"question_type": question.Type,
		}, "Failed to create listening question")
		response.WriteError(w, code, "Failed to create listening question")
		return
	}

	responseData := listeningDTO.ListeningQuestionResponse{
		ID:          question.ID,
		Type:        question.Type,
		Topic:       question.Topic,
		Instruction: question.Instruction,
		AudioURLs:   question.AudioURLs,
		ImageURLs:   question.ImageURLs,
		Transcript:  question.Transcript,
		MaxTime:     question.MaxTime,
		Version:     question.Version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseData)
}

func (h *ListeningQuestionHandler) GetListeningQuestionDetail(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtxValue := ctx.Value(constants.GinContextKey)
	if ginCtxValue == nil {
		h.logger.Error("listening_question_handler.get", nil, "GinContextKey not found in context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	ginCtx, ok := ginCtxValue.(*gin.Context)
	if !ok {
		h.logger.Error("listening_question_handler.get", nil, "Failed to convert context value to *gin.Context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("listening_question_handler.get.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID format")
		return
	}

	responseData, err := h.service.GetListeningQuestionDetail(ctx, id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, listeningService.ErrQuestionNotFound) {
			statusCode = http.StatusNotFound
		}
		h.logger.Error("listening_question_handler.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get listening question")
		response.WriteError(w, statusCode, "Failed to get listening question")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

func (h *ListeningQuestionHandler) UpdateListeningQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtxValue := ctx.Value(constants.GinContextKey)
	if ginCtxValue == nil {
		h.logger.Error("listening_question_handler.update", nil, "GinContextKey not found in context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	ginCtx, ok := ginCtxValue.(*gin.Context)
	if !ok {
		h.logger.Error("listening_question_handler.update", nil, "Failed to convert context value to *gin.Context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("listening_question_handler.update.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID format")
		return
	}

	var req listeningDTO.UpdateListeningQuestionFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_question_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.UpdateQuestion(ctx, id, req); err != nil {
		h.logger.Error("listening_question_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update listening question")

		statusCode := http.StatusInternalServerError
		if errors.Is(err, listeningService.ErrQuestionNotFound) {
			statusCode = http.StatusNotFound
		}
		response.WriteError(w, statusCode, "Failed to update listening question")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ListeningQuestionHandler) DeleteListeningQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtxValue := ctx.Value(constants.GinContextKey)
	if ginCtxValue == nil {
		h.logger.Error("listening_question_handler.delete", nil, "GinContextKey not found in context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	ginCtx, ok := ginCtxValue.(*gin.Context)
	if !ok {
		h.logger.Error("listening_question_handler.delete", nil, "Failed to convert context value to *gin.Context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("listening_question_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID format")
		return
	}

	if err := h.service.DeleteQuestion(ctx, id); err != nil {
		h.logger.Error("listening_question_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete listening question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete listening question")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ListeningQuestionHandler) GetListNewListeningQuestionByListVersionAndID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req listeningDTO.GetNewUpdatesListeningQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_question_handler.get_new_updates.decode", map[string]interface{}{
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
		versionChecks[i].ID = q.ListeningQuestionID
		versionChecks[i].Version = q.Version
	}

	questions, err := h.service.GetNewUpdatedQuestions(ctx, versionChecks)
	if err != nil {
		h.logger.Error("listening_question_handler.get_new_listening_questions", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get updated questions")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get updated questions")
		return
	}

	// Convert to response format
	responseData := make([]listeningDTO.ListeningQuestionDetail, len(questions))
	for i, q := range questions {
		responseData[i] = listeningDTO.ListeningQuestionDetail{
			ListeningQuestionResponse: listeningDTO.ListeningQuestionResponse{
				ID:          q.ID,
				Type:        q.Type,
				Topic:       q.Topic,
				Instruction: q.Instruction,
				AudioURLs:   q.AudioURLs,
				ImageURLs:   q.ImageURLs,
				Transcript:  q.Transcript,
				MaxTime:     q.MaxTime,
			},
		}
	}

	response.WriteJSON(w, http.StatusOK, gin.H{
		"success": true,
		"data":    responseData,
	})
}

func (h *ListeningQuestionHandler) GetService() *listeningService.ListeningQuestionService {
	return h.service
}

type GetListListeningByListIDRequest struct {
	QuestionIDs []string `json:"question_ids" validate:"required,min=1"`
}

func (h *ListeningQuestionHandler) GetListListeningByListID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req GetListListeningByListIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_question_handler.get_list_by_ids.decode", map[string]interface{}{
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
			h.logger.Error("listening_question_handler.get_list_by_ids.parse_id", map[string]interface{}{
				"error": err.Error(),
				"id":    idStr,
			}, "Invalid question ID format")
			response.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid question ID format: %s", idStr))
			return
		}
		questionIDs = append(questionIDs, id)
	}

	// Get questions with details
	questions, err := h.service.GetListeningByListID(ctx, questionIDs)
	if err != nil {
		h.logger.Error("listening_question_handler.get_list_by_ids", map[string]interface{}{
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

func (h *ListeningQuestionHandler) GetListListeningQuestiondetailPaganationWithFilter(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("listening_question_handler.search.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var filter listeningDTO.ListeningQuestionSearchFilter
	if err := ginCtx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("listening_question_handler.search.bind", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to bind query parameters")
		response.WriteError(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	questions, err := h.service.SearchQuestionsWithFilter(ctx, filter)
	if err != nil {
		h.logger.Error("listening_question_handler.search", map[string]interface{}{
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

func (h *ListeningQuestionHandler) DeleteAllListeningData(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteAllQuestions(ctx); err != nil {
		h.logger.Error("listening_question_handler.delete_all", map[string]interface{}{"error": err.Error()}, "Failed to delete all listening data")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete all listening data")
		return
	}
	response.WriteJSON(w, http.StatusOK, gin.H{"message": "All listening data deleted successfully"})
}
