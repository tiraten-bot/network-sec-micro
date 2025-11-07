#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "${ROOT_DIR}"
find . -name ".terraform" -prune -o -name "*.tf" -print | while read -r file; do
  terraform fmt "${file}"
done

