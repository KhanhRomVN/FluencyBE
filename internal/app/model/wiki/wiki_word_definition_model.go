package wiki

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type WikiWordDefinition struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WikiWordID       uuid.UUID      `gorm:"type:uuid;not null" json:"wiki_word_id"`
	Means            pq.StringArray `gorm:"type:text[];not null" json:"means"`
	IsMainDefinition bool           `gorm:"type:boolean;not null;default:false" json:"is_main_definition"`
	CreatedAt        time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	WikiWord *WikiWord `gorm:"foreignKey:WikiWordID" json:"wiki_word,omitempty"`
}
