package dto

import (
	"time"

	"github.com/google/uuid"
)

//==============================================================================
// * =-=-=-=-=-=-=-=-=-=-=-=-=-= Reading Question =-=-=-=-=-=-=-=-=-=-=-=-=-= *
//==============================================================================

//! ------------------------------------------------------------------------------
//! Base Reading Question Types
//! ------------------------------------------------------------------------------

type ReadingQuestionResponse struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Topic       []string  `json:"topic"`
	Instruction string    `json:"instruction"`
	Title       string    `json:"title"`
	Passages    []string  `json:"passages"`
	ImageURLs   []string  `json:"image_urls"`
	MaxTime     int       `json:"max_time"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

type ReadingQuestionDetail struct {
	ReadingQuestionResponse
	TrueFalse              []ReadingTrueFalseResponse             `json:"true_false,omitempty"`
	FillInTheBlankQuestion *ReadingFillInTheBlankQuestionResponse `json:"fill_in_the_blank_question,omitempty"`
	FillInTheBlankAnswers  []ReadingFillInTheBlankAnswerResponse  `json:"fill_in_the_blank_answers,omitempty"`
	ChoiceOneQuestion      *ReadingChoiceOneQuestionResponse      `json:"choice_one_question,omitempty"`
	ChoiceOneOptions       []ReadingChoiceOneOptionResponse       `json:"choice_one_options,omitempty"`
	ChoiceMultiQuestion    *ReadingChoiceMultiQuestionResponse    `json:"choice_multi_question,omitempty"`
	ChoiceMultiOptions     []ReadingChoiceMultiOptionResponse     `json:"choice_multi_options,omitempty"`
	Matching               []ReadingMatchingResponse              `json:"matching,omitempty"`
}

type CreateReadingQuestionRequest struct {
	Type        string   `json:"type" validate:"required"`
	Topic       []string `json:"topic" validate:"required,min=1"`
	Instruction string   `json:"instruction" validate:"required"`
	Title       string   `json:"title" validate:"required"`
	Passages    []string `json:"passages" validate:"required,min=1"`
	ImageURLs   []string `json:"image_urls"`
	MaxTime     int      `json:"max_time" validate:"required,min=1"`
}

type UpdateReadingQuestionFieldRequest struct {
	Field string      `json:"field" validate:"required,oneof=topic instruction title passages image_urls max_time"`
	Value interface{} `json:"value" validate:"required"`
}

type ReadingQuestionVersionCheck struct {
	ReadingQuestionID uuid.UUID `json:"reading_question_id" validate:"required"`
	Version           int       `json:"version" validate:"required"`
}

type GetNewUpdatesReadingQuestionRequest struct {
	Questions []ReadingQuestionVersionCheck `json:"questions" validate:"required,min=1"`
}

type ListReadingQuestionsPagination struct {
	Questions []ReadingQuestionDetail `json:"questions"`
	Total     int64                   `json:"total"`
	Page      int                     `json:"page"`
	PageSize  int                     `json:"page_size"`
}

type ReadingQuestionSearchRequest struct {
	Page         int    `form:"page" binding:"required,min=1"`
	PageSize     int    `form:"page_size" binding:"required,min=1,max=100"`
	QuestionType string `form:"question_type" binding:"required,oneof=basic complete uncomplete"`
	Type         string `form:"type" binding:"omitempty"`
	Topic        string `form:"topic" binding:"omitempty"`
}

type ReadingQuestionSearchFilter struct {
	Query       string `form:"query"`
	Type        string `form:"type"`
	Topic       string `form:"topic"`
	Instruction string `form:"instruction"`
	Title       string `form:"title"`
	Passages    string `form:"passages"`
	ImageURLs   string `form:"image_urls"`
	MaxTime     string `form:"max_time"`
	Metadata    string `form:"metadata"`
	Page        int    `form:"page" binding:"required,min=1"`
	PageSize    int    `form:"page_size" binding:"required,min=1,max=100"`
}

// ! ------------------------------------------------------------------------------
// ! True/False/Not Given
// ! ------------------------------------------------------------------------------
type ReadingTrueFalse struct {
	ID                uuid.UUID `json:"id"`
	ReadingQuestionID uuid.UUID `json:"reading_question_id"`
	Question          string    `json:"question"`
	Answer            string    `json:"answer"`
	Explain           string    `json:"explain"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
type ReadingTrueFalseResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
	Answer   string    `json:"answer"`
	Explain  string    `json:"explain"`
}

type CreateReadingTrueFalseRequest struct {
	ReadingQuestionID uuid.UUID `json:"reading_question_id" validate:"required"`
	Question          string    `json:"question" validate:"required"`
	Answer            string    `json:"answer" validate:"required,oneof=TRUE FALSE NOT GIVEN"`
	Explain           string    `json:"explain" validate:"required"`
}

type UpdateReadingTrueFalseRequest struct {
	ReadingTrueFalseID uuid.UUID `json:"reading_true_false_id" validate:"required"`
	Field              string    `json:"field" validate:"required,oneof=question answer explain"`
	Value              string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! - Fill In The Blank
// ! ------------------------------------------------------------------------------
// ? ------------------------------------------------------------------------------
// ? - Fill In The Blank Question
// ? ------------------------------------------------------------------------------
type ReadingFillInTheBlankQuestion struct {
	ID                uuid.UUID `json:"id"`
	ReadingQuestionID uuid.UUID `json:"reading_question_id"`
	Question          string    `json:"question"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ReadingFillInTheBlankQuestionResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
}

type CreateReadingFillInTheBlankQuestionRequest struct {
	ReadingQuestionID uuid.UUID `json:"reading_question_id" validate:"required"`
	Question          string    `json:"question" validate:"required"`
}

type UpdateReadingFillInTheBlankQuestionRequest struct {
	ReadingFillInTheBlankQuestionID uuid.UUID `json:"reading_fill_in_the_blank_question_id" validate:"required"`
	Field                           string    `json:"field" validate:"required,oneof=question"`
	Value                           string    `json:"value" validate:"required"`
}

// ? ------------------------------------------------------------------------------
// ? - Fill In The Blank Answer
// ? ------------------------------------------------------------------------------
type ReadingFillInTheBlankAnswer struct {
	ID                              uuid.UUID `json:"id"`
	ReadingFillInTheBlankQuestionID uuid.UUID `json:"reading_fill_in_the_blank_question_id"`
	Answer                          string    `json:"answer"`
	Explain                         string    `json:"explain"`
	CreatedAt                       time.Time `json:"created_at"`
	UpdatedAt                       time.Time `json:"updated_at"`
}

type ReadingFillInTheBlankAnswerResponse struct {
	ID      uuid.UUID `json:"id"`
	Answer  string    `json:"answer"`
	Explain string    `json:"explain"`
}

type CreateReadingFillInTheBlankAnswerRequest struct {
	ReadingFillInTheBlankQuestionID uuid.UUID `json:"reading_fill_in_the_blank_question_id" validate:"required"`
	Answer                          string    `json:"answer" validate:"required"`
	Explain                         string    `json:"explain" validate:"required"`
}

type UpdateReadingFillInTheBlankAnswerRequest struct {
	ReadingFillInTheBlankAnswerID uuid.UUID `json:"reading_fill_in_the_blank_answer_id" validate:"required"`
	Field                         string    `json:"field" validate:"required,oneof=answer explain"`
	Value                         string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! - Matching
// ! ------------------------------------------------------------------------------
type ReadingMatching struct {
	ID                uuid.UUID `json:"id"`
	ReadingQuestionID uuid.UUID `json:"reading_question_id"`
	Question          string    `json:"question"`
	Answer            string    `json:"answer"`
	Explain           string    `json:"explain"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ReadingMatchingResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
	Answer   string    `json:"answer"`
	Explain  string    `json:"explain"`
}

type CreateReadingMatchingRequest struct {
	ReadingQuestionID uuid.UUID `json:"reading_question_id" validate:"required"`
	Question          string    `json:"question" validate:"required"`
	Answer            string    `json:"answer" validate:"required"`
	Explain           string    `json:"explain" validate:"required"`
}

type UpdateReadingMatchingRequest struct {
	ReadingMatchingID uuid.UUID `json:"reading_matching_id" validate:"required"`
	Field             string    `json:"field" validate:"required,oneof=question answer explain"`
	Value             string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! Choice One
// ! ------------------------------------------------------------------------------
// ? ------------------------------------------------------------------------------
// ? - Choice One Question
// ? ------------------------------------------------------------------------------
type ReadingChoiceOneQuestion struct {
	ID                uuid.UUID `json:"id"`
	ReadingQuestionID uuid.UUID `json:"reading_question_id"`
	Question          string    `json:"question"`
	Explain           string    `json:"explain"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ReadingChoiceOneQuestionResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
	Explain  string    `json:"explain"`
}

type CreateReadingChoiceOneQuestionRequest struct {
	ReadingQuestionID uuid.UUID `json:"reading_question_id" validate:"required"`
	Question          string    `json:"question" validate:"required"`
	Explain           string    `json:"explain" validate:"required"`
}

type UpdateReadingChoiceOneQuestionRequest struct {
	ReadingChoiceOneQuestionID uuid.UUID `json:"reading_choice_one_question_id" validate:"required"`
	Field                      string    `json:"field" validate:"required,oneof=question explain"`
	Value                      string    `json:"value" validate:"required"`
}

// ? ------------------------------------------------------------------------------
// ? - Choice One Option
// ? ------------------------------------------------------------------------------
type ReadingChoiceOneOption struct {
	ID                         uuid.UUID `json:"id"`
	ReadingChoiceOneQuestionID uuid.UUID `json:"reading_choice_one_question_id"`
	Options                    string    `json:"options"`
	IsCorrect                  bool      `json:"is_correct"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
}

type ReadingChoiceOneOptionResponse struct {
	ID        uuid.UUID `json:"id"`
	Options   string    `json:"options"`
	IsCorrect bool      `json:"is_correct"`
}

type CreateReadingChoiceOneOptionRequest struct {
	ReadingChoiceOneQuestionID uuid.UUID `json:"reading_choice_one_question_id" validate:"required"`
	Options                    string    `json:"options" validate:"required"`
	IsCorrect                  bool      `json:"is_correct" validate:"required"`
}

type UpdateReadingChoiceOneOptionRequest struct {
	ReadingChoiceOneOptionID uuid.UUID `json:"reading_choice_one_option_id" validate:"required"`
	Field                    string    `json:"field" validate:"required,oneof=options is_correct"`
	Value                    string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! Choice Multi
// ! ------------------------------------------------------------------------------
// ? ------------------------------------------------------------------------------
// ? - Choice Multi Question
// ? ------------------------------------------------------------------------------
type ReadingChoiceMultiQuestion struct {
	ID                uuid.UUID `json:"id"`
	ReadingQuestionID uuid.UUID `json:"reading_question_id"`
	Question          string    `json:"question"`
	Explain           string    `json:"explain"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ReadingChoiceMultiQuestionResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
	Explain  string    `json:"explain"`
}

type CreateReadingChoiceMultiQuestionRequest struct {
	ReadingQuestionID uuid.UUID `json:"reading_question_id" validate:"required"`
	Question          string    `json:"question" validate:"required"`
	Explain           string    `json:"explain" validate:"required"`
}

type UpdateReadingChoiceMultiQuestionRequest struct {
	ReadingChoiceMultiQuestionID uuid.UUID `json:"reading_choice_multi_question_id" validate:"required"`
	Field                        string    `json:"field" validate:"required,oneof=question explain"`
	Value                        string    `json:"value" validate:"required"`
}

// ? ------------------------------------------------------------------------------
// ? - Choice Multi Option
// ? ------------------------------------------------------------------------------
type ReadingChoiceMultiOption struct {
	ID                           uuid.UUID `json:"id"`
	ReadingChoiceMultiQuestionID uuid.UUID `json:"reading_choice_multi_question_id"`
	Options                      string    `json:"options"`
	IsCorrect                    bool      `json:"is_correct"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
}

type ReadingChoiceMultiOptionResponse struct {
	ID        uuid.UUID `json:"id"`
	Options   string    `json:"options"`
	IsCorrect bool      `json:"is_correct"`
}

type CreateReadingChoiceMultiOptionRequest struct {
	ReadingChoiceMultiQuestionID uuid.UUID `json:"reading_choice_multi_question_id" validate:"required"`
	Options                      string    `json:"options" validate:"required"`
	IsCorrect                    bool      `json:"is_correct" validate:"required"`
}

type UpdateReadingChoiceMultiOptionRequest struct {
	ReadingChoiceMultiOptionID uuid.UUID `json:"reading_choice_multi_option_id" validate:"required"`
	Field                      string    `json:"field" validate:"required,oneof=options is_correct"`
	Value                      string    `json:"value" validate:"required"`
}
