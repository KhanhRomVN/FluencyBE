package writing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	writingDTO "fluencybe/internal/app/dto"
	writingHelper "fluencybe/internal/app/helper/writing"
	"fluencybe/internal/app/model/writing"
	searchClient "fluencybe/internal/app/opensearch"
	redisClient "fluencybe/internal/app/redis"
	writingRepository "fluencybe/internal/app/repository/writing"
	writingValidator "fluencybe/internal/app/validator"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"

	"github.com/google/uuid"
)

var (
	ErrQuestionNotFound = errors.New("writing question not found")
	ErrInvalidInput     = errors.New("invalid input")
)

type WritingQuestionService struct {
	repo                      *writingRepository.WritingQuestionRepository
	logger                    *logger.PrettyLogger
	redis                     *redisClient.WritingQuestionRedis
	search                    *searchClient.WritingQuestionSearch
	completion                *writingHelper.WritingQuestionCompletionHelper
	updater                   *writingHelper.WritingQuestionFieldUpdater
	questionUpdator           *writingHelper.WritingQuestionUpdator
	sentenceCompletionService *WritingSentenceCompletionService
	essayService              *WritingEssayService
}

func NewWritingQuestionService(
	repo *writingRepository.WritingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	search *searchClient.WritingQuestionSearch,
	sentenceCompletionService *WritingSentenceCompletionService,
	essayService *WritingEssayService,
	questionUpdator *writingHelper.WritingQuestionUpdator,
) *WritingQuestionService {
	return &WritingQuestionService{
		repo:                      repo,
		logger:                    logger,
		redis:                     redisClient.NewWritingQuestionRedis(cache, logger),
		search:                    search,
		completion:                writingHelper.NewWritingQuestionCompletionHelper(logger),
		updater:                   writingHelper.NewWritingQuestionFieldUpdater(logger),
		sentenceCompletionService: sentenceCompletionService,
		essayService:              essayService,
		questionUpdator:           questionUpdator,
	}
}

func (s *WritingQuestionService) CreateQuestion(ctx context.Context, question *writing.WritingQuestion) error {
	if err := writingValidator.ValidateWritingQuestion(question); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if err := s.repo.CreateWritingQuestion(ctx, question); err != nil {
		s.logger.Error("writing_question_service.create", map[string]interface{}{
			"error":         err.Error(),
			"question_type": question.Type,
		}, "Failed to create writing question")
		return err
	}

	questionDetail := writingDTO.WritingQuestionDetail{
		WritingQuestionResponse: writingDTO.WritingQuestionResponse{
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
	s.redis.SetCacheWritingQuestionDetail(ctx, &questionDetail, false)

	// Index to OpenSearch with correct status
	status := map[bool]string{true: "complete", false: "uncomplete"}[false]
	if err := s.search.UpsertWritingQuestion(ctx, &questionDetail, status); err != nil {
		s.logger.Error("writing_question_service.create.opensearch", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to index question in OpenSearch")
	}

	return nil
}

func (s *WritingQuestionService) DeleteAllQuestions(ctx context.Context) error {
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

	// Delete all records from writing_questions
	if err := tx.Exec("DELETE FROM writing_questions").Error; err != nil {
		tx.Rollback()
		s.logger.Error("writing_question_service.delete_all", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete all writing questions")
		return err
	}

	// Delete all Redis cache with pattern writing_question:*
	if err := s.redis.GetCache().DeletePattern(ctx, "writing_question:*"); err != nil {
		tx.Rollback()
		s.logger.Error("writing_question_service.delete_all.cache", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete Redis cache")
		return err
	}

	// Delete OpenSearch index
	if err := s.search.RemoveWritingQuestionsIndex(ctx); err != nil {
		tx.Rollback()
		s.logger.Error("writing_question_service.delete_all.search", map[string]interface{}{
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

func (s *WritingQuestionService) SearchQuestionsWithFilter(ctx context.Context, filter writingDTO.WritingQuestionSearchFilter) (*writingDTO.ListWritingQuestionsPagination, error) {
	s.logger.Debug("writing_question_service.search.start", map[string]interface{}{
		"filter": filter,
	}, "Starting question search")

	result, err := s.search.SearchQuestions(ctx, filter)
	if err != nil {
		s.logger.Error("writing_question_service.search", map[string]interface{}{
			"error":  err.Error(),
			"filter": filter,
		}, "Failed to search questions")
		return nil, fmt.Errorf("failed to search questions: %w", err)
	}

	s.logger.Debug("writing_question_service.search.complete", map[string]interface{}{
		"total_results": result.Total,
		"page":          result.Page,
		"page_size":     result.PageSize,
	}, "Search completed")

	return result, nil
}

func (s *WritingQuestionService) GetNewUpdatedQuestions(ctx context.Context, versionChecks []struct {
	ID      uuid.UUID
	Version int
}) ([]*writing.WritingQuestion, error) {
	questionsToRetrieve := make(map[uuid.UUID]int)

	// Check both complete and uncomplete cache keys
	for _, check := range versionChecks {
		completeKey := fmt.Sprintf("writing_question:%s:complete:%d", check.ID, check.Version)
		uncompleteKey := fmt.Sprintf("writing_question:%s:uncomplete:%d", check.ID, check.Version)

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
		return []*writing.WritingQuestion{}, nil
	}

	// Build query conditions
	conditions := make([]string, 0, len(questionsToRetrieve))
	values := make([]interface{}, 0, len(questionsToRetrieve)*2)
	for id, version := range questionsToRetrieve {
		conditions = append(conditions, "(id = ? AND version > ?)")
		values = append(values, id, version)
	}

	var questions []*writing.WritingQuestion
	query := s.repo.GetDB().WithContext(ctx).
		Where(strings.Join(conditions, " OR "), values...).
		Order("created_at DESC")

	if err := query.Find(&questions).Error; err != nil {
		s.logger.Error("writing_question_service.get_new_writing_questions", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get updated questions")
		return nil, fmt.Errorf("failed to get updated questions: %w", err)
	}

	return questions, nil
}

func (s *WritingQuestionService) GetWritingQuestionDetail(ctx context.Context, id uuid.UUID) (*writingDTO.WritingQuestionDetail, error) {
	question, err := s.repo.GetWritingQuestionByID(ctx, id)
	if err != nil {
		s.logger.Error("writing_question_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get writing question")
		return nil, err
	}

	response, err := s.buildQuestionDetail(ctx, question)
	if err != nil {
		s.logger.Error("writing_question_service.get.detail", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get question details")
		return nil, err
	}

	// Check completion status
	isComplete := s.completion.IsQuestionComplete(response)

	// Cache with status and version
	if err := s.redis.SetCacheWritingQuestionDetail(ctx, response, isComplete); err != nil {
		s.logger.Error("writing_question_service.get.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to cache question detail")
	}

	return response, nil
}

func (s *WritingQuestionService) buildQuestionDetail(ctx context.Context, question *writing.WritingQuestion) (*writingDTO.WritingQuestionDetail, error) {
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
		err = s.loadSentenceCompletionData(ctx, question.ID, response)
	case "ESSAY":
		err = s.loadEssayData(ctx, question.ID, response)
	default:
		return nil, fmt.Errorf("unknown question type: %s", question.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load %s data: %w", question.Type, err)
	}

	return response, nil
}

func (s *WritingQuestionService) loadSentenceCompletionData(ctx context.Context, questionID uuid.UUID, response *writingDTO.WritingQuestionDetail) error {
	sentences, err := s.sentenceCompletionService.GetByWritingQuestionID(ctx, questionID)
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

func (s *WritingQuestionService) loadEssayData(ctx context.Context, questionID uuid.UUID, response *writingDTO.WritingQuestionDetail) error {
	essays, err := s.essayService.GetByWritingQuestionID(ctx, questionID)
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

func (s *WritingQuestionService) UpdateQuestion(ctx context.Context, id uuid.UUID, update writingDTO.UpdateWritingQuestionFieldRequest) error {
	// First get the base question from database
	baseQuestion, err := s.repo.GetWritingQuestionByID(ctx, id)
	if err != nil {
		if errors.Is(err, writingRepository.ErrQuestionNotFound) {
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
	if err := s.repo.UpdateWritingQuestion(ctx, baseQuestion); err != nil {
		s.logger.Error("writing_question_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    baseQuestion.ID,
		}, "Failed to update question in database")
		return fmt.Errorf("failed to update question in database: %w", err)
	}

	// Use the updator to update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, baseQuestion); err != nil {
		s.logger.Error("writing_question_service.update.cache_and_search", map[string]interface{}{
			"error": err.Error(),
			"id":    baseQuestion.ID,
		}, "Failed to update cache and search")
		// Continue even if cache/search update fails
	}

	return nil
}

func (s *WritingQuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteWritingQuestion(ctx, id); err != nil {
		s.logger.Error("writing_question_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete writing question")
		return err
	}

	// Invalidate cache
	s.redis.RemoveWritingQuestionCacheEntries(ctx, id)

	// Delete from OpenSearch
	if err := s.search.DeleteWritingQuestionFromIndex(ctx, id); err != nil {
		s.logger.Error("writing_question_service.delete.opensearch", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question from OpenSearch")
		// Continue even if OpenSearch delete fails
	}

	return nil
}

func (s *WritingQuestionService) GetWritingByListID(ctx context.Context, ids []uuid.UUID) ([]*writingDTO.WritingQuestionDetail, error) {
	// Track which IDs need to be fetched from database
	missingIDs := make([]uuid.UUID, 0)
	result := make([]*writingDTO.WritingQuestionDetail, 0, len(ids))

	// Try to get from cache first
	for _, id := range ids {
		// Try both complete and uncomplete patterns
		pattern := fmt.Sprintf("writing_question:%s:*", id)
		keys, err := s.redis.GetCache().Keys(ctx, pattern)
		if err == nil && len(keys) > 0 {
			// Found in cache, try to get the data
			cachedData, err := s.redis.GetCache().Get(ctx, keys[0])
			if err == nil {
				var questionDetail writingDTO.WritingQuestionDetail
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

		var dbQuestions []*writing.WritingQuestion
		query := s.repo.GetDB().WithContext(ctx).
			Where(strings.Join(conditions, " OR "), values...).
			Find(&dbQuestions)

		if query.Error != nil {
			s.logger.Error("writing_question_service.get_by_list_id", map[string]interface{}{
				"error": query.Error.Error(),
			}, "Failed to get questions from database")
			return nil, query.Error
		}

		// Process and cache the database results
		for _, q := range dbQuestions {
			// Build complete response with details
			questionDetail, err := s.buildQuestionDetail(ctx, q)
			if err != nil {
				s.logger.Error("writing_question_service.get_by_list_id.build_response", map[string]interface{}{
					"error": err.Error(),
					"id":    q.ID,
				}, "Failed to build complete response")
				continue
			}

			// Determine completion status
			isComplete := s.completion.IsQuestionComplete(questionDetail)

			// Cache with status and version
			if err := s.redis.SetCacheWritingQuestionDetail(ctx, questionDetail, isComplete); err != nil {
				s.logger.Error("writing_question_service.get_by_list_id.cache", map[string]interface{}{
					"error": err.Error(),
					"id":    q.ID,
				}, "Failed to cache question")
			}

			result = append(result, questionDetail)
		}
	}

	return result, nil
}
