package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Teacher represents a teacher in the system.
type Teacher struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string    `json:"last_name" validate:"required,min=1,max=100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ValidateTeacher validates the Teacher struct.
func ValidateTeacher(teacher Teacher) error {
	validate := validator.New()
	return validate.Struct(teacher)
}
