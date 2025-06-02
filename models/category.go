package models

import (
	"github.com/go-playground/validator/v10"
)

// Category represents a category for documentation entries.
type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name" validate:"required,min=2,max=100"` // Unique handled by DB, but required for feedback
	Description *string   `json:"description"`                             // Pointer for nullable field
}

// ValidateCategory validates the Category struct.
func ValidateCategory(category Category) error {
	validate := validator.New()
	return validate.Struct(category)
}