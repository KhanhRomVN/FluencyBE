package writing

import (
	"context"
	"encoding/json"
	writingDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/writing"
	writingService "fluencybe/internal/app/service/writing"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"fmt"
	"net/http"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WritingQuestionHandler struct {
	service                   *writingService.WritingQuestionService
	sentenceCompletionService *writingService.WritingSentenceCompletionService
	essayService              *writingService.WritingEssayService
	logger                    *logger.PrettyLogger
}

func NewWritingQuestionHandler(
	service *writingService.WritingQuestionService,
	sentenceCompletionService *writingService.WritingSentenceCompletionService,
	essayService *writingService.WritingEssayService,
	logger *logger.PrettyLogger,
) *WritingQuestionHandler {
	return &WritingQuestionHandler{
		service:                   service,
		sentenceCompletionService: sentenceCompletionService,
		essayService:              essayService,
		logger:                    logger,
	}
}

func (h *WritingQuestionHandler) CreateWritingQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req writingDTO.CreateWritingQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("writing_question_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	question := &writing.WritingQuestion{
		ID:          uuid.New(),
		Type:        req.Type,
		Topic:       req.Topic,
		Instruction: req.Instruction,
		ImageURLs:   req.ImageURLs,
		MaxTime:     req.MaxTime,
	}

	if err := h.service.CreateQuestion(ctx, question); err != nil {
		h.logger.Error("writing_question_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create writing question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create writing question")
		return
	}

	responseData := writingDTO.WritingQuestionResponse{
		ID:          question.ID,
		Type:        question.Type,
		Topic:       question.Topic,
		Instruction: question.Instruction,
		ImageURLs:   question.ImageURLs,
		MaxTime:     question.MaxTime,
		Version:     question.Version,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *WritingQuestionHandler) GetWritingQuestionDetail(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("writing_question_handler.get.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("writing_question_handler.get.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID")
		return
	}

	questionDetail, err := h.service.GetWritingQuestionDetail(ctx, id)
	if err != nil {
		h.logger.Error("writing_question_handler.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get writing question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get writing question")
		return
	}

	response.WriteJSON(w, http.StatusOK, questionDetail)
}

func (h *WritingQuestionHandler) UpdateWritingQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("writing_question_handler.update.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("writing_question_handler.update.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID")
		return
	}

	var req writingDTO.UpdateWritingQuestionFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("writing_question_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.UpdateQuestion(ctx, id, req); err != nil {
		h.logger.Error("writing_question_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update writing question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update writing question")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *WritingQuestionHandler) DeleteWritingQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("writing_question_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("writing_question_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID")
		return
	}

	if err := h.service.DeleteQuestion(ctx, id); err != nil {
		h.logger.Error("writing_question_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete writing question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete writing question")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "Question deleted successfully"})
}

func (h *WritingQuestionHandler) GetListNewWritingQuestionByListVersionAndID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req writingDTO.GetNewUpdatesWritingQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("writing_question_handler.get_new_updates.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	versionChecks := make([]struct {
		ID      uuid.UUID
		Version int
	}, len(req.Questions))
	for i, q := range req.Questions {
		versionChecks[i].ID = q.WritingQuestionID
		versionChecks[i].Version = q.Version
	}

	questions, err := h.service.GetNewUpdatedQuestions(ctx, versionChecks)
	if err != nil {
		h.logger.Error("writing_question_handler.get_new_writing_questions", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get updated questions")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get updated questions")
		return
	}

	responseData := make([]writingDTO.WritingQuestionDetail, len(questions))
	for i, q := range questions {
		responseData[i] = writingDTO.WritingQuestionDetail{
			WritingQuestionResponse: writingDTO.WritingQuestionResponse{
				ID:          q.ID,
				Type:        q.Type,
				Topic:       q.Topic,
				Instruction: q.Instruction,
				ImageURLs:   q.ImageURLs,
				MaxTime:     q.MaxTime,
				Version:     q.Version,
			},
		}
	}

	response.WriteJSON(w, http.StatusOK, gin.H{
		"success": true,
		"data":    responseData,
	})
}

type GetListWritingByListIDRequest struct {
	QuestionIDs []string `json:"question_ids" validate:"required,min=1"`
}

func (h *WritingQuestionHandler) GetListWritingByListID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req GetListWritingByListIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("writing_question_handler.get_list_by_ids.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	questionIDs := make([]uuid.UUID, 0, len(req.QuestionIDs))
	for _, idStr := range req.QuestionIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.logger.Error("writing_question_handler.get_list_by_ids.parse_id", map[string]interface{}{
				"error": err.Error(),
				"id":    idStr,
			}, "Invalid question ID format")
			response.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid question ID format: %s", idStr))
			return
		}
		questionIDs = append(questionIDs, id)
	}

	questions, err := h.service.GetWritingByListID(ctx, questionIDs)
	if err != nil {
		h.logger.Error("writing_question_handler.get_list_by_ids", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get questions")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get questions")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{
		"success": true,
		"data":    questions,
	})
}

func (h *WritingQuestionHandler) GetListWritingQuestiondetailPaganationWithFilter(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("writing_question_handler.search.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var filter writingDTO.WritingQuestionSearchFilter
	if err := ginCtx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("writing_question_handler.search.bind", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to bind query parameters")
		response.WriteError(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 10
	}

	h.logger.Debug("writing_question_handler.search.params", map[string]interface{}{
		"type":        filter.Type,
		"topic":       filter.Topic,
		"instruction": filter.Instruction,
		"image_urls":  filter.ImageURLs,
		"max_time":    filter.MaxTime,
		"metadata":    filter.Metadata,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
	}, "Search parameters")

	questions, err := h.service.SearchQuestionsWithFilter(ctx, filter)
	if err != nil {
		h.logger.Error("writing_question_handler.search", map[string]interface{}{
			"error":  err.Error(),
			"filter": filter,
		}, "Failed to search questions")
		response.WriteError(w, http.StatusInternalServerError, "Failed to search questions")
		return
	}

	h.logger.Debug("writing_question_handler.search.results", map[string]interface{}{
		"total_results": questions.Total,
		"page":          questions.Page,
		"page_size":     questions.PageSize,
		"num_results":   len(questions.Questions),
	}, "Search results")

	response.WriteJSON(w, http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"questions": questions.Questions,
			"total":     questions.Total,
			"page":      questions.Page,
			"page_size": questions.PageSize,
		},
	})
}

func (h *WritingQuestionHandler) DeleteAllWritingData(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteAllQuestions(ctx); err != nil {
		h.logger.Error("writing_question_handler.delete_all", map[string]interface{}{"error": err.Error()}, "Failed to delete all writing data")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete all writing data")
		return
	}
	response.WriteJSON(w, http.StatusOK, gin.H{"message": "All writing data deleted successfully"})
}

func (h *WritingQuestionHandler) GetService() *writingService.WritingQuestionService {
	return h.service
}
