#!/bin/bash

# Build script for warrior service

set -e

echo "🔨 Building warrior service..."

# Run wire generation
echo "📦 Running Wire..."
cd cmd/warrior && wire && cd ../..

# Build the application
echo "🔨 Building application..."
go build -o bin/warrior cmd/warrior/main.go

echo "✅ Build completed successfully!"
echo "📦 Binary location: ./bin/warrior"

