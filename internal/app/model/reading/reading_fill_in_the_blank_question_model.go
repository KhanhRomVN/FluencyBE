package reading

import (
	"time"

	"github.com/google/uuid"
)

type ReadingFillInTheBlankQuestion struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ReadingQuestionID uuid.UUID `gorm:"type:uuid;not null" json:"reading_question_id"`
	Question          string    `gorm:"type:text;not null" json:"question"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
