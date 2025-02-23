package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	writingDTO "fluencybe/internal/app/dto"
	"fluencybe/pkg/logger"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type WritingQuestionSearch struct {
	client *opensearch.Client
	logger *logger.PrettyLogger
}

func NewWritingQuestionSearch(client *opensearch.Client, logger *logger.PrettyLogger) *WritingQuestionSearch {
	return &WritingQuestionSearch{
		client: client,
		logger: logger,
	}
}

func (s *WritingQuestionSearch) GetClient() *opensearch.Client {
	return s.client
}

func (s *WritingQuestionSearch) CreateWritingQuestionsIndex(ctx context.Context) error {
	createReq := opensearchapi.IndicesCreateRequest{
		Index: "writing_questions",
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
                }
            },
            "mappings": {
                "properties": {
                    "id": {
                        "type": "keyword"
                    },
                    "type": {
                        "type": "text",
                        "fields": {
                            "keyword": {
                                "type": "keyword",
                                "ignore_above": 256
                            }
                        }
                    },
                    "topic": {
                        "type": "text",
                        "analyzer": "case_insensitive",
                        "fields": {
                            "keyword": {
                                "type": "keyword",
                                "normalizer": "case_insensitive"
                            }
                        }
                    },
                    "instruction": {
                        "type": "text",
                        "analyzer": "case_insensitive"
                    },
                    "image_urls": {
                        "type": "keyword"
                    },
                    "max_time": {
                        "type": "integer"
                    },
                    "status": {
                        "type": "keyword"
                    },
                    "version": {
                        "type": "integer"
                    },
                    "sentence_completion": {
                        "type": "text",
                        "analyzer": "case_insensitive"
                    },
                    "essay": {
                        "type": "text",
                        "analyzer": "case_insensitive"
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

func (s *WritingQuestionSearch) RemoveWritingQuestionsIndex(ctx context.Context) error {
	deleteReq := opensearchapi.IndicesDeleteRequest{
		Index: []string{"writing_questions"},
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

func (s *WritingQuestionSearch) DeleteWritingQuestionFromIndex(ctx context.Context, id uuid.UUID) error {
	req := opensearchapi.DeleteRequest{
		Index:      "writing_questions",
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

func (s *WritingQuestionSearch) UpsertWritingQuestion(ctx context.Context, question *writingDTO.WritingQuestionDetail, status string) error {
	// Check if index exists first
	exists := opensearchapi.IndicesExistsRequest{
		Index: []string{"writing_questions"},
	}

	res, err := exists.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	// Create index only if it doesn't exist
	if res.StatusCode == 404 {
		if err := s.CreateWritingQuestionsIndex(ctx); err != nil {
			return fmt.Errorf("error creating index: %w", err)
		}

		if err := s.UpdateWritingQuestionsMapping(ctx); err != nil {
			return fmt.Errorf("error updating mapping: %w", err)
		}
	}

	doc := map[string]interface{}{
		"id":                  question.ID,
		"type":                question.Type,
		"topic":               question.Topic,
		"instruction":         question.Instruction,
		"image_urls":          question.ImageURLs,
		"max_time":            question.MaxTime,
		"status":              status,
		"version":             question.Version,
		"sentence_completion": marshalWritingQuestionToString(question.SentenceCompletion),
		"essay":               marshalWritingQuestionToString(question.Essay),
	}

	docJSON, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("error marshaling document: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      "writing_questions",
		DocumentID: question.ID.String(),
		Body:       bytes.NewReader(docJSON),
	}

	res, err = req.Do(ctx, s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}

	return nil
}

func marshalWritingQuestionToString(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func (s *WritingQuestionSearch) SearchQuestions(ctx context.Context, filter writingDTO.WritingQuestionSearchFilter) (*writingDTO.ListWritingQuestionsPagination, error) {
	from := (filter.Page - 1) * filter.PageSize

	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{},
			},
		},
		"from": from,
		"size": filter.PageSize,
	}

	if filter.Type != "" {
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"match": map[string]interface{}{
					"type": filter.Type,
				},
			},
		)
	}

	if filter.Topic != "" {
		topics := strings.Split(filter.Topic, ",")
		topicTerms := make([]interface{}, len(topics))
		for i, topic := range topics {
			topicTerms[i] = strings.TrimSpace(topic)
		}
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"terms": map[string]interface{}{
					"topic.keyword": topicTerms,
				},
			},
		)
	}

	searchJSON, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search body: %w", err)
	}

	searchReq := opensearchapi.SearchRequest{
		Index: []string{"writing_questions"},
		Body:  bytes.NewReader(searchJSON),
	}

	searchRes, err := searchReq.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer searchRes.Body.Close()

	var rawSearchResult struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(searchRes.Body).Decode(&rawSearchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	var questions []writingDTO.WritingQuestionDetail
	for _, hit := range rawSearchResult.Hits.Hits {
		var question writingDTO.WritingQuestionDetail

		question.ID, _ = uuid.Parse(hit.Source["id"].(string))
		question.Type = hit.Source["type"].(string)
		if topicRaw, ok := hit.Source["topic"].([]interface{}); ok {
			question.Topic = make([]string, len(topicRaw))
			for i, t := range topicRaw {
				question.Topic[i] = t.(string)
			}
		}
		question.Instruction = hit.Source["instruction"].(string)
		if imgURLs, ok := hit.Source["image_urls"].([]interface{}); ok {
			question.ImageURLs = make([]string, len(imgURLs))
			for i, url := range imgURLs {
				question.ImageURLs[i] = url.(string)
			}
		}
		if maxTime, ok := hit.Source["max_time"].(float64); ok {
			question.MaxTime = int(maxTime)
		}
		if version, ok := hit.Source["version"].(float64); ok {
			question.Version = int(version)
		}

		switch question.Type {
		case "SENTENCE_COMPLETION":
			if sentenceStr, ok := hit.Source["sentence_completion"].(string); ok && sentenceStr != "" {
				var sentence []writingDTO.WritingSentenceCompletionResponse
				if err := json.Unmarshal([]byte(sentenceStr), &sentence); err == nil {
					question.SentenceCompletion = sentence
				}
			}
		case "ESSAY":
			if essayStr, ok := hit.Source["essay"].(string); ok && essayStr != "" {
				var essay []writingDTO.WritingEssayResponse
				if err := json.Unmarshal([]byte(essayStr), &essay); err == nil {
					question.Essay = essay
				}
			}
		}

		questions = append(questions, question)
	}

	return &writingDTO.ListWritingQuestionsPagination{
		Questions: questions,
		Total:     rawSearchResult.Hits.Total.Value,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
	}, nil
}

func (s *WritingQuestionSearch) UpdateWritingQuestionsMapping(ctx context.Context) error {
	putMappingReq := opensearchapi.IndicesPutMappingRequest{
		Index: []string{"writing_questions"},
		Body: strings.NewReader(`{
            "properties": {
                "id": {
                    "type": "keyword"
                },
                "type": {
                    "type": "text",
                    "fields": {
                        "keyword": {
                            "type": "keyword",
                            "ignore_above": 256
                        }
                    }
                },
                "topic": {
                    "type": "text",
                    "analyzer": "case_insensitive",
                    "fields": {
                        "keyword": {
                            "type": "keyword",
                            "normalizer": "case_insensitive"
                        }
                    }
                },
                "instruction": {
                    "type": "text",
                    "analyzer": "case_insensitive"
                },
                "image_urls": {
                    "type": "keyword"
                },
                "max_time": {
                    "type": "integer"
                },
                "status": {
                    "type": "keyword"
                },
                "version": {
                    "type": "integer"
                },
                "sentence_completion": {
                    "type": "text",
                    "analyzer": "case_insensitive"
                },
                "essay": {
                    "type": "text",
                    "analyzer": "case_insensitive"
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
