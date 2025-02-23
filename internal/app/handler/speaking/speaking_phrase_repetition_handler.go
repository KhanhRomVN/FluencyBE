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

type SpeakingPhraseRepetitionHandler struct {
	service *speakingService.SpeakingPhraseRepetitionService
	logger  *logger.PrettyLogger
}

func NewSpeakingPhraseRepetitionHandler(
	service *speakingService.SpeakingPhraseRepetitionService,
	logger *logger.PrettyLogger,
) *SpeakingPhraseRepetitionHandler {
	return &SpeakingPhraseRepetitionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SpeakingPhraseRepetitionHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.CreateSpeakingPhraseRepetitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_phrase_repetition_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	phraseRepetition := &speaking.SpeakingPhraseRepetition{
		ID:                 uuid.New(),
		SpeakingQuestionID: req.SpeakingQuestionID,
		Phrase:             req.Phrase,
		Mean:               req.Mean,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := h.service.Create(ctx, phraseRepetition); err != nil {
		h.logger.Error("speaking_phrase_repetition_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create phrase repetition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create phrase repetition")
		return
	}

	responseData := speakingDTO.SpeakingPhraseRepetitionResponse{
		ID:     phraseRepetition.ID,
		Phrase: phraseRepetition.Phrase,
		Mean:   phraseRepetition.Mean,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *SpeakingPhraseRepetitionHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.UpdateSpeakingPhraseRepetitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_phrase_repetition_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingPhrase, err := h.service.GetByID(ctx, req.SpeakingPhraseRepetitionID)
	if err != nil {
		h.logger.Error("speaking_phrase_repetition_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingPhraseRepetitionID,
		}, "Failed to get existing phrase repetition")
		response.WriteError(w, http.StatusNotFound, "Phrase repetition not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "phrase":
		existingPhrase.Phrase = req.Value
	case "mean":
		existingPhrase.Mean = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingPhrase.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingPhrase); err != nil {
		h.logger.Error("speaking_phrase_repetition_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingPhraseRepetitionID,
		}, "Failed to update phrase repetition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update phrase repetition")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *SpeakingPhraseRepetitionHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("speaking_phrase_repetition_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("speaking_phrase_repetition_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid phrase repetition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid phrase repetition ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("speaking_phrase_repetition_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete phrase repetition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete phrase repetition")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Phrase repetition deleted successfully"})
}
