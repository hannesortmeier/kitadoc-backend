package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// User represents a user in the system.
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username" validate:"required,min=3,max=100" pii:"true"` // Unique handled by DB, but required for feedback
	PasswordHash string    `json:"password_hash" validate:"required"`                     // Exclude from JSON output, required for input
	Role         string    `json:"role" validate:"required,oneof=teacher admin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserDB is a struct that matches the users table in the database.
// PII fields are stored as encrypted strings.
type UserDB struct {
	ID           int
	Username     string
	UsernameHMAC string // Needed for lookup
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ValidateUser validates the User struct.
func ValidateUser(user User) error {
	validate := validator.New()
	return validate.Struct(user)
}
