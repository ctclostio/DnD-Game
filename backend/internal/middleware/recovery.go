package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

// Recovery middleware recovers from panics and returns a 500 error
func Recovery(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the error and stack trace
					if logger != nil {
						logger.Printf("Panic recovered: %v\n%s", err, debug.Stack())
					} else {
						log.Printf("Panic recovered: %v\n%s", err, debug.Stack())
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

// RecoveryWithConfig allows custom error handling
type RecoveryConfig struct {
	Logger        *log.Logger
	PrintStack    bool
	StackSize     int
	ErrorHandler  func(w http.ResponseWriter, r *http.Request, err interface{})
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
					
					// Log the error
					if config.Logger != nil {
						config.Logger.Printf("Panic recovered: %v\nRequest: %s %s\nStack trace:\n%s", 
							err, r.Method, r.URL.Path, stack)
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