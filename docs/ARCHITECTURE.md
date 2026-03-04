# IX-Agent-Notary — Architecture

## Problem statement
When an agent can call tools (code, CI/CD, cloud APIs, ticketing, secrets, production ops), the enterprise question is:

**What exactly did the agent do, under what policy, with what approvals — and can we prove it?**

IX-Agent-Notary exists to make **policy enforcement + verifiable evidence** first-class.

---

## Design goals
1. **Tamper-evident receipts** for every meaningful action (inputs, policy decision, outputs, timing, identity).
2. **Strict verification** so consumers can reject placeholders or unverifiable evidence.
3. **Small trusted boundary**: keep the enforcement + receipt/signing core reviewable.
4. **Composable storage**: directory store or append-only JSONL log patterns.

---

## Components
- **PolicyGate (enforcement)**  
  Evaluates a policy pack and returns allow/deny with structured evidence and reason.
- **Tool mediator / executor**  
  Executes (or simulates) tool actions through the PolicyGate decision.
- **Receipt composer + canonicalization**  
  Builds a receipt and canonicalizes payloads (RFC 8785 / JCS) before hashing/signing.
- **Signer**  
  Signs receipts (ed25519 in v0) and writes `integrity.signature`.
- **Verifier (consumer)**  
  Validates schema, hashes, signature, optional approvals, and optional chain linkage.

---

## Data flow (canonical)

```mermaid
flowchart LR
  subgraph AgentHost["Agent Host (untrusted logic)"]
    A["Agent / Tool Caller"]
  end

  subgraph Notary["IX-Agent-Notary (trusted runtime boundary)"]
    PG["PolicyGate (allow/deny + reason)"]
    EX["Tool Mediator/Executor"]
    RC["Receipt Composer"]
    SG["Signer"]
  end

  subgraph External["External Tools / Systems"]
    T["Tool/API Target"]
    ST["Receipt Store (dir / jsonl)"]
    V["Verifier (CLI or service)"]
  end

  A -->|"intent + context"| PG
  PG -->|"decision + evidence"| EX
  EX -->|"tool call"| T
  T -->|"result"| EX
  EX --> RC
  RC --> SG
  SG -->|"signed receipt"| ST
  ST --> V

Trust boundaries
Trusted boundary (minimum viable TCB)

These pieces must be small, reviewable, and hardened:

policy decision evaluator (PolicyGate)

receipt construction + canonicalization (Receipt Composer)

signing + key handling (Signer)

verification logic (Verifier)

Untrusted / assumed-compromisable

Assume these can be wrong or compromised:

agent orchestration logic

prompts / LLM outputs

upstream planner code

tool response content (capture as evidence; don’t trust as truth)

Principle: assume the agent is fallible; rely on enforcement + receipts, not “agent honesty.”

Receipt chain concept

Receipts can form a linked chain (like an evidence log):

each receipt may reference a parent_receipt_id

workflows share a trace_id

receipts include hashes of inputs/outputs and policy context

This makes it harder to drop/reorder steps without detection (when chain validation is enabled).

Key posture

No private keys are shipped in the repo.

Local evaluation keys are generated on the evaluator’s machine (scripts/gen_demo_assets.sh).

Production should use KMS/HSM-backed keys and an explicit trusted public-key allowlist.

See docs/KEY_MANAGEMENT.md.
