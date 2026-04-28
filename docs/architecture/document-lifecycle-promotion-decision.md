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
behavior, public API, or public OpenClerk interface. It keeps the minimized
OpenClerk skill guidance that stabilized the reference lane.

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

Ergonomics path: no promotion. The finalized trial-and-error repair reduced
`skills/openclerk/SKILL.md` from 372 to 312 lines while passing the refreshed
targeted lane. The natural lifecycle row completed with `none` failure
classification using 40 tools/commands, 6 assistant calls, and 76.40 wall
seconds. Scripted diff, restore, pending-review, stale-synthesis, inspection,
and validation rows also completed with `none`. That evidence shows current
primitives remain acceptable as reference workflows when the skill preserves
runner-call serialization and concise lifecycle rules.

## Follow-Up Policy

No implementation follow-up for document lifecycle controls is authorized by
this decision. Future follow-up may continue monitoring natural lifecycle
rollback ergonomics and may rerun the targeted lane after broader OpenClerk
skill changes.

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
