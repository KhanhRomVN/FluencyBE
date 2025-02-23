package dto

import (
	"time"

	"github.com/google/uuid"
)

//==============================================================================
// * =-=-=-=-=-=-=-=-=-=-=-=-=-= Grammar Question =-=-=-=-=-=-=-=-=-=-=-=-=-= *
//==============================================================================

//! ------------------------------------------------------------------------------
//! Base Grammar Question Types
//! ------------------------------------------------------------------------------

type GrammarQuestionResponse struct {
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

type GrammarQuestionDetail struct {
	GrammarQuestionResponse
	FillInTheBlankQuestion *GrammarFillInTheBlankQuestionResponse `json:"fill_in_the_blank_question,omitempty"`
	FillInTheBlankAnswers  []GrammarFillInTheBlankAnswerResponse  `json:"fill_in_the_blank_answers,omitempty"`
	ChoiceOneQuestion      *GrammarChoiceOneQuestionResponse      `json:"choice_one_question,omitempty"`
	ChoiceOneOptions       []GrammarChoiceOneOptionResponse       `json:"choice_one_options,omitempty"`
	ErrorIdentification    *GrammarErrorIdentificationResponse    `json:"error_identification,omitempty"`
	SentenceTransformation *GrammarSentenceTransformationResponse `json:"sentence_transformation,omitempty"`
}

type CreateGrammarQuestionRequest struct {
	Type        string   `json:"type" validate:"required"`
	Topic       []string `json:"topic" validate:"required,min=1"`
	Instruction string   `json:"instruction" validate:"required"`
	ImageURLs   []string `json:"image_urls"`
	MaxTime     int      `json:"max_time" validate:"required,min=1"`
}

type UpdateGrammarQuestionFieldRequest struct {
	Field string      `json:"field" validate:"required,oneof=topic instruction image_urls max_time"`
	Value interface{} `json:"value" validate:"required"`
}

type GrammarQuestionVersionCheck struct {
	GrammarQuestionID uuid.UUID `json:"grammar_question_id" validate:"required"`
	Version           int       `json:"version" validate:"required"`
}

type GetNewUpdatesGrammarQuestionRequest struct {
	Questions []GrammarQuestionVersionCheck `json:"questions" validate:"required,min=1"`
}

type ListGrammarQuestionsPagination struct {
	Questions []GrammarQuestionDetail `json:"questions"`
	Total     int64                   `json:"total"`
	Page      int                     `json:"page"`
	PageSize  int                     `json:"page_size"`
}

type GrammarQuestionSearchRequest struct {
	Page         int    `form:"page" binding:"required,min=1"`
	PageSize     int    `form:"page_size" binding:"required,min=1,max=100"`
	QuestionType string `form:"question_type" binding:"required,oneof=basic complete uncomplete"`
	Type         string `form:"type" binding:"omitempty"`
	Topic        string `form:"topic" binding:"omitempty"`
}

type GrammarQuestionSearchFilter struct {
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
// ! - Fill In The Blank
// ! ------------------------------------------------------------------------------
// ? ------------------------------------------------------------------------------
// ? - Fill In The Blank Question
// ? ------------------------------------------------------------------------------
type GrammarFillInTheBlankQuestion struct {
	ID                uuid.UUID `json:"id"`
	GrammarQuestionID uuid.UUID `json:"grammar_question_id"`
	Question          string    `json:"question"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type GrammarFillInTheBlankQuestionResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
}

type CreateGrammarFillInTheBlankQuestionRequest struct {
	GrammarQuestionID uuid.UUID `json:"grammar_question_id" validate:"required"`
	Question          string    `json:"question" validate:"required"`
}

type UpdateGrammarFillInTheBlankQuestionRequest struct {
	GrammarFillInTheBlankQuestionID uuid.UUID `json:"grammar_fill_in_the_blank_question_id" validate:"required"`
	Field                           string    `json:"field" validate:"required,oneof=question"`
	Value                           string    `json:"value" validate:"required"`
}

// ? ------------------------------------------------------------------------------
// ? - Fill In The Blank Answer
// ? ------------------------------------------------------------------------------
type GrammarFillInTheBlankAnswer struct {
	ID                              uuid.UUID `json:"id"`
	GrammarFillInTheBlankQuestionID uuid.UUID `json:"grammar_fill_in_the_blank_question_id"`
	Answer                          string    `json:"answer"`
	Explain                         string    `json:"explain"`
	CreatedAt                       time.Time `json:"created_at"`
	UpdatedAt                       time.Time `json:"updated_at"`
}

type GrammarFillInTheBlankAnswerResponse struct {
	ID      uuid.UUID `json:"id"`
	Answer  string    `json:"answer"`
	Explain string    `json:"explain"`
}

type CreateGrammarFillInTheBlankAnswerRequest struct {
	GrammarFillInTheBlankQuestionID uuid.UUID `json:"grammar_fill_in_the_blank_question_id" validate:"required"`
	Answer                          string    `json:"answer" validate:"required"`
	Explain                         string    `json:"explain" validate:"required"`
}

type UpdateGrammarFillInTheBlankAnswerRequest struct {
	GrammarFillInTheBlankAnswerID uuid.UUID `json:"grammar_fill_in_the_blank_answer_id" validate:"required"`
	Field                         string    `json:"field" validate:"required,oneof=answer explain"`
	Value                         string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! - Choice One
// ! ------------------------------------------------------------------------------
// ? ------------------------------------------------------------------------------
// ? - Choice One Question
// ? ------------------------------------------------------------------------------
type GrammarChoiceOneQuestion struct {
	ID                uuid.UUID `json:"id"`
	GrammarQuestionID uuid.UUID `json:"grammar_question_id"`
	Question          string    `json:"question"`
	Explain           string    `json:"explain"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type GrammarChoiceOneQuestionResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
	Explain  string    `json:"explain"`
}

type CreateGrammarChoiceOneQuestionRequest struct {
	GrammarQuestionID uuid.UUID `json:"grammar_question_id" validate:"required"`
	Question          string    `json:"question" validate:"required"`
	Explain           string    `json:"explain" validate:"required"`
}

type UpdateGrammarChoiceOneQuestionRequest struct {
	GrammarChoiceOneQuestionID uuid.UUID `json:"grammar_choice_one_question_id" validate:"required"`
	Field                      string    `json:"field" validate:"required,oneof=question explain"`
	Value                      string    `json:"value" validate:"required"`
}

// ? ------------------------------------------------------------------------------
// ? - Choice One Option
// ? ------------------------------------------------------------------------------
type GrammarChoiceOneOption struct {
	ID                         uuid.UUID `json:"id"`
	GrammarChoiceOneQuestionID uuid.UUID `json:"grammar_choice_one_question_id"`
	Options                    string    `json:"options"`
	IsCorrect                  bool      `json:"is_correct"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
}

type GrammarChoiceOneOptionResponse struct {
	ID        uuid.UUID `json:"id"`
	Options   string    `json:"options"`
	IsCorrect bool      `json:"is_correct"`
}

type CreateGrammarChoiceOneOptionRequest struct {
	GrammarChoiceOneQuestionID uuid.UUID `json:"grammar_choice_one_question_id" validate:"required"`
	Options                    string    `json:"options" validate:"required"`
	IsCorrect                  bool      `json:"is_correct" validate:"required"`
}

type UpdateGrammarChoiceOneOptionRequest struct {
	GrammarChoiceOneOptionID uuid.UUID `json:"grammar_choice_one_option_id" validate:"required"`
	Field                    string    `json:"field" validate:"required,oneof=options is_correct"`
	Value                    string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! - Error Identification
// ! ------------------------------------------------------------------------------
type GrammarErrorIdentification struct {
	ID                uuid.UUID `json:"id"`
	GrammarQuestionID uuid.UUID `json:"grammar_question_id"`
	ErrorSentence     string    `json:"error_sentence"`
	ErrorWord         string    `json:"error_word"`
	CorrectWord       string    `json:"correct_word"`
	Explain           string    `json:"explain"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type GrammarErrorIdentificationResponse struct {
	ID            uuid.UUID `json:"id"`
	ErrorSentence string    `json:"error_sentence"`
	ErrorWord     string    `json:"error_word"`
	CorrectWord   string    `json:"correct_word"`
	Explain       string    `json:"explain"`
}

type CreateGrammarErrorIdentificationRequest struct {
	GrammarQuestionID uuid.UUID `json:"grammar_question_id" validate:"required"`
	ErrorSentence     string    `json:"error_sentence" validate:"required"`
	ErrorWord         string    `json:"error_word" validate:"required"`
	CorrectWord       string    `json:"correct_word" validate:"required"`
	Explain           string    `json:"explain" validate:"required"`
}

type UpdateGrammarErrorIdentificationRequest struct {
	GrammarErrorIdentificationID uuid.UUID `json:"grammar_error_identification_id" validate:"required"`
	Field                        string    `json:"field" validate:"required,oneof=error_sentence error_word correct_word explain"`
	Value                        string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! - Sentence Transformation
// ! ------------------------------------------------------------------------------
type GrammarSentenceTransformation struct {
	ID                     uuid.UUID `json:"id"`
	GrammarQuestionID      uuid.UUID `json:"grammar_question_id"`
	OriginalSentence       string    `json:"original_sentence"`
	BeginningWord          string    `json:"beginning_word"`
	ExampleCorrectSentence string    `json:"example_correct_sentence"`
	Explain                string    `json:"explain"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type GrammarSentenceTransformationResponse struct {
	ID                     uuid.UUID `json:"id"`
	OriginalSentence       string    `json:"original_sentence"`
	BeginningWord          string    `json:"beginning_word"`
	ExampleCorrectSentence string    `json:"example_correct_sentence"`
	Explain                string    `json:"explain"`
}

type CreateGrammarSentenceTransformationRequest struct {
	GrammarQuestionID      uuid.UUID `json:"grammar_question_id" validate:"required"`
	OriginalSentence       string    `json:"original_sentence" validate:"required"`
	BeginningWord          string    `json:"beginning_word"`
	ExampleCorrectSentence string    `json:"example_correct_sentence" validate:"required"`
	Explain                string    `json:"explain" validate:"required"`
}

type UpdateGrammarSentenceTransformationRequest struct {
	GrammarSentenceTransformationID uuid.UUID `json:"grammar_sentence_transformation_id" validate:"required"`
	Field                           string    `json:"field" validate:"required,oneof=original_sentence beginning_word example_correct_sentence explain"`
	Value                           string    `json:"value" validate:"required"`
}
