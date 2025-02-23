package speaking

import (
	"context"
	speakingDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/speaking"
	searchClient "fluencybe/internal/app/opensearch"
	redisClient "fluencybe/internal/app/redis"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
)

type SpeakingQuestionUpdator struct {
	logger                            *logger.PrettyLogger
	redis                             *redisClient.SpeakingQuestionRedis
	search                            *searchClient.SpeakingQuestionSearch
	completion                        *SpeakingQuestionCompletionHelper
	wordRepetitionService             WordRepetitionService
	phraseRepetitionService           PhraseRepetitionService
	paragraphRepetitionService        ParagraphRepetitionService
	openParagraphService              OpenParagraphService
	conversationalRepetitionService   ConversationalRepetitionService
	conversationalRepetitionQAService ConversationalRepetitionQAService
	conversationalOpenService         ConversationalOpenService
}

func NewSpeakingQuestionUpdator(
	log *logger.PrettyLogger,
	cache cache.Cache,
	openSearch *opensearch.Client,
	wordRepetitionService WordRepetitionService,
	phraseRepetitionService PhraseRepetitionService,
	paragraphRepetitionService ParagraphRepetitionService,
	openParagraphService OpenParagraphService,
	conversationalRepetitionService ConversationalRepetitionService,
	conversationalRepetitionQAService ConversationalRepetitionQAService,
	conversationalOpenService ConversationalOpenService,
) *SpeakingQuestionUpdator {
	return &SpeakingQuestionUpdator{
		logger:                            log,
		redis:                             redisClient.NewSpeakingQuestionRedis(cache, log),
		search:                            searchClient.NewSpeakingQuestionSearch(openSearch, log),
		completion:                        NewSpeakingQuestionCompletionHelper(log),
		wordRepetitionService:             wordRepetitionService,
		phraseRepetitionService:           phraseRepetitionService,
		paragraphRepetitionService:        paragraphRepetitionService,
		openParagraphService:              openParagraphService,
		conversationalRepetitionService:   conversationalRepetitionService,
		conversationalRepetitionQAService: conversationalRepetitionQAService,
		conversationalOpenService:         conversationalOpenService,
	}
}

func (u *SpeakingQuestionUpdator) UpdateCacheAndSearch(ctx context.Context, question *speaking.SpeakingQuestion) error {
	questionDetail, err := u.buildQuestionDetail(ctx, question)
	if err != nil {
		return fmt.Errorf("failed to build question detail: %w", err)
	}

	isComplete := u.completion.IsQuestionComplete(questionDetail)

	if err := u.redis.UpdateCachedSpeakingQuestion(ctx, questionDetail, isComplete); err != nil {
		u.logger.Error("speaking_question_updator.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update cache")
	}

	status := map[bool]string{true: "complete", false: "uncomplete"}[isComplete]
	if err := u.search.UpsertSpeakingQuestion(ctx, questionDetail, status); err != nil {
		u.logger.Error("speaking_question_updator.update_search", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update question in OpenSearch")
	}

	return nil
}

func (u *SpeakingQuestionUpdator) buildQuestionDetail(ctx context.Context, question *speaking.SpeakingQuestion) (*speakingDTO.SpeakingQuestionDetail, error) {
	response := &speakingDTO.SpeakingQuestionDetail{
		SpeakingQuestionResponse: speakingDTO.SpeakingQuestionResponse{
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
	case "WORD_REPETITION":
		err = u.loadWordRepetitionData(ctx, question.ID, response)
	case "PHRASE_REPETITION":
		err = u.loadPhraseRepetitionData(ctx, question.ID, response)
	case "PARAGRAPH_REPETITION":
		err = u.loadParagraphRepetitionData(ctx, question.ID, response)
	case "OPEN_PARAGRAPH":
		err = u.loadOpenParagraphData(ctx, question.ID, response)
	case "CONVERSATIONAL_REPETITION":
		err = u.loadConversationalRepetitionData(ctx, question.ID, response)
	case "CONVERSATIONAL_OPEN":
		err = u.loadConversationalOpenData(ctx, question.ID, response)
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

func (u *SpeakingQuestionUpdator) loadWordRepetitionData(ctx context.Context, questionID uuid.UUID, response *speakingDTO.SpeakingQuestionDetail) error {
	words, err := u.wordRepetitionService.GetBySpeakingQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get word repetitions: %w", err)
	}

	response.WordRepetition = make([]speakingDTO.SpeakingWordRepetitionResponse, len(words))
	for i, word := range words {
		response.WordRepetition[i] = speakingDTO.SpeakingWordRepetitionResponse{
			ID:   word.ID,
			Word: word.Word,
			Mean: word.Mean,
		}
	}
	return nil
}

func (u *SpeakingQuestionUpdator) loadPhraseRepetitionData(ctx context.Context, questionID uuid.UUID, response *speakingDTO.SpeakingQuestionDetail) error {
	phrases, err := u.phraseRepetitionService.GetBySpeakingQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get phrase repetitions: %w", err)
	}

	response.PhraseRepetition = make([]speakingDTO.SpeakingPhraseRepetitionResponse, len(phrases))
	for i, phrase := range phrases {
		response.PhraseRepetition[i] = speakingDTO.SpeakingPhraseRepetitionResponse{
			ID:     phrase.ID,
			Phrase: phrase.Phrase,
			Mean:   phrase.Mean,
		}
	}
	return nil
}

func (u *SpeakingQuestionUpdator) loadParagraphRepetitionData(ctx context.Context, questionID uuid.UUID, response *speakingDTO.SpeakingQuestionDetail) error {
	paragraphs, err := u.paragraphRepetitionService.GetBySpeakingQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get paragraph repetitions: %w", err)
	}

	response.ParagraphRepetition = make([]speakingDTO.SpeakingParagraphRepetitionResponse, len(paragraphs))
	for i, paragraph := range paragraphs {
		response.ParagraphRepetition[i] = speakingDTO.SpeakingParagraphRepetitionResponse{
			ID:        paragraph.ID,
			Paragraph: paragraph.Paragraph,
			Mean:      paragraph.Mean,
		}
	}
	return nil
}

func (u *SpeakingQuestionUpdator) loadOpenParagraphData(ctx context.Context, questionID uuid.UUID, response *speakingDTO.SpeakingQuestionDetail) error {
	paragraphs, err := u.openParagraphService.GetBySpeakingQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get open paragraphs: %w", err)
	}

	response.OpenParagraph = make([]speakingDTO.SpeakingOpenParagraphResponse, len(paragraphs))
	for i, paragraph := range paragraphs {
		response.OpenParagraph[i] = speakingDTO.SpeakingOpenParagraphResponse{
			ID:                   paragraph.ID,
			Question:             paragraph.Question,
			ExamplePassage:       paragraph.ExamplePassage,
			MeanOfExamplePassage: paragraph.MeanOfExamplePassage,
		}
	}
	return nil
}

func (u *SpeakingQuestionUpdator) loadConversationalRepetitionData(ctx context.Context, questionID uuid.UUID, response *speakingDTO.SpeakingQuestionDetail) error {
	conversations, err := u.conversationalRepetitionService.GetBySpeakingQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get conversational repetitions: %w", err)
	}

	if len(conversations) > 0 {
		conversation := conversations[0]
		response.ConversationalRepetition = &speakingDTO.SpeakingConversationalRepetitionResponse{
			ID:       conversation.ID,
			Title:    conversation.Title,
			Overview: conversation.Overview,
		}

		qas, err := u.conversationalRepetitionQAService.GetBySpeakingConversationalRepetitionID(ctx, conversation.ID)
		if err != nil {
			return fmt.Errorf("failed to get conversational repetition QAs: %w", err)
		}

		response.ConversationalRepetitionQAs = make([]speakingDTO.SpeakingConversationalRepetitionQAResponse, len(qas))
		for i, qa := range qas {
			response.ConversationalRepetitionQAs[i] = speakingDTO.SpeakingConversationalRepetitionQAResponse{
				ID:             qa.ID,
				Question:       qa.Question,
				Answer:         qa.Answer,
				MeanOfQuestion: qa.MeanOfQuestion,
				MeanOfAnswer:   qa.MeanOfAnswer,
				Explain:        qa.Explain,
			}
		}
	}
	return nil
}

func (u *SpeakingQuestionUpdator) loadConversationalOpenData(ctx context.Context, questionID uuid.UUID, response *speakingDTO.SpeakingQuestionDetail) error {
	conversations, err := u.conversationalOpenService.GetBySpeakingQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("failed to get conversational opens: %w", err)
	}

	if len(conversations) > 0 {
		conversation := conversations[0]
		response.ConversationalOpen = &speakingDTO.SpeakingConversationalOpenResponse{
			ID:                  conversation.ID,
			Title:               conversation.Title,
			Overview:            conversation.Overview,
			ExampleConversation: conversation.ExampleConversation,
		}
	}
	return nil
}
