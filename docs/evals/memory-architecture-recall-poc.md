# Memory Architecture And Recall POC

## Scope

This POC compares memory recall candidate surfaces for `oc-uj2y.3`.

The POC does not add a new memory store. It uses existing deterministic runner
fixtures and the installed `memory_router_recall_report` behavior as the
promoted read-only baseline.

## Candidate Shapes

| Shape | What It Proves | What It Does Not Prove |
| --- | --- | --- |
| Current primitives only | Search, list, get, provenance, and projection checks can express memory evidence. | Acceptable routine UX; previous evidence showed high step count and prompt choreography. |
| Source-linked memory docs | Canonical markdown can hold durable memory policy and recall evidence. | Fast repeated recall without a report action. |
| `memory_router_recall_report` | One read-only action returns approved recall fields and no-bypass boundaries. | Autonomous memory writes or a general memory API. |
| Mem0 or external memory | Useful recall architecture comparison. | OpenClerk authority, privacy, freshness, and local-first boundaries. |

## Selected POC Surface

`memory_router_recall_report` remains selected:

```json
{"action":"memory_router_recall_report","memory_router_recall":{"query":"memory router temporal recall session promotion feedback weighting routing canonical docs","limit":10}}
```

The report returns:

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

## Taste Review

A normal user should not have to understand Mem0, router classification,
projection freshness, stale session observations, and provenance calls just to
ask a routine memory/router recall question. A single read-only runner report
is the simpler surface while OpenClerk keeps memory writes out of scope.
