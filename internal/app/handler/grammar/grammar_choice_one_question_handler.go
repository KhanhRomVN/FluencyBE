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

type GrammarChoiceOneQuestionHandler struct {
	questionService *grammarService.GrammarChoiceOneQuestionService
	optionService   *grammarService.GrammarChoiceOneOptionService
	logger          *logger.PrettyLogger
}

func NewGrammarChoiceOneQuestionHandler(
	questionService *grammarService.GrammarChoiceOneQuestionService,
	optionService *grammarService.GrammarChoiceOneOptionService,
	logger *logger.PrettyLogger,
) *GrammarChoiceOneQuestionHandler {
	return &GrammarChoiceOneQuestionHandler{
		questionService: questionService,
		optionService:   optionService,
		logger:          logger,
	}
}

func (h *GrammarChoiceOneQuestionHandler) CreateQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req grammarDTO.CreateGrammarChoiceOneQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("grammar_choice_one_question_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	question := &grammar.GrammarChoiceOneQuestion{
		ID:                uuid.New(),
		GrammarQuestionID: req.GrammarQuestionID,
		Question:          req.Question,
		Explain:           req.Explain,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := h.questionService.CreateQuestion(ctx, question); err != nil {
		h.logger.Error("grammar_choice_one_question_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create question")
		return
	}

	responseData := grammarDTO.GrammarChoiceOneQuestionResponse{
		ID:       question.ID,
		Question: question.Question,
		Explain:  question.Explain,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *GrammarChoiceOneQuestionHandler) UpdateQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req grammarDTO.UpdateGrammarChoiceOneQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("grammar_choice_one_question_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	question, err := h.questionService.GetQuestion(ctx, req.GrammarChoiceOneQuestionID)
	if err != nil {
		h.logger.Error("grammar_choice_one_question_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.GrammarChoiceOneQuestionID,
		}, "Failed to get question")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Question not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to get question")
		return
	}

	switch req.Field {
	case "question":
		question.Question = req.Value
	case "explain":
		question.Explain = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	question.UpdatedAt = time.Now()

	if err := h.questionService.UpdateQuestion(ctx, question); err != nil {
		h.logger.Error("grammar_choice_one_question_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.GrammarChoiceOneQuestionID,
		}, "Failed to update question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update question")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *GrammarChoiceOneQuestionHandler) DeleteQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("grammar_choice_one_question_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("grammar_choice_one_question_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID")
		return
	}

	if err := h.questionService.DeleteQuestion(ctx, id); err != nil {
		h.logger.Error("grammar_choice_one_question_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Question not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete question")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Question deleted successfully"})
}
