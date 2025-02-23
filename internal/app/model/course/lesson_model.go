package course

import (
	"time"

	"github.com/google/uuid"
)

type Lesson struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CourseID  uuid.UUID `gorm:"type:uuid;not null" json:"course_id"`
	Sequence  int       `gorm:"not null" json:"sequence"`
	Title     string    `gorm:"type:text;not null" json:"title"`
	Overview  string    `gorm:"type:text;not null" json:"overview"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Course    *Course   `gorm:"foreignKey:CourseID" json:"course,omitempty"`

	_ struct{} `gorm:"uniqueIndex:unique_lesson_sequence,composite:course_id,sequence"`
	_ struct{} `gorm:"uniqueIndex:unique_lesson_title_per_course,composite:course_id,title"`
}
