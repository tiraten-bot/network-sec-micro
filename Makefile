.PHONY: run install build wire clean proto security-scan

install:
	go mod download
	go mod tidy

wire:
	cd cmd/warrior && wire
	cd cmd/battle && wire
	cd cmd/battlespell && wire
	cd cmd/arena && wire
	cd cmd/arenaspell && wire
	cd cmd/heal && wire || echo "Wire generation skipped for heal (dependency issue)"
	cd cmd/repair && wire || echo "Wire generation skipped for repair (optional)"

# Generate protobuf code
proto:
	@echo "🔨 Generating protobuf code..."
	@mkdir -p api/proto/coin api/proto/weapon api/proto/warrior api/proto/battle api/proto/battlespell api/proto/arena api/proto/heal api/proto/dragon api/proto/enemy api/proto/repair
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/coin/coin.proto \
		api/proto/weapon/weapon.proto \
		api/proto/armor/armor.proto \
		api/proto/warrior/warrior.proto \
		api/proto/battle/battle.proto \
		api/proto/battlespell/battlespell.proto \
		api/proto/arena/arena.proto \
		api/proto/arenaspell/arenaspell.proto \
		api/proto/heal/heal.proto \
		api/proto/dragon/dragon.proto \
		api/proto/enemy/enemy.proto \
		api/proto/repair/repair.proto
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
	go build -o bin/armor cmd/armor/main.go

run:
	go run cmd/warrior/main.go

clean:
	rm -rf bin/
	rm -f api/proto/**/*.pb.go
	@echo "🧹 Cleaned binaries and generated protobuf files"

security-scan:
	@./scripts/security-scan.sh