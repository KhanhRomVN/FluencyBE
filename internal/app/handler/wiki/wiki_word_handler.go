package wiki

import (
	"context"
	"encoding/json"
	"fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/wiki"
	wikiSer "fluencybe/internal/app/service/wiki"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WikiWordHandler struct {
	service                     *wikiSer.WikiWordService
	wordDefinitionService       *wikiSer.WikiWordDefinitionService
	wordDefinitionSampleService *wikiSer.WikiWordDefinitionSampleService
	wordSynonymService          *wikiSer.WikiWordSynonymService
	wordAntonymService          *wikiSer.WikiWordAntonymService
	logger                      *logger.PrettyLogger
}

func NewWikiWordHandler(
	service *wikiSer.WikiWordService,
	wordDefinitionService *wikiSer.WikiWordDefinitionService,
	wordDefinitionSampleService *wikiSer.WikiWordDefinitionSampleService,
	wordSynonymService *wikiSer.WikiWordSynonymService,
	wordAntonymService *wikiSer.WikiWordAntonymService,
	logger *logger.PrettyLogger,
) *WikiWordHandler {
	return &WikiWordHandler{
		service:                     service,
		wordDefinitionService:       wordDefinitionService,
		wordDefinitionSampleService: wordDefinitionSampleService,
		wordSynonymService:          wordSynonymService,
		wordAntonymService:          wordAntonymService,
		logger:                      logger,
	}
}

func (h *WikiWordHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req dto.CreateWikiWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("wiki_word_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	word := &wiki.WikiWord{
		ID:            uuid.New(),
		Word:          req.Word,
		Pronunciation: req.Pronunciation,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.service.Create(ctx, word); err != nil {
		h.logger.Error("wiki_word_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create word")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create word")
		return
	}

	response.WriteJSON(w, http.StatusCreated, dto.WikiWordResponse{
		ID:            word.ID,
		Word:          word.Word,
		Pronunciation: word.Pronunciation,
		CreatedAt:     word.CreatedAt,
		UpdatedAt:     word.UpdatedAt,
	})
}

func (h *WikiWordHandler) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_handler.get.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_handler.get.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid word ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid word ID")
		return
	}

	word, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("wiki_word_handler.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get word")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get word")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.WikiWordResponse{
		ID:            word.ID,
		Word:          word.Word,
		Pronunciation: word.Pronunciation,
		CreatedAt:     word.CreatedAt,
		UpdatedAt:     word.UpdatedAt,
	})
}

func (h *WikiWordHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_handler.update.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_handler.update.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid word ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid word ID")
		return
	}

	var req dto.UpdateWikiWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("wiki_word_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.Update(ctx, id, req); err != nil {
		h.logger.Error("wiki_word_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update word")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update word")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *WikiWordHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid word ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid word ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("wiki_word_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete word")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete word")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Word deleted successfully"})
}

func (h *WikiWordHandler) Search(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_handler.search.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var params dto.WikiSearchParams
	if err := ginCtx.ShouldBindQuery(&params); err != nil {
		h.logger.Error("wiki_word_handler.search.bind", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to bind query parameters")
		response.WriteError(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	// Set default values if not provided
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 10
	}

	words, total, err := h.service.Search(ctx, params)
	if err != nil {
		h.logger.Error("wiki_word_handler.search", map[string]interface{}{
			"error":  err.Error(),
			"params": params,
		}, "Failed to search words")
		response.WriteError(w, http.StatusInternalServerError, "Failed to search words")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.WikiPaginationResponse{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Data:     words,
	})
}
