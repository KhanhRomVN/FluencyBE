package wiki

import (
	"time"

	"github.com/google/uuid"
)

type WikiWord struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Word          string    `gorm:"type:text;not null" json:"word"`
	Pronunciation string    `gorm:"type:text;not null" json:"pronunciation"`
	CreatedAt     time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
