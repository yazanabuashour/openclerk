---
decision_id: decision-local-artifact-ocr-platform-policy
decision_title: Local Artifact OCR Platform Policy
decision_status: accepted
decision_scope: artifact-candidate-plan-ocr-platform-policy
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/local-artifact-ocr-platform-policy-comparison.md, docs/evals/results/ockp-local-artifact-ocr-platform-policy-comparison.md, docs/architecture/local-ocr-review-extraction-candidate-decision.md
---
# Decision: Local Artifact OCR Platform Policy

## Status

Accepted as the explicit kill decision for `oc-hbu5`.

Supersession note: this historical kill decision was superseded on 2026-05-05
by [`ocr-module-final-decision.md`](ocr-module-final-decision.md), which keeps
hidden OpenClerk-owned core OCR killed but promotes explicit optional local OCR
review through `modules/tesseract-ocr/module.json`.

Do not add OpenClerk-owned OCR-capable local artifact extraction, model/provider
egress over local artifacts, local OCR runtime management, scanned-PDF OCR
fallback, Tesseract, OCRmyPDF, Poppler rendering, Go OCR bindings,
domain-specific OCR fields, parser pipelines, storage migrations, public APIs,
shipped skill behavior, or product implementation work from this decision.

Evidence:

- [`docs/evals/local-artifact-ocr-platform-policy-comparison.md`](../evals/local-artifact-ocr-platform-policy-comparison.md)
- [`docs/evals/results/ockp-local-artifact-ocr-platform-policy-comparison.md`](../evals/results/ockp-local-artifact-ocr-platform-policy-comparison.md)
- [`docs/architecture/local-ocr-review-extraction-candidate-decision.md`](local-ocr-review-extraction-candidate-decision.md)

## Decision

Select no OpenClerk-owned OCR extraction.

The local OCR/scanned-PDF artifact-candidate path is closed by an explicit kill
decision: OpenClerk will not own OCR-capable local artifact extraction in this
path. External OCR or multimodal reading may be used outside OpenClerk; the
reviewed text can then enter OpenClerk through existing supplied-content and
candidate-planning surfaces.

## Rejected Alternatives

Do not promote runner-owned model/provider egress for local artifacts. It would
need a broader provider, egress, credential, retention, audit, and private
artifact approval model before local files could be sent to a hosted or harness
model.

Do not promote a runner-owned local OCR runtime policy. It would add dependency
installation, language data, version reporting, page rendering, timeouts,
platform support, and maintenance burden that is disproportionate to the
current artifact-candidate path.

## Safety, Capability, UX

Safety pass: pass. The selected policy preserves runner-only access,
local-first behavior, no hidden OCR/model/parser truth, no hidden local-file
egress, unsupported-file rejection, current duplicate handling, and
approval-before-write.

Capability pass: pass for the product boundary and partial for OCR as a
feature. OpenClerk does not recover OCR text; it safely handles supplied
reviewed text and current text/markdown/text-bearing PDF candidates.

UX quality: acceptable. This is less convenient than model-assisted OCR, but it
is clearer and safer than adding hidden provider egress or heavyweight local
OCR runtime management to the current product.

## Compatibility

Existing behavior remains unchanged:

- `artifact_candidate_plan` remains read-only and supports only explicit
  UTF-8 text, markdown, and text-bearing PDF local artifacts.
- OCR/image parsing, scanned-PDF OCR fallback, and opaque binary parsing remain
  unsupported.
- Durable writes still require approved `create_document` or
  `ingest_source_url`.
- Committed docs and reports use repo-relative paths or neutral placeholders.

## Follow-Up

No product implementation work item is filed.

No additional OCR/scanned-PDF artifact-candidate follow-up is filed from this
decision. Any future OCR work must start as a new product direction, not as a
continuation of this path.
