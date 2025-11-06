#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "==> Running security scans from ${ROOT_DIR}"

run_or_skip() {
    local tool="$1"
    local description="$2"
    shift 2

    if ! command -v "$tool" >/dev/null 2>&1; then
        echo "[-] Skipping ${description}: ${tool} not installed"
        return 0
    fi

    echo "[+] ${description}"
    "$tool" "$@"
}

cd "${ROOT_DIR}"

# Go source vulnerability scanning
run_or_skip "govulncheck" "Go vulnerability check" ./...

# Go static analysis for common security issues
run_or_skip "gosec" "GoSec static analysis" ./...

# Container / filesystem scanning (requires Trivy)
if command -v trivy >/dev/null 2>&1; then
    echo "[+] Trivy filesystem scan (vulnerabilities & misconfigurations)"
    trivy fs --quiet --exit-code 1 --scanners vuln,config --security-checks vuln,config .
else
    echo "[-] Skipping Trivy scan: trivy not installed"
fi

echo "==> Security scans completed"

