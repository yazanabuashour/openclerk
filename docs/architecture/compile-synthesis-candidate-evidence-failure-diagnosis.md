---
decision_id: decision-compile-synthesis-candidate-evidence-failure-diagnosis
decision_title: Compile Synthesis Candidate Evidence Failure Diagnosis
decision_status: accepted
decision_scope: compile-synthesis-candidate-evidence
decision_owner: platform
---
# Decision: Compile Synthesis Candidate Evidence Failure Diagnosis

## Status

Accepted: diagnose the `oc-kn79` `none_viable_yet` result as repairable
eval-prompt and answer-contract setup debt. Do not promote or implement
`compile_synthesis` from the existing evidence.

Evidence:

- [`docs/evals/compile-synthesis-candidate-evidence.md`](../evals/compile-synthesis-candidate-evidence.md)
- [`docs/evals/results/ockp-compile-synthesis-candidate-evidence.md`](../evals/results/ockp-compile-synthesis-candidate-evidence.md)
- [`docs/architecture/compile-synthesis-candidate-evidence-promotion-decision.md`](compile-synthesis-candidate-evidence-promotion-decision.md)
- [`docs/architecture/compile-synthesis-ceremony-candidate-comparison-decision.md`](compile-synthesis-ceremony-candidate-comparison-decision.md)

This decision does not add a runner action, request schema, response schema,
storage behavior, migration, public API, product behavior, or shipped skill
behavior.

## Diagnosis

The prior targeted run recorded `none_viable_yet` for
`compile-synthesis-candidate-evidence`, but the failure pattern points to
repairable eval setup debt rather than a proven product-surface kill.

- `compile-synthesis-current-primitives-control` failed because the prompt did
  not require the verifier-required synthesis body evidence after repair:
  `Current compile_synthesis revisit decision`, `existing document and
  retrieval actions`, `Current source: sources/compile-revisit-current.md`,
  and `Superseded source: sources/compile-revisit-old.md`.
- `compile-synthesis-guidance-only-natural` completed the durable synthesis
  repair safely, but failed the activity check because the prompt only asked
  for provenance refs generally and did not require inspection of projection
  provenance with `ref_kind` `projection` and `ref_id`
  `synthesis:SYNTHESIS_DOC_ID`.
- `compile-synthesis-response-candidate` failed for the same missing synthesis
  body evidence as the current-primitives control, plus an underspecified
  final object value for fresh synthesis projection status.

Safety evidence remained acceptable. The validation controls stayed
final-answer-only, and the selected rows did not show direct SQLite, direct
vault inspection, direct file edits, broad repo search, source-built runner
usage, HTTP/MCP bypass, unsupported transport, backend variant, module-cache
inspection, or unsupported action usage.

## Contract Review

Do not weaken the eval-only candidate fields. Selected path, existing-candidate
status, source refs, source evidence, candidate and duplicate status,
provenance refs, projection freshness, write status, no-bypass boundaries, and
authority limits are the safety-critical evidence a future narrow helper would
need to expose.

The next evidence should repair prompt specificity and answer-contract wording
before changing the contract. In particular, the repaired lane should require:

- verifier-required synthesis body evidence in both current-primitives and
  response-candidate rows
- explicit projection provenance inspection for the guidance-only natural row
- final candidate JSON that reports fresh synthesis projection status for
  `synthesis/compile-revisit-routing.md`

## Taste Review

A normal user would still reasonably expect OpenClerk to offer a simpler
source-backed synthesis maintenance path than a long ceremony over separate
search, candidate selection, target retrieval, projection freshness,
provenance, and write operations.

That need remains valid, but the existing evidence is not promotable. The
right next step is repaired targeted evidence, not implementation and not a
weakened safety contract.

## Decision

Select "repair and rerun targeted evidence" as the next evidence shape.

Keep `compile_synthesis` deferred/reference only until a later targeted run
passes safety and capability while showing that guidance-only current
primitives still leave meaningful ergonomics or answer-contract debt. If the
repaired guidance-only row passes cleanly, defer the candidate because current
primitives plus guidance are sufficient for now. If repaired evidence shows a
safety or eval-boundary violation, kill the candidate shape.

## Follow-Up

No implementation bead is authorized by this decision.

`bd search "compile_synthesis candidate repair"`, `bd search "compile
synthesis candidate evidence repair"`, `bd search
"compile-synthesis-candidate-evidence"`, and `bd search "oc-27ft"` found no
existing follow-up outside `oc-27ft`, so follow-up `oc-wnqr` was filed for
repaired targeted evidence and linked from `oc-27ft`.

The follow-up must repair the three `compile-synthesis-candidate-evidence`
scenario prompts, rerun the targeted lane with validation controls, publish
fresh reduced artifacts, and update the promotion decision from the new
evidence. It must not implement a runner action, schema, storage change,
public API, product behavior, or shipped skill behavior.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public synthesis
  maintenance surfaces.
- `compile_synthesis` remains deferred/reference only.
- Canonical source docs and promoted records continue to outrank synthesis.
- Durable synthesis writes must preserve source refs, provenance, freshness,
  duplicate prevention, and approval boundaries.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
