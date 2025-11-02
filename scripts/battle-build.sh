#!/bin/bash

# Battle Service Build Script

set -e

echo "Building Battle service..."

cd "$(dirname "$0")/.."

# Generate Wire code (if wire_gen.go doesn't exist or wire.go is newer)
if [ ! -f "cmd/battle/wire_gen.go" ] || [ "cmd/battle/wire.go" -nt "cmd/battle/wire_gen.go" ]; then
    echo "Generating Wire code..."
    cd cmd/battle && wire || echo "Wire generation failed, using existing wire_gen.go"
    cd ../..
fi

# Build the binary
go build -o bin/battle ./cmd/battle

echo "Battle service built successfully!"
echo "Binary location: ./bin/battle"

