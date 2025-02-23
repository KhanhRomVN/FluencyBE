package dto

import (
	"time"

	"github.com/google/uuid"
)

// Word DTOs
type WikiWordResponse struct {
	ID            uuid.UUID `json:"id"`
	Word          string    `json:"word"`
	Pronunciation string    `json:"pronunciation"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateWikiWordRequest struct {
	Word          string `json:"word" validate:"required,min=1,max=100"`
	Pronunciation string `json:"pronunciation" validate:"required"`
}

type UpdateWikiWordRequest struct {
	Word          *string `json:"word,omitempty" validate:"omitempty,min=1,max=100"`
	Pronunciation *string `json:"pronunciation,omitempty" validate:"omitempty"`
}

// Word Definition DTOs
type WikiWordDefinitionResponse struct {
	ID               uuid.UUID   `json:"id"`
	WikiWordID       uuid.UUID   `json:"wiki_word_id"`
	Means            []string    `json:"means"`
	IsMainDefinition bool        `json:"is_main_definition"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	WikiWord         interface{} `json:"wiki_word,omitempty"`
}

type CreateWikiWordDefinitionRequest struct {
	WikiWordID       uuid.UUID `json:"wiki_word_id" validate:"required"`
	Means            []string  `json:"means" validate:"required,min=1"`
	IsMainDefinition bool      `json:"is_main_definition"`
}

type UpdateWikiWordDefinitionRequest struct {
	Means            []string `json:"means,omitempty" validate:"omitempty,min=1"`
	IsMainDefinition *bool    `json:"is_main_definition,omitempty"`
}

// Word Definition Sample DTOs
type WikiWordDefinitionSampleResponse struct {
	ID                   uuid.UUID   `json:"id"`
	WikiWordDefinitionID uuid.UUID   `json:"wiki_word_definition_id"`
	SampleSentence       string      `json:"sample_sentence"`
	SampleSentenceMean   string      `json:"sample_sentence_mean"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
	WikiWordDefinition   interface{} `json:"wiki_word_definition,omitempty"`
}

type CreateWikiWordDefinitionSampleRequest struct {
	WikiWordDefinitionID uuid.UUID `json:"wiki_word_definition_id" validate:"required"`
	SampleSentence       string    `json:"sample_sentence" validate:"required"`
	SampleSentenceMean   string    `json:"sample_sentence_mean" validate:"required"`
}

type UpdateWikiWordDefinitionSampleRequest struct {
	SampleSentence     *string `json:"sample_sentence,omitempty" validate:"omitempty"`
	SampleSentenceMean *string `json:"sample_sentence_mean,omitempty" validate:"omitempty"`
}

// Word Synonym DTOs
type WikiWordSynonymResponse struct {
	ID                   uuid.UUID   `json:"id"`
	WikiWordDefinitionID uuid.UUID   `json:"wiki_word_definition_id"`
	WikiSynonymID        uuid.UUID   `json:"wiki_synonym_id"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
	WikiWordDefinition   interface{} `json:"wiki_word_definition,omitempty"`
	WikiSynonym          interface{} `json:"wiki_synonym,omitempty"`
}

type CreateWikiWordSynonymRequest struct {
	WikiWordDefinitionID uuid.UUID `json:"wiki_word_definition_id" validate:"required"`
	WikiSynonymID        uuid.UUID `json:"wiki_synonym_id" validate:"required"`
}

// Word Antonym DTOs
type WikiWordAntonymResponse struct {
	ID                   uuid.UUID   `json:"id"`
	WikiWordDefinitionID uuid.UUID   `json:"wiki_word_definition_id"`
	WikiAntonymID        uuid.UUID   `json:"wiki_antonym_id"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
	WikiWordDefinition   interface{} `json:"wiki_word_definition,omitempty"`
	WikiAntonym          interface{} `json:"wiki_antonym,omitempty"`
}

type CreateWikiWordAntonymRequest struct {
	WikiWordDefinitionID uuid.UUID `json:"wiki_word_definition_id" validate:"required"`
	WikiAntonymID        uuid.UUID `json:"wiki_antonym_id" validate:"required"`
}

// Phrase DTOs
type WikiPhraseResponse struct {
	ID              uuid.UUID `json:"id"`
	Phrase          string    `json:"phrase"`
	Type            string    `json:"type"`
	DifficultyLevel int       `json:"difficulty_level"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateWikiPhraseRequest struct {
	Phrase          string `json:"phrase" validate:"required,min=2,max=255"`
	Type            string `json:"type" validate:"required,max=25"`
	DifficultyLevel int    `json:"difficulty_level" validate:"required,min=1,max=5"`
}

type UpdateWikiPhraseRequest struct {
	Phrase          *string `json:"phrase,omitempty" validate:"omitempty,min=2,max=255"`
	Type            *string `json:"type,omitempty" validate:"omitempty,max=25"`
	DifficultyLevel *int    `json:"difficulty_level,omitempty" validate:"omitempty,min=1,max=5"`
}

// Phrase Definition DTOs
type WikiPhraseDefinitionResponse struct {
	ID               uuid.UUID   `json:"id"`
	WikiPhraseID     uuid.UUID   `json:"wiki_phrase_id"`
	Mean             string      `json:"mean"`
	IsMainDefinition bool        `json:"is_main_definition"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	WikiPhrase       interface{} `json:"wiki_phrase,omitempty"`
}

type CreateWikiPhraseDefinitionRequest struct {
	WikiPhraseID     uuid.UUID `json:"wiki_phrase_id" validate:"required"`
	Mean             string    `json:"mean" validate:"required"`
	IsMainDefinition bool      `json:"is_main_definition"`
}

type UpdateWikiPhraseDefinitionRequest struct {
	Mean             *string `json:"mean,omitempty" validate:"omitempty"`
	IsMainDefinition *bool   `json:"is_main_definition,omitempty"`
}

// Phrase Definition Sample DTOs
type WikiPhraseDefinitionSampleResponse struct {
	ID                     uuid.UUID   `json:"id"`
	WikiPhraseDefinitionID uuid.UUID   `json:"wiki_phrase_definition_id"`
	SampleSentence         string      `json:"sample_sentence"`
	SampleSentenceMean     string      `json:"sample_sentence_mean"`
	CreatedAt              time.Time   `json:"created_at"`
	UpdatedAt              time.Time   `json:"updated_at"`
	WikiPhraseDefinition   interface{} `json:"wiki_phrase_definition,omitempty"`
}

type CreateWikiPhraseDefinitionSampleRequest struct {
	WikiPhraseDefinitionID uuid.UUID `json:"wiki_phrase_definition_id" validate:"required"`
	SampleSentence         string    `json:"sample_sentence" validate:"required"`
	SampleSentenceMean     string    `json:"sample_sentence_mean" validate:"required"`
}

type UpdateWikiPhraseDefinitionSampleRequest struct {
	SampleSentence     *string `json:"sample_sentence,omitempty" validate:"omitempty"`
	SampleSentenceMean *string `json:"sample_sentence_mean,omitempty" validate:"omitempty"`
}

// Common Query Parameters
type WikiSearchParams struct {
	Query     string `form:"query"`
	Page      int    `form:"page" binding:"required,min=1"`
	PageSize  int    `form:"page_size" binding:"required,min=1,max=100"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

type WikiPaginationResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Data     interface{} `json:"data"`
}
