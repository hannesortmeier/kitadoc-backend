package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Category represents a category for documentation entries.
type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name" validate:"required,min=2,max=100"` // Unique handled by DB, but required for feedback
	Description *string   `json:"description"`                            // Pointer for nullable field
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ValidateCategory validates the Category struct.
func ValidateCategory(category Category) error {
	validate := validator.New()
	return validate.Struct(category)
}

// StringPtr returns a pointer to the string value.
func StringPtr(s string) *string {
	return &s
}
