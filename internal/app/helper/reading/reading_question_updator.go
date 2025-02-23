package reading

import (
	"context"
	readingDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/reading"
	searchClient "fluencybe/internal/app/opensearch"
	redisClient "fluencybe/internal/app/redis"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
)

type ReadingQuestionUpdator struct {
	logger                     *logger.PrettyLogger
	redis                      *redisClient.ReadingQuestionRedis
	search                     *searchClient.ReadingQuestionSearch
	completion                 *ReadingQuestionCompletionHelper
	fillInBlankQuestionService FillInBlankQuestionService
	fillInBlankAnswerService   FillInBlankAnswerService
	choiceOneQuestionService   ChoiceOneQuestionService
	choiceOneOptionService     ChoiceOneOptionService
	choiceMultiQuestionService ChoiceMultiQuestionService
	choiceMultiOptionService   ChoiceMultiOptionService
	trueFalseService           TrueFalseService
	matchingService            MatchingService
}

func NewReadingQuestionUpdator(
	log *logger.PrettyLogger,
	cache cache.Cache,
	openSearch *opensearch.Client,
	fillInBlankQuestionService FillInBlankQuestionService,
	fillInBlankAnswerService FillInBlankAnswerService,
	choiceOneQuestionService ChoiceOneQuestionService,
	choiceOneOptionService ChoiceOneOptionService,
	choiceMultiQuestionService ChoiceMultiQuestionService,
	choiceMultiOptionService ChoiceMultiOptionService,
	trueFalseService TrueFalseService,
	matchingService MatchingService,
) *ReadingQuestionUpdator {
	return &ReadingQuestionUpdator{
		logger:                     log,
		redis:                      redisClient.NewReadingQuestionRedis(cache, log),
		search:                     searchClient.NewReadingQuestionSearch(openSearch, log),
		completion:                 NewReadingQuestionCompletionHelper(log),
		fillInBlankQuestionService: fillInBlankQuestionService,
		fillInBlankAnswerService:   fillInBlankAnswerService,
		choiceOneQuestionService:   choiceOneQuestionService,
		choiceOneOptionService:     choiceOneOptionService,
		choiceMultiQuestionService: choiceMultiQuestionService,
		choiceMultiOptionService:   choiceMultiOptionService,
		trueFalseService:           trueFalseService,
		matchingService:            matchingService,
	}
}

func (u *ReadingQuestionUpdator) UpdateCacheAndSearch(ctx context.Context, question *reading.ReadingQuestion) error {
	// Build complete question detail
	questionDetail, err := u.buildQuestionDetail(ctx, question)
	if err != nil {
		return fmt.Errorf("failed to build question detail: %w", err)
	}

	// Check completion status
	isComplete := u.completion.IsQuestionComplete(questionDetail)

	// Update Redis cache with the correct key structure
	if err := u.redis.UpdateCachedReadingQuestion(ctx, questionDetail, isComplete); err != nil {
		u.logger.Error("reading_question_updator.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update cache")
	}

	// Update OpenSearch
	status := map[bool]string{true: "complete", false: "uncomplete"}[isComplete]
	if err := u.search.IndexReadingQuestionDetail(ctx, questionDetail, status); err != nil {
		u.logger.Error("reading_question_updator.update_search", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update question in OpenSearch")
	}

	return nil
}

func (u *ReadingQuestionUpdator) buildQuestionDetail(ctx context.Context, question *reading.ReadingQuestion) (*readingDTO.ReadingQuestionDetail, error) {
	response := &readingDTO.ReadingQuestionDetail{
		ReadingQuestionResponse: readingDTO.ReadingQuestionResponse{
			ID:          question.ID,
			Type:        question.Type,
			Topic:       question.Topic,
			Instruction: question.Instruction,
			Title:       question.Title,
			Passages:    question.Passages,
			ImageURLs:   question.ImageURLs,
			MaxTime:     question.MaxTime,
			Version:     question.Version,
		},
	}

	var err error
	switch question.Type {
	case "FILL_IN_THE_BLANK":
		err = u.loadFillInTheBlankData(ctx, question.ID, response)
	case "CHOICE_ONE":
		err = u.loadChoiceOneData(ctx, question.ID, response)
	case "CHOICE_MULTI":
		err = u.loadChoiceMultiData(ctx, question.ID, response)
	case "TRUE_FALSE":
		err = u.loadTrueFalseData(ctx, question.ID, response)
	case "MATCHING":
		err = u.loadMatchingData(ctx, question.ID, response)
	default:
		u.logger.Error("buildQuestionDetail.unknown_type", map[string]interface{}{
			"type": question.Type,
		}, "Unknown question type")
		return nil, fmt.Errorf("unknown question type: %s", question.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load %s data: %w", question.Type, err)
	}

	return response, nil
}

func (u *ReadingQuestionUpdator) loadFillInTheBlankData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	u.logger.Debug("loadFillInTheBlankData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load fill in blank data")

	questions, err := u.fillInBlankQuestionService.GetQuestionsByReadingQuestionID(ctx, questionID)
	if err != nil {
		u.logger.Error("loadFillInTheBlankData.get_questions", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get fill in blank questions")
		return fmt.Errorf("failed to get fill in blank questions: %w", err)
	}

	if len(questions) == 0 {
		return nil
	}

	question := questions[0]
	response.FillInTheBlankQuestion = &readingDTO.ReadingFillInTheBlankQuestionResponse{
		ID:       question.ID,
		Question: question.Question,
	}

	answers, err := u.fillInBlankAnswerService.GetAnswersByReadingFillInTheBlankQuestionID(ctx, question.ID)
	if err != nil {
		u.logger.Error("loadFillInTheBlankData.get_answers", map[string]interface{}{
			"error":      err.Error(),
			"questionID": question.ID,
		}, "Failed to get fill in blank answers")
		return fmt.Errorf("failed to get fill in blank answers: %w", err)
	}

	response.FillInTheBlankAnswers = make([]readingDTO.ReadingFillInTheBlankAnswerResponse, len(answers))
	for i, answer := range answers {
		response.FillInTheBlankAnswers[i] = readingDTO.ReadingFillInTheBlankAnswerResponse{
			ID:      answer.ID,
			Answer:  answer.Answer,
			Explain: answer.Explain,
		}
	}

	return nil
}

func (u *ReadingQuestionUpdator) loadChoiceOneData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	u.logger.Debug("loadChoiceOneData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load choice one data")

	questions, err := u.choiceOneQuestionService.GetQuestionsByReadingQuestionID(ctx, questionID)
	if err != nil {
		u.logger.Error("loadChoiceOneData.get_questions", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get choice one questions")
		return fmt.Errorf("failed to get choice one questions: %w", err)
	}

	if len(questions) > 0 {
		question := questions[0]
		response.ChoiceOneQuestion = &readingDTO.ReadingChoiceOneQuestionResponse{
			ID:       question.ID,
			Question: question.Question,
			Explain:  question.Explain,
		}

		options, err := u.choiceOneOptionService.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			u.logger.Error("loadChoiceOneData.get_options", map[string]interface{}{
				"error":      err.Error(),
				"questionID": question.ID,
			}, "Failed to get choice one options")
			return fmt.Errorf("failed to get choice one options: %w", err)
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

func (u *ReadingQuestionUpdator) loadChoiceMultiData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	u.logger.Debug("loadChoiceMultiData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load choice multi data")

	questions, err := u.choiceMultiQuestionService.GetQuestionsByReadingQuestionID(ctx, questionID)
	if err != nil {
		u.logger.Error("loadChoiceMultiData.get_questions", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get choice multi questions")
		return fmt.Errorf("failed to get choice multi questions: %w", err)
	}

	if len(questions) > 0 {
		question := questions[0]
		response.ChoiceMultiQuestion = &readingDTO.ReadingChoiceMultiQuestionResponse{
			ID:       question.ID,
			Question: question.Question,
			Explain:  question.Explain,
		}

		options, err := u.choiceMultiOptionService.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			u.logger.Error("loadChoiceMultiData.get_options", map[string]interface{}{
				"error":      err.Error(),
				"questionID": question.ID,
			}, "Failed to get choice multi options")
			return fmt.Errorf("failed to get choice multi options: %w", err)
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

func (u *ReadingQuestionUpdator) loadTrueFalseData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	u.logger.Debug("loadTrueFalseData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load true/false data")

	trueFalses, err := u.trueFalseService.GetByReadingQuestionID(ctx, questionID)
	if err != nil {
		u.logger.Error("loadTrueFalseData.get_questions", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get true/false questions")
		return fmt.Errorf("failed to get true/false questions: %w", err)
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

func (u *ReadingQuestionUpdator) loadMatchingData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	u.logger.Debug("loadMatchingData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load matching data")

	matchings, err := u.matchingService.GetByReadingQuestionID(ctx, questionID)
	if err != nil {
		u.logger.Error("loadMatchingData.get_matchings", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get matching questions")
		return fmt.Errorf("failed to get matching questions: %w", err)
	}

	// Convert all matchings to response format
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

func (u *ReadingQuestionUpdator) BuildQuestionDetail(ctx context.Context, question *reading.ReadingQuestion) (*readingDTO.ReadingQuestionDetail, error) {
	response := &readingDTO.ReadingQuestionDetail{
		ReadingQuestionResponse: readingDTO.ReadingQuestionResponse{
			ID:          question.ID,
			Type:        question.Type,
			Topic:       question.Topic,
			Instruction: question.Instruction,
			Title:       question.Title,
			Passages:    question.Passages,
			ImageURLs:   question.ImageURLs,
			MaxTime:     question.MaxTime,
			Version:     question.Version,
		},
	}

	var err error
	switch question.Type {
	case "FILL_IN_THE_BLANK":
		err = u.loadFillInTheBlankData(ctx, question.ID, response)
	case "CHOICE_ONE":
		err = u.loadChoiceOneData(ctx, question.ID, response)
	case "CHOICE_MULTI":
		err = u.loadChoiceMultiData(ctx, question.ID, response)
	case "TRUE_FALSE":
		err = u.loadTrueFalseData(ctx, question.ID, response)
	case "MATCHING":
		err = u.loadMatchingData(ctx, question.ID, response)
	default:
		u.logger.Error("buildQuestionDetail.unknown_type", map[string]interface{}{
			"type": question.Type,
		}, "Unknown question type")
		return nil, fmt.Errorf("unknown question type: %s", question.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load %s data: %w", question.Type, err)
	}

	return response, nil
}
