# Structured Data And Canonical Stores POC

## Scope

This POC compares candidate surfaces for `oc-uj2y.4`: structured facts and
non-document canonical stores.

The POC keeps OpenClerk local-first and runner-only. It does not inspect raw
storage, add migrations, import external datasets, or create durable
structured writes.

## Candidate Shapes

| Shape | What It Proves | What It Does Not Prove |
| --- | --- | --- |
| Current primitives | `records_lookup`, `services_lookup`, `decisions_lookup`, `provenance_events`, and `projection_states` can express structured evidence safely. | Acceptable routine UX for candidate-store decisions. |
| Domain-specific typed actions | Services and decisions show the selective-domain pattern works when schema is stable. | That every structured domain deserves its own action. |
| `structured_store_report` | One read-only action can package current projection evidence and candidate-store boundaries. | Independent canonical tables or durable structured writes. |
| Independent SQLite canonical tables | Potential future shape for dense measurements or time series. | Approval, correction, provenance, freshness, duplicate handling, or markdown reconciliation. |
| External domain stores/connectors | Useful reference for future imports. | Routine local-first OpenClerk authority. |

## Selected POC Surface

`structured_store_report`:

```json
{"action":"structured_store_report","structured_store":{"domain":"records","query":"structured canonical record evidence","entity_type":"tool","limit":10}}
```

The report returns:

- `records`, `services`, or `decisions` evidence for the selected domain
- `projections` for the selected projection family
- `candidate_surfaces`
- `recommendation`
- `safety_pass`
- `capability_pass`
- `ux_quality`
- `evidence_posture`
- `validation_boundaries`
- `authority_limits`
- `agent_handoff`

## Taste Review

A normal user should not have to choose between hidden database authority and a
manual sequence of record lookup, projection freshness, provenance, and
candidate comparison just to ask whether a structured domain belongs in
OpenClerk. The promoted surface is a read-only report that makes the current
schema-backed evidence visible while refusing premature independent stores.
