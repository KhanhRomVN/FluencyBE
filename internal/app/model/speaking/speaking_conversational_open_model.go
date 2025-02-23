package speaking

import (
	"time"

	"github.com/google/uuid"
)

type SpeakingConversationalOpen struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SpeakingQuestionID  uuid.UUID `gorm:"type:uuid;not null" json:"speaking_question_id"`
	Title               string    `gorm:"type:text;not null" json:"title"`
	Overview            string    `gorm:"type:text;not null" json:"overview"`
	ExampleConversation string    `gorm:"type:text;not null" json:"example_conversation"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
