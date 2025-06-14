package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Enhanced context keys
const (
	SessionIDKey   contextKey = "session_id"
	CharacterIDKey contextKey = "character_id"
	ServiceKey     contextKey = "service"
	MethodKey      contextKey = "method"
)

// LoggerV2 is an enhanced logger with additional features
type LoggerV2 struct {
	*zerolog.Logger
	config ConfigV2
}

// ConfigV2 holds enhanced logger configuration
type ConfigV2 struct {
	Level        string  // Log level: debug, info, warn, error, fatal
	Pretty       bool    // Pretty print for development
	TimeFormat   string  // Time format
	CallerInfo   bool    // Include caller information
	StackTrace   bool    // Include stack trace for errors
	Output       string  // Output: stdout, stderr, file path
	MaxSize      int     // Max size in MB for file rotation (if Output is file)
	MaxBackups   int     // Max number of backup files
	MaxAge       int     // Max age in days for backup files
	Compress     bool    // Compress backup files
	SamplingRate float32 // Sampling rate for debug logs (0.0 to 1.0)
	ServiceName  string  // Service name to include in all logs
	Environment  string  // Environment: development, staging, production
	Fields       Fields  // Default fields to include in all logs
}

// Fields represents default fields
type Fields map[string]interface{}

// DefaultConfig returns a default configuration
func DefaultConfig() ConfigV2 {
	return ConfigV2{
		Level:        "info",
		Pretty:       false,
		TimeFormat:   time.RFC3339Nano,
		CallerInfo:   true,
		StackTrace:   true,
		Output:       "stdout",
		SamplingRate: 1.0,
		ServiceName:  "dnd-game-backend",
		Environment:  "development",
	}
}

// NewV2 creates a new enhanced logger
func NewV2(cfg ConfigV2) (*LoggerV2, error) {
	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure time format
	if cfg.TimeFormat != "" {
		zerolog.TimeFieldFormat = cfg.TimeFormat
	}

	// Enable stack trace marshaling
	if cfg.StackTrace {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	}

	// Configure output
	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// File output with rotation would require additional setup
		// For now, fallback to stdout
		output = os.Stdout
	}

	// Create logger with pretty printing if needed
	var zl zerolog.Logger
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: cfg.TimeFormat,
			FormatLevel: func(i interface{}) string {
				return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
			},
			FormatFieldName: func(i interface{}) string {
				return fmt.Sprintf("%s:", i)
			},
		}
	}

	// Create base logger
	zl = zerolog.New(output).With().Timestamp().Logger()

	// Add default fields
	if cfg.ServiceName != "" {
		zl = zl.With().Str("service", cfg.ServiceName).Logger()
	}
	if cfg.Environment != "" {
		zl = zl.With().Str("env", cfg.Environment).Logger()
	}

	// Add hostname
	if hostname, err := os.Hostname(); err == nil {
		zl = zl.With().Str("hostname", hostname).Logger()
	}

	// Add custom default fields
	for k, v := range cfg.Fields {
		zl = zl.With().Interface(k, v).Logger()
	}

	// Add caller info if enabled
	if cfg.CallerInfo {
		zl = zl.With().CallerWithSkipFrameCount(3).Logger()
	}

	// Configure sampling for debug logs
	if cfg.SamplingRate < 1.0 && level == zerolog.DebugLevel {
		sampled := zl.Sample(&zerolog.BasicSampler{N: uint32(1.0 / cfg.SamplingRate)})
		zl = sampled
	}

	return &LoggerV2{
		Logger: &zl,
		config: cfg,
	}, nil
}

// WithContext enriches the logger with context values
func (l *LoggerV2) WithContext(ctx context.Context) *LoggerV2 {
	zl := l.With()

	// Add all context values
	contextKeys := []struct {
		key  contextKey
		name string
	}{
		{RequestIDKey, "request_id"},
		{CorrelationIDKey, "correlation_id"},
		{UserIDKey, "user_id"},
		{SessionIDKey, "session_id"},
		{CharacterIDKey, "character_id"},
		{ServiceKey, "service"},
		{MethodKey, "method"},
	}

	for _, ck := range contextKeys {
		if value, ok := ctx.Value(ck.key).(string); ok && value != "" {
			zl = zl.Str(ck.name, value)
		}
	}

	logger := zl.Logger()
	return &LoggerV2{Logger: &logger, config: l.config}
}

// WithOperation adds operation context
func (l *LoggerV2) WithOperation(service, method string) *LoggerV2 {
	logger := l.With().
		Str("service", service).
		Str("method", method).
		Logger()
	return &LoggerV2{Logger: &logger, config: l.config}
}

// WithGameContext adds game-specific context
func (l *LoggerV2) WithGameContext(sessionID, characterID string) *LoggerV2 {
	zl := l.With()
	if sessionID != "" {
		zl = zl.Str("session_id", sessionID)
	}
	if characterID != "" {
		zl = zl.Str("character_id", characterID)
	}
	logger := zl.Logger()
	return &LoggerV2{Logger: &logger, config: l.config}
}

// LogHTTPRequest logs HTTP request details
func (l *LoggerV2) LogHTTPRequest(method, path string, statusCode int, duration time.Duration, fields ...map[string]interface{}) {
	event := l.Info().
		Str("method", method).
		Str("path", path).
		Int("status", statusCode).
		Dur("duration", duration)

	// Add additional fields if provided
	if len(fields) > 0 {
		for k, v := range fields[0] {
			event = event.Interface(k, v)
		}
	}

	// Log with appropriate level based on status code
	switch {
	case statusCode >= 500:
		event.Msg("HTTP request failed")
	case statusCode >= 400:
		event.Msg("HTTP request client error")
	case statusCode >= 300:
		event.Msg("HTTP request redirected")
	default:
		event.Msg("HTTP request completed")
	}
}

// LogDatabaseQuery logs database query details
func (l *LoggerV2) LogDatabaseQuery(query string, duration time.Duration, err error, args ...interface{}) {
	event := l.Debug().
		Str("query", truncateQuery(query)).
		Dur("duration", duration).
		Int("args_count", len(args))

	if err != nil {
		event.Err(err).Msg("Database query failed")
	} else {
		event.Msg("Database query executed")
	}
}

// LogAIOperation logs AI operation details
func (l *LoggerV2) LogAIOperation(operation string, provider string, duration time.Duration, tokens int, err error) {
	event := l.Info().
		Str("operation", operation).
		Str("provider", provider).
		Dur("duration", duration).
		Int("tokens", tokens)

	if err != nil {
		event.Err(err).Msg("AI operation failed")
	} else {
		event.Msg("AI operation completed")
	}
}

// LogWebSocketEvent logs WebSocket events
func (l *LoggerV2) LogWebSocketEvent(eventType string, clientID string, data interface{}) {
	l.Debug().
		Str("event_type", eventType).
		Str("client_id", clientID).
		Interface("data", data).
		Msg("WebSocket event")
}

// LogGameEvent logs game-specific events
func (l *LoggerV2) LogGameEvent(eventType string, sessionID string, details map[string]interface{}) {
	event := l.Info().
		Str("event_type", eventType).
		Str("session_id", sessionID)

	for k, v := range details {
		event = event.Interface(k, v)
	}

	event.Msg("Game event occurred")
}

// Helper functions

// truncateQuery truncates long queries for logging
func truncateQuery(query string) string {
	const maxLength = 200
	query = strings.TrimSpace(query)
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")
	query = strings.Join(strings.Fields(query), " ")

	if len(query) > maxLength {
		return query[:maxLength] + "..."
	}
	return query
}

// GetCaller returns the caller information
func GetCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// ContextWithUserID adds user ID to context
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// ContextWithSessionID adds session ID to context
func ContextWithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

// ContextWithCharacterID adds character ID to context
func ContextWithCharacterID(ctx context.Context, characterID string) context.Context {
	return context.WithValue(ctx, CharacterIDKey, characterID)
}

// GetRequestIDFromContext retrieves request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}
