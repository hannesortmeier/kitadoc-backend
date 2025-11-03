package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Child represents a child in the system.
type Child struct {
	ID                       int       `json:"id"`
	FirstName                string    `json:"first_name" validate:"required,min=1,max=100" pii:"true"`
	LastName                 string    `json:"last_name" validate:"required,min=1,max=100" pii:"true"`
	Birthdate                time.Time `json:"birthdate" validate:"required,childbirthdate" pii:"true"`
	Gender                   string    `json:"gender" validate:"required"`
	FamilyLanguage           string    `json:"family_language" validate:"required"`
	MigrationBackground      bool      `json:"migration_background"`
	AdmissionDate            time.Time `json:"admission_date" validate:"required"`
	ExpectedSchoolEnrollment time.Time `json:"expected_school_enrollment" validate:"gtfield=Birthdate"`
	Address                  string    `json:"address" validate:"required" pii:"true"`
	Parent1Name              string    `json:"parent1_name" validate:"required,min=1,max=200" pii:"true"`
	Parent2Name              string    `json:"parent2_name" validate:"required,min=1,max=200" pii:"true"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// ChildDB is a struct that matches the children table in the database.
// PII fields are stored as encrypted strings.
type ChildDB struct {
	ID                       int
	FirstName                string
	LastName                 string
	Birthdate                string
	Gender                   string
	FamilyLanguage           string
	MigrationBackground      bool
	AdmissionDate            time.Time
	ExpectedSchoolEnrollment time.Time
	Address                  string
	Parent1Name              string
	Parent2Name              string
	CreatedAt                time.Time
	UpdatedAt                time.Time
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
