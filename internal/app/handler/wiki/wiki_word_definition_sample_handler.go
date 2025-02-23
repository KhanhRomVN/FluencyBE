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

type WikiWordDefinitionSampleHandler struct {
	service *wikiSer.WikiWordDefinitionSampleService
	logger  *logger.PrettyLogger
}

func NewWikiWordDefinitionSampleHandler(
	service *wikiSer.WikiWordDefinitionSampleService,
	logger *logger.PrettyLogger,
) *WikiWordDefinitionSampleHandler {
	return &WikiWordDefinitionSampleHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WikiWordDefinitionSampleHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req dto.CreateWikiWordDefinitionSampleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("wiki_word_definition_sample_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sample := &wikiModel.WikiWordDefinitionSample{
		ID:                   uuid.New(),
		WikiWordDefinitionID: req.WikiWordDefinitionID,
		SampleSentence:       req.SampleSentence,
		SampleSentenceMean:   req.SampleSentenceMean,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := h.service.Create(ctx, sample); err != nil {
		h.logger.Error("wiki_word_definition_sample_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create sample")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create sample")
		return
	}

	response.WriteJSON(w, http.StatusCreated, dto.WikiWordDefinitionSampleResponse{
		ID:                   sample.ID,
		WikiWordDefinitionID: sample.WikiWordDefinitionID,
		SampleSentence:       sample.SampleSentence,
		SampleSentenceMean:   sample.SampleSentenceMean,
		CreatedAt:            sample.CreatedAt,
		UpdatedAt:            sample.UpdatedAt,
	})
}

func (h *WikiWordDefinitionSampleHandler) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_definition_sample_handler.get.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_definition_sample_handler.get.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid sample ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid sample ID")
		return
	}

	sample, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("wiki_word_definition_sample_handler.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get sample")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get sample")
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.WikiWordDefinitionSampleResponse{
		ID:                   sample.ID,
		WikiWordDefinitionID: sample.WikiWordDefinitionID,
		SampleSentence:       sample.SampleSentence,
		SampleSentenceMean:   sample.SampleSentenceMean,
		CreatedAt:            sample.CreatedAt,
		UpdatedAt:            sample.UpdatedAt,
	})
}

func (h *WikiWordDefinitionSampleHandler) GetByDefinitionID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_definition_sample_handler.get_by_definition.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	definitionIDStr := ginCtx.Param("definition_id")
	definitionID, err := uuid.Parse(definitionIDStr)
	if err != nil {
		h.logger.Error("wiki_word_definition_sample_handler.get_by_definition.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    definitionIDStr,
		}, "Invalid definition ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid definition ID")
		return
	}

	samples, err := h.service.GetByDefinitionID(ctx, definitionID)
	if err != nil {
		h.logger.Error("wiki_word_definition_sample_handler.get_by_definition", map[string]interface{}{
			"error":         err.Error(),
			"definition_id": definitionID,
		}, "Failed to get samples")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get samples")
		return
	}

	// Convert models to DTOs
	responses := make([]dto.WikiWordDefinitionSampleResponse, len(samples))
	for i, sample := range samples {
		responses[i] = dto.WikiWordDefinitionSampleResponse{
			ID:                   sample.ID,
			WikiWordDefinitionID: sample.WikiWordDefinitionID,
			SampleSentence:       sample.SampleSentence,
			SampleSentenceMean:   sample.SampleSentenceMean,
			CreatedAt:            sample.CreatedAt,
			UpdatedAt:            sample.UpdatedAt,
		}
	}

	response.WriteJSON(w, http.StatusOK, responses)
}

func (h *WikiWordDefinitionSampleHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_definition_sample_handler.update.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_definition_sample_handler.update.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid sample ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid sample ID")
		return
	}

	var req dto.UpdateWikiWordDefinitionSampleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("wiki_word_definition_sample_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.Update(ctx, id, req); err != nil {
		h.logger.Error("wiki_word_definition_sample_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update sample")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update sample")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *WikiWordDefinitionSampleHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("wiki_word_definition_sample_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("wiki_word_definition_sample_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid sample ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid sample ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("wiki_word_definition_sample_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete sample")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete sample")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Sample deleted successfully"})
}
