---
decision_id: decision-local-ocr-artifact-candidate-planning
decision_title: Local OCR Artifact Candidate Planning
decision_status: accepted
decision_scope: artifact-candidate-plan-local-ocr
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/local-ocr-artifact-candidate-planning-poc.md, docs/evals/results/ockp-local-ocr-artifact-fixture-evidence.md, docs/evals/results/ockp-local-ocr-review-contract-dependency-policy.md, docs/evals/results/ockp-artifact-local-path-candidate-plan-promotion.md, docs/architecture/local-artifact-candidate-plan-promotion-decision.md, docs/architecture/parser-ocr-artifact-ingestion-candidate-comparison-decision.md
---
# Decision: Local OCR Artifact Candidate Planning

## Status

Accepted as a non-promotion decision for `oc-d1pm`.

Do not add runtime OCR, scanned-PDF OCR fallback, image parsing, domain-specific
receipt or invoice OCR actions, generalized artifact ingestion, parser
pipelines, storage migrations, asset registry behavior, public APIs, shipped
skill behavior, or product implementation work from this evidence.

Evidence:

- [`docs/evals/local-ocr-artifact-candidate-planning-poc.md`](../evals/local-ocr-artifact-candidate-planning-poc.md)
- [`docs/evals/results/ockp-local-ocr-artifact-fixture-evidence.md`](../evals/results/ockp-local-ocr-artifact-fixture-evidence.md)
- [`docs/evals/results/ockp-local-ocr-review-contract-dependency-policy.md`](../evals/results/ockp-local-ocr-review-contract-dependency-policy.md)
- [`docs/evals/results/ockp-artifact-local-path-candidate-plan-promotion.md`](../evals/results/ockp-artifact-local-path-candidate-plan-promotion.md)
- [`docs/architecture/local-artifact-candidate-plan-promotion-decision.md`](local-artifact-candidate-plan-promotion-decision.md)
- [`docs/architecture/parser-ocr-artifact-ingestion-candidate-comparison-decision.md`](parser-ocr-artifact-ingestion-candidate-comparison-decision.md)

## Decision

Select `none viable yet` for implementation promotion. Keep image OCR and
scanned-PDF OCR unsupported in the current runner.

The best future candidate remains a read-only extension of the existing
`artifact_candidate_plan` action, not a new generalized artifact ingestion
surface:

```json
{"action":"artifact_candidate_plan","artifact":{"local_path":"<explicit-user-local-file>","artifact_kind":"receipt","text_extraction":"ocr_review","limit":5}}
```

That shape is not promoted here. It is only the preferred future candidate if a
later evidence pass proves local-first OCR dependency policy, confidence,
provenance, page/image refs, low-confidence correction, duplicate checks,
unsupported-file behavior, and approval-before-write.

## Candidate Comparison

| Candidate | Safety | Capability | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Extend `artifact_candidate_plan` with OCR review | Plausible but unproven; needs explicit extractor identity, local-first dependency policy, confidence, page/image refs, correction flow, and no-write boundary. | Best future fit for explicit local OCR planning. | Best taste fit because OCR is adjacent to current local artifact planning. | Defer. |
| Scanned-PDF OCR plus review | Plausible but unproven; text-bearing PDF is safe today, OCR fallback is not. | Strongest concrete gap after PDF text extraction. | Strong normal-user expectation, but review and uncertainty behavior are missing. | Defer. |
| Domain-specific receipt/invoice OCR actions | Risky before generic OCR review exists; field extraction can overstate parser authority. | Potential later convenience for high-value artifacts. | Could be simpler later, but premature now. | Defer. |
| `none viable yet` | Passes by preserving current unsupported boundaries and no hidden parser truth. | Does not add OCR capability. | Acceptable as a temporary boundary; need remains valid. | Select. |

## Safety, Capability, UX

Safety pass: pass for the selected non-promotion. Existing
`artifact_candidate_plan` local file planning remains read-only, limited to
UTF-8 text, markdown, and text-bearing PDF files, and continues to reject OCR,
opaque parsing, hidden file inspection, direct vault or SQLite access, fetches,
and durable writes before approval.

Capability pass: partial. Current behavior covers explicit local text,
markdown, and text-bearing PDF candidate planning. It does not cover image OCR
or scanned-PDF OCR candidate recovery, and this decision does not prove an OCR
candidate can preserve confidence, provenance, review, duplicate, and
correction behavior.

UX quality: partial. A normal user would reasonably expect scanned PDFs or
receipt images to become reviewable OpenClerk candidates, especially now that
text-bearing PDFs are supported. The evaluated OCR shapes remain too
underspecified to promote; deferral is safety-preserving but leaves valid taste
and capability pressure.

## Follow-Up

Search performed before close:

- `bd search "OCR scanned PDF artifact candidate plan" --status all`
- `bd search "local OCR artifact candidate planning" --status all`

No product implementation Bead should be filed from `oc-d1pm`. A follow-up may
gather evidence or compare fixtures, but implementation must remain blocked
until a later accepted promotion decision names the exact request/response
surface and passes the required gates.

Created follow-up:

- `oc-osmc`: gather local OCR artifact fixture evidence.

`oc-osmc` completed the fixture-evidence pass in
[`docs/evals/results/ockp-local-ocr-artifact-fixture-evidence.md`](../evals/results/ockp-local-ocr-artifact-fixture-evidence.md).
It selected `none viable yet`, filed no product implementation Bead, and
created deferred follow-up `oc-i8yk` for the next OCR review contract and
dependency-policy pass.

`oc-i8yk` completed that design pass in
[`docs/architecture/local-ocr-review-contract-dependency-policy-decision.md`](local-ocr-review-contract-dependency-policy-decision.md).
It selected model-assisted OCR review as the simplest future evidence
candidate, kept local OCR engines as the required local-first comparison, and
filed follow-up `oc-s3wg` for targeted promotion evidence. It filed no product
implementation Bead.

## Compatibility

Existing behavior remains unchanged:

- `artifact_candidate_plan` remains the local artifact planning surface for
  explicit UTF-8 text, markdown, and text-bearing PDF local files.
- OCR/image parsing, scanned-PDF OCR fallback, and opaque binary parsing remain
  unsupported.
- Durable writes still require approved `create_document` or
  `ingest_source_url`.
- Committed docs and reports must use repo-relative paths or neutral
  placeholders such as `<explicit-user-local-file>`.
