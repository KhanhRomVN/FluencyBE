package writing

import (
	"context"
	writingDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/writing"
	searchClient "fluencybe/internal/app/opensearch"
	redisClient "fluencybe/internal/app/redis"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
)

type WritingQuestionUpdator struct {
	logger                    *logger.PrettyLogger
	redis                     *redisClient.WritingQuestionRedis
	search                    *searchClient.WritingQuestionSearch
	completion                *WritingQuestionCompletionHelper
	sentenceCompletionService SentenceCompletionService
	essayService              EssayService
}

func NewWritingQuestionUpdator(
	log *logger.PrettyLogger,
	cache cache.Cache,
	openSearch *opensearch.Client,
	sentenceCompletionService SentenceCompletionService,
	essayService EssayService,
) *WritingQuestionUpdator {
	return &WritingQuestionUpdator{
		logger:                    log,
		redis:                     redisClient.NewWritingQuestionRedis(cache, log),
		search:                    searchClient.NewWritingQuestionSearch(openSearch, log),
		completion:                NewWritingQuestionCompletionHelper(log),
		sentenceCompletionService: sentenceCompletionService,
		essayService:              essayService,
	}
}

func (u *WritingQuestionUpdator) UpdateCacheAndSearch(ctx context.Context, question *writing.WritingQuestion) error {
	questionDetail, err := u.buildQuestionDetail(ctx, question)
	if err != nil {
		return fmt.Errorf("failed to build question detail: %w", err)
	}

	isComplete := u.completion.IsQuestionComplete(questionDetail)

	if err := u.redis.UpdateCachedWritingQuestion(ctx, questionDetail, isComplete); err != nil {
		u.logger.Error("writing_question_updator.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update cache")
	}

	status := map[bool]string{true: "complete", false: "uncomplete"}[isComplete]
	if err := u.search.UpsertWritingQuestion(ctx, questionDetail, status); err != nil {
		u.logger.Error("writing_question_updator.update_search", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update question in OpenSearch")
	}

	return nil
}

func (u *WritingQuestionUpdator) buildQuestionDetail(ctx context.Context, question *writing.WritingQuestion) (*writingDTO.WritingQuestionDetail, error) {
	response := &writingDTO.WritingQuestionDetail{
		WritingQuestionResponse: writingDTO.WritingQuestionResponse{
			ID:          question.ID,
			Type:        question.Type,
			Topic:       question.Topic,
			Instruction: question.Instruction,
			ImageURLs:   question.ImageURLs,
			MaxTime:     question.MaxTime,
			Version:     question.Version,
		},
	}

	var err error
	switch question.Type {
	case "SENTENCE_COMPLETION":
		err = u.loadSentenceCompletionData(ctx, question.ID, response)
	case "ESSAY":
		err = u.loadEssayData(ctx, question.ID, response)
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

func (u *WritingQuestionUpdator) loadSentenceCompletionData(ctx context.Context, questionID uuid.UUID, response *writingDTO.WritingQuestionDetail) error {
	sentences, err := u.sentenceCompletionService.GetByWritingQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get sentence completions: %w", err)
	}

	response.SentenceCompletion = make([]writingDTO.WritingSentenceCompletionResponse, len(sentences))
	for i, sentence := range sentences {
		response.SentenceCompletion[i] = writingDTO.WritingSentenceCompletionResponse{
			ID:                sentence.ID,
			ExampleSentence:   sentence.ExampleSentence,
			GivenPartSentence: sentence.GivenPartSentence,
			Position:          sentence.Position,
			RequiredWords:     sentence.RequiredWords,
			Explain:           sentence.Explain,
			MinWords:          sentence.MinWords,
			MaxWords:          sentence.MaxWords,
		}
	}
	return nil
}

func (u *WritingQuestionUpdator) loadEssayData(ctx context.Context, questionID uuid.UUID, response *writingDTO.WritingQuestionDetail) error {
	essays, err := u.essayService.GetByWritingQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get essays: %w", err)
	}

	response.Essay = make([]writingDTO.WritingEssayResponse, len(essays))
	for i, essay := range essays {
		response.Essay[i] = writingDTO.WritingEssayResponse{
			ID:             essay.ID,
			EssayType:      essay.EssayType,
			RequiredPoints: essay.RequiredPoints,
			MinWords:       essay.MinWords,
			MaxWords:       essay.MaxWords,
			SampleEssay:    essay.SampleEssay,
			Explain:        essay.Explain,
		}
	}
	return nil
}
