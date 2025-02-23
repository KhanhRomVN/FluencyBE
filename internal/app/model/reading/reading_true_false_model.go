package reading

import (
	"time"

	"github.com/google/uuid"
)

type ReadingTrueFalse struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ReadingQuestionID uuid.UUID `gorm:"type:uuid;not null" json:"reading_question_id"`
	Question          string    `gorm:"type:text;not null" json:"question"`
	Answer            string    `gorm:"type:text;not null" json:"answer"`
	Explain           string    `gorm:"type:text;not null" json:"explain"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
