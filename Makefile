.PHONY: run install build wire clean proto

install:
	go mod download
	go mod tidy

wire:
	cd cmd/warrior && wire
	cd cmd/battle && wire
	cd cmd/battlespell && wire

# Generate protobuf code
proto:
	@echo "🔨 Generating protobuf code..."
	@mkdir -p api/proto/coin api/proto/weapon api/proto/warrior api/proto/battle api/proto/battlespell
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/coin/coin.proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/weapon/weapon.proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/warrior/warrior.proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/battle/battle.proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/battlespell/battlespell.proto
	@echo "✅ Protobuf code generated!"

build:
	go build -o bin/warrior cmd/warrior/main.go
	go build -o bin/weapon cmd/weapon/main.go
	go build -o bin/coin cmd/coin/main.go
	go build -o bin/enemy cmd/enemy/main.go
	go build -o bin/dragon cmd/dragon/main.go
	go build -o bin/battle cmd/battle/main.go
	go build -o bin/battlespell cmd/battlespell/main.go

run:
	go run cmd/warrior/main.go

clean:
	rm -rf bin/
	rm -f api/proto/**/*.pb.go
	@echo "🧹 Cleaned binaries and generated protobuf files"