package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	readingDTO "fluencybe/internal/app/dto"
	"fluencybe/pkg/logger"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type ReadingSearchResult struct {
	Questions []readingDTO.ReadingQuestionDetail
	Total     int64
	Page      int
	PageSize  int
}

type ReadingQuestionSearch struct {
	client *opensearch.Client
	logger *logger.PrettyLogger
}

func NewReadingQuestionSearch(client *opensearch.Client, logger *logger.PrettyLogger) *ReadingQuestionSearch {
	return &ReadingQuestionSearch{
		client: client,
		logger: logger,
	}
}

func (s *ReadingQuestionSearch) GetClient() *opensearch.Client {
	return s.client
}

func (s *ReadingQuestionSearch) CreateReadingQuestionsIndex(ctx context.Context) error {
	createReq := opensearchapi.IndicesCreateRequest{
		Index: "reading_questions",
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
                            "char_filter": [],
                            "filter": ["lowercase"]
                        }
                    }
                }
            },
            "mappings": {
                "properties": {
                    "id": { "type": "text" },
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
                    "title": {
                        "type": "text",
                        "analyzer": "case_insensitive"
                    },
                    "passages": {
                        "type": "text",
                        "analyzer": "case_insensitive"
                    },
                    "image_urls": { "type": "keyword" },
                    "max_time": { "type": "integer" },
                    "status": { "type": "keyword" },
                    "version": { "type": "integer" }
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
func (s *ReadingQuestionSearch) RemoveReadingQuestionsIndex(ctx context.Context) error {
	deleteReq := opensearchapi.IndicesDeleteRequest{
		Index: []string{"reading_questions"},
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

func (s *ReadingQuestionSearch) DeleteReadingQuestionFromIndex(ctx context.Context, id uuid.UUID) error {
	req := opensearchapi.DeleteRequest{
		Index:      "reading_questions",
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

func (s *ReadingQuestionSearch) SearchQuestions(ctx context.Context, filter readingDTO.ReadingQuestionSearchFilter) (*readingDTO.ListReadingQuestionsPagination, error) {
	from := (filter.Page - 1) * filter.PageSize

	// Build search query
	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":    filter.Query,
							"fields":   []string{"instruction", "title", "passages"},
							"operator": "and",
						},
					},
					{
						"match": map[string]interface{}{
							"topic": map[string]interface{}{
								"query":    filter.Query,
								"operator": "and",
								"analyzer": "case_insensitive",
							},
						},
					},
				},
			},
		},
		"from": from,
		"size": filter.PageSize,
	}

	searchJSON, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search body: %w", err)
	}

	searchReq := opensearchapi.SearchRequest{
		Index: []string{"reading_questions"},
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
				Source readingDTO.ReadingQuestionDetail `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(searchRes.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	var questions []readingDTO.ReadingQuestionDetail
	for _, hit := range searchResult.Hits.Hits {
		question := hit.Source // Copy the base fields directly

		// Handle TrueFalse
		if len(hit.Source.TrueFalse) > 0 {
			question.TrueFalse = hit.Source.TrueFalse
		}

		// Handle FillInTheBlankQuestion
		if hit.Source.FillInTheBlankQuestion != nil {
			question.FillInTheBlankQuestion = hit.Source.FillInTheBlankQuestion
		}

		// Handle FillInTheBlankAnswers
		if len(hit.Source.FillInTheBlankAnswers) > 0 {
			question.FillInTheBlankAnswers = hit.Source.FillInTheBlankAnswers
		}

		// Handle ChoiceOneQuestion
		if hit.Source.ChoiceOneQuestion != nil {
			question.ChoiceOneQuestion = hit.Source.ChoiceOneQuestion
		}

		// Handle ChoiceOneOptions
		if len(hit.Source.ChoiceOneOptions) > 0 {
			question.ChoiceOneOptions = hit.Source.ChoiceOneOptions
		}

		// Handle ChoiceMultiQuestion
		if hit.Source.ChoiceMultiQuestion != nil {
			question.ChoiceMultiQuestion = hit.Source.ChoiceMultiQuestion
		}

		// Handle ChoiceMultiOptions
		if len(hit.Source.ChoiceMultiOptions) > 0 {
			question.ChoiceMultiOptions = hit.Source.ChoiceMultiOptions
		}

		// Handle Matching
		if len(hit.Source.Matching) > 0 {
			question.Matching = hit.Source.Matching
		}

		questions = append(questions, question)
	}

	return &readingDTO.ListReadingQuestionsPagination{
		Questions: questions,
		Total:     searchResult.Hits.Total.Value,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
	}, nil
}

func (s *ReadingQuestionSearch) IndexQuestions(ctx context.Context, questions []readingDTO.ReadingQuestionDetail) error {
	if len(questions) == 0 {
		return nil
	}

	var bulkBuilder strings.Builder
	for _, q := range questions {
		// Create action line
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": "reading_questions",
				"_id":    q.ID.String(),
			},
		}
		actionLine, err := json.Marshal(action)
		if err != nil {
			return fmt.Errorf("failed to marshal action: %w", err)
		}

		// Create document line
		docLine, err := json.Marshal(q)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %w", err)
		}

		// Add both lines to bulk request
		bulkBuilder.Write(actionLine)
		bulkBuilder.WriteString("\n")
		bulkBuilder.Write(docLine)
		bulkBuilder.WriteString("\n")
	}

	// Execute bulk request
	bulkReq := opensearchapi.BulkRequest{
		Body: strings.NewReader(bulkBuilder.String()),
	}
	bulkResp, err := bulkReq.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to bulk index documents: %w", err)
	}
	defer bulkResp.Body.Close()

	return nil
}

func (s *ReadingQuestionSearch) UpdateReadingQuestionsMapping(ctx context.Context) error {
	putMappingReq := opensearchapi.IndicesPutMappingRequest{
		Index: []string{"reading_questions"},
		Body: strings.NewReader(`{
            "properties": {
                "id": { "type": "text" },
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
                "title": {
                    "type": "text",
                    "analyzer": "case_insensitive"
                },
                "passages": {
                    "type": "text",
                    "analyzer": "case_insensitive"
                },
                "image_urls": { "type": "keyword" },
                "max_time": { "type": "integer" },
                "status": { "type": "keyword" },
                "version": { "type": "integer" }
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

func (s *ReadingQuestionSearch) IndexReadingQuestionDetail(ctx context.Context, question *readingDTO.ReadingQuestionDetail, status string) error {
	// Check if index exists
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{"reading_questions"},
	}

	res, err := existsReq.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("error checking index existence: %w", err)
	}

	// If index doesn't exist, create it
	if res.StatusCode == 404 {
		if err := s.CreateReadingQuestionsIndex(ctx); err != nil {
			return fmt.Errorf("error creating index: %w", err)
		}
	} else if res.IsError() {
		return fmt.Errorf("error checking index existence: %s", res.String())
	}

	// Add status and version to the document
	doc := map[string]interface{}{
		"id":                         question.ID,
		"type":                       question.Type,
		"topic":                      question.Topic,
		"instruction":                question.Instruction,
		"title":                      question.Title,
		"passages":                   question.Passages,
		"image_urls":                 question.ImageURLs,
		"max_time":                   question.MaxTime,
		"status":                     status,
		"version":                    question.Version,
		"true_false":                 marshalReadingQuestionToString(question.TrueFalse),
		"fill_in_the_blank_question": marshalReadingQuestionToString(question.FillInTheBlankQuestion),
		"fill_in_the_blank_answers":  marshalReadingQuestionToString(question.FillInTheBlankAnswers),
		"choice_one_question":        marshalReadingQuestionToString(question.ChoiceOneQuestion),
		"choice_one_options":         marshalReadingQuestionToString(question.ChoiceOneOptions),
		"choice_multi_question":      marshalReadingQuestionToString(question.ChoiceMultiQuestion),
		"choice_multi_options":       marshalReadingQuestionToString(question.ChoiceMultiOptions),
		"MATCHING":                   marshalReadingQuestionToString(question.Matching),
	}

	// Add debug logging for document
	s.logger.Debug("index_document", map[string]interface{}{
		"id":       question.ID,
		"max_time": question.MaxTime,
		"type":     question.Type,
	}, "Indexing document with max_time")

	// Update the index mapping to include version field
	if err := s.UpdateReadingQuestionsMapping(ctx); err != nil {
		return fmt.Errorf("error updating mapping: %w", err)
	}

	// Marshal the doc map to JSON
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("error marshaling document: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      "reading_questions",
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

func marshalReadingQuestionToString(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
