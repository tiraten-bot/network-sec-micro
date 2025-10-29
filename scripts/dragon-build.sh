#!/bin/bash

# Build script for dragon service

set -e

echo "🐉 Building dragon service..."

# Generate protobuf code first
echo "📦 Generating protobuf code..."
make proto

# Build the application
echo "🔨 Building application..."
go build -o bin/dragon cmd/dragon/main.go

echo "✅ Build completed successfully!"
echo "📦 Binary location: ./bin/dragon"
