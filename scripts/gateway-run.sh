#!/bin/bash

set -e

echo "🚀 Starting fiber-gateway..."

export GW_PORT=${GW_PORT:-8090}
export UPSTREAM_WARRIOR=${UPSTREAM_WARRIOR:-http://localhost:8080}
export UPSTREAM_ENEMY=${UPSTREAM_ENEMY:-http://localhost:8083}
export UPSTREAM_DRAGON=${UPSTREAM_DRAGON:-http://localhost:8084}
export UPSTREAM_WEAPON=${UPSTREAM_WEAPON:-http://localhost:8081}
export UPSTREAM_BATTLE=${UPSTREAM_BATTLE:-http://localhost:8085}

echo "📊 Env: GW_PORT=$GW_PORT"
echo "  WARRIOR=$UPSTREAM_WARRIOR"
echo "  ENEMY=$UPSTREAM_ENEMY"
echo "  DRAGON=$UPSTREAM_DRAGON"
echo "  WEAPON=$UPSTREAM_WEAPON"
echo "  BATTLE=$UPSTREAM_BATTLE"

go run fiber-gateway/main.go


