#!/bin/bash

# Battle Service Build Script

set -e

echo "Building Battle service..."

cd "$(dirname "$0")/.."

# Build the binary
go build -o bin/battle ./cmd/battle

echo "Battle service built successfully!"
echo "Binary location: ./bin/battle"

