package speaking

import (
	"time"

	"github.com/google/uuid"
)

type SpeakingOpenParagraph struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SpeakingQuestionID   uuid.UUID `gorm:"type:uuid;not null" json:"speaking_question_id"`
	Question             string    `gorm:"type:text;not null" json:"question"`
	ExamplePassage       string    `gorm:"type:text;not null" json:"example_passage"`
	MeanOfExamplePassage string    `gorm:"type:text;not null" json:"mean_of_example_passage"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
