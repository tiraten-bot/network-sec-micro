#!/bin/bash

# Run script for weapon service

set -e

echo "🚀 Starting weapon service..."

# Set default environment variables if not set
export MONGODB_URI=${MONGODB_URI:-mongodb://localhost:27017}
export MONGODB_DB=${MONGODB_DB:-weapon_db}
export PORT=${PORT:-8081}

echo "📊 Environment:"
echo "  MONGODB_URI: $MONGODB_URI"
echo "  MONGODB_DB: $MONGODB_DB"
echo "  PORT: $PORT"

# Run the application
go run cmd/weapon/main.go
