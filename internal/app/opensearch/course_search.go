package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	courseDTO "fluencybe/internal/app/dto"
	"fluencybe/pkg/logger"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type CourseSearch struct {
	client *opensearch.Client
	logger *logger.PrettyLogger
}

func NewCourseSearch(client *opensearch.Client, logger *logger.PrettyLogger) *CourseSearch {
	return &CourseSearch{
		client: client,
		logger: logger,
	}
}

func (s *CourseSearch) GetClient() *opensearch.Client {
	return s.client
}

func (s *CourseSearch) CreateCoursesIndex(ctx context.Context) error {
	createReq := opensearchapi.IndicesCreateRequest{
		Index: "courses",
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
                        "type": "keyword"
                    },
                    "title": {
                        "type": "text",
                        "analyzer": "case_insensitive",
                        "fields": {
                            "keyword": {
                                "type": "keyword",
                                "normalizer": "case_insensitive"
                            }
                        }
                    },
                    "overview": {
                        "type": "text",
                        "analyzer": "case_insensitive"
                    },
                    "skills": {
                        "type": "keyword"
                    },
                    "band": {
                        "type": "keyword"
                    },
                    "image_urls": {
                        "type": "keyword"
                    },
                    "course_book": {
                        "properties": {
                            "publishers": {
                                "type": "keyword"
                            },
                            "authors": {
                                "type": "keyword"
                            },
                            "publication_year": {
                                "type": "integer"
                            }
                        }
                    },
                    "course_other": {
                        "type": "object"
                    },
                    "lessons": {
                        "type": "nested",
                        "properties": {
                            "id": {
                                "type": "keyword"
                            },
                            "sequence": {
                                "type": "integer"
                            },
                            "title": {
                                "type": "text",
                                "analyzer": "case_insensitive"
                            },
                            "overview": {
                                "type": "text",
                                "analyzer": "case_insensitive"
                            },
                            "lesson_questions": {
                                "type": "nested",
                                "properties": {
                                    "id": {
                                        "type": "keyword"
                                    },
                                    "sequence": {
                                        "type": "integer"
                                    },
                                    "question_id": {
                                        "type": "keyword"
                                    },
                                    "question_type": {
                                        "type": "keyword"
                                    }
                                }
                            }
                        }
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

func (s *CourseSearch) RemoveCoursesIndex(ctx context.Context) error {
	deleteReq := opensearchapi.IndicesDeleteRequest{
		Index: []string{"courses"},
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

func (s *CourseSearch) DeleteCourseFromIndex(ctx context.Context, id uuid.UUID) error {
	req := opensearchapi.DeleteRequest{
		Index:      "courses",
		DocumentID: id.String(),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to delete course: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error deleting course: %s", res.String())
	}
	return nil
}

func (s *CourseSearch) UpsertCourse(ctx context.Context, course *courseDTO.CourseDetail) error {
	// Remove index if it exists
	if err := s.RemoveCoursesIndex(ctx); err != nil {
		s.logger.Warning("upsert_course.remove_index", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to remove existing index")
	}

	// Create new index
	if err := s.CreateCoursesIndex(ctx); err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}

	// Marshal the course to JSON
	docJSON, err := json.Marshal(course)
	if err != nil {
		return fmt.Errorf("error marshaling document: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      "courses",
		DocumentID: course.ID.String(),
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

func (s *CourseSearch) SearchCourses(ctx context.Context, filter courseDTO.CourseSearchRequest) (*courseDTO.ListCoursePagination, error) {
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
				"term": map[string]interface{}{
					"type": filter.Type,
				},
			},
		)
	}

	// Add title filter if provided
	if filter.Title != "" {
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"match": map[string]interface{}{
					"title": filter.Title,
				},
			},
		)
	}

	// Add skills filter if provided
	if filter.Skills != "" {
		skills := strings.Split(filter.Skills, ",")
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"terms": map[string]interface{}{
					"skills": skills,
				},
			},
		)
	}

	// Add band filter if provided
	if filter.Band != "" {
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"term": map[string]interface{}{
					"band": filter.Band,
				},
			},
		)
	}

	searchJSON, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search body: %w", err)
	}

	searchReq := opensearchapi.SearchRequest{
		Index: []string{"courses"},
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
				Source courseDTO.CourseDetail `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(searchRes.Body).Decode(&rawSearchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	courses := make([]courseDTO.CourseDetail, len(rawSearchResult.Hits.Hits))
	for i, hit := range rawSearchResult.Hits.Hits {
		courses[i] = hit.Source
	}

	return &courseDTO.ListCoursePagination{
		Courses:  courses,
		Total:    rawSearchResult.Hits.Total.Value,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}
