package response

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/pkg/errors"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// Response represents a standard API response
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorInfo represents error information in the response
type ErrorInfo struct {
	Type    errors.ErrorType `json:"type"`
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Details interface{}      `json:"details,omitempty"`
}

// Meta contains pagination and other metadata
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// RequestIDKey is the context key for request ID
type contextKey string

const RequestIDKey contextKey = "request_id"

// getRequestID extracts request ID from the request context
func getRequestID(r *http.Request) string {
	if id := r.Context().Value(RequestIDKey); id != nil {
		if reqID, ok := id.(string); ok {
			return reqID
		}
	}
	// Generate new ID if not found
	return uuid.New().String()
}

// JSON sends a successful JSON response
func JSON(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	response := Response{
		Success:   true,
		Data:      data,
		RequestID: getRequestID(r),
		Timestamp: time.Now().UTC(),
	}

	sendJSON(w, status, response)
}

// JSONWithMeta sends a successful JSON response with metadata
func JSONWithMeta(w http.ResponseWriter, r *http.Request, status int, data interface{}, meta *Meta) {
	response := Response{
		Success:   true,
		Data:      data,
		Meta:      meta,
		RequestID: getRequestID(r),
		Timestamp: time.Now().UTC(),
	}

	sendJSON(w, status, response)
}

// Error sends an error response
func Error(w http.ResponseWriter, r *http.Request, err error) {
	appErr := errors.GetAppError(err)

	// Log the error with context
	log := logger.GetLogger()
	requestID := getRequestID(r)

	// Log at appropriate level
	switch appErr.StatusCode {
	case http.StatusInternalServerError, http.StatusServiceUnavailable:
		log.Error().
			Str("request_id", requestID).
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Err(appErr.Internal).
			Msg(appErr.Message)
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		log.Warn().
			Str("request_id", requestID).
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Msg(appErr.Message)
	default:
		log.Info().
			Str("request_id", requestID).
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Msg(appErr.Message)
	}

	response := Response{
		Success:   false,
		Error:     appErrorToErrorInfo(appErr),
		RequestID: requestID,
		Timestamp: time.Now().UTC(),
	}

	sendJSON(w, appErr.StatusCode, response)
}

// ErrorWithCode sends an error response with a specific error code
func ErrorWithCode(w http.ResponseWriter, r *http.Request, code errors.ErrorCode, customMessage ...string) {
	message := errors.GetErrorMessage(code)
	if len(customMessage) > 0 && customMessage[0] != "" {
		message = customMessage[0]
	}

	// Determine error type and status code based on error code
	var errType errors.ErrorType
	var statusCode int

	switch code {
	// Authentication errors (401 Unauthorized)
	case errors.ErrCodeInvalidCredentials, errors.ErrCodeTokenExpired, errors.ErrCodeTokenInvalid,
		errors.ErrCodeSessionExpired, errors.ErrCodeCSRFTokenMismatch, errors.ErrCodeEmailNotVerified:
		errType = errors.ErrorTypeAuthentication
		statusCode = http.StatusUnauthorized

	// Authorization errors (403 Forbidden)
	case errors.ErrCodeInsufficientPrivilege, errors.ErrCodeCharacterNotOwned, errors.ErrCodeNotDM:
		errType = errors.ErrorTypeAuthorization
		statusCode = http.StatusForbidden

	// Not Found errors (404)
	case errors.ErrCodeUserNotFound, errors.ErrCodeCharacterNotFound, errors.ErrCodeSessionNotFound,
		errors.ErrCodeItemNotFound:
		errType = errors.ErrorTypeNotFound
		statusCode = http.StatusNotFound

	// Conflict errors (409)
	case errors.ErrCodeUserExists, errors.ErrCodeDuplicateEntry, errors.ErrCodeSessionFull,
		errors.ErrCodeSessionInProgress, errors.ErrCodeNotInSession, errors.ErrCodeCombatNotActive,
		errors.ErrCodeNotYourTurn, errors.ErrCodeItemAlreadyEquipped:
		errType = errors.ErrorTypeConflict
		statusCode = http.StatusConflict

	// Validation/Bad Request errors (400)
	case errors.ErrCodeValidationFailed, errors.ErrCodeInvalidInput, errors.ErrCodeMissingRequired,
		errors.ErrCodeInvalidPassword, errors.ErrCodeInvalidFormat, errors.ErrCodeOutOfRange,
		errors.ErrCodeCharacterLimitReached, errors.ErrCodeInvalidCharacterData, errors.ErrCodeInvalidTarget,
		errors.ErrCodeInsufficientRange, errors.ErrCodeNoActionsRemaining, errors.ErrCodeInventoryFull,
		errors.ErrCodeInsufficientFunds, errors.ErrCodeItemNotEquippable, errors.ErrCodeAIInvalidRequest:
		errType = errors.ErrorTypeValidation
		statusCode = http.StatusBadRequest

	// Rate Limit errors (429)
	case errors.ErrCodeRateLimitExceeded, errors.ErrCodeAIRateLimitExceeded:
		errType = errors.ErrorTypeRateLimit
		statusCode = http.StatusTooManyRequests

	// Service Unavailable errors (503)
	case errors.ErrCodeServiceUnavailable, errors.ErrCodeAIServiceUnavailable, errors.ErrCodeAIGenerationFailed:
		errType = errors.ErrorTypeServiceUnavailable
		statusCode = http.StatusServiceUnavailable

	// Internal Server errors (500)
	case errors.ErrCodeDatabaseError, errors.ErrCodeForeignKeyViolation, errors.ErrCodeDeadlock,
		errors.ErrCodeInternalError, errors.ErrCodeTimeout:
		errType = errors.ErrorTypeInternal
		statusCode = http.StatusInternalServerError

	default:
		errType = errors.ErrorTypeInternal
		statusCode = http.StatusInternalServerError
	}

	appErr := &errors.AppError{
		Type:       errType,
		Message:    message,
		Code:       string(code),
		StatusCode: statusCode,
	}

	Error(w, r, appErr)
}

// ValidationError sends a validation error response
func ValidationError(w http.ResponseWriter, r *http.Request, validationErrors *errors.ValidationErrors) {
	if validationErrors == nil || !validationErrors.HasErrors() {
		return
	}

	Error(w, r, validationErrors.ToAppError().WithCode(string(errors.ErrCodeValidationFailed)))
}

// NotFound sends a not found error response
func NotFound(w http.ResponseWriter, r *http.Request, resource string) {
	Error(w, r, errors.NewNotFoundError(resource).WithCode(string(errors.ErrCodeCharacterNotFound)))
}

// Unauthorized sends an unauthorized error response
func Unauthorized(w http.ResponseWriter, r *http.Request, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	Error(w, r, errors.NewAuthenticationError(message).WithCode(string(errors.ErrCodeTokenInvalid)))
}

// Forbidden sends a forbidden error response
func Forbidden(w http.ResponseWriter, r *http.Request, message string) {
	if message == "" {
		message = "Forbidden"
	}
	Error(w, r, errors.NewAuthorizationError(message).WithCode(string(errors.ErrCodeInsufficientPrivilege)))
}

// BadRequest sends a bad request error response
func BadRequest(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, errors.NewBadRequestError(message).WithCode(string(errors.ErrCodeInvalidInput)))
}

// InternalServerError sends an internal server error response
func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	Error(w, r, errors.NewInternalError("An unexpected error occurred", err).WithCode(string(errors.ErrCodeInternalError)))
}

// Helper functions

func appErrorToErrorInfo(appErr *errors.AppError) *ErrorInfo {
	return &ErrorInfo{
		Type:    appErr.Type,
		Code:    appErr.Code,
		Message: appErr.Message,
		Details: appErr.Details,
	}
}

func sendJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.GetLogger().Error().Err(err).Msg("Failed to encode JSON response")
	}
}
