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

	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
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
		// Add user ID to context
		r = addUserIDToContext(r)
		
		// Add session ID to context
		r = addSessionIDToContext(r)
		
		// Add character ID to context
		r = addCharacterIDToContext(r)

		next.ServeHTTP(w, r)
	})
}

func addUserIDToContext(r *http.Request) *http.Request {
	userID := extractUserIDFromRequest(r)
	if userID != "" {
		ctx := logger.ContextWithUserID(r.Context(), userID)
		return r.WithContext(ctx)
	}
	return r
}

func addSessionIDToContext(r *http.Request) *http.Request {
	sessionID := extractSessionID(r)
	if sessionID != "" {
		ctx := logger.ContextWithSessionID(r.Context(), sessionID)
		return r.WithContext(ctx)
	}
	return r
}

func extractSessionID(r *http.Request) string {
	if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
		return sessionID
	}
	return r.URL.Query().Get("session_id")
}

func addCharacterIDToContext(r *http.Request) *http.Request {
	characterID := extractCharacterID(r)
	if characterID != "" {
		ctx := logger.ContextWithCharacterID(r.Context(), characterID)
		return r.WithContext(ctx)
	}
	return r
}

func extractCharacterID(r *http.Request) string {
	// Check header first
	if characterID := r.Header.Get("X-Character-ID"); characterID != "" {
		return characterID
	}
	
	// Try to extract from path
	return extractCharacterIDFromPath(r.URL.Path)
}

func extractCharacterIDFromPath(path string) string {
	if !strings.Contains(path, "/characters/") {
		return ""
	}
	
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "characters" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// DatabaseQueryLogger logs database queries.
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

var sensitiveParams = []string{"password", "token", "secret", "api_key", "access_token", "refresh_token"}

func sanitizeQuery(query string) string {
	if query == "" {
		return ""
	}

	params := strings.Split(query, "&")
	sanitized := make([]string, 0, len(params))

	for _, param := range params {
		sanitizedParam := sanitizeQueryParam(param)
		sanitized = append(sanitized, sanitizedParam)
	}

	return strings.Join(sanitized, "&")
}

func sanitizeQueryParam(param string) string {
	parts := strings.SplitN(param, "=", 2)
	if len(parts) != 2 {
		return param
	}
	
	if isSensitiveParam(parts[0]) {
		return parts[0] + "=[REDACTED]"
	}
	return param
}

func isSensitiveParam(key string) bool {
	lowerKey := strings.ToLower(key)
	for _, sensitive := range sensitiveParams {
		if strings.Contains(lowerKey, sensitive) {
			return true
		}
	}
	return false
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
