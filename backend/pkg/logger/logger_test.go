package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		verify func(t *testing.T, logger *Logger)
	}{
		{
			name: "default config",
			config: Config{
				Level:      "info",
				Pretty:     false,
				TimeFormat: "",
			},
			verify: func(t *testing.T, logger *Logger) {
				assert.NotNil(t, logger)
				assert.NotNil(t, logger.Logger)
			},
		},
		{
			name: "debug level",
			config: Config{
				Level:  "debug",
				Pretty: false,
			},
			verify: func(t *testing.T, logger *Logger) {
				assert.NotNil(t, logger)
				// Debug level should be set
				assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())
			},
		},
		{
			name: "invalid level defaults to info",
			config: Config{
				Level:  "invalid",
				Pretty: false,
			},
			verify: func(t *testing.T, logger *Logger) {
				assert.NotNil(t, logger)
				assert.Equal(t, zerolog.InfoLevel, zerolog.GlobalLevel())
			},
		},
		{
			name: "pretty printing enabled",
			config: Config{
				Level:  "info",
				Pretty: true,
			},
			verify: func(t *testing.T, logger *Logger) {
				assert.NotNil(t, logger)
			},
		},
		{
			name: "custom time format",
			config: Config{
				Level:      "info",
				Pretty:     false,
				TimeFormat: "2006-01-02 15:04:05",
			},
			verify: func(t *testing.T, logger *Logger) {
				assert.NotNil(t, logger)
				assert.Equal(t, "2006-01-02 15:04:05", zerolog.TimeFieldFormat)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.config)
			tt.verify(t, logger)
		})
	}
}

func TestLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{
		Logger: &zl,
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, "test-request-id")
	ctx = context.WithValue(ctx, CorrelationIDKey, "test-correlation-id")
	ctx = context.WithValue(ctx, UserIDKey, "test-user-id")

	contextLogger := logger.WithContext(ctx)
	contextLogger.Info().Msg("test message")

	var logEntry map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "test-request-id", logEntry["request_id"])
	assert.Equal(t, "test-correlation-id", logEntry["correlation_id"])
	assert.Equal(t, "test-user-id", logEntry["user_id"])
	assert.Equal(t, "test message", logEntry["message"])
}

// testLoggerWith is a helper function to test logger context methods
func testLoggerWith(t *testing.T, withFunc func(*Logger) *Logger, expectedField, expectedValue string) {
	t.Helper()
	var buf bytes.Buffer
	logger := &Logger{
		Logger: func() *zerolog.Logger { zl := zerolog.New(&buf).With().Timestamp().Logger(); return &zl }(),
	}

	contextLogger := withFunc(logger)
	contextLogger.Info().Msg("test message")

	var logEntry map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, expectedValue, logEntry[expectedField])
}

func TestLogger_WithRequestID(t *testing.T) {
	testLoggerWith(t, func(l *Logger) *Logger {
		return l.WithRequestID("req-123")
	}, "request_id", "req-123")
}

func TestLogger_WithCorrelationID(t *testing.T) {
	testLoggerWith(t, func(l *Logger) *Logger {
		return l.WithCorrelationID("corr-123")
	}, "correlation_id", "corr-123")
}

func TestLogger_WithUserID(t *testing.T) {
	testLoggerWith(t, func(l *Logger) *Logger {
		return l.WithUserID("user-456")
	}, "user_id", "user-456")
}

func TestLogger_WithError(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		Logger: func() *zerolog.Logger { zl := zerolog.New(&buf).With().Timestamp().Logger(); return &zl }(),
	}

	testErr := assert.AnError
	errorLogger := logger.WithError(testErr)
	errorLogger.Error().Msg("error occurred")

	var logEntry map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, testErr.Error(), logEntry["error"])
}

func TestLogger_WithField(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		Logger: func() *zerolog.Logger { zl := zerolog.New(&buf).With().Timestamp().Logger(); return &zl }(),
	}

	fieldLogger := logger.WithField("custom_field", "custom_value")
	fieldLogger.Info().Msg("test message")

	var logEntry map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "custom_value", logEntry["custom_field"])
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		Logger: func() *zerolog.Logger { zl := zerolog.New(&buf).With().Timestamp().Logger(); return &zl }(),
	}

	fields := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
		"field3": true,
	}

	fieldsLogger := logger.WithFields(fields)
	fieldsLogger.Info().Msg("test message")

	var logEntry map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "value1", logEntry["field1"])
	assert.Equal(t, float64(42), logEntry["field2"])
	assert.Equal(t, true, logEntry["field3"])
}

func TestInit(t *testing.T) {
	// Reset global logger
	defaultLogger = nil

	cfg := Config{
		Level:  "debug",
		Pretty: false,
	}

	Init(cfg)

	assert.NotNil(t, defaultLogger)
	assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())
}

func TestGetLogger(t *testing.T) {
	// Test with uninitialized logger
	defaultLogger = nil
	logger := GetLogger()
	assert.NotNil(t, logger)

	// Test with initialized logger
	Init(Config{Level: "warn"})
	logger2 := GetLogger()
	assert.NotNil(t, logger2)
	assert.Equal(t, defaultLogger, logger2)
}

func TestGlobalLoggerFunctions(t *testing.T) {
	var buf bytes.Buffer

	// Initialize with a buffer logger for testing
	// Use SyncWriter to ensure all writes are captured
	writer := zerolog.SyncWriter(&buf)
	zl := zerolog.New(writer).With().Timestamp().Logger().Level(zerolog.DebugLevel)
	defaultLogger = &Logger{&zl}
	// Ensure global log level allows debug messages
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	tests := []struct {
		name    string
		logFunc func() *zerolog.Event
		level   string
		message string
	}{
		{
			name:    "Debug",
			logFunc: Debug,
			level:   "debug",
			message: "debug message",
		},
		{
			name:    "Info",
			logFunc: Info,
			level:   "info",
			message: "info message",
		},
		{
			name:    "Warn",
			logFunc: Warn,
			level:   "warn",
			message: "warn message",
		},
		{
			name:    "Error",
			logFunc: Error,
			level:   "error",
			message: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc().Msg(tt.message)

			// Check if buffer is empty (could happen if logger isn't properly initialized)
			if buf.Len() == 0 {
				t.Fatal("No log output generated")
			}

			var logEntry map[string]interface{}
			require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

			assert.Equal(t, tt.level, logEntry["level"])
			assert.Equal(t, tt.message, logEntry["message"])
		})
	}
}

func TestWithContext_Global(t *testing.T) {
	var buf bytes.Buffer
	// Create a sync writer to ensure all writes go to the same buffer
	writer := zerolog.SyncWriter(&buf)
	zl := zerolog.New(writer).With().Timestamp().Logger().Level(zerolog.InfoLevel)
	defaultLogger = &Logger{&zl}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, "global-request-id")

	// Use WithContext to add context fields
	WithContext(ctx).Info().Msg("context message")

	var logEntry map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "global-request-id", logEntry["request_id"])
	assert.Equal(t, "context message", logEntry["message"])
}

func TestLogger_ChainedOperations(t *testing.T) {
	var buf bytes.Buffer
	// Use SyncWriter to ensure all writes complete
	writer := zerolog.SyncWriter(&buf)
	zl := zerolog.New(writer).With().Timestamp().Logger().Level(zerolog.InfoLevel)
	logger := &Logger{
		Logger: &zl,
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Test chaining multiple operations
	logger.
		WithRequestID("req-chain").
		WithUserID("user-chain").
		WithField("operation", "test").
		WithFields(map[string]interface{}{
			"count":  10,
			"active": true,
		}).
		Info().
		Msg("chained operations")

	var logEntry map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "req-chain", logEntry["request_id"])
	assert.Equal(t, "user-chain", logEntry["user_id"])
	assert.Equal(t, "test", logEntry["operation"])
	assert.Equal(t, float64(10), logEntry["count"])
	assert.Equal(t, true, logEntry["active"])
}

func TestLogger_EmptyContext(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		Logger: func() *zerolog.Logger { zl := zerolog.New(&buf).With().Timestamp().Logger(); return &zl }(),
	}

	// Test with empty context values
	ctx := context.Background()
	contextLogger := logger.WithContext(ctx)
	contextLogger.Info().Msg("empty context")

	logOutput := buf.String()
	assert.NotContains(t, logOutput, "request_id")
	assert.NotContains(t, logOutput, "user_id")
}

func TestLogger_NilError(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger().Level(zerolog.InfoLevel)
	logger := &Logger{
		Logger: &zl,
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Test with nil error - WithError should handle nil gracefully
	errorLogger := logger.WithError(nil)
	errorLogger.Info().Msg("nil error test")

	// Parse the log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "nil error test", logEntry["message"])
	// When error is nil, the error field should not be present
	_, hasError := logEntry["error"]
	assert.False(t, hasError, "error field should not be present when error is nil")
}

func TestLogger_MultipleLogLevels(t *testing.T) {
	// Test that log level filtering works correctly
	levels := []struct {
		configLevel string
		testLevel   zerolog.Level
		shouldLog   bool
	}{
		{"debug", zerolog.DebugLevel, true},
		{"debug", zerolog.InfoLevel, true},
		{"info", zerolog.DebugLevel, false},
		{"info", zerolog.InfoLevel, true},
		{"warn", zerolog.InfoLevel, false},
		{"warn", zerolog.WarnLevel, true},
		{"error", zerolog.WarnLevel, false},
		{"error", zerolog.ErrorLevel, true},
	}

	for _, test := range levels {
		t.Run(test.configLevel+"_"+test.testLevel.String(), func(t *testing.T) {
			var buf bytes.Buffer

			// Create logger with specific level
			logger := New(Config{
				Level:  test.configLevel,
				Pretty: false,
			})

			// Override the writer for testing
			zl := zerolog.New(&buf).Level(test.testLevel)
			logger.Logger = &zl

			// Log at test level
			switch test.testLevel {
			case zerolog.DebugLevel:
				logger.Debug().Msg("test")
			case zerolog.InfoLevel:
				logger.Info().Msg("test")
			case zerolog.WarnLevel:
				logger.Warn().Msg("test")
			case zerolog.ErrorLevel:
				logger.Error().Msg("test")
			case zerolog.FatalLevel, zerolog.PanicLevel:
				// Fatal and Panic levels are not tested as they would terminate the test
				// These levels are included here to satisfy the exhaustive check
			case zerolog.NoLevel, zerolog.Disabled:
				// NoLevel and Disabled don't produce output
				// These levels are included here to satisfy the exhaustive check
			case zerolog.TraceLevel:
				logger.Trace().Msg("test")
			}

			if test.shouldLog {
				assert.NotEmpty(t, buf.String(), "Expected log output")
			} else {
				assert.Empty(t, buf.String(), "Expected no log output")
			}
		})
	}
}

func BenchmarkLogger_WithContext(b *testing.B) {
	logger := New(Config{Level: "info", Pretty: false})
	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, "bench-request-id")
	ctx = context.WithValue(ctx, CorrelationIDKey, "bench-correlation-id")
	ctx = context.WithValue(ctx, UserIDKey, "bench-user-id")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithContext(ctx).Info().Msg("benchmark message")
	}
}

func BenchmarkLogger_WithFields(b *testing.B) {
	logger := New(Config{Level: "info", Pretty: false})
	fields := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
		"field3": true,
		"field4": 3.14,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(fields).Info().Msg("benchmark message")
	}
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()

	ctx = ContextWithRequestID(ctx, "req-123")
	ctx = ContextWithCorrelationID(ctx, "corr-456")

	assert.Equal(t, "req-123", ctx.Value(RequestIDKey))
	assert.Equal(t, "corr-456", ctx.Value(CorrelationIDKey))
	assert.Equal(t, "corr-456", GetCorrelationIDFromContext(ctx))
}
