package handlers

// Common error messages used across handlers
const (
	// Request validation errors
	ErrInvalidRequestBody  = "Invalid request body"
	ErrInvalidSessionID    = "Invalid session ID"
	ErrInvalidCharacterID  = "Invalid character ID format"
	ErrInvalidUserID       = "Invalid user ID"
	ErrInvalidRaceID       = "Invalid race ID"
	
	// Not found errors
	ErrSessionNotFound = "Session not found"
	ErrCultureNotFound = "Culture not found"
	
	// Validation messages
	ErrSessionNameRequired = "session name is required"
)