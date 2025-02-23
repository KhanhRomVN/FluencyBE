package reading

import (
	"time"

	"github.com/google/uuid"
)

type ReadingFillInTheBlankAnswer struct {
	ID                              uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ReadingFillInTheBlankQuestionID uuid.UUID `gorm:"type:uuid;not null" json:"reading_fill_in_the_blank_question_id"`
	Answer                          string    `gorm:"type:text;not null" json:"answer"`
	Explain                         string    `gorm:"type:text;not null" json:"explain"`
	CreatedAt                       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
