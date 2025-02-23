package reading

import (
	"context"
	"encoding/json"
	"errors"
	readingDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/reading"
	readingService "fluencybe/internal/app/service/reading"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/response"
	"fmt"
	"net/http"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReadingQuestionHandler struct {
	service                    *readingService.ReadingQuestionService
	fillInBlankQuestionService *readingService.ReadingFillInTheBlankQuestionService
	fillInBlankAnswerService   *readingService.ReadingFillInTheBlankAnswerService
	choiceOneQuestionService   *readingService.ReadingChoiceOneQuestionService
	choiceOneOptionService     *readingService.ReadingChoiceOneOptionService
	choiceMultiQuestionService *readingService.ReadingChoiceMultiQuestionService
	choiceMultiOptionService   *readingService.ReadingChoiceMultiOptionService
	trueFalseService           *readingService.ReadingTrueFalseService
	matchingService            *readingService.ReadingMatchingService
	logger                     *logger.PrettyLogger
}

func NewReadingQuestionHandler(
	service *readingService.ReadingQuestionService,
	fillInBlankQuestionService *readingService.ReadingFillInTheBlankQuestionService,
	fillInBlankAnswerService *readingService.ReadingFillInTheBlankAnswerService,
	choiceOneQuestionService *readingService.ReadingChoiceOneQuestionService,
	choiceOneOptionService *readingService.ReadingChoiceOneOptionService,
	choiceMultiQuestionService *readingService.ReadingChoiceMultiQuestionService,
	choiceMultiOptionService *readingService.ReadingChoiceMultiOptionService,
	trueFalseService *readingService.ReadingTrueFalseService,
	matchingService *readingService.ReadingMatchingService,
	logger *logger.PrettyLogger,
) *ReadingQuestionHandler {
	return &ReadingQuestionHandler{
		service:                    service,
		fillInBlankQuestionService: fillInBlankQuestionService,
		fillInBlankAnswerService:   fillInBlankAnswerService,
		choiceOneQuestionService:   choiceOneQuestionService,
		choiceOneOptionService:     choiceOneOptionService,
		choiceMultiQuestionService: choiceMultiQuestionService,
		choiceMultiOptionService:   choiceMultiOptionService,
		trueFalseService:           trueFalseService,
		matchingService:            matchingService,
		logger:                     logger,
	}
}

func (h *ReadingQuestionHandler) CreateReadingQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req readingDTO.CreateReadingQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("reading_question_handler.create.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Invalid request format")
		response.WriteError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	question := &reading.ReadingQuestion{
		ID:          uuid.New(),
		Type:        req.Type,
		Topic:       req.Topic,
		Instruction: req.Instruction,
		Title:       req.Title,
		Passages:    req.Passages,
		ImageURLs:   req.ImageURLs,
		MaxTime:     req.MaxTime,
		Version:     1,
	}

	if err := h.service.CreateQuestion(ctx, question); err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, readingService.ErrInvalidInput) {
			code = http.StatusBadRequest
		}
		h.logger.Error("reading_question_handler.create", map[string]interface{}{
			"error":         err.Error(),
			"question_type": question.Type,
		}, "Failed to create reading question")
		response.WriteError(w, code, "Failed to create reading question")
		return
	}

	responseData := readingDTO.ReadingQuestionResponse{
		ID:          question.ID,
		Type:        question.Type,
		Topic:       question.Topic,
		Instruction: question.Instruction,
		Title:       question.Title,
		Passages:    question.Passages,
		ImageURLs:   question.ImageURLs,
		MaxTime:     question.MaxTime,
		Version:     question.Version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseData)
}

func (h *ReadingQuestionHandler) GetReadingQuestionType(ctx context.Context, questionID uuid.UUID, questionType string, response *readingDTO.ReadingQuestionDetail) error {
	switch questionType {
	case "FILL_IN_THE_BLANK":
		if err := h.loadFillInTheBlankData(ctx, questionID, response); err != nil {
			return fmt.Errorf("failed to load fill in blank data: %w", err)
		}
	case "CHOICE_ONE":
		if err := h.loadChoiceOneData(ctx, questionID, response); err != nil {
			return fmt.Errorf("failed to load choice one data: %w", err)
		}
	case "CHOICE_MULTI":
		if err := h.loadChoiceMultiData(ctx, questionID, response); err != nil {
			return fmt.Errorf("failed to load choice multi data: %w", err)
		}
	case "TRUE_FALSE":
		if err := h.loadTrueFalseData(ctx, questionID, response); err != nil {
			return fmt.Errorf("failed to load true/false data: %w", err)
		}
	case "MATCHING":
		if err := h.loadMatchingData(ctx, questionID, response); err != nil {
			return fmt.Errorf("failed to load matching data: %w", err)
		}
	default:
		return fmt.Errorf("unknown question type: %s", questionType)
	}
	return nil
}

func (h *ReadingQuestionHandler) loadFillInTheBlankData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	questions, err := h.fillInBlankQuestionService.GetQuestionsByReadingQuestionID(ctx, questionID)
	if err != nil {
		return err
	}

	if len(questions) > 0 {
		question := questions[0]
		response.FillInTheBlankQuestion = &readingDTO.ReadingFillInTheBlankQuestionResponse{
			ID:       question.ID,
			Question: question.Question,
		}

		answers, err := h.fillInBlankAnswerService.GetAnswersByReadingFillInTheBlankQuestionID(ctx, question.ID)
		if err != nil {
			return err
		}

		response.FillInTheBlankAnswers = make([]readingDTO.ReadingFillInTheBlankAnswerResponse, len(answers))
		for i, answer := range answers {
			response.FillInTheBlankAnswers[i] = readingDTO.ReadingFillInTheBlankAnswerResponse{
				ID:      answer.ID,
				Answer:  answer.Answer,
				Explain: answer.Explain,
			}
		}
	}
	return nil
}

func (h *ReadingQuestionHandler) loadChoiceOneData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	questions, err := h.choiceOneQuestionService.GetQuestionsByReadingQuestionID(ctx, questionID)
	if err != nil {
		return err
	}

	if len(questions) > 0 {
		question := questions[0]
		response.ChoiceOneQuestion = &readingDTO.ReadingChoiceOneQuestionResponse{
			ID:       question.ID,
			Question: question.Question,
			Explain:  question.Explain,
		}

		options, err := h.choiceOneOptionService.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			return err
		}

		response.ChoiceOneOptions = make([]readingDTO.ReadingChoiceOneOptionResponse, len(options))
		for i, option := range options {
			response.ChoiceOneOptions[i] = readingDTO.ReadingChoiceOneOptionResponse{
				ID:        option.ID,
				Options:   option.Options,
				IsCorrect: option.IsCorrect,
			}
		}
	}
	return nil
}

func (h *ReadingQuestionHandler) loadChoiceMultiData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	questions, err := h.choiceMultiQuestionService.GetQuestionsByReadingQuestionID(ctx, questionID)
	if err != nil {
		return err
	}

	if len(questions) > 0 {
		question := questions[0]
		response.ChoiceMultiQuestion = &readingDTO.ReadingChoiceMultiQuestionResponse{
			ID:       question.ID,
			Question: question.Question,
			Explain:  question.Explain,
		}

		options, err := h.choiceMultiOptionService.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			return err
		}

		response.ChoiceMultiOptions = make([]readingDTO.ReadingChoiceMultiOptionResponse, len(options))
		for i, option := range options {
			response.ChoiceMultiOptions[i] = readingDTO.ReadingChoiceMultiOptionResponse{
				ID:        option.ID,
				Options:   option.Options,
				IsCorrect: option.IsCorrect,
			}
		}
	}
	return nil
}

func (h *ReadingQuestionHandler) loadTrueFalseData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	trueFalses, err := h.trueFalseService.GetByReadingQuestionID(ctx, questionID)
	if err != nil {
		return err
	}

	response.TrueFalse = make([]readingDTO.ReadingTrueFalseResponse, len(trueFalses))
	for i, tf := range trueFalses {
		response.TrueFalse[i] = readingDTO.ReadingTrueFalseResponse{
			ID:       tf.ID,
			Question: tf.Question,
			Answer:   tf.Answer,
			Explain:  tf.Explain,
		}
	}
	return nil
}

func (h *ReadingQuestionHandler) loadMatchingData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	matchings, err := h.matchingService.GetByReadingQuestionID(ctx, questionID)
	if err != nil {
		h.logger.Error("loadMatchingData.get_qas", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get matching QAs")
		return err
	}

	if len(matchings) == 0 {
		return nil
	}

	response.Matching = make([]readingDTO.ReadingMatchingResponse, len(matchings))
	for i, matching := range matchings {
		response.Matching[i] = readingDTO.ReadingMatchingResponse{
			ID:       matching.ID,
			Question: matching.Question,
			Answer:   matching.Answer,
			Explain:  matching.Explain,
		}
	}
	return nil
}

func (h *ReadingQuestionHandler) GetReadingQuestionDetail(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtxValue := ctx.Value(constants.GinContextKey)
	if ginCtxValue == nil {
		h.logger.Error("reading_question_handler.get", nil, "GinContextKey not found in context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	ginCtx, ok := ginCtxValue.(*gin.Context)
	if !ok {
		h.logger.Error("reading_question_handler.get", nil, "Failed to convert context value to *gin.Context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("reading_question_handler.get.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID format")
		return
	}

	responseData, err := h.service.GetReadingQuestionDetail(ctx, id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, readingService.ErrQuestionNotFound) {
			statusCode = http.StatusNotFound
		}
		h.logger.Error("reading_question_handler.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get reading question")
		response.WriteError(w, statusCode, "Failed to get reading question")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

func (h *ReadingQuestionHandler) UpdateReadingQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtxValue := ctx.Value(constants.GinContextKey)
	if ginCtxValue == nil {
		h.logger.Error("reading_question_handler.update", nil, "GinContextKey not found in context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	ginCtx, ok := ginCtxValue.(*gin.Context)
	if !ok {
		h.logger.Error("reading_question_handler.update", nil, "Failed to convert context value to *gin.Context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("reading_question_handler.update.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID format")
		return
	}

	var req readingDTO.UpdateReadingQuestionFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("reading_question_handler.update.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.UpdateQuestion(ctx, id, req); err != nil {
		h.logger.Error("reading_question_handler.update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update reading question")

		statusCode := http.StatusInternalServerError
		if errors.Is(err, readingService.ErrQuestionNotFound) {
			statusCode = http.StatusNotFound
		}
		response.WriteError(w, statusCode, "Failed to update reading question")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ReadingQuestionHandler) DeleteReadingQuestion(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtxValue := ctx.Value(constants.GinContextKey)
	if ginCtxValue == nil {
		h.logger.Error("reading_question_handler.delete", nil, "GinContextKey not found in context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	ginCtx, ok := ginCtxValue.(*gin.Context)
	if !ok {
		h.logger.Error("reading_question_handler.delete", nil, "Failed to convert context value to *gin.Context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := ginCtx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("reading_question_handler.delete.parse_id", map[string]interface{}{
			"error": err.Error(),
			"id":    idStr,
		}, "Invalid question ID format")
		response.WriteError(w, http.StatusBadRequest, "Invalid question ID format")
		return
	}

	if err := h.service.DeleteQuestion(ctx, id); err != nil {
		h.logger.Error("reading_question_handler.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete reading question")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete reading question")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ReadingQuestionHandler) GetListNewReadingQuestionByListVersionAndID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req readingDTO.GetNewUpdatesReadingQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("reading_question_handler.get_new_updates.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert request to internal format
	versionChecks := make([]struct {
		ID      uuid.UUID
		Version int
	}, len(req.Questions))
	for i, q := range req.Questions {
		versionChecks[i].ID = q.ReadingQuestionID
		versionChecks[i].Version = q.Version
	}

	questions, err := h.service.GetNewUpdatedQuestions(ctx, versionChecks)
	if err != nil {
		h.logger.Error("reading_question_handler.get_new_reading_questions", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get updated questions")
		response.WriteError(w, http.StatusInternalServerError, "Failed to get updated questions")
		return
	}

	// Convert to response format
	responseData := make([]readingDTO.ReadingQuestionDetail, len(questions))
	for i, q := range questions {
		responseData[i] = readingDTO.ReadingQuestionDetail{
			ReadingQuestionResponse: readingDTO.ReadingQuestionResponse{
				ID:          q.ID,
				Type:        q.Type,
				Topic:       q.Topic,
				Instruction: q.Instruction,
				Title:       q.Title,
				Passages:    q.Passages,
				ImageURLs:   q.ImageURLs,
				MaxTime:     q.MaxTime,
				Version:     q.Version,
			},
		}
		if err := h.GetReadingQuestionType(ctx, q.ID, q.Type, &responseData[i]); err != nil {
			h.logger.Warning("reading_question_handler.get_new_reading_questions.load_type", map[string]interface{}{
				"error": err.Error(),
				"id":    q.ID,
				"type":  q.Type,
			}, "Failed to load question type data")
			continue
		}
	}

	response.WriteJSON(w, http.StatusOK, gin.H{
		"success": true,
		"data":    responseData,
	})
}

func (h *ReadingQuestionHandler) GetService() *readingService.ReadingQuestionService {
	return h.service
}

type GetListReadingByListIDRequest struct {
	QuestionIDs []string `json:"question_ids" validate:"required,min=1"`
}

func (h *ReadingQuestionHandler) GetListReadingQuestiondetailPaganationWithFilter(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("reading_question_handler.search.context", nil, "Failed to get gin context")
		response.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var filter readingDTO.ReadingQuestionSearchFilter
	if err := ginCtx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("reading_question_handler.search.bind", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to bind query parameters")
		response.WriteError(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	// Add debug logging for filter values
	h.logger.Debug("reading_question_handler.search.filter", map[string]interface{}{
		"type":        filter.Type,
		"topic":       filter.Topic,
		"instruction": filter.Instruction,
		"title":       filter.Title,
		"passages":    filter.Passages,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
	}, "Search filter values")

	result, err := h.service.SearchQuestionsWithFilter(ctx, filter)
	if err != nil {
		h.logger.Error("reading_question_handler.search", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to search questions")
		response.WriteError(w, http.StatusInternalServerError, "Failed to search questions")
		return
	}

	response.WriteJSON(w, http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

func (h *ReadingQuestionHandler) DeleteAllReadingData(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteAllQuestions(ctx); err != nil {
		h.logger.Error("reading_question_handler.delete_all", map[string]interface{}{"error": err.Error()}, "Failed to delete all reading data")
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete all reading data")
		return
	}
	response.WriteJSON(w, http.StatusOK, gin.H{"message": "All reading data deleted successfully"})
}

func (h *ReadingQuestionHandler) GetListReadingByListID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req GetListReadingByListIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("reading_question_handler.get_list_by_ids.decode", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to decode request body")
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert string IDs to UUIDs
	questionIDs := make([]uuid.UUID, 0, len(req.QuestionIDs))
	for _, idStr := range req.QuestionIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.logger.Error("reading_question_handler.get_list_by_ids.parse_id", map[string]interface{}{
				"error": err.Error(),
				"id":    idStr,
			}, "Invalid question ID format")
			response.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid question ID format: %s", idStr))
			return
		}
		questionIDs = append(questionIDs, id)
	}

	// Get questions with details
	questions, err := h.service.GetReadingByListID(ctx, questionIDs)
	if err != nil {
		h.logger.Error("reading_question_handler.get_list_by_ids", map[string]interface{}{
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
