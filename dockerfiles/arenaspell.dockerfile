# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate Wire code (best-effort)
RUN cd cmd/arenaspell && wire || true

# Build the arenaspell service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /arenaspell ./cmd/arenaspell

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy binary from builder
COPY --from=builder /arenaspell .

# Create non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8088/health || exit 1

EXPOSE 8088
EXPOSE 50056

CMD ["./arenaspell"]


