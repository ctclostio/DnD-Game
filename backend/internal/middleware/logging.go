package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// RequestLogger middleware logs all HTTP requests
func RequestLogger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add request ID and correlation ID to context
			ctx := logger.ContextWithRequestID(r.Context(), requestID)

			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = requestID
			}
			ctx = logger.ContextWithCorrelationID(ctx, correlationID)
			r = r.WithContext(ctx)

			// Create a response writer wrapper to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Add IDs to response header
			w.Header().Set("X-Request-ID", requestID)
			w.Header().Set("X-Correlation-ID", correlationID)

			// Log request start
			log.WithRequestID(requestID).
				WithCorrelationID(correlationID).
				WithFields(map[string]interface{}{
					"method":     r.Method,
					"path":       r.URL.Path,
					"query":      r.URL.RawQuery,
					"remote_ip":  getClientIP(r),
					"user_agent": r.UserAgent(),
				}).
				Info().
				Msg("Request started")

			// Process request
			next.ServeHTTP(rw, r)

			// Log request completion
			duration := time.Since(start)
			log.WithRequestID(requestID).
				WithCorrelationID(correlationID).
				WithFields(map[string]interface{}{
					"method":      r.Method,
					"path":        r.URL.Path,
					"status":      rw.statusCode,
					"duration_ms": duration.Milliseconds(),
					"bytes_sent":  rw.bytesWritten,
				}).
				Info().
				Msg("Request completed")
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	rw.bytesWritten += n
	return n, err
}

// CorrelationID middleware ensures all requests have a correlation ID
func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing correlation ID
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Add to context
		ctx := logger.ContextWithCorrelationID(r.Context(), correlationID)
		r = r.WithContext(ctx)

		// Add to response header
		w.Header().Set("X-Correlation-ID", correlationID)

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// ErrorLogger logs errors with context
func ErrorLogger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create custom response writer to intercept errors
			rw := &errorResponseWriter{
				ResponseWriter: w,
				log:            log.WithContext(r.Context()),
			}

			next.ServeHTTP(rw, r)
		})
	}
}

// errorResponseWriter logs errors when status code >= 400
type errorResponseWriter struct {
	http.ResponseWriter
	log         *logger.Logger
	wroteHeader bool
	statusCode  int
}

func (w *errorResponseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.statusCode = code
		w.wroteHeader = true

		// Log errors
		if code >= 400 {
			w.log.WithField("status_code", code).
				Error().
				Msg("HTTP error response")
		}
	}
	w.ResponseWriter.WriteHeader(code)
}
