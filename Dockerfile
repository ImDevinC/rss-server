# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o rss-server ./cmd/server/

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S podcast && adduser -S podcast -G podcast

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/rss-server /app/rss-server

# Copy web assets
COPY --from=builder /build/web /app/web

# Copy config (if exists)COPY --from=builder /build/config.yaml* /app/ 2>/dev/null || true

# Create data directories with correct permissions
RUN mkdir -p /app/data/audio /app/data/artwork && \
    chown -R podcast:podcast /app

# Switch to non-root user
USER podcast

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/feed.xml || exit 1

# Set environment variables
ENV PORT=8080

# Run the application
CMD ["/app/rss-server"]
