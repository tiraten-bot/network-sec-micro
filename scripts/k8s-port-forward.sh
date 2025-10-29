#!/bin/bash
set -euo pipefail

NS=${1:-network-sec}
SVC=${2:-fiber-gateway}
LOCAL=${3:-8090}
REMOTE=${4:-8090}

echo "🔁 Port-forward $SVC $LOCAL->$REMOTE (ns=$NS)"
kubectl port-forward -n "$NS" svc/$SVC "$LOCAL:$REMOTE"


