package listening

import (
	"time"

	"github.com/google/uuid"
)

type ListeningFillInTheBlankQuestion struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ListeningQuestionID uuid.UUID `gorm:"type:uuid;not null" json:"listening_question_id"`
	Question            string    `gorm:"type:text;not null" json:"question"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
