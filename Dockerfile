# Multi-stage build Dockerfile for CCany with pure Go SQLite
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget \
    curl

# Set timezone
ENV TZ=Asia/Shanghai

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from GoReleaser
COPY ccany ./ccany

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

# Health check - only check /health endpoint
HEALTHCHECK --interval=30s --timeout=10s --start-period=120s --retries=5 \
    CMD wget --no-verbose --tries=1 -O /dev/null http://localhost:8082/health || exit 1

# Environment variables for Claude Code compatibility
ENV CLAUDE_CODE_COMPATIBLE=true
ENV CLAUDE_CONFIG_PATH=/home/appuser/.claude.json

# Labels for better container management
LABEL org.opencontainers.image.title="CCany"
LABEL org.opencontainers.image.description="Claude-to-OpenAI API Proxy with Claude Code compatibility"
LABEL org.opencontainers.image.vendor="CCany"
LABEL org.opencontainers.image.licenses="MIT"

# Start the application
CMD ["./ccany"]
