package course

import (
	"time"

	"github.com/google/uuid"
)

type CourseOther struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CourseID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:unique_course_other" json:"course_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Course    *Course   `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}
