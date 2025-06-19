# Docker Security Improvements

## Overview
This document explains the security improvements made to our Docker configurations to address SonarCloud security hotspots.

## Security Issues Addressed

### 1. Recursive Copying (`COPY . .`)
**Issue**: Copying entire directories recursively can inadvertently include sensitive files like:
- Environment files (.env)
- Private keys
- Git history
- Local development configurations

**Solution**: 
- Replace `COPY . .` with specific COPY commands for required directories only
- Use `.dockerignore` files to exclude sensitive files as an additional layer of protection
- Copy only what's needed: `COPY backend/ ./backend/`, `COPY src/ ./src/`, etc.

### 2. Running as Root User
**Issue**: Running containers as root user poses security risks if the container is compromised.

**Solution**:
- Create non-root users in all Docker stages
- Switch to non-root user before executing commands
- Set proper file permissions and ownership

### 3. Glob Pattern Copying
**Issue**: Using glob patterns like `COPY *.json ./` can accidentally include unintended files.

**Solution**:
- Be explicit about which files to copy
- Use `.dockerignore` to exclude patterns
- Set ownership during COPY with `--chown` flag

## Best Practices Implemented

### 1. Multi-stage Builds
- Separate build dependencies from runtime
- Minimize final image size and attack surface
- Use scratch or minimal base images for production

### 2. Non-root User Execution
```dockerfile
# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Switch to non-root user
USER appuser
```

### 3. Explicit File Copying
```dockerfile
# Instead of: COPY . .
# Use:
COPY --chown=appuser:appgroup backend/ ./backend/
COPY --chown=appuser:appgroup data/ ./data/
```

### 4. Read-only File Systems
```dockerfile
# Make static files read-only
RUN chmod -R 444 /usr/share/nginx/html
```

### 5. Comprehensive .dockerignore
Both root and frontend directories now have `.dockerignore` files that exclude:
- Environment files
- Git directories
- IDE configurations
- Test files
- Build artifacts
- Sensitive credentials

## Security Validation

1. **Build Time**: 
   - No sensitive files are included in the Docker context
   - All operations run as non-root where possible

2. **Runtime**:
   - Containers run as non-root users
   - File permissions are restrictive
   - Only necessary files are included

3. **Image Scanning**:
   - Trivy scanner integrated in build pipeline
   - Regular security updates for base images

## Maintenance

- Regularly update base images to get security patches
- Review `.dockerignore` files when adding new sensitive files
- Audit file permissions and user access periodically
- Keep dependencies up to date