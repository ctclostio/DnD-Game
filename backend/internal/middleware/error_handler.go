package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/ctclostio/DnD-Game/backend/pkg/errors"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// ErrorHandler middleware handles errors in a consistent way
func ErrorHandler(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create custom response writer
			rw := &errorHandlerResponseWriter{
				ResponseWriter: w,
				log:            log.WithContext(r.Context()),
			}

			// Defer error handling
			defer func() {
				if rec := recover(); rec != nil {
					rw.handlePanic(rec, r)
				}
			}()

			next.ServeHTTP(rw, r)
		})
	}
}

// errorHandlerResponseWriter wraps response writer to handle errors
type errorHandlerResponseWriter struct {
	http.ResponseWriter
	log *logger.Logger
}

// handlePanic handles panic recovery
func (w *errorHandlerResponseWriter) handlePanic(rec interface{}, r *http.Request) {
	w.log.WithFields(map[string]interface{}{
		"panic":  rec,
		"method": r.Method,
		"path":   r.URL.Path,
	}).Error().Msg("Panic recovered")

	// Send error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	response := map[string]interface{}{
		"type":    errors.ErrorTypeInternal,
		"message": "Internal server error",
	}

	json.NewEncoder(w).Encode(response)
}

// SendError sends an error response
func SendError(w http.ResponseWriter, err error, log *logger.Logger) {
	appErr := errors.GetAppError(err)

	// Log the error
	logEntry := log.WithError(appErr.Internal).
		WithFields(map[string]interface{}{
			"error_type": appErr.Type,
			"error_code": appErr.Code,
		})

	// Log at appropriate level
	switch appErr.StatusCode {
	case http.StatusInternalServerError, http.StatusServiceUnavailable:
		logEntry.Error().Msg(appErr.Message)
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		logEntry.Warn().Msg(appErr.Message)
	default:
		logEntry.Info().Msg(appErr.Message)
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)

	// Prepare response (don't expose internal error details)
	response := map[string]interface{}{
		"type":    appErr.Type,
		"message": appErr.Message,
	}

	if appErr.Code != "" {
		response["code"] = appErr.Code
	}

	if appErr.Details != nil {
		response["details"] = appErr.Details
	}

	json.NewEncoder(w).Encode(response)
}

// SendSuccess sends a success response
func SendSuccess(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// ErrorHandlerFunc is a handler function that returns an error
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// WrapErrorHandler wraps an error-returning handler
func WrapErrorHandler(handler ErrorHandlerFunc, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			SendError(w, err, log.WithContext(r.Context()))
		}
	}
}
