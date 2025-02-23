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

type CourseHandler struct {
	service               *courseSer.CourseService
	courseBookService     *courseSer.CourseBookService
	courseOtherService    *courseSer.CourseOtherService
	lessonService         *courseSer.LessonService
	lessonQuestionService *courseSer.LessonQuestionService
	logger                *logger.PrettyLogger
}

func NewCourseHandler(
	service *courseSer.CourseService,
	courseBookService *courseSer.CourseBookService,
	courseOtherService *courseSer.CourseOtherService,
	lessonService *courseSer.LessonService,
	lessonQuestionService *courseSer.LessonQuestionService,
	logger *logger.PrettyLogger,
) *CourseHandler {
	return &CourseHandler{
		service:               service,
		courseBookService:     courseBookService,
		courseOtherService:    courseOtherService,
		lessonService:         lessonService,
		lessonQuestionService: lessonQuestionService,
		logger:                logger,
	}
}

func (h *CourseHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req courseDTO.CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("course_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	course := &course.Course{
		ID:        uuid.New(),
		Type:      req.Type,
		Title:     req.Title,
		Overview:  req.Overview,
		Skills:    req.Skills,
		Band:      req.Band,
		ImageURLs: req.ImageURLs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.service.Create(ctx, course); err != nil {
		h.logger.Error("course_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create course")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create course")
		return
	}

	responseData := courseDTO.CourseResponse{
		ID:        course.ID,
		Type:      course.Type,
		Title:     course.Title,
		Overview:  course.Overview,
		Skills:    course.Skills,
		Band:      course.Band,
		ImageURLs: course.ImageURLs,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *CourseHandler) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("course_handler.get.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("course_handler.get.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid course ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	courseDetail, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("course_handler.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get course")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get course")
		return
	}

	response.WriteJSON(w, http.StatusOK, courseDetail)
}

func (h *CourseHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("course_handler.update.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("course_handler.update.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid course ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	var req courseDTO.UpdateCourseFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("course_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.Update(ctx, id, req); err != nil {
		h.logger.Error("course_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update course")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update course")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *CourseHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("course_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("course_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid course ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("course_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete course")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete course")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Course deleted successfully"})
}

func (h *CourseHandler) Search(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("course_handler.search.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var filter courseDTO.CourseSearchRequest
	if err := ginCtx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("course_handler.search.bind", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to bind query parameters")
		response.WriteError(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	// Validate required fields
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 10
	}

	// Log search parameters for debugging
	h.logger.Debug("course_handler.search.params", map[string]interface{}{
		"type":      filter.Type,
		"title":     filter.Title,
		"skills":    filter.Skills,
		"band":      filter.Band,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	}, "Search parameters")

	result, err := h.service.SearchWithFilter(ctx, filter)
	if err != nil {
		h.logger.Error("course_handler.search", map[string]interface{}{
			"error":  err.Error(),
			"filter": filter,
		}, "Failed to search courses")
		response.WriteError(w, http.StatusInternalServerError, "Failed to search courses")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"courses":   result.Courses,
			"total":     result.Total,
			"page":      result.Page,
			"page_size": result.PageSize,
		},
	})
}

func (h *CourseHandler) DeleteAllCourseData(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteAllCourseData(ctx); err != nil {
		h.logger.Error("course_handler.delete_all", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete all course data")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete all course data")
		return
	}
	response.WriteJSON(w, http.StatusOK, gin.H{"message": "All course data deleted successfully"})
}
