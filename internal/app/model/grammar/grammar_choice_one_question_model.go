package grammar

import (
	"time"

	"github.com/google/uuid"
)

type GrammarChoiceOneQuestion struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	GrammarQuestionID uuid.UUID `gorm:"type:uuid;not null" json:"grammar_question_id"`
	Question          string    `gorm:"type:text;not null" json:"question"`
	Explain           string    `gorm:"type:text;not null" json:"explain"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
