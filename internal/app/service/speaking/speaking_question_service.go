package speaking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	speakingDTO "fluencybe/internal/app/dto"
	speakingHelper "fluencybe/internal/app/helper/speaking"
	"fluencybe/internal/app/model/speaking"
	searchClient "fluencybe/internal/app/opensearch"
	redisClient "fluencybe/internal/app/redis"
	speakingRepository "fluencybe/internal/app/repository/speaking"
	speakingValidator "fluencybe/internal/app/validator"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"

	"github.com/google/uuid"
)

var (
	ErrQuestionNotFound = errors.New("speaking question not found")
	ErrInvalidInput     = errors.New("invalid input")
)

type SpeakingQuestionService struct {
	repo                              *speakingRepository.SpeakingQuestionRepository
	logger                            *logger.PrettyLogger
	redis                             *redisClient.SpeakingQuestionRedis
	search                            *searchClient.SpeakingQuestionSearch
	completion                        *speakingHelper.SpeakingQuestionCompletionHelper
	updater                           *speakingHelper.SpeakingQuestionFieldUpdater
	questionUpdator                   *speakingHelper.SpeakingQuestionUpdator
	wordRepetitionService             *SpeakingWordRepetitionService
	phraseRepetitionService           *SpeakingPhraseRepetitionService
	paragraphRepetitionService        *SpeakingParagraphRepetitionService
	openParagraphService              *SpeakingOpenParagraphService
	conversationalRepetitionService   *SpeakingConversationalRepetitionService
	conversationalRepetitionQAService *SpeakingConversationalRepetitionQAService
	conversationalOpenService         *SpeakingConversationalOpenService
}

func NewSpeakingQuestionService(
	repo *speakingRepository.SpeakingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	search *searchClient.SpeakingQuestionSearch,
	wordRepetitionService *SpeakingWordRepetitionService,
	phraseRepetitionService *SpeakingPhraseRepetitionService,
	paragraphRepetitionService *SpeakingParagraphRepetitionService,
	openParagraphService *SpeakingOpenParagraphService,
	conversationalRepetitionService *SpeakingConversationalRepetitionService,
	conversationalRepetitionQAService *SpeakingConversationalRepetitionQAService,
	conversationalOpenService *SpeakingConversationalOpenService,
	questionUpdator *speakingHelper.SpeakingQuestionUpdator,
) *SpeakingQuestionService {
	return &SpeakingQuestionService{
		repo:                              repo,
		logger:                            logger,
		redis:                             redisClient.NewSpeakingQuestionRedis(cache, logger),
		search:                            search,
		completion:                        speakingHelper.NewSpeakingQuestionCompletionHelper(logger),
		updater:                           speakingHelper.NewSpeakingQuestionFieldUpdater(logger),
		wordRepetitionService:             wordRepetitionService,
		phraseRepetitionService:           phraseRepetitionService,
		paragraphRepetitionService:        paragraphRepetitionService,
		openParagraphService:              openParagraphService,
		conversationalRepetitionService:   conversationalRepetitionService,
		conversationalRepetitionQAService: conversationalRepetitionQAService,
		conversationalOpenService:         conversationalOpenService,
		questionUpdator:                   questionUpdator,
	}
}

func (s *SpeakingQuestionService) CreateQuestion(ctx context.Context, question *speaking.SpeakingQuestion) error {
	if err := speakingValidator.ValidateSpeakingQuestion(question); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if err := s.repo.CreateSpeakingQuestion(ctx, question); err != nil {
		s.logger.Error("speaking_question_service.create", map[string]interface{}{
			"error":         err.Error(),
			"question_type": question.Type,
		}, "Failed to create speaking question")
		return err
	}

	questionDetail := speakingDTO.SpeakingQuestionDetail{
		SpeakingQuestionResponse: speakingDTO.SpeakingQuestionResponse{
			ID:          question.ID,
			Type:        question.Type,
			Topic:       question.Topic,
			Instruction: question.Instruction,
			ImageURLs:   question.ImageURLs,
			MaxTime:     question.MaxTime,
			Version:     question.Version,
			CreatedAt:   question.CreatedAt,
			UpdatedAt:   question.UpdatedAt,
		},
	}

	// Cache the question with correct completion status
	s.redis.SetCacheSpeakingQuestionDetail(ctx, &questionDetail, false)

	// Index to OpenSearch with correct status
	status := map[bool]string{true: "complete", false: "uncomplete"}[false]
	if err := s.search.UpsertSpeakingQuestion(ctx, &questionDetail, status); err != nil {
		s.logger.Error("speaking_question_service.create.opensearch", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to index question in OpenSearch")
	}

	return nil
}

func (s *SpeakingQuestionService) DeleteAllQuestions(ctx context.Context) error {
	// Start transaction
	tx := s.repo.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete all records from speaking_questions
	if err := tx.Exec("DELETE FROM speaking_questions").Error; err != nil {
		tx.Rollback()
		s.logger.Error("speaking_question_service.delete_all", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete all speaking questions")
		return err
	}

	// Delete all Redis cache with pattern speaking_question:*
	if err := s.redis.GetCache().DeletePattern(ctx, "speaking_question:*"); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_question_service.delete_all.cache", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete Redis cache")
		return err
	}

	// Delete OpenSearch index
	if err := s.search.RemoveSpeakingQuestionsIndex(ctx); err != nil {
		tx.Rollback()
		s.logger.Error("speaking_question_service.delete_all.search", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete OpenSearch index")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SpeakingQuestionService) SearchQuestionsWithFilter(ctx context.Context, filter speakingDTO.SpeakingQuestionSearchFilter) (*speakingDTO.ListSpeakingQuestionsPagination, error) {
	// Add debug logging
	s.logger.Debug("speaking_question_service.search.start", map[string]interface{}{
		"filter": filter,
	}, "Starting question search")

	result, err := s.search.SearchQuestions(ctx, filter)
	if err != nil {
		s.logger.Error("speaking_question_service.search", map[string]interface{}{
			"error":  err.Error(),
			"filter": filter,
		}, "Failed to search questions")
		return nil, fmt.Errorf("failed to search questions: %w", err)
	}

	// Add debug logging for results
	s.logger.Debug("speaking_question_service.search.complete", map[string]interface{}{
		"total_results": result.Total,
		"page":          result.Page,
		"page_size":     result.PageSize,
	}, "Search completed")

	return result, nil
}

func (s *SpeakingQuestionService) GetNewUpdatedQuestions(ctx context.Context, versionChecks []struct {
	ID      uuid.UUID
	Version int
}) ([]*speaking.SpeakingQuestion, error) {
	questionsToRetrieve := make(map[uuid.UUID]int)

	// Check both complete and uncomplete cache keys
	for _, check := range versionChecks {
		completeKey := fmt.Sprintf("speaking_question:%s:complete:%d", check.ID, check.Version)
		uncompleteKey := fmt.Sprintf("speaking_question:%s:uncomplete:%d", check.ID, check.Version)

		// Try complete key first
		if _, err := s.redis.GetCache().Get(ctx, completeKey); err != nil {
			// If not found, try uncomplete key
			if _, err := s.redis.GetCache().Get(ctx, uncompleteKey); err != nil {
				// If not found in either cache, add to retrieval map
				questionsToRetrieve[check.ID] = check.Version
				continue
			}
		}
	}

	if len(questionsToRetrieve) == 0 {
		return []*speaking.SpeakingQuestion{}, nil
	}

	// Build query conditions
	conditions := make([]string, 0, len(questionsToRetrieve))
	values := make([]interface{}, 0, len(questionsToRetrieve)*2)
	for id, version := range questionsToRetrieve {
		conditions = append(conditions, "(id = ? AND version > ?)")
		values = append(values, id, version)
	}

	var questions []*speaking.SpeakingQuestion
	query := s.repo.GetDB().WithContext(ctx).
		Where(strings.Join(conditions, " OR "), values...).
		Order("created_at DESC")

	if err := query.Find(&questions).Error; err != nil {
		s.logger.Error("speaking_question_service.get_new_speaking_questions", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get updated questions")
		return nil, fmt.Errorf("failed to get updated questions: %w", err)
	}

	return questions, nil
}

func (s *SpeakingQuestionService) GetSpeakingQuestionDetail(ctx context.Context, id uuid.UUID) (*speakingDTO.SpeakingQuestionDetail, error) {
	question, err := s.repo.GetSpeakingQuestionByID(ctx, id)
	if err != nil {
		s.logger.Error("speaking_question_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get speaking question")
		return nil, err
	}

	response, err := s.buildQuestionDetail(ctx, question)
	if err != nil {
		s.logger.Error("speaking_question_service.get.detail", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get question details")
		return nil, err
	}

	// Check completion status
	isComplete := s.completion.IsQuestionComplete(response)

	// Cache with status and version
	if err := s.redis.SetCacheSpeakingQuestionDetail(ctx, response, isComplete); err != nil {
		s.logger.Error("speaking_question_service.get.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to cache question detail")
	}

	return response, nil
}

func (s *SpeakingQuestionService) buildQuestionDetail(ctx context.Context, question *speaking.SpeakingQuestion) (*speakingDTO.SpeakingQuestionDetail, error) {
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
		err = s.loadWordRepetitionData(ctx, question.ID, response)
	case "PHRASE_REPETITION":
		err = s.loadPhraseRepetitionData(ctx, question.ID, response)
	case "PARAGRAPH_REPETITION":
		err = s.loadParagraphRepetitionData(ctx, question.ID, response)
	case "OPEN_PARAGRAPH":
		err = s.loadOpenParagraphData(ctx, question.ID, response)
	case "CONVERSATIONAL_REPETITION":
		err = s.loadConversationalRepetitionData(ctx, question.ID, response)
	case "CONVERSATIONAL_OPEN":
		err = s.loadConversationalOpenData(ctx, question.ID, response)
	default:
		return nil, fmt.Errorf("unknown question type: %s", question.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load %s data: %w", question.Type, err)
	}

	return response, nil
}

func (s *SpeakingQuestionService) loadWordRepetitionData(ctx context.Context, questionID uuid.UUID, response *speakingDTO.SpeakingQuestionDetail) error {
	words, err := s.wordRepetitionService.GetBySpeakingQuestionID(ctx, questionID)
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

func (s *SpeakingQuestionService) loadPhraseRepetitionData(ctx context.Context, questionID uuid.UUID, response *speakingDTO.SpeakingQuestionDetail) error {
	phrases, err := s.phraseRepetitionService.GetBySpeakingQuestionID(ctx, questionID)
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

func (s *SpeakingQuestionService) loadParagraphRepetitionData(ctx context.Context, questionID uuid.UUID, response *speakingDTO.SpeakingQuestionDetail) error {
	paragraphs, err := s.paragraphRepetitionService.GetBySpeakingQuestionID(ctx, questionID)
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

func (s *SpeakingQuestionService) loadOpenParagraphData(ctx context.Context, questionID uuid.UUID, response *speakingDTO.SpeakingQuestionDetail) error {
	paragraphs, err := s.openParagraphService.GetBySpeakingQuestionID(ctx, questionID)
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

func (s *SpeakingQuestionService) loadConversationalRepetitionData(ctx context.Context, questionID uuid.UUID, response *speakingDTO.SpeakingQuestionDetail) error {
	conversations, err := s.conversationalRepetitionService.GetBySpeakingQuestionID(ctx, questionID)
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

		qas, err := s.conversationalRepetitionQAService.GetBySpeakingConversationalRepetitionID(ctx, conversation.ID)
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

func (s *SpeakingQuestionService) loadConversationalOpenData(ctx context.Context, questionID uuid.UUID, response *speakingDTO.SpeakingQuestionDetail) error {
	conversations, err := s.conversationalOpenService.GetBySpeakingQuestionID(ctx, questionID)
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

func (s *SpeakingQuestionService) UpdateQuestion(ctx context.Context, id uuid.UUID, update speakingDTO.UpdateSpeakingQuestionFieldRequest) error {
	// First get the base question from database
	baseQuestion, err := s.repo.GetSpeakingQuestionByID(ctx, id)
	if err != nil {
		if errors.Is(err, speakingRepository.ErrQuestionNotFound) {
			return ErrQuestionNotFound
		}
		return fmt.Errorf("failed to get question: %w", err)
	}

	// Update the base question fields
	if err := s.updater.UpdateField(baseQuestion, update); err != nil {
		return fmt.Errorf("failed to update field: %w", err)
	}

	// Update timestamps and version
	baseQuestion.UpdatedAt = time.Now()
	baseQuestion.Version++ // Increment version on update

	// Update in database
	if err := s.repo.UpdateSpeakingQuestion(ctx, baseQuestion); err != nil {
		s.logger.Error("speaking_question_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    baseQuestion.ID,
		}, "Failed to update question in database")
		return fmt.Errorf("failed to update question in database: %w", err)
	}

	// Use the updator to update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, baseQuestion); err != nil {
		s.logger.Error("speaking_question_service.update.cache_and_search", map[string]interface{}{
			"error": err.Error(),
			"id":    baseQuestion.ID,
		}, "Failed to update cache and search")
		// Continue even if cache/search update fails
	}

	return nil
}

func (s *SpeakingQuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteSpeakingQuestion(ctx, id); err != nil {
		s.logger.Error("speaking_question_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete speaking question")
		return err
	}

	// Invalidate cache
	s.redis.RemoveSpeakingQuestionCacheEntries(ctx, id)

	// Delete from OpenSearch
	if err := s.search.DeleteSpeakingQuestionFromIndex(ctx, id); err != nil {
		s.logger.Error("speaking_question_service.delete.opensearch", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question from OpenSearch")
		// Continue even if OpenSearch delete fails
	}

	return nil
}

func (s *SpeakingQuestionService) GetSpeakingByListID(ctx context.Context, ids []uuid.UUID) ([]*speakingDTO.SpeakingQuestionDetail, error) {
	// Track which IDs need to be fetched from database
	missingIDs := make([]uuid.UUID, 0)
	result := make([]*speakingDTO.SpeakingQuestionDetail, 0, len(ids))

	// Try to get from cache first
	for _, id := range ids {
		// Try both complete and uncomplete patterns
		pattern := fmt.Sprintf("speaking_question:%s:*", id)
		keys, err := s.redis.GetCache().Keys(ctx, pattern)
		if err == nil && len(keys) > 0 {
			// Found in cache, try to get the data
			cachedData, err := s.redis.GetCache().Get(ctx, keys[0])
			if err == nil {
				var questionDetail speakingDTO.SpeakingQuestionDetail
				if err := json.Unmarshal([]byte(cachedData), &questionDetail); err == nil {
					result = append(result, &questionDetail)
					continue
				}
			}
		}
		// Not found in cache or error occurred, add to missing IDs
		missingIDs = append(missingIDs, id)
	}

	// If we have missing IDs, fetch them from database
	if len(missingIDs) > 0 {
		conditions := make([]string, len(missingIDs))
		values := make([]interface{}, len(missingIDs))
		for i, id := range missingIDs {
			conditions[i] = "id = ?"
			values[i] = id
		}

		var dbQuestions []*speaking.SpeakingQuestion
		query := s.repo.GetDB().WithContext(ctx).
			Where(strings.Join(conditions, " OR "), values...).
			Find(&dbQuestions)

		if query.Error != nil {
			s.logger.Error("speaking_question_service.get_by_list_id", map[string]interface{}{
				"error": query.Error.Error(),
			}, "Failed to get questions from database")
			return nil, query.Error
		}

		// Process and cache the database results
		for _, q := range dbQuestions {
			// Build complete response with details
			questionDetail, err := s.buildQuestionDetail(ctx, q)
			if err != nil {
				s.logger.Error("speaking_question_service.get_by_list_id.build_response", map[string]interface{}{
					"error": err.Error(),
					"id":    q.ID,
				}, "Failed to build complete response")
				continue
			}

			// Determine completion status
			isComplete := s.completion.IsQuestionComplete(questionDetail)

			// Cache with status and version
			if err := s.redis.SetCacheSpeakingQuestionDetail(ctx, questionDetail, isComplete); err != nil {
				s.logger.Error("speaking_question_service.get_by_list_id.cache", map[string]interface{}{
					"error": err.Error(),
					"id":    q.ID,
				}, "Failed to cache question")
			}

			result = append(result, questionDetail)
		}
	}

	return result, nil
}
