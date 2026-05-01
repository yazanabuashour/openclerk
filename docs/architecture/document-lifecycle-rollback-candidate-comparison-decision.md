---
decision_id: decision-document-lifecycle-rollback-candidate-comparison
decision_title: Document Lifecycle Rollback Candidate Comparison
decision_status: accepted
decision_scope: document-lifecycle-rollback
decision_owner: platform
---
# Decision: Document Lifecycle Rollback Candidate Comparison

## Status

Accepted: select a future narrow lifecycle review/rollback candidate for
targeted promotion evidence.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, or shipped
skill behavior. It does not authorize implementation work.

Evidence:

- [`docs/evals/document-lifecycle-rollback-candidate-comparison-poc.md`](../evals/document-lifecycle-rollback-candidate-comparison-poc.md)
- [`docs/architecture/document-lifecycle-ceremony-promotion-decision.md`](document-lifecycle-ceremony-promotion-decision.md)
- [`docs/evals/results/ockp-high-touch-document-lifecycle-ceremony.md`](../evals/results/ockp-high-touch-document-lifecycle-ceremony.md)
- [`docs/architecture/document-lifecycle-promotion-decision.md`](document-lifecycle-promotion-decision.md)

## Decision

Select the candidate shape: a future narrow lifecycle review/rollback helper
or report surface that exposes target document, source evidence, before/after
summary, restore reason, provenance refs, projection freshness,
privacy/no-diff boundaries, no-bypass boundaries, and write status. Do not
implement the candidate yet.

The selected future candidate should evaluate an evidence-visible request shape
such as:

```json
{
  "action": "review_lifecycle_rollback",
  "lifecycle": {
    "target_path": "notes/history-review/restore-target.md",
    "source_refs": ["sources/history-review/restore-authority.md"],
    "restore_section": "Summary",
    "restore_reason": "unsafe accepted lifecycle summary",
    "mode": "review_then_restore"
  }
}
```

The future response candidate should expose target identity, source evidence,
before/after summary without raw private diffs, restore reason, provenance
refs, projection freshness, privacy/no-diff and no-bypass boundaries, and
write status.

Rejected alternatives:

- Guidance-only repair is too weak as the next step because the `oc-k8ba`
  natural row completed safely but still required 40 tools/commands and 10
  assistant calls for routine lifecycle rollback.
- No new surface is premature because the lifecycle rollback need remains real
  and normal users would reasonably expect OpenClerk to review and restore an
  unsafe accepted summary without a long ceremony.

## Safety, Capability, UX

Safety pass: pass. Existing evidence preserved canonical markdown authority,
source refs or citations, provenance, projection freshness, rollback target
accuracy, privacy-safe summaries, no raw private diff leakage, local-first
runner-only access, validation controls, and no-bypass boundaries. The
selected candidate must keep direct SQLite, direct vault inspection, direct
file edits, broad repo search, source-built runners, HTTP/MCP bypasses,
unsupported transports, backend variants, module-cache inspection, raw private
diff leakage, storage-root path leakage, and automatic authority escalation
outside the routine workflow.

Capability pass: pass for current primitives. The `oc-k8ba` scripted control
completed with classification `none`, and current `openclerk document` plus
`openclerk retrieval` primitives can express lifecycle review and rollback
safely.

UX quality: not acceptable enough to stop at reference pressure. The
`oc-k8ba` natural row completed with classification `none`, but it required 40
tools/commands, 10 assistant calls, and 42.29 wall seconds. The scripted
control still required 22 tools/commands and 7 assistant calls.

## Follow-Up

Follow-up `oc-cez4` was filed for targeted eval and promotion evidence for the
selected narrow lifecycle review/rollback candidate. Do not file an
implementation Bead from this comparison.

The follow-up must compare the selected candidate against current primitives
and guidance-only repair, then either promote an exact request/response
contract, defer, kill, or record `none viable yet`. Any later promotion
decision must name the exact response fields, compatibility expectations,
failure modes, validation behavior, and gates.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public document
  lifecycle surfaces.
- Markdown remains canonical authority; storage-level Git or sync history
  remains outside OpenClerk semantic lifecycle state.
- Source-sensitive lifecycle answers must preserve citations/source refs,
  provenance, freshness, rollback target accuracy, privacy boundaries, and
  no-bypass invariants.
- Raw private diffs, private artifact bodies, and machine-absolute paths remain
  forbidden in committed evidence.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
