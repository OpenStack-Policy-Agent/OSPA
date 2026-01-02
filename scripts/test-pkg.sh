#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

RUN_RACE="${RUN_RACE:-0}"
RUN_INTEGRATION="${RUN_INTEGRATION:-0}"

echo "==> Go toolchain: $(go version)"

echo "==> Running pkg unit tests with coverage"
RACE_FLAG=()
if [[ "${RUN_RACE}" == "1" ]]; then
  RACE_FLAG+=("-race")
fi

go test ./pkg/... -count=1 "${RACE_FLAG[@]}" -coverprofile=pkg.cover.out
go tool cover -func=pkg.cover.out | tail -n 1

echo "==> Wrote coverage profile: pkg.cover.out"

if [[ "${RUN_INTEGRATION}" == "1" ]]; then
  echo "==> Running pkg integration tests (tag=integration)"
  go test -tags=integration ./pkg/... -count=1 -coverprofile=pkg.integration.cover.out
  go tool cover -func=pkg.integration.cover.out | tail -n 1
  echo "==> Wrote integration coverage profile: pkg.integration.cover.out"
fi
