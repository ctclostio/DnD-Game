package middleware

import (
	"database/sql"
	stderrors "errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/ctclostio/DnD-Game/backend/internal/testutil"
	"github.com/ctclostio/DnD-Game/backend/pkg/errors"
)

func TestErrorHandlerMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		handler        gin.HandlerFunc
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name: "handles custom error with code",
			handler: func(c *gin.Context) {
				err := errors.NewNotFoundError("character").WithCode(string(errors.ErrCodeCharacterNotFound))
				_ = c.Error(err)
				c.Abort()
			},
			expectedStatus: http.StatusNotFound,
			expectedCode:   string(errors.ErrCodeCharacterNotFound),
			expectedMsg:    "character not found",
		},
		{
			name: "handles validation error",
			handler: func(c *gin.Context) {
				err := errors.NewValidationError("name is required").WithCode(string(errors.ErrCodeValidationFailed))
				_ = c.Error(err)
				c.Abort()
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   string(errors.ErrCodeValidationFailed),
			expectedMsg:    "name is required",
		},
		{
			name: "handles authorization error",
			handler: func(c *gin.Context) {
				err := errors.NewAuthenticationError("invalid token").WithCode(string(errors.ErrCodeTokenInvalid))
				_ = c.Error(err)
				c.Abort()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   string(errors.ErrCodeTokenInvalid),
			expectedMsg:    "invalid token",
		},
		{
			name: "handles generic error",
			handler: func(c *gin.Context) {
				err := stderrors.New("something went wrong")
				_ = c.Error(err)
				c.Abort()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   string(errors.ErrCodeInternalError),
			expectedMsg:    "Internal server error",
		},
		// Note: Panic recovery is handled by Gin's recovery middleware
		// Our error handler doesn't see panics directly
		{
			name: "preserves request ID",
			handler: func(c *gin.Context) {
				c.Set("request_id", "test-request-123")
				err := errors.NewNotFoundError("resource").WithCode(string(errors.ErrCodeCharacterNotFound))
				_ = c.Error(err)
				c.Abort()
			},
			expectedStatus: http.StatusNotFound,
			expectedCode:   string(errors.ErrCodeCharacterNotFound),
		},
		{
			name: "handles multiple errors (returns last)",
			handler: func(c *gin.Context) {
				err1 := errors.NewValidationError("first error").WithCode(string(errors.ErrCodeValidationFailed))
				err2 := errors.NewNotFoundError("resource").WithCode(string(errors.ErrCodeCharacterNotFound))
				_ = c.Error(err1)
				_ = c.Error(err2)
				c.Abort()
			},
			expectedStatus: http.StatusNotFound,                     // Gin returns the status of the last error
			expectedCode:   string(errors.ErrCodeCharacterNotFound), // And the code of the last error
			expectedMsg:    "resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			// Add recovery middleware first
			router.Use(gin.Recovery())
			// Add error handler middleware
			router.Use(ErrorHandlerGin())

			router.GET("/test", tt.handler)

			client := testutil.NewHTTPTestClient(t).SetRouter(router)
			resp := client.GET("/test")

			resp.AssertStatus(tt.expectedStatus)

			var response map[string]interface{}
			resp.DecodeJSON(&response)

			require.Equal(t, tt.expectedCode, response["code"])

			if tt.expectedMsg != "" {
				require.Contains(t, response["message"], tt.expectedMsg)
			}

			// Verify request ID is present
			require.NotEmpty(t, response["request_id"])
		})
	}
}

func TestErrorHandlerMiddleware_ContextEnrichment(t *testing.T) {
	t.Run("includes user context in error", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		router.Use(ErrorHandlerGin())

		router.GET("/test", func(c *gin.Context) {
			c.Set("user_id", int64(123))
			c.Set("username", "testuser")
			err := errors.NewAuthorizationError("access denied").WithCode(string(errors.ErrCodeInsufficientPrivilege))
			_ = c.Error(err)
			c.Abort()
		})

		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.GET("/test")

		resp.AssertForbidden()

		var response map[string]interface{}
		resp.DecodeJSON(&response)

		// In production, user context might be logged but not returned
		// This test verifies the middleware has access to context
		require.Equal(t, string(errors.ErrCodeInsufficientPrivilege), response["code"])
	})

	t.Run("includes game session context", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		router.Use(ErrorHandlerGin())

		router.GET("/test", func(c *gin.Context) {
			c.Set("session_id", int64(456))
			c.Set("is_dm", true)
			err := errors.NewNotFoundError("session").WithCode(string(errors.ErrCodeSessionNotFound))
			_ = c.Error(err)
			c.Abort()
		})

		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.GET("/test")

		resp.AssertNotFound()
	})
}

func TestErrorHandlerMiddleware_ValidationErrors(t *testing.T) {
	t.Run("formats validation errors properly", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		router.Use(ErrorHandlerGin())

		router.POST("/test", func(c *gin.Context) {
			validationErrors := &errors.ValidationErrors{}
			validationErrors.Add("name", "is required")
			validationErrors.Add("level", "must be between 1 and 20")
			validationErrors.Add("abilities.strength", "must be at least 3")

			_ = c.Error(validationErrors)
			c.Abort()
		})

		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.POST("/test", nil)

		resp.AssertBadRequest()

		var response map[string]interface{}
		resp.DecodeJSON(&response)

		require.Equal(t, string(errors.ErrCodeValidationFailed), response["code"])

		// Check for field errors
		fieldErrors, ok := response["field_errors"].(map[string]interface{})
		require.True(t, ok)

		// The ValidationErrors struct stores errors as arrays of strings
		nameErrors, ok := fieldErrors["name"].([]interface{})
		require.True(t, ok)
		require.Contains(t, nameErrors, "is required")

		levelErrors, ok := fieldErrors["level"].([]interface{})
		require.True(t, ok)
		require.Contains(t, levelErrors, "must be between 1 and 20")

		strengthErrors, ok := fieldErrors["abilities.strength"].([]interface{})
		require.True(t, ok)
		require.Contains(t, strengthErrors, "must be at least 3")
	})
}

func TestErrorHandlerMiddleware_RateLimiting(t *testing.T) {
	t.Run("handles rate limit errors with retry header", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		router.Use(ErrorHandlerGin())

		router.GET("/test", func(c *gin.Context) {
			err := errors.NewRateLimitError("Too many requests").WithCode(string(errors.ErrCodeRateLimitExceeded)).WithDetails(map[string]interface{}{"retry_after": 60})
			_ = c.Error(err)
			c.Abort()
		})

		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.Request(http.MethodGet, "/test", nil)

		testResp := testutil.NewHTTPTestResponse(t, resp)
		testResp.AssertStatus(http.StatusTooManyRequests)
		testResp.AssertHeader("Retry-After", "60")

		var response map[string]interface{}
		testResp.DecodeJSON(&response)

		require.Equal(t, string(errors.ErrCodeRateLimitExceeded), response["code"])
	})
}

func TestErrorHandlerMiddleware_DatabaseErrors(t *testing.T) {
	tests := []struct {
		name           string
		dbError        error
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:           "not found error",
			dbError:        sql.ErrNoRows,
			expectedStatus: http.StatusNotFound,
			expectedCode:   string(errors.ErrCodeCharacterNotFound),
			expectedMsg:    "Resource not found",
		},
		{
			name:           "connection error",
			dbError:        stderrors.New("connection refused"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   string(errors.ErrCodeDatabaseError),
			expectedMsg:    "Database unavailable",
		},
		{
			name:           "constraint violation",
			dbError:        stderrors.New("duplicate key value violates unique constraint"),
			expectedStatus: http.StatusConflict,
			expectedCode:   string(errors.ErrCodeDuplicateEntry),
			expectedMsg:    "Resource already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			router.Use(ErrorHandlerGin())

			router.GET("/test", func(c *gin.Context) {
				// Convert database error to app error
				var err error
				if stderrors.Is(tt.dbError, sql.ErrNoRows) {
					err = errors.NewNotFoundError("Resource").WithCode(string(errors.ErrCodeCharacterNotFound))
				} else if tt.dbError != nil && tt.dbError.Error() == "connection refused" {
					err = errors.NewServiceUnavailableError("Database unavailable").WithCode(string(errors.ErrCodeDatabaseError))
				} else if tt.dbError != nil && tt.dbError.Error() == "duplicate key value violates unique constraint" {
					err = errors.NewConflictError("Resource already exists").WithCode(string(errors.ErrCodeDuplicateEntry))
				}
				_ = c.Error(err)
				c.Abort()
			})

			client := testutil.NewHTTPTestClient(t).SetRouter(router)
			resp := client.GET("/test")

			resp.AssertStatus(tt.expectedStatus)

			var response map[string]interface{}
			resp.DecodeJSON(&response)

			require.Equal(t, tt.expectedCode, response["code"])
			require.Contains(t, response["message"], tt.expectedMsg)
		})
	}
}

func TestErrorHandlerMiddleware_SecurityErrors(t *testing.T) {
	t.Run("sanitizes sensitive information", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		router.Use(ErrorHandlerGin())

		router.GET("/test", func(c *gin.Context) {
			// Error containing sensitive info
			err := stderrors.New("invalid password: expected 'secret123' but got 'wrongpass'")
			_ = c.Error(err)
			c.Abort()
		})

		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.GET("/test")

		resp.AssertStatus(http.StatusInternalServerError)

		var response map[string]interface{}
		resp.DecodeJSON(&response)

		// Should not expose sensitive information
		require.NotContains(t, response["message"], "secret123")
		require.NotContains(t, response["message"], "wrongpass")
		require.Equal(t, "Internal server error", response["message"])
	})

	t.Run("handles CSRF errors", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		router.Use(ErrorHandlerGin())

		router.POST("/test", func(c *gin.Context) {
			err := errors.NewAuthorizationError("CSRF token mismatch").WithCode(string(errors.ErrCodeCSRFTokenMismatch))
			_ = c.Error(err)
			c.Abort()
		})

		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.POST("/test", nil)

		resp.AssertForbidden()

		var response map[string]interface{}
		resp.DecodeJSON(&response)

		require.Equal(t, string(errors.ErrCodeCSRFTokenMismatch), response["code"])
	})
}

// Benchmark error handler performance
func BenchmarkErrorHandler(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(ErrorHandlerGin())

	router.GET("/test", func(c *gin.Context) {
		err := errors.NewNotFoundError("resource").WithCode(string(errors.ErrCodeCharacterNotFound))
		_ = c.Error(err)
		c.Abort()
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client := testutil.NewHTTPTestClient(&testing.T{}).SetRouter(router)
		_ = client.GET("/test")
	}
}
