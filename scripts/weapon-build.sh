#!/bin/bash

# Build script for weapon service

set -e

echo "🔨 Building weapon service..."

# Build the application
echo "🔨 Building application..."
go build -o bin/weapon cmd/weapon/main.go

echo "✅ Build completed successfully!"
echo "📦 Binary location: ./bin/weapon"

