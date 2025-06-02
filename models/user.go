package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// User represents a user in the system.
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username" validate:"required,min=3,max=10"` // Unique handled by DB, but required for feedback
	PasswordHash string    `json:"-" validate:"required"` // Exclude from JSON output, required for input
	Role         string    `json:"role" validate:"required,oneof=teacher admin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ValidateUser validates the User struct.
func ValidateUser(user User) error {
	validate := validator.New()
	return validate.Struct(user)
}