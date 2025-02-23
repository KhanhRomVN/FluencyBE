package listening

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	listeningDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/listening"
	ListeningRepository "fluencybe/internal/app/repository/listening"
	listeningService "fluencybe/internal/app/service/listening"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"strings"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ListeningFillInTheBlankQuestionHandler struct {
	questionService       *listeningService.ListeningFillInTheBlankQuestionService
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository
	logger                *logger.PrettyLogger
}

func NewListeningFillInTheBlankQuestionHandler(
	questionService *listeningService.ListeningFillInTheBlankQuestionService,
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository,
	logger *logger.PrettyLogger,
) *ListeningFillInTheBlankQuestionHandler {
	return &ListeningFillInTheBlankQuestionHandler{
		questionService:       questionService,
		listeningQuestionRepo: listeningQuestionRepo,
		logger:                logger,
	}
}

func (h *ListeningFillInTheBlankQuestionHandler) CreateQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req listeningDTO.CreateListeningFillInTheBlankQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_fill_in_the_blank_question_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	question := &listening.ListeningFillInTheBlankQuestion{
		ID:                  uuid.New(),
		ListeningQuestionID: req.ListeningQuestionID,
		Question:            req.Question,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := h.questionService.CreateQuestion(ctx, question); err != nil {
		h.logger.Error("listening_fill_in_the_blank_question_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")

		if strings.Contains(err.Error(), "unique_listening_fill_in_the_blank_question") {
			response.WriteError(w, http.StatusBadRequest, "A fill-in-the-blank question already exists for this listening question")
			return
		}

		response.WriteError(w, http.StatusInternalServerError, "Failed to create question")
		return
	}

	responseData := listeningDTO.ListeningFillInTheBlankQuestionResponse{
		ID:       question.ID,
		Question: question.Question,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *ListeningFillInTheBlankQuestionHandler) UpdateQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req listeningDTO.UpdateListeningFillInTheBlankQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("listening_fill_in_the_blank_question_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	question, err := h.questionService.GetQuestion(ctx, req.ListeningFillInTheBlankQuestionID)
	if err != nil {
		h.logger.Error("listening_fill_in_the_blank_question_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ListeningFillInTheBlankQuestionID,
		}, "Failed to get question")
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteError(w, http.StatusNotFound, "Question not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to get question")
		return
	}

	// Update question field based on request
	switch req.Field {
	case "question":
		question.Question = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	question.UpdatedAt = time.Now()

	if err := h.questionService.UpdateQuestion(ctx, question); err != nil {
		h.logger.Error("listening_fill_in_the_blank_question_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.ListeningFillInTheBlankQuestionID,
		}, "Failed to update question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update question")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *ListeningFillInTheBlankQuestionHandler) DeleteQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("listening_fill_in_the_blank_question_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("listening_fill_in_the_blank_question_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID")
		return
	}

	if err := h.questionService.DeleteQuestion(ctx, id); err != nil {
		h.logger.Error("listening_fill_in_the_blank_question_handler.delete", map[string]interface{}{
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
