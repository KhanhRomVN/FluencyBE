package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	listeningDTO "fluencybe/internal/app/dto"
	"fluencybe/pkg/logger"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type ListeningSearchResult struct {
	Questions []listeningDTO.ListeningQuestionDetail
	Total     int64
	Page      int
	PageSize  int
}

type ListeningQuestionSearch struct {
	client *opensearch.Client
	logger *logger.PrettyLogger
}

func NewListeningQuestionSearch(client *opensearch.Client, logger *logger.PrettyLogger) *ListeningQuestionSearch {
	return &ListeningQuestionSearch{
		client: client,
		logger: logger,
	}
}

func (s *ListeningQuestionSearch) GetClient() *opensearch.Client {
	return s.client
}

func (s *ListeningQuestionSearch) CreateListeningQuestionsIndex(ctx context.Context) error {
	createReq := opensearchapi.IndicesCreateRequest{
		Index: "listening_questions",
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

func (s *ListeningQuestionSearch) RemoveListeningQuestionsIndex(ctx context.Context) error {
	deleteReq := opensearchapi.IndicesDeleteRequest{
		Index: []string{"listening_questions"},
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

func (s *ListeningQuestionSearch) DeleteListeningQuestionFromIndex(ctx context.Context, id uuid.UUID) error {
	req := opensearchapi.DeleteRequest{
		Index:      "listening_questions",
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

func (s *ListeningQuestionSearch) UpdateListeningQuestionsMapping(ctx context.Context) error {
	putMappingReq := opensearchapi.IndicesPutMappingRequest{
		Index: []string{"listening_questions"},
		Body: strings.NewReader(`{
			"properties": {
				"id": { "type": "keyword" },
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
				"audio_urls": { "type": "keyword" },
				"image_urls": { "type": "keyword" },
				"transcript": {
					"type": "text",
					"analyzer": "case_insensitive"
				},
				"max_time": { "type": "integer" },
				"status": { "type": "keyword" },
				"version": { "type": "integer" },
				"fill_in_the_blank_question": { "type": "text", "analyzer": "case_insensitive" },
				"fill_in_the_blank_answers": { "type": "text", "analyzer": "case_insensitive" },
				"choice_one_question": { "type": "text", "analyzer": "case_insensitive" },
				"choice_one_options": { "type": "text", "analyzer": "case_insensitive" },
				"choice_multi_question": { "type": "text", "analyzer": "case_insensitive" },
				"choice_multi_options": { "type": "text", "analyzer": "case_insensitive" },
				"map_labelling": { "type": "text", "analyzer": "case_insensitive" },
				"MATCHING": { "type": "text", "analyzer": "case_insensitive" }
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

func (s *ListeningQuestionSearch) UpsertListeningQuestion(ctx context.Context, question *listeningDTO.ListeningQuestionDetail, status string) error {
	// Check if index exists
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{"listening_questions"},
	}
	existsRes, err := existsReq.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	// If index doesn't exist, create it
	if existsRes.StatusCode == 404 {
		if err := s.CreateListeningQuestionsIndex(ctx); err != nil {
			return fmt.Errorf("error creating index: %w", err)
		}
		// Update mapping after creating index
		if err := s.UpdateListeningQuestionsMapping(ctx); err != nil {
			return fmt.Errorf("error updating mapping: %w", err)
		}
	}

	// Add status and version to the document
	doc := map[string]interface{}{
		"id":                         question.ID,
		"type":                       question.Type,
		"topic":                      question.Topic,
		"instruction":                question.Instruction,
		"audio_urls":                 question.AudioURLs,
		"image_urls":                 question.ImageURLs,
		"transcript":                 question.Transcript,
		"max_time":                   question.MaxTime, // Explicitly include max_time
		"status":                     status,
		"version":                    question.Version,
		"fill_in_the_blank_question": ConvertListeningQuestionToJSON(question.FillInTheBlankQuestion),
		"fill_in_the_blank_answers":  ConvertListeningQuestionToJSON(question.FillInTheBlankAnswers),
		"choice_one_question":        ConvertListeningQuestionToJSON(question.ChoiceOneQuestion),
		"choice_one_options":         ConvertListeningQuestionToJSON(question.ChoiceOneOptions),
		"choice_multi_question":      ConvertListeningQuestionToJSON(question.ChoiceMultiQuestion),
		"choice_multi_options":       ConvertListeningQuestionToJSON(question.ChoiceMultiOptions),
		"map_labelling":              ConvertListeningQuestionToJSON(question.MapLabelling),
		"MATCHING":                   ConvertListeningQuestionToJSON(question.Matching),
	}

	// Add debug logging for document
	s.logger.Debug("index_document", map[string]interface{}{
		"id":       question.ID,
		"max_time": question.MaxTime,
		"type":     question.Type,
	}, "Indexing document with max_time")

	// Update the index mapping to include version field
	if err := s.UpdateListeningQuestionsMapping(ctx); err != nil {
		return fmt.Errorf("error updating mapping: %w", err)
	}

	// Marshal the doc map to JSON
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("error marshaling document: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      "listening_questions",
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

func ConvertListeningQuestionToJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
