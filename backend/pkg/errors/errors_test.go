package errors

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("Invalid input")

	assert.Equal(t, ErrorTypeValidation, err.Type)
	assert.Equal(t, "Invalid input", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	assert.Nil(t, err.Internal)
}

func TestNewAuthenticationError(t *testing.T) {
	err := NewAuthenticationError("Invalid token")

	assert.Equal(t, ErrorTypeAuthentication, err.Type)
	assert.Equal(t, "Invalid token", err.Message)
	assert.Equal(t, http.StatusUnauthorized, err.StatusCode)
}

func TestNewAuthorizationError(t *testing.T) {
	err := NewAuthorizationError("Access denied")

	assert.Equal(t, ErrorTypeAuthorization, err.Type)
	assert.Equal(t, "Access denied", err.Message)
	assert.Equal(t, http.StatusForbidden, err.StatusCode)
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("User")

	assert.Equal(t, ErrorTypeNotFound, err.Type)
	assert.Equal(t, "User not found", err.Message)
	assert.Equal(t, http.StatusNotFound, err.StatusCode)
}

func TestNewInternalError(t *testing.T) {
	originalErr := assert.AnError
	err := NewInternalError("Something went wrong", originalErr)

	assert.Equal(t, ErrorTypeInternal, err.Type)
	assert.Equal(t, "Something went wrong", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, originalErr, err.Internal)
}

func TestAppError_WithDetails(t *testing.T) {
	err := NewValidationError("Validation failed")
	details := map[string]interface{}{
		"field":  "email",
		"reason": "invalid format",
	}

	err.WithDetails(details)

	assert.Equal(t, details, err.Details)
}

func TestAppError_WithCode(t *testing.T) {
	err := NewNotFoundError("Resource")
	err.WithCode("RESOURCE_404")

	assert.Equal(t, "RESOURCE_404", err.Code)
}

func TestAppError_WithInternal(t *testing.T) {
	err := NewBadRequestError("Bad request")
	internalErr := assert.AnError

	err.WithInternal(internalErr)

	assert.Equal(t, internalErr, err.Internal)
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name:     "without internal error",
			err:      NewValidationError("Invalid email"),
			expected: "VALIDATION_ERROR: Invalid email",
		},
		{
			name:     "with internal error",
			err:      NewInternalError("Database error", assert.AnError),
			expected: "INTERNAL_ERROR: Database error (internal: assert.AnError general error for testing)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestAppError_ToJSON(t *testing.T) {
	err := NewValidationError("Invalid input").
		WithCode("VAL001").
		WithDetails(map[string]interface{}{
			"field": "email",
			"value": "invalid",
		})

	jsonBytes := err.ToJSON()

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(jsonBytes, &result))

	assert.Equal(t, string(ErrorTypeValidation), result["type"])
	assert.Equal(t, "Invalid input", result["message"])
	assert.Equal(t, "VAL001", result["code"])
	assert.NotNil(t, result["details"])
}

func TestIsAppError(t *testing.T) {
	appErr := NewValidationError("test")
	normalErr := assert.AnError

	assert.True(t, IsAppError(appErr))
	assert.False(t, IsAppError(normalErr))
}

func TestGetAppError(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		expected *AppError
	}{
		{
			name:     "app error input",
			input:    NewValidationError("test"),
			expected: NewValidationError("test"),
		},
		{
			name:  "normal error input",
			input: assert.AnError,
			expected: &AppError{
				Type:       ErrorTypeInternal,
				Message:    "An unexpected error occurred",
				StatusCode: http.StatusInternalServerError,
				Internal:   assert.AnError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAppError(tt.input)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Message, result.Message)
			assert.Equal(t, tt.expected.StatusCode, result.StatusCode)
		})
	}
}

func TestValidationErrors(t *testing.T) {
	ve := &ValidationErrors{}

	// Test adding errors
	ve.Add("email", "is required")
	ve.Add("email", "must be valid format")
	ve.Add("password", "is too short")

	assert.True(t, ve.HasErrors())
	assert.Len(t, ve.Errors, 2)
	assert.Len(t, ve.Errors["email"], 2)
	assert.Len(t, ve.Errors["password"], 1)

	// Test error string
	errStr := ve.Error()
	assert.Contains(t, errStr, "email")
	assert.Contains(t, errStr, "password")
}

func TestValidationErrors_ToAppError(t *testing.T) {
	ve := &ValidationErrors{}

	// Test with no errors
	assert.Nil(t, ve.ToAppError())

	// Test with errors
	ve.Add("name", "is required")
	ve.Add("age", "must be positive")

	appErr := ve.ToAppError()
	require.NotNil(t, appErr)

	assert.Equal(t, ErrorTypeValidation, appErr.Type)
	assert.Equal(t, "Validation failed", appErr.Message)
	assert.Equal(t, http.StatusBadRequest, appErr.StatusCode)
	assert.NotNil(t, appErr.Details)

	// Check details
	nameErrors, ok := appErr.Details["name"].([]string)
	require.True(t, ok)
	assert.Contains(t, nameErrors, "is required")

	ageErrors, ok := appErr.Details["age"].([]string)
	require.True(t, ok)
	assert.Contains(t, ageErrors, "must be positive")
}

func TestGetErrorMessage(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{ErrCodeInvalidCredentials, "Invalid username or password"},
		{ErrCodeUserNotFound, "User not found"},
		{ErrCodeCharacterNotFound, "Character not found"},
		{ErrorCode("UNKNOWN"), "Unknown error"},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			assert.Equal(t, tt.expected, GetErrorMessage(tt.code))
		})
	}
}
