package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// KitaMasterdata represents the master data of the kindergarten.
type KitaMasterdata struct {
	Name        string    `json:"name" validate:"required"`
	Street      string    `json:"street" validate:"required"`
	HouseNumber string    `json:"house_number" validate:"required"`
	PostalCode  string    `json:"postal_code" validate:"required"`
	City        string    `json:"city" validate:"required"`
	PhoneNumber string    `json:"phone_number" validate:"required"`
	Email       string    `json:"email" validate:"required,email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ValidateKitaMasterdata validates the KitaMasterdata struct.
func ValidateKitaMasterdata(data KitaMasterdata) error {
	validate := validator.New()
	return validate.Struct(data)
}
