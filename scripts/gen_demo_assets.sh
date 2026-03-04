#!/usr/bin/env bash
set -euo pipefail

# Generates local dev keys + example receipts that STRICTLY verify.
# Safe for public repos because outputs are gitignored.

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

mkdir -p keys/dev examples/receipts

seed="keys/dev/dev-key-001.seed"
pub="keys/dev/dev-key-001.pub"
kid="dev-key-001"

if [[ ! -f "$seed" || ! -f "$pub" ]]; then
  echo "== generating dev keypair =="
  go run ./cmd/ix-an keygen --out-seed "$seed" --out-pub "$pub"
else
  echo "== dev keypair exists =="
fi

echo "== generating example receipts (overwrite) =="
go run ./cmd/ix-an simulate --path docs/demo.txt       --out examples/receipts/minimal.receipt.json   --key "$seed" --key-id "$kid"
go run ./cmd/ix-an simulate --path .env               --out examples/receipts/denied.receipt.json    --key "$seed" --key-id "$kid"
go run ./cmd/ix-an simulate --path docs/approved.txt  --out examples/receipts/approved.receipt.json  --key "$seed" --key-id "$kid" \
  --approve --approver you@example.com --approval-type ticket

echo "== strict verify of generated examples =="
go run ./cmd/ix-an verify examples/receipts/minimal.receipt.json  --strict-hashes --strict-signature
go run ./cmd/ix-an verify examples/receipts/denied.receipt.json   --strict-hashes --strict-signature
go run ./cmd/ix-an verify examples/receipts/approved.receipt.json --strict-hashes --strict-signature --strict-approvals

echo "OK: demo assets generated under examples/receipts (gitignored)"
