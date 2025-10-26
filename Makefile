.PHONY: run install migrate

install:
	go mod download
	go mod tidy

run:
	go run cmd/warrior/main.go

build:
	go build -o bin/warrior cmd/warrior/main.go
