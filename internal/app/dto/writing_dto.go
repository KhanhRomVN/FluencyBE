package dto

import (
	"time"

	"github.com/google/uuid"
)

//==============================================================================
// * =-=-=-=-=-=-=-=-=-=-=-=-=-= Writing Question =-=-=-=-=-=-=-=-=-=-=-=-=-= *
//==============================================================================

//! ------------------------------------------------------------------------------
//! Base Writing Question Types
//! ------------------------------------------------------------------------------

type WritingQuestionResponse struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Topic       []string  `json:"topic"`
	Instruction string    `json:"instruction"`
	ImageURLs   []string  `json:"image_urls"`
	MaxTime     int       `json:"max_time"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

type WritingQuestionDetail struct {
	WritingQuestionResponse
	SentenceCompletion []WritingSentenceCompletionResponse `json:"sentence_completion,omitempty"`
	Essay              []WritingEssayResponse              `json:"essay,omitempty"`
}

type CreateWritingQuestionRequest struct {
	Type        string   `json:"type" validate:"required,oneof=SENTENCE_COMPLETION ESSAY"`
	Topic       []string `json:"topic" validate:"required,min=1"`
	Instruction string   `json:"instruction" validate:"required"`
	ImageURLs   []string `json:"image_urls"`
	MaxTime     int      `json:"max_time" validate:"required,min=1"`
}

type UpdateWritingQuestionFieldRequest struct {
	Field string      `json:"field" validate:"required,oneof=topic instruction image_urls max_time"`
	Value interface{} `json:"value" validate:"required"`
}

type WritingQuestionVersionCheck struct {
	WritingQuestionID uuid.UUID `json:"writing_question_id" validate:"required"`
	Version           int       `json:"version" validate:"required"`
}

type GetNewUpdatesWritingQuestionRequest struct {
	Questions []WritingQuestionVersionCheck `json:"questions" validate:"required,min=1"`
}

type ListWritingQuestionsPagination struct {
	Questions []WritingQuestionDetail `json:"questions"`
	Total     int64                   `json:"total"`
	Page      int                     `json:"page"`
	PageSize  int                     `json:"page_size"`
}

type WritingQuestionSearchFilter struct {
	Type        string `form:"type"`
	Topic       string `form:"topic"`
	Instruction string `form:"instruction"`
	ImageURLs   string `form:"image_urls"`
	MaxTime     string `form:"max_time"`
	Metadata    string `form:"metadata"`
	Page        int    `form:"page" binding:"required,min=1"`
	PageSize    int    `form:"page_size" binding:"required,min=1,max=100"`
}

// ! ------------------------------------------------------------------------------
// ! Sentence Completion
// ! ------------------------------------------------------------------------------
type WritingSentenceCompletionResponse struct {
	ID                uuid.UUID `json:"id"`
	ExampleSentence   string    `json:"example_sentence"`
	GivenPartSentence string    `json:"given_part_sentence"`
	Position          string    `json:"position"`
	RequiredWords     []string  `json:"required_words"`
	Explain           string    `json:"explain"`
	MinWords          int       `json:"min_words"`
	MaxWords          int       `json:"max_words"`
}

type CreateWritingSentenceCompletionRequest struct {
	WritingQuestionID uuid.UUID `json:"writing_question_id" validate:"required"`
	ExampleSentence   string    `json:"example_sentence" validate:"required"`
	GivenPartSentence string    `json:"given_part_sentence" validate:"required"`
	Position          string    `json:"position" validate:"required,oneof=start end"`
	RequiredWords     []string  `json:"required_words" validate:"required,min=1"`
	Explain           string    `json:"explain" validate:"required"`
	MinWords          int       `json:"min_words" validate:"required,min=1"`
	MaxWords          int       `json:"max_words" validate:"required,gtefield=MinWords"`
}

type UpdateWritingSentenceCompletionRequest struct {
	WritingSentenceCompletionID uuid.UUID   `json:"writing_sentence_completion_id" validate:"required"`
	Field                       string      `json:"field" validate:"required,oneof=example_sentence given_part_sentence position required_words explain min_words max_words"`
	Value                       interface{} `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! Essay
// ! ------------------------------------------------------------------------------
type WritingEssayResponse struct {
	ID             uuid.UUID `json:"id"`
	EssayType      string    `json:"essay_type"`
	RequiredPoints []string  `json:"required_points"`
	MinWords       int       `json:"min_words"`
	MaxWords       int       `json:"max_words"`
	SampleEssay    string    `json:"sample_essay"`
	Explain        string    `json:"explain"`
}

type CreateWritingEssayRequest struct {
	WritingQuestionID uuid.UUID `json:"writing_question_id" validate:"required"`
	EssayType         string    `json:"essay_type" validate:"required"`
	RequiredPoints    []string  `json:"required_points" validate:"required,min=1"`
	MinWords          int       `json:"min_words" validate:"required,min=1"`
	MaxWords          int       `json:"max_words" validate:"required,gtefield=MinWords"`
	SampleEssay       string    `json:"sample_essay" validate:"required"`
	Explain           string    `json:"explain" validate:"required"`
}

type UpdateWritingEssayRequest struct {
	WritingEssayID uuid.UUID   `json:"writing_essay_id" validate:"required"`
	Field          string      `json:"field" validate:"required,oneof=essay_type required_points min_words max_words sample_essay explain"`
	Value          interface{} `json:"value" validate:"required"`
}
