#!/bin/bash

set -euo pipefail

OVERLAY=${1:-base}

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
KUSTOMIZE_PATH="$ROOT_DIR/k8s/$OVERLAY"

if [ ! -d "$KUSTOMIZE_PATH" ]; then
  echo "❌ Overlay not found: $KUSTOMIZE_PATH"
  echo "Usage: $0 [base|overlays/dev|overlays/staging|overlays/prod]"
  exit 1
fi

echo "🚀 Applying k8s manifests: $KUSTOMIZE_PATH"
kubectl apply -k "$KUSTOMIZE_PATH"

echo "✅ Applied. Current resources in namespace 'network-sec':"
kubectl get all -n network-sec


