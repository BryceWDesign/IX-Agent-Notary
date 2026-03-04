# Receipt Store (v0)

IX-Agent-Notary supports a minimal “receipt store” pattern for keeping evidence.

The goal is simple: receipts should be easy to write, easy to verify, and hard to tamper with.

---

## 1) Directory store

Store receipts as individual `.json` files in a directory.

Verify a directory strictly:

```bash
go run ./cmd/ix-an verify-dir examples/receipts --strict-approvals

What this enforces (strict by default in verify-dir):

schema validity

strict core hashes

strict signature verification

optional chain verification when parent_receipt_id exists

2) Append-only log store (JSONL)

You can ingest receipts into an append-only JSON Lines log.

Append (strictly verified before ingest):
go run ./cmd/ix-an store append --in /tmp/approved.receipt.json --log /tmp/receipts.jsonl

Verify the entire log:
go run ./cmd/ix-an store verify-log --log /tmp/receipts.jsonl

Notes:

The JSONL log is “append-only” by convention; IX-Agent-Notary detects tampering via signatures (and optionally chain pointers).

A real deployment should put the log behind immutability controls (WORM / S3 Object Lock / append-only DB).
