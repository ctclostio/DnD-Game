package testhelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
)

// HTTPTestCase represents a standard HTTP test case structure
type HTTPTestCase struct {
	Name           string
	Method         string
	Path           string
	Body           interface{}
	PathVars       map[string]string
	Headers        map[string]string
	UserID         string
	Role           string
	ExpectedStatus int
	ExpectedError  string
	ValidateBody   func(*testing.T, map[string]interface{})
}

// CreateTestRequest creates a new HTTP test request with JSON body
func CreateTestRequest(method, path string, body interface{}) *http.Request {
	var bodyReader *bytes.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// CreateAuthenticatedRequest creates a request with authentication context
func CreateAuthenticatedRequest(method, path string, body interface{}, userID, role string) *http.Request {
	req := CreateTestRequest(method, path, body)

	claims := &auth.Claims{
		UserID:   userID,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     role,
		Type:     auth.AccessToken,
	}

	ctx := context.WithValue(req.Context(), auth.UserContextKey, claims)
	return req.WithContext(ctx)
}

// SetPathVars adds gorilla/mux path variables to a request
func SetPathVars(req *http.Request, vars map[string]string) *http.Request {
	if len(vars) > 0 {
		return mux.SetURLVars(req, vars)
	}
	return req
}

// DecodeResponseBody decodes the response body into a map
func DecodeResponseBody(t *testing.T, recorder *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()

	var result map[string]interface{}
	err := json.NewDecoder(recorder.Body).Decode(&result)
	require.NoError(t, err, "Failed to decode response body")

	return result
}

// DecodeResponseBodyInto decodes the response body into a provided struct
func DecodeResponseBodyInto(t *testing.T, recorder *httptest.ResponseRecorder, v interface{}) {
	t.Helper()

	err := json.NewDecoder(recorder.Body).Decode(v)
	require.NoError(t, err, "Failed to decode response body")
}

// AssertJSONResponse checks status code and optionally validates response body
func AssertJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()

	assert.Equal(t, expectedStatus, recorder.Code, "Unexpected status code")

	if recorder.Body.Len() > 0 {
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"),
			"Expected JSON content type")
	}
}

// AssertErrorResponse validates an error response
func AssertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int, expectedError string) {
	t.Helper()

	assert.Equal(t, expectedStatus, recorder.Code)

	if expectedError != "" {
		body := DecodeResponseBody(t, recorder)
		assert.Contains(t, body["error"], expectedError)
	}
}

// ExecuteTestCase runs a complete HTTP test case
func ExecuteTestCase(t *testing.T, tc HTTPTestCase, handler http.HandlerFunc) {
	t.Helper()

	// Create request
	var req *http.Request
	if tc.UserID != "" {
		req = CreateAuthenticatedRequest(tc.Method, tc.Path, tc.Body, tc.UserID, tc.Role)
	} else {
		req = CreateTestRequest(tc.Method, tc.Path, tc.Body)
	}

	// Add path variables
	req = SetPathVars(req, tc.PathVars)

	// Add custom headers
	for key, value := range tc.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	recorder := httptest.NewRecorder()
	handler(recorder, req)

	// Validate response
	AssertJSONResponse(t, recorder, tc.ExpectedStatus)

	if tc.ExpectedError != "" {
		AssertErrorResponse(t, recorder, tc.ExpectedStatus, tc.ExpectedError)
	} else if tc.ValidateBody != nil && recorder.Code < 400 {
		body := DecodeResponseBody(t, recorder)
		tc.ValidateBody(t, body)
	}
}

// RunHTTPTestCases executes a slice of test cases
func RunHTTPTestCases(t *testing.T, testCases []HTTPTestCase, handler http.HandlerFunc) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ExecuteTestCase(t, tc, handler)
		})
	}
}

// CreateMultipartRequest creates a multipart form request for file uploads
func CreateMultipartRequest(method, path string, fields map[string]string, files map[string][]byte) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add fields
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, err
		}
	}

	// Add files
	for fieldname, content := range files {
		part, err := writer.CreateFormFile(fieldname, "test.file")
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(content); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}
