package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	grammarDTO "fluencybe/internal/app/dto"
	"fluencybe/pkg/logger"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type GrammarSearchResult struct {
	Questions []grammarDTO.GrammarQuestionDetail
	Total     int64
	Page      int
	PageSize  int
}

type GrammarQuestionSearch struct {
	client *opensearch.Client
	logger *logger.PrettyLogger
}

func NewGrammarQuestionSearch(client *opensearch.Client, logger *logger.PrettyLogger) *GrammarQuestionSearch {
	return &GrammarQuestionSearch{
		client: client,
		logger: logger,
	}
}

func (s *GrammarQuestionSearch) GetClient() *opensearch.Client {
	return s.client
}

func (s *GrammarQuestionSearch) CreateGrammarQuestionsIndex(ctx context.Context) error {
	createReq := opensearchapi.IndicesCreateRequest{
		Index: "grammar_questions",
		Body: strings.NewReader(`{
            "settings": {
                "analysis": {
                    "analyzer": {
                        "case_insensitive": {
                            "type": "custom",
                            "tokenizer": "standard",
                            "filter": ["lowercase"]
                        }
                    },
                    "normalizer": {
                        "case_insensitive": {
                            "type": "custom",
                            "filter": ["lowercase"]
                        }
                    }
                },
                "number_of_shards": 1,
                "number_of_replicas": 1
            },
            "mappings": {
                "properties": {
                    "choice_one_options": {
                        "analyzer": "case_insensitive",
                        "type": "text"
                    },
                    "choice_one_question": {
                        "analyzer": "case_insensitive", 
                        "type": "text"
                    },
                    "error_identification": {
                        "analyzer": "case_insensitive",
                        "type": "text"
                    },
                    "fill_in_the_blank_answers": {
                        "analyzer": "case_insensitive",
                        "type": "text"
                    },
                    "fill_in_the_blank_question": {
                        "analyzer": "case_insensitive",
                        "type": "text"
                    },
                    "id": {
                        "type": "keyword"
                    },
                    "image_urls": {
                        "type": "keyword"
                    },
                    "instruction": {
                        "analyzer": "case_insensitive",
                        "type": "text"
                    },
                    "max_time": {
                        "type": "integer"
                    },
                    "sentence_transformation": {
                        "analyzer": "case_insensitive",
                        "type": "text"
                    },
                    "status": {
                        "type": "keyword"
                    },
                    "topic": {
                        "fields": {
                            "keyword": {
                                "type": "keyword",
                                "normalizer": "case_insensitive"
                            }
                        },
                        "analyzer": "case_insensitive",
                        "type": "text"
                    },
                    "type": {
                        "fields": {
                            "keyword": {
                                "type": "keyword",
                                "ignore_above": 256
                            }
                        },
                        "type": "text"
                    },
                    "version": {
                        "type": "integer"
                    }
                }
            }
        }`),
	}

	res, err := createReq.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error creating index: %s", res.String())
	}

	return nil
}

func (s *GrammarQuestionSearch) RemoveGrammarQuestionsIndex(ctx context.Context) error {
	deleteReq := opensearchapi.IndicesDeleteRequest{
		Index: []string{"grammar_questions"},
	}

	res, err := deleteReq.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("error deleting index: %s", res.String())
	}

	return nil
}

func (s *GrammarQuestionSearch) DeleteGrammarQuestionFromIndex(ctx context.Context, id uuid.UUID) error {
	req := opensearchapi.DeleteRequest{
		Index:      "grammar_questions",
		DocumentID: id.String(),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to delete question: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error deleting question: %s", res.String())
	}
	return nil
}

func (s *GrammarQuestionSearch) UpdateGrammarQuestionsMapping(ctx context.Context) error {
	putMappingReq := opensearchapi.IndicesPutMappingRequest{
		Index: []string{"grammar_questions"},
		Body: strings.NewReader(`{
			"properties": {
				"choice_one_options": {
					"analyzer": "case_insensitive",
					"type": "text"
				},
				"choice_one_question": {
					"analyzer": "case_insensitive",
					"type": "text"
				},
				"error_identification": {
					"analyzer": "case_insensitive",
					"type": "text"
				},
				"fill_in_the_blank_answers": {
					"analyzer": "case_insensitive",
					"type": "text"
				},
				"fill_in_the_blank_question": {
					"analyzer": "case_insensitive",
					"type": "text"
				},
				"id": {
					"type": "keyword"
				},
				"image_urls": {
					"type": "keyword"
				},
				"instruction": {
					"analyzer": "case_insensitive",
					"type": "text"
				},
				"max_time": {
					"type": "integer"
				},
				"sentence_transformation": {
					"analyzer": "case_insensitive",
					"type": "text"
				},
				"status": {
					"type": "keyword"
				},
				"topic": {
					"fields": {
						"keyword": {
							"type": "keyword",
							"normalizer": "case_insensitive"
						}
					},
					"analyzer": "case_insensitive",
					"type": "text"
				},
				"type": {
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					},
					"type": "text"
				},
				"version": {
					"type": "integer"
				}
			}
		}`),
	}

	res, err := putMappingReq.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error updating mapping: %s", res.String())
	}

	return nil
}

func (s *GrammarQuestionSearch) UpsertGrammarQuestion(ctx context.Context, question *grammarDTO.GrammarQuestionDetail, status string) error {
	// First check if index exists
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{"grammar_questions"},
	}

	existsRes, err := existsReq.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	// If index doesn't exist, create it
	if existsRes.StatusCode == 404 {
		if err := s.CreateGrammarQuestionsIndex(ctx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Add status and version to the document
	doc := map[string]interface{}{
		"id":                         question.ID,
		"type":                       question.Type,
		"topic":                      question.Topic,
		"instruction":                question.Instruction,
		"image_urls":                 question.ImageURLs,
		"max_time":                   question.MaxTime,
		"status":                     status,
		"version":                    question.Version,
		"fill_in_the_blank_question": ConvertGrammarQuestionToJSON(question.FillInTheBlankQuestion),
		"fill_in_the_blank_answers":  ConvertGrammarQuestionToJSON(question.FillInTheBlankAnswers),
		"choice_one_question":        ConvertGrammarQuestionToJSON(question.ChoiceOneQuestion),
		"choice_one_options":         ConvertGrammarQuestionToJSON(question.ChoiceOneOptions),
		"error_identification":       ConvertGrammarQuestionToJSON(question.ErrorIdentification),
		"sentence_transformation":    ConvertGrammarQuestionToJSON(question.SentenceTransformation),
	}

	// Add debug logging for document
	s.logger.Debug("index_document", map[string]interface{}{
		"id":       question.ID,
		"max_time": question.MaxTime,
		"type":     question.Type,
	}, "Indexing document with max_time")

	// Marshal the doc map to JSON
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("error marshaling document: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      "grammar_questions",
		DocumentID: question.ID.String(),
		Body:       bytes.NewReader(docJSON),
	}

	indexRes, err := req.Do(ctx, s.client)
	if err != nil {
		return err
	}
	defer indexRes.Body.Close()

	if indexRes.IsError() {
		return fmt.Errorf("error indexing document: %s", indexRes.String())
	}

	return nil
}

func ConvertGrammarQuestionToJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func (s *GrammarQuestionSearch) SearchQuestions(ctx context.Context, filter grammarDTO.GrammarQuestionSearchFilter) (*grammarDTO.ListGrammarQuestionsPagination, error) {
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
		case "CHOICE_ONE":
			metadataQuery = map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query": filter.Metadata,
					"fields": []string{
						"choice_one_question",
						"choice_one_options",
					},
				},
			}
		case "ERROR_IDENTIFICATION":
			metadataQuery = map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query": filter.Metadata,
					"fields": []string{
						"error_identification",
					},
				},
			}
		case "SENTENCE_TRANSFORMATION":
			metadataQuery = map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query": filter.Metadata,
					"fields": []string{
						"sentence_transformation",
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

	searchJSON, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search body: %w", err)
	}

	// Execute search
	searchReq := opensearchapi.SearchRequest{
		Index: []string{"grammar_questions"},
		Body:  bytes.NewReader(searchJSON),
	}

	searchRes, err := searchReq.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer searchRes.Body.Close()

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
					ImageURLs              []string  `json:"image_urls"`
					MaxTime                int       `json:"max_time"`
					Version                int       `json:"version"`
					Status                 string    `json:"status"`
					ErrorIdentification    string    `json:"error_identification,omitempty"`
					ChoiceOneQuestion      string    `json:"choice_one_question,omitempty"`
					ChoiceOneOptions       string    `json:"choice_one_options,omitempty"`
					FillInTheBlankQuestion string    `json:"fill_in_the_blank_question,omitempty"`
					FillInTheBlankAnswers  string    `json:"fill_in_the_blank_answers,omitempty"`
					SentenceTransformation string    `json:"sentence_transformation,omitempty"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(searchRes.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	questions := make([]grammarDTO.GrammarQuestionDetail, len(searchResult.Hits.Hits))
	for i, hit := range searchResult.Hits.Hits {
		// Luôn tạo câu hỏi với thông tin cơ bản
		question := grammarDTO.GrammarQuestionDetail{
			GrammarQuestionResponse: grammarDTO.GrammarQuestionResponse{
				ID:          hit.Source.ID,
				Type:        hit.Source.Type,
				Topic:       hit.Source.Topic,
				Instruction: hit.Source.Instruction,
				ImageURLs:   hit.Source.ImageURLs,
				MaxTime:     hit.Source.MaxTime,
				Version:     hit.Source.Version,
			},
		}

		// Parse additional fields based on question type
		switch hit.Source.Type {
		case "FILL_IN_THE_BLANK":
			if hit.Source.FillInTheBlankQuestion != "" {
				var fillInBlankQuestion grammarDTO.GrammarFillInTheBlankQuestionResponse
				if err := json.Unmarshal([]byte(hit.Source.FillInTheBlankQuestion), &fillInBlankQuestion); err == nil {
					// Chỉ gán nếu có Question
					if fillInBlankQuestion.Question != "" {
						question.FillInTheBlankQuestion = &fillInBlankQuestion
					}
				}
			}
			if hit.Source.FillInTheBlankAnswers != "" {
				var fillInBlankAnswers []grammarDTO.GrammarFillInTheBlankAnswerResponse
				if err := json.Unmarshal([]byte(hit.Source.FillInTheBlankAnswers), &fillInBlankAnswers); err == nil {
					// Chỉ gán nếu có answers
					if len(fillInBlankAnswers) > 0 {
						question.FillInTheBlankAnswers = fillInBlankAnswers
					}
				}
			}
		case "CHOICE_ONE":
			if hit.Source.ChoiceOneQuestion != "" {
				var choiceOneQuestion grammarDTO.GrammarChoiceOneQuestionResponse
				if err := json.Unmarshal([]byte(hit.Source.ChoiceOneQuestion), &choiceOneQuestion); err == nil {
					// Chỉ gán nếu có Question
					if choiceOneQuestion.Question != "" {
						question.ChoiceOneQuestion = &choiceOneQuestion
					}
				}
			}
			if hit.Source.ChoiceOneOptions != "" {
				var choiceOneOptions []grammarDTO.GrammarChoiceOneOptionResponse
				if err := json.Unmarshal([]byte(hit.Source.ChoiceOneOptions), &choiceOneOptions); err == nil {
					// Chỉ gán nếu có options
					if len(choiceOneOptions) > 0 {
						question.ChoiceOneOptions = choiceOneOptions
					}
				}
			}
		case "ERROR_IDENTIFICATION":
			if hit.Source.ErrorIdentification != "" {
				var errorIdentification grammarDTO.GrammarErrorIdentificationResponse
				if err := json.Unmarshal([]byte(hit.Source.ErrorIdentification), &errorIdentification); err == nil {
					// Chỉ gán nếu có ErrorSentence
					if errorIdentification.ErrorSentence != "" {
						question.ErrorIdentification = &errorIdentification
					}
				}
			}
		case "SENTENCE_TRANSFORMATION":
			if hit.Source.SentenceTransformation != "" {
				var sentenceTransformation grammarDTO.GrammarSentenceTransformationResponse
				if err := json.Unmarshal([]byte(hit.Source.SentenceTransformation), &sentenceTransformation); err == nil {
					// Chỉ gán nếu có OriginalSentence
					if sentenceTransformation.OriginalSentence != "" {
						question.SentenceTransformation = &sentenceTransformation
					}
				}
			}
		}

		questions[i] = question
	}

	return &grammarDTO.ListGrammarQuestionsPagination{
		Questions: questions,
		Total:     searchResult.Hits.Total.Value,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
	}, nil
}

// Helper function to check if question has valid data
func hasValidData(question *grammarDTO.GrammarQuestionDetail) bool {
	switch question.Type {
	case "FILL_IN_THE_BLANK":
		return question.FillInTheBlankQuestion != nil && question.FillInTheBlankQuestion.Question != ""
	case "CHOICE_ONE":
		return question.ChoiceOneQuestion != nil && question.ChoiceOneQuestion.Question != ""
	case "ERROR_IDENTIFICATION":
		return question.ErrorIdentification != nil && question.ErrorIdentification.ErrorSentence != ""
	case "SENTENCE_TRANSFORMATION":
		return question.SentenceTransformation != nil && question.SentenceTransformation.OriginalSentence != ""
	default:
		return false
	}
}
