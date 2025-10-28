#!/bin/bash

# Run script for coin service

set -e

echo "🚀 Starting coin service..."

# Set default environment variables if not set
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-postgres}
export DB_NAME=${DB_NAME:-warrior_db}
export DB_SSLMODE=${DB_SSLMODE:-disable}
export GRPC_PORT=${GRPC_PORT:-50051}
export KAFKA_BROKERS=${KAFKA_BROKERS:-localhost:9092}

echo "📊 Environment:"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_USER: $DB_USER"
echo "  DB_NAME: $DB_NAME"
echo "  GRPC_PORT: $GRPC_PORT"
echo "  KAFKA_BROKERS: $KAFKA_BROKERS"

# Run the application
go run cmd/coin/main.go

