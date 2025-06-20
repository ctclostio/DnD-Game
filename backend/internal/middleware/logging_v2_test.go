package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/middleware"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// Header constants
const (
	headerRequestID     = "X-Request-ID"
	headerCorrelationID = "X-Correlation-ID"
)

func TestLoggingMiddleware_CorrelationID(t *testing.T) {
	log := createTestLogger(t)
	handler := createIDVerificationHandler()
	mw := middleware.LoggingMiddleware(log)
	wrappedHandler := mw(handler)

	tests := []struct {
		name              string
		requestID         string
		correlationID     string
		expectedReqID     bool
		expectedCorrID    bool
		checkCorrIDEquals bool // Should correlation ID equal request ID
	}{
		{
			name:              "With both IDs provided",
			requestID:         "req-123",
			correlationID:     "corr-456",
			expectedReqID:     true,
			expectedCorrID:    true,
			checkCorrIDEquals: false,
		},
		{
			name:              "With only request ID",
			requestID:         "req-789",
			correlationID:     "",
			expectedReqID:     true,
			expectedCorrID:    true,
			checkCorrIDEquals: true,
		},
		{
			name:              "With neither ID",
			requestID:         "",
			correlationID:     "",
			expectedReqID:     true,
			expectedCorrID:    true,
			checkCorrIDEquals: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(tt.requestID, tt.correlationID)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			verifyResponseIDs(t, rr, tt)
			verifyContextPropagation(t, rr)
		})
	}
}

func createTestLogger(t *testing.T) *logger.LoggerV2 {
	logConfig := logger.ConfigV2{
		Level:        "debug",
		Pretty:       false,
		ServiceName:  "test",
		Environment:  "test",
		SamplingRate: 1.0,
	}
	log, err := logger.NewV2(&logConfig)
	require.NoError(t, err)
	return log
}

func createIDVerificationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqID := logger.GetRequestIDFromContext(ctx)
		corrID := logger.GetCorrelationIDFromContext(ctx)

		w.Header().Set("X-Context-Request-ID", reqID)
		w.Header().Set("X-Context-Correlation-ID", corrID)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}
}

func createTestRequest(requestID, correlationID string) *http.Request {
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	if requestID != "" {
		req.Header.Set(headerRequestID, requestID)
	}
	if correlationID != "" {
		req.Header.Set(headerCorrelationID, correlationID)
	}
	return req
}

func verifyResponseIDs(t *testing.T, rr *httptest.ResponseRecorder, tt struct {
	name              string
	requestID         string
	correlationID     string
	expectedReqID     bool
	expectedCorrID    bool
	checkCorrIDEquals bool
}) {
	respReqID := rr.Header().Get(headerRequestID)
	respCorrID := rr.Header().Get(headerCorrelationID)

	assert.NotEmpty(t, respReqID, "Response should have X-Request-ID")
	assert.NotEmpty(t, respCorrID, "Response should have X-Correlation-ID")

	if tt.requestID != "" {
		assert.Equal(t, tt.requestID, respReqID)
	}
	if tt.correlationID != "" {
		assert.Equal(t, tt.correlationID, respCorrID)
	}
	if tt.checkCorrIDEquals {
		assert.Equal(t, respReqID, respCorrID, "Correlation ID should equal Request ID when not provided")
	}
}

func verifyContextPropagation(t *testing.T, rr *httptest.ResponseRecorder) {
	respReqID := rr.Header().Get(headerRequestID)
	respCorrID := rr.Header().Get(headerCorrelationID)
	ctxReqID := rr.Header().Get("X-Context-Request-ID")
	ctxCorrID := rr.Header().Get("X-Context-Correlation-ID")

	assert.Equal(t, respReqID, ctxReqID, "Context should have same request ID as response")
	assert.Equal(t, respCorrID, ctxCorrID, "Context should have same correlation ID as response")
}

func TestRequestContextMiddleware(t *testing.T) {
	// Create test handler that checks context
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Context should be enriched by middleware
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Apply middleware
	wrappedHandler := middleware.RequestContextMiddleware(handler)

	// Basic test to ensure middleware doesn't break
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Session-ID", "session-123")
	req.Header.Set("X-Character-ID", "char-789")

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
