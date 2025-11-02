#!/bin/bash

set -e

echo "🔨 Building fiber-gateway..."

GO111MODULE=on go build -o bin/fiber-gateway fiber-gateway/main.go

echo "✅ Built: bin/fiber-gateway"


