package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationMiddleware_Validate(t *testing.T) {
	vm := NewValidationMiddleware()

	// Test struct for validation
	type CreateCharacterRequest struct {
		Name  string `json:"name" validate:"required,min=1,max=50"`
		Race  string `json:"race" validate:"required"`
		Class string `json:"class" validate:"required"`
	}

	tests := []struct {
		name           string
		method         string
		body           interface{}
		targetStruct   interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "valid POST request",
			method: http.MethodPost,
			body: CreateCharacterRequest{
				Name:  "Aragorn",
				Race:  "Human",
				Class: "Fighter",
			},
			targetStruct:   &CreateCharacterRequest{},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "invalid POST request - missing required field",
			method: http.MethodPost,
			body: map[string]interface{}{
				"race":  "Human",
				"class": "Fighter",
				// Missing name
			},
			targetStruct:   &CreateCharacterRequest{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "GET request - no validation",
			method: http.MethodGet,
			body:   nil,
			targetStruct:   &CreateCharacterRequest{},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "PUT request with valid data",
			method: http.MethodPut,
			body: CreateCharacterRequest{
				Name:  "Legolas",
				Race:  "Elf",
				Class: "Ranger",
			},
			targetStruct:   &CreateCharacterRequest{},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var body []byte
			var err error
			if tt.body != nil {
				body, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(tt.method, "/test", bytes.NewReader(body))
			if len(body) > 0 {
				req.Header.Set("Content-Type", "application/json")
			}

			// Create response recorder
			rec := httptest.NewRecorder()

			// Create handler with validation middleware
			handler := vm.Validate(tt.targetStruct)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}))

			// Execute request
			handler.ServeHTTP(rec, req)

			// Check response
			if tt.expectedError {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.Contains(t, rec.Body.String(), "error")
			} else {
				assert.Equal(t, http.StatusOK, rec.Code)
			}
		})
	}
}

func TestValidationMiddleware_StructValidation(t *testing.T) {
	vm := NewValidationMiddleware()

	type TestRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Age      int    `json:"age" validate:"required,min=18,max=100"`
		Username string `json:"username" validate:"required,alphanum,min=3,max=20"`
	}

	tests := []struct {
		name          string
		request       TestRequest
		expectedError bool
		errorContains string
	}{
		{
			name: "valid request",
			request: TestRequest{
				Email:    "test@example.com",
				Age:      25,
				Username: "testuser123",
			},
			expectedError: false,
		},
		{
			name: "invalid email",
			request: TestRequest{
				Email:    "not-an-email",
				Age:      25,
				Username: "testuser123",
			},
			expectedError: true,
			errorContains: "email",
		},
		{
			name: "age too young",
			request: TestRequest{
				Email:    "test@example.com",
				Age:      17,
				Username: "testuser123",
			},
			expectedError: true,
			errorContains: "age",
		},
		{
			name: "username too short",
			request: TestRequest{
				Email:    "test@example.com",
				Age:      25,
				Username: "ab",
			},
			expectedError: true,
			errorContains: "username",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler := vm.Validate(&TestRequest{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rec, req)

			if tt.expectedError {
				assert.Equal(t, http.StatusBadRequest, rec.Code)
				assert.Contains(t, rec.Body.String(), tt.errorContains)
			} else {
				assert.Equal(t, http.StatusOK, rec.Code)
			}
		})
	}
}

func TestValidationMiddleware_NestedStructValidation(t *testing.T) {
	vm := NewValidationMiddleware()

	type Address struct {
		Street  string `json:"street" validate:"required"`
		City    string `json:"city" validate:"required"`
		ZipCode string `json:"zip_code" validate:"required,len=5,numeric"`
	}

	type UserProfile struct {
		Name    string  `json:"name" validate:"required"`
		Email   string  `json:"email" validate:"required,email"`
		Address Address `json:"address" validate:"required"`
	}

	validProfile := UserProfile{
		Name:  "John Doe",
		Email: "john@example.com",
		Address: Address{
			Street:  "123 Main St",
			City:    "New York",
			ZipCode: "10001",
		},
	}

	body, err := json.Marshal(validProfile)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := vm.Validate(&UserProfile{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}