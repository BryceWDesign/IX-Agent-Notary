# IX-Agent-Notary — Threat Model (v0)

This document explains what IX-Agent-Notary is defending against, what it does **not** claim to solve, and the controls that make it enterprise-evaluable.

---

## Scope

IX-Agent-Notary is a **policy enforcement + evidence** layer for agent/tool execution.

It is designed to produce **tamper-evident, cryptographically signed receipts** that can be verified independently.

---

## Assets we protect

1. **Receipt integrity**
   - Receipts must not be modifiable without detection.

2. **Policy decision integrity**
   - The allow/deny decision and matched-rule evidence must be accurate and non-forgeable.

3. **Key material**
   - Signing keys must remain confidential; public keys must be authentic for verification.

4. **Auditability / traceability**
   - Receipts must be correlatable across steps (trace IDs, parent linkage) without easy truncation.

5. **Data minimization**
   - Receipts should avoid leaking secrets while still providing verifiable evidence (hash + redaction patterns).

---

## Trust boundaries

### Trusted (must be hardened / reviewable)
- PolicyGate evaluator (decision + matched rules)
- Receipt composer + canonicalization (RFC8785-JCS)
- Signing + key handling
- Verifier logic

### Untrusted (assumed fallible or compromis-able)
- Agent orchestration code / prompts / LLM outputs
- Tool targets and tool responses (captured as evidence, not trusted as “truth”)
- Receipt storage/log sink (may be tampered with; verification should detect)

---

## Attacker capabilities assumed

- Can manipulate agent prompts, tool selection, and inputs.
- Can attempt to inject or coerce the agent into unsafe actions.
- Can attempt to modify stored receipts (log tampering).
- Can attempt to replay old receipts.
- Can attempt to downgrade verification (skip signature checks, accept placeholders).
- May gain partial access to a host running agent logic.

---

## Primary threats & mitigations

### T1: Receipt tampering (edit fields after the fact)
**Threat:** An attacker edits `action`, `policy`, or `result` to hide behavior.  
**Mitigations:**
- Canonicalization (RFC8785-JCS) + signature verification
- Verifier fails if signature invalid
- Strict mode requires signatures and real hashes

### T2: Receipt fabrication (invent a receipt that never happened)
**Threat:** Create a “fake allowed action” receipt.  
**Mitigations:**
- Only the Notary runtime signs receipts
- Verifier requires a resolvable public key and valid signature
- Key IDs map to known public keys

### T3: Placeholder hashes/signatures in “production”
**Threat:** Receipts pass review even though they contain `PLACEHOLDER_*`.  
**Mitigations:**
- `--strict-hashes` and `--strict-signature` verification modes
- CI policy can require strict verification for acceptance

### T4: Policy bypass (execute tool without PolicyGate)
**Threat:** Agent calls tools directly and never produces receipts.  
**Mitigations:**
- Architectural control: place Notary as the execution mediator (control plane or CI gate)
- Enterprise deployment must enforce “tools only reachable through Notary” (network/IAM design)

### T5: Policy forgery (claim a decision came from a different policy)
**Threat:** Swap `policy_id` or “matched rules” to justify an action.  
**Mitigations:**
- Policy decision is covered by the receipt signature
- Verifier checks schema and (in future) can verify policy packs by hash/ID mapping

### T6: Receipt truncation / step dropping
**Threat:** Remove “bad” receipts from the middle of a workflow.  
**Mitigations (v0 + roadmap):**
- Trace IDs exist now
- Roadmap: enforce parent linkage + append-only receipt store semantics + monotonic step expectations

### T7: Replay attacks
**Threat:** Reuse an old “allowed” receipt to justify a new action.  
**Mitigations (v0 + roadmap):**
- Receipts contain timestamps + trace IDs
- Roadmap: include nonces / execution instance IDs / tool challenge-response bindings

### T8: Signing key compromise
**Threat:** Attacker steals signing key and forges receipts.  
**Mitigations:**
- Dev keys are demo-only (explicit)
- Production guidance: store keys in KMS/HSM, rotate keys, key-scoped usage
- Roadmap: support key rotation + key transparency / revocation lists

### T9: Verifier compromise / bad verifier configuration
**Threat:** A verifier is configured to skip checks.  
**Mitigations:**
- “Strict mode” flags make verification posture explicit
- Enterprise guidance should mandate strict verification in CI / SOC ingestion

### T10: Sensitive data leaks via receipts
**Threat:** Secrets land in `action.parameters` or `result.output`.  
**Mitigations:**
- Redaction pattern: store redacted objects + hash the canonicalized content
- Roadmap: built-in redaction policies + field allowlists

---

## What IX-Agent-Notary does NOT claim (important)

- It does **not** guarantee tool outputs are truthful or safe.
- It does **not** prevent misuse if policy allows a dangerous action.
- It does **not** replace IAM; it complements IAM with action-level evidence.
- It does **not** magically secure a fully compromised host (you still need OS, IAM, and network controls).

---

## Security posture goals (what “good” looks like)

For an enterprise pilot to be credible:
- Notary is the only path to tool execution (architectural enforcement)
- Receipts are verified in strict mode in CI and/or in log ingestion pipelines
- Keys are rotated and protected (KMS/HSM)
- Policies are explicit allowlists with approvals for high-risk actions

This is the posture IX-Agent-Notary is designed to support.
