package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// Recovery middleware recovers from panics and returns a 500 error
func Recovery(log *logger.LoggerV2) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the error with context and stack trace
					if log != nil {
						log.WithContext(r.Context()).
							Error().
							Interface("panic", err).
							Str("method", r.Method).
							Str("path", r.URL.Path).
							Str("stack_trace", string(debug.Stack())).
							Msg("Panic recovered in HTTP handler")
					}

					// Return 500 Internal Server Error
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					// Send a generic error message (don't expose internal details)
					fmt.Fprintf(w, `{"error":"Internal server error"}`)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryConfig allows custom error handling
type RecoveryConfig struct {
	Logger       *logger.LoggerV2
	PrintStack   bool
	StackSize    int
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err interface{})
}

// RecoveryWithConfig creates a recovery middleware with custom configuration
func RecoveryWithConfig(config RecoveryConfig) func(http.Handler) http.Handler {
	// Set defaults
	if config.StackSize == 0 {
		config.StackSize = 4 << 10 // 4KB
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get stack trace
					stack := make([]byte, config.StackSize)
					length := runtime.Stack(stack, config.PrintStack)
					stack = stack[:length]

					// Log the error with structured logging
					if config.Logger != nil {
						config.Logger.WithContext(r.Context()).
							Error().
							Interface("panic", err).
							Str("method", r.Method).
							Str("path", r.URL.Path).
							Str("remote_addr", r.RemoteAddr).
							Str("user_agent", r.UserAgent()).
							Str("stack_trace", string(stack)).
							Bool("full_stack", config.PrintStack).
							Msg("Panic recovered in HTTP handler")
					}

					// Handle the error
					if config.ErrorHandler != nil {
						config.ErrorHandler(w, r, err)
					} else {
						// Default error handling
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)
						fmt.Fprintf(w, `{"error":"Internal server error"}`)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
