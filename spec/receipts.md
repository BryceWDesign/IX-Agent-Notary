# IX-Agent-Notary Receipt Specification (Draft)

Status: **Draft**  
Current spec version: **0.1.0**

This document defines the receipt format emitted by IX-Agent-Notary. A receipt is a structured, tamper-evident record of a tool or action request, the policy decision that governed it, the resulting effects, and the cryptographic material needed for independent verification.

## 1) Goals

A receipt must let an independent verifier answer:

1. **What happened?**
2. **Who requested it?**
3. **Who mediated or executed it?**
4. **Was it allowed or denied, and why?**
5. **What was produced or changed?**
6. **Can I trust this record cryptographically?**
7. **Can I follow it across multiple workflow steps?**

## 2) Terminology

- **Agent**: orchestration logic requesting actions; treat as fallible or untrusted
- **Notary runtime**: the enforcement boundary that evaluates policy and emits signed receipts
- **Tool**: any invoked capability such as a CLI, API call, CI job, or filesystem action
- **Receipt**: a signed JSON document
- **Verifier**: software that checks schema, hashes, signatures, approvals, and chain linkage

## 3) Canonicalization and signing rules

Receipts are JSON objects signed over a **canonical byte representation**.

### 3.1 Canonical JSON

- Canonicalization algorithm: **JCS (JSON Canonicalization Scheme, RFC 8785)**
- The canonical JSON bytes are the input to hashing and signing

### 3.2 Hashing

- Default digest algorithm in v0: `SHA-256`
- Hashes are encoded according to `integrity.hash.encoding`
- Current implementation uses `base64url` encoding for receipt core hashes

### 3.3 Receipt signature

Receipts include a signature envelope at `integrity.signature`.

Required fields:

- `integrity.signature.alg`
- `integrity.signature.key_id`
- `integrity.signature.value`

The signature is computed over canonicalized receipt JSON while excluding the field `integrity.signature.value` itself.

Rule: verifiers must recompute the canonical form and verify the signature. They must not trust stored bytes.

## 4) Receipt object

A receipt is a JSON object with these top-level fields:

- `receipt_version` — string, required
- `receipt_id` — string, required
- `time` — object, required
- `trace` — object, required
- `actor` — object, required
- `notary` — object, required
- `action` — object, required
- `policy` — object, required
- `result` — object, required
- `integrity` — object, required

### 4.1 `time`

Required:

- `time.requested_at` — RFC3339 timestamp
- `time.decided_at` — RFC3339 timestamp
- `time.completed_at` — RFC3339 timestamp

### 4.2 `trace`

Required:

- `trace.trace_id` — stable workflow ID
- `trace.step` — integer step index, starting at `1`

Optional:

- `trace.parent_receipt_id` — parent receipt ID for a chained workflow

### 4.3 `actor`

Required:

- `actor.type`
- `actor.id`

Optional:

- `actor.display`
- `actor.session_id`

### 4.4 `notary`

Required:

- `notary.runtime`
- `notary.version`
- `notary.instance_id`
- `notary.environment`

### 4.5 `action`

Required:

- `action.kind`
- `action.tool`
- `action.operation`
- `action.parameters`
- `action.parameters_hash`

`action.parameters_hash` is the hash of canonicalized `action.parameters`.

Rule: if parameters contain sensitive material, the receipt may store a redacted form in `action.parameters`, while preserving the canonical hash for integrity linkage.

### 4.6 `policy`

Required:

- `policy.policy_id`
- `policy.decision` — `allow` or `deny`
- `policy.reason`
- `policy.rules`
- `policy.approvals`

#### 4.6.1 `policy.rules`

`policy.rules` is an array of objects with:

- `rule_id` — required
- `effect` — required, `allow` or `deny`
- `explanation` — optional

#### 4.6.2 `policy.approvals`

`policy.approvals` is an array and may be empty.

Each approval object contains these required fields:

- `approval_id`
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

Optional approval fields:

- `notes`
- `signature`

#### 4.6.3 Approval signature

If an approval carries a signature, it must contain:

- `signature.alg`
- `signature.key_id`
- `signature.value`

Current implementation signs the approval payload using canonical JSON and excludes `signature.value` from the signed payload, mirroring the receipt-signature pattern.

A verifier may choose to enforce approval signatures only in strict approval mode.

#### 4.6.4 Optional policy integrity fields

The receipt may also carry:

- `policy.policy_hash`
- `policy.policy_source`
- `policy.context_hashes`

`policy_hash` is intended to bind the policy decision to exact policy content, not just a policy ID string.

### 4.7 `result`

Required:

- `result.status`
- `result.summary`
- `result.output`
- `result.output_hash`

Common values for `result.status` include `success`, `failure`, and `denied`.

`result.output_hash` is the hash of canonicalized `result.output`.

### 4.8 `integrity`

Required:

- `integrity.canonicalization` — currently `RFC8785-JCS`
- `integrity.hash`
- `integrity.signature`

#### 4.8.1 `integrity.hash`

Required fields:

- `integrity.hash.alg`
- `integrity.hash.encoding`

#### 4.8.2 `integrity.signature`

Required fields:

- `integrity.signature.alg`
- `integrity.signature.key_id`
- `integrity.signature.value`

Optional:

- embedded public key material for demo scenarios, though that is discouraged for production systems

## 5) Versioning rules

- backward-compatible additions should bump the minor version
- breaking changes should bump the major version once the format stabilizes
- every receipt must include `receipt_version`

## 6) Chain semantics

When `trace.parent_receipt_id` is present, strict chain verification expects:

- the parent receipt to exist
- parent and child to share the same `trace.trace_id`
- the parent step to be exactly one less than the child step
- the root receipt to terminate at `step = 1`

This is meant to make evidence truncation or malformed parent linkage detectable.

## 7) Examples

Example receipts are generated locally by:

```bash
bash scripts/gen_demo_assets.sh
```

Generated outputs typically include:

- `examples/receipts/minimal.receipt.json`
- `examples/receipts/denied.receipt.json`
- `examples/receipts/approved.receipt.json`
- `examples/receipts/chain.root.receipt.json`
- `examples/receipts/chain.child.receipt.json`

These are evaluation artifacts and are gitignored by design.

## 8) Design intent

This receipt format intentionally separates:

- **policy evidence**
- **action evidence**
- **result evidence**
- **cryptographic integrity**
- **workflow linkage**

That separation is what makes the record useful to auditors, platform teams, and incident responders instead of only to the runtime that created it.
