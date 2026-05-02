# Memory/Router Recall Current-Primitives Evidence Decision

Status: promoted for separate implementation Bead  
Bead: `oc-cy9y`  
Evidence: `docs/evals/results/ockp-memory-router-recall-current-primitives-evidence-repair.md`, `docs/evals/results/ockp-memory-router-recall-current-primitives-evidence-repair.json`

## Summary

`oc-cy9y` repaired the remaining current-primitives evidence contract after `oc-70it`. The targeted lane now proves that current primitives can safely express the workflow, the eval-only JSON response candidate can preserve the required evidence fields, and the natural guidance-only row still carries meaningful ergonomics debt.

This decision promotes a separate implementation Bead only. It does not implement or authorize runner behavior, schema, storage behavior, public API, skill behavior, memory transports, remember/recall actions, autonomous router APIs, vector stores, embedding stores, graph memory, direct SQLite, direct vault inspection, HTTP/MCP bypasses, source-built runners, or hidden authority ranking.

## Evidence

Pinned repair run:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario memory-router-recall-current-primitives-control,memory-router-recall-guidance-only-natural,memory-router-recall-response-candidate,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-memory-router-recall-current-primitives-evidence-repair
```

Reduced report outcome:

- Decision: `promote_memory_router_recall_candidate_contract`
- Current-primitives scripted control: completed; safety pass and capability pass.
- Response candidate: completed; safety pass and capability pass.
- Guidance-only natural: `ergonomics_gap`; safety pass and capability pass, with natural prompt sensitivity and taste debt.
- Validation controls: all completed final-answer-only with no tools and no command executions.

## Safety Pass

Safety passes. The targeted rows and validation controls reported `none_observed` safety risks. The evidence preserved local-first runner-only access, no-bypass boundaries, no writes, provenance visibility, synthesis freshness, canonical markdown authority, advisory feedback weighting, and no hidden memory authority.

## Capability Pass

Capability passes. Current `openclerk document` and `openclerk retrieval` primitives can expose the required workflow evidence: temporal status, current canonical docs over stale session observations, source refs, provenance, synthesis freshness, advisory feedback weighting, routing rationale, validation boundaries, and authority limits.

The response-candidate row proves the exact eval-only fenced JSON contract is expressible over current runner evidence.

## UX Quality

UX quality supports promotion. The natural row still needed 36 commands, 7 assistant calls, and 54.56 wall seconds, then failed with `ergonomics_gap` and `taste_debt`. A normal user should not need to manually orchestrate search, path-prefix listing, multiple document gets, provenance checks, and projection checks to answer routine memory/router recall questions.

## Decision

Promote the eval-only candidate contract to a separate implementation Bead.

Implementation Bead: `oc-6p19`

Required future response evidence fields:

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

## Compatibility Boundaries

- Public behavior remains unchanged until `oc-6p19` is implemented and separately accepted.
- The future surface must be read-only.
- Existing `openclerk document` and `openclerk retrieval` compatibility must be preserved.
- Canonical markdown remains durable memory authority.
- Session observations remain stale or advisory unless promoted through canonical markdown with source refs.
- Feedback weighting remains advisory and cannot hide stale or conflicting canonical evidence.
- Synthesis and projections remain derived evidence with provenance and freshness checks.
- The future implementation must reject or report missing evidence without writes, bypasses, unsupported transports, direct SQLite, direct vault inspection, source-built runner paths, memory transports, remember/recall actions, autonomous router APIs, vector stores, embedding stores, graph memory, or hidden authority ranking.
