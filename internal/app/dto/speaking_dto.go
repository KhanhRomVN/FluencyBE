package dto

import (
	"time"

	"github.com/google/uuid"
)

//==============================================================================
// * =-=-=-=-=-=-=-=-=-=-=-=-=-= Speaking Question =-=-=-=-=-=-=-=-=-=-=-=-=-= *
//==============================================================================

//! ------------------------------------------------------------------------------
//! Base Speaking Question Types
//! ------------------------------------------------------------------------------

type SpeakingQuestionResponse struct {
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

type SpeakingQuestionDetail struct {
	SpeakingQuestionResponse
	WordRepetition              []SpeakingWordRepetitionResponse             `json:"word_repetition,omitempty"`
	PhraseRepetition            []SpeakingPhraseRepetitionResponse           `json:"phrase_repetition,omitempty"`
	ParagraphRepetition         []SpeakingParagraphRepetitionResponse        `json:"paragraph_repetition,omitempty"`
	OpenParagraph               []SpeakingOpenParagraphResponse              `json:"open_paragraph,omitempty"`
	ConversationalRepetition    *SpeakingConversationalRepetitionResponse    `json:"conversational_repetition,omitempty"`
	ConversationalRepetitionQAs []SpeakingConversationalRepetitionQAResponse `json:"conversational_repetition_qas,omitempty"`
	ConversationalOpen          *SpeakingConversationalOpenResponse          `json:"conversational_open,omitempty"`
}

type CreateSpeakingQuestionRequest struct {
	Type        string   `json:"type" validate:"required,oneof=WORD_REPETITION PHRASE_REPETITION PARAGRAPH_REPETITION OPEN_PARAGRAPH CONVERSATIONAL_REPETITION CONVERSATIONAL_OPEN"`
	Topic       []string `json:"topic" validate:"required,min=1"`
	Instruction string   `json:"instruction" validate:"required"`
	ImageURLs   []string `json:"image_urls"`
	MaxTime     int      `json:"max_time" validate:"required,min=1"`
}

type UpdateSpeakingQuestionFieldRequest struct {
	Field string      `json:"field" validate:"required,oneof=topic instruction image_urls max_time"`
	Value interface{} `json:"value" validate:"required"`
}

type SpeakingQuestionVersionCheck struct {
	SpeakingQuestionID uuid.UUID `json:"speaking_question_id" validate:"required"`
	Version            int       `json:"version" validate:"required"`
}

type GetNewUpdatesSpeakingQuestionRequest struct {
	Questions []SpeakingQuestionVersionCheck `json:"questions" validate:"required,min=1"`
}

type ListSpeakingQuestionsPagination struct {
	Questions []SpeakingQuestionDetail `json:"questions"`
	Total     int64                    `json:"total"`
	Page      int                      `json:"page"`
	PageSize  int                      `json:"page_size"`
}

type SpeakingQuestionSearchRequest struct {
	Page         int    `form:"page" binding:"required,min=1"`
	PageSize     int    `form:"page_size" binding:"required,min=1,max=100"`
	QuestionType string `form:"question_type" binding:"required,oneof=basic complete uncomplete"`
	Type         string `form:"type" binding:"omitempty"`
	Topic        string `form:"topic" binding:"omitempty"`
}

type SpeakingQuestionSearchFilter struct {
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
// ! Word Repetition
// ! ------------------------------------------------------------------------------
type SpeakingWordRepetitionResponse struct {
	ID   uuid.UUID `json:"id"`
	Word string    `json:"word"`
	Mean string    `json:"mean"`
}

type CreateSpeakingWordRepetitionRequest struct {
	SpeakingQuestionID uuid.UUID `json:"speaking_question_id" validate:"required"`
	Word               string    `json:"word" validate:"required"`
	Mean               string    `json:"mean" validate:"required"`
}

type UpdateSpeakingWordRepetitionRequest struct {
	SpeakingWordRepetitionID uuid.UUID `json:"speaking_word_repetition_id" validate:"required"`
	Field                    string    `json:"field" validate:"required,oneof=word mean"`
	Value                    string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! Phrase Repetition
// ! ------------------------------------------------------------------------------
type SpeakingPhraseRepetitionResponse struct {
	ID     uuid.UUID `json:"id"`
	Phrase string    `json:"phrase"`
	Mean   string    `json:"mean"`
}

type CreateSpeakingPhraseRepetitionRequest struct {
	SpeakingQuestionID uuid.UUID `json:"speaking_question_id" validate:"required"`
	Phrase             string    `json:"phrase" validate:"required"`
	Mean               string    `json:"mean" validate:"required"`
}

type UpdateSpeakingPhraseRepetitionRequest struct {
	SpeakingPhraseRepetitionID uuid.UUID `json:"speaking_phrase_repetition_id" validate:"required"`
	Field                      string    `json:"field" validate:"required,oneof=phrase mean"`
	Value                      string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! Paragraph Repetition
// ! ------------------------------------------------------------------------------
type SpeakingParagraphRepetitionResponse struct {
	ID        uuid.UUID `json:"id"`
	Paragraph string    `json:"paragraph"`
	Mean      string    `json:"mean"`
}

type CreateSpeakingParagraphRepetitionRequest struct {
	SpeakingQuestionID uuid.UUID `json:"speaking_question_id" validate:"required"`
	Paragraph          string    `json:"paragraph" validate:"required"`
	Mean               string    `json:"mean" validate:"required"`
}

type UpdateSpeakingParagraphRepetitionRequest struct {
	SpeakingParagraphRepetitionID uuid.UUID `json:"speaking_paragraph_repetition_id" validate:"required"`
	Field                         string    `json:"field" validate:"required,oneof=paragraph mean"`
	Value                         string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! Open Paragraph
// ! ------------------------------------------------------------------------------
type SpeakingOpenParagraphResponse struct {
	ID                   uuid.UUID `json:"id"`
	Question             string    `json:"question"`
	ExamplePassage       string    `json:"example_passage"`
	MeanOfExamplePassage string    `json:"mean_of_example_passage"`
}

type CreateSpeakingOpenParagraphRequest struct {
	SpeakingQuestionID   uuid.UUID `json:"speaking_question_id" validate:"required"`
	Question             string    `json:"question" validate:"required"`
	ExamplePassage       string    `json:"example_passage" validate:"required"`
	MeanOfExamplePassage string    `json:"mean_of_example_passage" validate:"required"`
}

type UpdateSpeakingOpenParagraphRequest struct {
	SpeakingOpenParagraphID uuid.UUID `json:"speaking_open_paragraph_id" validate:"required"`
	Field                   string    `json:"field" validate:"required,oneof=question example_passage mean_of_example_passage"`
	Value                   string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! Conversational Repetition
// ! ------------------------------------------------------------------------------
type SpeakingConversationalRepetitionResponse struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Overview string    `json:"overview"`
}

type CreateSpeakingConversationalRepetitionRequest struct {
	SpeakingQuestionID uuid.UUID `json:"speaking_question_id" validate:"required"`
	Title              string    `json:"title" validate:"required"`
	Overview           string    `json:"overview" validate:"required"`
}

type UpdateSpeakingConversationalRepetitionRequest struct {
	SpeakingConversationalRepetitionID uuid.UUID `json:"speaking_conversational_repetition_id" validate:"required"`
	Field                              string    `json:"field" validate:"required,oneof=title overview"`
	Value                              string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! Conversational Repetition QA
// ! ------------------------------------------------------------------------------
type SpeakingConversationalRepetitionQAResponse struct {
	ID             uuid.UUID `json:"id"`
	Question       string    `json:"question"`
	Answer         string    `json:"answer"`
	MeanOfQuestion string    `json:"mean_of_question"`
	MeanOfAnswer   string    `json:"mean_of_answer"`
	Explain        string    `json:"explain"`
}

type CreateSpeakingConversationalRepetitionQARequest struct {
	SpeakingConversationalRepetitionID uuid.UUID `json:"speaking_conversational_repetition_id" validate:"required"`
	Question                           string    `json:"question" validate:"required"`
	Answer                             string    `json:"answer" validate:"required"`
	MeanOfQuestion                     string    `json:"mean_of_question" validate:"required"`
	MeanOfAnswer                       string    `json:"mean_of_answer" validate:"required"`
	Explain                            string    `json:"explain" validate:"required"`
}

type UpdateSpeakingConversationalRepetitionQARequest struct {
	SpeakingConversationalRepetitionQAID uuid.UUID `json:"speaking_conversational_repetition_qa_id" validate:"required"`
	Field                                string    `json:"field" validate:"required,oneof=question answer mean_of_question mean_of_answer explain"`
	Value                                string    `json:"value" validate:"required"`
}

// ! ------------------------------------------------------------------------------
// ! Conversational Open
// ! ------------------------------------------------------------------------------
type SpeakingConversationalOpenResponse struct {
	ID                  uuid.UUID `json:"id"`
	Title               string    `json:"title"`
	Overview            string    `json:"overview"`
	ExampleConversation string    `json:"example_conversation"`
}

type CreateSpeakingConversationalOpenRequest struct {
	SpeakingQuestionID  uuid.UUID `json:"speaking_question_id" validate:"required"`
	Title               string    `json:"title" validate:"required"`
	Overview            string    `json:"overview" validate:"required"`
	ExampleConversation string    `json:"example_conversation" validate:"required"`
}

type UpdateSpeakingConversationalOpenRequest struct {
	SpeakingConversationalOpenID uuid.UUID `json:"speaking_conversational_open_id" validate:"required"`
	Field                        string    `json:"field" validate:"required,oneof=title overview example_conversation"`
	Value                        string    `json:"value" validate:"required"`
}
