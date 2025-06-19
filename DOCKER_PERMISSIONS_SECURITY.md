# Docker File Permissions Security Guide

## Overview
This document explains the security improvements made to address SonarCloud's "write permissions" security hotspots in Docker containers.

## Security Principle
**Never give non-root users write access to application code or configuration files in production containers.**

## Issues Addressed

### 1. Using --chown During COPY
**Problem**: Commands like `COPY --chown=user:group file ./` give the non-root user ownership with write permissions by default.

**Solution**: 
- Use `--chmod` to set read-only permissions during copy
- Change ownership separately after setting restrictive permissions
- Only grant write permissions where absolutely necessary (e.g., runtime directories)

### 2. Improper Permission Patterns

#### ❌ Insecure Pattern:
```dockerfile
# This gives write permissions to non-root user
COPY --chown=appuser:appgroup ./src ./src
```

#### ✅ Secure Pattern:
```dockerfile
# Copy as read-only first
COPY --chmod=444 ./src ./src
# Then change ownership if needed
RUN chown -R appuser:appgroup ./src
```

## Permission Guidelines by File Type

### Application Code (src, public, etc.)
- **Permissions**: 444 (read-only for all)
- **Rationale**: Source code should never be modified at runtime

### Configuration Files (*.json, *.conf)
- **Permissions**: 444 or 644 (read-only or read/write for owner only)
- **Rationale**: Config files should not be modified in production

### Executables (binaries, scripts)
- **Permissions**: 550 or 755 (execute for owner/group)
- **Rationale**: Need execute permission but not write

### Runtime Directories (/var/run, /var/cache)
- **Permissions**: 755 (write access needed)
- **Rationale**: Application needs to write logs, PIDs, cache

## Stage-Specific Considerations

### Build Stage
- Write permissions acceptable as build artifacts are temporary
- Should still run as non-root user when possible

### Development Stage
- Write permissions acceptable for hot reloading
- Security is less critical in development

### Test/Analysis Stages
- Read-only permissions preferred
- Only node_modules may need write access for some tools

### Production Stage
- **Strictest permissions required**
- All application files should be read-only
- Only runtime directories should have write access

## Implementation Examples

### Frontend Production:
```dockerfile
# Copy built files as read-only
COPY --from=builder --chmod=444 /app/build /usr/share/nginx/html

# Nginx configs as read-only
COPY --chmod=444 nginx.conf /etc/nginx/nginx.conf

# Runtime directories need write access
RUN chown appuser:appgroup /var/run/nginx.pid /var/cache/nginx
```

### Backend Production:
```dockerfile
# Binary with execute permissions only
COPY --from=builder --chmod=550 /app/server ./server

# Data files as read-only
COPY --from=builder --chmod=440 /app/data ./data

# Change ownership after setting permissions
RUN chown -R appuser:appgroup /app
```

## Security Benefits

1. **Defense in Depth**: Even if container is compromised, attacker cannot modify application code
2. **Prevents Tampering**: Malicious code injection is prevented
3. **Audit Trail**: File modifications would be immediately apparent
4. **Compliance**: Meets security standards for immutable infrastructure

## Validation

Use the provided script to validate permissions:
```bash
./scripts/validate-docker-security.sh
```

This checks for:
- No `COPY . .` commands
- Proper use of USER directive
- Correct permission patterns
- Comprehensive .dockerignore files