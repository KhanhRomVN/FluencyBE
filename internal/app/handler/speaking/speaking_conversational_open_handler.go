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

type SpeakingConversationalOpenHandler struct {
	service *speakingService.SpeakingConversationalOpenService
	logger  *logger.PrettyLogger
}

func NewSpeakingConversationalOpenHandler(
	service *speakingService.SpeakingConversationalOpenService,
	logger *logger.PrettyLogger,
) *SpeakingConversationalOpenHandler {
	return &SpeakingConversationalOpenHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SpeakingConversationalOpenHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.CreateSpeakingConversationalOpenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_conversational_open_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	conversationalOpen := &speaking.SpeakingConversationalOpen{
		ID:                  uuid.New(),
		SpeakingQuestionID:  req.SpeakingQuestionID,
		Title:               req.Title,
		Overview:            req.Overview,
		ExampleConversation: req.ExampleConversation,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := h.service.Create(ctx, conversationalOpen); err != nil {
		h.logger.Error("speaking_conversational_open_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create conversational open")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create conversational open")
		return
	}

	responseData := speakingDTO.SpeakingConversationalOpenResponse{
		ID:                  conversationalOpen.ID,
		Title:               conversationalOpen.Title,
		Overview:            conversationalOpen.Overview,
		ExampleConversation: conversationalOpen.ExampleConversation,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *SpeakingConversationalOpenHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.UpdateSpeakingConversationalOpenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_conversational_open_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingOpen, err := h.service.GetByID(ctx, req.SpeakingConversationalOpenID)
	if err != nil {
		h.logger.Error("speaking_conversational_open_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingConversationalOpenID,
		}, "Failed to get existing conversational open")
		response.WriteError(w, http.StatusNotFound, "Conversational open not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "title":
		existingOpen.Title = req.Value
	case "overview":
		existingOpen.Overview = req.Value
	case "example_conversation":
		existingOpen.ExampleConversation = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingOpen.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingOpen); err != nil {
		h.logger.Error("speaking_conversational_open_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingConversationalOpenID,
		}, "Failed to update conversational open")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update conversational open")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *SpeakingConversationalOpenHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("speaking_conversational_open_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("speaking_conversational_open_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid conversational open ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid conversational open ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("speaking_conversational_open_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete conversational open")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete conversational open")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Conversational open deleted successfully"})
}
