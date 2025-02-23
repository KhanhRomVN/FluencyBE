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

type GrammarErrorIdentificationHandler struct {
	service *grammarService.GrammarErrorIdentificationService
	logger  *logger.PrettyLogger
}

func NewGrammarErrorIdentificationHandler(
	service *grammarService.GrammarErrorIdentificationService,
	logger *logger.PrettyLogger,
) *GrammarErrorIdentificationHandler {
	return &GrammarErrorIdentificationHandler{
		service: service,
		logger:  logger,
	}
}

func (h *GrammarErrorIdentificationHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req grammarDTO.CreateGrammarErrorIdentificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("grammar_error_identification_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	identification := &grammar.GrammarErrorIdentification{
		ID:                uuid.New(),
		GrammarQuestionID: req.GrammarQuestionID,
		ErrorSentence:     req.ErrorSentence,
		ErrorWord:         req.ErrorWord,
		CorrectWord:       req.CorrectWord,
		Explain:           req.Explain,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := h.service.Create(ctx, identification); err != nil {
		h.logger.Error("grammar_error_identification_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create error identification")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create error identification")
		return
	}

	responseData := grammarDTO.GrammarErrorIdentificationResponse{
		ID:            identification.ID,
		ErrorSentence: identification.ErrorSentence,
		ErrorWord:     identification.ErrorWord,
		CorrectWord:   identification.CorrectWord,
		Explain:       identification.Explain,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *GrammarErrorIdentificationHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req grammarDTO.UpdateGrammarErrorIdentificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("grammar_error_identification_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	identification, err := h.service.GetByID(ctx, req.GrammarErrorIdentificationID)
	if err != nil {
		h.logger.Error("grammar_error_identification_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.GrammarErrorIdentificationID,
		}, "Failed to get error identification")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Error identification not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to get error identification")
		return
	}

	switch req.Field {
	case "error_sentence":
		identification.ErrorSentence = req.Value
	case "error_word":
		identification.ErrorWord = req.Value
	case "correct_word":
		identification.CorrectWord = req.Value
	case "explain":
		identification.Explain = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	identification.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, identification); err != nil {
		h.logger.Error("grammar_error_identification_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.GrammarErrorIdentificationID,
		}, "Failed to update error identification")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update error identification")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *GrammarErrorIdentificationHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("grammar_error_identification_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("grammar_error_identification_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid error identification ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid error identification ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("grammar_error_identification_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete error identification")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Error identification not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete error identification")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Error identification deleted successfully"})
}
