# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN cd cmd/heal && go build -o /app/bin/heal .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 heal && \
    adduser -D -u 1000 -G heal heal

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/heal .

# Change ownership to heal user
RUN chown heal:heal /app/heal

# Switch to non-root user
USER heal

# Expose port (gRPC)
EXPOSE 50058

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:50058/health || exit 1

# Run the application
CMD ["./heal"]

