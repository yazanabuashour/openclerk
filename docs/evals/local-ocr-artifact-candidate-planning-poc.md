# Local OCR Artifact Candidate Planning POC

## Status

Implemented candidate-comparison framing for `oc-d1pm`.

This document compares OCR and scanned-PDF candidate surfaces only. It does not
add a runner action, OCR engine, parser dependency, schema, storage migration,
asset registry behavior, public API, product behavior, shipped skill behavior,
or implementation authorization.

Governing evidence:

- [`docs/evals/results/ockp-local-ocr-artifact-fixture-evidence.md`](results/ockp-local-ocr-artifact-fixture-evidence.md)
- [`docs/evals/results/ockp-artifact-local-path-candidate-plan-promotion.md`](results/ockp-artifact-local-path-candidate-plan-promotion.md)
- [`docs/architecture/local-artifact-candidate-plan-promotion-decision.md`](../architecture/local-artifact-candidate-plan-promotion-decision.md)
- [`docs/architecture/parser-ocr-artifact-ingestion-candidate-comparison-decision.md`](../architecture/parser-ocr-artifact-ingestion-candidate-comparison-decision.md)
- [`docs/architecture/generalized-artifact-ingestion-promotion-decision.md`](../architecture/generalized-artifact-ingestion-promotion-decision.md)

## Baseline

`artifact_candidate_plan` is the promoted local artifact planning surface for
explicit `artifact.local_path` inputs, limited to UTF-8 text, markdown, and
text-bearing PDF files. It is read-only, inspects only the explicit local file,
returns candidate content/provenance/confidence/duplicate evidence, and keeps
durable writes behind approved `create_document` or `ingest_source_url`
requests.

Current unsupported boundaries are still part of the baseline:

- image OCR
- scanned-PDF OCR
- opaque binary parsing
- hidden local file inspection
- direct vault or SQLite access
- durable writes before approval

## Candidate Workflows

| Candidate | Shape | Strength | Risk |
| --- | --- | --- | --- |
| Extend `artifact_candidate_plan` with OCR review | Add an explicit read-only option such as `text_extraction: "ocr_review"` on `artifact.local_path`, returning local artifact metadata, extractor identity, page/image refs, extracted text preview, confidence, uncertainty notes, duplicate evidence, and no-write handoff. | Best taste fit: normal users would expect OCR to be adjacent to existing local artifact candidate planning, not a separate ingestion surface. It preserves approval-before-write if OCR output stays candidate text only. | Not proven yet: no local OCR dependency policy, confidence calibration, page/image provenance contract, correction workflow, resource limits, or repeated eval evidence. |
| Scanned-PDF extraction plus review | Keep PDF handling inside `artifact_candidate_plan`, adding OCR only when text extraction returns no text and the user explicitly requests review-mode extraction. | Targets the most natural gap left by text-bearing PDF support and avoids a broad image parser surface. | Still needs OCR dependency, confidence, provenance, page mapping, low-confidence correction, duplicate behavior, and unsupported-file behavior before promotion. |
| Domain-specific OCR artifact actions | Add receipt/invoice-specific candidate planners that OCR and normalize fields before proposing documents. | Could improve high-value workflows where field confidence, totals, vendors, and dates matter. | Too narrow and too authority-heavy without stronger evidence; risks hiding parser truth behind domain fields before the generic OCR review contract is proven. |
| `none viable yet` | Keep OCR/scanned-PDF unsupported while gathering targeted evidence and fixtures. | Safest current outcome; preserves existing promoted behavior and avoids unsupported OCR authority. | Does not solve the valid user expectation that scanned PDFs or receipt images should become reviewable OpenClerk candidates. |

## Selected Outcome

Select `none viable yet` for implementation promotion, while recording the best
future candidate as an extension of existing `artifact_candidate_plan`.

If later evidence justifies promotion, the candidate request should stay on the
existing read-only planning action rather than introducing generalized artifact
ingestion:

```json
{"action":"artifact_candidate_plan","artifact":{"local_path":"<explicit-user-local-file>","artifact_kind":"receipt","text_extraction":"ocr_review","limit":5}}
```

The future response must remain candidate evidence only. OCR text must not
become canonical knowledge until the user reviews or corrects it and approves a
durable `create_document` or `ingest_source_url` write.

## Evidence Scorecard

| Evidence | Safety | Capability | UX quality |
| --- | --- | --- | --- |
| Existing local text/markdown/text-PDF candidate planning | Passed: read-only explicit local file inspection, no OCR, no fetch, no durable write, local artifact metadata, duplicate evidence, and approved next-create handoff. | Passed for UTF-8 text, markdown, and text-bearing PDF. | Good: one existing action handles explicit local artifact planning without direct agent file parsing. |
| Image OCR candidate planning | Not proven: no OCR engine, dependency policy, resource limits, confidence calibration, provenance envelope, or correction workflow is documented. | Gap remains: current runner rejects image/OCR artifacts instead of producing reviewable candidate text. | Valid need, but not promotion-ready because a normal user also needs visible uncertainty and review before write. |
| Scanned-PDF OCR candidate planning | Not proven: text-bearing PDF extraction is implemented, but OCR fallback for scanned pages has no page-level provenance or confidence contract. | Gap remains for scanned PDFs with no extractable text. | Strongest future UX pressure because scanned PDFs are adjacent to the promoted PDF path. |
| Domain-specific receipt/invoice OCR | Not proven: field-level extraction could overstate authority without a generic OCR review envelope. | Possible future capability after OCR review is safe. | Defer; domain convenience should not precede confidence, provenance, and correction behavior. |
| Bypass and unsupported-file behavior | Passed in the current baseline: OCR/image parsing and opaque binary parsing remain unsupported. | Passed for rejection, not for extraction. | Acceptable as a safety boundary, but it leaves a real capability need. |

## Required Gates Before Promotion

A later implementation promotion must name:

- exact request and response fields, including any `text_extraction` option
- OCR dependency and local-first runtime policy
- supported file types, size/page limits, and unsupported-file errors
- extractor identity, page/image refs, confidence, and uncertainty notes
- correction workflow for low-confidence or disputed text
- duplicate checks before any next-create request
- approval-before-write handoff through existing durable actions
- Go tests for `internal/runner/artifact_candidate_plan.go` and
  `internal/runner/runner_document_test.go`
- targeted eval scenarios for scanned PDF, image OCR, low confidence,
  duplicate risk, and bypass rejection

## Conclusion

Do not file a product implementation work item from `oc-d1pm`. OCR and scanned-PDF
candidate planning remain valid OpenClerk needs, and the best future surface is
an explicit review-mode extension of `artifact_candidate_plan`, but the current
evidence does not prove confidence, provenance, correction, duplicate, or
review behavior strongly enough to promote runtime OCR.

Follow-up fixture evidence in
[`docs/evals/results/ockp-local-ocr-artifact-fixture-evidence.md`](results/ockp-local-ocr-artifact-fixture-evidence.md)
keeps this outcome unchanged: `none viable yet` remains selected, no product
implementation work item is justified, and deferred follow-up `oc-i8yk` tracks the
next OCR/scanned-PDF contract pass for a concrete OCR dependency and correction
workflow.
