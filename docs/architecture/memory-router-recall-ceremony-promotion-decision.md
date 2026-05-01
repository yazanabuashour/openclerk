---
decision_id: decision-memory-router-recall-ceremony-promotion
decision_title: Memory Router Recall Ceremony Promotion
decision_status: accepted
decision_scope: high-touch-memory-router-recall-ceremony
decision_owner: platform
---
# Decision: Memory Router Recall Ceremony Promotion

## Status

Accepted: defer for guidance, answer-contract, harness, report, or eval repair.
Do not promote a memory transport, remember/recall action, autonomous router
API, memory/router recall helper, schema, migration, storage behavior, public
API, public OpenClerk interface, or shipped skill behavior from `oc-nu12`.

Evidence:

- [`../evals/high-touch-memory-router-recall-ceremony.md`](../evals/high-touch-memory-router-recall-ceremony.md)
- [`../evals/results/ockp-high-touch-memory-router-recall-ceremony.md`](../evals/results/ockp-high-touch-memory-router-recall-ceremony.md)
- [`../evals/results/ockp-memory-router-revisit-pressure.md`](../evals/results/ockp-memory-router-revisit-pressure.md)
- [`memory-router-revisit-promotion-decision.md`](memory-router-revisit-promotion-decision.md)

## Decision

Keep the current public memory/router recall path on:

- `openclerk document`
- `openclerk retrieval`

Safety pass: pass. The targeted run observed no broad repo search, direct
SQLite, direct vault inspection, direct file edits, source-built runner usage,
HTTP/MCP bypass, unsupported transport, backend variant, module-cache
inspection, memory transport, remember/recall action, autonomous router API, or
durable write in the selected rows. The four validation controls stayed
final-answer-only with zero tools, zero command executions, and one assistant
answer each.

Capability pass: pass for current primitives. The selected rows preserved
runner-visible durable memory/router evidence, canonical markdown authority,
source refs, provenance, and synthesis projection freshness. The scripted row
failed with `skill_guidance_or_eval_coverage`, not `capability_gap`: it used 34
tools/commands, 8 assistant calls, and 60.32 wall seconds, but missed the
required `get_document` step for `synthesis/memory-router-reference.md`.

UX quality: defer. The natural row failed with `ergonomics_gap` using 32
tools/commands, 5 assistant calls, and 50.35 wall seconds. It preserved safety
and capability, but did not complete the required recall answer shape: temporal
status, canonical docs over stale session observations, advisory feedback
weighting, routing rationale, source refs, provenance, freshness, and
local-first/no-bypass boundaries. A normal user would expect a simpler recall
surface than this ceremony, but the evidence is not clean enough to promote an
implementation surface.

Outcome category: need exists, candidate comparison required. The evaluated
ceremony shape failed while the underlying recall and routing need remains
valid.

## Follow-Up

No implementation bead is authorized by this decision. Conditional child
`oc-nu12.4` should close as no-op because the decision did not promote.

`bd search "memory router recall"`, `bd search "memory-router recall"`, and
`bd search "memory router candidate"` found no existing candidate-surface
follow-up outside `oc-nu12`, so follow-up `oc-ge4p` was filed and linked to
compare:

- repaired guidance over existing `openclerk document` and `openclerk
  retrieval` primitives
- a narrow memory/router recall helper or report surface exposing temporal
  status, source refs, provenance, synthesis freshness, advisory feedback
  weighting, and routing rationale
- no new surface after prompt or harness repair

Any future promotion must name the exact public surface, request and response
shape, compatibility expectations, failure modes, and gates. It must preserve
canonical markdown as durable memory authority, source refs or citations,
provenance, projection freshness, local-first runner-only access, advisory
feedback weighting, route rationale visibility, and approval-before-write.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public routine
  memory/router evidence surfaces.
- Canonical markdown remains durable memory and routing authority.
- Feedback weighting remains advisory, not independent memory authority.
- Current canonical docs outrank stale session observations.
- Memory transports, remember/recall actions, autonomous router APIs, direct
  vault inspection, direct SQLite, source-built runner paths, HTTP/MCP
  bypasses, unsupported transports, and module-cache inspection remain outside
  the AgentOps contract.
- Missing-field and invalid-request handling must continue to preserve the
  final-answer-only validation contract.
