# Structured Logging & Correlation ID Implementation

## Overview
This document describes the implementation of structured logging with correlation IDs throughout the D&D Game backend.

**Implementation Status**: ✅ Complete (January 11, 2025)

## Components Implemented

### 1. Request/Correlation ID Middleware
**Location**: `internal/middleware/logging_v2.go`
- Generates or extracts Request ID from `X-Request-ID` header
- Generates or extracts Correlation ID from `X-Correlation-ID` header
- Falls back to using Request ID as Correlation ID if not provided
- Adds both IDs to request context and response headers

### 2. Enhanced Logger (LoggerV2)
**Location**: `pkg/logger/logger_enhanced.go`
- Supports structured logging with zerolog
- Context-aware logging that automatically includes:
  - Request ID
  - Correlation ID
  - User ID
  - Session ID
  - Character ID
- Configurable log levels and output formats

### 3. Database Query Logging
**Location**: `internal/database/connection.go`
- Integrated logging directly into the DB type
- Logs all database operations when logger is configured
- Includes query text, execution time, and argument count
- Preserves context (including correlation IDs) in logs
- Truncates long queries for readability
- Supports both regular and rebind query methods

### 4. WebSocket Logging
**Location**: `internal/websocket/handler_v2.go`
- Logs WebSocket connection attempts, upgrades, and authentication
- Includes client metadata (IP, user agent, origin)
- Tracks connection lifecycle with structured events

### 5. Business Event Logging
- Structured logging available in all handlers via context
- LoggerV2 automatically includes correlation IDs from context
- Handlers can log important business events with full context
- Example use cases:
  - Registration attempts and outcomes
  - Login attempts and outcomes
  - Game session lifecycle events
  - Combat actions and outcomes

## Context Propagation Flow

1. **HTTP Request arrives**
   ```
   Client Request → RequestIDMiddleware → LoggingMiddleware → Handler
   ```

2. **Context enrichment**
   - Request ID generated/extracted
   - Correlation ID generated/extracted
   - User ID added after authentication
   - Session/Character IDs added from request data

3. **Service layer**
   - Services receive context from handlers
   - Context passed to all database operations
   - Context passed to external service calls

4. **Database operations**
   - LoggedDB extracts context for all queries
   - Logs include correlation ID for tracing

5. **WebSocket connections**
   - Initial HTTP context preserved
   - Correlation ID maintained across WebSocket lifecycle

## Configuration

### Middleware Setup (main.go)
```go
r.Use(middleware.RequestIDMiddleware)
r.Use(middleware.RequestContextMiddleware)
r.Use(middleware.LoggingMiddleware(log))
```

### Logger Configuration
```go
logConfig := logger.ConfigV2{
    Level:        "info",
    Pretty:       false,
    CallerInfo:   true,
    ServiceName:  "dnd-game-backend",
    Environment:  "production",
}
```

### Database Logging
```go
// Initialize database with logging support
db, repos, err := database.InitializeWithLogging(cfg, log)
// The database will automatically log all queries with correlation IDs
```

## Log Format Examples

### HTTP Request Log
```json
{
  "level": "info",
  "time": "2025-01-11T10:30:45Z",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "path": "/api/v1/auth/login",
  "status": 200,
  "duration_ms": 145,
  "message": "HTTP request completed successfully"
}
```

### Database Query Log
```json
{
  "level": "debug",
  "time": "2025-01-11T10:30:45Z",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user-123",
  "query": "SELECT * FROM users WHERE username = ?",
  "duration_ms": 5,
  "args_count": 1,
  "message": "Database query executed"
}
```

### Business Event Log
```json
{
  "level": "info",
  "time": "2025-01-11T10:30:45Z",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user-123",
  "username": "player1",
  "email": "player1@example.com",
  "message": "User registered successfully"
}
```

## Best Practices

1. **Always pass context**: Ensure context is passed through all function calls
2. **Log at appropriate levels**:
   - DEBUG: Detailed operational info (queries, internal state)
   - INFO: Business events (login, registration, game actions)
   - WARN: Recoverable issues (validation failures, rate limits)
   - ERROR: Unrecoverable errors requiring attention
3. **Avoid logging sensitive data**: Never log passwords, tokens, or PII
4. **Include relevant context**: Add fields that help debug issues
5. **Use structured fields**: Prefer structured fields over string concatenation

## Monitoring & Observability

With this implementation, you can:
1. **Trace requests**: Follow a request through all layers using correlation ID
2. **Debug issues**: See exact database queries and timings
3. **Monitor performance**: Track slow queries and requests
4. **Audit trail**: Track important business events
5. **Correlate errors**: Link errors across services and layers

## Future Enhancements

1. **OpenTelemetry Integration**: Export traces to Jaeger/Zipkin
2. **Metrics Collection**: Add Prometheus metrics for monitoring
3. **Log Aggregation**: Ship logs to ELK stack or similar
4. **Sampling**: Implement adaptive sampling for high-traffic scenarios
5. **Distributed Tracing**: Extend correlation IDs to external services