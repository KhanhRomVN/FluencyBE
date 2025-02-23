package speaking

import (
	"time"

	"github.com/google/uuid"
)

type SpeakingConversationalRepetitionQA struct {
	ID                                 uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SpeakingConversationalRepetitionID uuid.UUID `gorm:"type:uuid;not null" json:"speaking_conversational_repetition_id"`
	Question                           string    `gorm:"type:text;not null" json:"question"`
	Answer                             string    `gorm:"type:text;not null" json:"answer"`
	MeanOfQuestion                     string    `gorm:"type:text;not null" json:"mean_of_question"`
	MeanOfAnswer                       string    `gorm:"type:text;not null" json:"mean_of_answer"`
	Explain                            string    `gorm:"type:text;not null" json:"explain"`
	CreatedAt                          time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                          time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
