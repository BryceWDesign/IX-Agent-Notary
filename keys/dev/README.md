# Dev keys (local evaluation)

This directory exists so evaluators have a standard place for **locally-generated** dev keys.

Important:
- **No private keys are committed to this repo.**
- `*.seed` and `*.pub` under `keys/dev/` are gitignored by design.
- Treat anything under `keys/dev/` as **demo-only**. Do not use in production.

---

## Generate dev keys + demo receipts (recommended)

Run:

```bash
bash scripts/gen_demo_assets.sh

That script will:

generate a local ed25519 keypair (seed + public key), and

generate example receipts under examples/receipts/, and

strictly verify the results.

Outputs (gitignored):

keys/dev/dev-key-001.seed

keys/dev/dev-key-001.pub

examples/receipts/*.json

