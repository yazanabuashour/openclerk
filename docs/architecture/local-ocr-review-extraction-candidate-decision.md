---
decision_id: decision-local-ocr-review-extraction-candidates
decision_title: Local OCR Review Extraction Candidates
decision_status: accepted
decision_scope: artifact-candidate-plan-ocr-extraction-candidates
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/local-ocr-review-extraction-candidates.md, docs/evals/results/ockp-local-ocr-review-extraction-candidates.md, docs/evals/results/ockp-local-ocr-review-contract-dependency-policy.md, docs/architecture/local-ocr-review-contract-dependency-policy-decision.md
---
# Decision: Local OCR Review Extraction Candidates

## Status

Accepted as the durable non-promotion decision for `oc-s3wg` and the local
OCR/scanned-PDF artifact-candidate path.

Supersession note: this historical non-promotion was superseded on
2026-05-05 by [`ocr-module-final-decision.md`](ocr-module-final-decision.md),
which promotes the optional local `modules/tesseract-ocr/module.json` module
for explicit `artifact_candidate_plan` OCR review. The constraints below remain
historical evidence for why hidden core OCR, model egress, OCR caches, and
domain-field extraction were not promoted.

Do not add runtime OCR, multimodal model calls, scanned-PDF OCR fallback,
image parsing, Tesseract, OCRmyPDF, Poppler rendering, Go OCR bindings,
domain-specific receipt or invoice OCR fields, parser pipelines, storage
migrations, public APIs, shipped skill behavior, or product implementation work
from this decision.

Evidence:

- [`docs/evals/local-ocr-review-extraction-candidates.md`](../evals/local-ocr-review-extraction-candidates.md)
- [`docs/evals/results/ockp-local-ocr-review-extraction-candidates.md`](../evals/results/ockp-local-ocr-review-extraction-candidates.md)
- [`docs/evals/results/ockp-local-ocr-review-contract-dependency-policy.md`](../evals/results/ockp-local-ocr-review-contract-dependency-policy.md)
- [`docs/architecture/local-ocr-review-contract-dependency-policy-decision.md`](local-ocr-review-contract-dependency-policy-decision.md)

## Decision

Select current unsupported OCR behavior as the final outcome for this path.

The evaluated OCR candidates are not promoted:

- model-assisted OCR review remains reference pressure only
- local Tesseract/OCRmyPDF-style extraction is not promoted
- Poppler-only fallback is killed as an OCR candidate because it is not OCR
- Go-native OCR bindings are not promoted
- receipt/invoice-specific OCR fields are future-only outside this path

Outcome category: durable non-promotion. The capability need is real, but this
specific path is closed because no candidate preserves all required safety,
provenance, dependency, correction, duplicate, unsupported-file, and
approval-before-write gates.

## Safety, Capability, UX

Safety pass: pass for non-promotion. Existing production behavior preserves
runner-only access, local-first behavior, no hidden model or parser truth,
explicit unsupported-file rejection, no direct vault or SQLite access, no
unsupported transports, no durable writes before approval, and existing
duplicate checks for supplied or supported extracted text.

Capability pass: partial. OpenClerk still does not recover OCR text from
scanned images or scanned PDFs. It continues to support explicit supplied
content, UTF-8 text, markdown, and text-bearing PDF local artifact planning.

UX quality: acceptable as a final boundary for this path. A normal user would
prefer OCR review, and model-assisted extraction remains the simplest taste
reference, but production OCR would require a broader runner-owned
provider/egress or local OCR dependency policy that this path does not prove.

## Rejected Alternatives

Do not promote model-assisted OCR as the simple path yet. It would be the best
UX if the runner had an accepted model/provider policy for local artifacts, but
silent local-file egress to a hosted model would violate local-first and
runner-only expectations.

Do not promote local OCR dependencies from this path. Tesseract and OCRmyPDF
were unavailable in the local environment, and Poppler alone is not OCR. A
future local OCR track would need dependency installation, version reporting,
language data, page rendering, timeout, and fixture policy before product work.

Do not promote receipt/invoice-specific OCR fields. Field extraction would
overstate authority before generic reviewed OCR text, confidence, uncertainty,
correction, and duplicate behavior exist.

## Compatibility

Existing behavior remains unchanged:

- `artifact_candidate_plan` remains read-only and supports only explicit
  UTF-8 text, markdown, and text-bearing PDF local artifacts.
- Draft `artifact.text_extraction` remains unrecognized by the current runner.
- OCR/image parsing, scanned-PDF OCR fallback, and opaque binary parsing remain
  unsupported.
- Durable writes still require approved `create_document` or
  `ingest_source_url`.
- Committed docs and reports use repo-relative paths or neutral placeholders
  such as `<run-root>` and `<explicit-user-local-file>`.

## Follow-Up

No product implementation work item is filed.

Created broader prerequisite follow-up:

- `oc-hbu5`: compare platform policies for OCR-capable local artifact
  extraction.

`oc-hbu5` completed that policy comparison in
[`docs/architecture/local-artifact-ocr-platform-policy-decision.md`](local-artifact-ocr-platform-policy-decision.md).
It explicitly killed OpenClerk-owned OCR extraction for this path and filed no
product implementation work item.
