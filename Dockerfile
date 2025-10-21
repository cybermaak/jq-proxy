# Multi-stage Dockerfile for JQ Proxy Service

# Build stage
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates (needed for go modules and HTTPS)
RUN apk add --no-cache git ca-certificates tzdata

# Create appuser for security
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o proxy cmd/proxy/main.go

# Final stage - using alpine for health checks
FROM alpine:latest

# Install ca-certificates and wget for health checks
RUN apk --no-cache add ca-certificates wget

# Create appuser for security
RUN adduser -D -g '' appuser

# Copy the binary
COPY --from=builder /build/proxy /app/proxy

# Create configs directory for volume mounting
RUN mkdir -p /app/configs

# Use non-root user
USER appuser

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
ENTRYPOINT ["/app/proxy"]
CMD ["-config", "/app/configs/docker.json"]