#!/bin/bash

# Run script for heal service

set -e

echo "🚀 Starting heal service..."

# Set default environment variables if not set
export HEAL_USE_POSTGRES=${HEAL_USE_POSTGRES:-1}
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-postgres}
export DB_NAME_HEAL=${DB_NAME_HEAL:-heal_db}
export DB_SSLMODE=${DB_SSLMODE:-disable}
export WARRIOR_GRPC_ADDR=${WARRIOR_GRPC_ADDR:-localhost:50052}
export COIN_GRPC_ADDR=${COIN_GRPC_ADDR:-localhost:50051}
export KAFKA_BROKERS=${KAFKA_BROKERS:-localhost:9092}
export REDIS_ADDR=${REDIS_ADDR:-localhost:6379}
export GRPC_PORT=${GRPC_PORT:-50058}

echo "📊 Environment:"
echo "  HEAL_USE_POSTGRES: $HEAL_USE_POSTGRES"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_USER: $DB_USER"
echo "  DB_NAME_HEAL: $DB_NAME_HEAL"
echo "  WARRIOR_GRPC_ADDR: $WARRIOR_GRPC_ADDR"
echo "  COIN_GRPC_ADDR: $COIN_GRPC_ADDR"
echo "  KAFKA_BROKERS: $KAFKA_BROKERS"
echo "  REDIS_ADDR: $REDIS_ADDR"
echo "  GRPC_PORT: $GRPC_PORT"

# Run the application
go run cmd/heal/main.go

