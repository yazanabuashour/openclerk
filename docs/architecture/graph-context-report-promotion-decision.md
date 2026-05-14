---
decision_id: decision-graph-context-report-promotion
decision_title: Graph Context Report Promotion
decision_status: accepted
decision_scope: graph-context-report
decision_owner: platform
source_refs: docs/evals/graph-context-report-implementation.md, docs/evals/results/ockp-graph-context-report-implementation.md, docs/architecture/graph-semantics-revisit-promotion-decision.md
---
# Decision: Graph Context Report Promotion

## Status

Accepted: promote the narrow read-only `graph_context_report` retrieval action
for routine relationship graph context.

This promotes a report-shaped workflow action only. It does not promote richer
graph semantics, semantic-label graph storage, a graph database, schema or
migration work, graph memory, hidden authority ranking, write behavior, direct
vault inspection, direct SQLite inspection, source-built runner usage, HTTP/MCP
bypasses, unsupported transports, or another public command.

Evidence:

- [`docs/evals/graph-context-report-implementation.md`](../evals/graph-context-report-implementation.md)
- [`docs/evals/results/ockp-graph-context-report-implementation.md`](../evals/results/ockp-graph-context-report-implementation.md)
- [`docs/architecture/graph-semantics-revisit-promotion-decision.md`](graph-semantics-revisit-promotion-decision.md)

## Decision

Promote `graph_context_report` under `openclerk retrieval`.

The action packages source document identity, cited canonical markdown
relationship text, outgoing links, incoming backlinks, nearby graph evidence,
graph projection freshness, provenance refs, candidate-surface comparison,
validation boundaries, authority limits, and `graph_context.agent_handoff`.

Keep existing primitives for explicit drill-down:

- `search`
- `document_links`
- `graph_neighborhood`
- `provenance_events`
- `projection_states`
- `openclerk document` `list_documents` and `get_document`

## Safety, Capability, UX

Safety pass: pass when the targeted implementation lane records no writes, no
bypasses, no direct vault/SQLite/source inspection, no unsupported transports,
no graph memory, no hidden authority ranking, and no semantic-label graph
truth. Canonical markdown remains semantic relationship authority.

Capability pass: pass when the report returns source identity, canonical
relationship text with citations, links/backlinks, nearby graph context, graph
projection freshness, provenance refs, validation boundaries, and authority
limits from runner-visible evidence.

UX quality: promotion is justified because current primitives plus help can
express the workflow safely but remain ceremonial for normal routine
relationship inspection. The report action reduces the routine path to a
single read-only retrieval action while preserving the same authority and
provenance discipline.

## Compatibility

Existing document and retrieval primitives remain supported and remain the
fallback for explicit inspection. `graph_context_report` is a packaging
surface over existing evidence; graph evidence remains derived navigation
context, not semantic relationship truth.

This decision closes the current `oc-gy2s` evidence question with a promotion
outcome. It does not require more evidence before implementation acceptance.
