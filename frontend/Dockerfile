# Build stage
FROM node:24-alpine AS builder

WORKDIR /app

# Copy package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY frontend/ ./

# Build the application
RUN npm run build

# Development stage
FROM node:24-alpine AS development

WORKDIR /app

# Copy package files
COPY frontend/package*.json ./

# Install all dependencies (including devDependencies)
RUN npm install

# Copy source code
COPY frontend/ ./

# Expose port
EXPOSE 3000

# Start development server
CMD ["npm", "start"]

# Production stage
FROM nginx:alpine AS production

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy built files from builder with proper ownership
COPY --from=builder --chown=appuser:appgroup /app/build /usr/share/nginx/html

# Copy nginx configuration
COPY --chown=appuser:appgroup frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Update nginx config to run as non-root and set read-only permissions
RUN sed -i 's/user nginx;/user appuser;/g' /etc/nginx/nginx.conf && \
    touch /var/run/nginx.pid && \
    chown -R appuser:appgroup /var/run/nginx.pid /var/cache/nginx /var/log/nginx /etc/nginx/conf.d && \
    chmod -R 444 /usr/share/nginx/html && \
    chmod 444 /etc/nginx/conf.d/default.conf

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Start nginx
CMD ["nginx", "-g", "daemon off;"]