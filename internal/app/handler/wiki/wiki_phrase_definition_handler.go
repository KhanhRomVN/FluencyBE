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

type WikiPhraseDefinitionHandler struct {
	service *wikiSer.WikiPhraseDefinitionService
	logger  *logger.PrettyLogger
}

func NewWikiPhraseDefinitionHandler(
	service *wikiSer.WikiPhraseDefinitionService,
	logger *logger.PrettyLogger,
) *WikiPhraseDefinitionHandler {
	return &WikiPhraseDefinitionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WikiPhraseDefinitionHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req dto.CreateWikiPhraseDefinitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("wiki_phrase_definition_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	definition := &wikiModel.WikiPhraseDefinition{
		ID:               uuid.New(),
		WikiPhraseID:     req.WikiPhraseID,
		Mean:             req.Mean,
		IsMainDefinition: req.IsMainDefinition,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := h.service.Create(ctx, definition); err != nil {
		h.logger.Error("wiki_phrase_definition_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create definition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create definition")
		return
	}

	response.WriteJSON(w, http.StatusCreated, dto.WikiPhraseDefinitionResponse{
		ID:               definition.ID,
		WikiPhraseID:     definition.WikiPhraseID,
		Mean:             definition.Mean,
		IsMainDefinition: definition.IsMainDefinition,
		CreatedAt:        definition.CreatedAt,
		UpdatedAt:        definition.UpdatedAt,
	})
}

func (h *WikiPhraseDefinitionHandler) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_phrase_definition_handler.get.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_phrase_definition_handler.get.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid definition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid definition ID")
		return
	}

	definition, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("wiki_phrase_definition_handler.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get definition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get definition")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.WikiPhraseDefinitionResponse{
		ID:               definition.ID,
		WikiPhraseID:     definition.WikiPhraseID,
		Mean:             definition.Mean,
		IsMainDefinition: definition.IsMainDefinition,
		CreatedAt:        definition.CreatedAt,
		UpdatedAt:        definition.UpdatedAt,
	})
}

func (h *WikiPhraseDefinitionHandler) GetByPhraseID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_phrase_definition_handler.get_by_phrase.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	phraseIDStr := ginCtx.Param("phrase_id")
	phraseID, err := uuid.Parse(phraseIDStr)
	if err != nil {
		h.logger.Error("wiki_phrase_definition_handler.get_by_phrase.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    phraseIDStr,
		}, "Invalid phrase ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid phrase ID")
		return
	}

	definitions, err := h.service.GetByPhraseID(ctx, phraseID)
	if err != nil {
		h.logger.Error("wiki_phrase_definition_handler.get_by_phrase", map[string]interface{}{
			"error":     err.Error(),
			"phrase_id": phraseID,
		}, "Failed to get definitions")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get definitions")
		return
	}

	// Convert models to DTOs
	responses := make([]dto.WikiPhraseDefinitionResponse, len(definitions))
	for i, def := range definitions {
		responses[i] = dto.WikiPhraseDefinitionResponse{
			ID:               def.ID,
			WikiPhraseID:     def.WikiPhraseID,
			Mean:             def.Mean,
			IsMainDefinition: def.IsMainDefinition,
			CreatedAt:        def.CreatedAt,
			UpdatedAt:        def.UpdatedAt,
		}
	}

	response.WriteJSON(w, http.StatusOK, responses)
}

func (h *WikiPhraseDefinitionHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_phrase_definition_handler.update.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_phrase_definition_handler.update.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid definition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid definition ID")
		return
	}

	var req dto.UpdateWikiPhraseDefinitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("wiki_phrase_definition_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.Update(ctx, id, req); err != nil {
		h.logger.Error("wiki_phrase_definition_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update definition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update definition")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *WikiPhraseDefinitionHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_phrase_definition_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_phrase_definition_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid definition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid definition ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("wiki_phrase_definition_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete definition")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete definition")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Definition deleted successfully"})
}
