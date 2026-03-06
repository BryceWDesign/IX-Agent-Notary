# Approvals (v0)

Enterprises do not just want “the policy allowed it.” They want governance evidence:

- who approved
- what exactly was approved
- when it was approved
- whether the approval can be independently verified

IX-Agent-Notary models that evidence as structured objects inside `policy.approvals[]`.

## Approval object

Each approval is a JSON object with these required fields:

- `approval_id` — unique identifier for the approval record
- `type` — one of:
  - `human`
  - `ticket`
  - `breakglass`
- `status` — one of:
  - `requested`
  - `approved`
  - `denied`
  - `expired`
  - `revoked`
- `approver` — object with:
  - `type`
  - `id`
  - optional `display`
- `scope` — object with:
  - `kind`
  - `tool`
  - `operation`
  - optional `resource`
- `time` — object with:
  - `requested_at`
  - `decided_at`
  - optional `expires_at`

Optional fields:

- `notes`
- `signature`

## Approval signature object

When present, `signature` must be an object with:

- `alg`
- `key_id`
- `value`

In the current implementation, approval signing uses the same canonicalization rule as receipt signing:

- canonical JSON via **RFC8785-JCS**
- sign the approval payload while excluding `signature.value`
- store the signature as base64url text

## Current implementation behavior

### What the simulator does

The simulator can embed a single approval record when `--approve` is provided.

Example:

```bash
go run ./cmd/ix-an simulate \
  --path docs/approved.txt \
  --out /tmp/approved.receipt.json \
  --approve \
  --approver you@example.com \
  --approval-type ticket
```

In the current demo flow, the simulator signs the approval object using the same signing key used for the receipt.

### What strict approval verification means

Use `--strict-approvals` when verifying a receipt:

```bash
go run ./cmd/ix-an verify \
  --strict-hashes \
  --strict-signature \
  --strict-approvals \
  /tmp/approved.receipt.json
```

Under strict approval verification:

- if approvals are present, each approval must include a signature
- each approval signature must verify successfully
- malformed or unsigned approvals cause verification failure

Without `--strict-approvals`, approvals may still be present in the receipt, but the verifier will not require signatures on them.

## Why approvals matter

Approvals turn receipts into governance artifacts instead of plain execution logs.

That matters for:

- SOC 2 and ISO 27001 evidence trails
- change-management linkage
- break-glass event recording
- higher-assurance workflows where risky actions need separate human or ticket authorization

## Practical interpretation

An approval does **not** replace policy.  
It complements policy.

The pattern is:

1. policy says whether the action class is even eligible
2. approvals carry governance context for actions that require explicit sign-off
3. the receipt binds policy evidence, approval evidence, and execution evidence into one verifiable record

## v0 scope

Current v0 support is intentionally small:

- one or more structured approval objects in `policy.approvals[]`
- optional approval signatures
- strict verification mode for approval signatures

Future enterprise-grade extensions could add:

- quorum approvals
- separate approval trust domains
- external ticket-system binding
- expiry and revocation workflows tied to policy engines
