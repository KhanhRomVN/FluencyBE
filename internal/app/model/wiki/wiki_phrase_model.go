package wiki

import (
	"time"

	"github.com/google/uuid"
)

type WikiPhrase struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Phrase          string    `gorm:"type:text;not null" json:"phrase"`
	Type            string    `gorm:"type:varchar(25);not null" json:"type"`
	DifficultyLevel int       `gorm:"type:int" json:"difficulty_level"`
	CreatedAt       time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
