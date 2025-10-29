#!/bin/bash
set -euo pipefail
"$(dirname "$0")/k8s-apply.sh" overlays/dev


