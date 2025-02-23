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

type SpeakingParagraphRepetitionHandler struct {
	service *speakingService.SpeakingParagraphRepetitionService
	logger  *logger.PrettyLogger
}

func NewSpeakingParagraphRepetitionHandler(
	service *speakingService.SpeakingParagraphRepetitionService,
	logger *logger.PrettyLogger,
) *SpeakingParagraphRepetitionHandler {
	return &SpeakingParagraphRepetitionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SpeakingParagraphRepetitionHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.CreateSpeakingParagraphRepetitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_paragraph_repetition_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	paragraphRepetition := &speaking.SpeakingParagraphRepetition{
		ID:                 uuid.New(),
		SpeakingQuestionID: req.SpeakingQuestionID,
		Paragraph:          req.Paragraph,
		Mean:               req.Mean,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := h.service.Create(ctx, paragraphRepetition); err != nil {
		h.logger.Error("speaking_paragraph_repetition_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create paragraph repetition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create paragraph repetition")
		return
	}

	responseData := speakingDTO.SpeakingParagraphRepetitionResponse{
		ID:        paragraphRepetition.ID,
		Paragraph: paragraphRepetition.Paragraph,
		Mean:      paragraphRepetition.Mean,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *SpeakingParagraphRepetitionHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.UpdateSpeakingParagraphRepetitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_paragraph_repetition_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingParagraph, err := h.service.GetByID(ctx, req.SpeakingParagraphRepetitionID)
	if err != nil {
		h.logger.Error("speaking_paragraph_repetition_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingParagraphRepetitionID,
		}, "Failed to get existing paragraph repetition")
		response.WriteError(w, http.StatusNotFound, "Paragraph repetition not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "paragraph":
		existingParagraph.Paragraph = req.Value
	case "mean":
		existingParagraph.Mean = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingParagraph.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingParagraph); err != nil {
		h.logger.Error("speaking_paragraph_repetition_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingParagraphRepetitionID,
		}, "Failed to update paragraph repetition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update paragraph repetition")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *SpeakingParagraphRepetitionHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("speaking_paragraph_repetition_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("speaking_paragraph_repetition_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid paragraph repetition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid paragraph repetition ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("speaking_paragraph_repetition_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete paragraph repetition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete paragraph repetition")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Paragraph repetition deleted successfully"})
}
