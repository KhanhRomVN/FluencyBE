package grammar

import (
	"context"
	grammarDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/grammar"
	searchClient "fluencybe/internal/app/opensearch"
	redisClient "fluencybe/internal/app/redis"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
)

type GrammarQuestionUpdator struct {
	logger                        *logger.PrettyLogger
	redis                         *redisClient.GrammarQuestionRedis
	search                        *searchClient.GrammarQuestionSearch
	completion                    *GrammarQuestionCompletionHelper
	fillInBlankQuestionService    FillInBlankQuestionService
	fillInBlankAnswerService      FillInBlankAnswerService
	choiceOneQuestionService      ChoiceOneQuestionService
	choiceOneOptionService        ChoiceOneOptionService
	errorIdentificationService    ErrorIdentificationService
	sentenceTransformationService SentenceTransformationService
}

func NewGrammarQuestionUpdator(
	log *logger.PrettyLogger,
	cache cache.Cache,
	openSearch *opensearch.Client,
	fillInBlankQuestionService FillInBlankQuestionService,
	fillInBlankAnswerService FillInBlankAnswerService,
	choiceOneQuestionService ChoiceOneQuestionService,
	choiceOneOptionService ChoiceOneOptionService,
	errorIdentificationService ErrorIdentificationService,
	sentenceTransformationService SentenceTransformationService,
) *GrammarQuestionUpdator {
	return &GrammarQuestionUpdator{
		logger:                        log,
		redis:                         redisClient.NewGrammarQuestionRedis(cache, log),
		search:                        searchClient.NewGrammarQuestionSearch(openSearch, log),
		completion:                    NewGrammarQuestionCompletionHelper(log),
		fillInBlankQuestionService:    fillInBlankQuestionService,
		fillInBlankAnswerService:      fillInBlankAnswerService,
		choiceOneQuestionService:      choiceOneQuestionService,
		choiceOneOptionService:        choiceOneOptionService,
		errorIdentificationService:    errorIdentificationService,
		sentenceTransformationService: sentenceTransformationService,
	}
}

func (u *GrammarQuestionUpdator) UpdateCacheAndSearch(ctx context.Context, question *grammar.GrammarQuestion) error {
	questionDetail, err := u.buildQuestionDetail(ctx, question)
	if err != nil {
		return fmt.Errorf("failed to build question detail: %w", err)
	}

	isComplete := u.completion.IsQuestionComplete(questionDetail)

	if err := u.redis.UpdateCachedGrammarQuestion(ctx, questionDetail, isComplete); err != nil {
		u.logger.Error("grammar_question_updator.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update cache")
	}

	status := map[bool]string{true: "complete", false: "uncomplete"}[isComplete]
	if err := u.search.UpsertGrammarQuestion(ctx, questionDetail, status); err != nil {
		u.logger.Error("grammar_question_updator.update_search", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update question in OpenSearch")
	}

	return nil
}

func (u *GrammarQuestionUpdator) buildQuestionDetail(ctx context.Context, question *grammar.GrammarQuestion) (*grammarDTO.GrammarQuestionDetail, error) {
	response := &grammarDTO.GrammarQuestionDetail{
		GrammarQuestionResponse: grammarDTO.GrammarQuestionResponse{
			ID:          question.ID,
			Type:        string(question.Type),
			Topic:       question.Topic,
			Instruction: question.Instruction,
			ImageURLs:   question.ImageURLs,
			MaxTime:     question.MaxTime,
			Version:     question.Version,
		},
	}

	var err error
	switch question.Type {
	case grammar.FillInTheBlank:
		err = u.loadFillInTheBlankData(ctx, question.ID, response)
	case grammar.ChoiceOne:
		err = u.loadChoiceOneData(ctx, question.ID, response)
	case grammar.ErrorIdentification:
		err = u.loadErrorIdentificationData(ctx, question.ID, response)
	case grammar.SentenceTransformation:
		err = u.loadSentenceTransformationData(ctx, question.ID, response)
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

func (u *GrammarQuestionUpdator) loadFillInTheBlankData(ctx context.Context, questionID uuid.UUID, response *grammarDTO.GrammarQuestionDetail) error {
	questions, err := u.fillInBlankQuestionService.GetQuestionsByGrammarQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get fill in blank questions: %w", err)
	}

	if len(questions) == 0 {
		return nil
	}

	question := questions[0]
	response.FillInTheBlankQuestion = &grammarDTO.GrammarFillInTheBlankQuestionResponse{
		ID:       question.ID,
		Question: question.Question,
	}

	answers, err := u.fillInBlankAnswerService.GetAnswersByGrammarFillInTheBlankQuestionID(ctx, question.ID)
	if err != nil {
		return fmt.Errorf("failed to get fill in blank answers: %w", err)
	}

	response.FillInTheBlankAnswers = make([]grammarDTO.GrammarFillInTheBlankAnswerResponse, len(answers))
	for i, answer := range answers {
		response.FillInTheBlankAnswers[i] = grammarDTO.GrammarFillInTheBlankAnswerResponse{
			ID:      answer.ID,
			Answer:  answer.Answer,
			Explain: answer.Explain,
		}
	}

	return nil
}

func (u *GrammarQuestionUpdator) loadChoiceOneData(ctx context.Context, questionID uuid.UUID, response *grammarDTO.GrammarQuestionDetail) error {
	questions, err := u.choiceOneQuestionService.GetQuestionsByGrammarQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get choice one questions: %w", err)
	}

	if len(questions) == 0 {
		return nil
	}

	question := questions[0]
	response.ChoiceOneQuestion = &grammarDTO.GrammarChoiceOneQuestionResponse{
		ID:       question.ID,
		Question: question.Question,
		Explain:  question.Explain,
	}

	options, err := u.choiceOneOptionService.GetOptionsByQuestionID(ctx, question.ID)
	if err != nil {
		return fmt.Errorf("failed to get choice one options: %w", err)
	}

	response.ChoiceOneOptions = make([]grammarDTO.GrammarChoiceOneOptionResponse, len(options))
	for i, option := range options {
		response.ChoiceOneOptions[i] = grammarDTO.GrammarChoiceOneOptionResponse{
			ID:        option.ID,
			Options:   option.Options,
			IsCorrect: option.IsCorrect,
		}
	}

	return nil
}

func (u *GrammarQuestionUpdator) loadErrorIdentificationData(ctx context.Context, questionID uuid.UUID, response *grammarDTO.GrammarQuestionDetail) error {
	identifications, err := u.errorIdentificationService.GetByGrammarQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get error identification: %w", err)
	}

	if len(identifications) == 0 {
		return nil
	}

	identification := identifications[0]
	response.ErrorIdentification = &grammarDTO.GrammarErrorIdentificationResponse{
		ID:            identification.ID,
		ErrorSentence: identification.ErrorSentence,
		ErrorWord:     identification.ErrorWord,
		CorrectWord:   identification.CorrectWord,
		Explain:       identification.Explain,
	}

	return nil
}

func (u *GrammarQuestionUpdator) loadSentenceTransformationData(ctx context.Context, questionID uuid.UUID, response *grammarDTO.GrammarQuestionDetail) error {
	transformations, err := u.sentenceTransformationService.GetByGrammarQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get sentence transformation: %w", err)
	}

	if len(transformations) == 0 {
		return nil
	}

	transformation := transformations[0]
	response.SentenceTransformation = &grammarDTO.GrammarSentenceTransformationResponse{
		ID:                     transformation.ID,
		OriginalSentence:       transformation.OriginalSentence,
		BeginningWord:          transformation.BeginningWord,
		ExampleCorrectSentence: transformation.ExampleCorrectSentence,
		Explain:                transformation.Explain,
	}

	return nil
}
