package middleware

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/your-org/dnd-game/internal/testutil"
	"github.com/your-org/dnd-game/pkg/errors"
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
				err := errors.NewError(errors.CodeCharacterNotFound, "character not found")
				c.Error(err)
				c.Abort()
			},
			expectedStatus: http.StatusNotFound,
			expectedCode:   errors.CodeCharacterNotFound,
			expectedMsg:    "character not found",
		},
		{
			name: "handles validation error",
			handler: func(c *gin.Context) {
				err := errors.NewValidationError("name", "is required")
				c.Error(err)
				c.Abort()
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   errors.CodeValidationFailed,
			expectedMsg:    "name is required",
		},
		{
			name: "handles authorization error",
			handler: func(c *gin.Context) {
				err := errors.NewError(errors.CodeUnauthorized, "invalid token")
				c.Error(err)
				c.Abort()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   errors.CodeUnauthorized,
			expectedMsg:    "invalid token",
		},
		{
			name: "handles generic error",
			handler: func(c *gin.Context) {
				err := errors.New("something went wrong")
				c.Error(err)
				c.Abort()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   errors.CodeInternalError,
			expectedMsg:    "Internal server error",
		},
		{
			name: "handles panic recovery",
			handler: func(c *gin.Context) {
				panic("unexpected panic")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   errors.CodeInternalError,
			expectedMsg:    "Internal server error",
		},
		{
			name: "preserves request ID",
			handler: func(c *gin.Context) {
				c.Set("request_id", "test-request-123")
				err := errors.NewError(errors.CodeNotFound, "not found")
				c.Error(err)
				c.Abort()
			},
			expectedStatus: http.StatusNotFound,
			expectedCode:   errors.CodeNotFound,
		},
		{
			name: "handles multiple errors (returns first)",
			handler: func(c *gin.Context) {
				err1 := errors.NewError(errors.CodeValidationFailed, "first error")
				err2 := errors.NewError(errors.CodeNotFound, "second error")
				c.Error(err1)
				c.Error(err2)
				c.Abort()
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   errors.CodeValidationFailed,
			expectedMsg:    "first error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			
			// Add recovery middleware first
			router.Use(gin.Recovery())
			// Add error handler middleware
			router.Use(ErrorHandler())
			
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
		
		router.Use(ErrorHandler())
		
		router.GET("/test", func(c *gin.Context) {
			c.Set("user_id", int64(123))
			c.Set("username", "testuser")
			err := errors.NewError(errors.CodeForbidden, "access denied")
			c.Error(err)
			c.Abort()
		})
		
		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.GET("/test")
		
		resp.AssertForbidden()
		
		var response map[string]interface{}
		resp.DecodeJSON(&response)
		
		// In production, user context might be logged but not returned
		// This test verifies the middleware has access to context
		require.Equal(t, errors.CodeForbidden, response["code"])
	})

	t.Run("includes game session context", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		
		router.Use(ErrorHandler())
		
		router.GET("/test", func(c *gin.Context) {
			c.Set("session_id", int64(456))
			c.Set("is_dm", true)
			err := errors.NewError(errors.CodeGameSessionNotFound, "session not found")
			c.Error(err)
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
		
		router.Use(ErrorHandler())
		
		router.POST("/test", func(c *gin.Context) {
			validationErrors := errors.NewValidationErrors()
			validationErrors.AddFieldError("name", "is required")
			validationErrors.AddFieldError("level", "must be between 1 and 20")
			validationErrors.AddFieldError("abilities.strength", "must be at least 3")
			
			c.Error(validationErrors)
			c.Abort()
		})
		
		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.POST("/test", nil)
		
		resp.AssertBadRequest()
		
		var response map[string]interface{}
		resp.DecodeJSON(&response)
		
		require.Equal(t, errors.CodeValidationFailed, response["code"])
		
		// Check for field errors
		fieldErrors, ok := response["field_errors"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "is required", fieldErrors["name"])
		require.Equal(t, "must be between 1 and 20", fieldErrors["level"])
		require.Equal(t, "must be at least 3", fieldErrors["abilities.strength"])
	})
}

func TestErrorHandlerMiddleware_RateLimiting(t *testing.T) {
	t.Run("handles rate limit errors with retry header", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		
		router.Use(ErrorHandler())
		
		router.GET("/test", func(c *gin.Context) {
			err := errors.NewRateLimitError(60) // 60 seconds until retry
			c.Error(err)
			c.Abort()
		})
		
		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.Request(http.MethodGet, "/test", nil)
		
		testResp := testutil.NewHTTPTestResponse(t, resp)
		testResp.AssertStatus(http.StatusTooManyRequests)
		testResp.AssertHeader("Retry-After", "60")
		
		var response map[string]interface{}
		testResp.DecodeJSON(&response)
		
		require.Equal(t, errors.CodeRateLimitExceeded, response["code"])
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
			expectedCode:   errors.CodeNotFound,
			expectedMsg:    "Resource not found",
		},
		{
			name:           "connection error",
			dbError:        errors.New("connection refused"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   errors.CodeDatabaseError,
			expectedMsg:    "Database unavailable",
		},
		{
			name:           "constraint violation",
			dbError:        errors.New("duplicate key value violates unique constraint"),
			expectedStatus: http.StatusConflict,
			expectedCode:   errors.CodeDuplicateEntry,
			expectedMsg:    "Resource already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			
			router.Use(ErrorHandler())
			
			router.GET("/test", func(c *gin.Context) {
				// Wrap database error
				err := errors.WrapDatabaseError(tt.dbError)
				c.Error(err)
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
		
		router.Use(ErrorHandler())
		
		router.GET("/test", func(c *gin.Context) {
			// Error containing sensitive info
			err := errors.New("invalid password: expected 'secret123' but got 'wrongpass'")
			c.Error(err)
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
		
		router.Use(ErrorHandler())
		
		router.POST("/test", func(c *gin.Context) {
			err := errors.NewError(errors.CodeCSRFTokenInvalid, "CSRF token mismatch")
			c.Error(err)
			c.Abort()
		})
		
		client := testutil.NewHTTPTestClient(t).SetRouter(router)
		resp := client.POST("/test", nil)
		
		resp.AssertForbidden()
		
		var response map[string]interface{}
		resp.DecodeJSON(&response)
		
		require.Equal(t, errors.CodeCSRFTokenInvalid, response["code"])
	})
}

// Benchmark error handler performance
func BenchmarkErrorHandler(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	
	router.Use(ErrorHandler())
	
	router.GET("/test", func(c *gin.Context) {
		err := errors.NewError(errors.CodeNotFound, "not found")
		c.Error(err)
		c.Abort()
	})
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		client := testutil.NewHTTPTestClient(&testing.T{}).SetRouter(router)
		_ = client.GET("/test")
	}
}