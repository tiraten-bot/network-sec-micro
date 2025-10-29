#!/bin/bash

# Docker compose run script

set -e

echo "🐳 Starting services with Docker Compose..."

docker-compose up -d

echo "✅ Services started!"
echo ""
echo "📊 Status:"
docker-compose ps

echo ""
echo "📝 Logs: docker-compose logs -f fiber-gateway"
echo "🛑 Stop: docker-compose down"
