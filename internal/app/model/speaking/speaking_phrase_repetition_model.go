package speaking

import (
	"time"

	"github.com/google/uuid"
)

type SpeakingPhraseRepetition struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SpeakingQuestionID uuid.UUID `gorm:"type:uuid;not null" json:"speaking_question_id"`
	Phrase             string    `gorm:"type:text;not null" json:"phrase"`
	Mean               string    `gorm:"type:text;not null" json:"mean"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
