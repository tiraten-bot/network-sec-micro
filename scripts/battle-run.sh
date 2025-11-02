#!/bin/bash

# Battle Service Run Script

set -e

cd "$(dirname "$0")/.."

# Default values
export PORT=${PORT:-8085}
export MONGODB_URI=${MONGODB_URI:-mongodb://localhost:27017}
export MONGODB_DB=${MONGODB_DB:-battle_db}
export WARRIOR_GRPC_ADDR=${WARRIOR_GRPC_ADDR:-localhost:50052}
export COIN_GRPC_ADDR=${COIN_GRPC_ADDR:-localhost:50051}
export KAFKA_BROKERS=${KAFKA_BROKERS:-localhost:9092}
export GIN_MODE=${GIN_MODE:-debug}

echo "Starting Battle service on port $PORT..."
echo "MongoDB: $MONGODB_URI"
echo "Warrior gRPC: $WARRIOR_GRPC_ADDR"
echo "Coin gRPC: $COIN_GRPC_ADDR"
echo "Kafka: $KAFKA_BROKERS"

# Run the service
./bin/battle

