That script creates:

keys/dev/dev-key-001.seed (0600, gitignored)

keys/dev/dev-key-001.pub (gitignored)

example receipts under examples/receipts/ (gitignored)

What a real deployment should do
1) Store signing keys in KMS/HSM

Keep ed25519 private key material in a hardware-backed store when possible

Limit use to “sign receipt” operations only

Require IAM authorization and (optionally) approvals for signing in high-risk environments

2) Rotate keys (and keep receipts verifiable forever)

Receipts reference:

integrity.signature.key_id

A practical approach:

Keep key_id immutable for a key version (e.g. notary-prod-2026-03)

Rotate by issuing a new key_id

Publish public keys for historical key_ids so old receipts remain verifiable

3) Treat public keys as an allowlist

Your SOC/CI should verify signatures only against:

a trusted set of public keys

with known key_ids

and (optionally) a revocation list

4) Make verifier posture explicit (strict by default in CI)

Always run:

strict core hashes

strict signature verification

strict approvals (if you require approvals)

strict chain verification (when parent linkage exists)

Threats this mitigates

Receipt tampering (invalid signature)

Receipt fabrication (unknown key_id / unknown public key)

“audit theater” placeholders (strict verifier rejects)

Silent policy drift (use policy_hash when enabled)

See also:

docs/THREAT_MODEL.md

docs/POLICY_INTEGRITY.md

