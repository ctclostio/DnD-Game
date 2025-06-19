package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// Recovery middleware recovers from panics and returns a 500 error.
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
					if _, err := fmt.Fprintf(w, `{"error":"Internal server error"}`); err != nil {
						fmt.Printf("failed to write error response: %v\n", err)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryConfig allows custom error handling.
type RecoveryConfig struct {
	Logger       *logger.LoggerV2
	PrintStack   bool
	StackSize    int
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err interface{})
}

// RecoveryWithConfig creates a recovery middleware with custom configuration.
func RecoveryWithConfig(config RecoveryConfig) func(http.Handler) http.Handler {
	// Set defaults
	if config.StackSize == 0 {
		config.StackSize = 4 << 10 // 4KB
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer handlePanic(w, r, &config)
			next.ServeHTTP(w, r)
		})
	}
}

// handlePanic recovers from panics and handles the error
func handlePanic(w http.ResponseWriter, r *http.Request, config *RecoveryConfig) {
	if err := recover(); err != nil {
		stack := captureStackTrace(config.StackSize, config.PrintStack)
		logPanic(r, err, stack, config)
		respondToError(w, r, err, config)
	}
}

// captureStackTrace captures the current stack trace
func captureStackTrace(stackSize int, fullStack bool) []byte {
	stack := make([]byte, stackSize)
	length := runtime.Stack(stack, fullStack)
	return stack[:length]
}

// logPanic logs the panic with structured logging
func logPanic(r *http.Request, err interface{}, stack []byte, config *RecoveryConfig) {
	if config.Logger == nil {
		return
	}
	
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

// respondToError sends the error response to the client
func respondToError(w http.ResponseWriter, r *http.Request, err interface{}, config *RecoveryConfig) {
	if config.ErrorHandler != nil {
		config.ErrorHandler(w, r, err)
		return
	}
	
	// Default error handling
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	if _, writeErr := fmt.Fprintf(w, `{"error":"Internal server error"}`); writeErr != nil {
		fmt.Printf("failed to write error response: %v\n", writeErr)
	}
}
