#!/bin/bash

# Build script for enemy service

set -e

echo "👹 Building enemy service..."

# Generate protobuf code first
echo "📦 Generating protobuf code..."
make proto

# Build the application
echo "🔨 Building application..."
go build -o bin/enemy cmd/enemy/main.go

echo "✅ Build completed successfully!"
echo "📦 Binary location: ./bin/enemy"
