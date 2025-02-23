package listening

import (
	"time"

	"github.com/google/uuid"
)

type ListeningFillInTheBlankAnswer struct {
	ID                                uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ListeningFillInTheBlankQuestionID uuid.UUID `gorm:"type:uuid;not null" json:"listening_fill_in_the_blank_question_id"`
	Answer                            string    `gorm:"type:text;not null" json:"answer"`
	Explain                           string    `gorm:"size:255;not null" json:"explain"`
	CreatedAt                         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
