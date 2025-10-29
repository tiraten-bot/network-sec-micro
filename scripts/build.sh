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

echo "✅ All services built successfully!"
echo "📦 Binaries location: ./bin/"
echo "   - warrior (HTTP API)"
echo "   - weapon (HTTP API)" 
echo "   - coin (gRPC API)"
echo "   - enemy (HTTP API)"
