package reading

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	readingDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/reading"
	readingService "fluencybe/internal/app/service/reading"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"strconv"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReadingChoiceMultiOptionHandler struct {
	service *readingService.ReadingChoiceMultiOptionService
	logger  *logger.PrettyLogger
}

func NewReadingChoiceMultiOptionHandler(
	service *readingService.ReadingChoiceMultiOptionService,
	logger *logger.PrettyLogger,
) *ReadingChoiceMultiOptionHandler {
	return &ReadingChoiceMultiOptionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ReadingChoiceMultiOptionHandler) CreateOption(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req readingDTO.CreateReadingChoiceMultiOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("reading_choice_multi_option_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	option := &reading.ReadingChoiceMultiOption{
		ID:                           uuid.New(),
		ReadingChoiceMultiQuestionID: req.ReadingChoiceMultiQuestionID,
		Options:                      req.Options,
		IsCorrect:                    req.IsCorrect,
		CreatedAt:                    time.Now(),
		UpdatedAt:                    time.Now(),
	}

	if err := h.service.CreateOption(ctx, option); err != nil {
		h.logger.Error("reading_choice_multi_option_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create option")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create option")
		return
	}

	responseData := readingDTO.ReadingChoiceMultiOptionResponse{
		ID:        option.ID,
		Options:   option.Options,
		IsCorrect: option.IsCorrect,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *ReadingChoiceMultiOptionHandler) UpdateOption(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req readingDTO.UpdateReadingChoiceMultiOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("reading_choice_multi_option_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	option, err := h.service.GetOption(ctx, req.ReadingChoiceMultiOptionID)
	if err != nil {
		h.logger.Error("reading_choice_multi_option_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ReadingChoiceMultiOptionID,
		}, "Failed to get option")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Option not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to get option")
		return
	}

	switch req.Field {
	case "options":
		option.Options = req.Value
	case "is_correct":
		isCorrect, err := strconv.ParseBool(req.Value)
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "Invalid boolean value for is_correct")
			return
		}
		option.IsCorrect = isCorrect
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	option.UpdatedAt = time.Now()

	if err := h.service.UpdateOption(ctx, option); err != nil {
		h.logger.Error("reading_choice_multi_option_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ReadingChoiceMultiOptionID,
		}, "Failed to update option")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update option")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *ReadingChoiceMultiOptionHandler) DeleteOption(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("reading_choice_multi_option_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("reading_choice_multi_option_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid option ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid option ID")
		return
	}

	if err := h.service.DeleteOption(ctx, id); err != nil {
		h.logger.Error("reading_choice_multi_option_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete option")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Option not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete option")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Option deleted successfully"})
}
