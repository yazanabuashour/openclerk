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
- document history and review controls
- new public runner actions

The default decision is to keep each capability as reference or deferred.
Promotion requires targeted AgentOps evidence through one of two paths:

- **Capability gap:** the existing `openclerk document` and
  `openclerk retrieval` actions cannot safely express the workflow.
- **Ergonomics gap:** the existing actions can technically express the
  workflow, but repeated evidence shows the workflow is too slow, too many
  steps, too scripted, too error-prone, or too guidance-dependent for routine
  AgentOps use.

## Shared Rubric

Use the same decision rubric for every deferred capability:

- **Promote via capability gap** when repeated targeted AgentOps eval failures
  show the existing document/retrieval workflow is structurally insufficient,
  not merely underspecified, missing data, or missing eval coverage.
- **Promote via ergonomics gap** when the current workflow is expressible but
  repeated targeted evidence shows unacceptable step count, latency, prompt
  brittleness, retry risk, or guidance dependence, and a proposed surface
  reduces that cost without weakening authority, citations, provenance,
  freshness, or local-first operation.
- **Defer** when current runner actions pass with acceptable ergonomics,
  failures are data hygiene, ordinary skill-guidance gaps, or eval coverage
  gaps, or the evidence is too narrow to justify a production surface. Treat
  skill guidance as ordinary only when clarifying existing runner usage is
  enough; repeated need for workflow-specific prompt choreography or skill
  intervention is ergonomics-gap evidence.
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

## Ergonomics Evidence

Every targeted POC and eval for a deferred capability should report both
technical expressibility and UX acceptability. The minimum ergonomics scorecard
is:

- tool or command count for the current workflow and candidate surface
- assistant calls and wall time
- amount of prompt specificity required to make the workflow pass
- whether a natural user-intent prompt passes without scripting every runner
  step
- retry or brittleness indicators, including duplicate creation, skipped
  freshness inspection, dropped citations, or wrong target selection
- failure classification separated into data hygiene, ordinary skill guidance,
  eval coverage, capability gap, ergonomics gap, or contract violation
- authority, provenance, freshness, privacy, and bypass risks introduced by
  any proposed surface

Ergonomics promotion is not a shortcut around safety. A smoother surface should
be killed or deferred if it creates a second truth system, hides provenance,
drops citations, hides freshness, or normalizes lower-level bypasses.

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

### Document History And Review Controls

Promotion requires evidence that semantic document lifecycle control is needed
for agent-authored durable edits beyond storage-level Git, sync, snapshots, or
backups. The candidate must explain what changed, why it changed, which source
refs or citations justified it, which before/after content hashes or references
identify the change, which actor or source produced it, what review state
applies, and what derived projections or synthesis pages became stale.

Defer the candidate if existing document and retrieval workflows can express
the scenario through registered documents, append or replace-section,
provenance events, projection states, and operator review with acceptable
ergonomics. Defer also when failures are ordinary missing skill guidance, data
hygiene, thin dogfooding evidence, or eval coverage gaps.

Kill the candidate if it duplicates Git as byte-level history, hides review
state, enables hidden autonomous rewrites, drops source refs or citations,
hides provenance or freshness, exposes private raw diffs in public artifacts,
or requires routine direct SQLite, direct vault inspection, source-built runner
paths, HTTP/MCP bypasses, backend variants, module-cache inspection, or ad hoc
runtime programs.

### New Public Runner Actions

Promotion requires repeated failures that show existing multi-step document and
retrieval workflows cannot express the needed behavior, or repeated ergonomics
evidence that those workflows are too costly for routine use despite being
technically expressible. Any proposed action must include an exact JSON request
shape, backward compatibility expectations, failure modes, and targeted eval
gates.

Kill or defer the candidate if the existing actions pass with acceptable
ergonomics, the pressure comes from ordinary missing skill guidance, or the
proposed action would create a second authority surface.

## Prompt And Eval Pattern

Future POCs for deferred capabilities must follow this pattern:

1. Start with a control prompt that solves the workflow using only
   `openclerk document` and `openclerk retrieval`.
2. Add at least one natural user-intent prompt and one scripted-control prompt.
   The natural prompt measures UX and brittleness; the scripted control proves
   whether current primitives can still work with exact instructions.
3. Require the agent to use runner JSON evidence and preserve citations,
   source refs, provenance, and freshness where relevant.
4. Record tool/command count, assistant calls, wall time, prompt specificity,
   and retry or brittleness signals.
5. Classify failures as data hygiene, ordinary skill guidance, eval coverage,
   capability gap, ergonomics gap, or contract violation.
6. Record targeted evidence under `docs/evals/results/` using repo-relative
   paths and `<run-root>` placeholders.
7. End with an explicit decision: promote, defer, kill, or keep as reference.
8. If promoted, file a separate implementation Bead that names the exact
   surface and gates.

This keeps capability pressure measurable without letting interesting reference
behavior become production scope by default.
