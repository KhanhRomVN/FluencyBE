package grammar

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type GrammarQuestionType string

const (
	FillInTheBlank         GrammarQuestionType = "FILL_IN_THE_BLANK"
	ChoiceOne              GrammarQuestionType = "CHOICE_ONE"
	ErrorIdentification    GrammarQuestionType = "ERROR_IDENTIFICATION"
	SentenceTransformation GrammarQuestionType = "SENTENCE_TRANSFORMATION"
)

type GrammarQuestion struct {
	ID          uuid.UUID           `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Type        GrammarQuestionType `gorm:"type:grammar_question_type;not null" json:"type"`
	Topic       pq.StringArray      `gorm:"type:varchar(100)[];not null" json:"topic"`
	Instruction string              `gorm:"type:text;not null" json:"instruction"`
	ImageURLs   pq.StringArray      `gorm:"type:text[];not null;default:'{}'" json:"image_urls"`
	MaxTime     int                 `gorm:"not null" json:"max_time"`
	Version     int                 `gorm:"not null;default:1" json:"version"`
	CreatedAt   time.Time           `gorm:"autoCreateTime;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time           `gorm:"autoUpdateTime;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
