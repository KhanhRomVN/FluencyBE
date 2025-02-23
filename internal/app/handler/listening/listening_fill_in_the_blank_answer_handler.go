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

type ListeningFillInTheBlankAnswerHandler struct {
	service *listeningService.ListeningFillInTheBlankAnswerService
	logger  *logger.PrettyLogger
}

func NewListeningFillInTheBlankAnswerHandler(
	service *listeningService.ListeningFillInTheBlankAnswerService,
	logger *logger.PrettyLogger,
) *ListeningFillInTheBlankAnswerHandler {
	return &ListeningFillInTheBlankAnswerHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ListeningFillInTheBlankAnswerHandler) CreateAnswer(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req listeningDTO.CreateListeningFillInTheBlankAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_fill_in_the_blank_answer_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	answer := &listening.ListeningFillInTheBlankAnswer{
		ID:                                uuid.New(),
		ListeningFillInTheBlankQuestionID: req.ListeningFillInTheBlankQuestionID,
		Answer:                            req.Answer,
		Explain:                           req.Explain,
		CreatedAt:                         time.Now(),
		UpdatedAt:                         time.Now(),
	}

	if err := h.service.CreateAnswer(ctx, answer); err != nil {
		h.logger.Error("listening_fill_in_the_blank_answer_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create answer")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create answer")
		return
	}

	responseData := listeningDTO.ListeningFillInTheBlankAnswerResponse{
		ID:      answer.ID,
		Answer:  answer.Answer,
		Explain: answer.Explain,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *ListeningFillInTheBlankAnswerHandler) UpdateAnswer(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req listeningDTO.UpdateListeningFillInTheBlankAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_fill_in_the_blank_answer_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	answer, err := h.service.GetAnswer(ctx, req.ListeningFillInTheBlankAnswerID)
	if err != nil {
		h.logger.Error("listening_fill_in_the_blank_answer_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ListeningFillInTheBlankAnswerID,
		}, "Failed to get answer")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Answer not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to get answer")
		return
	}

	if req.Field == "answer" {
		answer.Answer = req.Value
	} else if req.Field == "explain" {
		answer.Explain = req.Value
	} else {
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}
	answer.UpdatedAt = time.Now()

	if err := h.service.UpdateAnswer(ctx, answer); err != nil {
		h.logger.Error("listening_fill_in_the_blank_answer_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ListeningFillInTheBlankAnswerID,
		}, "Failed to update answer")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update answer")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *ListeningFillInTheBlankAnswerHandler) DeleteAnswer(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("listening_fill_in_the_blank_answer_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("listening_fill_in_the_blank_answer_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid answer ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid answer ID")
		return
	}

	if err := h.service.DeleteAnswer(ctx, id); err != nil {
		h.logger.Error("listening_fill_in_the_blank_answer_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete answer")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Answer not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete answer")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Answer deleted successfully"})
}
