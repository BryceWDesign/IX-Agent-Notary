# Policies

This folder contains **policy packs** for the PolicyGate evaluator.

- `demo.policy.json` is a deliberately small allowlist policy:
  - Denies writes to `.env`
  - Allows writes only under `docs/`
  - Default effect: **deny**

This is not meant to be a full policy language yet—just a credible v0 that proves:
1) deterministic allow/deny decisions
2) a machine-verifiable receipt contains the policy decision + matched rule evidence
