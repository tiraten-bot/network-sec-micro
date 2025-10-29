#!/bin/bash

# Run script for dragon service

set -e

echo "🐉 Starting dragon service..."

# Set default environment variables if not set
export MONGODB_URI=${MONGODB_URI:-mongodb://localhost:27017}
export MONGODB_DATABASE=${MONGODB_DATABASE:-dragon_db}
export WARRIOR_GRPC_HOST=${WARRIOR_GRPC_HOST:-localhost:50052}
export KAFKA_BROKERS=${KAFKA_BROKERS:-localhost:9092}
export GIN_MODE=${GIN_MODE:-debug}
export PORT=${PORT:-8084}

echo "📊 Environment:"
echo "  MONGODB_URI: $MONGODB_URI"
echo "  MONGODB_DATABASE: $MONGODB_DATABASE"
echo "  WARRIOR_GRPC_HOST: $WARRIOR_GRPC_HOST"
echo "  KAFKA_BROKERS: $KAFKA_BROKERS"
echo "  PORT: $PORT"

# Run the application
go run cmd/dragon/main.go
