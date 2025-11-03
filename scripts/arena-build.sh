#!/bin/bash

set -e

echo "🔨 Building Arena service..."

# Generate Wire code if needed
cd cmd/arena
if [ -f "wire.go" ]; then
    echo "Generating Wire code..."
    wire || true
fi
cd ../..

# Build the arena service
echo "Building arena binary..."
go build -o bin/arena cmd/arena/main.go

echo "✅ Arena service built successfully!"

