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

Ergonomics path: defer. The repaired natural-intent row completed with
classification `none` while preserving the existing synthesis candidate,
single-line `source_refs`, current and superseded source status, source paths,
`## Sources`, `## Freshness`, projection freshness, and no-bypass boundaries.
It required 34 tool/command calls, 8 assistant calls, and 65.51 wall seconds,
so the current workflow is still high-touch, but the repaired evidence no
longer shows a repeated ergonomics failure that justifies a new public runner
surface. The scripted control also completed with classification `none` using
28 tool/command calls, 9 assistant calls, and 57.12 wall seconds.

Validation controls still require repair. The missing document path,
unsupported lower-level, and unsupported transport controls stayed
final-answer-only, but the uncoached negative-limit request still used tools
and classified as `skill_guidance_or_eval_coverage`.

## oc-4qlx Repair Addendum

`oc-4qlx` repaired the synthesis compile natural-intent evidence while leaving
negative-limit validation coverage as remaining guidance/eval repair work. It
does not promote or implement `compile_synthesis`.

Current primitives can express the workflow safely. The scripted-control row
continues to prove that `openclerk document` and `openclerk retrieval` can
discover the candidate, inspect projection/provenance/freshness evidence,
update the existing synthesis document, preserve source authority, and avoid
duplicates without bypassing the installed runner contract.

The current UX is acceptable enough to keep without promotion. The repaired
natural-intent row still shows non-trivial latency and call count, but it
completed from outcome-level guidance rather than a step-by-step runner script,
reported classification `none`, and preserved the same authority,
provenance, freshness, and duplicate-prevention invariants as the scripted
control.

`compile_synthesis` remains deferred/reference only. No implementation
follow-up is authorized by this decision, and validation repair remains outside
that implementation surface.

## Follow-Up Policy

No implementation follow-up for `compile_synthesis` is authorized by this
decision. `oc-4qlx` repaired the natural-intent evidence and reaffirmed the
defer/reference outcome; it is not an implementation authorization. A future
promotion issue may be opened only after repaired targeted evidence shows one
of these conditions:

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
