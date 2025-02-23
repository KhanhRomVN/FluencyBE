package wiki

import (
	"time"

	"github.com/google/uuid"
)

type WikiPhraseDefinition struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WikiPhraseID     uuid.UUID `gorm:"type:uuid;not null" json:"wiki_phrase_id"`
	Mean             string    `gorm:"type:text;not null" json:"mean"`
	IsMainDefinition bool      `gorm:"type:boolean;not null;default:false" json:"is_main_definition"`
	CreatedAt        time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	WikiPhrase *WikiPhrase `gorm:"foreignKey:WikiPhraseID" json:"wiki_phrase,omitempty"`
}
