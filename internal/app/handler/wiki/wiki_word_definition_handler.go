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

type WikiWordDefinitionHandler struct {
	service *wikiSer.WikiWordDefinitionService
	logger  *logger.PrettyLogger
}

func NewWikiWordDefinitionHandler(
	service *wikiSer.WikiWordDefinitionService,
	logger *logger.PrettyLogger,
) *WikiWordDefinitionHandler {
	return &WikiWordDefinitionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WikiWordDefinitionHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req dto.CreateWikiWordDefinitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("wiki_word_definition_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	definition := &wikiModel.WikiWordDefinition{
		ID:               uuid.New(),
		WikiWordID:       req.WikiWordID,
		Means:            req.Means,
		IsMainDefinition: req.IsMainDefinition,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := h.service.Create(ctx, definition); err != nil {
		h.logger.Error("wiki_word_definition_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create definition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create definition")
		return
	}

	response.WriteJSON(w, http.StatusCreated, dto.WikiWordDefinitionResponse{
		ID:               definition.ID,
		WikiWordID:       definition.WikiWordID,
		Means:            definition.Means,
		IsMainDefinition: definition.IsMainDefinition,
		CreatedAt:        definition.CreatedAt,
		UpdatedAt:        definition.UpdatedAt,
	})
}

func (h *WikiWordDefinitionHandler) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_definition_handler.get.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_definition_handler.get.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid definition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid definition ID")
		return
	}

	definition, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("wiki_word_definition_handler.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get definition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get definition")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.WikiWordDefinitionResponse{
		ID:               definition.ID,
		WikiWordID:       definition.WikiWordID,
		Means:            definition.Means,
		IsMainDefinition: definition.IsMainDefinition,
		CreatedAt:        definition.CreatedAt,
		UpdatedAt:        definition.UpdatedAt,
	})
}

func (h *WikiWordDefinitionHandler) GetByWordID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_definition_handler.get_by_word.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	wordIDStr := ginCtx.Param("word_id")
	wordID, err := uuid.Parse(wordIDStr)
	if err != nil {
		h.logger.Error("wiki_word_definition_handler.get_by_word.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    wordIDStr,
		}, "Invalid word ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid word ID")
		return
	}

	definitions, err := h.service.GetByWordID(ctx, wordID)
	if err != nil {
		h.logger.Error("wiki_word_definition_handler.get_by_word", map[string]interface{}{
			"error":   err.Error(),
			"word_id": wordID,
		}, "Failed to get definitions")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get definitions")
		return
	}

	// Convert models to DTOs
	responses := make([]dto.WikiWordDefinitionResponse, len(definitions))
	for i, def := range definitions {
		responses[i] = dto.WikiWordDefinitionResponse{
			ID:               def.ID,
			WikiWordID:       def.WikiWordID,
			Means:            def.Means,
			IsMainDefinition: def.IsMainDefinition,
			CreatedAt:        def.CreatedAt,
			UpdatedAt:        def.UpdatedAt,
		}
	}

	response.WriteJSON(w, http.StatusOK, responses)
}

func (h *WikiWordDefinitionHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_definition_handler.update.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_definition_handler.update.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid definition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid definition ID")
		return
	}

	var req dto.UpdateWikiWordDefinitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("wiki_word_definition_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.Update(ctx, id, req); err != nil {
		h.logger.Error("wiki_word_definition_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update definition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update definition")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *WikiWordDefinitionHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_definition_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_definition_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid definition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid definition ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("wiki_word_definition_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete definition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete definition")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Definition deleted successfully"})
}
