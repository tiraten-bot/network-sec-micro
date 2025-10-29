# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN cd fiber-gateway && go build -o /app/bin/fiber-gateway .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1000 fibergw && \
    adduser -D -u 1000 -G fibergw fibergw

WORKDIR /app

COPY --from=builder /app/bin/fiber-gateway .

RUN chown fibergw:fibergw /app/fiber-gateway
USER fibergw

EXPOSE 8090

HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

ENV GW_PORT=8090

CMD ["./fiber-gateway"]


