package wiki

import (
	"time"

	"github.com/google/uuid"
)

type WikiWordAntonym struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WikiWordDefinitionID uuid.UUID `gorm:"type:uuid;not null" json:"wiki_word_definition_id"`
	WikiAntonymID        uuid.UUID `gorm:"type:uuid;not null" json:"wiki_antonym_id"`
	CreatedAt            time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	WikiWordDefinition *WikiWordDefinition `gorm:"foreignKey:WikiWordDefinitionID" json:"wiki_word_definition,omitempty"`
	WikiAntonym        *WikiWord           `gorm:"foreignKey:WikiAntonymID" json:"wiki_antonym,omitempty"`
}
