package middleware

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/your-username/dnd-game/backend/pkg/logger"
)

// LoggingMiddleware logs all HTTP requests with enhanced context
func LoggingMiddleware(log *logger.LoggerV2) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Extract or generate request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Extract correlation ID
			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = requestID // Use request ID as correlation ID if not provided
			}

			// Add IDs to context
			ctx := r.Context()
			ctx = logger.ContextWithRequestID(ctx, requestID)
			ctx = logger.ContextWithCorrelationID(ctx, correlationID)
			r = r.WithContext(ctx)

			// Create logger with context
			reqLog := log.WithContext(ctx)

			// Create response writer wrapper
			rw := &loggingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				startTime:      start,
			}

			// Add IDs to response headers
			w.Header().Set("X-Request-ID", requestID)
			w.Header().Set("X-Correlation-ID", correlationID)

			// Log request start
			reqLog.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("query", sanitizeQuery(r.URL.RawQuery)).
				Str("remote_ip", getClientIPV2(r)).
				Str("user_agent", r.UserAgent()).
				Str("referer", r.Referer()).
				Int64("content_length", r.ContentLength).
				Msg("HTTP request started")

			// Process request
			next.ServeHTTP(rw, r)

			// Calculate duration
			duration := time.Since(start)

			// Determine log level based on status code and duration
			event := reqLog.Info()
			if rw.statusCode >= 500 {
				event = reqLog.Error()
			} else if rw.statusCode >= 400 {
				event = reqLog.Warn()
			} else if duration > 5*time.Second {
				event = reqLog.Warn()
			}

			// Log request completion
			event.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", rw.statusCode).
				Dur("duration", duration).
				Int64("duration_ms", duration.Milliseconds()).
				Int("bytes_sent", rw.bytesWritten).
				Str("remote_ip", getClientIPV2(r)).
				Msg(getRequestMessage(rw.statusCode))
		})
	}
}

// loggingResponseWriter captures response details
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
	wroteHeader  bool
	startTime    time.Time
}

func (rw *loggingResponseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *loggingResponseWriter) Write(data []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(data)
	rw.bytesWritten += n
	return n, err
}

// Hijack implements http.Hijacker for WebSocket support
func (rw *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("response writer does not support hijacking")
}

// RequestContextMiddleware adds request context values
func RequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from auth token if present
		userID := extractUserIDFromRequest(r)
		if userID != "" {
			ctx := logger.ContextWithUserID(r.Context(), userID)
			r = r.WithContext(ctx)
		}

		// Extract session ID from headers or query
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			sessionID = r.URL.Query().Get("session_id")
		}
		if sessionID != "" {
			ctx := logger.ContextWithSessionID(r.Context(), sessionID)
			r = r.WithContext(ctx)
		}

		// Extract character ID from headers or path
		characterID := r.Header.Get("X-Character-ID")
		if characterID == "" {
			// Try to extract from path (e.g., /characters/{id})
			if strings.Contains(r.URL.Path, "/characters/") {
				parts := strings.Split(r.URL.Path, "/")
				for i, part := range parts {
					if part == "characters" && i+1 < len(parts) {
						characterID = parts[i+1]
						break
					}
				}
			}
		}
		if characterID != "" {
			ctx := logger.ContextWithCharacterID(r.Context(), characterID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// DatabaseQueryLogger logs database queries
type DatabaseQueryLogger struct {
	log *logger.LoggerV2
}

func NewDatabaseQueryLogger(log *logger.LoggerV2) *DatabaseQueryLogger {
	return &DatabaseQueryLogger{log: log}
}

func (d *DatabaseQueryLogger) LogQuery(ctx context.Context, query string, args []interface{}, duration time.Duration, err error) {
	log := d.log.WithContext(ctx)

	event := log.Logger.Debug().
		Str("query", truncateQuery(query)).
		Dur("duration", duration).
		Int("args_count", len(args))

	if err != nil {
		event.Err(err).Msg("Database query failed")
	} else {
		event.Msg("Database query executed")
	}
}

// Helper functions

func getClientIPV2(r *http.Request) string {
	// Check CF-Connecting-IP for Cloudflare
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	// Check X-Forwarded-For
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Take the first IP if there are multiple
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return forwarded
	}

	// Check X-Real-IP
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

func sanitizeQuery(query string) string {
	if query == "" {
		return ""
	}

	// Remove sensitive parameters
	sensitiveParams := []string{"password", "token", "secret", "api_key", "access_token", "refresh_token"}

	params := strings.Split(query, "&")
	sanitized := make([]string, 0, len(params))

	for _, param := range params {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 2 {
			key := strings.ToLower(parts[0])
			isSensitive := false
			for _, sensitive := range sensitiveParams {
				if strings.Contains(key, sensitive) {
					isSensitive = true
					break
				}
			}
			if isSensitive {
				sanitized = append(sanitized, parts[0]+"=[REDACTED]")
			} else {
				sanitized = append(sanitized, param)
			}
		} else {
			sanitized = append(sanitized, param)
		}
	}

	return strings.Join(sanitized, "&")
}

func truncateQuery(query string) string {
	const maxLength = 500
	query = strings.TrimSpace(query)
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")
	query = strings.Join(strings.Fields(query), " ")

	if len(query) > maxLength {
		return query[:maxLength] + "..."
	}
	return query
}

func getRequestMessage(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "HTTP request failed with server error"
	case statusCode >= 400:
		return "HTTP request failed with client error"
	case statusCode >= 300:
		return "HTTP request redirected"
	case statusCode >= 200:
		return "HTTP request completed successfully"
	default:
		return "HTTP request completed"
	}
}

func extractUserIDFromRequest(r *http.Request) string {
	// This would typically extract from JWT token
	// For now, check context if auth middleware has already set it
	if userID, ok := r.Context().Value("user_id").(string); ok {
		return userID
	}
	return ""
}
