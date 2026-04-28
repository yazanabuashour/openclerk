---
decision_id: adr-memory-router-revisit
decision_title: Memory And Autonomous Router Revisit
decision_status: accepted
decision_scope: memory-router
decision_owner: platform
---
# ADR: Memory And Autonomous Router Revisit

## Status

Accepted as a deferred-capability revisit track.

This ADR frames memory and autonomous routing as evidence gathering only. It
does not add runner actions, schemas, storage behavior, migrations, skill
behavior, memory transports, autonomous router APIs, or public APIs.

## Context

OpenClerk already keeps memory and autonomous routing as reference/deferred
pressure in
[`knowledge-configuration-v1-adr.md`](knowledge-configuration-v1-adr.md) and
[`../evals/results/ockp-memory-router-reference-poc.md`](../evals/results/ockp-memory-router-reference-poc.md).
That reference showed useful session material can be promoted by writing
canonical markdown with source refs, while temporal policy, feedback weighting,
and routing choice remain grounded in existing AgentOps document and retrieval
actions.

The revisit asks whether that reference posture still holds under the
deferred-capability promotion rubric in
[`deferred-capability-promotion-gates.md`](deferred-capability-promotion-gates.md).
The track tests both:

- `capability_gap`: whether existing `openclerk document` and
  `openclerk retrieval` actions are structurally insufficient for temporal
  recall, session promotion, feedback weighting, and routing-choice workflows.
- `ergonomics_gap`: whether existing actions can express the workflow but are
  too slow, too scripted, too brittle, too guidance-dependent, or too costly
  for routine AgentOps use.

## Decision

Use targeted ADR, POC, eval, and decision artifacts before any implementation
work. The default outcome is to keep memory and autonomous routing as
reference/deferred unless targeted evidence proves a capability gap or repeated
ergonomics gap.

The current public surface remains:

- `openclerk document`
- `openclerk retrieval`

Canonical markdown remains the authority for durable memory-like facts,
temporal status, feedback interpretation, and routing rationale. Session
observations, feedback weights, search results, and synthesis are advisory or
derived evidence; they must not outrank current canonical documents, source
refs, provenance, or freshness.

## Options

| Option | Description | Promotion posture |
| --- | --- | --- |
| Keep current primitives | Use search, `list_documents`, `get_document`, source-linked synthesis, provenance events, and projection freshness over canonical markdown. | Default/reference if natural and scripted pressure pass with acceptable ergonomics. |
| Strengthen guidance/evals | Keep the runner unchanged while improving skill wording, prompts, or targeted verifier coverage. | Use when failures are ordinary guidance, answer-contract, or eval-coverage gaps. |
| Add a narrow memory/router surface | Add a promoted runner action that packages temporal recall, candidate source selection, routing rationale, citations/source refs, provenance, and freshness. | Consider only if scripted controls prove current primitives are insufficient, or repeated natural rows show unacceptable UX while scripted controls pass. |
| Add memory-first recall or hidden autonomous routing | Let remembered state or autonomous route selection become authority independent of canonical markdown. | Kill unless it can preserve canonical authority, citations, provenance, freshness, local-first operation, and no-bypass invariants. |

## Invariants

Any future promoted surface must preserve:

- AgentOps-only routine operation through installed runner JSON.
- Canonical markdown authority for durable memory-like claims and routing
  rationale.
- Citations, source refs, or stable source identifiers for source-sensitive
  claims.
- Inspectable provenance and projection freshness.
- Local-first operation.
- No broad repo search, direct vault inspection, direct SQLite, source-built
  runner paths, HTTP/MCP bypasses, backend variants, module-cache inspection,
  memory transports, remember/recall actions, autonomous router APIs, or ad
  hoc lower-level transports for routine tasks.
- Final-answer-only handling for invalid no-tools requests.

## Non-Goals

This ADR does not:

- define a promoted memory, recall, or autonomous router runner action
- add vector storage, embeddings, graph memory, memory transports, or router
  state
- add schemas, migrations, background jobs, indexes, parser pipelines, or
  public APIs
- make feedback weight, session observation, synthesis, search ranking, or
  route choice more authoritative than canonical markdown
- relax citation, source ref, provenance, freshness, duplicate-prevention, or
  validation requirements

## Promotion Gates

Promotion via `capability_gap` requires repeated scripted-control failures
showing current document/retrieval primitives cannot safely express
memory/router workflows while preserving authority, source refs, provenance,
freshness, and bypass boundaries.

Promotion via `ergonomics_gap` requires repeated natural-intent failures or
unacceptable UX cost where scripted controls still pass. Evidence must show
high step count, latency, prompt brittleness, retries, wrong route selection,
missing source refs, skipped provenance/freshness checks, or workflow-specific
guidance dependence that a proposed surface would reduce without weakening the
invariants above.

Defer when failures are guidance, answer-contract, eval coverage, data hygiene,
partial evidence, one-off ergonomics pressure, or insufficient scripted-control
evidence.

Keep as reference when current primitives pass with acceptable ergonomics and
the proposed surface would mostly repackage document/retrieval workflows.

Kill the candidate if it makes memory first-class authority, hides citations
or freshness, hides provenance, silently routes across sources, creates a
second truth system, or encourages bypasses.

## Evidence Plan

The POC comparison is
[`../evals/memory-router-revisit-comparison-poc.md`](../evals/memory-router-revisit-comparison-poc.md).
The targeted reduced report is
[`../evals/results/ockp-memory-router-revisit-pressure.md`](../evals/results/ockp-memory-router-revisit-pressure.md).
The final promotion decision is
[`memory-router-revisit-promotion-decision.md`](memory-router-revisit-promotion-decision.md).
