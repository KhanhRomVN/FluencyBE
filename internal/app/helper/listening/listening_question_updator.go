package listening

import (
	"context"
	listeningDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/listening"
	searchClient "fluencybe/internal/app/opensearch"
	redisClient "fluencybe/internal/app/redis"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
)

type ListeningQuestionUpdator struct {
	logger                     *logger.PrettyLogger
	redis                      *redisClient.ListeningQuestionRedis
	search                     *searchClient.ListeningQuestionSearch
	completion                 *ListeningQuestionCompletionHelper
	fillInBlankQuestionService FillInBlankQuestionService
	fillInBlankAnswerService   FillInBlankAnswerService
	choiceOneQuestionService   ChoiceOneQuestionService
	choiceOneOptionService     ChoiceOneOptionService
	choiceMultiQuestionService ChoiceMultiQuestionService
	choiceMultiOptionService   ChoiceMultiOptionService
	mapLabellingService        MapLabellingService
	matchingService            MatchingService
}

func NewListeningQuestionUpdator(
	log *logger.PrettyLogger,
	cache cache.Cache,
	openSearch *opensearch.Client,
	fillInBlankQuestionService FillInBlankQuestionService,
	fillInBlankAnswerService FillInBlankAnswerService,
	choiceOneQuestionService ChoiceOneQuestionService,
	choiceOneOptionService ChoiceOneOptionService,
	choiceMultiQuestionService ChoiceMultiQuestionService,
	choiceMultiOptionService ChoiceMultiOptionService,
	mapLabellingService MapLabellingService,
	matchingService MatchingService,
) *ListeningQuestionUpdator {
	return &ListeningQuestionUpdator{
		logger:                     log,
		redis:                      redisClient.NewListeningQuestionRedis(cache, log),
		search:                     searchClient.NewListeningQuestionSearch(openSearch, log),
		completion:                 NewListeningQuestionCompletionHelper(log),
		fillInBlankQuestionService: fillInBlankQuestionService,
		fillInBlankAnswerService:   fillInBlankAnswerService,
		choiceOneQuestionService:   choiceOneQuestionService,
		choiceOneOptionService:     choiceOneOptionService,
		choiceMultiQuestionService: choiceMultiQuestionService,
		choiceMultiOptionService:   choiceMultiOptionService,
		mapLabellingService:        mapLabellingService,
		matchingService:            matchingService,
	}
}

func (u *ListeningQuestionUpdator) UpdateCacheAndSearch(ctx context.Context, question *listening.ListeningQuestion) error {
	// Build complete question detail
	questionDetail, err := u.buildQuestionDetail(ctx, question)
	if err != nil {
		return fmt.Errorf("failed to build question detail: %w", err)
	}

	// Check completion status
	isComplete := u.completion.IsQuestionComplete(questionDetail)

	// Update Redis cache with the correct key structure
	if err := u.redis.UpdateCachedListeningQuestion(ctx, questionDetail, isComplete); err != nil {
		u.logger.Error("listening_question_updator.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update cache")
	}

	// Update OpenSearch
	status := map[bool]string{true: "complete", false: "uncomplete"}[isComplete]
	if err := u.search.UpsertListeningQuestion(ctx, questionDetail, status); err != nil {
		u.logger.Error("listening_question_updator.update_search", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update question in OpenSearch")
	}

	return nil
}

func (u *ListeningQuestionUpdator) buildQuestionDetail(ctx context.Context, question *listening.ListeningQuestion) (*listeningDTO.ListeningQuestionDetail, error) {
	response := &listeningDTO.ListeningQuestionDetail{

		ListeningQuestionResponse: listeningDTO.ListeningQuestionResponse{
			ID:          question.ID,
			Type:        question.Type,
			Topic:       question.Topic,
			Instruction: question.Instruction,
			AudioURLs:   question.AudioURLs,
			ImageURLs:   question.ImageURLs,
			Transcript:  question.Transcript,
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
	case "MAP_LABELLING":
		err = u.loadMapLabellingData(ctx, question.ID, response)
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

func (u *ListeningQuestionUpdator) loadFillInTheBlankData(ctx context.Context, questionID uuid.UUID, response *listeningDTO.ListeningQuestionDetail) error {
	u.logger.Debug("loadFillInTheBlankData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load fill in blank data")

	questions, err := u.fillInBlankQuestionService.GetQuestionsByListeningQuestionID(ctx, questionID)
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
	response.FillInTheBlankQuestion = &listeningDTO.ListeningFillInTheBlankQuestionResponse{
		ID:       question.ID,
		Question: question.Question,
	}

	answers, err := u.fillInBlankAnswerService.GetAnswersByListeningFillInTheBlankQuestionID(ctx, question.ID)
	if err != nil {
		u.logger.Error("loadFillInTheBlankData.get_answers", map[string]interface{}{
			"error":      err.Error(),
			"questionID": question.ID,
		}, "Failed to get fill in blank answers")
		return fmt.Errorf("failed to get fill in blank answers: %w", err)
	}

	response.FillInTheBlankAnswers = make([]listeningDTO.ListeningFillInTheBlankAnswerResponse, len(answers))
	for i, answer := range answers {
		response.FillInTheBlankAnswers[i] = listeningDTO.ListeningFillInTheBlankAnswerResponse{
			ID:      answer.ID,
			Answer:  answer.Answer,
			Explain: answer.Explain,
		}
	}

	return nil
}

func (u *ListeningQuestionUpdator) loadChoiceOneData(ctx context.Context, questionID uuid.UUID, response *listeningDTO.ListeningQuestionDetail) error {
	u.logger.Debug("loadChoiceOneData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load choice one data")

	questions, err := u.choiceOneQuestionService.GetQuestionsByListeningQuestionID(ctx, questionID)
	if err != nil {
		u.logger.Error("loadChoiceOneData.get_questions", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get choice one questions")
		return fmt.Errorf("failed to get choice one questions: %w", err)
	}

	if len(questions) == 0 {
		return nil
	}

	question := questions[0]
	response.ChoiceOneQuestion = &listeningDTO.ListeningChoiceOneQuestionResponse{
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

	response.ChoiceOneOptions = make([]listeningDTO.ListeningChoiceOneOptionResponse, len(options))
	for i, option := range options {
		response.ChoiceOneOptions[i] = listeningDTO.ListeningChoiceOneOptionResponse{
			ID:        option.ID,
			Options:   option.Options,
			IsCorrect: option.IsCorrect,
		}
	}

	return nil
}

func (u *ListeningQuestionUpdator) loadChoiceMultiData(ctx context.Context, questionID uuid.UUID, response *listeningDTO.ListeningQuestionDetail) error {
	u.logger.Debug("loadChoiceMultiData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load choice multi data")

	questions, err := u.choiceMultiQuestionService.GetQuestionsByListeningQuestionID(ctx, questionID)
	if err != nil {
		u.logger.Error("loadChoiceMultiData.get_questions", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get choice multi questions")
		return fmt.Errorf("failed to get choice multi questions: %w", err)
	}

	if len(questions) == 0 {
		return nil
	}

	question := questions[0]
	response.ChoiceMultiQuestion = &listeningDTO.ListeningChoiceMultiQuestionResponse{
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

	response.ChoiceMultiOptions = make([]listeningDTO.ListeningChoiceMultiOptionResponse, len(options))
	for i, option := range options {
		response.ChoiceMultiOptions[i] = listeningDTO.ListeningChoiceMultiOptionResponse{
			ID:        option.ID,
			Options:   option.Options,
			IsCorrect: option.IsCorrect,
		}
	}

	return nil
}

func (u *ListeningQuestionUpdator) loadMapLabellingData(ctx context.Context, questionID uuid.UUID, response *listeningDTO.ListeningQuestionDetail) error {
	u.logger.Debug("loadMapLabellingData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load map labelling data")

	// Get map labelling QAs directly
	qas, err := u.mapLabellingService.GetByListeningQuestionID(ctx, questionID)
	if err != nil {
		u.logger.Error("loadMapLabellingData.get_qas", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get map labelling QAs")
		return fmt.Errorf("failed to get map labelling QAs: %w", err)
	}

	if len(qas) == 0 {
		u.logger.Debug("loadMapLabellingData.not_found", map[string]interface{}{
			"questionID": questionID,
		}, "No map labelling QAs found")
		return nil
	}

	// Convert QAs to response format
	response.MapLabelling = make([]listeningDTO.ListeningMapLabellingResponse, len(qas))
	for i, qa := range qas {
		response.MapLabelling[i] = listeningDTO.ListeningMapLabellingResponse{
			ID:       qa.ID,
			Question: qa.Question,
			Answer:   qa.Answer,
			Explain:  qa.Explain,
		}
		u.logger.Debug("loadMapLabellingData.qa_detail", map[string]interface{}{
			"index":    i,
			"qaID":     qa.ID,
			"question": qa.Question,
			"answer":   qa.Answer,
		}, "Processed QA item")
	}

	u.logger.Debug("loadMapLabellingData.success", map[string]interface{}{
		"questionID": questionID,
		"qasCount":   len(qas),
	}, "Successfully loaded map labelling data")

	return nil
}

func (u *ListeningQuestionUpdator) loadMatchingData(ctx context.Context, questionID uuid.UUID, response *listeningDTO.ListeningQuestionDetail) error {
	u.logger.Debug("loadMatchingData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load matching data")

	// Get matching QAs directly
	qas, err := u.matchingService.GetByListeningQuestionID(ctx, questionID)
	if err != nil {
		u.logger.Error("loadMatchingData.get_qas", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get matching QAs")
		return fmt.Errorf("failed to get matching QAs: %w", err)
	}

	if len(qas) == 0 {
		u.logger.Debug("loadMatchingData.not_found", map[string]interface{}{
			"questionID": questionID,
		}, "No matching QAs found")
		return nil
	}

	// Convert QAs to response format
	response.Matching = make([]listeningDTO.ListeningMatchingResponse, len(qas))
	for i, qa := range qas {
		response.Matching[i] = listeningDTO.ListeningMatchingResponse{
			ID:       qa.ID,
			Question: qa.Question,
			Answer:   qa.Answer,
			Explain:  qa.Explain,
		}
		u.logger.Debug("loadMatchingData.qa_detail", map[string]interface{}{
			"index":    i,
			"qaID":     qa.ID,
			"question": qa.Question,
			"answer":   qa.Answer,
		}, "Processed QA item")
	}

	u.logger.Debug("loadMatchingData.success", map[string]interface{}{
		"questionID": questionID,
		"qasCount":   len(qas),
	}, "Successfully loaded matching data")

	return nil
}
