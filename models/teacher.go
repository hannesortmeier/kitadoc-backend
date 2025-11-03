package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Teacher represents a teacher in the system.
type Teacher struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name" validate:"required,min=1,max=100" pii:"true"`
	LastName  string    `json:"last_name" validate:"required,min=1,max=100" pii:"true"`
	Username  string    `json:"username" validate:"required,min=1,max=100" pii:"true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TeacherDB is a struct that matches the teachers table in the database.
// PII fields are stored as encrypted strings.
type TeacherDB struct {
	ID        int
	FirstName string
	LastName  string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ValidateTeacher validates the Teacher struct.
func ValidateTeacher(teacher Teacher) error {
	validate := validator.New()
	return validate.Struct(teacher)
}
