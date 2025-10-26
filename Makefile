.PHONY: run install build wire clean

install:
	go mod download
	go mod tidy

wire:
	cd cmd/warrior && wire

build:
	go build -o bin/warrior cmd/warrior/main.go

run:
	go run cmd/warrior/main.go

clean:
	rm -rf bin/