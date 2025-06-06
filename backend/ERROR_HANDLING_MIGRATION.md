# Error Handling Migration Guide

This guide explains how to migrate from the old error handling patterns to the new standardized system.

## Overview

The new error handling system provides:
- Consistent error responses across all endpoints
- Proper error codes for debugging
- Request ID tracking
- Structured validation errors
- Better logging with context
- Type-safe error handling

## Quick Migration Examples

### 1. Basic Handler Migration

**OLD:**
```go
func (h *CharacterHandler) GetCharacter(w http.ResponseWriter, r *http.Request) {
    characterID := chi.URLParam(r, "id")
    
    character, err := h.characterService.GetCharacterByID(r.Context(), characterID)
    if err != nil {
        sendErrorResponse(w, http.StatusNotFound, "Character not found")
        return
    }
    
    sendJSONResponse(w, http.StatusOK, character)
}
```

**NEW:**
```go
func (h *CharacterHandler) GetCharacter(w http.ResponseWriter, r *http.Request) error {
    characterID := chi.URLParam(r, "id")
    
    character, err := h.characterService.GetCharacterByID(r.Context(), characterID)
    if err != nil {
        return errors.NewNotFoundError("character").WithCode(string(errors.ErrCodeCharacterNotFound))
    }
    
    response.JSON(w, r, http.StatusOK, character)
    return nil
}

// In routes:
r.Get("/{id}", middleware.Handler(h.GetCharacter))
```

### 2. Validation Error Migration

**OLD:**
```go
func (h *CharacterHandler) CreateCharacter(w http.ResponseWriter, r *http.Request) {
    var req CreateCharacterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    // Manual validation
    if req.Name == "" {
        sendErrorResponse(w, http.StatusBadRequest, "Name is required")
        return
    }
    
    // ... rest of handler
}
```

**NEW:**
```go
func (h *CharacterHandler) CreateCharacter(w http.ResponseWriter, r *http.Request) error {
    var req CreateCharacterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return errors.NewBadRequestError("Invalid request body").WithCode(string(errors.ErrCodeInvalidInput))
    }
    
    // Structured validation
    validationErrors := &errors.ValidationErrors{}
    if req.Name == "" {
        validationErrors.Add("name", "Name is required")
    }
    if req.Level < 1 || req.Level > 20 {
        validationErrors.Add("level", "Level must be between 1 and 20")
    }
    
    if validationErrors.HasErrors() {
        return validationErrors.ToAppError()
    }
    
    // ... rest of handler
    
    response.JSON(w, r, http.StatusCreated, character)
    return nil
}
```

### 3. Authentication Error Migration

**OLD:**
```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // ... authentication logic
    
    if !valid {
        sendErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
        return
    }
}
```

**NEW:**
```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
    // ... authentication logic
    
    if !valid {
        return errors.NewAuthenticationError("Invalid credentials").
            WithCode(string(errors.ErrCodeInvalidCredentials))
    }
    
    return nil
}
```

### 4. Service Error Handling

**OLD Service:**
```go
func (s *CharacterService) GetCharacterByID(ctx context.Context, id string) (*models.Character, error) {
    character, err := s.repo.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("character not found")
    }
    return character, nil
}
```

**NEW Service:**
```go
func (s *CharacterService) GetCharacterByID(ctx context.Context, id string) (*models.Character, error) {
    character, err := s.repo.GetByID(id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.NewNotFoundError("character").
                WithCode(string(errors.ErrCodeCharacterNotFound))
        }
        return nil, errors.NewInternalError("Failed to retrieve character", err).
            WithCode(string(errors.ErrCodeDatabaseError))
    }
    return character, nil
}
```

## Error Response Format

All error responses now follow this structure:

```json
{
  "success": false,
  "error": {
    "type": "NOT_FOUND",
    "code": "CHAR001",
    "message": "Character not found",
    "details": {
      "character_id": "123e4567-e89b-12d3-a456-426614174000"
    }
  },
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-01-06T15:04:05.123Z"
}
```

## Handler Types

### 1. Basic Handler
```go
func handler(w http.ResponseWriter, r *http.Request) error {
    // Handler logic
    return nil
}

// Route registration
r.Get("/path", middleware.Handler(handler))
```

### 2. Authenticated Handler
```go
func handler(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
    // Handler logic with guaranteed userID
    return nil
}

// Route registration
r.Get("/path", middleware.AuthenticatedHandler(handler))
```

### 3. Game Session Handler
```go
func handler(w http.ResponseWriter, r *http.Request, userID, sessionID uuid.UUID) error {
    // Handler logic with user and session context
    return nil
}

// Route registration
r.Get("/path", middleware.GameSessionHandler(handler))
```

### 4. DM Only Handler
```go
func handler(w http.ResponseWriter, r *http.Request, userID, sessionID uuid.UUID) error {
    // Handler logic for DM only actions
    return nil
}

// Route registration
r.Get("/path", middleware.DMOnlyHandler(handler))
```

## Common Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| AUTH001 | Invalid credentials | 401 |
| AUTH002 | Token expired | 401 |
| AUTH003 | Invalid token | 401 |
| AUTH004 | Insufficient privileges | 403 |
| USER001 | User not found | 404 |
| CHAR001 | Character not found | 404 |
| CHAR002 | Character limit reached | 400 |
| GAME001 | Game session not found | 404 |
| GAME005 | Only DM can perform this action | 403 |
| VAL001 | Validation failed | 400 |
| DB001 | Database error | 500 |
| AI001 | AI service unavailable | 503 |

## Best Practices

1. **Always use error codes**: They help with debugging and client-side error handling
   ```go
   return errors.NewNotFoundError("character").WithCode(string(errors.ErrCodeCharacterNotFound))
   ```

2. **Add context to errors**: Include relevant IDs or details
   ```go
   return errors.NewNotFoundError("character").
       WithCode(string(errors.ErrCodeCharacterNotFound)).
       WithDetails(map[string]interface{}{
           "character_id": characterID,
           "user_id": userID,
       })
   ```

3. **Use structured validation**: Group related validation errors
   ```go
   validationErrors := &errors.ValidationErrors{}
   // Add all validation errors
   if validationErrors.HasErrors() {
       return validationErrors.ToAppError()
   }
   ```

4. **Wrap internal errors**: Don't expose internal error details
   ```go
   if err != nil {
       return errors.NewInternalError("Operation failed", err)
   }
   ```

5. **Use appropriate error types**: Match the error type to the situation
   - `NewAuthenticationError` for auth failures
   - `NewAuthorizationError` for permission issues
   - `NewValidationError` for input validation
   - `NewNotFoundError` for missing resources
   - `NewConflictError` for duplicates
   - `NewInternalError` for server errors

## Migration Checklist

- [ ] Update handler signatures to return `error`
- [ ] Replace `sendErrorResponse` with appropriate error returns
- [ ] Replace `sendJSONResponse` with `response.JSON`
- [ ] Add error codes to all errors
- [ ] Update route registration to use new handler wrappers
- [ ] Implement structured validation where needed
- [ ] Update service layer to return `AppError` types
- [ ] Add request ID logging
- [ ] Update API documentation with error codes
- [ ] Test error responses match new format

## Gradual Migration

The migration helpers in `migration_helpers.go` allow for gradual migration:

1. **Phase 1**: Update imports and use migration helpers
2. **Phase 2**: Convert handlers to new signatures one by one
3. **Phase 3**: Update services to return proper errors
4. **Phase 4**: Remove migration helpers and old functions

## Testing

Update your tests to check for the new error format:

```go
func TestGetCharacter_NotFound(t *testing.T) {
    // ... setup
    
    req := httptest.NewRequest("GET", "/characters/invalid-id", nil)
    rec := httptest.NewRecorder()
    
    handler.GetCharacter(rec, req)
    
    assert.Equal(t, http.StatusNotFound, rec.Code)
    
    var response map[string]interface{}
    json.Unmarshal(rec.Body.Bytes(), &response)
    
    assert.False(t, response["success"].(bool))
    assert.Equal(t, "NOT_FOUND", response["error"].(map[string]interface{})["type"])
    assert.Equal(t, "CHAR001", response["error"].(map[string]interface{})["code"])
}
```