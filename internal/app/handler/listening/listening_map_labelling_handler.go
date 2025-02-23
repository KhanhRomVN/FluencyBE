package listening

import (
	"context"
	"encoding/json"
	listeningDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/listening"
	listeningService "fluencybe/internal/app/service/listening"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ListeningMapLabellingHandler struct {
	service *listeningService.ListeningMapLabellingService
	logger  *logger.PrettyLogger
}

func NewListeningMapLabellingHandler(
	service *listeningService.ListeningMapLabellingService,
	logger *logger.PrettyLogger,
) *ListeningMapLabellingHandler {
	return &ListeningMapLabellingHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ListeningMapLabellingHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req listeningDTO.CreateListeningMapLabellingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_map_labelling_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	mapLabelling := &listening.ListeningMapLabelling{
		ID:                  uuid.New(),
		ListeningQuestionID: req.ListeningQuestionID,
		Question:            req.Question,
		Answer:              req.Answer,
		Explain:             req.Explain,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := h.service.Create(ctx, mapLabelling); err != nil {
		h.logger.Error("listening_map_labelling_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create map labelling")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create map labelling")
		return
	}

	responseData := listeningDTO.ListeningMapLabellingResponse{
		ID:       mapLabelling.ID,
		Question: mapLabelling.Question,
		Answer:   mapLabelling.Answer,
		Explain:  mapLabelling.Explain,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *ListeningMapLabellingHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req listeningDTO.UpdateListeningMapLabellingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_map_labelling_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingLabelling, err := h.service.GetByID(ctx, req.ListeningMapLabellingID)
	if err != nil {
		h.logger.Error("listening_map_labelling_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ListeningMapLabellingID,
		}, "Failed to get existing labelling")
		response.WriteError(w, http.StatusNotFound, "Map labelling not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "question":
		existingLabelling.Question = req.Value
	case "answer":
		existingLabelling.Answer = req.Value
	case "explain":
		existingLabelling.Explain = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingLabelling.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingLabelling); err != nil {
		h.logger.Error("listening_map_labelling_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ListeningMapLabellingID,
		}, "Failed to update map labelling")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update map labelling")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *ListeningMapLabellingHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("listening_map_labelling_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("listening_map_labelling_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid map labelling ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid map labelling ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("listening_map_labelling_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete map labelling")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete map labelling")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Map labelling deleted successfully"})
}
