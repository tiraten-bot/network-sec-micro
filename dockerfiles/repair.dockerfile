FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/repair ./cmd/repair/main.go

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/repair /repair
ENV PORT=50061
EXPOSE 50061
ENTRYPOINT ["/repair"]


