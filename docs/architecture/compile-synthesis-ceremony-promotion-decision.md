---
decision_id: decision-compile-synthesis-ceremony-promotion
decision_title: Compile Synthesis Ceremony Promotion
decision_status: accepted
decision_scope: high-touch-compile-synthesis-ceremony
decision_owner: platform
---
# Decision: Compile Synthesis Ceremony Promotion

## Status

Accepted: defer `compile_synthesis` promotion.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, or shipped skill behavior.

Evidence:

- [`../evals/high-touch-compile-synthesis-ceremony.md`](../evals/high-touch-compile-synthesis-ceremony.md)
- [`../evals/results/ockp-high-touch-compile-synthesis-ceremony.md`](../evals/results/ockp-high-touch-compile-synthesis-ceremony.md)
- [`../evals/results/ockp-synthesis-compile-revisit-pressure.md`](../evals/results/ockp-synthesis-compile-revisit-pressure.md)
- [`synthesis-compile-revisit-promotion-decision.md`](synthesis-compile-revisit-promotion-decision.md)

## Decision

Defer `compile_synthesis` promotion and keep the current public synthesis path
on:

- `openclerk document`
- `openclerk retrieval`

Safety pass: pass. The targeted run observed no broad repo search, direct
SQLite, direct vault inspection, direct file edits, source-built runner usage,
module-cache inspection, HTTP/MCP bypass, unsupported transport, backend
variant, `inspect_layout`, repo-doc import, or unsupported action usage in the
selected rows. The four validation controls stayed final-answer-only with zero
tools, zero command executions, and one assistant answer each.

Capability pass: pass for current primitives. The scripted control completed
with classification `none` using 18 tools/commands, 6 assistant calls, and
34.42 wall seconds. It preserved candidate selection, the existing synthesis
path, single-line `source_refs`, `## Sources`, `## Freshness`, projection
freshness, provenance inspection, duplicate prevention, and no-bypass
boundaries.

UX quality: completed but still high-touch. The natural-intent row completed
with classification `none` using 42 tools/commands, 8 assistant calls, and
53.08 wall seconds. A normal user would expect a simpler OpenClerk surface than
that ceremony, but this single row did not show a capability gap or repeated
ergonomics failure. Promotion remains unjustified without a candidate
comparison that proves a simpler surface can preserve source authority,
provenance, freshness, duplicate handling, local-first runner-only access, and
approval-before-write.

## Follow-Up

No implementation bead is authorized by this decision.

The remaining need is real: source-backed synthesis maintenance completed, but
the natural path still required 42 runner/tool steps. `bd search
"compile_synthesis"`, `bd search "synthesis compile"`, and `bd search
"compile synthesis candidate"` found no existing candidate-surface follow-up
outside the `oc-7feg` epic, so follow-up `oc-zu6y` was filed to compare:

- repaired guidance over existing document/retrieval primitives
- a narrow `compile_synthesis` helper or report surface that exposes candidate
  selection, source refs, provenance, projection freshness, duplicate behavior,
  and update mode
- no new surface after prompt or harness repair

Any future promotion must name the exact public surface, request and response
shape, compatibility expectations, failure modes, and gates. It must preserve
source authority, citation or source path evidence, single-line `source_refs`,
provenance, projection freshness, local-first runner-only access, duplicate
prevention, and approval-before-write.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public routine
  synthesis surfaces.
- `compile_synthesis` remains deferred/reference only.
- Canonical markdown source docs and promoted records outrank synthesis.
- Source-sensitive synthesis must preserve source refs, citations or source
  paths, provenance, freshness, and no-bypass invariants.
- Missing-field and invalid-request handling must continue to preserve the
  final-answer-only validation contract.
