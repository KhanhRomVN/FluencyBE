package wiki

import (
	"time"

	"github.com/google/uuid"
)

type WikiPhraseDefinitionSample struct {
	ID                     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WikiPhraseDefinitionID uuid.UUID `gorm:"type:uuid;not null" json:"wiki_phrase_definition_id"`
	SampleSentence         string    `gorm:"type:text;not null" json:"sample_sentence"`
	SampleSentenceMean     string    `gorm:"type:text;not null" json:"sample_sentence_mean"`
	CreatedAt              time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt              time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	WikiPhraseDefinition *WikiPhraseDefinition `gorm:"foreignKey:WikiPhraseDefinitionID" json:"wiki_phrase_definition,omitempty"`
}
