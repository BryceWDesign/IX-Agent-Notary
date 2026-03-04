#!/usr/bin/env bash
set -euo pipefail

echo "== gofmt check =="
unformatted="$(gofmt -l .)"
if [[ -n "${unformatted}" ]]; then
  echo "gofmt required on:"
  echo "${unformatted}"
  exit 1
fi

echo "== go vet =="
go vet ./...

echo "== go test =="
go test ./...

echo "== generate demo assets (keys + receipts) =="
bash scripts/gen_demo_assets.sh

echo "== verify generated examples directory (strict; chain check is ok even if depth=0) =="
go run ./cmd/ix-an verify-dir examples/receipts --strict-approvals
