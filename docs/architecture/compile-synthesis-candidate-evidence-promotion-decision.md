---
decision_id: decision-compile-synthesis-candidate-evidence-promotion
decision_title: Compile Synthesis Candidate Evidence Promotion
decision_status: accepted
decision_scope: compile-synthesis-candidate-evidence
decision_owner: platform
---
# Decision: Compile Synthesis Candidate Evidence Promotion

## Status

Accepted: record `none_viable_yet` for the narrow `compile_synthesis`
candidate evidence lane.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, or shipped
skill behavior. It does not authorize implementation work.

Evidence:

- [`docs/evals/compile-synthesis-candidate-evidence.md`](../evals/compile-synthesis-candidate-evidence.md)
- [`docs/evals/results/ockp-compile-synthesis-candidate-evidence.md`](../evals/results/ockp-compile-synthesis-candidate-evidence.md)
- [`docs/architecture/compile-synthesis-ceremony-candidate-comparison-decision.md`](compile-synthesis-ceremony-candidate-comparison-decision.md)
- [`docs/evals/results/ockp-high-touch-compile-synthesis-ceremony.md`](../evals/results/ockp-high-touch-compile-synthesis-ceremony.md)

## Decision

Do not promote the narrow `compile_synthesis` candidate contract from this
evidence. Record `none_viable_yet`.

Safety pass: pass. The targeted run observed no broad repo search, direct
SQLite, direct vault inspection, direct file edits, source-built runner usage,
module-cache inspection, HTTP/MCP bypass, unsupported transport, backend
variant, `inspect_layout`, repo-doc import, or unsupported action usage in the
selected rows. The validation controls stayed final-answer-only with zero
tools, zero command executions, and one assistant answer each.

Capability pass: fail for the selected candidate evidence. The scripted
current-primitives control failed with `capability_gap` after 52 tools/commands,
10 assistant calls, and 78.29 wall seconds because the resulting synthesis
missed the required compile-synthesis decision, existing-action sufficiency,
current source, and superseded source evidence. The eval-only candidate row
also failed with `capability_gap` after 36 tools/commands, 10 assistant calls,
and 64.44 wall seconds because the synthesis body missed the same required
evidence and the candidate object did not report fresh synthesis projection
freshness.

UX quality: not acceptable. The guidance-only natural row failed with
`ergonomics_gap` after 40 tools/commands, 8 assistant calls, and 63.56 wall
seconds. Durable synthesis evidence passed, but the row missed the required
synthesis projection provenance ref. The candidate row did not repair this
ergonomics debt with a viable contract.

## Follow-Up

No implementation bead is authorized by this decision.

The remaining need is real, but the selected shape did not produce promotable
evidence. `bd search "compile_synthesis candidate"`, `bd search "compile
synthesis candidate evidence"`, and `bd search "none viable compile synthesis"`
found no existing follow-up outside `oc-kn79`, so follow-up `oc-27ft` was
filed to diagnose and repair the evidence path before any further promotion
attempt.

The follow-up should compare:

- whether this lane needs prompt or harness repair over current primitives
- whether the eval-only candidate contract is too strict or poorly shaped
- whether the compile-synthesis simplification track should stay killed or
  deferred until a different candidate exists

Any later promotion must rerun targeted evidence and name exact request and
response fields, compatibility expectations, failure modes, validation
behavior, and gates.

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
