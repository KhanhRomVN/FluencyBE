package course

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Course struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Type      string         `gorm:"type:varchar(50);not null" json:"type"`
	Title     string         `gorm:"type:text;not null;uniqueIndex:unique_course_title" json:"title"`
	Overview  string         `gorm:"type:text;not null" json:"overview"`
	Skills    pq.StringArray `gorm:"type:text[];not null" json:"skills"`
	Band      string         `gorm:"type:text;not null" json:"band"`
	ImageURLs pq.StringArray `gorm:"type:text[];not null" json:"image_urls"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}
