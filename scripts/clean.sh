#!/bin/bash

# Clean script

set -e

echo "🧹 Cleaning build artifacts..."

# Remove build directory
rm -rf bin/

# Remove go test cache
go clean -testcache

# Remove wire generated files
find . -name "wire_gen.go" -type f -delete

# Optional: Stop docker containers
read -p "Stop Docker containers? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker-compose down -v
fi

echo "✅ Clean completed!"
