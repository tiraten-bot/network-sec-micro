#!/bin/bash
set -e
echo "✨ Building arenaspell service..."
go build -o bin/arenaspell cmd/arenaspell/main.go
echo "✅ Built bin/arenaspell"


