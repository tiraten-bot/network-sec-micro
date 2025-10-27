#!/bin/bash

# Build script for all services

set -e

echo "🔨 Building all services..."

# Build warrior service
echo "📦 Building warrior service..."
cd cmd/warrior && wire && cd ../..
go build -o bin/warrior cmd/warrior/main.go

# Build weapon service
echo "🔨 Building weapon service..."
go build -o bin/weapon cmd/weapon/main.go

echo "✅ All services built successfully!"
echo "📦 Binaries location: ./bin/"
