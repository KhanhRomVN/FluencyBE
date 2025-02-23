package writing

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type WritingSentenceCompletion struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	WritingQuestionID uuid.UUID      `gorm:"type:uuid;not null" json:"writing_question_id"`
	ExampleSentence   string         `gorm:"type:text;not null" json:"example_sentence"`
	GivenPartSentence string         `gorm:"type:text;not null" json:"given_part_sentence"`
	Position          string         `gorm:"type:varchar(10);not null" json:"position"`
	RequiredWords     pq.StringArray `gorm:"type:text[];not null" json:"required_words"`
	Explain           string         `gorm:"type:text;not null" json:"explain"`
	MinWords          int            `gorm:"not null" json:"min_words"`
	MaxWords          int            `gorm:"not null" json:"max_words"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}
