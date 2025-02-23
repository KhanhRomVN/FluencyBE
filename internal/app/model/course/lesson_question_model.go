package course

import (
	"time"

	"github.com/google/uuid"
)

type LessonQuestion struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	LessonID     uuid.UUID `gorm:"type:uuid;not null" json:"lesson_id"`
	Sequence     int       `gorm:"not null" json:"sequence"`
	QuestionID   uuid.UUID `gorm:"type:uuid;not null" json:"question_id"`
	QuestionType string    `gorm:"type:varchar(50);not null" json:"question_type"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Lesson       *Lesson   `gorm:"foreignKey:LessonID" json:"lesson,omitempty"`

	_ struct{} `gorm:"uniqueIndex:unique_question_sequence,composite:lesson_id,sequence"`
	_ struct{} `gorm:"uniqueIndex:unique_question_per_lesson,composite:lesson_id,question_id"`
}
