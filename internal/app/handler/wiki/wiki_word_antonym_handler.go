package wiki

import (
	"context"
	"encoding/json"
	"fluencybe/internal/app/dto"
	wikiModel "fluencybe/internal/app/model/wiki"
	wikiSer "fluencybe/internal/app/service/wiki"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WikiWordAntonymHandler struct {
	service *wikiSer.WikiWordAntonymService
	logger  *logger.PrettyLogger
}

func NewWikiWordAntonymHandler(
	service *wikiSer.WikiWordAntonymService,
	logger *logger.PrettyLogger,
) *WikiWordAntonymHandler {
	return &WikiWordAntonymHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WikiWordAntonymHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req dto.CreateWikiWordAntonymRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("wiki_word_antonym_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	antonym := &wikiModel.WikiWordAntonym{
		ID:                   uuid.New(),
		WikiWordDefinitionID: req.WikiWordDefinitionID,
		WikiAntonymID:        req.WikiAntonymID,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := h.service.Create(ctx, antonym); err != nil {
		h.logger.Error("wiki_word_antonym_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create antonym")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create antonym")
		return
	}

	response.WriteJSON(w, http.StatusCreated, dto.WikiWordAntonymResponse{
		ID:                   antonym.ID,
		WikiWordDefinitionID: antonym.WikiWordDefinitionID,
		WikiAntonymID:        antonym.WikiAntonymID,
		CreatedAt:            antonym.CreatedAt,
		UpdatedAt:            antonym.UpdatedAt,
	})
}

func (h *WikiWordAntonymHandler) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_antonym_handler.get.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_antonym_handler.get.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid antonym ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid antonym ID")
		return
	}

	antonym, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("wiki_word_antonym_handler.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get antonym")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get antonym")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.WikiWordAntonymResponse{
		ID:                   antonym.ID,
		WikiWordDefinitionID: antonym.WikiWordDefinitionID,
		WikiAntonymID:        antonym.WikiAntonymID,
		CreatedAt:            antonym.CreatedAt,
		UpdatedAt:            antonym.UpdatedAt,
	})
}

func (h *WikiWordAntonymHandler) GetByDefinitionID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_antonym_handler.get_by_definition.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	definitionIDStr := ginCtx.Param("definition_id")
	definitionID, err := uuid.Parse(definitionIDStr)
	if err != nil {
		h.logger.Error("wiki_word_antonym_handler.get_by_definition.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    definitionIDStr,
		}, "Invalid definition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid definition ID")
		return
	}

	antonyms, err := h.service.GetByDefinitionID(ctx, definitionID)
	if err != nil {
		h.logger.Error("wiki_word_antonym_handler.get_by_definition", map[string]interface{}{
			"error":         err.Error(),
			"definition_id": definitionID,
		}, "Failed to get antonyms")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get antonyms")
		return
	}

	// Convert models to DTOs
	responses := make([]dto.WikiWordAntonymResponse, len(antonyms))
	for i, antonym := range antonyms {
		responses[i] = dto.WikiWordAntonymResponse{
			ID:                   antonym.ID,
			WikiWordDefinitionID: antonym.WikiWordDefinitionID,
			WikiAntonymID:        antonym.WikiAntonymID,
			CreatedAt:            antonym.CreatedAt,
			UpdatedAt:            antonym.UpdatedAt,
		}
	}

	response.WriteJSON(w, http.StatusOK, responses)
}

func (h *WikiWordAntonymHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_antonym_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_antonym_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid antonym ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid antonym ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("wiki_word_antonym_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete antonym")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete antonym")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Antonym deleted successfully"})
}
