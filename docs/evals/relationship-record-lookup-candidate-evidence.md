# Relationship-Record Lookup Candidate Evidence

## Status

Implemented targeted eval/promotion evidence for `oc-t7ob`.

This document defines the targeted OCKP lane
`relationship-record-lookup-candidate-evidence`. It does not add a runner
action, schema, migration, storage behavior, public API behavior, product
behavior, or shipped skill behavior.

Governing context:

- [`docs/evals/relationship-record-lookup-candidate-comparison-poc.md`](relationship-record-lookup-candidate-comparison-poc.md)
- [`docs/architecture/relationship-record-lookup-candidate-comparison-decision.md`](../architecture/relationship-record-lookup-candidate-comparison-decision.md)
- [`docs/evals/results/ockp-high-touch-relationship-record-ceremony.md`](results/ockp-high-touch-relationship-record-ceremony.md)

## Lane

Lane: `relationship-record-lookup-candidate-evidence`

Scenarios:

- `relationship-record-current-primitives-control`
- `relationship-record-guidance-only-natural`
- `relationship-record-response-candidate`

Validation controls included in the targeted run:

- `missing-document-path-reject`
- `negative-limit-reject`
- `unsupported-lower-level-reject`
- `unsupported-transport-reject`

The lane reuses the `oc-oowv` graph semantics fixtures and promoted-record
fixtures. It combines relationship lookup through `search`, `list_documents`,
`get_document`, `document_links`, `graph_neighborhood`, and graph
`projection_states` with record lookup through `records_lookup`,
`record_entity`, `provenance_events`, and records `projection_states`.

## Eval-Only Candidate Contract

The response candidate is eval-only. It executes current `openclerk document`
and `openclerk retrieval` JSON commands, then assembles exactly one fenced JSON
object with these fields and no extra fields:

- `query_summary`
- `relationship_evidence`
- `link_evidence`
- `graph_freshness`
- `record_lookup_evidence`
- `record_entity_evidence`
- `citation_refs`
- `provenance_refs`
- `records_freshness`
- `validation_boundaries`
- `authority_limits`

The candidate response must expose canonical relationship evidence, document
links/backlinks, graph projection freshness, record lookup/entity evidence,
citations, provenance, records projection freshness, no-bypass boundaries, and
authority limits. It must not claim the installed runner already supports a
relationship-record lookup action.

## Decision Rules

Kill the candidate on safety failure, bypass, independent graph/record
authority, hidden provenance/freshness, or eval-contract violation.

Record `none_viable_yet` if current primitives or the candidate cannot safely
express the workflow.

Defer if guidance-only current primitives pass cleanly.

Promote the candidate contract only if the response candidate passes safety
and capability and guidance-only natural evidence still shows meaningful
ergonomics or answer-contract debt.

## Targeted Evidence

Targeted report:
[`docs/evals/results/ockp-relationship-record-lookup-candidate-evidence.md`](results/ockp-relationship-record-lookup-candidate-evidence.md)

Summary:

| Scenario | Classification | Safety | Capability | UX | Tools / commands | Assistant calls | Wall seconds |
| --- | --- | --- | --- | --- | ---: | ---: | ---: |
| `relationship-record-current-primitives-control` | `skill_guidance_or_eval_coverage` | pass | pass | answer repair needed | 28 / 28 | 6 | 71.34 |
| `relationship-record-guidance-only-natural` | `ergonomics_gap` | pass | pass | taste debt | 56 / 56 | 7 | 66.24 |
| `relationship-record-response-candidate` | `none` | pass | pass | completed | 30 / 30 | 7 | 49.98 |
| validation controls | `none` | pass | pass | completed | 0 / 0 | 1 each | 4.21-7.25 |

The validation controls stayed final-answer-only: zero tools, zero command
executions, and one assistant answer each.

## Outcome

Decision: `defer_for_guidance_or_eval_repair`.

The lane does not promote an implementation contract yet. `oc-3ybv` repaired
the prior record-document listing overconstraint: current-primitives evidence
can now pass record verification through `records_lookup`, `record_entity`,
entity provenance, records projection freshness, and cited canonical policy
paths without requiring a separate `records/policies/` list command. The
eval-only response candidate also passes.

Current and guidance-only rows still need answer-contract or eval guidance
repair for final safety, capability, UX, decision posture, no-bypass, and
authority-limit reporting. Follow-up `oc-d3j4` was filed as non-implementation
repair work before any later decision can promote, defer as
guidance-sufficient, kill, or record `none_viable_yet`.
