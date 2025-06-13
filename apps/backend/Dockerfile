# Multi-stage build for Go backend services
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build all backend services
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notification-service ./cmd/notification-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o seed-db ./cmd/seed-db
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o seed-user ./cmd/seed-user
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o db-tools ./cmd/db-tools
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o test-auth-server ./cmd/test-auth-server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# Copy the binaries from builder stage
COPY --from=builder /app/notification-service .
COPY --from=builder /app/seed-db .
COPY --from=builder /app/seed-user .
COPY --from=builder /app/db-tools .
COPY --from=builder /app/test-auth-server .

# Add non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership of application files
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

# Default command - can be overridden
CMD ["./notification-service"] 