#!/usr/bin/env bash
set -euo pipefail

export PATH="${PATH}:/usr/local/go/bin"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

echo "==> Running pkg unit tests with coverage"
go test ./pkg/... -count=1 -race -coverprofile=pkg.cover.out
go tool cover -func=pkg.cover.out | tail -n 1

echo "==> Wrote coverage profile: pkg.cover.out"

if [[ "${RUN_INTEGRATION:-}" == "1" ]]; then
  echo "==> Running pkg integration tests (tag=integration)"
  go test -tags=integration ./pkg/... -count=1 -coverprofile=pkg.integration.cover.out
  go tool cover -func=pkg.integration.cover.out | tail -n 1
  echo "==> Wrote integration coverage profile: pkg.integration.cover.out"
fi
