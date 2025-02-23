package listening

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	listeningDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/listening"
	listeningService "fluencybe/internal/app/service/listening"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ListeningChoiceOneQuestionHandler struct {
	questionService *listeningService.ListeningChoiceOneQuestionService
	optionService   *listeningService.ListeningChoiceOneOptionService
	logger          *logger.PrettyLogger
}

func NewListeningChoiceOneQuestionHandler(
	questionService *listeningService.ListeningChoiceOneQuestionService,
	optionService *listeningService.ListeningChoiceOneOptionService,
	logger *logger.PrettyLogger,
) *ListeningChoiceOneQuestionHandler {
	return &ListeningChoiceOneQuestionHandler{
		questionService: questionService,
		optionService:   optionService,
		logger:          logger,
	}
}

func (h *ListeningChoiceOneQuestionHandler) CreateQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req listeningDTO.CreateListeningChoiceOneQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_choice_one_question_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	question := &listening.ListeningChoiceOneQuestion{
		ID:                  uuid.New(),
		ListeningQuestionID: req.ListeningQuestionID,
		Question:            req.Question,
		Explain:             req.Explain,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := h.questionService.CreateQuestion(ctx, question); err != nil {
		h.logger.Error("listening_choice_one_question_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create question")
		return
	}

	responseData := listeningDTO.ListeningChoiceOneQuestionResponse{
		ID:       question.ID,
		Question: question.Question,
		Explain:  question.Explain,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *ListeningChoiceOneQuestionHandler) UpdateQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req listeningDTO.UpdateListeningChoiceOneQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_choice_one_question_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	question, err := h.questionService.GetQuestion(ctx, req.ListeningChoiceOneQuestionID)
	if err != nil {
		h.logger.Error("listening_choice_one_question_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ListeningChoiceOneQuestionID,
		}, "Failed to get question")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Question not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to get question")
		return
	}

	switch req.Field {
	case "question":
		question.Question = req.Value
	case "explain":
		question.Explain = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	question.UpdatedAt = time.Now()

	if err := h.questionService.UpdateQuestion(ctx, question); err != nil {
		h.logger.Error("listening_choice_one_question_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ListeningChoiceOneQuestionID,
		}, "Failed to update question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update question")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *ListeningChoiceOneQuestionHandler) DeleteQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("listening_choice_one_question_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("listening_choice_one_question_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID")
		return
	}

	if err := h.questionService.DeleteQuestion(ctx, id); err != nil {
		h.logger.Error("listening_choice_one_question_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Question not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete question")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Question deleted successfully"})
}
