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
RUN cd cmd/coin && go build -o /app/bin/coin .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 coin && \
    adduser -D -u 1000 -G coin coin

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/coin .

# Change ownership to coin user
RUN chown coin:coin /app/coin

# Switch to non-root user
USER coin

# Expose port (gRPC)
EXPOSE 50051

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:50051/health || exit 1

# Run the application
CMD ["./coin"]

