# Build stage
FROM node:24-alpine AS builder

# Create non-root user for build
RUN addgroup -g 1001 -S buildgroup && \
    adduser -u 1001 -S builduser -G buildgroup

WORKDIR /app

# Copy package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy only necessary source code
# Excludes files listed in .dockerignore
COPY --chown=builduser:buildgroup frontend/public ./public
COPY --chown=builduser:buildgroup frontend/src ./src
COPY --chown=builduser:buildgroup frontend/tsconfig.json ./
COPY --chown=builduser:buildgroup frontend/.eslintrc* ./

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
COPY frontend/package*.json ./

# Install all dependencies (including devDependencies)
RUN npm install

# Copy only necessary source code with proper ownership
COPY --chown=devuser:devgroup frontend/public ./public
COPY --chown=devuser:devgroup frontend/src ./src
COPY --chown=devuser:devgroup frontend/tsconfig.json ./
COPY --chown=devuser:devgroup frontend/.eslintrc* ./

# Change ownership of node_modules
RUN chown -R devuser:devgroup node_modules

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