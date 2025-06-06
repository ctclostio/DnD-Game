package middleware

import (
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/your-username/dnd-game/backend/pkg/errors"
	"github.com/your-username/dnd-game/backend/pkg/logger"
)

// RecoveryMiddleware returns a middleware that recovers from panics
func RecoveryMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with structured logging
					log.WithContext(r.Context()).
						WithFields(map[string]interface{}{
							"panic":       err,
							"stack_trace": string(debug.Stack()),
							"method":      r.Method,
							"path":        r.URL.Path,
							"remote_ip":   getClientIP(r),
							"user_agent":  r.UserAgent(),
						}).
						Error().
						Msg("Panic recovered")
					
					// Send error response
					appErr := errors.NewInternalError("Internal server error", nil)
					SendError(w, appErr, log.WithContext(r.Context()))
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryWithConfigNew allows custom error handling with structured logging
type RecoveryConfigNew struct {
	Logger       *logger.Logger
	PrintStack   bool
	StackSize    int
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err interface{}, log *logger.Logger)
}

// RecoveryWithConfigNew creates a recovery middleware with custom configuration
func RecoveryWithConfigNew(config RecoveryConfigNew) func(http.Handler) http.Handler {
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
					
					// Create log entry
					logEntry := config.Logger.WithContext(r.Context()).
						WithFields(map[string]interface{}{
							"panic":      err,
							"method":     r.Method,
							"path":       r.URL.Path,
							"remote_ip":  getClientIP(r),
							"user_agent": r.UserAgent(),
						})
					
					if config.PrintStack {
						logEntry = logEntry.WithField("stack_trace", string(stack))
					}
					
					// Log the panic
					logEntry.Error().Msg("Panic recovered")
					
					// Handle the error
					if config.ErrorHandler != nil {
						config.ErrorHandler(w, r, err, config.Logger.WithContext(r.Context()))
					} else {
						// Default error handling
						appErr := errors.NewInternalError("Internal server error", nil)
						SendError(w, appErr, config.Logger.WithContext(r.Context()))
					}
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}