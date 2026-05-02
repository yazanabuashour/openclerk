# Memory/Router Recall Report Implementation Eval

`oc-6p19` implements the promoted read-only memory/router recall report as an
`openclerk retrieval` action. This eval verifies the installed surface rather
than re-running candidate comparison.

## Surface

The implemented request shape is:

```json
{
  "action": "memory_router_recall_report",
  "memory_router_recall": {
    "query": "memory router temporal recall session promotion feedback weighting routing canonical docs",
    "limit": 10
  }
}
```

The `memory_router_recall` result contains exactly the approved report fields:

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

## Boundaries

The action is read-only and uses existing runner-visible evidence: search,
canonical memory/router documents, session observation provenance, and synthesis
projection freshness. It does not add writes, memory transports,
`remember`/`recall` actions, autonomous router APIs, vector stores, embedding
stores, graph memory, direct SQLite, direct vault inspection, HTTP/MCP bypasses,
unsupported transports, source-built runner paths, storage changes, schema
changes, or hidden authority ranking.

Missing evidence is reported in the response fields and validation boundaries
without bypassing the runner or attempting repair.

## Targeted Run

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario memory-router-recall-report-action-control,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-memory-router-recall-report-implementation
```

The reduced report must be published as:

- `docs/evals/results/ockp-memory-router-recall-report-implementation.md`
- `docs/evals/results/ockp-memory-router-recall-report-implementation.json`
