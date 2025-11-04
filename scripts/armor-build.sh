#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
go build -o bin/armor cmd/armor/main.go
echo "armor built -> bin/armor"


