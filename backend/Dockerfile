# Build stage
FROM golang:1.21-alpine AS builder

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
FROM alpine:3.19

# Install ca-certificates and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create app directory
WORKDIR /app

# Copy the binary from builder with execute permissions
COPY --from=builder --chmod=550 /app/backend/cmd/server/server .
# Copy data directory as read-only
COPY --from=builder --chmod=440 /app/data ./data

# Change ownership after setting permissions
RUN chown appuser:appgroup server && \
    chown -R appuser:appgroup ./data

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./server"]