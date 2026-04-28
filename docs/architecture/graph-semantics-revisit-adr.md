---
decision_id: adr-graph-semantics-revisit
decision_title: Graph Semantics Revisit
decision_status: accepted
decision_scope: graph-semantics
decision_owner: platform
---
# ADR: Graph Semantics Revisit

## Status

Accepted as a deferred-capability revisit track.

This ADR frames the graph semantics revisit as evidence gathering only. It does
not add runner actions, schemas, storage behavior, migrations, skill behavior,
or public APIs.

## Context

OpenClerk already has a reference graph semantics POC in
[`knowledge-configuration-v1-adr.md`](knowledge-configuration-v1-adr.md) and
[`../evals/results/ockp-graph-semantics-reference-poc.md`](../evals/results/ockp-graph-semantics-reference-poc.md).
That reference kept semantic relationship meaning in canonical markdown and
treated graph output as derived structural navigation with citations and
projection freshness.

The revisit asks whether that reference decision still holds under the
deferred-capability promotion rubric in
[`deferred-capability-promotion-gates.md`](deferred-capability-promotion-gates.md).
The track tests both:

- `capability_gap`: whether existing `openclerk document` and
  `openclerk retrieval` actions are structurally insufficient for
  relationship-shaped tasks.
- `ergonomics_gap`: whether existing actions can express the workflow but are
  too slow, too scripted, too brittle, too guidance-dependent, or too costly for
  routine AgentOps use.

## Decision

Use targeted ADR, POC, eval, and decision artifacts before any implementation
work. The default outcome is to keep graph semantics as reference/deferred
unless repeated targeted evidence proves a capability gap or ergonomics gap.

The current public surface remains:

- `openclerk document`
- `openclerk retrieval`

Canonical markdown remains the authority for semantic relationship meaning.
Graph output may expose structural links, backlinks, graph neighborhoods, and
projection freshness, but it must not become an independent truth system.

## Options

| Option | Description | Promotion posture |
| --- | --- | --- |
| Keep current primitives | Use search, `get_document`, `document_links`, backlinks, `graph_neighborhood`, provenance, and projection freshness over canonical markdown. | Default/reference if natural and scripted pressure pass with acceptable ergonomics. |
| Add a narrow graph query surface | Add a promoted runner action that packages relationship search, links, backlinks, graph neighborhood, citations, and freshness. | Consider only if repeated natural-intent rows show unacceptable UX while scripted controls prove current primitives are technically sufficient. |
| Add semantic-label graph authority | Store or infer relationship labels as graph truth independent of markdown. | Kill unless it can preserve canonical markdown authority, citations, provenance, freshness, and no-bypass invariants. |

## Invariants

Any future promoted surface must preserve:

- AgentOps-only routine operation through installed runner JSON.
- Canonical markdown authority for relationship meaning.
- Citations, source refs, or stable source identifiers for source-sensitive
  claims.
- Inspectable provenance and graph projection freshness.
- Local-first operation.
- No broad repo search, direct vault inspection, direct SQLite, source-built
  runner paths, HTTP/MCP bypasses, backend variants, module-cache inspection, or
  ad hoc lower-level transports for routine tasks.
- Final-answer-only handling for invalid no-tools requests.

## Non-Goals

This ADR does not:

- Define a promoted graph runner action.
- Add semantic edge-label storage.
- Add a graph database, migration, index schema, or public API.
- Replace markdown links, backlinks, search, provenance, or freshness checks.
- Make graph state more authoritative than canonical docs or promoted records.

## Promotion Gates

Promotion via `capability_gap` requires repeated scripted-control failures that
show current document/retrieval primitives cannot safely express
relationship-shaped graph workflows while preserving citations, provenance, and
freshness.

Promotion via `ergonomics_gap` requires repeated natural-intent failures or
unacceptable UX cost where the scripted control still passes. Evidence must
show high step count, latency, prompt brittleness, retries, or workflow-specific
guidance dependence that a proposed surface would reduce without weakening the
invariants above.

Defer or keep as reference when current primitives pass with acceptable
ergonomics, when failures are data hygiene, ordinary skill guidance, eval
coverage, or partial evidence, or when the proposed surface would mostly be a
more complicated way to do docs retrieval.

Kill the candidate if semantic edges become independent truth, lack source
evidence, hide stale graph state, drop citations, hide provenance/freshness, or
encourage bypasses.

## Evidence Plan

The POC comparison is
[`../evals/graph-semantics-revisit-comparison-poc.md`](../evals/graph-semantics-revisit-comparison-poc.md).
The targeted reduced report is
[`../evals/results/ockp-graph-semantics-revisit-pressure.md`](../evals/results/ockp-graph-semantics-revisit-pressure.md).
The final promotion decision is
[`graph-semantics-revisit-promotion-decision.md`](graph-semantics-revisit-promotion-decision.md).
