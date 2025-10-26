#!/bin/bash

# Run script for warrior service

set -e

echo "🚀 Starting warrior service..."

# Set default environment variables if not set
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-postgres}
export DB_NAME=${DB_NAME:-warrior_db}
export PORT=${PORT:-8080}

echo "📊 Environment:"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_USER: $DB_USER"
echo "  DB_NAME: $DB_NAME"
echo "  PORT: $PORT"

# Run the application
go run cmd/warrior/main.go
