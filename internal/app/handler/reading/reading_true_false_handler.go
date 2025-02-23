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

type ReadingTrueFalseHandler struct {
	service *readingService.ReadingTrueFalseService
	logger  *logger.PrettyLogger
}

func NewReadingTrueFalseHandler(
	service *readingService.ReadingTrueFalseService,
	logger *logger.PrettyLogger,
) *ReadingTrueFalseHandler {
	return &ReadingTrueFalseHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ReadingTrueFalseHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req readingDTO.CreateReadingTrueFalseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("reading_true_false_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	trueFalse := &reading.ReadingTrueFalse{
		ID:                uuid.New(),
		ReadingQuestionID: req.ReadingQuestionID,
		Question:          req.Question,
		Answer:            req.Answer,
		Explain:           req.Explain,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := h.service.Create(ctx, trueFalse); err != nil {
		h.logger.Error("reading_true_false_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create true/false question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create true/false question")
		return
	}

	responseData := readingDTO.ReadingTrueFalseResponse{
		ID:       trueFalse.ID,
		Question: trueFalse.Question,
		Answer:   trueFalse.Answer,
		Explain:  trueFalse.Explain,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *ReadingTrueFalseHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req readingDTO.UpdateReadingTrueFalseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("reading_true_false_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	trueFalse, err := h.service.GetByID(ctx, req.ReadingTrueFalseID)
	if err != nil {
		h.logger.Error("reading_true_false_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ReadingTrueFalseID,
		}, "Failed to get true/false question")
		response.WriteError(w, http.StatusNotFound, "True/false question not found")
		return
	}

	switch req.Field {
	case "question":
		trueFalse.Question = req.Value
	case "answer":
		if req.Value != "TRUE" && req.Value != "FALSE" && req.Value != "NOT GIVEN" {
			response.WriteError(w, http.StatusBadRequest, "Invalid answer value")
			return
		}
		trueFalse.Answer = req.Value
	case "explain":
		trueFalse.Explain = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	trueFalse.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, trueFalse); err != nil {
		h.logger.Error("reading_true_false_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ReadingTrueFalseID,
		}, "Failed to update true/false question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update true/false question")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *ReadingTrueFalseHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("reading_true_false_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("reading_true_false_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid true/false question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid true/false question ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("reading_true_false_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete true/false question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete true/false question")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "True/false question deleted successfully"})
}
