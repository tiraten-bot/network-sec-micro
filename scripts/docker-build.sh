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

# Build enemy service
echo "🐳 Building enemy service..."
docker build -f dockerfiles/enemy.dockerfile -t enemy:latest .

echo "✅ All Docker images built successfully!"
echo "📦 Images: warrior:latest, weapon:latest, coin:latest, enemy:latest"
echo ""
echo "🚀 To run with Docker Compose: docker-compose up -d"
