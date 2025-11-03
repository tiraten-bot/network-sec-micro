#!/bin/bash
set -e

export ARENASPELL_HTTP_ADDR=${ARENASPELL_HTTP_ADDR:-:8088}
export ARENASPELL_GRPC_PORT=${ARENASPELL_GRPC_PORT:-50056}
export MONGODB_URI=${MONGODB_URI:-mongodb://localhost:27017}
export MONGODB_DB=${MONGODB_DB:-arenaspell_db}

./bin/arenaspell


