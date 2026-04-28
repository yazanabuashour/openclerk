---
decision_id: decision-document-lifecycle-promotion
decision_title: Document Lifecycle Promotion
decision_status: accepted
decision_scope: document-lifecycle
decision_owner: platform
---
# Decision: Document Lifecycle Promotion

## Status

Accepted: defer document lifecycle promotion and keep the refreshed pressure
lane as reference evidence.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, or shipped skill behavior.

Evidence:

- [`document-history-review-controls-adr.md`](document-history-review-controls-adr.md)
- [`../evals/document-history-review-controls-poc.md`](../evals/document-history-review-controls-poc.md)
- [`../evals/results/ockp-document-lifecycle-pressure.md`](../evals/results/ockp-document-lifecycle-pressure.md)
- [`../evals/results/ockp-document-history-review-controls-poc.md`](../evals/results/ockp-document-history-review-controls-poc.md)
- [`../evals/results/ockp-document-diff-review-path-guidance.md`](../evals/results/ockp-document-diff-review-path-guidance.md)

## Decision

Do not promote semantic document history, semantic diff, pending-review queue,
restore/rollback, stale-derived-state, private artifact lifecycle, storage
migration, or new public runner/API surface from this evidence.

The current promoted public surface remains:

- `openclerk document`
- `openclerk retrieval`

Capability path: no promotion. The refreshed targeted lane completed scripted
controls for history inspection, semantic diff review, restore/rollback, stale
synthesis inspection, and validation/bypass handling. Current primitives can
express those workflows safely when the task is explicit, while preserving
canonical authority, citations/source refs, provenance, projection freshness,
privacy boundaries, local-first operation, and no-bypass rules.

Ergonomics path: defer for repair. The natural lifecycle intent row failed as
`ergonomics_gap` with 48 tools/commands, 10 assistant calls, and 73.61 wall
seconds. The pending-review row failed as `skill_guidance` even though
runner-visible evidence existed. The stale synthesis scripted control also
remained high-touch at 52 tools/commands, 14 assistant calls, and 175.68 wall
seconds. That evidence shows real workflow pressure, but not repeated repaired
natural-intent failures sufficient to promote a public lifecycle surface.

## Follow-Up Policy

No implementation follow-up for document lifecycle controls is authorized by
this decision. A future follow-up may repair skill guidance or eval coverage
for natural lifecycle rollback and pending-review answer quality, then rerun
the targeted lane.

A future promotion issue may be opened only after refreshed evidence shows one
of these conditions:

- repeated `capability_gap` or `runner_capability_gap` failures where current
  primitives cannot preserve authority, citations/source refs, provenance,
  freshness, privacy, local-first operation, operator visibility, and bypass
  prevention
- repeated `ergonomics_gap` failures under natural intent after validation
  controls stay final-answer-only and scripted controls continue to pass
- a proposed lifecycle request and response contract that exposes source
  evidence, before/after references or hashes, review state, restore reason,
  provenance, projection freshness, stale-derived-state impact, and private
  artifact handling rather than hiding them behind a write result

Until then, document lifecycle maintenance remains an AgentOps runner workflow
using `search`, `list_documents`, `get_document`, `create_document` for review
notes, `replace_section`, `append_document`, `provenance_events`, and
`projection_states`.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` are still the public routine
  document lifecycle surfaces.
- Markdown remains canonical authority; storage-level Git or sync history
  remains outside OpenClerk semantic lifecycle state.
- Source-sensitive lifecycle answers must preserve citations/source refs,
  provenance, freshness, and no-bypass invariants.
- Public reports must use repo-relative paths or neutral placeholders such as
  `<run-root>` and must not expose raw private diffs or private artifact bodies.
- Missing-field and invalid-request handling must continue to preserve the
  final-answer-only validation contract.
