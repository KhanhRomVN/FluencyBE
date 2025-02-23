package course

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type CourseBook struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CourseID        uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:unique_course_book" json:"course_id"`
	Publishers      pq.StringArray `gorm:"type:text[];not null" json:"publishers"`
	Authors         pq.StringArray `gorm:"type:text[];not null" json:"authors"`
	PublicationYear int            `gorm:"not null" json:"publication_year"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	Course          *Course        `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}
