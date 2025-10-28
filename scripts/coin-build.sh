#!/bin/bash

# Build script for coin service

set -e

echo "🔨 Building coin service..."

# Generate protobuf code first
echo "📦 Generating protobuf code..."
make proto

# Build the application
echo "🔨 Building application..."
go build -o bin/coin cmd/coin/main.go

echo "✅ Build completed successfully!"
echo "📦 Binary location: ./bin/coin"

