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

type CourseOtherHandler struct {
	service *courseSer.CourseOtherService
	logger  *logger.PrettyLogger
}

func NewCourseOtherHandler(
	service *courseSer.CourseOtherService,
	logger *logger.PrettyLogger,
) *CourseOtherHandler {
	return &CourseOtherHandler{
		service: service,
		logger:  logger,
	}
}

func (h *CourseOtherHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req courseDTO.CreateCourseOtherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("course_other_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	courseOther := &course.CourseOther{
		ID:        uuid.New(),
		CourseID:  req.CourseID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.service.Create(ctx, courseOther); err != nil {
		h.logger.Error("course_other_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create course other")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create course other")
		return
	}

	responseData := courseDTO.CourseOtherResponse{
		ID:       courseOther.ID,
		CourseID: courseOther.CourseID,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *CourseOtherHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("course_other_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("course_other_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid course other ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid course other ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("course_other_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete course other")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete course other")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Course other deleted successfully"})
}
