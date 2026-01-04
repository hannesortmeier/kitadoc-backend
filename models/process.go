package models

import (
	_ "github.com/go-playground/validator/v10"
	"time"
)

// Process represents a audio transcription and analysis process.
type Process struct {
	ProcessId int       `json:"process_id"`
	Status    string    `json:"status" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
}
