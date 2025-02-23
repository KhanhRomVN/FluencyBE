package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	speakingDTO "fluencybe/internal/app/dto"
	"fluencybe/pkg/logger"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type SpeakingSearchResult struct {
	Questions []speakingDTO.SpeakingQuestionDetail
	Total     int64
	Page      int
	PageSize  int
}

type SpeakingQuestionSearch struct {
	client *opensearch.Client
	logger *logger.PrettyLogger
}

func NewSpeakingQuestionSearch(client *opensearch.Client, logger *logger.PrettyLogger) *SpeakingQuestionSearch {
	return &SpeakingQuestionSearch{
		client: client,
		logger: logger,
	}
}

func (s *SpeakingQuestionSearch) GetClient() *opensearch.Client {
	return s.client
}

func (s *SpeakingQuestionSearch) CreateSpeakingQuestionsIndex(ctx context.Context) error {
	createReq := opensearchapi.IndicesCreateRequest{
		Index: "speaking_questions",
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
                    "word_repetition": {
                        "type": "text",
                        "analyzer": "case_insensitive"
                    },
                    "phrase_repetition": {
                        "type": "text",
                        "analyzer": "case_insensitive"
                    },
                    "paragraph_repetition": {
                        "type": "text",
                        "analyzer": "case_insensitive"
                    },
                    "open_paragraph": {
                        "type": "text",
                        "analyzer": "case_insensitive"
                    },
                    "conversational_repetition": {
                        "type": "text",
                        "analyzer": "case_insensitive"
                    },
                    "conversational_repetition_qas": {
                        "type": "text",
                        "analyzer": "case_insensitive"
                    },
                    "conversational_open": {
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

func (s *SpeakingQuestionSearch) RemoveSpeakingQuestionsIndex(ctx context.Context) error {
	deleteReq := opensearchapi.IndicesDeleteRequest{
		Index: []string{"speaking_questions"},
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

func (s *SpeakingQuestionSearch) DeleteSpeakingQuestionFromIndex(ctx context.Context, id uuid.UUID) error {
	req := opensearchapi.DeleteRequest{
		Index:      "speaking_questions",
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

func (s *SpeakingQuestionSearch) UpsertSpeakingQuestion(ctx context.Context, question *speakingDTO.SpeakingQuestionDetail, status string) error {
	// Check if index exists
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{"speaking_questions"},
	}
	existsRes, err := existsReq.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	// Create index and mapping only if it doesn't exist
	if existsRes.StatusCode == 404 {
		if err := s.CreateSpeakingQuestionsIndex(ctx); err != nil {
			return fmt.Errorf("error creating index: %w", err)
		}

		// Update the index mapping to include version field
		if err := s.UpdateSpeakingQuestionsMapping(ctx); err != nil {
			return fmt.Errorf("error updating mapping: %w", err)
		}
	}

	// Add status and version to the document
	doc := map[string]interface{}{
		"id":                            question.ID,
		"type":                          question.Type,
		"topic":                         question.Topic,
		"instruction":                   question.Instruction,
		"image_urls":                    question.ImageURLs,
		"max_time":                      question.MaxTime,
		"status":                        status,
		"version":                       question.Version,
		"word_repetition":               marshalSpeakingQuestionToString(question.WordRepetition),
		"phrase_repetition":             marshalSpeakingQuestionToString(question.PhraseRepetition),
		"paragraph_repetition":          marshalSpeakingQuestionToString(question.ParagraphRepetition),
		"open_paragraph":                marshalSpeakingQuestionToString(question.OpenParagraph),
		"conversational_repetition":     marshalSpeakingQuestionToString(question.ConversationalRepetition),
		"conversational_repetition_qas": marshalSpeakingQuestionToString(question.ConversationalRepetitionQAs),
		"conversational_open":           marshalSpeakingQuestionToString(question.ConversationalOpen),
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
		Index:      "speaking_questions",
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

func marshalSpeakingQuestionToString(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func (s *SpeakingQuestionSearch) SearchQuestions(ctx context.Context, filter speakingDTO.SpeakingQuestionSearchFilter) (*speakingDTO.ListSpeakingQuestionsPagination, error) {
	from := (filter.Page - 1) * filter.PageSize

	// Build search query
	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{},
			},
		},
		"from": from,
		"size": filter.PageSize,
	}

	// Add type filter if provided
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

	// Add topic filter if provided
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
		Index: []string{"speaking_questions"},
		Body:  bytes.NewReader(searchJSON),
	}

	searchRes, err := searchReq.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer searchRes.Body.Close()

	// First unmarshal into a raw map to handle the string fields
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

	var questions []speakingDTO.SpeakingQuestionDetail
	for _, hit := range rawSearchResult.Hits.Hits {
		var question speakingDTO.SpeakingQuestionDetail

		// Handle base fields
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

		// Handle specialized fields based on question type
		switch question.Type {
		case "WORD_REPETITION":
			if wordRepStr, ok := hit.Source["word_repetition"].(string); ok && wordRepStr != "" {
				var wordRep []speakingDTO.SpeakingWordRepetitionResponse
				if err := json.Unmarshal([]byte(wordRepStr), &wordRep); err == nil {
					question.WordRepetition = wordRep
				}
			}

		case "PHRASE_REPETITION":
			if phraseRepStr, ok := hit.Source["phrase_repetition"].(string); ok && phraseRepStr != "" {
				var phraseRep []speakingDTO.SpeakingPhraseRepetitionResponse
				if err := json.Unmarshal([]byte(phraseRepStr), &phraseRep); err == nil {
					question.PhraseRepetition = phraseRep
				}
			}

		case "PARAGRAPH_REPETITION":
			if paraRepStr, ok := hit.Source["paragraph_repetition"].(string); ok && paraRepStr != "" {
				var paraRep []speakingDTO.SpeakingParagraphRepetitionResponse
				if err := json.Unmarshal([]byte(paraRepStr), &paraRep); err == nil {
					question.ParagraphRepetition = paraRep
				}
			}

		case "OPEN_PARAGRAPH":
			if openParaStr, ok := hit.Source["open_paragraph"].(string); ok && openParaStr != "" {
				var openPara []speakingDTO.SpeakingOpenParagraphResponse
				if err := json.Unmarshal([]byte(openParaStr), &openPara); err == nil {
					question.OpenParagraph = openPara
				}
			}

		case "CONVERSATIONAL_REPETITION":
			if convRepStr, ok := hit.Source["conversational_repetition"].(string); ok && convRepStr != "" {
				var convRep speakingDTO.SpeakingConversationalRepetitionResponse
				if err := json.Unmarshal([]byte(convRepStr), &convRep); err == nil {
					question.ConversationalRepetition = &convRep
				}
			}
			if convRepQAStr, ok := hit.Source["conversational_repetition_qas"].(string); ok && convRepQAStr != "" {
				var convRepQAs []speakingDTO.SpeakingConversationalRepetitionQAResponse
				if err := json.Unmarshal([]byte(convRepQAStr), &convRepQAs); err == nil {
					question.ConversationalRepetitionQAs = convRepQAs
				}
			}

		case "CONVERSATIONAL_OPEN":
			if convOpenStr, ok := hit.Source["conversational_open"].(string); ok && convOpenStr != "" {
				var convOpen speakingDTO.SpeakingConversationalOpenResponse
				if err := json.Unmarshal([]byte(convOpenStr), &convOpen); err == nil {
					question.ConversationalOpen = &convOpen
				}
			}
		}

		questions = append(questions, question)
	}

	return &speakingDTO.ListSpeakingQuestionsPagination{
		Questions: questions,
		Total:     rawSearchResult.Hits.Total.Value,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
	}, nil
}

func (s *SpeakingQuestionSearch) UpdateSpeakingQuestionsMapping(ctx context.Context) error {
	putMappingReq := opensearchapi.IndicesPutMappingRequest{
		Index: []string{"speaking_questions"},
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
                "word_repetition": {
                    "type": "text",
                    "analyzer": "case_insensitive"
                },
                "phrase_repetition": {
                    "type": "text",
                    "analyzer": "case_insensitive"
                },
                "paragraph_repetition": {
                    "type": "text",
                    "analyzer": "case_insensitive"
                },
                "open_paragraph": {
                    "type": "text",
                    "analyzer": "case_insensitive"
                },
                "conversational_repetition": {
                    "type": "text",
                    "analyzer": "case_insensitive"
                },
                "conversational_repetition_qas": {
                    "type": "text",
                    "analyzer": "case_insensitive"
                },
                "conversational_open": {
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
