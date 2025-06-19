# Docker Glob Pattern Security Guide

## Overview
This document explains why avoiding glob patterns in Docker COPY commands is a security best practice.

## Security Risk
Glob patterns (wildcards like `*`, `?`, `[...]`) in COPY commands can inadvertently include sensitive files in Docker images.

## Examples of Risky Patterns

### ❌ Insecure Patterns:
```dockerfile
# Could match unintended files
COPY .eslintrc* ./              # Might match .eslintrc.backup, .eslintrc.old
COPY package*.json ./           # Might match package-secrets.json
COPY config.* ./                # Might match config.backup, config.secret
COPY *.conf ./                  # Might match sensitive.conf
```

### ✅ Secure Patterns:
```dockerfile
# Explicitly list files
COPY .eslintrc.js ./
COPY package.json package-lock.json ./
COPY config.yaml ./
COPY nginx.conf ./
```

## Why This Matters

1. **Unintended File Inclusion**: Glob patterns might match temporary files, backups, or sensitive files
   - `.env*` might include `.env.production` with secrets
   - `*.json` might include `secrets.json`
   - `config.*` might include `config.backup` with old passwords

2. **Build Context Pollution**: Development environments often contain:
   - Backup files (`.bak`, `.old`, `.backup`)
   - Editor temporary files (`.swp`, `~`)
   - Local configuration overrides
   - Test data files

3. **Supply Chain Security**: Explicit file listing ensures:
   - Predictable build outputs
   - Easier security audits
   - Clear understanding of what's in the image

## Best Practices

### 1. Be Explicit
Always specify exact filenames:
```dockerfile
# Instead of
COPY src/* ./src/

# Use
COPY src/index.js src/app.js src/utils.js ./src/
```

### 2. Use .dockerignore
Even with explicit COPY commands, maintain a comprehensive .dockerignore:
```
.env*
*.secret
*.key
*.pem
**/backup/
**/temp/
```

### 3. Handle Optional Files
For optional files, use conditional logic in build scripts rather than globs:
```dockerfile
# Instead of optional globs
COPY package-lock.json* ./

# Use explicit copy (file must exist)
COPY package-lock.json ./
```

### 4. Multi-Stage Builds
Use multi-stage builds to further limit what gets into final images:
```dockerfile
# Build stage can be less strict
FROM node:20 AS builder
COPY . .
RUN npm run build

# Production stage must be explicit
FROM node:20-alpine
COPY --from=builder /app/dist ./dist
COPY package.json package-lock.json ./
```

## Validation
Check for glob patterns in Dockerfiles:
```bash
# Find potential glob patterns
grep -n "COPY.*[*?\[]" Dockerfile

# Validate all Dockerfiles
find . -name "Dockerfile*" -exec grep -l "COPY.*[*?\[]" {} \;
```

## Migration Guide

When removing glob patterns:

1. **Identify actual files**:
   ```bash
   ls -la frontend/.eslintrc*
   ls -la frontend/package*.json
   ```

2. **Update COPY commands** to list files explicitly

3. **Test builds** to ensure all required files are included

4. **Update CI/CD** if it expects certain patterns

## Summary

Avoiding glob patterns in Docker COPY commands:
- Prevents accidental inclusion of sensitive files
- Makes builds more predictable and auditable
- Follows the principle of least privilege
- Improves supply chain security

Always prefer explicit file listing over convenience of wildcards.