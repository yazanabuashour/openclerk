---
decision_id: decision-compile-synthesis-repeated-ergonomics-checkpoint
decision_title: Compile Synthesis Repeated Ergonomics Checkpoint
decision_status: accepted
decision_scope: compile-synthesis-candidate-evidence
decision_owner: platform
---
# Decision: Compile Synthesis Repeated Ergonomics Checkpoint

## Status

Accepted: close the conditional repeated-ergonomics evidence follow-up with no
new eval run and no implementation work.

This checkpoint does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, shipped
skill behavior, or eval harness behavior.

Evidence:

- [`docs/architecture/compile-synthesis-candidate-evidence-promotion-decision.md`](compile-synthesis-candidate-evidence-promotion-decision.md)
- [`docs/evals/results/ockp-compile-synthesis-candidate-evidence.md`](../evals/results/ockp-compile-synthesis-candidate-evidence.md)
- [`docs/architecture/compile-synthesis-candidate-evidence-failure-diagnosis.md`](compile-synthesis-candidate-evidence-failure-diagnosis.md)

## Checkpoint

`oc-ghcs` was conditional on a recurring natural-intent ergonomics or
answer-contract debt signal after the repaired `oc-wnqr` evidence. No such
new trigger is present.

The latest accepted result remains
`defer_guidance_only_current_primitives_sufficient`: current primitives,
guidance-only natural intent, and the eval-only response candidate all passed
the repaired targeted lane. Guidance-only current primitives were sufficient
for that pressure, so the candidate contract remains deferred rather than
promoted.

Beads searches before closure, using `--status all` where closed follow-ups
must remain visible, found no newer trigger beyond `oc-ghcs` itself:

- `bd search "compile synthesis ergonomics" --status all` returned only
  `oc-ghcs`
- `bd search "compile_synthesis ergonomics" --status all` returned only
  `oc-ghcs`
- `bd search "compile synthesis natural intent"` returned no issues
- `bd search "compile_synthesis repeated ergonomics"` returned no issues
- `bd search "compile_synthesis candidate promotion evidence"` returned no
  issues

## Decision

Do not run another targeted eval from this checkpoint. Do not file an
implementation bead. Do not promote `compile_synthesis`.

The valid future trigger remains the one already recorded in the promotion
decision: stronger repeated evidence that natural guidance over current
`openclerk document` and `openclerk retrieval` primitives leaves meaningful
ergonomics or answer-contract debt while a candidate contract preserves
safety and capability.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public synthesis
  maintenance surfaces.
- `compile_synthesis` remains deferred/reference only.
- Canonical markdown source docs and promoted records outrank synthesis.
- Durable synthesis writes must preserve source refs, provenance, freshness,
  duplicate prevention, and approval boundaries.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
