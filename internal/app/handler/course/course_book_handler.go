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

type CourseBookHandler struct {
	service *courseSer.CourseBookService
	logger  *logger.PrettyLogger
}

func NewCourseBookHandler(
	service *courseSer.CourseBookService,
	logger *logger.PrettyLogger,
) *CourseBookHandler {
	return &CourseBookHandler{
		service: service,
		logger:  logger,
	}
}

func (h *CourseBookHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req courseDTO.CreateCourseBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("course_book_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	courseBook := &course.CourseBook{
		ID:              uuid.New(),
		CourseID:        req.CourseID,
		Publishers:      req.Publishers,
		Authors:         req.Authors,
		PublicationYear: req.PublicationYear,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.service.Create(ctx, courseBook); err != nil {
		h.logger.Error("course_book_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create course book")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create course book")
		return
	}

	responseData := courseDTO.CourseBookResponse{
		ID:              courseBook.ID,
		CourseID:        courseBook.CourseID,
		Publishers:      courseBook.Publishers,
		Authors:         courseBook.Authors,
		PublicationYear: courseBook.PublicationYear,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *CourseBookHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req courseDTO.UpdateCourseBookFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("course_book_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingBook, err := h.service.GetByID(ctx, req.CourseBookID)
	if err != nil {
		h.logger.Error("course_book_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.CourseBookID,
		}, "Failed to get existing course book")
		response.WriteError(w, http.StatusNotFound, "Course book not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "publishers":
		if publishers, ok := req.Value.([]string); ok {
			existingBook.Publishers = publishers
		} else {
			response.WriteError(w, http.StatusBadRequest, "Invalid publishers format")
			return
		}
	case "authors":
		if authors, ok := req.Value.([]string); ok {
			existingBook.Authors = authors
		} else {
			response.WriteError(w, http.StatusBadRequest, "Invalid authors format")
			return
		}
	case "publication_year":
		if year, ok := req.Value.(float64); ok {
			existingBook.PublicationYear = int(year)
		} else {
			response.WriteError(w, http.StatusBadRequest, "Invalid publication year format")
			return
		}
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingBook.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingBook); err != nil {
		h.logger.Error("course_book_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.CourseBookID,
		}, "Failed to update course book")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update course book")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *CourseBookHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("course_book_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("course_book_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid course book ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid course book ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("course_book_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete course book")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete course book")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Course book deleted successfully"})
}
