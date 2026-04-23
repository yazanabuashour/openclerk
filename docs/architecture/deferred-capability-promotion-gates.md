# Deferred Capability Promotion Gates

## Status

Accepted as the decision method for deferred OpenClerk capabilities.

This document defines how OpenClerk decides whether to promote capabilities
that are intentionally outside the v1 AgentOps runner slice. It is a gate
contract, not a new public API.

## Scope

These gates apply to:

- Mem0 or a memory API
- autonomous router behavior
- semantic graph layers as truth
- broad contradiction engines
- new public runner actions

The default decision is to keep each capability as reference or deferred.
Promotion requires targeted AgentOps evidence that the existing
`openclerk document` and `openclerk retrieval` actions are structurally
insufficient.

## Shared Rubric

Use the same decision rubric for every deferred capability:

- **Promote** only when repeated targeted AgentOps eval failures show the
  existing document/retrieval workflow is structurally insufficient, not merely
  awkward, underspecified, missing data, or missing skill guidance.
- **Defer** when current runner actions pass, failures are data hygiene, skill
  guidance, or eval coverage gaps, or the evidence is too narrow to justify a
  production surface.
- **Kill** when the capability mostly duplicates docs retrieval, weakens source
  authority, hides provenance or freshness, increases duplicate/conflicting
  truth, or encourages routine bypasses.
- **Keep as reference** when the capability is useful benchmark pressure but
  does not justify implementation.

No promoted implementation work should be filed from this document alone. A
separate follow-up Bead is allowed only after a targeted eval report and
decision note identify the exact promoted surface and its gates.

## Required Invariants

Every candidate must preserve the current AgentOps invariants:

- citations, source refs, or stable source identifiers remain attached to
  source-sensitive claims
- provenance and projection freshness stay inspectable
- canonical docs and promoted canonical records outrank synthesis, memory,
  graph state, and routing choices
- routine agents do not use direct SQLite, broad repo search, source-built
  runner paths, HTTP/MCP bypasses, backend variants, module-cache inspection,
  or ad hoc runtime programs
- invalid routine requests still preserve the OpenClerk no-tools contract:
  missing required fields clarify, while invalid limits and bypass requests
  reject

If a candidate cannot preserve these invariants, kill or defer it.

## Capability-Specific Proof Obligations

### Mem0 Or Memory API

Promotion requires evidence that repeated recall improves real workflows after
canonicalization. Memory must remain recall, not authority. The candidate must
expose source refs, promotion path, temporal status, and stale or superseded
state before memory-derived output is trusted.

Kill or defer the candidate if it introduces memory-first truth, hides stale
canonical evidence behind ranking, cannot cite canonical docs or records, or
requires routine agents to use memory transports outside AgentOps.

### Autonomous Router

Promotion requires evidence that routing improves correctness over explicit
runner-action choice while staying explainable and audited. The candidate must
show why each source was chosen and must not invent precedence rules separate
from canonical docs, promoted records, provenance, and freshness.

Kill or defer the candidate if it becomes a hidden classifier, performs opaque
multi-store fanout, silently promotes memory, or routes around the runner.

### Semantic Graph Layer

Promotion requires evidence that richer graph semantics beat search, markdown
links, backlinks, and existing `graph_neighborhood` for relationship-shaped
tasks. Canonical markdown must remain the semantic authority, and graph output
must preserve source refs plus projection freshness.

Kill or defer the candidate if semantic edges become independent truth, lack
source evidence, hide stale graph state, or behave like a more complicated way
to do docs retrieval.

### Broad Contradiction Engine

Promotion requires evidence that a broader contradiction workflow beats the
existing source-sensitive audit path without inventing unsupported semantic
truth. Current-source conflicts with no runner-visible authority must remain
unresolved and explainable rather than forced to a winner.

Kill or defer the candidate if it makes arbitrary semantic contradiction
claims, drops source paths, hides supersession/freshness evidence, or creates
unrepairable conflict state.

### New Public Runner Actions

Promotion requires repeated failures that show existing multi-step document and
retrieval workflows are structurally too many steps or cannot express the
needed behavior. Any proposed action must include an exact JSON request shape,
backward compatibility expectations, failure modes, and targeted eval gates.

Kill or defer the candidate if the existing actions pass, the pressure comes
from missing skill guidance, or the proposed action would create a second
authority surface.

## Prompt And Eval Pattern

Future POCs for deferred capabilities must follow this pattern:

1. Start with a control prompt that solves the workflow using only
   `openclerk document` and `openclerk retrieval`.
2. Add pressure prompts only for the specific suspected failure mode.
3. Require the agent to use runner JSON evidence and preserve citations,
   source refs, provenance, and freshness where relevant.
4. Classify failures as data hygiene, skill guidance, eval coverage, or runner
   capability gaps.
5. Record targeted evidence under `docs/evals/results/` using repo-relative
   paths and `<run-root>` placeholders.
6. End with an explicit decision: promote, defer, kill, or keep as reference.
7. If promoted, file a separate implementation Bead that names the exact
   surface and gates.

This keeps capability pressure measurable without letting interesting reference
behavior become production scope by default.
