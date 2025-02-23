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
	"strconv"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WikiPhraseHandler struct {
	service *wikiSer.WikiPhraseService
	logger  *logger.PrettyLogger
}

func NewWikiPhraseHandler(
	service *wikiSer.WikiPhraseService,
	logger *logger.PrettyLogger,
) *WikiPhraseHandler {
	return &WikiPhraseHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WikiPhraseHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req dto.CreateWikiPhraseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("wiki_phrase_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	phrase := &wikiModel.WikiPhrase{
		ID:              uuid.New(),
		Phrase:          req.Phrase,
		Type:            req.Type,
		DifficultyLevel: req.DifficultyLevel,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.service.Create(ctx, phrase); err != nil {
		h.logger.Error("wiki_phrase_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create phrase")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create phrase")
		return
	}

	response.WriteJSON(w, http.StatusCreated, dto.WikiPhraseResponse{
		ID:              phrase.ID,
		Phrase:          phrase.Phrase,
		Type:            phrase.Type,
		DifficultyLevel: phrase.DifficultyLevel,
		CreatedAt:       phrase.CreatedAt,
		UpdatedAt:       phrase.UpdatedAt,
	})
}

func (h *WikiPhraseHandler) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_phrase_handler.get.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_phrase_handler.get.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid phrase ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid phrase ID")
		return
	}

	phrase, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("wiki_phrase_handler.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get phrase")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get phrase")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.WikiPhraseResponse{
		ID:              phrase.ID,
		Phrase:          phrase.Phrase,
		Type:            phrase.Type,
		DifficultyLevel: phrase.DifficultyLevel,
		CreatedAt:       phrase.CreatedAt,
		UpdatedAt:       phrase.UpdatedAt,
	})
}

func (h *WikiPhraseHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_phrase_handler.update.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_phrase_handler.update.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid phrase ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid phrase ID")
		return
	}

	var req dto.UpdateWikiPhraseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("wiki_phrase_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.Update(ctx, id, req); err != nil {
		h.logger.Error("wiki_phrase_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update phrase")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update phrase")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *WikiPhraseHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_phrase_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_phrase_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid phrase ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid phrase ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("wiki_phrase_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete phrase")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete phrase")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Phrase deleted successfully"})
}

func (h *WikiPhraseHandler) Search(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_phrase_handler.search.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var params dto.WikiSearchParams
	if err := ginCtx.ShouldBindQuery(&params); err != nil {
		h.logger.Error("wiki_phrase_handler.search.bind", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to bind query parameters")
		response.WriteError(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	// Get optional filters
	phraseType := ginCtx.Query("type")
	difficultyLevelStr := ginCtx.Query("difficulty_level")

	var difficultyLevel *int
	if difficultyLevelStr != "" {
		level, err := strconv.Atoi(difficultyLevelStr)
		if err != nil {
			h.logger.Error("wiki_phrase_handler.search.parse_difficulty", map[string]interface{}{
				"error": err.Error(),
				"value": difficultyLevelStr,
			}, "Invalid difficulty level format")
			response.WriteError(w, http.StatusBadRequest, "Invalid difficulty level")
			return
		}
		difficultyLevel = &level
	}

	phrases, total, err := h.service.SearchWithFilter(ctx, params, phraseType, difficultyLevel)
	if err != nil {
		h.logger.Error("wiki_phrase_handler.search", map[string]interface{}{
			"error":  err.Error(),
			"params": params,
		}, "Failed to search phrases")
		response.WriteError(w, http.StatusInternalServerError, "Failed to search phrases")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.WikiPaginationResponse{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Data:     phrases,
	})
}
