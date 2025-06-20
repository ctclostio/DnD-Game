package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ctclostio/DnD-Game/backend/pkg/errors"
)

// ErrorHandlerGin returns a Gin middleware for handling errors
func ErrorHandlerGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request
		c.Next()

		// Check if there are any errors
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		requestID := getRequestID(c)

		// Handle different error types
		switch e := err.(type) {
		case *errors.AppError:
			handleAppError(c, e, requestID)
		case *errors.ValidationErrors:
			handleValidationError(c, e, requestID)
		default:
			handleGenericError(c, requestID)
		}
	}
}

func getRequestID(c *gin.Context) interface{} {
	requestID, exists := c.Get("request_id")
	if exists && requestID != nil {
		return requestID
	}

	requestID = c.GetHeader("X-Request-ID")
	if requestID == "" {
		return "unknown"
	}
	return requestID
}

func handleAppError(c *gin.Context, e *errors.AppError, requestID interface{}) {
	response := buildAppErrorResponse(e, requestID)

	// Check for rate limit error and add retry header
	if e.Type == errors.ErrorTypeRateLimit {
		setRetryAfterHeader(c, e)
	}

	c.JSON(e.StatusCode, response)
}

func buildAppErrorResponse(e *errors.AppError, requestID interface{}) gin.H {
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

	return response
}

func setRetryAfterHeader(c *gin.Context, e *errors.AppError) {
	if retryAfter, ok := e.Details["retry_after"].(int); ok {
		c.Header("Retry-After", strconv.Itoa(retryAfter))
	}
}

func handleValidationError(c *gin.Context, e *errors.ValidationErrors, requestID interface{}) {
	response := gin.H{
		"code":         string(errors.ErrCodeValidationFailed),
		"message":      "Validation failed",
		"field_errors": e.Errors,
		"request_id":   requestID,
	}
	c.JSON(http.StatusBadRequest, response)
}

func handleGenericError(c *gin.Context, requestID interface{}) {
	response := gin.H{
		"code":       string(errors.ErrCodeInternalError),
		"message":    "Internal server error",
		"request_id": requestID,
	}
	c.JSON(http.StatusInternalServerError, response)
}
