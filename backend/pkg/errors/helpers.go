package errors

import (
	"database/sql"
	"strings"
)

// Database error helpers

// IsNotFound checks if an error indicates a not found condition
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	// Check for sql.ErrNoRows
	if err == sql.ErrNoRows {
		return true
	}

	// Check for AppError with NotFound type
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == ErrorTypeNotFound
	}

	// Check for common not found error messages
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "no rows") ||
		strings.Contains(errMsg, "does not exist")
}

// IsDuplicate checks if an error indicates a duplicate/conflict condition
func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}

	// Check for AppError with Conflict type
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == ErrorTypeConflict
	}

	// Check for common duplicate error messages (PostgreSQL, MySQL, etc.)
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "duplicate") ||
		strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "already exists") ||
		strings.Contains(errMsg, "violates unique")
}

// IsTimeout checks if an error indicates a timeout condition
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}

	// Check for common timeout error messages
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "deadline exceeded") ||
		strings.Contains(errMsg, "context canceled")
}

// IsPermissionDenied checks if an error indicates a permission/authorization issue
func IsPermissionDenied(err error) bool {
	if err == nil {
		return false
	}

	// Check for AppError with Authorization type
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == ErrorTypeAuthorization
	}

	// Check for common permission error messages
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "permission denied") ||
		strings.Contains(errMsg, "access denied") ||
		strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "unauthorized")
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}

	// Check for AppError with Validation type
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == ErrorTypeValidation
	}

	// Check for ValidationErrors type
	if _, ok := err.(*ValidationErrors); ok {
		return true
	}

	return false
}

// WrapDatabaseError converts common database errors to AppError types
func WrapDatabaseError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Handle not found
	if IsNotFound(err) {
		return NewNotFoundError(operation).WithInternal(err)
	}

	// Handle duplicates
	if IsDuplicate(err) {
		return NewConflictError("Resource already exists").WithInternal(err)
	}

	// Handle timeouts
	if IsTimeout(err) {
		return NewServiceUnavailableError("Database operation timed out").
			WithCode(string(ErrCodeTimeout)).
			WithInternal(err)
	}

	// Default to internal error
	return NewInternalError("Database operation failed", err).
		WithCode(string(ErrCodeDatabaseError))
}

// Error chain helpers

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	// If it's already an AppError, add context to the message
	if appErr, ok := err.(*AppError); ok {
		appErr.Message = message + ": " + appErr.Message
		return appErr
	}

	// Create new internal error
	return NewInternalError(message, err)
}

// Wrapf wraps an error with formatted message
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	message := format
	if len(args) > 0 {
		message = strings.TrimSpace(format) + " " + strings.TrimSpace(args[0].(string))
	}

	return Wrap(err, message)
}

// Cause returns the underlying cause of the error
func Cause(err error) error {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*AppError); ok && appErr.Internal != nil {
		return appErr.Internal
	}

	return err
}
