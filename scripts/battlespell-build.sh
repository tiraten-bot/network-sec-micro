#!/bin/bash

# BattleSpell Service Build Script

set -e

echo "Building BattleSpell service..."

cd "$(dirname "$0")/.."

# Generate Wire code (if wire_gen.go doesn't exist or wire.go is newer)
if [ ! -f "cmd/battlespell/wire_gen.go" ] || [ "cmd/battlespell/wire.go" -nt "cmd/battlespell/wire_gen.go" ]; then
    echo "Generating Wire code..."
    cd cmd/battlespell && wire || echo "Wire generation failed, using existing wire_gen.go"
    cd ../..
fi

# Build the binary
go build -o bin/battlespell ./cmd/battlespell

echo "BattleSpell service built successfully!"
echo "Binary location: ./bin/battlespell"

