# Build stage
FROM golang:1.24.5-alpine AS builder

# Install build dependencies including CGO for SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=1 is required for SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o spacenet .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite

# Create non-root user
RUN addgroup -g 1001 -S spacenet && \
    adduser -u 1001 -S spacenet -G spacenet

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/spacenet .

# Create data directory for SQLite database
RUN mkdir -p /app/data && chown -R spacenet:spacenet /app

# Switch to non-root user
USER spacenet

# Expose the default HTTP port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
CMD ["./spacenet", "--http-port", "8080", "--database", "/app/data/spacenet.db"]