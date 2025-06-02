package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Group represents a group in the system.
type Group struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" validate:"required,min=2,max=100"` // Unique handled by DB, but required for feedback
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ValidateGroup validates the Group struct.
func ValidateGroup(group Group) error {
	validate := validator.New()
	return validate.Struct(group)
}
