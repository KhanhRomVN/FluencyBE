package grammar

import (
	"context"
	"encoding/json"
	"errors"
	grammarDTO "fluencybe/internal/app/dto"
	grammarHelper "fluencybe/internal/app/helper/grammar"
	"fluencybe/internal/app/model/grammar"
	searchClient "fluencybe/internal/app/opensearch"
	redisClient "fluencybe/internal/app/redis"
	GrammarRepository "fluencybe/internal/app/repository/grammar"
	grammarValidator "fluencybe/internal/app/validator"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
)

var (
	ErrQuestionNotFound = errors.New("grammar question not found")
	ErrInvalidInput     = errors.New("invalid input")
)

type GrammarQuestionService struct {
	repo                          *GrammarRepository.GrammarQuestionRepository
	logger                        *logger.PrettyLogger
	redis                         *redisClient.GrammarQuestionRedis
	search                        *searchClient.GrammarQuestionSearch
	completion                    *grammarHelper.GrammarQuestionCompletionHelper
	updater                       *grammarHelper.GrammarQuestionFieldUpdater
	questionUpdator               *grammarHelper.GrammarQuestionUpdator
	fillInBlankQuestionService    *GrammarFillInTheBlankQuestionService
	fillInBlankAnswerService      *GrammarFillInTheBlankAnswerService
	choiceOneQuestionService      *GrammarChoiceOneQuestionService
	choiceOneOptionService        *GrammarChoiceOneOptionService
	errorIdentificationService    *GrammarErrorIdentificationService
	sentenceTransformationService *GrammarSentenceTransformationService
}

func NewGrammarQuestionService(
	repo *GrammarRepository.GrammarQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	openSearch *opensearch.Client,
	fillInBlankQuestionService *GrammarFillInTheBlankQuestionService,
	fillInBlankAnswerService *GrammarFillInTheBlankAnswerService,
	choiceOneQuestionService *GrammarChoiceOneQuestionService,
	choiceOneOptionService *GrammarChoiceOneOptionService,
	errorIdentificationService *GrammarErrorIdentificationService,
	sentenceTransformationService *GrammarSentenceTransformationService,
	questionUpdator *grammarHelper.GrammarQuestionUpdator,
) *GrammarQuestionService {
	return &GrammarQuestionService{
		repo:                          repo,
		logger:                        logger,
		redis:                         redisClient.NewGrammarQuestionRedis(cache, logger),
		search:                        searchClient.NewGrammarQuestionSearch(openSearch, logger),
		completion:                    grammarHelper.NewGrammarQuestionCompletionHelper(logger),
		updater:                       grammarHelper.NewGrammarQuestionFieldUpdater(logger),
		fillInBlankQuestionService:    fillInBlankQuestionService,
		fillInBlankAnswerService:      fillInBlankAnswerService,
		choiceOneQuestionService:      choiceOneQuestionService,
		choiceOneOptionService:        choiceOneOptionService,
		errorIdentificationService:    errorIdentificationService,
		sentenceTransformationService: sentenceTransformationService,
		questionUpdator:               questionUpdator,
	}
}

func (s *GrammarQuestionService) CreateQuestion(ctx context.Context, question *grammar.GrammarQuestion) error {
	if err := grammarValidator.ValidateGrammarQuestion(question); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Create in database
	if err := s.repo.CreateGrammarQuestion(ctx, question); err != nil {
		s.logger.Error("grammar_question_service.create", map[string]interface{}{
			"error":         err.Error(),
			"question_type": question.Type,
		}, "Failed to create grammar question")
		return err
	}

	questionDetail := grammarDTO.GrammarQuestionDetail{
		GrammarQuestionResponse: grammarDTO.GrammarQuestionResponse{
			ID:          question.ID,
			Type:        string(question.Type),
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
	s.redis.SetCacheGrammarQuestionDetail(ctx, &questionDetail, false)

	// Index to OpenSearch with correct status
	status := map[bool]string{true: "complete", false: "uncomplete"}[false]
	if err := s.search.UpsertGrammarQuestion(ctx, &questionDetail, status); err != nil {
		s.logger.Error("grammar_question_service.create.opensearch", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to index question in OpenSearch")
		return err
	}

	return nil
}

func (s *GrammarQuestionService) GetGrammarQuestionDetail(ctx context.Context, id uuid.UUID) (*grammarDTO.GrammarQuestionDetail, error) {
	start := time.Now()
	defer func() {
		s.logger.Info("grammar_question_service.get.timing", map[string]interface{}{
			"id":          id,
			"duration_ms": time.Since(start).Milliseconds(),
		}, "Question retrieval timing")
	}()

	// Try to get from cache with both complete and uncomplete status
	pattern := fmt.Sprintf("grammar_question:%s:*", id)
	keys, err := s.redis.GetCache().Keys(ctx, pattern)
	if err == nil && len(keys) > 0 {
		// Found in cache, try to get the data
		cachedData, err := s.redis.GetCache().Get(ctx, keys[0])
		if err == nil {
			var response grammarDTO.GrammarQuestionDetail
			if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
				return &response, nil
			}
		}
	}

	// If not in cache or error, get from database
	question, err := s.repo.GetGrammarQuestionByID(ctx, id)
	if err != nil {
		s.logger.Error("grammar_question_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get grammar question")
		return nil, err
	}

	response, err := s.getGrammarQuestionDetail(ctx, question)
	if err != nil {
		s.logger.Error("grammar_question_service.get.detail", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get question details")
		return nil, err
	}

	// Check completion status
	isComplete := s.completion.IsQuestionComplete(response)

	// Cache with status and version
	if err := s.redis.SetCacheGrammarQuestionDetail(ctx, response, isComplete); err != nil {
		s.logger.Error("grammar_question_service.get.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to cache question detail")
	}

	return response, nil
}

func (s *GrammarQuestionService) UpdateQuestion(ctx context.Context, id uuid.UUID, update grammarDTO.UpdateGrammarQuestionFieldRequest) error {
	// First get the base question from database
	baseQuestion, err := s.repo.GetGrammarQuestionByID(ctx, id)
	if err != nil {
		if errors.Is(err, GrammarRepository.ErrQuestionNotFound) {
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
	if err := s.repo.UpdateGrammarQuestion(ctx, baseQuestion); err != nil {
		s.logger.Error("grammar_question_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    baseQuestion.ID,
		}, "Failed to update question in database")
		return fmt.Errorf("failed to update question in database: %w", err)
	}

	// Use the updator to update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, baseQuestion); err != nil {
		s.logger.Error("grammar_question_service.update.cache_and_search", map[string]interface{}{
			"error": err.Error(),
			"id":    baseQuestion.ID,
		}, "Failed to update cache and search")
		// Continue even if cache/search update fails
	}

	return nil
}

func (s *GrammarQuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	// Delete from database
	if err := s.repo.DeleteGrammarQuestion(ctx, id); err != nil {
		s.logger.Error("grammar_question_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question")
		return err
	}

	// Invalidate cache
	s.redis.RemoveGrammarQuestionCacheEntries(ctx, id)

	// Delete from OpenSearch
	if err := s.search.DeleteGrammarQuestionFromIndex(ctx, id); err != nil {
		s.logger.Error("grammar_question_service.delete.opensearch", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question from OpenSearch")
		// Continue even if OpenSearch delete fails
	}

	return nil
}

func (s *GrammarQuestionService) GetNewUpdatedQuestions(ctx context.Context, versionChecks []struct {
	ID      uuid.UUID
	Version int
}) ([]*grammar.GrammarQuestion, error) {
	questionsToRetrieve := make(map[uuid.UUID]int)

	// Check both complete and uncomplete cache keys
	for _, check := range versionChecks {
		completeKey := fmt.Sprintf("grammar_question:%s:complete:%d", check.ID, check.Version)
		uncompleteKey := fmt.Sprintf("grammar_question:%s:uncomplete:%d", check.ID, check.Version)

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
		return []*grammar.GrammarQuestion{}, nil
	}

	// Build query conditions
	conditions := make([]string, 0, len(questionsToRetrieve))
	values := make([]interface{}, 0, len(questionsToRetrieve)*2)
	for id, version := range questionsToRetrieve {
		conditions = append(conditions, "(id = ? AND version > ?)")
		values = append(values, id, version)
	}

	var questions []*grammar.GrammarQuestion
	query := s.repo.GetDB().WithContext(ctx).
		Where(strings.Join(conditions, " OR "), values...).
		Order("created_at DESC")

	if err := query.Find(&questions).Error; err != nil {
		s.logger.Error("grammar_question_service.get_new_grammar_questions", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get updated questions")
		return nil, fmt.Errorf("failed to get updated questions: %w", err)
	}

	return questions, nil
}

func (s *GrammarQuestionService) getGrammarQuestionDetail(ctx context.Context, question *grammar.GrammarQuestion) (*grammarDTO.GrammarQuestionDetail, error) {
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
		err = s.loadFillInTheBlankData(ctx, question.ID, response)
	case grammar.ChoiceOne:
		err = s.loadChoiceOneData(ctx, question.ID, response)
	case grammar.ErrorIdentification:
		err = s.loadErrorIdentificationData(ctx, question.ID, response)
	case grammar.SentenceTransformation:
		err = s.loadSentenceTransformationData(ctx, question.ID, response)
	default:
		return nil, fmt.Errorf("unknown question type: %s", question.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load %s data: %w", question.Type, err)
	}

	return response, nil
}

func (s *GrammarQuestionService) loadFillInTheBlankData(ctx context.Context, questionID uuid.UUID, response *grammarDTO.GrammarQuestionDetail) error {
	questions, err := s.fillInBlankQuestionService.GetQuestionsByGrammarQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadFillInTheBlankData.get_questions", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get fill in blank questions")
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

	answers, err := s.fillInBlankAnswerService.GetAnswersByGrammarFillInTheBlankQuestionID(ctx, question.ID)
	if err != nil {
		s.logger.Error("loadFillInTheBlankData.get_answers", map[string]interface{}{
			"error":      err.Error(),
			"questionID": question.ID,
		}, "Failed to get fill in blank answers")
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

func (s *GrammarQuestionService) loadChoiceOneData(ctx context.Context, questionID uuid.UUID, response *grammarDTO.GrammarQuestionDetail) error {
	questions, err := s.choiceOneQuestionService.GetQuestionsByGrammarQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadChoiceOneData.get_questions", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get choice one questions")
		return fmt.Errorf("failed to get choice one questions: %w", err)
	}

	if len(questions) > 0 {
		question := questions[0]
		response.ChoiceOneQuestion = &grammarDTO.GrammarChoiceOneQuestionResponse{
			ID:       question.ID,
			Question: question.Question,
			Explain:  question.Explain,
		}

		options, err := s.choiceOneOptionService.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			s.logger.Error("loadChoiceOneData.get_options", map[string]interface{}{
				"error":      err.Error(),
				"questionID": question.ID,
			}, "Failed to get choice one options")
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
	}

	return nil
}

func (s *GrammarQuestionService) loadErrorIdentificationData(ctx context.Context, questionID uuid.UUID, response *grammarDTO.GrammarQuestionDetail) error {
	identifications, err := s.errorIdentificationService.GetByGrammarQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadErrorIdentificationData.get_identifications", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get error identifications")
		return fmt.Errorf("failed to get error identifications: %w", err)
	}

	if len(identifications) > 0 {
		identification := identifications[0]
		response.ErrorIdentification = &grammarDTO.GrammarErrorIdentificationResponse{
			ID:            identification.ID,
			ErrorSentence: identification.ErrorSentence,
			ErrorWord:     identification.ErrorWord,
			CorrectWord:   identification.CorrectWord,
			Explain:       identification.Explain,
		}
	}

	return nil
}

func (s *GrammarQuestionService) loadSentenceTransformationData(ctx context.Context, questionID uuid.UUID, response *grammarDTO.GrammarQuestionDetail) error {
	transformations, err := s.sentenceTransformationService.GetByGrammarQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadSentenceTransformationData.get_transformations", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get sentence transformations")
		return fmt.Errorf("failed to get sentence transformations: %w", err)
	}

	if len(transformations) > 0 {
		transformation := transformations[0]
		response.SentenceTransformation = &grammarDTO.GrammarSentenceTransformationResponse{
			ID:                     transformation.ID,
			OriginalSentence:       transformation.OriginalSentence,
			BeginningWord:          transformation.BeginningWord,
			ExampleCorrectSentence: transformation.ExampleCorrectSentence,
			Explain:                transformation.Explain,
		}
	}

	return nil
}

func (s *GrammarQuestionService) GetGrammarByListID(ctx context.Context, ids []uuid.UUID) ([]*grammarDTO.GrammarQuestionDetail, error) {
	// Track which IDs need to be fetched from database
	missingIDs := make([]uuid.UUID, 0)
	result := make([]*grammarDTO.GrammarQuestionDetail, 0, len(ids))

	// Try to get from cache first
	for _, id := range ids {
		// Try both complete and uncomplete patterns
		pattern := fmt.Sprintf("grammar_question:%s:*", id)
		keys, err := s.redis.GetCache().Keys(ctx, pattern)
		if err == nil && len(keys) > 0 {
			// Found in cache, try to get the data
			cachedData, err := s.redis.GetCache().Get(ctx, keys[0])
			if err == nil {
				var questionDetail grammarDTO.GrammarQuestionDetail
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

		var dbQuestions []*grammar.GrammarQuestion
		query := s.repo.GetDB().WithContext(ctx).
			Where(strings.Join(conditions, " OR "), values...).
			Find(&dbQuestions)

		if query.Error != nil {
			s.logger.Error("grammar_question_service.get_by_list_id", map[string]interface{}{
				"error": query.Error.Error(),
			}, "Failed to get questions from database")
			return nil, query.Error
		}

		// Process and cache the database results
		for _, q := range dbQuestions {
			// Build complete response with details
			questionDetail, err := s.getGrammarQuestionDetail(ctx, q)
			if err != nil {
				s.logger.Error("grammar_question_service.get_by_list_id.build_response", map[string]interface{}{
					"error": err.Error(),
					"id":    q.ID,
				}, "Failed to build complete response")
				continue
			}

			// Determine completion status
			isComplete := s.completion.IsQuestionComplete(questionDetail)
			status := "uncomplete"
			if isComplete {
				status = "complete"
			}

			// Cache with status and version
			cacheKey := fmt.Sprintf("grammar_question:%s:%s:%d", q.ID, status, q.Version)
			if questionJSON, err := json.Marshal(questionDetail); err == nil {
				if err := s.redis.GetCache().Set(ctx, cacheKey, string(questionJSON), 24*time.Hour); err != nil {
					s.logger.Error("grammar_question_service.get_by_list_id.cache", map[string]interface{}{
						"error": err.Error(),
						"id":    q.ID,
					}, "Failed to cache question")
				}
			}

			result = append(result, questionDetail)
		}
	}

	return result, nil
}

func (s *GrammarQuestionService) SearchQuestionsWithFilter(ctx context.Context, filter grammarDTO.GrammarQuestionSearchFilter) (*grammarDTO.ListGrammarQuestionsPagination, error) {
	// Add debug logging
	s.logger.Debug("grammar_question_service.search.start", map[string]interface{}{
		"filter": filter,
	}, "Starting question search")

	result, err := s.search.SearchQuestions(ctx, filter)
	if err != nil {
		s.logger.Error("grammar_question_service.search", map[string]interface{}{
			"error":  err.Error(),
			"filter": filter,
		}, "Failed to search questions")
		return nil, fmt.Errorf("failed to search questions: %w", err)
	}

	// Add debug logging for results
	s.logger.Debug("grammar_question_service.search.complete", map[string]interface{}{
		"total_results": result.Total,
		"page":          result.Page,
		"page_size":     result.PageSize,
		"num_results":   len(result.Questions),
	}, "Search completed")

	// Return the paginated results
	return &grammarDTO.ListGrammarQuestionsPagination{
		Questions: result.Questions,
		Total:     result.Total,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
	}, nil
}

func (s *GrammarQuestionService) DeleteAllQuestions(ctx context.Context) error {
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

	// Delete all records from grammar_questions
	if err := tx.Exec("DELETE FROM grammar_questions").Error; err != nil {
		tx.Rollback()
		s.logger.Error("grammar_question_service.delete_all", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete all grammar questions")
		return err
	}

	// Delete all Redis cache with pattern grammar_question:*
	if err := s.redis.GetCache().DeletePattern(ctx, "grammar_question:*"); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_question_service.delete_all.cache", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete Redis cache")
		return err
	}

	// Delete OpenSearch index
	if err := s.search.RemoveGrammarQuestionsIndex(ctx); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_question_service.delete_all.search", map[string]interface{}{
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
