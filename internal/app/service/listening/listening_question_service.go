package listening

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	listeningDTO "fluencybe/internal/app/dto"
	listeningHelper "fluencybe/internal/app/helper/listening"
	"fluencybe/internal/app/model/listening"
	searchClient "fluencybe/internal/app/opensearch"
	redisClient "fluencybe/internal/app/redis"
	ListeningRepository "fluencybe/internal/app/repository/listening"
	listeningValidator "fluencybe/internal/app/validator"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

var (
	ErrQuestionNotFound = errors.New("listening question not found")
	ErrInvalidInput     = errors.New("invalid input")
)

type ListeningQuestionService struct {
	repo                              *ListeningRepository.ListeningQuestionRepository
	logger                            *logger.PrettyLogger
	redis                             *redisClient.ListeningQuestionRedis
	search                            *searchClient.ListeningQuestionSearch
	completion                        *listeningHelper.ListeningQuestionCompletionHelper
	updater                           *listeningHelper.ListeningQuestionFieldUpdater
	questionUpdator                   *listeningHelper.ListeningQuestionUpdator
	fillInBlankQuestionService        *ListeningFillInTheBlankQuestionService
	fillInBlankAnswerService          *ListeningFillInTheBlankAnswerService
	choiceOneQuestionService          *ListeningChoiceOneQuestionService
	choiceOneOptionService            *ListeningChoiceOneOptionService
	choiceMultiQuestionService        *ListeningChoiceMultiQuestionService
	choiceMultiOptionService          *ListeningChoiceMultiOptionService
	mapLabellingQuestionAnswerService *ListeningMapLabellingService
	matchingQuestionAnswerService     *ListeningMatchingService
}

func NewListeningQuestionService(
	repo *ListeningRepository.ListeningQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	openSearch *opensearch.Client,
	fillInBlankQuestionService *ListeningFillInTheBlankQuestionService,
	fillInBlankAnswerService *ListeningFillInTheBlankAnswerService,
	choiceOneQuestionService *ListeningChoiceOneQuestionService,
	choiceOneOptionService *ListeningChoiceOneOptionService,
	choiceMultiQuestionService *ListeningChoiceMultiQuestionService,
	choiceMultiOptionService *ListeningChoiceMultiOptionService,
	mapLabellingQuestionAnswerService *ListeningMapLabellingService,
	matchingQuestionAnswerService *ListeningMatchingService,
	questionUpdator *listeningHelper.ListeningQuestionUpdator,

) *ListeningQuestionService {
	return &ListeningQuestionService{
		repo:                              repo,
		logger:                            logger,
		redis:                             redisClient.NewListeningQuestionRedis(cache, logger),
		search:                            searchClient.NewListeningQuestionSearch(openSearch, logger),
		completion:                        listeningHelper.NewListeningQuestionCompletionHelper(logger),
		updater:                           listeningHelper.NewListeningQuestionFieldUpdater(logger),
		fillInBlankQuestionService:        fillInBlankQuestionService,
		fillInBlankAnswerService:          fillInBlankAnswerService,
		choiceOneQuestionService:          choiceOneQuestionService,
		choiceOneOptionService:            choiceOneOptionService,
		choiceMultiQuestionService:        choiceMultiQuestionService,
		choiceMultiOptionService:          choiceMultiOptionService,
		mapLabellingQuestionAnswerService: mapLabellingQuestionAnswerService,
		matchingQuestionAnswerService:     matchingQuestionAnswerService,
		questionUpdator:                   questionUpdator,
	}
}

func (s *ListeningQuestionService) CreateQuestion(ctx context.Context, question *listening.ListeningQuestion) error {
	if question == nil {
		return ErrInvalidInput
	}

	// Validate question
	if err := listeningValidator.ValidateListeningQuestion(question); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Create in database
	if err := s.repo.CreateListeningQuestion(ctx, question); err != nil {
		s.logger.Error("listening_question_service.create", map[string]interface{}{
			"error":         err.Error(),
			"question_type": question.Type,
		}, "Failed to create listening question")
		return err
	}

	questionDetail := listeningDTO.ListeningQuestionDetail{
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
			CreatedAt:   question.CreatedAt,
			UpdatedAt:   question.UpdatedAt,
		},
	}

	// Cache the question with correct completion status
	s.redis.SetCacheListeningQuestionDetail(ctx, &questionDetail, false)

	// Index to OpenSearch with correct status
	status := map[bool]string{true: "complete", false: "uncomplete"}[false]
	if err := s.search.UpsertListeningQuestion(ctx, &questionDetail, status); err != nil {
		s.logger.Error("listening_question_service.create.opensearch", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to index question in OpenSearch")
		return err
	}

	return nil
}

func (s *ListeningQuestionService) GetListeningQuestionDetail(ctx context.Context, id uuid.UUID) (*listeningDTO.ListeningQuestionDetail, error) {
	start := time.Now()
	defer func() {
		s.logger.Info("listening_question_service.get.timing", map[string]interface{}{
			"id":          id,
			"duration_ms": time.Since(start).Milliseconds(),
		}, "Question retrieval timing")
	}()

	// Try to get from cache with both complete and uncomplete status
	pattern := fmt.Sprintf("listening_question:%s:*", id)
	keys, err := s.redis.GetCache().Keys(ctx, pattern)
	if err == nil && len(keys) > 0 {
		// Found in cache, try to get the data
		cachedData, err := s.redis.GetCache().Get(ctx, keys[0])
		if err == nil {
			var response listeningDTO.ListeningQuestionDetail
			if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
				return &response, nil
			}
		}
	}

	// If not in cache or error, get from database
	question, err := s.repo.GetListeningQuestionByID(ctx, id)
	if err != nil {
		s.logger.Error("listening_question_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get listening question")
		return nil, err
	}

	response, err := s.getListeningQuestionDetail(ctx, question)
	if err != nil {
		s.logger.Error("listening_question_service.get.detail", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get question details")
		return nil, err
	}

	// Check completion status
	isComplete := s.completion.IsQuestionComplete(response)

	// Cache with status and version
	if err := s.redis.SetCacheListeningQuestionDetail(ctx, response, isComplete); err != nil {
		s.logger.Error("listening_question_service.get.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to cache question detail")
	}

	return response, nil
}

func (s *ListeningQuestionService) UpdateQuestion(ctx context.Context, id uuid.UUID, update listeningDTO.UpdateListeningQuestionFieldRequest) error {
	// First get the base question from database
	baseQuestion, err := s.repo.GetListeningQuestionByID(ctx, id)
	if err != nil {
		if errors.Is(err, ListeningRepository.ErrQuestionNotFound) {
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
	if err := s.repo.UpdateListeningQuestion(ctx, baseQuestion); err != nil {
		s.logger.Error("listening_question_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    baseQuestion.ID,
		}, "Failed to update question in database")
		return fmt.Errorf("failed to update question in database: %w", err)
	}

	// Use the updator to update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, baseQuestion); err != nil {
		s.logger.Error("listening_question_service.update.cache_and_search", map[string]interface{}{
			"error": err.Error(),
			"id":    baseQuestion.ID,
		}, "Failed to update cache and search")
		// Continue even if cache/search update fails
	}

	return nil
}

func (s *ListeningQuestionService) GetNewUpdatedQuestions(ctx context.Context, versionChecks []struct {
	ID      uuid.UUID
	Version int
}) ([]*listening.ListeningQuestion, error) {
	questionsToRetrieve := make(map[uuid.UUID]int)

	// Check both complete and uncomplete cache keys
	for _, check := range versionChecks {
		completeKey := fmt.Sprintf("listening_question:%s:complete:%d", check.ID, check.Version)
		uncompleteKey := fmt.Sprintf("listening_question:%s:uncomplete:%d", check.ID, check.Version)

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
		return []*listening.ListeningQuestion{}, nil
	}

	// Build query conditions
	conditions := make([]string, 0, len(questionsToRetrieve))
	values := make([]interface{}, 0, len(questionsToRetrieve)*2)
	for id, version := range questionsToRetrieve {
		conditions = append(conditions, "(id = ? AND version > ?)")
		values = append(values, id, version)
	}

	var questions []*listening.ListeningQuestion
	query := s.repo.GetDB().WithContext(ctx).
		Where(strings.Join(conditions, " OR "), values...).
		Order("created_at DESC")

	if err := query.Find(&questions).Error; err != nil {
		s.logger.Error("listening_question_service.get_new_listening_questions", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get updated questions")
		return nil, fmt.Errorf("failed to get updated questions: %w", err)
	}

	return questions, nil
}

func (s *ListeningQuestionService) getListeningQuestionDetail(ctx context.Context, question *listening.ListeningQuestion) (*listeningDTO.ListeningQuestionDetail, error) {
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

	switch question.Type {
	case "FILL_IN_THE_BLANK":
		if err := s.loadFillInTheBlankData(ctx, question.ID, response); err != nil {
			return nil, fmt.Errorf("failed to load fill in blank data: %w", err)
		}
	case "CHOICE_ONE":
		if err := s.loadChoiceOneData(ctx, question.ID, response); err != nil {
			return nil, fmt.Errorf("failed to load choice one data: %w", err)
		}
	case "CHOICE_MULTI":
		if err := s.loadChoiceMultiData(ctx, question.ID, response); err != nil {
			return nil, fmt.Errorf("failed to load choice multi data: %w", err)
		}
	case "MAP_LABELLING":
		if err := s.loadMapLabellingData(ctx, question.ID, response); err != nil {
			return nil, fmt.Errorf("failed to load map labelling data: %w", err)
		}
	case "MATCHING":
		if err := s.loadMatchingData(ctx, question.ID, response); err != nil {
			return nil, fmt.Errorf("failed to load matching data: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown question type: %s", question.Type)
	}

	return response, nil
}

func (s *ListeningQuestionService) loadFillInTheBlankData(ctx context.Context, questionID uuid.UUID, response *listeningDTO.ListeningQuestionDetail) error {
	// Add debug logging
	s.logger.Debug("loadFillInTheBlankData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load fill in blank data")

	questions, err := s.fillInBlankQuestionService.GetQuestionsByListeningQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadFillInTheBlankData.get_questions", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get fill in blank questions")
		return fmt.Errorf("failed to get fill in blank questions: %w", err)
	}

	s.logger.Debug("loadFillInTheBlankData.questions", map[string]interface{}{
		"questionCount": len(questions),
		"questionID":    questionID,
	}, "Retrieved fill in blank questions")

	if len(questions) == 0 {
		s.logger.Warning("loadFillInTheBlankData.no_questions", map[string]interface{}{
			"questionID": questionID,
		}, "No fill in blank questions found")
		return nil
	}

	question := questions[0]
	response.FillInTheBlankQuestion = &listeningDTO.ListeningFillInTheBlankQuestionResponse{
		ID:       question.ID,
		Question: question.Question,
	}

	s.logger.Debug("loadFillInTheBlankData.question_loaded", map[string]interface{}{
		"questionID":       question.ID,
		"parentQuestionID": questionID,
	}, "Loaded fill in blank question")

	answers, err := s.fillInBlankAnswerService.GetAnswersByListeningFillInTheBlankQuestionID(ctx, question.ID)
	if err != nil {
		s.logger.Error("loadFillInTheBlankData.get_answers", map[string]interface{}{
			"error":      err.Error(),
			"questionID": question.ID,
		}, "Failed to get fill in blank answers")
		return fmt.Errorf("failed to get fill in blank answers: %w", err)
	}

	s.logger.Debug("loadFillInTheBlankData.answers", map[string]interface{}{
		"answerCount": len(answers),
		"questionID":  question.ID,
	}, "Retrieved fill in blank answers")

	response.FillInTheBlankAnswers = make([]listeningDTO.ListeningFillInTheBlankAnswerResponse, len(answers))
	for i, answer := range answers {
		response.FillInTheBlankAnswers[i] = listeningDTO.ListeningFillInTheBlankAnswerResponse{
			ID:      answer.ID,
			Answer:  answer.Answer,
			Explain: answer.Explain,
		}
	}

	s.logger.Debug("loadFillInTheBlankData.complete", map[string]interface{}{
		"questionID":    questionID,
		"answersLoaded": len(response.FillInTheBlankAnswers),
	}, "Successfully loaded fill in blank data")

	return nil
}

func (s *ListeningQuestionService) loadChoiceOneData(ctx context.Context, questionID uuid.UUID, response *listeningDTO.ListeningQuestionDetail) error {
	// Add debug logging
	s.logger.Debug("loadChoiceOneData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load choice one data")

	questions, err := s.choiceOneQuestionService.GetQuestionsByListeningQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadChoiceOneData.get_questions", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get choice one questions")
		return fmt.Errorf("failed to get choice one questions: %w", err)
	}

	s.logger.Debug("loadChoiceOneData.questions", map[string]interface{}{
		"questionCount": len(questions),
		"questionID":    questionID,
	}, "Retrieved choice one questions")

	if len(questions) > 0 {
		question := questions[0]
		response.ChoiceOneQuestion = &listeningDTO.ListeningChoiceOneQuestionResponse{
			ID:       question.ID,
			Question: question.Question,
			Explain:  question.Explain,
		}

		s.logger.Debug("loadChoiceOneData.question_loaded", map[string]interface{}{
			"questionID":       question.ID,
			"parentQuestionID": questionID,
		}, "Loaded choice one question")

		options, err := s.choiceOneOptionService.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			s.logger.Error("loadChoiceOneData.get_options", map[string]interface{}{
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

		s.logger.Debug("loadChoiceOneData.complete", map[string]interface{}{
			"questionID":    questionID,
			"optionsLoaded": len(response.ChoiceOneOptions),
		}, "Successfully loaded choice one data")
	}

	return nil
}

func (s *ListeningQuestionService) loadChoiceMultiData(ctx context.Context, questionID uuid.UUID, response *listeningDTO.ListeningQuestionDetail) error {
	s.logger.Debug("loadChoiceMultiData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load choice multi data")

	questions, err := s.choiceMultiQuestionService.GetQuestionsByListeningQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadChoiceMultiData.get_questions", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get choice multi questions")
		return fmt.Errorf("failed to get choice multi questions: %w", err)
	}

	s.logger.Debug("loadChoiceMultiData.questions", map[string]interface{}{
		"questionCount": len(questions),
		"questionID":    questionID,
	}, "Retrieved choice multi questions")

	if len(questions) > 0 {
		question := questions[0]
		response.ChoiceMultiQuestion = &listeningDTO.ListeningChoiceMultiQuestionResponse{
			ID:       question.ID,
			Question: question.Question,
			Explain:  question.Explain,
		}

		s.logger.Debug("loadChoiceMultiData.question_loaded", map[string]interface{}{
			"questionID":       question.ID,
			"parentQuestionID": questionID,
		}, "Loaded choice multi question")

		options, err := s.choiceMultiOptionService.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			s.logger.Error("loadChoiceMultiData.get_options", map[string]interface{}{
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

		s.logger.Debug("loadChoiceMultiData.complete", map[string]interface{}{
			"questionID":    questionID,
			"optionsLoaded": len(response.ChoiceMultiOptions),
		}, "Successfully loaded choice multi data")
	}

	return nil
}

func (s *ListeningQuestionService) loadMapLabellingData(ctx context.Context, questionID uuid.UUID, response *listeningDTO.ListeningQuestionDetail) error {
	s.logger.Debug("loadMapLabellingData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load map labelling data")

	// Get map labelling questions directly - no need for parent record
	qas, err := s.mapLabellingQuestionAnswerService.GetByListeningQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadMapLabellingData.get_qas", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get map labelling QAs")
		return fmt.Errorf("failed to get map labelling QAs: %w", err)
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
	}

	s.logger.Debug("loadMapLabellingData.complete", map[string]interface{}{
		"questionID": questionID,
		"qasLoaded":  len(response.MapLabelling),
	}, "Successfully loaded map labelling data")

	return nil
}

func (s *ListeningQuestionService) loadMatchingData(ctx context.Context, questionID uuid.UUID, response *listeningDTO.ListeningQuestionDetail) error {
	s.logger.Debug("loadMatchingData.start", map[string]interface{}{
		"questionID": questionID,
		"type":       response.Type,
	}, "Starting to load matching data")

	// Get matching questions directly - no need for parent record
	qas, err := s.matchingQuestionAnswerService.GetByListeningQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadMatchingData.get_qas", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get matching QAs")
		return fmt.Errorf("failed to get matching QAs: %w", err)
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
	}

	s.logger.Debug("loadMatchingData.complete", map[string]interface{}{
		"questionID": questionID,
		"qasLoaded":  len(response.Matching),
	}, "Successfully loaded matching data")

	return nil
}

func (s *ListeningQuestionService) GetListeningByListID(ctx context.Context, ids []uuid.UUID) ([]*listeningDTO.ListeningQuestionDetail, error) {
	// Track which IDs need to be fetched from database
	missingIDs := make([]uuid.UUID, 0)
	result := make([]*listeningDTO.ListeningQuestionDetail, 0, len(ids))

	// Try to get from cache first
	for _, id := range ids {
		// Try both complete and uncomplete patterns
		pattern := fmt.Sprintf("listening_question:%s:*", id)
		keys, err := s.redis.GetCache().Keys(ctx, pattern)
		if err == nil && len(keys) > 0 {
			// Found in cache, try to get the data
			cachedData, err := s.redis.GetCache().Get(ctx, keys[0])
			if err == nil {
				var questionDetail listeningDTO.ListeningQuestionDetail
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

		var dbQuestions []*listening.ListeningQuestion
		query := s.repo.GetDB().WithContext(ctx).
			Where(strings.Join(conditions, " OR "), values...).
			Find(&dbQuestions)

		if query.Error != nil {
			s.logger.Error("listening_question_service.get_by_list_id", map[string]interface{}{
				"error": query.Error.Error(),
			}, "Failed to get questions from database")
			return nil, query.Error
		}

		// Process and cache the database results
		for _, q := range dbQuestions {
			// Build complete response with details
			questionDetail, err := s.getListeningQuestionDetail(ctx, q)
			if err != nil {
				s.logger.Error("listening_question_service.get_by_list_id.build_response", map[string]interface{}{
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
			cacheKey := fmt.Sprintf("listening_question:%s:%s:%d", q.ID, status, q.Version)
			if questionJSON, err := json.Marshal(questionDetail); err == nil {
				if err := s.redis.GetCache().Set(ctx, cacheKey, string(questionJSON), 24*time.Hour); err != nil {
					s.logger.Error("listening_question_service.get_by_list_id.cache", map[string]interface{}{
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

func (s *ListeningQuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	// Delete from database
	if err := s.repo.DeleteListeningQuestion(ctx, id); err != nil {
		s.logger.Error("listening_question_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question")
		return err
	}

	// Invalidate cache
	s.redis.RemoveListeningQuestionCacheEntries(ctx, id)

	// Delete from OpenSearch
	if err := s.search.DeleteListeningQuestionFromIndex(ctx, id); err != nil {
		s.logger.Error("listening_question_service.delete.opensearch", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question from OpenSearch")
		// Continue even if OpenSearch delete fails
	}

	return nil
}

func (s *ListeningQuestionService) SearchQuestionsWithFilter(ctx context.Context, filter listeningDTO.ListeningQuestionSearchFilter) (*listeningDTO.ListListeningQuestionsPagination, error) {
	// Calculate offset
	from := (filter.Page - 1) * filter.PageSize

	// Build bool query
	boolQuery := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": []map[string]interface{}{},
		},
	}

	// Add type filter
	if filter.Type != "" {
		boolQuery["bool"].(map[string]interface{})["must"] = append(
			boolQuery["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"match": map[string]interface{}{
					"type": filter.Type,
				},
			},
		)
	}

	// Add topic filter
	if filter.Topic != "" {
		topics := strings.Split(filter.Topic, ",")
		topicTerms := make([]interface{}, len(topics))
		for i, topic := range topics {
			topicTerms[i] = strings.TrimSpace(topic)
		}
		boolQuery["bool"].(map[string]interface{})["must"] = append(
			boolQuery["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"terms": map[string]interface{}{
					"topic.keyword": topicTerms,
				},
			},
		)
	}

	// Add instruction filter
	if filter.Instruction != "" {
		boolQuery["bool"].(map[string]interface{})["must"] = append(
			boolQuery["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"match": map[string]interface{}{
					"instruction": filter.Instruction,
				},
			},
		)
	}

	// Add audio_urls filter
	if filter.AudioURLs != "" {
		boolQuery["bool"].(map[string]interface{})["must"] = append(
			boolQuery["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"term": map[string]interface{}{
					"audio_urls": filter.AudioURLs,
				},
			},
		)
	}

	// Add image_urls filter
	if filter.ImageURLs != "" {
		boolQuery["bool"].(map[string]interface{})["must"] = append(
			boolQuery["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"term": map[string]interface{}{
					"image_urls": filter.ImageURLs,
				},
			},
		)
	}

	// Add transcript filter
	if filter.Transcript != "" {
		boolQuery["bool"].(map[string]interface{})["must"] = append(
			boolQuery["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"match": map[string]interface{}{
					"transcript": filter.Transcript,
				},
			},
		)
	}

	// Add max_time range filter
	if filter.MaxTime != "" {
		parts := strings.Split(filter.MaxTime, "-")
		if len(parts) == 2 {
			minTime, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			maxTime, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil {
				rangeQuery := map[string]interface{}{
					"range": map[string]interface{}{
						"max_time": map[string]interface{}{
							"gte": minTime,
							"lte": maxTime,
						},
					},
				}
				boolQuery["bool"].(map[string]interface{})["must"] = append(
					boolQuery["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
					rangeQuery,
				)
			}
		}
	}

	// Add metadata filter based on type
	if filter.Metadata != "" && filter.Type != "" {
		metadataQuery := map[string]interface{}{}

		switch filter.Type {
		case "FILL_IN_THE_BLANK":
			metadataQuery = map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query": filter.Metadata,
					"fields": []string{
						"fill_in_the_blank_question",
						"fill_in_the_blank_answers",
					},
				},
			}
		case "CHOICE_ONE", "CHOICE_MULTI":
			metadataQuery = map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query": filter.Metadata,
					"fields": []string{
						filter.Type + "_question",
						filter.Type + "_options",
					},
				},
			}
		case "MAP_LABELLING", "MATCHING":
			metadataQuery = map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query": filter.Metadata,
					"fields": []string{
						filter.Type + "_questions_and_answers",
					},
				},
			}
		}

		if len(metadataQuery) > 0 {
			boolQuery["bool"].(map[string]interface{})["must"] = append(
				boolQuery["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
				metadataQuery,
			)
		}
	}

	// Build final search body
	searchBody := map[string]interface{}{
		"query":            boolQuery,
		"from":             from,
		"size":             filter.PageSize,
		"track_total_hits": true,
	}

	// Add debug logging
	searchJSON, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search body: %w", err)
	}

	s.logger.Debug("opensearch_query_details", map[string]interface{}{
		"query_json": string(searchJSON),
		"type":       filter.Type,
		"topics":     filter.Topic,
		"max_time":   filter.MaxTime,
		"page":       filter.Page,
		"page_size":  filter.PageSize,
		"bool_query": boolQuery,
	}, "OpenSearch query details")

	// Execute search
	searchReq := opensearchapi.SearchRequest{
		Index: []string{"listening_questions"},
		Body:  bytes.NewReader(searchJSON),
	}

	searchRes, err := searchReq.Do(ctx, s.search.GetClient())
	if err != nil {
		s.logger.Error("opensearch_search_error", map[string]interface{}{
			"error": err.Error(),
			"query": string(searchJSON),
		}, "Failed to execute OpenSearch query")
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer searchRes.Body.Close()

	// Add more detailed debug logging for response
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(searchRes.Body); err != nil {
		s.logger.Error("opensearch_read_error", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to read OpenSearch response")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	searchRes.Body = ioutil.NopCloser(&buf)

	s.logger.Debug("opensearch_response_details", map[string]interface{}{
		"response": buf.String(),
	}, "OpenSearch response details")

	var searchResult struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source struct {
					ID                              uuid.UUID `json:"id"`
					Type                            string    `json:"type"`
					Topic                           []string  `json:"topic"`
					Instruction                     string    `json:"instruction"`
					AudioURLs                       []string  `json:"audio_urls"`
					ImageURLs                       []string  `json:"image_urls"`
					Transcript                      string    `json:"transcript"`
					MaxTime                         int       `json:"max_time"`
					Version                         int       `json:"version"`
					Status                          string    `json:"status"`
					ChoiceMultiQuestion             string    `json:"choice_multi_question"`
					ChoiceMultiOptions              string    `json:"choice_multi_options"`
					ChoiceOneQuestion               string    `json:"choice_one_question"`
					ChoiceOneOptions                string    `json:"choice_one_options"`
					FillInTheBlankQuestion          string    `json:"fill_in_the_blank_question"`
					FillInTheBlankAnswers           string    `json:"fill_in_the_blank_answers"`
					MapLabelling                    string    `json:"map_labelling"`
					MapLabellingQuestionsAndAnswers string    `json:"map_labelling_questions_and_answers"`
					Matching                        string    `json:"MATCHING"`
					MatchingQuestionsAndAnswers     string    `json:"matching_questions_and_answers"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(&buf).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	var questions []listeningDTO.ListeningQuestionDetail
	for _, hit := range searchResult.Hits.Hits {
		question := listeningDTO.ListeningQuestionDetail{
			ListeningQuestionResponse: listeningDTO.ListeningQuestionResponse{
				ID:          hit.Source.ID,
				Type:        hit.Source.Type,
				Topic:       hit.Source.Topic,
				Instruction: hit.Source.Instruction,
				AudioURLs:   hit.Source.AudioURLs,
				ImageURLs:   hit.Source.ImageURLs,
				Transcript:  hit.Source.Transcript,
				MaxTime:     hit.Source.MaxTime,
				Version:     hit.Source.Version,
			},
		}

		// Parse additional fields based on question type
		switch hit.Source.Type {
		case "CHOICE_MULTI":
			if hit.Source.ChoiceMultiQuestion != "" && hit.Source.ChoiceMultiQuestion != "null" {
				var choiceMultiQuestion listeningDTO.ListeningChoiceMultiQuestionResponse
				if err := json.Unmarshal([]byte(hit.Source.ChoiceMultiQuestion), &choiceMultiQuestion); err == nil {
					question.ChoiceMultiQuestion = &choiceMultiQuestion
				}
			}
			if hit.Source.ChoiceMultiOptions != "" && hit.Source.ChoiceMultiOptions != "null" {
				var choiceMultiOptions []listeningDTO.ListeningChoiceMultiOptionResponse
				if err := json.Unmarshal([]byte(hit.Source.ChoiceMultiOptions), &choiceMultiOptions); err == nil {
					question.ChoiceMultiOptions = choiceMultiOptions
				}
			}
		case "CHOICE_ONE":
			if hit.Source.ChoiceOneQuestion != "" && hit.Source.ChoiceOneQuestion != "null" {
				var choiceOneQuestion listeningDTO.ListeningChoiceOneQuestionResponse
				if err := json.Unmarshal([]byte(hit.Source.ChoiceOneQuestion), &choiceOneQuestion); err == nil {
					question.ChoiceOneQuestion = &choiceOneQuestion
				}
			}
			if hit.Source.ChoiceOneOptions != "" && hit.Source.ChoiceOneOptions != "null" {
				var choiceOneOptions []listeningDTO.ListeningChoiceOneOptionResponse
				if err := json.Unmarshal([]byte(hit.Source.ChoiceOneOptions), &choiceOneOptions); err == nil {
					question.ChoiceOneOptions = choiceOneOptions
				}
			}
		case "FILL_IN_THE_BLANK":
			if hit.Source.FillInTheBlankQuestion != "" && hit.Source.FillInTheBlankQuestion != "null" {
				var fillInBlankQuestion listeningDTO.ListeningFillInTheBlankQuestionResponse
				if err := json.Unmarshal([]byte(hit.Source.FillInTheBlankQuestion), &fillInBlankQuestion); err == nil {
					question.FillInTheBlankQuestion = &fillInBlankQuestion
				}
			}
			if hit.Source.FillInTheBlankAnswers != "" && hit.Source.FillInTheBlankAnswers != "null" {
				var fillInBlankAnswers []listeningDTO.ListeningFillInTheBlankAnswerResponse
				if err := json.Unmarshal([]byte(hit.Source.FillInTheBlankAnswers), &fillInBlankAnswers); err == nil {
					question.FillInTheBlankAnswers = fillInBlankAnswers
				}
			}
		}

		questions = append(questions, question)
	}

	result := &listeningDTO.ListListeningQuestionsPagination{
		Questions: questions,
		Total:     searchResult.Hits.Total.Value,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
	}

	return result, nil
}

func (s *ListeningQuestionService) DeleteAllQuestions(ctx context.Context) error {
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

	// Delete all records from listening_questions
	if err := tx.Exec("DELETE FROM listening_questions").Error; err != nil {
		tx.Rollback()
		s.logger.Error("listening_question_service.delete_all", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete all listening questions")
		return err
	}

	// Delete all Redis cache with pattern listening_question:*
	if err := s.redis.GetCache().DeletePattern(ctx, "listening_question:*"); err != nil {
		tx.Rollback()
		s.logger.Error("listening_question_service.delete_all.cache", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete Redis cache")
		return err
	}

	// Delete OpenSearch index
	if err := s.search.RemoveListeningQuestionsIndex(ctx); err != nil {
		tx.Rollback()
		s.logger.Error("listening_question_service.delete_all.search", map[string]interface{}{
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
