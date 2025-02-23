package grammar

import (
	"time"

	"github.com/google/uuid"
)

type GrammarSentenceTransformation struct {
	ID                     uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	GrammarQuestionID      uuid.UUID `gorm:"type:uuid;not null" json:"grammar_question_id"`
	OriginalSentence       string    `gorm:"type:text;not null" json:"original_sentence"`
	BeginningWord          string    `gorm:"type:text" json:"beginning_word"`
	ExampleCorrectSentence string    `gorm:"type:text;not null" json:"example_correct_sentence"`
	Explain                string    `gorm:"type:text;not null" json:"explain"`
	CreatedAt              time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
