#!/bin/bash

# Docker build script for all services

set -e

echo "🐳 Building Docker images..."

# Build warrior service
echo "🐳 Building warrior service..."
docker build -f dockerfiles/warrior.dockerfile -t warrior:latest .

# Build weapon service
echo "🐳 Building weapon service..."
docker build -f dockerfiles/weapon.dockerfile -t weapon:latest .

# Build coin service
echo "🐳 Building coin service..."
docker build -f dockerfiles/coin.dockerfile -t coin:latest .

echo "✅ All Docker images built successfully!"
echo "📦 Images: warrior:latest, weapon:latest, coin:latest"
echo ""
echo "🚀 To run with Docker Compose: docker-compose up -d"
