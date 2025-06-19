# Build stage
FROM node:24-alpine AS builder

# Create non-root user for build
RUN addgroup -g 1001 -S buildgroup && \
    adduser -u 1001 -S builduser -G buildgroup

WORKDIR /app

# Copy package files
COPY frontend/package.json frontend/package-lock.json ./

# Install dependencies
RUN npm ci

# Copy only necessary source code as read-only
# Excludes files listed in .dockerignore
COPY --chmod=444 frontend/public ./public
COPY --chmod=444 frontend/src ./src
COPY --chmod=444 frontend/tsconfig.json ./
COPY --chmod=444 frontend/.eslintrc.js ./

# Set proper ownership and permissions
# Source files should be read-only, build directory needs write access
RUN chown -R builduser:buildgroup /app && \
    chmod -R 644 /app/public /app/src && \
    chmod 644 /app/tsconfig.json && \
    chmod -R 755 /app/node_modules

# Switch to non-root user for build
USER builduser

# Build the application
RUN npm run build

# Development stage
FROM node:24-alpine AS development

# Create non-root user
RUN addgroup -g 1001 -S devgroup && \
    adduser -u 1001 -S devuser -G devgroup

WORKDIR /app

# Copy package files
COPY frontend/package.json frontend/package-lock.json ./

# Install all dependencies (including devDependencies)
RUN npm install

# Copy only necessary source code with proper permissions
# Development needs write access for hot reloading
COPY --chmod=755 frontend/public ./public
COPY --chmod=755 frontend/src ./src
COPY --chmod=644 frontend/tsconfig.json ./
COPY --chmod=644 frontend/.eslintrc.js ./

# Set ownership for development (write access needed)
RUN chown -R devuser:devgroup /app

# Switch to non-root user
USER devuser

# Expose port
EXPOSE 3000

# Start development server
CMD ["npm", "start"]

# Production stage
FROM nginx:alpine AS production

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy built files from builder as read-only
COPY --from=builder --chmod=444 /app/build /usr/share/nginx/html

# Copy nginx configuration as read-only
COPY --chmod=444 frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Update nginx config to run as non-root
# Set proper ownership for runtime directories only
RUN sed -i 's/user nginx;/user appuser;/g' /etc/nginx/nginx.conf && \
    touch /var/run/nginx.pid && \
    chown -R appuser:appgroup /var/run/nginx.pid /var/cache/nginx /var/log/nginx && \
    chown -R appuser:appgroup /usr/share/nginx/html && \
    chown appuser:appgroup /etc/nginx/conf.d/default.conf

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Start nginx
CMD ["nginx", "-g", "daemon off;"]