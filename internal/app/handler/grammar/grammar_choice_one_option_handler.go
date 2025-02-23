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
	"strconv"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GrammarChoiceOneOptionHandler struct {
	service *grammarService.GrammarChoiceOneOptionService
	logger  *logger.PrettyLogger
}

func NewGrammarChoiceOneOptionHandler(
	service *grammarService.GrammarChoiceOneOptionService,
	logger *logger.PrettyLogger,
) *GrammarChoiceOneOptionHandler {
	return &GrammarChoiceOneOptionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *GrammarChoiceOneOptionHandler) CreateOption(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req grammarDTO.CreateGrammarChoiceOneOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("grammar_choice_one_option_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	option := &grammar.GrammarChoiceOneOption{
		ID:                         uuid.New(),
		GrammarChoiceOneQuestionID: req.GrammarChoiceOneQuestionID,
		Options:                    req.Options,
		IsCorrect:                  req.IsCorrect,
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
	}

	if err := h.service.CreateOption(ctx, option); err != nil {
		h.logger.Error("grammar_choice_one_option_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create option")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create option")
		return
	}

	responseData := grammarDTO.GrammarChoiceOneOptionResponse{
		ID:        option.ID,
		Options:   option.Options,
		IsCorrect: option.IsCorrect,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *GrammarChoiceOneOptionHandler) UpdateOption(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req grammarDTO.UpdateGrammarChoiceOneOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("grammar_choice_one_option_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	option, err := h.service.GetOption(ctx, req.GrammarChoiceOneOptionID)
	if err != nil {
		h.logger.Error("grammar_choice_one_option_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.GrammarChoiceOneOptionID,
		}, "Failed to get option")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Option not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to get option")
		return
	}

	switch req.Field {
	case "options":
		option.Options = req.Value
	case "is_correct":
		isCorrect, err := strconv.ParseBool(req.Value)
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "Invalid boolean value for is_correct")
			return
		}
		option.IsCorrect = isCorrect
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	option.UpdatedAt = time.Now()

	if err := h.service.UpdateOption(ctx, option); err != nil {
		h.logger.Error("grammar_choice_one_option_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.GrammarChoiceOneOptionID,
		}, "Failed to update option")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update option")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *GrammarChoiceOneOptionHandler) DeleteOption(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("grammar_choice_one_option_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("grammar_choice_one_option_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid option ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid option ID")
		return
	}

	if err := h.service.DeleteOption(ctx, id); err != nil {
		h.logger.Error("grammar_choice_one_option_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete option")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Option not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete option")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Option deleted successfully"})
}
