package grammar

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	grammarDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/grammar"
	grammarService "fluencybe/internal/app/service/grammar"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GrammarSentenceTransformationHandler struct {
	service *grammarService.GrammarSentenceTransformationService
	logger  *logger.PrettyLogger
}

func NewGrammarSentenceTransformationHandler(
	service *grammarService.GrammarSentenceTransformationService,
	logger *logger.PrettyLogger,
) *GrammarSentenceTransformationHandler {
	return &GrammarSentenceTransformationHandler{
		service: service,
		logger:  logger,
	}
}

func (h *GrammarSentenceTransformationHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req grammarDTO.CreateGrammarSentenceTransformationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("grammar_sentence_transformation_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	transformation := &grammar.GrammarSentenceTransformation{
		ID:                     uuid.New(),
		GrammarQuestionID:      req.GrammarQuestionID,
		OriginalSentence:       req.OriginalSentence,
		BeginningWord:          req.BeginningWord,
		ExampleCorrectSentence: req.ExampleCorrectSentence,
		Explain:                req.Explain,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := h.service.Create(ctx, transformation); err != nil {
		h.logger.Error("grammar_sentence_transformation_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create sentence transformation")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create sentence transformation")
		return
	}

	responseData := grammarDTO.GrammarSentenceTransformationResponse{
		ID:                     transformation.ID,
		OriginalSentence:       transformation.OriginalSentence,
		BeginningWord:          transformation.BeginningWord,
		ExampleCorrectSentence: transformation.ExampleCorrectSentence,
		Explain:                transformation.Explain,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *GrammarSentenceTransformationHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req grammarDTO.UpdateGrammarSentenceTransformationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("grammar_sentence_transformation_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	transformation, err := h.service.GetByID(ctx, req.GrammarSentenceTransformationID)
	if err != nil {
		h.logger.Error("grammar_sentence_transformation_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.GrammarSentenceTransformationID,
		}, "Failed to get sentence transformation")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Sentence transformation not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to get sentence transformation")
		return
	}

	switch req.Field {
	case "original_sentence":
		transformation.OriginalSentence = req.Value
	case "beginning_word":
		transformation.BeginningWord = req.Value
	case "example_correct_sentence":
		transformation.ExampleCorrectSentence = req.Value
	case "explain":
		transformation.Explain = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	transformation.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, transformation); err != nil {
		h.logger.Error("grammar_sentence_transformation_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.GrammarSentenceTransformationID,
		}, "Failed to update sentence transformation")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update sentence transformation")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *GrammarSentenceTransformationHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("grammar_sentence_transformation_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("grammar_sentence_transformation_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid sentence transformation ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid sentence transformation ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("grammar_sentence_transformation_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete sentence transformation")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Sentence transformation not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete sentence transformation")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Sentence transformation deleted successfully"})
}
