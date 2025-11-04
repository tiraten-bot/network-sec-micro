#!/bin/bash

# Build script for repair service

set -e

echo "🔨 Building repair service..."

echo "🔧 Generating protos..."
make proto

echo "🔨 Building application..."
go build -o bin/repair cmd/repair/main.go

echo "✅ Build completed successfully!"
echo "📦 Binary location: ./bin/repair"


