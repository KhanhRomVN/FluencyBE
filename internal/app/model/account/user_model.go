package account

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Username  string    `gorm:"uniqueIndex;not null;size:255" json:"username"`
	Password  string    `gorm:"not null;type:text" json:"-"`
	Type      string    `gorm:"type:ENUM('basic', 'google', 'facebook', 'github');not null" json:"type"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Username == "" {
		return errors.New("username is required")
	}
	if u.Password == "" {
		return errors.New("password is required")
	}
	if u.Type == "" {
		return errors.New("user type is required")
	}
	return nil
}
