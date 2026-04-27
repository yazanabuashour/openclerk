---
decision_id: decision-generalized-artifact-ingestion-promotion
decision_title: Generalized Artifact Ingestion Promotion
decision_status: accepted
decision_scope: artifact-ingestion
decision_owner: platform
---
# Decision: Generalized Artifact Ingestion Promotion

## Status

Accepted: defer generalized artifact ingestion promotion. Keep the targeted
lane as evidence pressure while repairing data hygiene and eval coverage.

`oc-no2` has been reopened for an ergonomics-gate refresh. The historical
decision below remains the current recorded outcome until `oc-res` repairs the
PDF fixture gap and a refreshed decision evaluates both capability-gap and
ergonomics-gap evidence.

Evidence:

- [`generalized-artifact-ingestion-adr.md`](generalized-artifact-ingestion-adr.md)
- [`../evals/artifact-ingestion-architecture-options-poc.md`](../evals/artifact-ingestion-architecture-options-poc.md)
- [`../evals/results/ockp-heterogeneous-artifact-ingestion-pressure.md`](../evals/results/ockp-heterogeneous-artifact-ingestion-pressure.md)

## Decision

Do not promote `ingest_artifact`, artifact-specific ingestion actions, parser
pipelines, storage migrations, or new public APIs from this evidence.

The current promoted public surface remains:

- `openclerk document`
- `openclerk retrieval`
- existing `ingest_source_url` for PDF source URLs

The targeted evidence does not show repeated `runner_capability_gap` failures.
It shows passing coverage for markdown-transcribed transcripts, invoice/receipt
authority retrieval, mixed-artifact synthesis freshness, missing source hints,
unsupported native video rejection, and bypass rejection, plus one
data-hygiene/eval-coverage failure in PDF source URL pressure that needs repair
before any stronger claim.

Native video/YouTube ingestion, OCR-heavy receipts, local file import, and
other parser-backed artifact workflows remain deferred. They require a later
promotion decision with repeated `runner_capability_gap` evidence and an exact
request/response surface.

## Promotion Rubric

| Outcome | Standard |
| --- | --- |
| Promote | Repeated targeted scenarios show `runner_capability_gap`, and the exact promoted surface preserves authority, citations, provenance, freshness, local-first storage, and AgentOps. |
| Defer | Failures are data hygiene, skill guidance, eval coverage, partial evidence, or one-off assistant handling. |
| Kill | The capability creates a second truth surface, weakens provenance/freshness, hides citations, increases duplicate authority, or requires routine bypasses. |
| Keep as reference | Existing document/retrieval workflows are sufficient, but the lane remains useful pressure for guidance and future design. |

The current decision is **defer** generalized artifact ingestion promotion and
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
