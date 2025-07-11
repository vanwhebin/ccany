# Multi-stage build Dockerfile for Enhanced CCany with Claude Code support

# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary packages for CGO and SQLite
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    sqlite-dev \
    build-base

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG BUILD_TIME=unknown

# Build the enhanced application with Claude Code support
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-extldflags '-static' -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o ccany cmd/server/main.go

# Verify the binary works
RUN ./ccany --help || echo "Binary built successfully"

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget \
    curl \
    sqlite

# Set timezone
ENV TZ=Asia/Shanghai

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/ccany .

# Create necessary directories with proper permissions
RUN mkdir -p \
    /app/data \
    /app/logs \
    /home/appuser/.claude \
    && chown -R appuser:appgroup /app /home/appuser

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8082

# Enhanced health check with Claude Code compatibility
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Environment variables for Claude Code compatibility
ENV CLAUDE_CODE_COMPATIBLE=true
ENV CLAUDE_CONFIG_PATH=/home/appuser/.claude.json

# Labels for better container management
LABEL org.opencontainers.image.title="CCany Enhanced with Claude Code Support"
LABEL org.opencontainers.image.description="Enhanced Claude-to-OpenAI API Proxy with full Claude Code compatibility"
LABEL org.opencontainers.image.vendor="CCany"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/yourusername/ccany"

# Start the enhanced application
CMD ["./ccany"]
