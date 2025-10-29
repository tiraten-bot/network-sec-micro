.PHONY: run install build wire clean proto

install:
	go mod download
	go mod tidy

wire:
	cd cmd/warrior && wire

# Generate protobuf code
proto:
	@echo "🔨 Generating protobuf code..."
	@mkdir -p api/proto/coin api/proto/weapon api/proto/warrior
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/coin/coin.proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/weapon/weapon.proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/warrior/warrior.proto
	@echo "✅ Protobuf code generated!"

build:
	go build -o bin/warrior cmd/warrior/main.go
	go build -o bin/weapon cmd/weapon/main.go
	go build -o bin/coin cmd/coin/main.go
	go build -o bin/enemy cmd/enemy/main.go

run:
	go run cmd/warrior/main.go

clean:
	rm -rf bin/
	rm -f api/proto/**/*.pb.go
	@echo "🧹 Cleaned binaries and generated protobuf files"