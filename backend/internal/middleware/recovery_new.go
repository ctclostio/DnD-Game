package middleware

import (
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/ctclostio/DnD-Game/backend/pkg/errors"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

// RecoveryMiddleware returns a middleware that recovers from panics.
func RecoveryMiddleware(log *logger.LoggerV2) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with structured logging.
					log.WithContext(r.Context()).
						Error().
						Interface("panic", err).
						Str("stack_trace", string(debug.Stack())).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Str("remote_ip", getClientIP(r)).
						Str("user_agent", r.UserAgent()).
						Msg("Panic recovered")

					// Send error response.
					appErr := errors.NewInternalError("Internal server error", nil)
					response.Error(w, r, appErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryWithConfigNew allows custom error handling with structured logging.
type RecoveryConfigNew struct {
	Logger       *logger.LoggerV2
	PrintStack   bool
	StackSize    int
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err interface{}, log *logger.LoggerV2)
}

// RecoveryWithConfigNew creates a recovery middleware with custom configuration.
func RecoveryWithConfigNew(config RecoveryConfigNew) func(http.Handler) http.Handler {
	// Set defaults.
	if config.StackSize == 0 {
		config.StackSize = 4 << 10 // 4KB
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get stack trace.
					stack := make([]byte, config.StackSize)
					length := runtime.Stack(stack, config.PrintStack)
					stack = stack[:length]

					// Create log entry.
					logEvent := config.Logger.WithContext(r.Context()).
						Error().
						Interface("panic", err).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Str("remote_ip", getClientIP(r)).
						Str("user_agent", r.UserAgent())

					if config.PrintStack {
						logEvent = logEvent.Str("stack_trace", string(stack))
					}

					// Log the panic.
					logEvent.Msg("Panic recovered")

					// Handle the error.
					if config.ErrorHandler != nil {
						config.ErrorHandler(w, r, err, config.Logger.WithContext(r.Context()))
					} else {
						// Default error handling.
						appErr := errors.NewInternalError("Internal server error", nil)
						response.Error(w, r, appErr)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
