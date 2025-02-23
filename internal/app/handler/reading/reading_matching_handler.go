package reading

import (
	"context"
	"encoding/json"
	readingDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/reading"
	readingService "fluencybe/internal/app/service/reading"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReadingMatchingHandler struct {
	service *readingService.ReadingMatchingService
	logger  *logger.PrettyLogger
}

func NewReadingMatchingHandler(
	service *readingService.ReadingMatchingService,
	logger *logger.PrettyLogger,
) *ReadingMatchingHandler {
	return &ReadingMatchingHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ReadingMatchingHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req readingDTO.CreateReadingMatchingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("reading_matching_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	matching := &reading.ReadingMatching{
		ID:                uuid.New(),
		ReadingQuestionID: req.ReadingQuestionID,
		Question:          req.Question,
		Answer:            req.Answer,
		Explain:           req.Explain,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := h.service.Create(ctx, matching); err != nil {
		h.logger.Error("reading_matching_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create matching")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create matching")
		return
	}

	responseData := readingDTO.ReadingMatchingResponse{
		ID:       matching.ID,
		Question: matching.Question,
		Answer:   matching.Answer,
		Explain:  matching.Explain,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *ReadingMatchingHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req readingDTO.UpdateReadingMatchingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("reading_matching_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingMatching, err := h.service.GetByID(ctx, req.ReadingMatchingID)
	if err != nil {
		h.logger.Error("reading_matching_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ReadingMatchingID,
		}, "Failed to get existing matching")
		response.WriteError(w, http.StatusNotFound, "Matching not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "question":
		existingMatching.Question = req.Value
	case "answer":
		existingMatching.Answer = req.Value
	case "explain":
		existingMatching.Explain = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingMatching.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingMatching); err != nil {
		h.logger.Error("reading_matching_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ReadingMatchingID,
		}, "Failed to update matching")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update matching")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *ReadingMatchingHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("reading_matching_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("reading_matching_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid matching ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid matching ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("reading_matching_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete matching")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete matching")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Matching deleted successfully"})
}
