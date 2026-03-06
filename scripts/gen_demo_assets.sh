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

echo "== generating real chained example receipts =="
chain_root="examples/receipts/chain.root.receipt.json"
chain_child="examples/receipts/chain.child.receipt.json"

go run ./cmd/ix-an simulate --path docs/chain-root.txt --out "$chain_root" --key "$seed" --key-id "$kid"

chain_receipt_id="$(grep -m1 '"receipt_id"' "$chain_root" | sed -E 's/.*"receipt_id": "([^"]+)".*/\1/')"
chain_trace_id="$(grep -m1 '"trace_id"' "$chain_root" | sed -E 's/.*"trace_id": "([^"]+)".*/\1/')"

if [[ -z "$chain_receipt_id" || -z "$chain_trace_id" ]]; then
  echo "FAIL: unable to extract chain metadata from $chain_root" >&2
  exit 1
fi

go run ./cmd/ix-an simulate \
  --path docs/chain-child.txt \
  --out "$chain_child" \
  --key "$seed" \
  --key-id "$kid" \
  --trace-id "$chain_trace_id" \
  --step 2 \
  --parent-receipt-id "$chain_receipt_id"

echo "== strict verify of generated examples =="
go run ./cmd/ix-an verify --strict-hashes --strict-signature examples/receipts/minimal.receipt.json
go run ./cmd/ix-an verify --strict-hashes --strict-signature examples/receipts/denied.receipt.json
go run ./cmd/ix-an verify --strict-hashes --strict-signature --strict-approvals examples/receipts/approved.receipt.json
go run ./cmd/ix-an verify --strict-hashes --strict-signature --strict-chain "$chain_child"

echo "OK: demo assets generated under examples/receipts (gitignored)"
