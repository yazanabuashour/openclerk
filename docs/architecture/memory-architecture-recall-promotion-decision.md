---
decision_id: decision-memory-architecture-recall
decision_status: accepted
decision_scope: memory-architecture-recall
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/architecture/memory-architecture-recall-adr.md, docs/evals/memory-architecture-recall-poc.md, docs/evals/results/ockp-memory-architecture-recall-track.md, docs/architecture/memory-router-recall-report-implementation-decision.md
---

# Memory Architecture And Recall Promotion Decision

## Decision

Accept the existing `memory_router_recall_report` as the promoted
source-linked memory recall surface for `oc-uj2y.3`.

No additional product behavior is required in this epic because the selected
surface is already implemented, documented, exposed in runner help and skill
guidance, and covered by tests and eval evidence.

## Safety Pass

Pass. The selected report is read-only and forbids writes, memory transports,
`remember`/`recall` actions, autonomous router APIs, vector stores, embedding
stores, graph memory, direct SQLite, direct vault inspection, HTTP/MCP
bypasses, source-built runners, unsupported transports, and hidden authority
ranking.

## Capability Pass

Pass. The report exposes the approved memory recall evidence fields:

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

Pass. The current promoted report avoids the high-step current-primitives
ceremony while keeping memory authority explicit and source-linked.

## Conditional Implementation

Already present:

- runner JSON action `memory_router_recall_report`
- request object `memory_router_recall`
- read-only execution path
- approved report fields
- runner help and skill action index
- tests in `internal/runner/runner_retrieval_test.go`
- implementation decision and eval report under committed docs

No schema, storage, projection, or skill expansion is needed in this epic.

## Iteration Gate

Future memory work should compare source-linked memory docs, internal derived
memory projections, and Mem0-style recall only after evidence shows repeated
recall needs that the report cannot satisfy. Any future write path must require
approval before durable writes and preserve source refs, freshness, provenance,
privacy boundaries, and canonical override behavior.
