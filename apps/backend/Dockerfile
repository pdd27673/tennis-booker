# Multi-stage build for Go backend services
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build all backend services with optimized flags
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o notification-service ./cmd/notification-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o retention-service ./cmd/retention-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o seed-db ./cmd/seed-db
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o seed-user ./cmd/seed-user
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o db-tools ./cmd/db-tools
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o test-auth-server ./cmd/test-auth-server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget curl

WORKDIR /app

# Copy all binaries from builder stage
COPY --from=builder /app/notification-service .
COPY --from=builder /app/retention-service .
COPY --from=builder /app/server .
COPY --from=builder /app/seed-db .
COPY --from=builder /app/seed-user .
COPY --from=builder /app/db-tools .
COPY --from=builder /app/test-auth-server .

# Create directories for logs and metrics
RUN mkdir -p /var/log /var/metrics

# Add non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership of application files and directories
RUN chown -R appuser:appgroup /app /var/log /var/metrics

USER appuser

EXPOSE 8080

# Health check for general service
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pgrep -f "notification-service|server|retention-service" > /dev/null || exit 1

# Default command - can be overridden via docker-compose or command line
CMD ["./notification-service"] 