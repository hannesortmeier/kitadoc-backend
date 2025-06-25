package models

import "time"

// AudioRecording represents an audio recording associated with a documentation entry.
type AudioRecording struct {
	ID                   int       `json:"id"`
	DocumentationEntryID int       `json:"documentation_entry_id" validate:"required"`
	FilePath             string    `json:"file_path" validate:"required"`
	DurationSeconds      int       `json:"duration_seconds" validate:"min=1"`
	CreatedAt            time.Time `json:"created_at"`
}
