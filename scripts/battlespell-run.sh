#!/bin/bash

# BattleSpell Service Run Script

set -e

echo "Running BattleSpell service..."

cd "$(dirname "$0")/.."

# Check if binary exists, if not build it
if [ ! -f "bin/battlespell" ]; then
    echo "Binary not found, building..."
    ./scripts/battlespell-build.sh
fi

# Set environment variables if not set
export MONGODB_URI=${MONGODB_URI:-mongodb://localhost:27017}
export MONGODB_DB=${MONGODB_DB:-battlespell_db}
export BATTLE_GRPC_ADDR=${BATTLE_GRPC_ADDR:-localhost:50053}
export PORT=${PORT:-8086}
export GRPC_PORT=${GRPC_PORT:-50054}
export GIN_MODE=${GIN_MODE:-debug}

echo "Starting BattleSpell service..."
echo "MongoDB URI: $MONGODB_URI"
echo "MongoDB DB: $MONGODB_DB"
echo "Battle gRPC: $BATTLE_GRPC_ADDR"
echo "HTTP Port: $PORT"
echo "gRPC Port: $GRPC_PORT"

./bin/battlespell

