package dto

import (
	"time"

	"github.com/google/uuid"
)

//==============================================================================
// * =-=-=-=-=-=-=-=-=-=-=-=-=-= Listening Question =-=-=-=-=-=-=-=-=-=-=-=-=-= *
//==============================================================================

//! ------------------------------------------------------------------------------
//! Base Listening Question Types
//! ------------------------------------------------------------------------------

type ListeningQuestionResponse struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Topic       []string  `json:"topic"`
	Instruction string    `json:"instruction"`
	AudioURLs   []string  `json:"audio_urls"`
	ImageURLs   []string  `json:"image_urls"`
	Transcript  string    `json:"transcript"`
	MaxTime     int       `json:"max_time"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

type ListeningQuestionDetail struct {
	ListeningQuestionResponse
	FillInTheBlankQuestion *ListeningFillInTheBlankQuestionResponse `json:"fill_in_the_blank_question,omitempty"`
	FillInTheBlankAnswers  []ListeningFillInTheBlankAnswerResponse  `json:"fill_in_the_blank_answers,omitempty"`
	ChoiceOneQuestion      *ListeningChoiceOneQuestionResponse      `json:"choice_one_question,omitempty"`
	ChoiceOneOptions       []ListeningChoiceOneOptionResponse       `json:"choice_one_options,omitempty"`
	ChoiceMultiQuestion    *ListeningChoiceMultiQuestionResponse    `json:"choice_multi_question,omitempty"`
	ChoiceMultiOptions     []ListeningChoiceMultiOptionResponse     `json:"choice_multi_options,omitempty"`
	MapLabelling           []ListeningMapLabellingResponse          `json:"map_labelling,omitempty"`
	Matching               []ListeningMatchingResponse              `json:"matching,omitempty"`
}

type CreateListeningQuestionRequest struct {
	Type        string   `json:"type" validate:"required"`
	Topic       []string `json:"topic" validate:"required,min=1"`
	Instruction string   `json:"instruction" validate:"required"`
	AudioURLs   []string `json:"audio_urls" validate:"required,min=1"`
	ImageURLs   []string `json:"image_urls"`
	Transcript  string   `json:"transcript" validate:"required"`
	MaxTime     int      `json:"max_time" validate:"required,min=1"`
}

type UpdateListeningQuestionFieldRequest struct {
	Field string      `json:"field" validate:"required,oneof=topic instruction audio_urls image_urls transcript max_time"`
	Value interface{} `json:"value" validate:"required"`
}

type ListeningQuestionVersionCheck struct {
	ListeningQuestionID uuid.UUID `json:"listening_question_id" validate:"required"`
	Version             int       `json:"version" validate:"required"`
}

type GetNewUpdatesListeningQuestionRequest struct {
	Questions []ListeningQuestionVersionCheck `json:"questions" validate:"required,min=1"`
}

type ListListeningQuestionsPagination struct {
	Questions []ListeningQuestionDetail `json:"questions"`
	Total     int64                     `json:"total"`
	Page      int                       `json:"page"`
	PageSize  int                       `json:"page_size"`
}

type ListeningQuestionSearchRequest struct {
	Page         int    `form:"page" binding:"required,min=1"`
	PageSize     int    `form:"page_size" binding:"required,min=1,max=100"`
	QuestionType string `form:"question_type" binding:"required,oneof=basic complete uncomplete"`
	Type         string `form:"type" binding:"omitempty"`
	Topic        string `form:"topic" binding:"omitempty"`
}

type ListeningQuestionSearchFilter struct {
	Type        string `form:"type"`
	Topic       string `form:"topic"`
	Instruction string `form:"instruction"`
	AudioURLs   string `form:"audio_urls"`
	ImageURLs   string `form:"image_urls"`
	Transcript  string `form:"transcript"`
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
type ListeningFillInTheBlankQuestion struct {
	ID                  uuid.UUID `json:"id"`
	ListeningQuestionID uuid.UUID `json:"listening_question_id"`
	Question            string    `json:"question"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ListeningFillInTheBlankQuestionResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
}

type CreateListeningFillInTheBlankQuestionRequest struct {
	ListeningQuestionID uuid.UUID `json:"listening_question_id" validate:"required"`
	Question            string    `json:"question" validate:"required"`
}

type UpdateListeningFillInTheBlankQuestionRequest struct {
	ListeningFillInTheBlankQuestionID uuid.UUID `json:"listening_fill_in_the_blank_question_id" validate:"required"`
	Field                             string    `json:"field" validate:"required,oneof=question"`
	Value                             string    `json:"value" validate:"required"`
}

// ? ------------------------------------------------------------------------------
// ? - Fill In The Blank Answer
// ? ------------------------------------------------------------------------------
type ListeningFillInTheBlankAnswer struct {
	ID                                uuid.UUID `json:"id"`
	ListeningFillInTheBlankQuestionID uuid.UUID `json:"listening_fill_in_the_blank_question_id"`
	Answer                            string    `json:"answer"`
	Explain                           string    `json:"explain"`
	CreatedAt                         time.Time `json:"created_at"`
	UpdatedAt                         time.Time `json:"updated_at"`
}

type ListeningFillInTheBlankAnswerResponse struct {
	ID      uuid.UUID `json:"id"`
	Answer  string    `json:"answer"`
	Explain string    `json:"explain"`
}

type CreateListeningFillInTheBlankAnswerRequest struct {
	ListeningFillInTheBlankQuestionID uuid.UUID `json:"listening_fill_in_the_blank_question_id" validate:"required"`
	Answer                            string    `json:"answer" validate:"required"`
	Explain                           string    `json:"explain" validate:"required"`
}

type UpdateListeningFillInTheBlankAnswerRequest struct {
	ListeningFillInTheBlankAnswerID uuid.UUID `json:"listening_fill_in_the_blank_answer_id" validate:"required"`
	Field                           string    `json:"field" validate:"required,oneof=answer explain"`
	Value                           string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! - Choice One
// ! ------------------------------------------------------------------------------
// ? ------------------------------------------------------------------------------
// ? - Choice One Question
// ? ------------------------------------------------------------------------------
type ListeningChoiceOneQuestion struct {
	ID                  uuid.UUID `json:"id"`
	ListeningQuestionID uuid.UUID `json:"listening_question_id"`
	Question            string    `json:"question"`
	Explain             string    `json:"explain"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ListeningChoiceOneQuestionResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
	Explain  string    `json:"explain"`
}

type CreateListeningChoiceOneQuestionRequest struct {
	ListeningQuestionID uuid.UUID `json:"listening_question_id" validate:"required"`
	Question            string    `json:"question" validate:"required"`
	Explain             string    `json:"explain" validate:"required"`
}

type UpdateListeningChoiceOneQuestionRequest struct {
	ListeningChoiceOneQuestionID uuid.UUID `json:"listening_choice_one_question_id" validate:"required"`
	Field                        string    `json:"field" validate:"required,oneof=question correct_option_id explain"`
	Value                        string    `json:"value" validate:"required"`
}

// ? ------------------------------------------------------------------------------
// ? - Choice One Option
// ? ------------------------------------------------------------------------------
type ListeningChoiceOneOption struct {
	ID                           uuid.UUID `json:"id"`
	ListeningChoiceOneQuestionID uuid.UUID `json:"listening_choice_one_question_id"`
	Options                      string    `json:"options"`
	IsCorrect                    bool      `json:"is_correct"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
}

type ListeningChoiceOneOptionResponse struct {
	ID        uuid.UUID `json:"id"`
	Options   string    `json:"options"`
	IsCorrect bool      `json:"is_correct"`
}

type CreateListeningChoiceOneOptionRequest struct {
	ListeningChoiceOneQuestionID uuid.UUID `json:"listening_choice_one_question_id" validate:"required"`
	Options                      string    `json:"options" validate:"required"`
	IsCorrect                    bool      `json:"is_correct" validate:"required"`
}

type UpdateListeningChoiceOneOptionRequest struct {
	ListeningChoiceOneOptionID uuid.UUID `json:"listening_choice_one_option_id" validate:"required"`
	Field                      string    `json:"field" validate:"required,oneof=options is_correct"`
	Value                      string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! - Choice Multi
// ! ------------------------------------------------------------------------------
// ? ------------------------------------------------------------------------------
// ? - Choice Multi Question
// ? ------------------------------------------------------------------------------
type ListeningChoiceMultiQuestion struct {
	ID                  uuid.UUID `json:"id"`
	ListeningQuestionID uuid.UUID `json:"listening_question_id"`
	Question            string    `json:"question"`
	Explain             string    `json:"explain"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ListeningChoiceMultiQuestionResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
	Explain  string    `json:"explain"`
}

type CreateListeningChoiceMultiQuestionRequest struct {
	ListeningQuestionID uuid.UUID `json:"listening_question_id" validate:"required"`
	Question            string    `json:"question" validate:"required"`
	Explain             string    `json:"explain" validate:"required"`
}

type UpdateListeningChoiceMultiQuestionRequest struct {
	ListeningChoiceMultiQuestionID uuid.UUID `json:"listening_choice_multi_question_id" validate:"required"`
	Field                          string    `json:"field" validate:"required,oneof=question explain"`
	Value                          string    `json:"value" validate:"required"`
}

// ? ------------------------------------------------------------------------------
// ? - Choice Multi Option
// ? ------------------------------------------------------------------------------
type ListeningChoiceMultiOption struct {
	ID                             uuid.UUID `json:"id"`
	ListeningChoiceMultiQuestionID uuid.UUID `json:"listening_choice_multi_question_id"`
	Options                        string    `json:"options"`
	IsCorrect                      bool      `json:"is_correct"`
	CreatedAt                      time.Time `json:"created_at"`
	UpdatedAt                      time.Time `json:"updated_at"`
}

type ListeningChoiceMultiOptionResponse struct {
	ID        uuid.UUID `json:"id"`
	Options   string    `json:"options"`
	IsCorrect bool      `json:"is_correct"`
}

type CreateListeningChoiceMultiOptionRequest struct {
	ListeningChoiceMultiQuestionID uuid.UUID `json:"listening_choice_multi_question_id" validate:"required"`
	Options                        string    `json:"options" validate:"required"`
	IsCorrect                      bool      `json:"is_correct" validate:"required"`
}

type UpdateListeningChoiceMultiOptionRequest struct {
	ListeningChoiceMultiOptionID uuid.UUID `json:"listening_choice_multi_option_id" validate:"required"`
	Field                        string    `json:"field" validate:"required,oneof=options is_correct"`
	Value                        string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! - Map Labelling
// ! ------------------------------------------------------------------------------
type ListeningMapLabelling struct {
	ID                  uuid.UUID `json:"id"`
	ListeningQuestionID uuid.UUID `json:"listening_question_id"`
	Question            string    `json:"question"`
	Answer              string    `json:"answer"`
	Explain             string    `json:"explain"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ListeningMapLabellingResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
	Answer   string    `json:"answer"`
	Explain  string    `json:"explain"`
}

type CreateListeningMapLabellingRequest struct {
	ListeningQuestionID uuid.UUID `json:"listening_question_id" validate:"required"`
	Question            string    `json:"question" validate:"required"`
	Answer              string    `json:"answer" validate:"required"`
	Explain             string    `json:"explain" validate:"required"`
}

type UpdateListeningMapLabellingRequest struct {
	ListeningMapLabellingID uuid.UUID `json:"listening_map_labelling_id" validate:"required"`
	Field                   string    `json:"field" validate:"required,oneof=question answer explain"`
	Value                   string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! - Matching
// ! ------------------------------------------------------------------------------
type ListeningMatching struct {
	ID                  uuid.UUID `json:"id"`
	ListeningQuestionID uuid.UUID `json:"listening_question_id"`
	Question            string    `json:"question"`
	Answer              string    `json:"answer"`
	Explain             string    `json:"explain"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ListeningMatchingResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
	Answer   string    `json:"answer"`
	Explain  string    `json:"explain"`
}

type CreateListeningMatchingRequest struct {
	ListeningQuestionID uuid.UUID `json:"listening_question_id" validate:"required"`
	Question            string    `json:"question" validate:"required"`
	Answer              string    `json:"answer" validate:"required"`
	Explain             string    `json:"explain" validate:"required"`
}

type UpdateListeningMatchingRequest struct {
	ListeningMatchingID uuid.UUID `json:"listening_matching_id" validate:"required"`
	Field               string    `json:"field" validate:"required,oneof=question answer explain"`
	Value               string    `json:"value" validate:"required"`
}
