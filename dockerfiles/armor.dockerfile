FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN go mod download && go build -o /out/armor cmd/armor/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=build /out/armor /usr/local/bin/armor
ENV PORT=8089 GRPC_PORT=50059 MONGODB_URI=mongodb://mongodb:27017 MONGODB_DB=armor_db
EXPOSE 8089 50059
ENTRYPOINT ["/usr/local/bin/armor"]


