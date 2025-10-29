#!/bin/bash

# Build script for all services

set -e

echo "🔨 Building all services..."

# Generate protobuf code first
echo "📦 Generating protobuf code..."
make proto

# Build weapon service
echo "🔨 Building weapon service..."
go build -o bin/weapon cmd/weapon/main.go

# Build coin service
echo "💰 Building coin service..."
go build -o bin/coin cmd/coin/main.go

# Build enemy service
echo "👹 Building enemy service..."
go build -o bin/enemy cmd/enemy/main.go

echo "✅ Services built successfully!"
echo "📦 Binaries location: ./bin/"
echo "   - weapon (HTTP API)" 
echo "   - coin (gRPC API)"
echo "   - enemy (HTTP API)"
echo ""
echo "⚠️  Warrior service skipped due to Wire dependency issues"
echo "   To build warrior: cd cmd/warrior && go build -o ../../bin/warrior main.go"
