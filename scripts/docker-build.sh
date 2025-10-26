#!/bin/bash

# Docker build script

set -e

echo "🐳 Building warrior Docker image..."

docker build -f dockerfiles/warrior.dockerfile -t warrior:latest .

echo "✅ Docker image built successfully!"
echo "📦 Image: warrior:latest"
echo ""
echo "🚀 To run: docker run -p 8080:8080 warrior:latest"
