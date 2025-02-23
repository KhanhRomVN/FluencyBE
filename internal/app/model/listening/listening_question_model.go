package listening

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ListeningQuestion struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Type        string         `gorm:"type:varchar(50);not null" json:"type"`
	Topic       pq.StringArray `gorm:"type:varchar(100)[];not null" json:"topic"`
	Instruction string         `gorm:"type:text;not null" json:"instruction"`
	AudioURLs   pq.StringArray `gorm:"type:text[];not null" json:"audio_urls"`
	ImageURLs   pq.StringArray `gorm:"type:text[];not null" json:"image_urls"`
	Transcript  string         `gorm:"type:text;not null" json:"transcript"`
	MaxTime     int            `gorm:"not null" json:"max_time"`
	Version     int            `gorm:"not null;default:1" json:"version"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
