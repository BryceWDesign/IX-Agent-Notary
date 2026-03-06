# Enterprise Pilot Guide (v0)

This is the shortest serious path to evaluate IX-Agent-Notary in a real environment.

## Goal

Prove that agent/tool execution can be:

- **policy-enforced**
- **auditable**
- **tamper-evident**
- **governance-aware**
- **incident-friendly**

In other words: prove that the organization can independently reconstruct and verify what happened.

## Recommended pilot architecture

### Minimal

Agent → Notary wrapper → Tool(s)  
Notary emits receipts → receipt store → verifier in CI or ingest pipeline

### Better

Agent → tool gateway → Notary enforcement and signing → Tool(s)  
Receipts → immutable or append-only store → SIEM or pipeline verifies on ingest

## Pilot sequence

### 1) Start with a narrow allowlist

Begin with deny-by-default and a tiny safe allowlist.

Use `policy/demo.policy.json` as a pattern:

- allow only a safe prefix such as `docs/`
- explicitly deny sensitive targets such as `.env`
- keep the policy small enough to review line by line

### 2) Make Notary the required path

The single most important architectural requirement is this:

> Tools must not be reachable unless the call passes through the notary control point.

If the agent can bypass the notary, the receipts become optional storytelling instead of mandatory evidence.

### 3) Run strict verification in CI

A credible pilot should gate merges or ingest on verification, including:

- schema validity
- strict hashes
- strict signatures
- strict chain verification when parent links exist
- strict approvals when governance evidence is required

This repo already demonstrates that posture through:

- `.github/workflows/ci.yml`
- `scripts/ci.sh`

### 4) Store receipts in a way that is easy to verify

Directory model:

```bash
go run ./cmd/ix-an verify-dir --strict-approvals examples/receipts
```

JSONL log model:

```bash
go run ./cmd/ix-an store append --in /tmp/approved.receipt.json --log /tmp/receipts.jsonl
go run ./cmd/ix-an store verify-log --log /tmp/receipts.jsonl
```

In production, place that storage behind immutability controls such as WORM or object lock.

### 5) Tie decisions to exact policy content

Require `policy.policy_hash` in receipts so a buyer can tell exactly which policy content produced the decision, not just a policy ID string.

### 6) Add approvals for higher-risk actions

For actions outside the normal low-risk allowlist, require structured approvals and verify them.

Demo example:

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

## Success criteria

A serious buyer should be able to confirm all of the following:

- receipts alone are enough to reconstruct the important parts of execution
- tampering causes verification failure
- policy identity is bound to the evidence
- approval evidence is structured and verifiable
- the notary can be made the mandatory path to execution

## What usually convinces security and platform teams

The repo becomes much more credible when evaluators can see that it:

- builds cleanly from a fresh clone
- tests cleanly without hidden local assets
- generates demo evidence locally
- rejects malformed or unverifiable receipts
- can be slotted into existing CI or log-ingest controls

## Near-term enterprise extensions

Likely next steps after a pilot:

- KMS or HSM signing
- approval trust-domain separation
- transparency-log or immutable-registry integration
- tool adapters for real APIs and cloud controls
- policy-pack distribution and attestation workflows
