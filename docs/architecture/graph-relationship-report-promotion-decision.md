---
decision_id: decision-graph-relationship-report-promotion
decision_title: Graph Relationship Report Promotion
decision_status: accepted
decision_scope: graph-relationship-report
decision_owner: platform
source_refs: docs/evals/graph-relationship-report-implementation.md, docs/evals/results/ockp-graph-relationship-report-implementation.md, docs/architecture/graph-product-story-promotion-decision.md, docs/architecture/graph-relationship-maintenance-plan-promotion-decision.md
follow_up_work: none
---
# Decision: Graph Relationship Report Promotion

## Status

Accepted: promote the narrow read-only `graph_relationship_report` retrieval
action.

This closes `oc-sl7c` by combining the deferred read-only graph needs from
`oc-tfms`: relationship/path finding, direct-vs-derived relationship
reporting, typed relationship candidates from canonical markdown, and limited
stale/contradictory/orphaned graph audits.

No follow-up work is required for those read-only deferred needs. The
separate approval-gated maintenance-plan need is resolved by
[`docs/architecture/graph-relationship-maintenance-plan-promotion-decision.md`](graph-relationship-maintenance-plan-promotion-decision.md).

Evidence:

- [`docs/evals/graph-relationship-report-implementation.md`](../evals/graph-relationship-report-implementation.md)
- [`docs/evals/results/ockp-graph-relationship-report-implementation.md`](../evals/results/ockp-graph-relationship-report-implementation.md)
- [`docs/architecture/graph-product-story-promotion-decision.md`](graph-product-story-promotion-decision.md)

## Decision

Promote `graph_relationship_report` under `openclerk retrieval`.

The action packages relationship paths, direct relationship evidence, derived
relationship evidence, typed relationship candidates, limited audit findings,
graph projection freshness, provenance refs, validation boundaries, authority
limits, and `graph_relationship.agent_handoff`.

`graph_context_report` remains the broad context baseline. Existing primitives
remain the fallback for explicit drill-down and rejection repair.

## Safety, Capability, UX

Safety pass: pass. The selected action is read-only, local-first, runner-only,
citation-bearing, and explicit that canonical markdown remains relationship
authority. It does not write, inspect the vault or SQLite directly, use
source-built runners, use unsupported transports, create graph memory, rank
authority, add durable semantic graph storage, or claim semantic-label graph
truth.

Capability pass: pass. The selected action covers all `oc-sl7c` read-only
needs in one response: paths, direct-vs-derived evidence, typed candidates
from cited markdown wording, stale projection status, orphaned context status,
simple contradictory relationship wording status, graph projection freshness,
and provenance refs.

UX quality: pass. A normal user asking for relationship paths, relationship
types, or graph audit status should not need to manually compose
`graph_context_report`, `document_links`, `graph_neighborhood`,
`projection_states`, and `provenance_events`. One combined report is simpler
than split specialized reports because the candidate and audit slices share
the same source evidence.

## Authority Model

Canonical markdown remains semantic relationship authority.
`typed_relationship_candidates` are labels suggested from cited markdown text,
not durable graph facts. `derived_relationships` come from graph edges and
backlinks as navigation evidence, not independent relationship inference.
Limited contradiction findings do not replace broader source-sensitive audit
work.

## Candidate Outcome

| Candidate | Outcome | Rationale |
| --- | --- | --- |
| `current_primitives_plus_graph_context_report` | Keep as fallback/reference. | Safe and capable, but still too ceremonial for the deferred user-facing graph report needs. |
| `graph_relationship_report` | Promote. | Best read-only surface: combines useful behaviors without creating new durable authority. |
| `split_specialized_reports` | Do not select. | More API surface and more agent routing work for the same underlying evidence. |
