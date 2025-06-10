# Structured Logging Implementation Locations

## Overview
This document identifies locations in the backend code where structured logging should be implemented to replace existing logging patterns or add missing logs.

## Critical Issues Found

### 1. Files Using log.Printf/log.Println
These files are using standard library logging instead of the structured logger:

- **internal/database/init.go**
  - Lines 37, 47, 55: Using log.Printf/log.Println for database connection and migration status
  - Should use structured logger with fields for retry attempts, connection details

- **internal/websocket/handler.go**
  - Lines 71, 76, 100, 108, 129, 140: Using log.Printf for WebSocket events
  - Critical auth flow logging that should include structured fields

- **internal/websocket/hub.go**
  - Lines 57, 66, 72, 102: Using log.Printf for client connection events
  - Should include structured fields for client ID, room ID, etc.

- **internal/services/ai_battle_map_generator.go**
  - Lines 42, 48: Error logging without context
  - Should use structured logging with request context

- **internal/services/ai_campaign_manager.go**
  - Similar pattern of using log.Printf for errors

### 2. Silent Error Handling (No Logging)
Many service methods return errors without logging them, making debugging difficult:

#### Authentication & Security Operations
- **internal/services/user.go**
  - Register/Login operations have no logging for security events
  - Failed login attempts should be logged with structured fields
  - User creation/registration should be logged

- **internal/services/refresh_token.go**
  - Token operations (create/validate/revoke) have no logging
  - Security-critical operations that must be logged

#### Database Operations
- **internal/database/user_repository.go**
  - Database errors wrapped but not logged
  - Should log with query context and parameters (sanitized)

- **internal/database/character_repository.go**
  - CRUD operations without logging
  - Should log character creation/updates with user context

#### Game Operations
- **internal/services/combat.go**
  - Combat state changes not logged
  - Should log combat events for analytics

- **internal/services/game_session.go**
  - Session lifecycle events not logged
  - Should log session creation/join/leave events

### 3. Handler Error Responses Without Logging
Handlers return errors to clients but don't log them for monitoring:

- **internal/handlers/auth.go**
  - Lines 52-59, 68-70, 74-76: Errors sent to client but not logged
  - Authentication failures should be logged for security monitoring

- **internal/handlers/character.go**
  - Character CRUD operations errors not logged
  - Should log with user context and operation details

### 4. Critical Operations Missing Logs
Operations that should always be logged but currently aren't:

#### Startup/Shutdown
- **cmd/server/main.go**
  - Server startup configuration not logged
  - Should log config values (sanitized) and initialization status

#### WebSocket Events
- **internal/websocket/dm_assistant_handler.go**
  - DM assistant commands not logged
  - Should log command execution with context

#### AI Service Calls
- **internal/services/ai_*.go files**
  - AI provider calls not logged with latency/cost metrics
  - Should add structured logging for monitoring AI usage

### 5. Database Migration Logging
- **internal/database/migrate.go**
  - Migration steps should have detailed logging
  - Success/failure of each migration should be logged

## Recommended Logging Patterns

### 1. Authentication/Security Events
```go
logger.Info("user login attempt",
    "username", username,
    "ip", clientIP,
    "user_agent", userAgent,
    "success", success,
    "reason", failureReason,
)
```

### 2. Database Operations
```go
logger.Debug("executing query",
    "operation", "user_create",
    "duration_ms", duration.Milliseconds(),
    "affected_rows", rowsAffected,
)
```

### 3. Error Logging
```go
logger.Error("operation failed",
    "operation", operationName,
    "user_id", userID,
    "error", err.Error(),
    "stack_trace", debug.Stack(),
)
```

### 4. AI Service Calls
```go
logger.Info("ai service call",
    "provider", providerName,
    "operation", operationType,
    "tokens_used", tokenCount,
    "latency_ms", latency.Milliseconds(),
    "cost", estimatedCost,
)
```

## Priority Order for Implementation

1. **High Priority (Security & Debugging)**
   - Authentication handlers and services
   - Database error logging
   - WebSocket connection/auth events

2. **Medium Priority (Operations)**
   - Game session lifecycle
   - Combat events
   - AI service calls

3. **Low Priority (Nice to Have)**
   - Character CRUD operations
   - Dice roll logging
   - Inventory changes

## Next Steps

1. Import the enhanced logger package in identified files
2. Replace log.Printf/Println with structured logger calls
3. Add logging to silent error paths
4. Ensure sensitive data is not logged (passwords, tokens)
5. Add correlation IDs for request tracing