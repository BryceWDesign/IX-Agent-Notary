# IX-Agent-Notary

Proof-carrying agent/tool actions: **policy enforcement + cryptographically signed receipts** that an independent verifier can check in CI, log ingestion, or incident response.

## Why this exists

As soon as an AI agent can touch real systems such as repos, CI/CD, cloud APIs, ticketing, secrets, or production ops, the real enterprise question becomes:

**What exactly did the agent do, under what policy, with what approvals, and can we prove it?**

IX-Agent-Notary is a small trust layer meant to make that answer **machine-verifiable** instead of narrative.

## What it does

1. A tool action is evaluated by **PolicyGate** using an allow/deny policy.
2. The notary emits a **receipt** containing actor, action, timing, policy decision, hashes, and optional approvals.
3. The receipt is canonicalized with **RFC 8785 / JCS** and signed with **ed25519** in v0.
4. A verifier can independently check schema validity, hashes, signatures, approvals, and optional chain linkage.

This is not “trust me” logging. It is evidence that can be independently rejected if it is malformed, unsigned, tampered with, or incomplete.

## 10-minute evaluation

Run the same local validation path that CI runs:

```bash
bash scripts/ci.sh
```

That script:

- enforces `gofmt`
- runs `go vet`
- builds the CLI
- runs `go test ./...`
- generates local dev keys and demo receipts
- verifies generated receipts strictly
- verifies the generated directory strictly

Generate demo assets only:

```bash
bash scripts/gen_demo_assets.sh
```

Verify a generated directory of receipts strictly:

```bash
go run ./cmd/ix-an verify-dir --strict-approvals examples/receipts
```

Important notes:

- This repo intentionally ships **no private keys**.
- This repo intentionally ships **no pre-generated receipts**.
- Demo keys and demo receipts are generated locally and are **gitignored by design**.

## Core capabilities (v0)

- Receipt schema: `spec/receipt.schema.json`
- Draft receipt spec: `spec/receipts.md`
- Policy evaluation pattern: `policy/demo.policy.json`
- Receipt signing: `ed25519`
- Canonical JSON: `RFC8785-JCS`
- Strict verification: schema + hashes + signature + optional approvals + optional chain
- Governance evidence via structured approvals: `docs/APPROVALS.md`
- Receipt storage patterns: directory store + append-only JSONL log: `docs/STORE.md`

## CLI quickstart

All commands below use `go run` directly, so no installation step is required.

### Verify one receipt

```bash
go run ./cmd/ix-an verify --strict-hashes --strict-signature /tmp/allow.receipt.json
```

### Verify a directory of receipts

`verify-dir` is strict on hashes and signatures by design. Add `--strict-approvals` when you want approval signatures enforced too.

```bash
go run ./cmd/ix-an verify-dir --strict-approvals examples/receipts
```

### Simulate a tool action and emit a signed receipt

```bash
go run ./cmd/ix-an simulate --path docs/demo.txt --out /tmp/allow.receipt.json
go run ./cmd/ix-an verify --strict-hashes --strict-signature /tmp/allow.receipt.json
```

### Simulate a denied action

```bash
go run ./cmd/ix-an simulate --path .env --out /tmp/deny.receipt.json
go run ./cmd/ix-an verify --strict-hashes --strict-signature /tmp/deny.receipt.json
```

### Simulate a receipt with governance approval evidence

```bash
go run ./cmd/ix-an simulate \
  --path docs/approved.txt \
  --out /tmp/approved.receipt.json \
  --approve \
  --approver you@example.com \
  --approval-type ticket

go run ./cmd/ix-an verify \
  --strict-hashes \
  --strict-signature \
  --strict-approvals \
  /tmp/approved.receipt.json
```

### Simulate a chained child receipt

A chained child receipt must share the parent trace ID and must reference the parent receipt ID.

```bash
go run ./cmd/ix-an simulate --path docs/chain-root.txt --out /tmp/chain.root.receipt.json
```

Then generate the child using the parent’s `receipt_id` and `trace.trace_id`:

```bash
go run ./cmd/ix-an simulate \
  --path docs/chain-child.txt \
  --out /tmp/chain.child.receipt.json \
  --trace-id <parent-trace-id> \
  --step 2 \
  --parent-receipt-id <parent-receipt-id>
```

Verify the child with strict chain validation:

```bash
go run ./cmd/ix-an verify \
  --strict-chain \
  --chain-dir /tmp \
  /tmp/chain.child.receipt.json
```

For a complete end-to-end chained example without manual extraction, use:

```bash
bash scripts/gen_demo_assets.sh
```

### Append-only JSONL log

Append a receipt after strict verification:

```bash
go run ./cmd/ix-an store append --in /tmp/approved.receipt.json --log /tmp/receipts.jsonl
```

Verify the entire log:

```bash
go run ./cmd/ix-an store verify-log --log /tmp/receipts.jsonl
```

## Where this fits

IX-Agent-Notary is the enforcement-and-evidence layer that sits between **agents** and **tools**.

It is meant to:

- prevent unsafe calls through policy
- produce verifiable receipts for audit, compliance, and incident response
- give buyers a narrow, reviewable trust boundary instead of asking them to trust the agent stack itself

It is **not** a full agent framework.  
It is **not** a SIEM.  
It is the part you want to be able to verify.

## Document map

Start here:

- Architecture: `docs/ARCHITECTURE.md`
- Threat model: `docs/THREAT_MODEL.md`
- Key management: `docs/KEY_MANAGEMENT.md`
- Approvals: `docs/APPROVALS.md`
- Receipt store: `docs/STORE.md`
- Policy integrity: `docs/POLICY_INTEGRITY.md`
- Enterprise pilot guide: `docs/ENTERPRISE_PILOT.md`
- Design partner notes: `docs/DESIGN_PARTNER.md`

## License and commercial use

IX-Agent-Notary is source-available for evaluation under `LICENSE`.

If you want to use it in production or any commercial context, you need a separate commercial license. See `COMMERCIAL.md` for trigger conditions and contact guidance.

## Security

Please report security issues according to `SECURITY.md`.

If you are evaluating agent governance and need receipts that security or compliance teams can independently verify, start with:

- `docs/ENTERPRISE_PILOT.md`
- `docs/THREAT_MODEL.md`
- `docs/KEY_MANAGEMENT.md`
