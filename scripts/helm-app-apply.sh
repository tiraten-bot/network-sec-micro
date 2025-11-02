#!/bin/bash
set -euo pipefail

NS=${NS:-network-sec}
RELEASE=${RELEASE:-network-sec}
ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)

kubectl create ns "$NS" >/dev/null 2>&1 || true

helm upgrade --install "$RELEASE" "$ROOT_DIR/k8s/helm/network-sec" \
  --namespace "$NS" \
  --set namespace="$NS"

echo "✅ Helm app deployed: $RELEASE in ns=$NS"


