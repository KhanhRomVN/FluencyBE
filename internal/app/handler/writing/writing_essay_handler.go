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

type WritingEssayHandler struct {
	service *writingService.WritingEssayService
	logger  *logger.PrettyLogger
}

func NewWritingEssayHandler(
	service *writingService.WritingEssayService,
	logger *logger.PrettyLogger,
) *WritingEssayHandler {
	return &WritingEssayHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WritingEssayHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req writingDTO.CreateWritingEssayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("writing_essay_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	essay := &writing.WritingEssay{
		ID:                uuid.New(),
		WritingQuestionID: req.WritingQuestionID,
		EssayType:         req.EssayType,
		RequiredPoints:    req.RequiredPoints,
		MinWords:          req.MinWords,
		MaxWords:          req.MaxWords,
		SampleEssay:       req.SampleEssay,
		Explain:           req.Explain,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := h.service.Create(ctx, essay); err != nil {
		h.logger.Error("writing_essay_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create essay")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create essay")
		return
	}

	responseData := writingDTO.WritingEssayResponse{
		ID:             essay.ID,
		EssayType:      essay.EssayType,
		RequiredPoints: essay.RequiredPoints,
		MinWords:       essay.MinWords,
		MaxWords:       essay.MaxWords,
		SampleEssay:    essay.SampleEssay,
		Explain:        essay.Explain,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *WritingEssayHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req writingDTO.UpdateWritingEssayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("writing_essay_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingEssay, err := h.service.GetByID(ctx, req.WritingEssayID)
	if err != nil {
		h.logger.Error("writing_essay_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.WritingEssayID,
		}, "Failed to get existing essay")
		response.WriteError(w, http.StatusNotFound, "Essay not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "essay_type":
		existingEssay.EssayType = req.Value.(string)
	case "required_points":
		existingEssay.RequiredPoints = req.Value.([]string)
	case "min_words":
		existingEssay.MinWords = int(req.Value.(float64))
	case "max_words":
		existingEssay.MaxWords = int(req.Value.(float64))
	case "sample_essay":
		existingEssay.SampleEssay = req.Value.(string)
	case "explain":
		existingEssay.Explain = req.Value.(string)
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingEssay.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingEssay); err != nil {
		h.logger.Error("writing_essay_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.WritingEssayID,
		}, "Failed to update essay")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update essay")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *WritingEssayHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("writing_essay_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("writing_essay_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid essay ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid essay ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("writing_essay_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete essay")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete essay")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Essay deleted successfully"})
}
