package models

import "errors"

// Common errors.
var (
	// ErrNotFound is returned when a requested resource is not found.
	ErrNotFound = errors.New("not found")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrDuplicate is returned when attempting to create a duplicate resource.
	ErrDuplicate = errors.New("duplicate resource")
)
