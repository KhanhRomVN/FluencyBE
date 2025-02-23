package writing

import (
	"context"
	"encoding/json"
	writingDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/writing"
	writingService "fluencybe/internal/app/service/writing"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WritingSentenceCompletionHandler struct {
	service *writingService.WritingSentenceCompletionService
	logger  *logger.PrettyLogger
}

func NewWritingSentenceCompletionHandler(
	service *writingService.WritingSentenceCompletionService,
	logger *logger.PrettyLogger,
) *WritingSentenceCompletionHandler {
	return &WritingSentenceCompletionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WritingSentenceCompletionHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req writingDTO.CreateWritingSentenceCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("writing_sentence_completion_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sentence := &writing.WritingSentenceCompletion{
		ID:                uuid.New(),
		WritingQuestionID: req.WritingQuestionID,
		ExampleSentence:   req.ExampleSentence,
		GivenPartSentence: req.GivenPartSentence,
		Position:          req.Position,
		RequiredWords:     req.RequiredWords,
		Explain:           req.Explain,
		MinWords:          req.MinWords,
		MaxWords:          req.MaxWords,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := h.service.Create(ctx, sentence); err != nil {
		h.logger.Error("writing_sentence_completion_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create sentence completion")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create sentence completion")
		return
	}

	responseData := writingDTO.WritingSentenceCompletionResponse{
		ID:                sentence.ID,
		ExampleSentence:   sentence.ExampleSentence,
		GivenPartSentence: sentence.GivenPartSentence,
		Position:          sentence.Position,
		RequiredWords:     sentence.RequiredWords,
		Explain:           sentence.Explain,
		MinWords:          sentence.MinWords,
		MaxWords:          sentence.MaxWords,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *WritingSentenceCompletionHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req writingDTO.UpdateWritingSentenceCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("writing_sentence_completion_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingSentence, err := h.service.GetByID(ctx, req.WritingSentenceCompletionID)
	if err != nil {
		h.logger.Error("writing_sentence_completion_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.WritingSentenceCompletionID,
		}, "Failed to get existing sentence completion")
		response.WriteError(w, http.StatusNotFound, "Sentence completion not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "example_sentence":
		existingSentence.ExampleSentence = req.Value.(string)
	case "given_part_sentence":
		existingSentence.GivenPartSentence = req.Value.(string)
	case "position":
		existingSentence.Position = req.Value.(string)
	case "required_words":
		existingSentence.RequiredWords = req.Value.([]string)
	case "explain":
		existingSentence.Explain = req.Value.(string)
	case "min_words":
		existingSentence.MinWords = int(req.Value.(float64))
	case "max_words":
		existingSentence.MaxWords = int(req.Value.(float64))
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingSentence.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingSentence); err != nil {
		h.logger.Error("writing_sentence_completion_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.WritingSentenceCompletionID,
		}, "Failed to update sentence completion")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update sentence completion")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *WritingSentenceCompletionHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("writing_sentence_completion_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("writing_sentence_completion_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid sentence completion ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid sentence completion ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("writing_sentence_completion_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete sentence completion")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete sentence completion")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Sentence completion deleted successfully"})
}
