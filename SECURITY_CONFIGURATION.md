# Security Configuration Guide

## Overview
This document outlines the security configurations and best practices implemented in the D&D Game backend.

## Environment-Based Security

### Production Mode (Default)
The server now defaults to production mode for safety. To run in development mode, explicitly set `ENV=development`.

```bash
# Production (default)
./server

# Development mode
ENV=development ./server
```

### Security Features by Environment

| Feature | Production | Development |
|---------|------------|-------------|
| CSRF Cookie Secure Flag | Always `true` | `false` (for local testing) |
| Swagger Documentation | Disabled | Enabled at `/swagger` |
| Security Headers HSTS | Enabled | Disabled |
| CSP unsafe-inline | Disabled | Enabled |
| Mock AI Provider | Forbidden | Allowed |
| Database SSL | Required | Optional |
| JWT Secret Length | Min 64 chars | Min 32 chars |
| Debug Logging | Disabled | Enabled |

## Cookie Security

### CSRF Protection
- **Production**: CSRF cookies always have `Secure=true` flag
- **Development**: Secure flag is disabled for local HTTP testing
- **Implementation**: Based on environment, not TLS detection

```go
http.SetCookie(w, &http.Cookie{
    Name:     csrfCookieName,
    Value:    token,
    Path:     "/",
    HttpOnly: false, // Must be readable by JavaScript
    Secure:   isProduction, // Environment-based
    SameSite: http.SameSiteStrictMode,
    MaxAge:   int(csrfTokenTTL.Seconds()),
})
```

## Production Validations

The following validations are enforced when `ENV=production`:

1. **Mock Providers Forbidden**
   - AI provider cannot be "mock"
   - Prevents accidental deployment with test services

2. **Database Security**
   - SSL mode must not be "disable"
   - Ensures encrypted database connections

3. **JWT Security**
   - Secret must be at least 64 characters (vs 32 in dev)
   - Provides stronger token security

## API Documentation Security

### Swagger/OpenAPI
- **Production**: Completely disabled
- **Development**: Available at `/swagger` and `/api/v1/swagger.json`
- **Rationale**: Prevents exposure of API structure in production

## Error Handling Security

### Stack Traces
- Never exposed to clients in any environment
- Logged internally with appropriate log levels
- Generic error messages sent to clients

### Error Response Structure
```json
{
  "success": false,
  "error": {
    "type": "validation",
    "code": "VALIDATION_FAILED",
    "message": "User-friendly message",
    "details": {} // Optional, never includes stack traces
  },
  "requestId": "uuid",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Security Headers

Configured in `middleware/security.go`:

### Production Headers
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- `Content-Security-Policy`: Strict policy without unsafe-inline
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`

### Development Headers
- HSTS disabled for local testing
- CSP allows unsafe-inline for development tools
- Other security headers remain active

## Configuration Best Practices

### Required Environment Variables
```bash
# Production deployment checklist
ENV=production
JWT_SECRET=<minimum-64-character-secret>
DB_PASSWORD=<strong-password>
DB_SSLMODE=require
AI_PROVIDER=openai|anthropic|openrouter
AI_API_KEY=<real-api-key>
```

### Development Mode Warning
When running in development mode, the server logs prominent warnings:
```
⚠️  SERVER IS RUNNING IN DEVELOPMENT MODE - NOT SUITABLE FOR PRODUCTION
⚠️  Security features are relaxed. Set ENV=production for production use
```

## Docker Security

See [DOCKER_PERMISSIONS_SECURITY.md](DOCKER_PERMISSIONS_SECURITY.md) and [DOCKER_GLOB_SECURITY.md](DOCKER_GLOB_SECURITY.md) for container-specific security configurations.

## Monitoring and Compliance

### SonarCloud Integration
- Security hotspots are regularly reviewed
- Code smells and vulnerabilities tracked
- Current security rating maintained

### Security Checklist
- [ ] Environment set to production
- [ ] All required environment variables configured
- [ ] JWT secret is sufficiently long (64+ chars)
- [ ] Database SSL enabled
- [ ] Real AI provider configured (not mock)
- [ ] Docker images built with proper permissions
- [ ] No debug endpoints exposed
- [ ] Error messages don't leak sensitive data

## Incident Response

If a security issue is discovered:
1. Check if running in production mode
2. Verify all security validations are passing
3. Review logs for any sensitive data exposure
4. Update configuration as needed
5. Redeploy with proper settings

## Future Improvements

- [ ] Add rate limiting by IP in production
- [ ] Implement API key rotation
- [ ] Add security audit logging
- [ ] Implement Content Security Policy reporting
- [ ] Add automated security testing in CI/CD