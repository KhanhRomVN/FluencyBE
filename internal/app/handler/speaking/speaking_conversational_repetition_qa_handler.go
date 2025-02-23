package speaking

import (
	"context"
	"encoding/json"
	speakingDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/speaking"
	speakingService "fluencybe/internal/app/service/speaking"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"net/http"
	"time"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SpeakingConversationalRepetitionQAHandler struct {
	service *speakingService.SpeakingConversationalRepetitionQAService
	logger  *logger.PrettyLogger
}

func NewSpeakingConversationalRepetitionQAHandler(
	service *speakingService.SpeakingConversationalRepetitionQAService,
	logger *logger.PrettyLogger,
) *SpeakingConversationalRepetitionQAHandler {
	return &SpeakingConversationalRepetitionQAHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SpeakingConversationalRepetitionQAHandler) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.CreateSpeakingConversationalRepetitionQARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_conversational_repetition_qa_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	qa := &speaking.SpeakingConversationalRepetitionQA{
		ID:                                 uuid.New(),
		SpeakingConversationalRepetitionID: req.SpeakingConversationalRepetitionID,
		Question:                           req.Question,
		Answer:                             req.Answer,
		MeanOfQuestion:                     req.MeanOfQuestion,
		MeanOfAnswer:                       req.MeanOfAnswer,
		Explain:                            req.Explain,
		CreatedAt:                          time.Now(),
		UpdatedAt:                          time.Now(),
	}

	if err := h.service.Create(ctx, qa); err != nil {
		h.logger.Error("speaking_conversational_repetition_qa_handler.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create QA")
		response.WriteError(w, http.StatusInternalServerError, "Failed to create QA")
		return
	}

	responseData := speakingDTO.SpeakingConversationalRepetitionQAResponse{
		ID:             qa.ID,
		Question:       qa.Question,
		Answer:         qa.Answer,
		MeanOfQuestion: qa.MeanOfQuestion,
		MeanOfAnswer:   qa.MeanOfAnswer,
		Explain:        qa.Explain,
	}

	response.WriteJSON(w, http.StatusCreated, responseData)
}

func (h *SpeakingConversationalRepetitionQAHandler) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req speakingDTO.UpdateSpeakingConversationalRepetitionQARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("speaking_conversational_repetition_qa_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing record first
	existingQA, err := h.service.GetByID(ctx, req.SpeakingConversationalRepetitionQAID)
	if err != nil {
		h.logger.Error("speaking_conversational_repetition_qa_handler.update.get", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingConversationalRepetitionQAID,
		}, "Failed to get existing QA")
		response.WriteError(w, http.StatusNotFound, "QA not found")
		return
	}

	// Update only the specified field
	switch req.Field {
	case "question":
		existingQA.Question = req.Value
	case "answer":
		existingQA.Answer = req.Value
	case "mean_of_question":
		existingQA.MeanOfQuestion = req.Value
	case "mean_of_answer":
		existingQA.MeanOfAnswer = req.Value
	case "explain":
		existingQA.Explain = req.Value
	default:
		response.WriteError(w, http.StatusBadRequest, "Invalid field")
		return
	}

	existingQA.UpdatedAt = time.Now()

	if err := h.service.Update(ctx, existingQA); err != nil {
		h.logger.Error("speaking_conversational_repetition_qa_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    req.SpeakingConversationalRepetitionQAID,
		}, "Failed to update QA")
		response.WriteError(w, http.StatusInternalServerError, "Failed to update QA")
		return
	}

	response.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *SpeakingConversationalRepetitionQAHandler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("speaking_conversational_repetition_qa_handler.delete.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("speaking_conversational_repetition_qa_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid QA ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid QA ID")
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("speaking_conversational_repetition_qa_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete QA")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete QA")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{"message": "QA deleted successfully"})
}
