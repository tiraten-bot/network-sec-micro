#!/bin/bash

set -e

echo "🚀 Starting Arena service..."

# Set environment variables
export MONGODB_URI=${MONGODB_URI:-"mongodb://localhost:27017"}
export MONGODB_DB=${MONGODB_DB:-"arena_db"}
export WARRIOR_GRPC_ADDR=${WARRIOR_GRPC_ADDR:-"localhost:50052"}
export KAFKA_BROKERS=${KAFKA_BROKERS:-"localhost:9092"}
export GIN_MODE=${GIN_MODE:-"debug"}
export PORT=${PORT:-"8087"}
export GRPC_PORT=${GRPC_PORT:-"50055"}

# Run the service
exec ./bin/arena

