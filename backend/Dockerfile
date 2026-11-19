# Backend Dockerfile - Multi-stage build

# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/api

# Development stage
FROM golang:1.25-alpine AS development

RUN apk add --no-cache git make gcc musl-dev air

WORKDIR /app

# Install air for hot reload
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

# Production stage
FROM alpine:3.19 AS production

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -u 1000 -s /bin/sh appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /server /app/server

# Copy configs
COPY --from=builder /app/migrations ./migrations

# Change ownership
RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/ready || exit 1

CMD ["/app/server"]