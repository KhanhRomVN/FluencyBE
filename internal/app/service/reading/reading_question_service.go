package reading

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	readingDTO "fluencybe/internal/app/dto"
	readingHelper "fluencybe/internal/app/helper/reading"
	"fluencybe/internal/app/model/reading"
	searchClient "fluencybe/internal/app/opensearch"
	redisClient "fluencybe/internal/app/redis"
	ReadingRepository "fluencybe/internal/app/repository/reading"
	readingValidator "fluencybe/internal/app/validator"
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
	ErrQuestionNotFound = errors.New("reading question not found")
	ErrInvalidInput     = errors.New("invalid input")
)

type ReadingQuestionService struct {
	repo                       *ReadingRepository.ReadingQuestionRepository
	logger                     *logger.PrettyLogger
	redis                      *redisClient.ReadingQuestionRedis
	search                     *searchClient.ReadingQuestionSearch
	completion                 *readingHelper.ReadingQuestionCompletionHelper
	updater                    *readingHelper.ReadingQuestionFieldUpdater
	questionUpdator            *readingHelper.ReadingQuestionUpdator
	fillInBlankQuestionService *ReadingFillInTheBlankQuestionService
	fillInBlankAnswerService   *ReadingFillInTheBlankAnswerService
	choiceOneQuestionService   *ReadingChoiceOneQuestionService
	choiceOneOptionService     *ReadingChoiceOneOptionService
	choiceMultiQuestionService *ReadingChoiceMultiQuestionService
	choiceMultiOptionService   *ReadingChoiceMultiOptionService
	trueFalseService           *ReadingTrueFalseService
	matchingService            *ReadingMatchingService
}

func NewReadingQuestionService(
	repo *ReadingRepository.ReadingQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	openSearch *opensearch.Client,
	fillInBlankQuestionService *ReadingFillInTheBlankQuestionService,
	fillInBlankAnswerService *ReadingFillInTheBlankAnswerService,
	choiceOneQuestionService *ReadingChoiceOneQuestionService,
	choiceOneOptionService *ReadingChoiceOneOptionService,
	choiceMultiQuestionService *ReadingChoiceMultiQuestionService,
	choiceMultiOptionService *ReadingChoiceMultiOptionService,
	trueFalseService *ReadingTrueFalseService,
	matchingService *ReadingMatchingService,
	questionUpdator *readingHelper.ReadingQuestionUpdator,
) *ReadingQuestionService {
	return &ReadingQuestionService{
		repo:                       repo,
		logger:                     logger,
		redis:                      redisClient.NewReadingQuestionRedis(cache, logger),
		search:                     searchClient.NewReadingQuestionSearch(openSearch, logger),
		completion:                 readingHelper.NewReadingQuestionCompletionHelper(logger),
		updater:                    readingHelper.NewReadingQuestionFieldUpdater(logger),
		fillInBlankQuestionService: fillInBlankQuestionService,
		fillInBlankAnswerService:   fillInBlankAnswerService,
		choiceOneQuestionService:   choiceOneQuestionService,
		choiceOneOptionService:     choiceOneOptionService,
		choiceMultiQuestionService: choiceMultiQuestionService,
		choiceMultiOptionService:   choiceMultiOptionService,
		trueFalseService:           trueFalseService,
		matchingService:            matchingService,
		questionUpdator:            questionUpdator,
	}
}

func (s *ReadingQuestionService) CreateQuestion(ctx context.Context, question *reading.ReadingQuestion) error {
	if question == nil {
		return ErrInvalidInput
	}

	// Validate question
	if err := readingValidator.ValidateReadingQuestion(question); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Create in database
	if err := s.repo.CreateReadingQuestion(ctx, question); err != nil {
		s.logger.Error("reading_question_service.create", map[string]interface{}{
			"error":         err.Error(),
			"question_type": question.Type,
		}, "Failed to create reading question")
		return err
	}

	// Use questionUpdator to build complete question detail and update cache/search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, question); err != nil {
		s.logger.Error("reading_question_service.create.cache_and_search", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update cache and search")
		return err
	}

	return nil
}

func (s *ReadingQuestionService) GetReadingQuestionDetail(ctx context.Context, id uuid.UUID) (*readingDTO.ReadingQuestionDetail, error) {
	// Try to get from cache with both complete and uncomplete status
	pattern := fmt.Sprintf("reading_question:%s:*", id)
	keys, err := s.redis.GetCache().Keys(ctx, pattern)
	if err == nil && len(keys) > 0 {
		// Found in cache, try to get the data
		cachedData, err := s.redis.GetCache().Get(ctx, keys[0])
		if err == nil {
			var response readingDTO.ReadingQuestionDetail
			if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
				return &response, nil
			}
		}
	}

	// If not in cache or error, get from database
	question, err := s.repo.GetReadingQuestionByID(ctx, id)
	if err != nil {
		s.logger.Error("reading_question_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get reading question")
		return nil, err
	}

	response, err := s.getReadingQuestionDetail(ctx, question)
	if err != nil {
		s.logger.Error("reading_question_service.get.detail", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get question details")
		return nil, err
	}

	// Check completion status
	isComplete := s.completion.IsQuestionComplete(response)

	// Cache with status and version
	if err := s.redis.SetCacheReadingQuestionDetail(ctx, response, isComplete); err != nil {
		s.logger.Error("reading_question_service.get.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to cache question detail")
	}

	return response, nil
}

func (s *ReadingQuestionService) getReadingQuestionDetail(ctx context.Context, question *reading.ReadingQuestion) (*readingDTO.ReadingQuestionDetail, error) {
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
	case "TRUE_FALSE":
		if err := s.loadTrueFalseData(ctx, question.ID, response); err != nil {
			return nil, fmt.Errorf("failed to load true/false data: %w", err)
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

func (s *ReadingQuestionService) loadFillInTheBlankData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	questions, err := s.fillInBlankQuestionService.GetQuestionsByReadingQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadFillInTheBlankData.get_questions", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get fill in blank questions")
		return fmt.Errorf("failed to get fill in blank questions: %w", err)
	}

	if len(questions) > 0 {
		question := questions[0]
		response.FillInTheBlankQuestion = &readingDTO.ReadingFillInTheBlankQuestionResponse{
			ID:       question.ID,
			Question: question.Question,
		}

		answers, err := s.fillInBlankAnswerService.GetAnswersByReadingFillInTheBlankQuestionID(ctx, question.ID)
		if err != nil {
			s.logger.Error("loadFillInTheBlankData.get_answers", map[string]interface{}{
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
	}

	return nil
}

func (s *ReadingQuestionService) loadChoiceOneData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	questions, err := s.choiceOneQuestionService.GetQuestionsByReadingQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadChoiceOneData.get_questions", map[string]interface{}{
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

		options, err := s.choiceOneOptionService.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			s.logger.Error("loadChoiceOneData.get_options", map[string]interface{}{
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

func (s *ReadingQuestionService) loadChoiceMultiData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	questions, err := s.choiceMultiQuestionService.GetQuestionsByReadingQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadChoiceMultiData.get_questions", map[string]interface{}{
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

		options, err := s.choiceMultiOptionService.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			s.logger.Error("loadChoiceMultiData.get_options", map[string]interface{}{
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

func (s *ReadingQuestionService) loadTrueFalseData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	trueFalses, err := s.trueFalseService.GetByReadingQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadTrueFalseData.get_questions", map[string]interface{}{
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

func (s *ReadingQuestionService) loadMatchingData(ctx context.Context, questionID uuid.UUID, response *readingDTO.ReadingQuestionDetail) error {
	matchings, err := s.matchingService.GetByReadingQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("loadMatchingData.get_matchings", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get matching questions")
		return fmt.Errorf("failed to get matching questions: %w", err)
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

func (s *ReadingQuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	// Delete from database
	if err := s.repo.DeleteReadingQuestion(ctx, id); err != nil {
		s.logger.Error("listening_question_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question")
		return err
	}

	// Invalidate cache
	s.redis.RemoveReadingQuestionCacheEntries(ctx, id)

	// Delete from OpenSearch
	if err := s.search.DeleteReadingQuestionFromIndex(ctx, id); err != nil {
		s.logger.Error("listening_question_service.delete.opensearch", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question from OpenSearch")
		// Continue even if OpenSearch delete fails
	}

	return nil
}

func (s *ReadingQuestionService) SearchQuestionsWithFilter(ctx context.Context, filter readingDTO.ReadingQuestionSearchFilter) (*readingDTO.ListReadingQuestionsPagination, error) {
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

	// Add title filter
	if filter.Title != "" {
		boolQuery["bool"].(map[string]interface{})["must"] = append(
			boolQuery["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"match": map[string]interface{}{
					"title": filter.Title,
				},
			},
		)
	}

	// Add passages filter
	if filter.Passages != "" {
		boolQuery["bool"].(map[string]interface{})["must"] = append(
			boolQuery["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"match": map[string]interface{}{
					"passages": filter.Passages,
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
		case "TRUE_FALSE":
			metadataQuery = map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query": filter.Metadata,
					"fields": []string{
						"true_false_question",
						"true_false_answer",
					},
				},
			}
		case "MATCHING":
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
		Index: []string{"reading_questions"},
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
					ID                     uuid.UUID `json:"id"`
					Type                   string    `json:"type"`
					Topic                  []string  `json:"topic"`
					Instruction            string    `json:"instruction"`
					Title                  string    `json:"title"`
					Passages               []string  `json:"passages"`
					ImageURLs              []string  `json:"image_urls"`
					MaxTime                int       `json:"max_time"`
					Version                int       `json:"version"`
					Status                 string    `json:"status"`
					TrueFalse              string    `json:"true_false"`
					ChoiceMultiQuestion    string    `json:"choice_multi_question"`
					ChoiceMultiOptions     string    `json:"choice_multi_options"`
					ChoiceOneQuestion      string    `json:"choice_one_question"`
					ChoiceOneOptions       string    `json:"choice_one_options"`
					FillInTheBlankQuestion string    `json:"fill_in_the_blank_question"`
					FillInTheBlankAnswers  string    `json:"fill_in_the_blank_answers"`
					Matching               string    `json:"MATCHING"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(&buf).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	var questions []readingDTO.ReadingQuestionDetail
	for _, hit := range searchResult.Hits.Hits {
		question := readingDTO.ReadingQuestionDetail{
			ReadingQuestionResponse: readingDTO.ReadingQuestionResponse{
				ID:          hit.Source.ID,
				Type:        hit.Source.Type,
				Topic:       hit.Source.Topic,
				Instruction: hit.Source.Instruction,
				Title:       hit.Source.Title,
				Passages:    hit.Source.Passages,
				ImageURLs:   hit.Source.ImageURLs,
				MaxTime:     hit.Source.MaxTime,
				Version:     hit.Source.Version,
			},
		}

		// Parse additional fields based on question type
		switch hit.Source.Type {
		case "TRUE_FALSE":
			if hit.Source.TrueFalse != "" && hit.Source.TrueFalse != "null" {
				var trueFalse []readingDTO.ReadingTrueFalseResponse
				if err := json.Unmarshal([]byte(hit.Source.TrueFalse), &trueFalse); err == nil {
					question.TrueFalse = trueFalse
				}
			}
		case "CHOICE_MULTI":
			if hit.Source.ChoiceMultiQuestion != "" && hit.Source.ChoiceMultiQuestion != "null" {
				var choiceMultiQuestion readingDTO.ReadingChoiceMultiQuestionResponse
				if err := json.Unmarshal([]byte(hit.Source.ChoiceMultiQuestion), &choiceMultiQuestion); err == nil {
					question.ChoiceMultiQuestion = &choiceMultiQuestion
				}
			}
			if hit.Source.ChoiceMultiOptions != "" && hit.Source.ChoiceMultiOptions != "null" {
				var choiceMultiOptions []readingDTO.ReadingChoiceMultiOptionResponse
				if err := json.Unmarshal([]byte(hit.Source.ChoiceMultiOptions), &choiceMultiOptions); err == nil {
					question.ChoiceMultiOptions = choiceMultiOptions
				}
			}
		case "CHOICE_ONE":
			if hit.Source.ChoiceOneQuestion != "" && hit.Source.ChoiceOneQuestion != "null" {
				var choiceOneQuestion readingDTO.ReadingChoiceOneQuestionResponse
				if err := json.Unmarshal([]byte(hit.Source.ChoiceOneQuestion), &choiceOneQuestion); err == nil {
					question.ChoiceOneQuestion = &choiceOneQuestion
				}
			}
			if hit.Source.ChoiceOneOptions != "" && hit.Source.ChoiceOneOptions != "null" {
				var choiceOneOptions []readingDTO.ReadingChoiceOneOptionResponse
				if err := json.Unmarshal([]byte(hit.Source.ChoiceOneOptions), &choiceOneOptions); err == nil {
					question.ChoiceOneOptions = choiceOneOptions
				}
			}
		case "FILL_IN_THE_BLANK":
			if hit.Source.FillInTheBlankQuestion != "" && hit.Source.FillInTheBlankQuestion != "null" {
				var fillInBlankQuestion readingDTO.ReadingFillInTheBlankQuestionResponse
				if err := json.Unmarshal([]byte(hit.Source.FillInTheBlankQuestion), &fillInBlankQuestion); err == nil {
					question.FillInTheBlankQuestion = &fillInBlankQuestion
				}
			}
			if hit.Source.FillInTheBlankAnswers != "" && hit.Source.FillInTheBlankAnswers != "null" {
				var fillInBlankAnswers []readingDTO.ReadingFillInTheBlankAnswerResponse
				if err := json.Unmarshal([]byte(hit.Source.FillInTheBlankAnswers), &fillInBlankAnswers); err == nil {
					question.FillInTheBlankAnswers = fillInBlankAnswers
				}
			}
		}

		questions = append(questions, question)
	}

	result := &readingDTO.ListReadingQuestionsPagination{
		Questions: questions,
		Total:     searchResult.Hits.Total.Value,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
	}

	return result, nil
}

func (s *ReadingQuestionService) DeleteAllQuestions(ctx context.Context) error {
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

	// Delete all records from reading_questions
	if err := tx.Exec("DELETE FROM reading_questions").Error; err != nil {
		tx.Rollback()
		s.logger.Error("reading_question_service.delete_all", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete all reading questions")
		return err
	}

	// Delete all Redis cache with pattern reading_question:*
	if err := s.redis.GetCache().DeletePattern(ctx, "reading_question:*"); err != nil {
		tx.Rollback()
		s.logger.Error("reading_question_service.delete_all.cache", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to delete Redis cache")
		return err
	}

	// Delete OpenSearch index
	if err := s.search.RemoveReadingQuestionsIndex(ctx); err != nil {
		tx.Rollback()
		s.logger.Error("reading_question_service.delete_all.search", map[string]interface{}{
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

func (s *ReadingQuestionService) GetNewUpdatedQuestions(ctx context.Context, versionChecks []struct {
	ID      uuid.UUID
	Version int
}) ([]*reading.ReadingQuestion, error) {
	questionsToRetrieve := make(map[uuid.UUID]int)

	// Check both complete and uncomplete cache keys
	for _, check := range versionChecks {
		completeKey := fmt.Sprintf("reading_question:%s:complete:%d", check.ID, check.Version)
		uncompleteKey := fmt.Sprintf("reading_question:%s:uncomplete:%d", check.ID, check.Version)

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
		return []*reading.ReadingQuestion{}, nil
	}

	return s.repo.GetNewUpdatedQuestions(ctx, versionChecks)
}

func (s *ReadingQuestionService) GetReadingByListID(ctx context.Context, ids []uuid.UUID) ([]*readingDTO.ReadingQuestionDetail, error) {
	// Track which IDs need to be fetched from database
	missingIDs := make([]uuid.UUID, 0)
	result := make([]*readingDTO.ReadingQuestionDetail, 0, len(ids))

	// Try to get from cache first
	for _, id := range ids {
		// Try both complete and uncomplete patterns
		pattern := fmt.Sprintf("reading_question:%s:*", id)
		keys, err := s.redis.GetCache().Keys(ctx, pattern)
		if err == nil && len(keys) > 0 {
			// Found in cache, try to get the data
			cachedData, err := s.redis.GetCache().Get(ctx, keys[0])
			if err == nil {
				var questionDetail readingDTO.ReadingQuestionDetail
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

		var dbQuestions []*reading.ReadingQuestion
		query := s.repo.GetDB().WithContext(ctx).
			Where(strings.Join(conditions, " OR "), values...).
			Find(&dbQuestions)

		if query.Error != nil {
			s.logger.Error("reading_question_service.get_by_list_id", map[string]interface{}{
				"error": query.Error.Error(),
			}, "Failed to get questions from database")
			return nil, query.Error
		}

		// Process and cache the database results
		for _, q := range dbQuestions {
			// Build complete response with details
			questionDetail, err := s.getReadingQuestionDetail(ctx, q)
			if err != nil {
				s.logger.Error("reading_question_service.get_by_list_id.build_response", map[string]interface{}{
					"error": err.Error(),
					"id":    q.ID,
				}, "Failed to build complete response")
				continue
			}

			// Determine completion status
			isComplete := s.completion.IsQuestionComplete(questionDetail)

			// Cache with status and version
			if err := s.redis.SetCacheReadingQuestionDetail(ctx, questionDetail, isComplete); err != nil {
				s.logger.Error("reading_question_service.get_by_list_id.cache", map[string]interface{}{
					"error": err.Error(),
					"id":    q.ID,
				}, "Failed to cache question")
			}

			result = append(result, questionDetail)
		}
	}

	return result, nil
}

func (s *ReadingQuestionService) UpdateQuestion(ctx context.Context, id uuid.UUID, update readingDTO.UpdateReadingQuestionFieldRequest) error {
	// First get the base question from database
	baseQuestion, err := s.repo.GetReadingQuestionByID(ctx, id)
	if err != nil {
		if errors.Is(err, ReadingRepository.ErrQuestionNotFound) {
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
	if err := s.repo.UpdateReadingQuestion(ctx, baseQuestion); err != nil {
		s.logger.Error("Reading_question_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    baseQuestion.ID,
		}, "Failed to update question in database")
		return fmt.Errorf("failed to update question in database: %w", err)
	}

	// Use the updator to update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, baseQuestion); err != nil {
		s.logger.Error("Reading_question_service.update.cache_and_search", map[string]interface{}{
			"error": err.Error(),
			"id":    baseQuestion.ID,
		}, "Failed to update cache and search")
		// Continue even if cache/search update fails
	}

	return nil
}
