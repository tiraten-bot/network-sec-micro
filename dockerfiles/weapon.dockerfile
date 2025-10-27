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
RUN cd cmd/weapon && go build -o /app/bin/weapon .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1000 weapon && \
    adduser -D -u 1000 -G weapon weapon

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/weapon .

# Change ownership to weapon user
RUN chown weapon:weapon /app/weapon

# Switch to non-root user
USER weapon

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run the application
CMD ["./weapon"]

