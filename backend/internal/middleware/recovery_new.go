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
					// Log the panic with structured logging
					log.WithContext(r.Context()).
						Error().
						Interface("panic", err).
						Str("stack_trace", string(debug.Stack())).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Str("remote_ip", getClientIP(r)).
						Str("user_agent", r.UserAgent()).
						Msg("Panic recovered")

					// Send error response
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
	// Set defaults
	config = setRecoveryDefaults(config)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer handlePanicWithConfig(w, r, &config)
			next.ServeHTTP(w, r)
		})
	}
}

func setRecoveryDefaults(config RecoveryConfigNew) RecoveryConfigNew {
	if config.StackSize == 0 {
		config.StackSize = 4 << 10 // 4KB
	}
	return config
}

func handlePanicWithConfig(w http.ResponseWriter, r *http.Request, config *RecoveryConfigNew) {
	if err := recover(); err != nil {
		logPanicWithConfig(r, err, config)
		respondToError(w, r, err, config)
	}
}

func logPanicWithConfig(r *http.Request, err interface{}, config *RecoveryConfigNew) {
	stack := captureStackTrace(config.StackSize, config.PrintStack)

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

	logEvent.Msg("Panic recovered")
}

func captureStackTrace(stackSize int, fullStack bool) []byte {
	stack := make([]byte, stackSize)
	length := runtime.Stack(stack, fullStack)
	return stack[:length]
}

func respondToError(w http.ResponseWriter, r *http.Request, err interface{}, config *RecoveryConfigNew) {
	if config.ErrorHandler != nil {
		config.ErrorHandler(w, r, err, config.Logger.WithContext(r.Context()))
	} else {
		appErr := errors.NewInternalError("Internal server error", nil)
		response.Error(w, r, appErr)
	}
}
