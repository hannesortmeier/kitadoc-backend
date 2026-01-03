package models

import (
	_ "github.com/go-playground/validator/v10"
)

// Process represents the response from a request to kitadoc-audioproc.
type Process struct {
	ProcessId int    `json:"process_id"`
	Status    string `json:"status" validate:"required"`
}
