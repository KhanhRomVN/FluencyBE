package speaking

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type SpeakingQuestion struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Type        string         `gorm:"type:speaking_question_type;not null" json:"type"`
	Topic       pq.StringArray `gorm:"type:varchar(100)[];not null" json:"topic"`
	Instruction string         `gorm:"type:text;not null" json:"instruction"`
	ImageURLs   pq.StringArray `gorm:"type:text[];not null" json:"image_urls"`
	MaxTime     int            `gorm:"not null" json:"max_time"`
	Version     int            `gorm:"not null;default:1" json:"version"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}
