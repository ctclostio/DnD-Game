package logger

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ContextKey for storing request ID in context
type contextKey string

const (
	RequestIDKey     contextKey = "request_id"
	CorrelationIDKey contextKey = "correlation_id"
	UserIDKey        contextKey = "user_id"
)

// Logger wraps zerolog logger with additional functionality
type Logger struct {
	*zerolog.Logger
}

// Config holds logger configuration
type Config struct {
	Level      string
	Pretty     bool
	TimeFormat string
}

// New creates a new logger instance
func New(cfg Config) *Logger {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure time format
	if cfg.TimeFormat != "" {
		zerolog.TimeFieldFormat = cfg.TimeFormat
	} else {
		zerolog.TimeFieldFormat = time.RFC3339
	}

	// Create logger
	var zl zerolog.Logger
	if cfg.Pretty {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		zl = zerolog.New(output).With().Timestamp().Logger()
	} else {
		zl = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return &Logger{&zl}
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	zl := l.Logger.With()

	// Add request ID if present
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		zl = zl.Str("request_id", requestID)
	}

	// Add correlation ID if present
	if corrID, ok := ctx.Value(CorrelationIDKey).(string); ok && corrID != "" {
		zl = zl.Str("correlation_id", corrID)
	}

	// Add user ID if present
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		zl = zl.Str("user_id", userID)
	}

	logger := zl.Logger()
	return &Logger{&logger}
}

// WithRequestID adds request ID to logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	logger := l.Logger.With().Str("request_id", requestID).Logger()
	return &Logger{&logger}
}

// WithCorrelationID adds correlation ID to logger
func (l *Logger) WithCorrelationID(correlationID string) *Logger {
	logger := l.Logger.With().Str("correlation_id", correlationID).Logger()
	return &Logger{&logger}
}

// WithUserID adds user ID to logger
func (l *Logger) WithUserID(userID string) *Logger {
	logger := l.Logger.With().Str("user_id", userID).Logger()
	return &Logger{&logger}
}

// WithError adds error to logger
func (l *Logger) WithError(err error) *Logger {
	logger := l.Logger.With().Err(err).Logger()
	return &Logger{&logger}
}

// WithField adds a field to logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	logger := l.Logger.With().Interface(key, value).Logger()
	return &Logger{&logger}
}

// WithFields adds multiple fields to logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	logContext := l.Logger.With()
	for k, v := range fields {
		logContext = logContext.Interface(k, v)
	}
	logger := logContext.Logger()
	return &Logger{&logger}
}

// Global logger instance
var (
	defaultLogger *Logger
	loggerMutex   sync.Mutex
)

// Init initializes the global logger
func Init(cfg Config) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	
	defaultLogger = New(cfg)
	// Set global logger for zerolog
	log.Logger = *defaultLogger.Logger
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	
	if defaultLogger == nil {
		// Initialize with default config if not set
		defaultLogger = New(Config{
			Level:  "info",
			Pretty: false,
		})
		// Set global logger for zerolog
		log.Logger = *defaultLogger.Logger
	}
	return defaultLogger
}

// Helper functions for global logger

// Debug logs a debug message
func Debug() *zerolog.Event {
	return GetLogger().Logger.Debug()
}

// Info logs an info message
func Info() *zerolog.Event {
	return GetLogger().Logger.Info()
}

// Warn logs a warning message
func Warn() *zerolog.Event {
	return GetLogger().Logger.Warn()
}

// Error logs an error message
func Error() *zerolog.Event {
	return GetLogger().Logger.Error()
}

// Fatal logs a fatal message and exits
func Fatal() *zerolog.Event {
	return GetLogger().Logger.Fatal()
}

// WithContext returns a logger with context
func WithContext(ctx context.Context) *Logger {
	return GetLogger().WithContext(ctx)
}

// ContextWithRequestID adds request ID to context
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// ContextWithCorrelationID adds correlation ID to context
func ContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetCorrelationIDFromContext retrieves correlation ID from context
func GetCorrelationIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}
