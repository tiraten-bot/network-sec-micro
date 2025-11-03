#!/bin/bash

# Build script for heal service

set -e

echo "🔨 Building heal service..."

# Generate protobuf code first
echo "📦 Generating protobuf code..."
make proto

# Build the application
echo "🔨 Building application..."
go build -o bin/heal cmd/heal/main.go

echo "✅ Build completed successfully!"
echo "📦 Binary location: ./bin/heal"

