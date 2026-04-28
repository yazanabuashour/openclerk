---
decision_id: decision-generalized-artifact-ingestion-promotion
decision_title: Generalized Artifact Ingestion Promotion
decision_status: accepted
decision_scope: artifact-ingestion
decision_owner: platform
---
# Decision: Generalized Artifact Ingestion Promotion

## Status

Accepted: keep generalized artifact ingestion as reference pressure and defer
promotion of a broader production surface.

`oc-res` and `oc-04h` repaired the PDF fixture and transport gaps, so the
refreshed heterogeneous artifact lane can now evaluate agent behavior instead
of fixture reachability. The final `oc-no2` decision evaluates both promotion
paths from the deferred-capability gates: current primitives can safely express
the targeted workflows, and the observed ergonomics are not strong enough to
justify a new public runner surface.

Evidence:

- [`generalized-artifact-ingestion-adr.md`](generalized-artifact-ingestion-adr.md)
- [`../evals/artifact-ingestion-architecture-options-poc.md`](../evals/artifact-ingestion-architecture-options-poc.md)
- [`../evals/results/ockp-heterogeneous-artifact-ingestion-pressure.md`](../evals/results/ockp-heterogeneous-artifact-ingestion-pressure.md)

## Decision

Do not promote `ingest_artifact`, artifact-specific ingestion actions, parser
pipelines, storage migrations, local file ingestion, OCR receipt ingestion,
native video/YouTube ingestion, or new public APIs from this evidence.

The current promoted public surface remains:

- `openclerk document`
- `openclerk retrieval`
- existing `ingest_source_url` for PDF source URLs

Capability path: no promotion. The refreshed targeted evidence reports
`fixture_preflight: passed` for both PDF rows and no remaining
`runner_capability_gap` classification. It shows passing coverage for scripted
PDF source URL ingestion, natural PDF source URL intent,
markdown-transcribed transcripts, invoice/receipt authority retrieval,
mixed-artifact synthesis freshness, missing source hints, unsupported native
video rejection, and bypass rejection.

Ergonomics path: no promotion. The natural-intent PDF and mixed-artifact rows
are tool-heavy and high latency, but they completed without retries or contract
violations. The evidence remains useful pressure for future guidance and design,
not enough proof that a generalized artifact ingestion action would reduce
routine AgentOps cost while preserving authority, citations, provenance,
freshness, and local-first operation.

Native video/YouTube ingestion, OCR-heavy receipts, local file import, and
other parser-backed artifact workflows remain deferred. They require a later
promotion decision with repeated capability-gap or ergonomics-gap evidence and
an exact request/response surface.

## Promotion Rubric

| Outcome | Standard |
| --- | --- |
| Promote | Repeated targeted scenarios show `runner_capability_gap`, and the exact promoted surface preserves authority, citations, provenance, freshness, local-first storage, and AgentOps. |
| Defer | Failures are data hygiene, skill guidance, eval coverage, partial evidence, or one-off assistant handling. |
| Kill | The capability creates a second truth surface, weakens provenance/freshness, hides citations, increases duplicate authority, or requires routine bypasses. |
| Keep as reference | Existing document/retrieval workflows are sufficient, but the lane remains useful pressure for guidance and future design. |

The current decision is **keep as reference** for the targeted heterogeneous
artifact lane, **defer** generalized artifact ingestion promotion, and
**defer** native parser-backed ingestion surfaces.

## Required Gates For Future Promotion

A future promotion must name:

- exact runner action names and JSON request/response shapes
- supported artifact kinds, URI schemes, path-hint rules, asset-hint rules, and
  update semantics
- compatibility with existing `ingest_source_url` create/update/conflict
  behavior
- duplicate, partial-success, parser-failure, unsupported-kind, and stale
  synthesis failure modes
- citation mapping from parsed artifact content to canonical markdown chunks
- provenance events and projection freshness semantics
- targeted eval scenarios that repeatedly fail without the proposed surface

No follow-up implementation Beads are filed from this decision because the
evidence does not justify promotion. Follow-up repair work may be filed only
for eval data hygiene or verifier coverage, not for a production ingestion
surface.
