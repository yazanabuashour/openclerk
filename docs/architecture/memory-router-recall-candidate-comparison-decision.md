---
decision_id: decision-memory-router-recall-candidate-comparison
decision_title: Memory/Router Recall Candidate Comparison
decision_status: accepted
decision_scope: memory-router-recall
decision_owner: platform
---
# Decision: Memory/Router Recall Candidate Comparison

## Status

Accepted: select a future narrow memory/router recall helper or report
candidate for targeted promotion evidence.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, memory
transport, remember/recall action, autonomous router API, or shipped skill
behavior. It does not authorize implementation work.

Evidence:

- [`docs/evals/memory-router-recall-candidate-comparison-poc.md`](../evals/memory-router-recall-candidate-comparison-poc.md)
- [`docs/architecture/memory-router-recall-ceremony-promotion-decision.md`](memory-router-recall-ceremony-promotion-decision.md)
- [`docs/evals/results/ockp-high-touch-memory-router-recall-ceremony.md`](../evals/results/ockp-high-touch-memory-router-recall-ceremony.md)
- [`docs/architecture/memory-router-revisit-promotion-decision.md`](memory-router-revisit-promotion-decision.md)
- [`docs/evals/results/ockp-memory-router-revisit-pressure.md`](../evals/results/ockp-memory-router-revisit-pressure.md)

## Decision

Select the candidate shape: a future read-only memory/router recall helper or
report surface that exposes temporal status, canonical evidence refs,
stale/session status, source refs or citations, provenance refs, synthesis
freshness, advisory feedback weighting, routing rationale, validation
boundaries, and authority limits. Do not implement the candidate yet.

Rejected alternatives:

- Guidance-only repair is too weak as the next step because the `oc-nu12`
  natural row preserved safety and capability but failed with an
  `ergonomics_gap` after 32 tools/commands, 5 assistant calls, and 50.35 wall
  seconds.
- No new surface is premature because the memory/router recall need remains
  real, the scripted row was still answer-shape fragile after 34
  tools/commands and 8 assistant calls, and normal users would reasonably
  expect one safe recall/routing answer without a manual evidence ceremony.

## Safety, Capability, UX

Safety pass: pass. Existing evidence preserves canonical markdown authority,
current canonical docs over stale session observations, source refs or
citations, provenance, synthesis projection freshness, advisory feedback
weighting, local-first runner-only access, validation controls, no-bypass
boundaries, and no durable-write shortcut. The selected candidate must not add
memory transports, remember/recall actions, autonomous router APIs, vector
stores, embedding stores, graph memory, hidden authority ranking, direct
SQLite, direct vault inspection, source-built runners, HTTP/MCP bypasses, or
unsupported transports.

Capability pass: pass for current primitives. The `oc-nu12` evidence did not
show a `capability_gap`; current `openclerk document` plus `openclerk
retrieval` primitives can express the workflow safely when the agent completes
the required evidence and answer steps.

UX quality: not acceptable enough to stop at reference pressure. The `oc-nu12`
natural row failed with `ergonomics_gap`, and the scripted row failed with
`skill_guidance_or_eval_coverage` after missing required synthesis-document
evidence. Prior revisit evidence completed safely, but still required 26
tools/commands and 66.91 wall seconds for the natural row.

## Follow-Up

File one follow-up Bead for targeted eval and promotion evidence for the
selected narrow memory/router recall candidate. Do not file an implementation
Bead.

Follow-up `oc-fnhj` must compare the selected candidate against current
primitives and guidance-only repair, then either promote an exact
request/response contract, defer, kill, or record `none viable yet`. Any later
promotion decision must name the exact response fields, compatibility
expectations, failure modes, validation behavior, authority limits, and gates.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public
  memory/router evidence surfaces.
- Canonical markdown remains durable memory and routing authority.
- Current canonical docs outrank stale session observations.
- Feedback weighting remains advisory.
- Routing rationale must stay grounded in existing document/retrieval evidence.
- Memory/router candidate work remains read-only unless a later promotion
  decision explicitly authorizes durable write behavior.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
