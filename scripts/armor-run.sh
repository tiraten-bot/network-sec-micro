#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export PORT=${PORT:-8089}
export GRPC_PORT=${GRPC_PORT:-50059}
export MONGODB_URI=${MONGODB_URI:-mongodb://localhost:27017}
export MONGODB_DB=${MONGODB_DB:-armor_db}
go run cmd/armor/main.go


