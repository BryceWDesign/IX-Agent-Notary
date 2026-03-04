# Key Management (v0)

IX-Agent-Notary receipts are only as trustworthy as their signing keys.

This doc is **operational guidance** for serious evaluation (and future productionization).

---

## What exists today (in this repo)
- Receipts are signed with **ed25519**.
- Demo keys under `keys/dev/` exist **only** to make local verification easy.
- The verifier can also be given a public key directly via `--pubkey`.

**Important:** demo keys are not secure. They are for evaluation only.

---

## What a real deployment should do

### 1) Store signing keys in KMS/HSM
- Keep ed25519 private key material in a hardware-backed store if possible
- Limit use to “sign receipt” operations only
- Require IAM authorization and (optionally) approvals for signing in high-risk environments

### 2) Rotate keys (and keep receipts verifiable forever)
Receipts reference:
- `integrity.signature.key_id`

A practical approach:
- Keep `key_id` immutable for a key version (e.g. `notary-prod-2026-03`)
- Rotate by issuing a new `key_id`
- Publish public keys for all historical key_ids so old receipts remain verifiable

### 3) Treat public keys as an “allowlist”
Your SOC/CI should verify signatures only against:
- a trusted set of public keys
- with known key_ids
- and (optionally) a revocation list

### 4) Make verifier posture explicit (strict by default in CI)
- Always run:
  - strict core hashes
  - strict signature verification
  - strict chain verification (when parent linkage exists)
- Example:
  - `ix-an verify-dir <dir>` (strict by default)
  - `ix-an store verify-log --log <file>` (strict)

---

## Threats this mitigates
- Receipt tampering (invalid signature)
- Receipt fabrication (unknown key_id / unknown public key)
- “audit theater” placeholders (strict verifier rejects)
- Silent policy drift (use `policy_hash`)

See also:
- `docs/THREAT_MODEL.md`
- `docs/POLICY_INTEGRITY.md`
