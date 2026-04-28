---
decision_id: adr-document-history-review-controls
decision_title: Evidence-Gated Document History And Review Controls
decision_status: deferred
decision_scope: document-lifecycle
decision_owner: platform
---
# ADR: Evidence-Gated Document History And Review Controls

## Status

Deferred after refreshed document lifecycle pressure; kept as reference
evidence for future document-lifecycle design.

This ADR defines how OpenClerk evaluates document history, review, restore,
rollback, and stale-derived-state controls after v0.1.0 and after generalized
artifact-ingestion pressure. It does not add a public runner action, JSON
schema, storage migration, storage behavior, or API.

The product vision is recorded in
[`openclerk-document-post-v0.1.0.md`](openclerk-document-post-v0.1.0.md). The
targeted POC/eval contract is recorded in
[`../evals/document-history-review-controls-poc.md`](../evals/document-history-review-controls-poc.md).
The refreshed promotion decision is recorded in
[`document-lifecycle-promotion-decision.md`](document-lifecycle-promotion-decision.md).

## Context

OpenClerk v1 follows the AgentOps pattern: routine agents use the installed
`openclerk` runner, `openclerk document`, `openclerk retrieval`, and the
OpenClerk skill. Canonical markdown, source-linked synthesis, promoted records,
provenance events, projection freshness, and final-answer-only rejection gates
are the proven slice.

The next document-lifecycle pressure is agent-authored durable edits. When an
agent changes a lasting document, OpenClerk eventually needs to answer what
changed, why it changed, what evidence justified the edit, what prior content
was replaced, whether the change is accepted, pending review, restored, or
superseded, and which derived projections became stale.

Generalized artifact-ingestion pressure increases the privacy and provenance
stakes. Document lifecycle evidence may refer to PDFs, transcripts, receipts,
or other private artifacts, but committed public artifacts must not expose raw
private diffs, private artifact bodies, storage-root paths, or machine-absolute
paths. Lifecycle controls must preserve canonical authority, citations or
source refs, provenance, freshness, local-first operation, operator visibility,
and no-bypass invariants.

That pressure must not turn into hidden autonomous rewrites, a Git clone inside
OpenClerk, or a new public runner surface before evidence shows the current
document and retrieval workflows are either structurally insufficient or
unacceptably costly for routine use.

## Options Considered

- **Storage-only Git or sync history:** keeps byte-level recovery outside
  OpenClerk and remains useful for commits, snapshots, and sync rollback. It
  does not explain semantic document lifecycle state, source evidence, review
  status, derived freshness, or artifact privacy through AgentOps.
- **Semantic OpenClerk revision history:** records OpenClerk-authored document
  changes as knowledge lifecycle events with stable before/after references,
  content hashes, actor/source metadata, edit summaries, source refs,
  provenance, and freshness links. This remains a candidate only if targeted
  evidence shows a capability or ergonomics gap.
- **Pending review queues:** holds agent-authored changes for operator review
  before they become accepted durable knowledge. This may be useful when edits
  are high risk or source-sensitive, but it must not hide review state or make
  autonomous rewrites routine.
- **Restore and rollback controls:** allow an operator or agent to restore an
  OpenClerk-managed accepted revision with an explicit reason and evidence
  trail. This must stay semantic lifecycle control, not storage-level history
  replacement.
- **No new capability:** keep v1 document create, list, get, append,
  replace-section, retrieval, provenance, and projection freshness workflows.
  This remains the default unless targeted evals show a true capability gap or
  repeated unacceptable ergonomics under natural intent.

## Candidate Semantic Model

If promoted later, the model should remain semantic and evidence-linked rather
than file-system centric. The candidate lifecycle record should cover:

- revision records for OpenClerk-authored document changes
- stable content hashes for before and after states
- edit summaries that explain why a change was made
- actor and source metadata for human, agent, import, or projection-originated
  changes
- source refs, citations, provenance events, and projection freshness used
  during the edit
- before/after references that allow diff inspection without committing raw
  private content into public artifacts
- review state such as pending, accepted, rejected, restored, or superseded
- restore or rollback intent, reason, and evidence for OpenClerk-managed
  changes
- private artifact handling that exposes stable references, hashes, summaries,
  or citations instead of raw private bodies in committed reports

These are product-level semantics only. Exact request and response shapes are
out of scope until a later implementation Bead has targeted evidence and
compatibility review.

## Decision

Defer document history and review controls. Keep the refreshed lane as
reference pressure and repair guidance/eval gaps before reconsidering
promotion.

The refreshed report
[`../evals/results/ockp-document-lifecycle-pressure.md`](../evals/results/ockp-document-lifecycle-pressure.md)
evaluated both accepted promotion paths:

- **Capability path:** no promotion. Scripted controls completed for history
  inspection, semantic diff review, restore/rollback, stale synthesis
  inspection, and validation/bypass handling. Existing `openclerk document` and
  `openclerk retrieval` workflows can express those tasks while preserving
  source refs, provenance, freshness, privacy, local-first operation, and
  no-bypass boundaries.
- **Ergonomics path:** defer for repair. The latest guidance repair improved
  parts of the lane, but the committed full run still classified natural
  lifecycle intent as `ergonomics_gap` after 12 tools/commands, 4 assistant
  calls, and 38.70 wall seconds. Diff review remained `skill_guidance`, and
  pending review was reclassified as `data_hygiene` durable-target pressure
  after final-answer guidance passed. That is real pressure but not enough to
  promote a new public runner surface before additional guidance/eval repair
  and repeated evidence.

No public runner action, request or response schema, storage migration, storage
API, semantic history table, review queue, rollback API, or public OpenClerk
interface is promoted by this ADR.

## Promotion Gates

- **Promote for capability** only if repeated targeted AgentOps eval failures
  show the current document, retrieval, provenance, and projection freshness
  workflows cannot safely express needed history inspection, diff review,
  pending review, restore, rollback, private artifact handling, or
  stale-derived-output workflows.
- **Promote for ergonomics** only if scripted controls prove current
  primitives can work, but repeated natural-intent rows show the workflow is
  too slow, too many steps, too brittle, too retry-prone, or too dependent on
  step-by-step guidance for routine use.
- **Defer** when failures are awkward but expressible workflows, one-off
  natural-intent failures, missing skill guidance, data hygiene, thin evidence,
  missing eval coverage, or insufficient dogfooding pressure.
- **Kill** if the candidate duplicates Git or sync history, hides provenance or
  projection freshness, drops citations or source refs, enables hidden
  autonomous rewrites, makes review state invisible, weakens canonical markdown
  authority, exposes raw private diffs in committed artifacts, or requires
  direct SQLite, direct vault inspection, HTTP/MCP, source-built runner paths,
  backend variants, module-cache inspection, or ad hoc runtime programs.

## Invariants

- AgentOps remains the production agent surface.
- Canonical markdown and promoted records remain source authority.
- Source-sensitive claims retain citations, source refs, or stable source
  identifiers.
- Provenance and projection freshness remain inspectable through
  runner-visible evidence.
- Artifact-derived lifecycle evidence must stay local-first and must not leak
  raw private artifact bodies or raw private diffs into committed reports.
- Public artifacts use repo-relative paths or neutral placeholders.
- New public runner actions require separate targeted evidence, compatibility
  review, and an implementation Bead.
