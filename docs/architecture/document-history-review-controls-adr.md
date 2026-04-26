---
decision_id: adr-document-history-review-controls
decision_title: Evidence-Gated Document History And Review Controls
decision_status: deferred
decision_scope: document-lifecycle
decision_owner: platform
---
# ADR: Evidence-Gated Document History And Review Controls

## Status

Deferred after targeted AgentOps eval evidence; kept as reference evidence for
future document-lifecycle pressure.

This ADR defines how OpenClerk will evaluate document history and review
controls after v0.1.0. It does not add a public runner action, JSON schema,
storage migration, or API.

The product vision is recorded in
[`openclerk-document-post-v0.1.0.md`](openclerk-document-post-v0.1.0.md). The
targeted POC/eval contract is recorded in
[`../evals/document-history-review-controls-poc.md`](../evals/document-history-review-controls-poc.md).

## Context

OpenClerk v1 follows the AgentOps pattern: routine agents use the installed
`openclerk` runner, `openclerk document`, `openclerk retrieval`, and the
OpenClerk skill. Canonical markdown, source-linked synthesis, promoted records,
provenance events, projection freshness, and final-answer-only rejection gates
are the proven slice.

The next document-lifecycle pressure is agent-authored durable edits. When an
agent changes a lasting document, OpenClerk eventually needs a runner-visible
way to answer what changed, why it changed, what evidence justified the edit,
what prior content was replaced, and whether the change is accepted, pending
review, restored, or superseded.

That pressure must not turn into hidden autonomous rewrites, a Git clone inside
OpenClerk, or a new public runner surface before evidence shows the current
document and retrieval workflows are structurally insufficient.

## Options Considered

- **Storage-only Git or sync history:** keeps byte-level recovery outside
  OpenClerk and remains useful for commits, snapshots, and sync rollback. It
  does not explain semantic document lifecycle state, source evidence, review
  status, or derived freshness through AgentOps.
- **Semantic OpenClerk revision history:** records OpenClerk-authored document
  changes as knowledge lifecycle events with stable before/after references,
  content hashes, actor/source metadata, edit summaries, source refs,
  provenance, and freshness links. This is the main candidate, but it requires
  targeted evidence before promotion.
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
  This remains the default unless dogfooding and targeted evals show a true
  capability gap.

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

These are product-level semantics only. Exact request and response shapes are
out of scope until a later implementation Bead has targeted evidence and
compatibility review.

## Decision

Defer document history and review controls until dogfooding and targeted
AgentOps evals justify promotion.

Before any implementation issue is filed, a reduced eval report and decision
note must show that existing `openclerk document` and `openclerk retrieval`
workflows are structurally insufficient for reliable document lifecycle
control while preserving the v1 invariants.

### Post-POC Decision

Decision: **defer**.

The targeted POC report
[`../evals/results/ockp-document-history-review-controls-poc.md`](../evals/results/ockp-document-history-review-controls-poc.md)
keeps document history and review controls as non-release-blocking reference
evidence. The scenarios showed that history inspection, restore and rollback
pressure, pending-change review pressure, stale synthesis after revision, and
final-answer-only validation pressure are expressible through existing runner
behavior while preserving citations, source refs, provenance, and projection
freshness evidence.

The only original POC failure was `document-diff-review-pressure`, classified
as skill guidance and eval coverage around vault-relative path use rather than
a runner capability gap. The focused follow-up report
[`../evals/results/ockp-document-diff-review-path-guidance.md`](../evals/results/ockp-document-diff-review-path-guidance.md)
resolved that failure by hardening guidance and verifier coverage for logical
vault-relative paths.

No public runner action, request or response schema, storage migration, storage
API, or public OpenClerk interface is promoted by this decision. Because the
outcome is defer and the only discovered follow-up gap was resolved by
`oc-d8w`, no implementation Bead or additional decision-created follow-up Bead
is required before closing `oc-da8`.

## Promotion Gates

- **Promote** only if repeated targeted AgentOps eval failures show the current
  document, retrieval, provenance, and projection freshness workflows cannot
  safely express needed history inspection, diff review, pending review,
  restore, rollback, or stale-derived-output workflows.
- **Defer** when failures are awkward but expressible workflows, missing skill
  guidance, data hygiene, thin evidence, missing eval coverage, or insufficient
  dogfooding pressure.
- **Kill** if the candidate duplicates Git or sync history, hides provenance or
  projection freshness, drops citations or source refs, enables hidden
  autonomous rewrites, makes review state invisible, weakens canonical markdown
  authority, or requires direct SQLite, vault inspection, HTTP/MCP,
  source-built runner paths, backend variants, module-cache inspection, or ad
  hoc runtime programs.

## Invariants

- AgentOps remains the production agent surface.
- Canonical markdown and promoted records remain source authority.
- Source-sensitive claims retain citations, source refs, or stable source
  identifiers.
- Provenance and projection freshness remain inspectable through runner-visible
  evidence.
- Public artifacts use repo-relative paths or neutral placeholders and must not
  expose raw private document diffs.
- New public runner actions require separate targeted evidence, compatibility
  review, and an implementation Bead.
