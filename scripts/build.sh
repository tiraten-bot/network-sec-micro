#!/bin/bash

# Build script for all services

set -e

echo "🔨 Building all services..."

# Generate protobuf code first
echo "📦 Generating protobuf code..."
make proto

# Build warrior service (manual DI due to Wire dependency issues)
echo "📦 Building warrior service..."
go build -o bin/warrior cmd/warrior/main.go

# Build weapon service
echo "🔨 Building weapon service..."
go build -o bin/weapon cmd/weapon/main.go

# Build coin service
echo "💰 Building coin service..."
go build -o bin/coin cmd/coin/main.go

# Build enemy service
echo "👹 Building enemy service..."
go build -o bin/enemy cmd/enemy/main.go

# Build dragon service
echo "🐉 Building dragon service..."
go build -o bin/dragon cmd/dragon/main.go

# Build battle service
echo "⚔️ Building battle service..."
go build -o bin/battle cmd/battle/main.go

# Build battlespell service
echo "✨ Building battlespell service..."
go build -o bin/battlespell cmd/battlespell/main.go

# Build arenaspell service
echo "✨ Building arenaspell service..."
go build -o bin/arenaspell cmd/arenaspell/main.go

# Build arena service
echo "🏟️ Building arena service..."
go build -o bin/arena cmd/arena/main.go

# Build fiber-gateway
echo "🌐 Building fiber-gateway..."
go build -o bin/fiber-gateway ./fiber-gateway

echo "✅ All services built successfully!"
echo "📦 Binaries location: ./bin/"
echo "   - warrior (HTTP API)"
echo "   - weapon (HTTP API)" 
echo "   - coin (gRPC API)"
echo "   - enemy (HTTP API)"
echo "   - dragon (HTTP API)"
echo "   - battle (HTTP API)"
echo "   - battlespell (HTTP API + gRPC)"
echo "   - arena (HTTP API + gRPC)"
echo "   - fiber-gateway (API Gateway)"
