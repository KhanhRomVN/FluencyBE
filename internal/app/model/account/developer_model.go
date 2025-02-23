package account

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidUsername = errors.New("username must be 3-50 characters and alphanumeric")
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
)

type Developer struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Username  string    `gorm:"uniqueIndex;not null;size:255" json:"username"`
	Password  string    `gorm:"not null;type:text" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (d *Developer) Validate() error {
	if d.Email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(d.Email) {
		return ErrInvalidEmail
	}

	if len(d.Username) < 3 || len(d.Username) > 50 {
		return ErrInvalidUsername
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !usernameRegex.MatchString(d.Username) {
		return ErrInvalidUsername
	}

	return nil
}
