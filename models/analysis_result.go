package models

import (
	_ "github.com/go-playground/validator/v10"
)

// AnalysisResult represents the response from a request to kitadoc-audioproc.
type AnalysisResult struct {
	NumberOfEntries int                   `json:"number_of_entries" validate:"required"`
	AnalysisResults []ChildAnalysisObject `json:"analysis_results" validate:"required,dive"`
}

// ChildAnalysisObject represents the analysis result for a child.
type ChildAnalysisObject struct {
	ChildID              int              `json:"child_id" validate:"required"`
	FirstName            string           `json:"first_name" validate:"required"`
	LastName             string           `json:"last_name" validate:"required"`
	TranscriptionSummary string           `json:"transcription_summary" validate:"required"`
	Category             AnalysisCategory `json:"analysis_category" validate:"required"`
}

// Category represents a category in the analysis result.
type AnalysisCategory struct {
	AnalysisCategoryID   int    `json:"category_id" validate:"required"`
	AnalysisCategoryName string `json:"category_name" validate:"required"`
}
