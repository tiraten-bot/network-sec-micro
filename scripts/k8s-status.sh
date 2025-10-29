#!/bin/bash
set -euo pipefail

NS=${1:-network-sec}

echo "📊 Deployments:" && kubectl get deploy -n "$NS"
echo "\n📦 Pods:" && kubectl get pods -n "$NS"
echo "\n🔌 Services:" && kubectl get svc -n "$NS"
echo "\n🌐 Ingress:" && kubectl get ingress -n "$NS"


