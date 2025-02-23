package listening

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	listeningDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/listening"
	listeningService "fluencybe/internal/app/service/listening"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"strconv"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ListeningChoiceMultiOptionHandler struct {
	service *listeningService.ListeningChoiceMultiOptionService
	logger  *logger.PrettyLogger
}

func NewListeningChoiceMultiOptionHandler(
	service *listeningService.ListeningChoiceMultiOptionService,
	logger *logger.PrettyLogger,
) *ListeningChoiceMultiOptionHandler {
	return &ListeningChoiceMultiOptionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ListeningChoiceMultiOptionHandler) CreateOption(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req listeningDTO.CreateListeningChoiceMultiOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_choice_multi_option_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	option := &listening.ListeningChoiceMultiOption{
		ID:                            uuid.New(),
		ListeningChoiceMultiQuestionID: req.ListeningChoiceMultiQuestionID,
		Options:                       req.Options,
		IsCorrect:                     req.IsCorrect,
		CreatedAt:                     time.Now(),
		UpdatedAt:                     time.Now(),
	}

	if err := h.service.CreateOption(ctx, option); err != nil {
		h.logger.Error("listening_choice_multi_option_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create option")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create option")
		return
	}

	responseData := listeningDTO.ListeningChoiceMultiOptionResponse{
		ID:        option.ID,
		Options:   option.Options,
		IsCorrect: option.IsCorrect,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *ListeningChoiceMultiOptionHandler) UpdateOption(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req listeningDTO.UpdateListeningChoiceMultiOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_choice_multi_option_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	option, err := h.service.GetOption(ctx, req.ListeningChoiceMultiOptionID)
	if err != nil {
		h.logger.Error("listening_choice_multi_option_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ListeningChoiceMultiOptionID,
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
		h.logger.Error("listening_choice_multi_option_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ListeningChoiceMultiOptionID,
		}, "Failed to update option")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update option")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *ListeningChoiceMultiOptionHandler) DeleteOption(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("listening_choice_multi_option_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("listening_choice_multi_option_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid option ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid option ID")
		return
	}

	if err := h.service.DeleteOption(ctx, id); err != nil {
		h.logger.Error("listening_choice_multi_option_handler.delete", map[string]interface{}{
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
