---
decision_id: decision-synthesis-compile-revisit-promotion
decision_title: Synthesis Compile Revisit Promotion
decision_status: accepted
decision_scope: synthesis-compile-revisit
decision_owner: platform
---
# Decision: Synthesis Compile Revisit Promotion

## Status

Accepted: defer `compile_synthesis` promotion.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, or shipped skill behavior.

Evidence:

- [`synthesis-compile-revisit-adr.md`](synthesis-compile-revisit-adr.md)
- [`../evals/synthesis-compile-revisit-comparison-poc.md`](../evals/synthesis-compile-revisit-comparison-poc.md)
- [`../evals/results/ockp-synthesis-compile-revisit-pressure.md`](../evals/results/ockp-synthesis-compile-revisit-pressure.md)
- [`../evals/results/ockp-synthesis-compiler-pressure.md`](../evals/results/ockp-synthesis-compiler-pressure.md)
- [`../evals/results/ockp-synthesis-maintenance-ergonomics.md`](../evals/results/ockp-synthesis-maintenance-ergonomics.md)

## Decision

Defer `compile_synthesis` promotion and keep the existing `openclerk document`
and `openclerk retrieval` runner workflow as the public synthesis path.

Capability path: no promotion. The scripted control in
`docs/evals/results/ockp-synthesis-compile-revisit-pressure.md` completed
through existing runner primitives while preserving canonical source authority,
single-line `source_refs`, `## Sources`, `## Freshness`, projection freshness,
provenance inspection, duplicate prevention, and no-bypass boundaries. Current
primitives can express the workflow safely when the task is explicit.

Ergonomics path: defer. The natural-intent row classified as
`ergonomics_gap`, with 16 commands, 9 assistant calls, and missing required
current/superseded source status lines in the repaired synthesis. The scripted
control also required 48 commands and 8 assistant calls. This is real
ergonomics pressure, but the same reduced run also reported a
`skill_guidance_or_eval_coverage` failure for `negative-limit-reject`, so the
evidence lane is not clean enough to promote a new public runner surface.

## Follow-Up Policy

No implementation follow-up for `compile_synthesis` is authorized by this
decision. Guidance/eval repair is tracked as `oc-4qlx`; it is not an
implementation authorization. A future promotion issue may be opened only after
repaired targeted evidence shows one of these conditions:

- repeated `capability_gap` failures where the current primitives cannot
  preserve authority, citations, provenance, freshness, and duplicate
  prevention
- repeated `ergonomics_gap` failures under natural intent after validation
  controls stay final-answer-only and scripted controls continue to pass
- a proposed `compile_synthesis` request and response contract that exposes
  source evidence, candidate selection, provenance, projection freshness,
  update mode, duplicate behavior, and failure classification rather than
  hiding them behind a write result

Until then, synthesis maintenance remains an AgentOps runner workflow using
`search`, `list_documents`, `get_document`, `projection_states`,
`provenance_events`, `replace_section`, and `append_document`.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` are still the public routine
  synthesis surfaces.
- `compile_synthesis` remains a deferred reference shape only.
- Canonical markdown source docs and promoted records outrank synthesis.
- Source-sensitive synthesis must preserve source refs, citations or source
  paths, provenance, freshness, and no-bypass invariants.
- Missing-field and invalid-request handling must continue to preserve the
  final-answer-only validation contract.
