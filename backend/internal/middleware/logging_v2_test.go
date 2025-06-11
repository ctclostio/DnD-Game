package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/middleware"
	"github.com/your-username/dnd-game/backend/pkg/logger"
)

func TestLoggingMiddleware_CorrelationID(t *testing.T) {
	// Create logger for testing
	logConfig := logger.ConfigV2{
		Level:        "debug",
		Pretty:       false,
		ServiceName:  "test",
		Environment:  "test",
		SamplingRate: 1.0, // Ensure no sampling issues
	}
	log, err := logger.NewV2(logConfig)
	require.NoError(t, err)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context has IDs
		ctx := r.Context()
		
		// Get IDs from context using logger functions
		reqID := logger.GetRequestIDFromContext(ctx)
		corrID := logger.GetCorrelationIDFromContext(ctx)
		
		// Write them to response for verification
		w.Header().Set("X-Context-Request-ID", reqID)
		w.Header().Set("X-Context-Correlation-ID", corrID)
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply logging middleware
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
			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.requestID != "" {
				req.Header.Set("X-Request-ID", tt.requestID)
			}
			if tt.correlationID != "" {
				req.Header.Set("X-Correlation-ID", tt.correlationID)
			}
			
			// Create response recorder
			rr := httptest.NewRecorder()
			
			// Serve request
			wrappedHandler.ServeHTTP(rr, req)
			
			// Check response headers
			respReqID := rr.Header().Get("X-Request-ID")
			respCorrID := rr.Header().Get("X-Correlation-ID")
			
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
			
			// Verify context propagation
			ctxReqID := rr.Header().Get("X-Context-Request-ID")
			ctxCorrID := rr.Header().Get("X-Context-Correlation-ID")
			
			assert.Equal(t, respReqID, ctxReqID, "Context should have same request ID as response")
			assert.Equal(t, respCorrID, ctxCorrID, "Context should have same correlation ID as response")
		})
	}
}

func TestRequestContextMiddleware(t *testing.T) {
	// Create test handler that checks context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Context should be enriched by middleware
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Apply middleware
	wrappedHandler := middleware.RequestContextMiddleware(handler)
	
	// Basic test to ensure middleware doesn't break
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Session-ID", "session-123")
	req.Header.Set("X-Character-ID", "char-789")
	
	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
}