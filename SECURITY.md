# Security Configuration

This document outlines the security measures implemented in the D&D Game application.

## Security Headers

The application implements comprehensive security headers to protect against common web vulnerabilities:

### Content Security Policy (CSP)
- Restricts resource loading to trusted sources
- Prevents XSS attacks by blocking inline scripts (in production)
- Configured differently for development and production environments

### HTTP Strict Transport Security (HSTS)
- Forces HTTPS connections in production
- Prevents protocol downgrade attacks
- Enabled with `includeSubDomains` and `preload` directives

### Other Security Headers
- `X-Content-Type-Options: nosniff` - Prevents MIME type sniffing
- `X-Frame-Options: DENY` - Prevents clickjacking attacks
- `X-XSS-Protection: 1; mode=block` - Additional XSS protection
- `Referrer-Policy: strict-origin-when-cross-origin` - Controls referrer information
- `Permissions-Policy` - Restricts browser features

## WebSocket Security

### Origin Validation
- Strict origin checking for WebSocket connections
- Configurable allowed origins via environment variables
- Development mode allows localhost connections

### Authentication Flow
1. Client connects to WebSocket without token in URL
2. Server sends `auth_required` message
3. Client responds with authentication token
4. Server validates token and establishes authenticated connection
5. All subsequent messages are authenticated

### Token Security
- Tokens are never transmitted in URLs (prevents logging/caching issues)
- Tokens are sent via secure WebSocket messages after connection
- Automatic token refresh on expiration

## CORS Configuration

### Strict CORS Policy
- Whitelisted origins only
- Specific allowed headers
- Credentials support with proper validation
- Different configurations for development and production

### Configuration
```env
# Development
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080

# Production
PRODUCTION_ORIGIN=https://yourdomain.com
```

## Authentication & Authorization

### JWT Token Management
- Short-lived access tokens (15 minutes default)
- Long-lived refresh tokens (7 days default)
- Secure token storage in httpOnly cookies (production)
- Automatic token refresh mechanism

### CSRF Protection
- Token-based CSRF protection for state-changing operations
- Double-submit cookie pattern
- Validation on all POST/PUT/DELETE requests

## Rate Limiting

### Endpoint-Specific Limits
- Authentication endpoints: 5 requests/minute
- API endpoints: 100 requests/minute
- Configurable via environment variables

### Implementation
- In-memory rate limiting with exponential backoff
- Per-IP address tracking
- Graceful degradation under load

## Environment Variables

### Required Security Configuration
```env
# JWT Secret (minimum 32 characters)
JWT_SECRET=your-very-secure-jwt-secret-here-min-32-chars

# Token Durations
ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=7d

# Production Origin
PRODUCTION_ORIGIN=https://yourdomain.com

# Rate Limits
RATE_LIMIT_AUTH=5
RATE_LIMIT_API=100
```

### Production-Only Settings
```env
# Enable HSTS
ENABLE_HSTS=true

# CSP Reporting
CSP_REPORT_URI=https://yourdomain.com/csp-report

# Secure Cookies
SESSION_SECURE=true
SESSION_HTTPONLY=true
SESSION_SAMESITE=strict
```

## Best Practices

1. **Never commit sensitive data**
   - Use `.env` files for secrets
   - Keep `.env` out of version control
   - Use `.env.example` as a template

2. **Regular Security Updates**
   - Keep dependencies updated
   - Monitor security advisories
   - Run security audits regularly

3. **Logging and Monitoring**
   - Never log sensitive data (tokens, passwords)
   - Monitor failed authentication attempts
   - Set up alerts for suspicious activity

4. **Data Validation**
   - Validate all user input
   - Use parameterized queries
   - Sanitize output to prevent XSS

## Security Checklist for Deployment

- [ ] Generate strong JWT secret (32+ characters)
- [ ] Set production environment variables
- [ ] Configure HTTPS/TLS certificates
- [ ] Update CORS allowed origins
- [ ] Enable all security headers
- [ ] Configure rate limiting thresholds
- [ ] Set up monitoring and alerting
- [ ] Review and update dependencies
- [ ] Enable secure cookie settings
- [ ] Configure firewall rules
- [ ] Set up backup and recovery procedures
- [ ] Document incident response plan

## Reporting Security Issues

If you discover a security vulnerability, please email security@yourdomain.com with:

1. Description of the vulnerability
2. Steps to reproduce
3. Potential impact
4. Suggested fix (if any)

Please do not create public issues for security vulnerabilities.