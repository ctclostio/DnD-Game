# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY backend/ ./backend/
COPY data/ ./data/

# Build the application
WORKDIR /app/backend/cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create app directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/backend/cmd/server/server .
# Copy data directory with proper ownership
COPY --from=builder --chown=appuser:appgroup /app/data ./data

# Change ownership of the binary
RUN chown appuser:appgroup server

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./server"]