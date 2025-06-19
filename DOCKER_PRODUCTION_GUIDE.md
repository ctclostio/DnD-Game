# Docker Production Deployment Guide

## Overview
This guide ensures secure production deployments by using the correct Docker build stages and configurations.

## Security Requirements

### 1. Use Correct Docker Stages

#### Frontend Production Build
```bash
# CORRECT: Build the production stage
docker build -t dnd-frontend:prod --target production -f frontend/Dockerfile.optimized .

# WRONG: Never use development stage for production
# docker build -t dnd-frontend:prod --target development -f frontend/Dockerfile.optimized .
```

#### Backend Production Build
```bash
# CORRECT: Build the final stage
docker build -t dnd-backend:prod --target final -f backend/Dockerfile.optimized .

# WRONG: Never use development stage for production
# docker build -t dnd-backend:prod --target development -f backend/Dockerfile.optimized .
```

### 2. Production Stage Characteristics

#### Frontend Production Stage
- **Stage Name**: `production` (Stage 4 in Dockerfile.optimized)
- **Base Image**: `nginx:alpine-slim`
- **Features**:
  - No source code, only compiled build artifacts
  - No development dependencies
  - No source maps (disabled in webpack.prod.js)
  - Read-only file permissions
  - Non-root user execution
  - No NODE_ENV=development

#### Backend Production Stage
- **Stage Name**: `final` (Stage 4 in Dockerfile.optimized)
- **Base Image**: `scratch` (minimal)
- **Features**:
  - Only compiled binary, no source code
  - No development tools (no dlv debugger, no air hot reload)
  - No shell access (using scratch image)
  - Non-root user execution

### 3. Environment Variables for Production

#### Required Production Settings
```bash
# Backend
ENV=production
JWT_SECRET=<minimum-64-character-secret>
DB_SSLMODE=require
AI_PROVIDER=openai  # Not "mock"

# Frontend (set at runtime via runtime-env.sh)
REACT_APP_API_URL=https://api.yourdomain.com
REACT_APP_ENVIRONMENT=production
```

## Docker Compose for Production

### Production docker-compose.yml
```yaml
version: '3.8'

services:
  frontend:
    image: dnd-frontend:prod
    build:
      context: .
      dockerfile: frontend/Dockerfile.optimized
      target: production  # IMPORTANT: Specify production stage
    environment:
      - REACT_APP_ENVIRONMENT=production
      - REACT_APP_API_URL=https://api.yourdomain.com
    ports:
      - "80:3000"
    restart: unless-stopped

  backend:
    image: dnd-backend:prod
    build:
      context: .
      dockerfile: backend/Dockerfile.optimized
      target: final  # IMPORTANT: Specify final stage
    environment:
      - ENV=production
      - JWT_SECRET=${JWT_SECRET}
      - DB_HOST=${DB_HOST}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_SSLMODE=require
    ports:
      - "8080:8080"
    restart: unless-stopped
```

## Security Checklist

### Pre-Deployment Verification
- [ ] Frontend built with `--target production`
- [ ] Backend built with `--target final`
- [ ] Source maps disabled in webpack.prod.js
- [ ] NODE_ENV not set to development
- [ ] ENV=production for backend
- [ ] No debug tools in production images
- [ ] No source code in production images
- [ ] All secrets provided via environment variables

### Build Verification Commands
```bash
# Verify frontend doesn't contain source maps
docker run --rm dnd-frontend:prod ls /usr/share/nginx/html/js/ | grep -v ".map$"

# Verify backend doesn't contain source code
docker run --rm --entrypoint sh dnd-backend:prod -c "ls -la" 2>&1 | grep -q "exec: sh: not found" && echo "✅ No shell in production image"

# Check image sizes (production should be smaller)
docker images | grep dnd-
```

## Common Mistakes to Avoid

### 1. Wrong Target Stage
```bash
# ❌ WRONG - Uses development stage
docker build -t myapp:prod .

# ✅ CORRECT - Explicitly targets production stage
docker build -t myapp:prod --target production .
```

### 2. Development Environment Variables
```yaml
# ❌ WRONG - Development settings in production
environment:
  - NODE_ENV=development
  - DEBUG=true

# ✅ CORRECT - Production settings
environment:
  - NODE_ENV=production
  # No DEBUG variable
```

### 3. Including Source Maps
```javascript
// ❌ WRONG - webpack.prod.js with source maps
module.exports = {
  devtool: 'source-map',
  // ...
}

// ✅ CORRECT - No source maps in production
module.exports = {
  // No devtool property
  // ...
}
```

## CI/CD Pipeline Configuration

### GitHub Actions Example
```yaml
- name: Build Frontend Production
  run: |
    docker build \
      --target production \
      -t dnd-frontend:${{ github.sha }} \
      -f frontend/Dockerfile.optimized \
      .

- name: Build Backend Production
  run: |
    docker build \
      --target final \
      -t dnd-backend:${{ github.sha }} \
      -f backend/Dockerfile.optimized \
      .
```

## Monitoring and Alerts

### Security Indicators
1. **Image Size**: Production images should be significantly smaller
2. **Running Processes**: Only nginx/server process, no development tools
3. **Environment Variables**: Verify NODE_ENV is not "development"
4. **Network Traffic**: No source map requests in browser DevTools

### Container Security Scan
```bash
# Scan for vulnerabilities
docker scan dnd-frontend:prod
docker scan dnd-backend:prod

# Check for exposed secrets
docker history dnd-frontend:prod --no-trunc | grep -i secret
docker history dnd-backend:prod --no-trunc | grep -i secret
```

## Troubleshooting

### Issue: Application runs in development mode
**Solution**: Ensure you're using the correct Docker stage and ENV variables

### Issue: Source maps visible in browser
**Solution**: Rebuild with updated webpack.prod.js (no devtool property)

### Issue: Debug endpoints accessible
**Solution**: Verify backend ENV=production and rebuild with correct stage

## References
- [Docker Multi-Stage Builds](https://docs.docker.com/develop/develop-images/multistage-build/)
- [OWASP Docker Security](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)
- [Node.js Production Best Practices](https://github.com/goldbergyoni/nodebestpractices)