---
decision_id: decision-memory-router-recall-report-implementation
decision_status: accepted
decision_scope: memory-router-recall-report
decision_owner: agentops
decision_date: 2026-05-01
source_refs: docs/architecture/memory-router-recall-current-primitives-evidence-decision.md, docs/evals/memory-router-recall-report-implementation.md, docs/evals/results/ockp-memory-router-recall-report-implementation.md
---

# Memory/Router Recall Report Implementation Decision

## Decision

Accept `memory_router_recall_report` as the promoted read-only memory/router
recall report action under `openclerk retrieval`.

The action implements the `oc-cy9y` promotion decision without adding a new
command, write workflow, storage behavior, schema, memory transport,
`remember`/`recall` action, autonomous router API, vector store, embedding
store, graph memory, direct SQLite access, direct vault inspection, HTTP/MCP
bypass, source-built runner path, unsupported transport, or hidden authority
ranking.

## Safety Pass

Safety passes when the targeted implementation run completes with no writes,
no bypasses, no prohibited transport, and validation controls remain
final-answer-only. Missing evidence is reported in the read-only response
instead of repaired or fetched through a bypass.

## Capability Pass

Capability passes when the report returns the approved fields from
runner-visible search, canonical memory/router docs, session observation
provenance, and synthesis projection freshness:

- `query_summary`
- `temporal_status`
- `canonical_evidence_refs`
- `stale_session_status`
- `feedback_weighting`
- `routing_rationale`
- `provenance_refs`
- `synthesis_freshness`
- `validation_boundaries`
- `authority_limits`

## UX Quality

UX quality improves over the prior current-primitives ceremony because a normal
caller can request the report with one retrieval action instead of coordinating
search, list, get, provenance, projection, and answer-shape choreography.

## Compatibility

Existing `openclerk document` and `openclerk retrieval` actions remain
compatible. Canonical markdown remains durable memory authority; synthesis is
derived evidence with provenance and freshness; feedback remains advisory; and
the report does not approve durable writes or autonomous routing decisions.
