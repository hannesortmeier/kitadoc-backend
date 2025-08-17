package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Child represents a child in the system.
type Child struct {
	ID                       int       `json:"id"`
	FirstName                string    `json:"first_name" validate:"required,min=1,max=100"`
	LastName                 string    `json:"last_name" validate:"required,min=1,max=100"`
	Birthdate                time.Time `json:"birthdate" validate:"required,childbirthdate"` // Custom validation for age range
	Gender                   string    `json:"gender" validate:"required"`
	FamilyLanguage           string    `json:"family_language" validate:"required"`
	MigrationBackground      bool      `json:"migration_background"`
	AdmissionDate            time.Time `json:"admission_date" validate:"required"`
	ExpectedSchoolEnrollment time.Time `json:"expected_school_enrollment" validate:"gtfield=Birthdate"`
	Address                  string    `json:"address" validate:"required"`
	Parent1Name              string    `json:"parent1_name" validate:"required,min=1,max=200"`
	Parent2Name              string    `json:"parent2_name" validate:"required,min=1,max=200"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// ValidateChild validates the Child struct.
func ValidateChild(child Child) error {
	validate := validator.New()
	validate.RegisterValidation("childbirthdate", ValidateChildBirthdate) //nolint:errcheck
	return validate.Struct(child)
}

// ValidateChildBirthdate is a custom validator for child's birthdate.
func ValidateChildBirthdate(fl validator.FieldLevel) bool {
	birthdate, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}

	now := time.Now()
	minBirthdate := now.AddDate(-8, 0, 0) // Max 8 years old
	maxBirthdate := now                   // Must already be born

	// Birthdate must be after minBirthdate
	return birthdate.After(minBirthdate) && birthdate.Before(maxBirthdate)
}
