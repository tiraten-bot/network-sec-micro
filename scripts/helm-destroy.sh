#!/bin/bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR/k8s/helm"

if ! command -v helmfile >/dev/null 2>&1; then
  echo "❌ helmfile not found. Install: https://github.com/helmfile/helmfile"
  exit 1
fi

echo "🧹 Destroying Helm releases..."
helmfile destroy
echo "✅ Done"


