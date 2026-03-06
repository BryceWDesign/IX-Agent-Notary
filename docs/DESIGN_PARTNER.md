# Design Partner

If you are evaluating agent governance seriously, this document is the fastest path from “interesting repo” to “real pilot.”

## Best-fit design partner profile

IX-Agent-Notary is a strong fit when you need one or more of these:

- a mandatory policy gate in front of agent tool use
- tamper-evident receipts for audit or incident review
- stronger evidence around approvals or higher-risk actions
- a narrow trust layer security teams can actually inspect

## What a good first pilot looks like

The cleanest first pilot is small:

- one agent or automation path
- one tool plane
- one allow/deny policy pack
- one receipt destination
- two or three audit questions you care about

Examples:

- “Can we prove when an agent tried to write outside allowed repo paths?”
- “Can we show which policy caused an allow or deny?”
- “Can we require approval evidence for higher-risk actions?”
- “Can we reject unverifiable receipts at ingest time?”

## What to bring to the conversation

Bring these inputs:

1. your tool plane
2. your environment boundary
3. your must-have audit questions
4. your receipt destination
5. your compliance or evidence requirements

That is enough to tell whether a pilot is real or just theoretical.

## Engagement shape

Typical sequence:

### Phase 1 — evaluation fit
- define one narrow workflow
- define one deny-by-default policy boundary
- define what evidence must be captured

### Phase 2 — pilot implementation
- enforce the no-bypass path
- emit signed receipts for the chosen workflow
- verify receipts in CI, ingest, or review pipeline

### Phase 3 — production hardening decision
- key posture review
- storage integrity posture review
- approval workflow review
- commercial licensing decision

## What a partner should expect back

A serious design-partner effort should produce:

- a pilot architecture that is actually deployable
- a clear list of trust assumptions
- a policy pack aligned to the chosen workflow
- receipt examples that answer real audit questions
- a practical hardening gap list for production

## How to start

Open:

- `.github/ISSUE_TEMPLATE/commercial-licensing.md`

Use the title:

- `Commercial licensing / design partner`

If you want to avoid public detail, open a minimal issue that says:

- `Requesting private commercial channel`

Primary public business contact path:

- `https://www.linkedin.com/in/brycewdesign/`
