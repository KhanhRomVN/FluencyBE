package dto

import (
	"time"

	"github.com/google/uuid"
)

//==============================================================================
// * =-=-=-=-=-=-=-=-=-=-=-=-=-= Account Management =-=-=-=-=-=-=-=-=-=-=-=-=-= *
//==============================================================================

//------------------------------------------------------------------------------
// * Authentication DTOs
//------------------------------------------------------------------------------

// ? Registration
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Password string `json:"password" validate:"required,min=8,max=100"`
	Type     string `json:"type" validate:"required,oneof=basic google facebook github"`
}

// ? Login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
	ID    uuid.UUID `json:"id"`
	Token string    `json:"token"`
}

//------------------------------------------------------------------------------
// * User Profile DTOs
//------------------------------------------------------------------------------

// ? User Information
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// ? Profile Updates
type UpdateRequest struct {
	Email    *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50,alphanum"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8,max=100"`
}

type FieldUpdateRequest struct {
	Field string `json:"field" validate:"required,oneof=username email password"`
	Value string `json:"value" validate:"required"`
}

//------------------------------------------------------------------------------
// * Error Handling DTOs
//------------------------------------------------------------------------------

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
