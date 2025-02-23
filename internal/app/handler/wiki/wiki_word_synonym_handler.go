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

type WikiWordSynonymHandler struct {
	service *wikiSer.WikiWordSynonymService
	logger  *logger.PrettyLogger
}

func NewWikiWordSynonymHandler(
	service *wikiSer.WikiWordSynonymService,
	logger *logger.PrettyLogger,
) *WikiWordSynonymHandler {
	return &WikiWordSynonymHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WikiWordSynonymHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req dto.CreateWikiWordSynonymRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("wiki_word_synonym_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	synonym := &wikiModel.WikiWordSynonym{
		ID:                   uuid.New(),
		WikiWordDefinitionID: req.WikiWordDefinitionID,
		WikiSynonymID:        req.WikiSynonymID,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := h.service.Create(ctx, synonym); err != nil {
		h.logger.Error("wiki_word_synonym_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create synonym")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create synonym")
		return
	}

	response.WriteJSON(w, http.StatusCreated, dto.WikiWordSynonymResponse{
		ID:                   synonym.ID,
		WikiWordDefinitionID: synonym.WikiWordDefinitionID,
		WikiSynonymID:        synonym.WikiSynonymID,
		CreatedAt:            synonym.CreatedAt,
		UpdatedAt:            synonym.UpdatedAt,
	})
}

func (h *WikiWordSynonymHandler) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_synonym_handler.get.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_synonym_handler.get.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid synonym ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid synonym ID")
		return
	}

	synonym, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("wiki_word_synonym_handler.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get synonym")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get synonym")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.WikiWordSynonymResponse{
		ID:                   synonym.ID,
		WikiWordDefinitionID: synonym.WikiWordDefinitionID,
		WikiSynonymID:        synonym.WikiSynonymID,
		CreatedAt:            synonym.CreatedAt,
		UpdatedAt:            synonym.UpdatedAt,
	})
}

func (h *WikiWordSynonymHandler) GetByDefinitionID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_synonym_handler.get_by_definition.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	definitionIDStr := ginCtx.Param("definition_id")
	definitionID, err := uuid.Parse(definitionIDStr)
	if err != nil {
		h.logger.Error("wiki_word_synonym_handler.get_by_definition.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    definitionIDStr,
		}, "Invalid definition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid definition ID")
		return
	}

	synonyms, err := h.service.GetByDefinitionID(ctx, definitionID)
	if err != nil {
		h.logger.Error("wiki_word_synonym_handler.get_by_definition", map[string]interface{}{
			"error":         err.Error(),
			"definition_id": definitionID,
		}, "Failed to get synonyms")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get synonyms")
		return
	}

	// Convert models to DTOs
	responses := make([]dto.WikiWordSynonymResponse, len(synonyms))
	for i, synonym := range synonyms {
		responses[i] = dto.WikiWordSynonymResponse{
			ID:                   synonym.ID,
			WikiWordDefinitionID: synonym.WikiWordDefinitionID,
			WikiSynonymID:        synonym.WikiSynonymID,
			CreatedAt:            synonym.CreatedAt,
			UpdatedAt:            synonym.UpdatedAt,
		}
	}

	response.WriteJSON(w, http.StatusOK, responses)
}

func (h *WikiWordSynonymHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_synonym_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_synonym_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid synonym ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid synonym ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("wiki_word_synonym_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete synonym")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete synonym")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Synonym deleted successfully"})
}
