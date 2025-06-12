package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// HTTPTestClient provides utilities for testing HTTP endpoints
type HTTPTestClient struct {
	t      *testing.T
	router *gin.Engine
	token  string
}

// NewHTTPTestClient creates a new HTTP test client
func NewHTTPTestClient(t *testing.T) *HTTPTestClient {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	return &HTTPTestClient{
		t:      t,
		router: router,
	}
}

// SetRouter sets the gin router for testing
func (c *HTTPTestClient) SetRouter(router *gin.Engine) *HTTPTestClient {
	c.router = router
	return c
}

// WithAuth sets an authentication token
func (c *HTTPTestClient) WithAuth(token string) *HTTPTestClient {
	c.token = token
	return c
}

// WithUser creates a JWT token for the given user ID
func (c *HTTPTestClient) WithUser(userID int64) *HTTPTestClient {
	jwtManager := TestJWTManager()
	tokenPair, err := jwtManager.GenerateTokenPair(fmt.Sprintf("%d", userID), "testuser", "test@example.com", "player")
	require.NoError(c.t, err)
	c.token = tokenPair.AccessToken
	return c
}

// Request makes an HTTP request and returns the response
func (c *HTTPTestClient) Request(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(c.t, err)
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	w := httptest.NewRecorder()
	c.router.ServeHTTP(w, req)

	return w
}

// GET makes a GET request
func (c *HTTPTestClient) GET(path string) *HTTPTestResponse {
	w := c.Request(http.MethodGet, path, nil)
	return NewHTTPTestResponse(c.t, w)
}

// POST makes a POST request
func (c *HTTPTestClient) POST(path string, body interface{}) *HTTPTestResponse {
	w := c.Request(http.MethodPost, path, body)
	return NewHTTPTestResponse(c.t, w)
}

// PUT makes a PUT request
func (c *HTTPTestClient) PUT(path string, body interface{}) *HTTPTestResponse {
	w := c.Request(http.MethodPut, path, body)
	return NewHTTPTestResponse(c.t, w)
}

// DELETE makes a DELETE request
func (c *HTTPTestClient) DELETE(path string) *HTTPTestResponse {
	w := c.Request(http.MethodDelete, path, nil)
	return NewHTTPTestResponse(c.t, w)
}

// HTTPTestResponse wraps the response for easy assertions
type HTTPTestResponse struct {
	t        *testing.T
	recorder *httptest.ResponseRecorder
	body     []byte
}

// NewHTTPTestResponse creates a new test response wrapper
func NewHTTPTestResponse(t *testing.T, recorder *httptest.ResponseRecorder) *HTTPTestResponse {
	return &HTTPTestResponse{
		t:        t,
		recorder: recorder,
		body:     recorder.Body.Bytes(),
	}
}

// AssertStatus asserts the response status code
func (r *HTTPTestResponse) AssertStatus(expected int) *HTTPTestResponse {
	require.Equal(r.t, expected, r.recorder.Code,
		"Expected status %d, got %d. Body: %s",
		expected, r.recorder.Code, string(r.body))
	return r
}

// AssertOK asserts status 200
func (r *HTTPTestResponse) AssertOK() *HTTPTestResponse {
	return r.AssertStatus(http.StatusOK)
}

// AssertCreated asserts status 201
func (r *HTTPTestResponse) AssertCreated() *HTTPTestResponse {
	return r.AssertStatus(http.StatusCreated)
}

// AssertBadRequest asserts status 400
func (r *HTTPTestResponse) AssertBadRequest() *HTTPTestResponse {
	return r.AssertStatus(http.StatusBadRequest)
}

// AssertUnauthorized asserts status 401
func (r *HTTPTestResponse) AssertUnauthorized() *HTTPTestResponse {
	return r.AssertStatus(http.StatusUnauthorized)
}

// AssertForbidden asserts status 403
func (r *HTTPTestResponse) AssertForbidden() *HTTPTestResponse {
	return r.AssertStatus(http.StatusForbidden)
}

// AssertNotFound asserts status 404
func (r *HTTPTestResponse) AssertNotFound() *HTTPTestResponse {
	return r.AssertStatus(http.StatusNotFound)
}

// AssertHeader asserts a response header value
func (r *HTTPTestResponse) AssertHeader(key, value string) *HTTPTestResponse {
	actual := r.recorder.Header().Get(key)
	require.Equal(r.t, value, actual,
		"Expected header %s to be %s, got %s", key, value, actual)
	return r
}

// AssertJSON asserts the response body matches the expected JSON
func (r *HTTPTestResponse) AssertJSON(expected interface{}) *HTTPTestResponse {
	expectedJSON, err := json.Marshal(expected)
	require.NoError(r.t, err)

	require.JSONEq(r.t, string(expectedJSON), string(r.body))
	return r
}

// AssertJSONPath asserts a specific path in the JSON response
func (r *HTTPTestResponse) AssertJSONPath(path string, expected interface{}) *HTTPTestResponse {
	var data map[string]interface{}
	err := json.Unmarshal(r.body, &data)
	require.NoError(r.t, err)

	// Simple path implementation (can be enhanced with a proper JSON path library)
	actual := data[path]
	require.Equal(r.t, expected, actual)
	return r
}

// AssertBodyContains asserts the response body contains a string
func (r *HTTPTestResponse) AssertBodyContains(substr string) *HTTPTestResponse {
	require.Contains(r.t, string(r.body), substr)
	return r
}

// AssertBodyNotContains asserts the response body does not contain a string
func (r *HTTPTestResponse) AssertBodyNotContains(substr string) *HTTPTestResponse {
	require.NotContains(r.t, string(r.body), substr)
	return r
}

// DecodeJSON decodes the response body into the given interface
func (r *HTTPTestResponse) DecodeJSON(v interface{}) *HTTPTestResponse {
	err := json.Unmarshal(r.body, v)
	require.NoError(r.t, err, "Failed to decode JSON: %s", string(r.body))
	return r
}

// GetBody returns the response body as bytes
func (r *HTTPTestResponse) GetBody() []byte {
	return r.body
}

// GetBodyString returns the response body as a string
func (r *HTTPTestResponse) GetBodyString() string {
	return string(r.body)
}

// HTTPTestCase represents a single HTTP test case
type HTTPTestCase struct {
	Name           string
	Method         string
	Path           string
	Body           interface{}
	Headers        map[string]string
	Auth           bool
	UserID         int64
	ExpectedStatus int
	ExpectedBody   interface{}
	ExpectedError  string
	Setup          func()
	Cleanup        func()
}

// RunHTTPTestCases runs multiple HTTP test cases
func RunHTTPTestCases(t *testing.T, router *gin.Engine, cases []HTTPTestCase) {
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Setup != nil {
				tc.Setup()
			}
			if tc.Cleanup != nil {
				defer tc.Cleanup()
			}

			client := NewHTTPTestClient(t).SetRouter(router)

			if tc.Auth && tc.UserID > 0 {
				client = client.WithUser(tc.UserID)
			}

			// Make request
			resp := client.Request(tc.Method, tc.Path, tc.Body)
			testResp := NewHTTPTestResponse(t, resp)

			// Assert status
			testResp.AssertStatus(tc.ExpectedStatus)

			// Assert body if provided
			if tc.ExpectedBody != nil {
				testResp.AssertJSON(tc.ExpectedBody)
			}

			// Assert error message if provided
			if tc.ExpectedError != "" {
				testResp.AssertBodyContains(tc.ExpectedError)
			}
		})
	}
}

// MockHTTPContext creates a mock gin context for unit testing
func MockHTTPContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	var reqBody io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	return c, w
}

// SetContextUser sets a user ID in the gin context
func SetContextUser(c *gin.Context, userID int64) {
	c.Set("user_id", userID)
}

// SetContextValue sets a value in the gin context
func SetContextValue(c *gin.Context, key string, value interface{}) {
	c.Set(key, value)
}

// AssertErrorResponse asserts an error response structure
func AssertErrorResponseWithCode(t *testing.T, w *httptest.ResponseRecorder, expectedCode string, expectedStatus int) {
	require.Equal(t, expectedStatus, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	require.Equal(t, expectedCode, response["code"])
	require.NotEmpty(t, response["message"])
}
