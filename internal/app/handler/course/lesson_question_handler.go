package course

import (
	"context"
	"encoding/json"
	courseDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/course"
	courseSer "fluencybe/internal/app/service/course"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LessonQuestionHandler struct {
	service *courseSer.LessonQuestionService
	logger  *logger.PrettyLogger
}

func NewLessonQuestionHandler(
	service *courseSer.LessonQuestionService,
	logger *logger.PrettyLogger,
) *LessonQuestionHandler {
	return &LessonQuestionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *LessonQuestionHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req courseDTO.CreateLessonQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("lesson_question_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	lessonQuestion := &course.LessonQuestion{
		ID:           uuid.New(),
		LessonID:     req.LessonID,
		Sequence:     req.Sequence,
		QuestionID:   req.QuestionID,
		QuestionType: req.QuestionType,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.service.Create(ctx, lessonQuestion); err != nil {
		h.logger.Error("lesson_question_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create lesson question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create lesson question")
		return
	}

	responseData := courseDTO.LessonQuestionResponse{
		ID:           lessonQuestion.ID,
		LessonID:     lessonQuestion.LessonID,
		Sequence:     lessonQuestion.Sequence,
		QuestionID:   lessonQuestion.QuestionID,
		QuestionType: lessonQuestion.QuestionType,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *LessonQuestionHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req courseDTO.UpdateLessonQuestionFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("lesson_question_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingQuestion, err := h.service.GetByID(ctx, req.LessonQuestionID)
	if err != nil {
		h.logger.Error("lesson_question_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.LessonQuestionID,
		}, "Failed to get existing lesson question")
		response.WriteError(w, http.StatusNotFound, "Lesson question not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "sequence":
		if sequence, ok := req.Value.(float64); ok {
			existingQuestion.Sequence = int(sequence)
		} else {
			response.WriteError(w, http.StatusBadRequest, "Invalid sequence format")
			return
		}
	case "question_id":
		if questionID, ok := req.Value.(string); ok {
			id, err := uuid.Parse(questionID)
			if err != nil {
				response.WriteError(w, http.StatusBadRequest, "Invalid question ID format")
				return
			}
			existingQuestion.QuestionID = id
		} else {
			response.WriteError(w, http.StatusBadRequest, "Invalid question ID format")
			return
		}
	case "question_type":
		if questionType, ok := req.Value.(string); ok {
			existingQuestion.QuestionType = questionType
		} else {
			response.WriteError(w, http.StatusBadRequest, "Invalid question type format")
			return
		}
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingQuestion.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingQuestion); err != nil {
		h.logger.Error("lesson_question_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.LessonQuestionID,
		}, "Failed to update lesson question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update lesson question")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *LessonQuestionHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("lesson_question_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("lesson_question_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid lesson question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid lesson question ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("lesson_question_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete lesson question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete lesson question")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Lesson question deleted successfully"})
}

func (h *LessonQuestionHandler) SwapSequence(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req courseDTO.SwapSequenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("lesson_question_handler.swap.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.SwapSequence(ctx, req.ID1, req.ID2); err != nil {
		h.logger.Error("lesson_question_handler.swap", map[string]interface{}{
			"error": err.Error(),
			"id1":   req.ID1,
			"id2":   req.ID2,
		}, "Failed to swap lesson question sequences")
		response.WriteError(w, http.StatusInternalServerError, "Failed to swap sequences")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Sequences swapped successfully"})
}
