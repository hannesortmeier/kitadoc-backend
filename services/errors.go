package services

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInternal           = errors.New("internal server error")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrChildReportGenerationFailed = errors.New("child report generation failed")
	ErrFileUploadFailed   = errors.New("file upload failed")
	ErrBulkImportFailed   = errors.New("bulk import failed")
)