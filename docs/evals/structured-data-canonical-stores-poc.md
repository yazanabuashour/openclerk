# Structured Data And Canonical Stores POC

## Scope

This POC compares candidate surfaces for `oc-uj2y.4`: structured facts and
non-document canonical stores.

The POC keeps OpenClerk local-first and runner-only. It does not inspect raw
storage, add migrations, import external datasets, or create durable
structured writes.

Required references:

- docs/architecture/agent-knowledge-plane.md
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Candidate Shapes

| Shape | What It Proves | What It Does Not Prove |
| --- | --- | --- |
| Current primitives | `records_lookup`, `services_lookup`, `decisions_lookup`, `provenance_events`, and `projection_states` can express structured evidence safely. | Acceptable routine UX for candidate-store decisions. |
| Domain-specific typed actions | Services and decisions show the selective-domain pattern works when schema is stable. | That every structured domain deserves its own action. |
| `structured_store_report` | One read-only action can package current projection evidence and candidate-store boundaries. | Independent canonical tables or durable structured writes. |
| Independent SQLite canonical tables | Potential future shape for dense measurements or time series. | Approval, correction, provenance, freshness, duplicate handling, or markdown reconciliation. |
| External domain stores/connectors | Useful reference for future imports. | Routine local-first OpenClerk authority. |

## Evidence Examples

Projection or report is enough for:

- service registry lookup, where markdown service docs are visible authority
  and the typed projection only improves filters
- decision record lookup, where ADR/decision markdown remains the canonical
  record and the projection exposes status, owner, scope, and freshness
- generic promoted records, where facts can be rebuilt from frontmatter and
  `## Facts` sections

An independent canonical store would require a separate domain track for:

- dense measurements or time-series observations
- correction-heavy imported rows with stable external IDs
- high-volume inventory or finance facts where markdown records cannot
  faithfully represent identity, deduplication, and correction semantics

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

## Closure

Safety pass, capability pass, and UX quality are recorded separately in
`docs/evals/results/ockp-structured-data-canonical-stores.md`. Remaining work
is represented by linked beads:

- `oc-tnnw.2.3` eval for safety, capability, and UX quality.
- `oc-tnnw.2.4` promotion decision.
- `oc-tnnw.2.5` conditional implementation only if promoted.
- `oc-tnnw.2.6` iteration and follow-up bead creation.
