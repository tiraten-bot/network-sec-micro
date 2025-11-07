#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage: ${0} <environment> [extra terraform args]" >&2
  exit 1
fi

ENVIRONMENT="$1"
shift || true

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STACK_DIR="${ROOT_DIR}/environments/${ENVIRONMENT}"

if [[ ! -d "${STACK_DIR}" ]]; then
  echo "Unknown environment: ${ENVIRONMENT}" >&2
  exit 1
fi

cd "${STACK_DIR}"
terraform init
terraform apply "$@"

