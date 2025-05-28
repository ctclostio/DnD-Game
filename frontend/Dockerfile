# Build stage
FROM node:18-alpine AS builder

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
FROM node:18-alpine AS development

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

# Copy built files from builder
COPY --from=builder /app/build /usr/share/nginx/html

# Copy nginx configuration if needed
# COPY nginx.conf /etc/nginx/nginx.conf

# Expose port
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]