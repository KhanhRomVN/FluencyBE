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

type LessonHandler struct {
	service *courseSer.LessonService
	logger  *logger.PrettyLogger
}

func NewLessonHandler(
	service *courseSer.LessonService,
	logger *logger.PrettyLogger,
) *LessonHandler {
	return &LessonHandler{
		service: service,
		logger:  logger,
	}
}

func (h *LessonHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req courseDTO.CreateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("lesson_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	lesson := &course.Lesson{
		ID:        uuid.New(),
		CourseID:  req.CourseID,
		Sequence:  req.Sequence,
		Title:     req.Title,
		Overview:  req.Overview,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.service.Create(ctx, lesson); err != nil {
		h.logger.Error("lesson_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create lesson")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create lesson")
		return
	}

	responseData := courseDTO.LessonResponse{
		ID:       lesson.ID,
		CourseID: lesson.CourseID,
		Sequence: lesson.Sequence,
		Title:    lesson.Title,
		Overview: lesson.Overview,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *LessonHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req courseDTO.UpdateLessonFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("lesson_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingLesson, err := h.service.GetByID(ctx, req.LessonID)
	if err != nil {
		h.logger.Error("lesson_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.LessonID,
		}, "Failed to get existing lesson")
		response.WriteError(w, http.StatusNotFound, "Lesson not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "sequence":
		if sequence, ok := req.Value.(float64); ok {
			existingLesson.Sequence = int(sequence)
		} else {
			response.WriteError(w, http.StatusBadRequest, "Invalid sequence format")
			return
		}
	case "title":
		if title, ok := req.Value.(string); ok {
			existingLesson.Title = title
		} else {
			response.WriteError(w, http.StatusBadRequest, "Invalid title format")
			return
		}
	case "overview":
		if overview, ok := req.Value.(string); ok {
			existingLesson.Overview = overview
		} else {
			response.WriteError(w, http.StatusBadRequest, "Invalid overview format")
			return
		}
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingLesson.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingLesson); err != nil {
		h.logger.Error("lesson_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.LessonID,
		}, "Failed to update lesson")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update lesson")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *LessonHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("lesson_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("lesson_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid lesson ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid lesson ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("lesson_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete lesson")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete lesson")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Lesson deleted successfully"})
}

func (h *LessonHandler) SwapSequence(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req courseDTO.SwapSequenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("lesson_handler.swap.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.SwapSequence(ctx, req.ID1, req.ID2); err != nil {
		h.logger.Error("lesson_handler.swap", map[string]interface{}{
			"error": err.Error(),
			"id1":   req.ID1,
			"id2":   req.ID2,
		}, "Failed to swap lesson sequences")
		response.WriteError(w, http.StatusInternalServerError, "Failed to swap sequences")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Sequences swapped successfully"})
}
