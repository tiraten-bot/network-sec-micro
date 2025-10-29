#!/bin/bash
set -euo pipefail

NS=${NS:-network-sec}
RELEASE=${RELEASE:-network-sec}

helm uninstall "$RELEASE" -n "$NS" || true
echo "✅ Helm app removed: $RELEASE in ns=$NS"


