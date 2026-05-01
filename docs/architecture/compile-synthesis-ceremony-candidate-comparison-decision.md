---
decision_id: decision-compile-synthesis-ceremony-candidate-comparison
decision_title: Compile Synthesis Ceremony Candidate Comparison
decision_status: accepted
decision_scope: compile-synthesis-ceremony
decision_owner: platform
---
# Decision: Compile Synthesis Ceremony Candidate Comparison

## Status

Accepted: select a future narrow `compile_synthesis` candidate for targeted
promotion evidence.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, or shipped
skill behavior. It does not authorize implementation work.

Evidence:

- [`docs/evals/compile-synthesis-ceremony-candidate-comparison-poc.md`](../evals/compile-synthesis-ceremony-candidate-comparison-poc.md)
- [`docs/architecture/compile-synthesis-ceremony-promotion-decision.md`](compile-synthesis-ceremony-promotion-decision.md)
- [`docs/evals/results/ockp-high-touch-compile-synthesis-ceremony.md`](../evals/results/ockp-high-touch-compile-synthesis-ceremony.md)
- [`docs/architecture/synthesis-compile-revisit-promotion-decision.md`](synthesis-compile-revisit-promotion-decision.md)

## Decision

Select the candidate shape: a future narrow `compile_synthesis` helper or
report surface that exposes candidate selection, source refs, provenance,
projection freshness, duplicate behavior, update mode, and write status. Do
not implement the candidate yet.

The selected future candidate should evaluate the existing deferred request
shape:

```json
{
  "action": "compile_synthesis",
  "synthesis": {
    "path": "synthesis/example.md",
    "title": "Example",
    "source_refs": ["sources/source-a.md", "sources/source-b.md"],
    "body": "# Example\n\n## Summary\n...\n\n## Sources\n...\n\n## Freshness\n...",
    "mode": "create_or_update"
  }
}
```

The future response candidate should expose selected path, existing-candidate
or duplicate status, source evidence, single-line `source_refs`, provenance
refs, projection freshness, write status, and no-bypass boundaries.

Rejected alternatives:

- Guidance-only repair is too weak as the next step because the `oc-7feg`
  natural row completed safely but still required 42 tools/commands for a
  routine synthesis maintenance request.
- No new surface is premature because the synthesis maintenance need remains
  real and normal users would reasonably expect OpenClerk to preserve source
  refs, sources, freshness, and duplicate prevention without a long ceremony.

## Safety, Capability, UX

Safety pass: pass. Existing evidence preserved source authority, single-line
`source_refs`, `## Sources`, `## Freshness`, provenance/freshness inspection,
duplicate prevention, local-first runner-only access, validation controls, and
no-bypass boundaries. The selected candidate must keep direct SQLite, direct
vault inspection, direct file edits, broad repo search, source-built runners,
HTTP/MCP bypasses, unsupported transports, backend variants, module-cache
inspection, unsupported write actions, and authority escalation outside the
routine workflow.

Capability pass: pass for current primitives. The `oc-7feg` scripted control
completed with classification `none`, and current `openclerk document` plus
`openclerk retrieval` primitives can express the workflow safely.

UX quality: not acceptable enough to stop at reference pressure. The
`oc-7feg` natural row completed with classification `none`, but it required 42
tools/commands, 8 assistant calls, and 53.08 wall seconds. Prior repaired
synthesis revisit evidence also completed safely while remaining high-touch at
34 tools/commands, 12 assistant calls, and 105.24 wall seconds.

## Follow-Up

File one follow-up Bead for targeted eval and promotion evidence for the
selected narrow `compile_synthesis` candidate. Do not file an implementation
Bead.

The follow-up must compare the selected candidate against current primitives
and guidance-only repair, then either promote an exact request/response
contract, defer, kill, or record `none viable yet`. Any later promotion
decision must name the exact response fields, compatibility expectations,
failure modes, validation behavior, and gates.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public synthesis
  maintenance surfaces.
- `compile_synthesis` remains deferred/reference only.
- Canonical markdown source docs and promoted records outrank synthesis.
- Dependent synthesis writes remain durable writes that must preserve source
  refs, provenance, freshness, duplicate prevention, and approval boundaries.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
