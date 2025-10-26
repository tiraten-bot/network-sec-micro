#!/bin/bash

# Test script

set -e

echo "🧪 Running tests..."

# Run all tests
go test ./... -v -cover

echo "✅ Tests completed!"
