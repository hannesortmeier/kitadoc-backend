package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Assignment represents an assignment of a child to a teacher.
type Assignment struct {
	ID        int        `json:"id"`
	ChildID   int        `json:"child_id" validate:"required"`
	TeacherID int        `json:"teacher_id" validate:"required"`
	StartDate time.Time  `json:"start_date" validate:"required"`
	EndDate   *time.Time `json:"end_date" validate:"omitempty,gtfield=StartDate"` // Optional, but if present, must be after StartDate
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ValidateAssignment validates the Assignment struct.
func ValidateAssignment(assignment Assignment) error {
	validate := validator.New()
	return validate.Struct(assignment)
}