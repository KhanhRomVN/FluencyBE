package writing

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type WritingEssay struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	WritingQuestionID uuid.UUID      `gorm:"type:uuid;not null" json:"writing_question_id"`
	EssayType         string         `gorm:"type:varchar(50);not null" json:"essay_type"`
	RequiredPoints    pq.StringArray `gorm:"type:text[];not null" json:"required_points"`
	MinWords          int            `gorm:"not null" json:"min_words"`
	MaxWords          int            `gorm:"not null" json:"max_words"`
	SampleEssay       string         `gorm:"type:text;not null" json:"sample_essay"`
	Explain           string         `gorm:"type:text;not null" json:"explain"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}
