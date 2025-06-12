package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation indicates a validation error
	ErrorTypeValidation ErrorType = "VALIDATION_ERROR"
	// ErrorTypeAuthorization indicates an authorization error
	ErrorTypeAuthorization ErrorType = "AUTHORIZATION_ERROR"
	// ErrorTypeAuthentication indicates an authentication error
	ErrorTypeAuthentication ErrorType = "AUTHENTICATION_ERROR"
	// ErrorTypeNotFound indicates a resource not found error
	ErrorTypeNotFound ErrorType = "NOT_FOUND"
	// ErrorTypeConflict indicates a conflict error (e.g., duplicate)
	ErrorTypeConflict ErrorType = "CONFLICT"
	// ErrorTypeInternal indicates an internal server error
	ErrorTypeInternal ErrorType = "INTERNAL_ERROR"
	// ErrorTypeRateLimit indicates rate limit exceeded
	ErrorTypeRateLimit ErrorType = "RATE_LIMIT_EXCEEDED"
	// ErrorTypeBadRequest indicates a bad request
	ErrorTypeBadRequest ErrorType = "BAD_REQUEST"
	// ErrorTypeServiceUnavailable indicates service is unavailable
	ErrorTypeServiceUnavailable ErrorType = "SERVICE_UNAVAILABLE"
)

// AppError represents an application error
type AppError struct {
	Type       ErrorType              `json:"type"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
	Internal   error                  `json:"-"` // Internal error not exposed to client
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s (internal: %v)", e.Type, e.Message, e.Internal)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithInternal adds internal error information
func (e *AppError) WithInternal(err error) *AppError {
	e.Internal = err
	return e
}

// WithCode adds an error code
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

// ToJSON converts error to JSON
func (e *AppError) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// Common error constructors

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewAuthenticationError creates an authentication error
func NewAuthenticationError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeAuthentication,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewAuthorizationError creates an authorization error
func NewAuthorizationError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeAuthorization,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// NewInternalError creates an internal error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Internal:   err,
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeRateLimit,
		Message:    message,
		StatusCode: http.StatusTooManyRequests,
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewServiceUnavailableError creates a service unavailable error
func NewServiceUnavailableError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeServiceUnavailable,
		Message:    message,
		StatusCode: http.StatusServiceUnavailable,
	}
}

// IsAppError checks if error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError converts error to AppError if possible
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return NewInternalError("An unexpected error occurred", err)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors map[string][]string `json:"errors"`
}

// Error implements the error interface
func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "validation errors"
	}

	// Create a simple string representation of the errors
	var messages []string
	for field, errs := range v.Errors {
		for _, err := range errs {
			messages = append(messages, fmt.Sprintf("%s: %s", field, err))
		}
	}

	if len(messages) == 1 {
		return messages[0]
	}

	return fmt.Sprintf("validation errors: %v", messages)
}

// Add adds a validation error for a field
func (v *ValidationErrors) Add(field, message string) {
	if v.Errors == nil {
		v.Errors = make(map[string][]string)
	}
	v.Errors[field] = append(v.Errors[field], message)
}

// HasErrors checks if there are any validation errors
func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

// ToAppError converts validation errors to AppError
func (v *ValidationErrors) ToAppError() *AppError {
	if !v.HasErrors() {
		return nil
	}

	details := make(map[string]interface{})
	for field, messages := range v.Errors {
		details[field] = messages
	}

	return NewValidationError("Validation failed").WithDetails(details)
}
