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

type SpeakingOpenParagraphHandler struct {
	service *speakingService.SpeakingOpenParagraphService
	logger  *logger.PrettyLogger
}

func NewSpeakingOpenParagraphHandler(
	service *speakingService.SpeakingOpenParagraphService,
	logger *logger.PrettyLogger,
) *SpeakingOpenParagraphHandler {
	return &SpeakingOpenParagraphHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SpeakingOpenParagraphHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.CreateSpeakingOpenParagraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_open_paragraph_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	openParagraph := &speaking.SpeakingOpenParagraph{
		ID:                   uuid.New(),
		SpeakingQuestionID:   req.SpeakingQuestionID,
		Question:             req.Question,
		ExamplePassage:       req.ExamplePassage,
		MeanOfExamplePassage: req.MeanOfExamplePassage,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := h.service.Create(ctx, openParagraph); err != nil {
		h.logger.Error("speaking_open_paragraph_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create open paragraph")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create open paragraph")
		return
	}

	responseData := speakingDTO.SpeakingOpenParagraphResponse{
		ID:                   openParagraph.ID,
		Question:             openParagraph.Question,
		ExamplePassage:       openParagraph.ExamplePassage,
		MeanOfExamplePassage: openParagraph.MeanOfExamplePassage,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *SpeakingOpenParagraphHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.UpdateSpeakingOpenParagraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_open_paragraph_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingParagraph, err := h.service.GetByID(ctx, req.SpeakingOpenParagraphID)
	if err != nil {
		h.logger.Error("speaking_open_paragraph_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingOpenParagraphID,
		}, "Failed to get existing open paragraph")
		response.WriteError(w, http.StatusNotFound, "Open paragraph not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "question":
		existingParagraph.Question = req.Value
	case "example_passage":
		existingParagraph.ExamplePassage = req.Value
	case "mean_of_example_passage":
		existingParagraph.MeanOfExamplePassage = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingParagraph.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingParagraph); err != nil {
		h.logger.Error("speaking_open_paragraph_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingOpenParagraphID,
		}, "Failed to update open paragraph")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update open paragraph")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *SpeakingOpenParagraphHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("speaking_open_paragraph_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("speaking_open_paragraph_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid open paragraph ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid open paragraph ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("speaking_open_paragraph_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete open paragraph")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete open paragraph")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Open paragraph deleted successfully"})
}
