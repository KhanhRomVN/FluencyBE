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

type SpeakingWordRepetitionHandler struct {
	service *speakingService.SpeakingWordRepetitionService
	logger  *logger.PrettyLogger
}

func NewSpeakingWordRepetitionHandler(
	service *speakingService.SpeakingWordRepetitionService,
	logger *logger.PrettyLogger,
) *SpeakingWordRepetitionHandler {
	return &SpeakingWordRepetitionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SpeakingWordRepetitionHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.CreateSpeakingWordRepetitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_word_repetition_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	wordRepetition := &speaking.SpeakingWordRepetition{
		ID:                 uuid.New(),
		SpeakingQuestionID: req.SpeakingQuestionID,
		Word:               req.Word,
		Mean:               req.Mean,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := h.service.Create(ctx, wordRepetition); err != nil {
		h.logger.Error("speaking_word_repetition_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create word repetition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create word repetition")
		return
	}

	responseData := speakingDTO.SpeakingWordRepetitionResponse{
		ID:   wordRepetition.ID,
		Word: wordRepetition.Word,
		Mean: wordRepetition.Mean,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *SpeakingWordRepetitionHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.UpdateSpeakingWordRepetitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_word_repetition_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingWord, err := h.service.GetByID(ctx, req.SpeakingWordRepetitionID)
	if err != nil {
		h.logger.Error("speaking_word_repetition_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingWordRepetitionID,
		}, "Failed to get existing word repetition")
		response.WriteError(w, http.StatusNotFound, "Word repetition not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "word":
		existingWord.Word = req.Value
	case "mean":
		existingWord.Mean = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingWord.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingWord); err != nil {
		h.logger.Error("speaking_word_repetition_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingWordRepetitionID,
		}, "Failed to update word repetition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update word repetition")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *SpeakingWordRepetitionHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("speaking_word_repetition_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("speaking_word_repetition_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid word repetition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid word repetition ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("speaking_word_repetition_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete word repetition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete word repetition")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Word repetition deleted successfully"})
}
