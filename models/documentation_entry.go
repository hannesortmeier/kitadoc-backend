package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// DocumentationEntry represents a behavioral documentation entry.
type DocumentationEntry struct {
	ID                     int       `json:"id"`
	ChildID                int       `json:"child_id" validate:"required"`
	TeacherID              int       `json:"teacher_id" validate:"required"`
	CategoryID             int       `json:"category_id" validate:"required"`
	ObservationDate        time.Time `json:"observation_date" validate:"required,iso8601date"` // Assuming ISO8601 format for date
	ObservationDescription string    `json:"observation_description" validate:"required,min=10"`
	IsApproved             bool      `json:"is_approved"`
	ApprovedByUserID       *int      `json:"approved_by_teacher_id"` // Pointer for nullable foreign key
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// ValidateDocumentationEntry validates the DocumentationEntry struct.
func ValidateDocumentationEntry(entry DocumentationEntry) error {
	validate := validator.New()
	validate.RegisterValidation("iso8601date", ValidateISO8601Date) //nolint:errcheck
	return validate.Struct(entry)
}

// ValidateISO8601Date is a custom validator for ISO8601 date format.
// This is a placeholder; actual ISO8601 validation might be more complex
// depending on the exact format expected (e.g., "YYYY-MM-DD").
// For simplicity, this just checks if it's a valid time.Time and not in the future.
func ValidateISO8601Date(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	// Check if the date is not in the future
	return !date.After(time.Now())
}
