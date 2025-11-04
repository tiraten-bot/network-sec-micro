#!/bin/bash

# Run script for repair service

set -e

echo "🚀 Starting repair service..."

export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-postgres}
export DB_NAME_REPAIR=${DB_NAME_REPAIR:-repair_db}
export KAFKA_BROKERS=${KAFKA_BROKERS:-localhost:9092}
export WEAPON_GRPC_ADDR=${WEAPON_GRPC_ADDR:-localhost:50057}

echo "📊 Environment:"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_NAME_REPAIR: $DB_NAME_REPAIR"
echo "  KAFKA_BROKERS: $KAFKA_BROKERS"
echo "  WEAPON_GRPC_ADDR: $WEAPON_GRPC_ADDR"

go run cmd/repair/main.go


