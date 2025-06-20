package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/pkg/errors"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

// Common constants
const (
	xRequestIDHeader = "X-Request-ID"
	authenticationRequiredMsg = "Authentication required"
)

// ErrorHandlerV2 is the enhanced error handling middleware
func ErrorHandlerV2(log *logger.LoggerV2) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add request ID to context if not present
			ctx := r.Context()
			requestID := r.Header.Get(xRequestIDHeader)
			if requestID == "" {
				requestID = uuid.New().String()
			}
			ctx = context.WithValue(ctx, response.RequestIDKey, requestID)
			r = r.WithContext(ctx)

			// Add request ID to response header
			w.Header().Set(xRequestIDHeader, requestID)

			// Create a custom response writer to capture panic
			rw := &panicCapturingResponseWriter{
				ResponseWriter: w,
				log:            log,
				requestID:      requestID,
			}

			// Defer panic recovery
			defer func() {
				if rec := recover(); rec != nil {
					rw.handlePanic(rec, r)
				}
			}()

			next.ServeHTTP(rw, r)
		})
	}
}

// panicCapturingResponseWriter captures panics and converts them to proper error responses
type panicCapturingResponseWriter struct {
	http.ResponseWriter
	log       *logger.LoggerV2
	requestID string
}

func (w *panicCapturingResponseWriter) handlePanic(rec interface{}, r *http.Request) {
	// Log the panic with stack trace
	stackTrace := string(debug.Stack())
	w.log.WithContext(r.Context()).
		Error().
		Str("request_id", w.requestID).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Interface("panic", rec).
		Str("stack_trace", stackTrace).
		Msg("Panic recovered in request handler")

	// Create internal error
	err := errors.NewInternalError("An unexpected error occurred", fmt.Errorf("panic: %v", rec)).
		WithCode(string(errors.ErrCodeInternalError))

	// Send error response using the standard response package
	response.Error(w, r, err)
}

// RequestIDMiddleware ensures every request has a unique ID
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(xRequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to context
		ctx := context.WithValue(r.Context(), response.RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Add to response header
		w.Header().Set(xRequestIDHeader, requestID)

		next.ServeHTTP(w, r)
	})
}

// HandlerFunc is an enhanced handler function that can return an error
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Handler wraps a HandlerFunc to handle errors consistently
func Handler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			response.Error(w, r, err)
		}
	}
}

// AuthenticatedHandler wraps a handler that requires authentication
type AuthenticatedHandlerFunc func(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error

// AuthenticatedHandler wraps an authenticated handler function
func AuthenticatedHandler(h AuthenticatedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from context (set by auth middleware)
		userID, ok := r.Context().Value("user_id").(uuid.UUID)
		if !ok {
			response.Unauthorized(w, r, authenticationRequiredMsg)
			return
		}

		if err := h(w, r, userID); err != nil {
			response.Error(w, r, err)
		}
	}
}

// GameSessionHandler wraps a handler that requires game session context
type GameSessionHandlerFunc func(w http.ResponseWriter, r *http.Request, userID, sessionID uuid.UUID) error

// GameSessionHandler wraps a game session handler function
func GameSessionHandler(h GameSessionHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from context
		userID, ok := r.Context().Value("user_id").(uuid.UUID)
		if !ok {
			response.Unauthorized(w, r, authenticationRequiredMsg)
			return
		}

		// Extract session ID from context (set by game session middleware)
		sessionID, ok := r.Context().Value("session_id").(uuid.UUID)
		if !ok {
			response.ErrorWithCode(w, r, errors.ErrCodeNotInSession)
			return
		}

		if err := h(w, r, userID, sessionID); err != nil {
			response.Error(w, r, err)
		}
	}
}

// DMOnlyHandler wraps a handler that requires DM privileges
type DMOnlyHandlerFunc func(w http.ResponseWriter, r *http.Request, userID, sessionID uuid.UUID) error

// DMOnlyHandler wraps a DM-only handler function
func DMOnlyHandler(h DMOnlyHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from context
		userID, ok := r.Context().Value("user_id").(uuid.UUID)
		if !ok {
			response.Unauthorized(w, r, authenticationRequiredMsg)
			return
		}

		// Extract session ID from context
		sessionID, ok := r.Context().Value("session_id").(uuid.UUID)
		if !ok {
			response.ErrorWithCode(w, r, errors.ErrCodeNotInSession)
			return
		}

		// Check if user is DM (set by game session middleware)
		isDM, ok := r.Context().Value("is_dm").(bool)
		if !ok || !isDM {
			response.ErrorWithCode(w, r, errors.ErrCodeNotDM)
			return
		}

		if err := h(w, r, userID, sessionID); err != nil {
			response.Error(w, r, err)
		}
	}
}
