#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage: ${0} <environment>" >&2
  exit 1
fi

ENVIRONMENT="$1"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STACK_DIR="${ROOT_DIR}/environments/${ENVIRONMENT}"

if [[ ! -d "${STACK_DIR}" ]]; then
  echo "Unknown environment: ${ENVIRONMENT}" >&2
  exit 1
fi

cd "${STACK_DIR}"
terraform init >/dev/null
terraform plan -detailed-exitcode >/tmp/terraform-plan.out || EXIT=$?

if [[ "${EXIT:-0}" -eq 0 ]]; then
  echo "No drift detected."
elif [[ "${EXIT:-0}" -eq 2 ]]; then
  echo "Drift detected. See plan output:"
  cat /tmp/terraform-plan.out
else
  echo "Terraform plan failed." >&2
  exit "${EXIT:-1}"
fi

