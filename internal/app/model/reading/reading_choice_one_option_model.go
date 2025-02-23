package reading

import (
	"time"

	"github.com/google/uuid"
)

type ReadingChoiceOneOption struct {
	ID                         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ReadingChoiceOneQuestionID uuid.UUID `gorm:"type:uuid;not null" json:"reading_choice_one_question_id"`
	Options                    string    `gorm:"type:text;not null" json:"options"`
	IsCorrect                  bool      `gorm:"not null;default:false" json:"is_correct"`
	CreatedAt                  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
