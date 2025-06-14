package middleware

import (
	"net/http"
	"strconv"

	"github.com/ctclostio/DnD-Game/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ErrorHandlerGin returns a Gin middleware for handling errors.
func ErrorHandlerGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request.
		c.Next()

		// Check if there are any errors.
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Get request ID from context.
			requestID, _ := c.Get("request_id")
			if requestID == nil {
				requestID = c.GetHeader("X-Request-ID")
				if requestID == "" {
					requestID = "unknown"
				}
			}

			// Handle different error types.
			switch e := err.(type) {
			case *errors.AppError:
				// Handle AppError.
				response := gin.H{
					"code":       e.Code,
					"message":    e.Message,
					"request_id": requestID,
				}

				if e.Details != nil {
					for k, v := range e.Details {
						response[k] = v
					}
				}

				// Check for rate limit error and add retry header.
				if e.Type == errors.ErrorTypeRateLimit {
					if retryAfter, ok := e.Details["retry_after"].(int); ok {
						c.Header("Retry-After", strconv.Itoa(retryAfter))
					}
				}

				c.JSON(e.StatusCode, response)
				return

			case *errors.ValidationErrors:
				// Handle validation errors.
				response := gin.H{
					"code":         string(errors.ErrCodeValidationFailed),
					"message":      "Validation failed",
					"field_errors": e.Errors,
					"request_id":   requestID,
				}

				c.JSON(http.StatusBadRequest, response)
				return

			default:
				// Handle generic errors.
				response := gin.H{
					"code":       string(errors.ErrCodeInternalError),
					"message":    "Internal server error",
					"request_id": requestID,
				}

				c.JSON(http.StatusInternalServerError, response)
				return
			}
		}
	}
}
