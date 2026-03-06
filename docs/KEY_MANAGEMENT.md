# Key Management (v0)

Receipts are only as trustworthy as the signing keys behind them.

This repo intentionally **does not** ship any private signing material. Private keys must never be committed to a public repository.

## Local evaluation

Generate a local dev keypair and local example receipts:

```bash
bash scripts/gen_demo_assets.sh
```

That creates gitignored local artifacts such as:

- `keys/dev/dev-key-001.seed` — private ed25519 seed
- `keys/dev/dev-key-001.pub` — public key
- `examples/receipts/*.json` — generated example receipts

Verify the generated receipts strictly:

```bash
go run ./cmd/ix-an verify-dir --strict-approvals examples/receipts
```

## Manual local flow

Generate a local keypair:

```bash
go run ./cmd/ix-an keygen --out-seed keys/dev/dev-key-001.seed --out-pub keys/dev/dev-key-001.pub
```

Create a signed receipt with that key:

```bash
go run ./cmd/ix-an simulate \
  --path docs/demo.txt \
  --out /tmp/allow.receipt.json \
  --key keys/dev/dev-key-001.seed \
  --key-id dev-key-001
```

Verify it with the matching public key:

```bash
go run ./cmd/ix-an verify \
  --strict-hashes \
  --strict-signature \
  --pubkey keys/dev/dev-key-001.pub \
  /tmp/allow.receipt.json
```

Using `--pubkey` is the most explicit and deterministic way to verify a receipt during evaluation.

## Baseline production posture

### 1) Store private keys in KMS or HSM

Private key material should be hardware-backed or at least controlled by a hardened signing service.

Minimum posture:

- no raw private keys in public repos
- no broad filesystem access to signing keys
- sign-only permission boundary where possible
- IAM and change control around key usage

### 2) Publish a trusted public-key allowlist

Verification should only accept signatures from a curated set of trusted public keys mapped to known `key_id` values.

At minimum, production needs:

- an explicit list of trusted public keys
- stable `key_id` naming
- a process for updating trust when keys rotate or are revoked

### 3) Rotate keys without breaking old verification

Receipts carry:

- `integrity.signature.key_id`

Recommended pattern:

- treat `key_id` as a specific key version, not a floating alias
- rotate by minting a new key and a new `key_id`
- keep historical public keys available so older receipts remain verifiable

### 4) Separate trust domains where appropriate

Higher-assurance deployments may want:

- one trust domain for receipt signing
- another trust domain for approval signing
- separate operational ownership for each

That reduces the blast radius of a single compromise.

## What good key hygiene protects against

Good key handling helps defend against:

- receipt tampering
- fabricated receipts
- unverifiable “audit theater”
- silent evidence drift
- accidental trust in unknown signing identities

## Practical rule

If a buyer cannot answer **which keys are trusted, where they live, how they rotate, and how old receipts stay verifiable**, the evaluation is not production-credible yet.

## Related documents

- `docs/THREAT_MODEL.md`
- `docs/POLICY_INTEGRITY.md`
- `docs/APPROVALS.md`
