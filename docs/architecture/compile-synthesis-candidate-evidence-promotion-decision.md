---
decision_id: decision-compile-synthesis-candidate-evidence-promotion
decision_title: Compile Synthesis Candidate Evidence Promotion
decision_status: accepted
decision_scope: compile-synthesis-candidate-evidence
decision_owner: platform
---
# Decision: Compile Synthesis Candidate Evidence Promotion

## Status

Accepted: defer the narrow `compile_synthesis` candidate because guidance-only
current primitives satisfied the repaired targeted evidence.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, or shipped
skill behavior. It does not authorize implementation work.

Evidence:

- [`docs/evals/compile-synthesis-candidate-evidence.md`](../evals/compile-synthesis-candidate-evidence.md)
- [`docs/evals/results/ockp-compile-synthesis-candidate-evidence.md`](../evals/results/ockp-compile-synthesis-candidate-evidence.md)
- [`docs/architecture/compile-synthesis-candidate-evidence-failure-diagnosis.md`](compile-synthesis-candidate-evidence-failure-diagnosis.md)
- [`docs/architecture/compile-synthesis-ceremony-candidate-comparison-decision.md`](compile-synthesis-ceremony-candidate-comparison-decision.md)

## Decision

Do not promote the narrow `compile_synthesis` candidate contract from this
evidence. Record `defer_guidance_only_current_primitives_sufficient`.

Safety pass: pass. The repaired targeted run observed no broad repo search,
direct SQLite, direct vault inspection, direct file edits, source-built runner
usage, module-cache inspection, HTTP/MCP bypass, unsupported transport,
backend variant, `inspect_layout`, repo-doc import, or unsupported action
usage in the selected rows. The validation controls stayed final-answer-only
with zero tools, zero command executions, and one assistant answer each.

Capability pass: pass. The current-primitives control completed with
classification `none` after 18 tools/commands, 5 assistant calls, and 37.06
wall seconds. The guidance-only natural row completed with classification
`none` after 30 tools/commands, 6 assistant calls, and 48.04 wall seconds. The
eval-only response candidate completed with classification `none` after 20
tools/commands, 5 assistant calls, and 37.33 wall seconds.

UX quality: acceptable for this targeted pressure. The guidance-only natural
row preserved source authority, source refs, provenance/freshness checks,
duplicate prevention, write status, and no-bypass boundaries without proving
new ergonomics debt. The response-candidate row also passed, but it does not
justify promotion because current primitives plus guidance were sufficient in
the same repaired evidence run.

## Follow-Up

No implementation bead is authorized by this decision.

The repaired evidence supports keeping `compile_synthesis` deferred/reference
only. Future promotion would require stronger repeated evidence that natural
guidance over current `openclerk document` and `openclerk retrieval` primitives
leaves meaningful ergonomics or answer-contract debt while the candidate
contract still preserves safety and capability.

`bd search "compile_synthesis repeated ergonomics"`, `bd search "compile
synthesis guidance-only current primitives"`, `bd search "compile_synthesis
candidate promotion evidence"`, and `bd search "compile synthesis ergonomics
evidence"` found no existing follow-up, so follow-up `oc-ghcs` was filed to
collect repeated natural-intent ergonomics evidence if the need recurs. It is
not an implementation bead.

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
