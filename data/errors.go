package data

import "errors"

var (
	ErrNotFound     = errors.New("record not found")
	ErrConflict     = errors.New("record conflict")
	ErrInvalidInput = errors.New("invalid input")
)
