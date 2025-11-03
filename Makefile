.PHONY: run install build wire clean proto

install:
	go mod download
	go mod tidy

wire:
	cd cmd/warrior && wire
	cd cmd/battle && wire
	cd cmd/battlespell && wire
	cd cmd/arena && wire
	cd cmd/arenaspell && wire

# Generate protobuf code
proto:
	@echo "🔨 Generating protobuf code..."
	@mkdir -p api/proto/coin api/proto/weapon api/proto/warrior api/proto/battle api/proto/battlespell api/proto/arena api/proto/heal
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/coin/coin.proto \
		api/proto/warrior/warrior.proto \
		api/proto/battle/battle.proto \
		api/proto/battlespell/battlespell.proto \
		api/proto/arena/arena.proto \
		api/proto/arenaspell/arenaspell.proto \
		api/proto/heal/heal.proto
	@echo "✅ Protobuf code generated!"

build:
	go build -o bin/warrior cmd/warrior/main.go
	go build -o bin/weapon cmd/weapon/main.go
	go build -o bin/coin cmd/coin/main.go
	go build -o bin/enemy cmd/enemy/main.go
	go build -o bin/dragon cmd/dragon/main.go
	go build -o bin/battle cmd/battle/main.go
	go build -o bin/battlespell cmd/battlespell/main.go
	go build -o bin/arena cmd/arena/main.go
	go build -o bin/heal cmd/heal/main.go

run:
	go run cmd/warrior/main.go

clean:
	rm -rf bin/
	rm -f api/proto/**/*.pb.go
	@echo "🧹 Cleaned binaries and generated protobuf files"