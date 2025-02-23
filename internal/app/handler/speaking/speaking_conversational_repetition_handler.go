package speaking

import (
	"context"
	"encoding/json"
	speakingDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/speaking"
	speakingService "fluencybe/internal/app/service/speaking"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SpeakingConversationalRepetitionHandler struct {
	service *speakingService.SpeakingConversationalRepetitionService
	logger  *logger.PrettyLogger
}

func NewSpeakingConversationalRepetitionHandler(
	service *speakingService.SpeakingConversationalRepetitionService,
	logger *logger.PrettyLogger,
) *SpeakingConversationalRepetitionHandler {
	return &SpeakingConversationalRepetitionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SpeakingConversationalRepetitionHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.CreateSpeakingConversationalRepetitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_conversational_repetition_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	conversationalRepetition := &speaking.SpeakingConversationalRepetition{
		ID:                 uuid.New(),
		SpeakingQuestionID: req.SpeakingQuestionID,
		Title:              req.Title,
		Overview:           req.Overview,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := h.service.Create(ctx, conversationalRepetition); err != nil {
		h.logger.Error("speaking_conversational_repetition_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create conversational repetition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create conversational repetition")
		return
	}

	responseData := speakingDTO.SpeakingConversationalRepetitionResponse{
		ID:       conversationalRepetition.ID,
		Title:    conversationalRepetition.Title,
		Overview: conversationalRepetition.Overview,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *SpeakingConversationalRepetitionHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.UpdateSpeakingConversationalRepetitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_conversational_repetition_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingRepetition, err := h.service.GetByID(ctx, req.SpeakingConversationalRepetitionID)
	if err != nil {
		h.logger.Error("speaking_conversational_repetition_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingConversationalRepetitionID,
		}, "Failed to get existing conversational repetition")
		response.WriteError(w, http.StatusNotFound, "Conversational repetition not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "title":
		existingRepetition.Title = req.Value
	case "overview":
		existingRepetition.Overview = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingRepetition.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingRepetition); err != nil {
		h.logger.Error("speaking_conversational_repetition_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingConversationalRepetitionID,
		}, "Failed to update conversational repetition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update conversational repetition")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *SpeakingConversationalRepetitionHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("speaking_conversational_repetition_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("speaking_conversational_repetition_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid conversational repetition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid conversational repetition ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("speaking_conversational_repetition_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete conversational repetition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete conversational repetition")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Conversational repetition deleted successfully"})
}
