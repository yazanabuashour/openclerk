---
decision_id: decision-ocr-module-final
decision_title: OCR Module Final Decision
decision_status: accepted
decision_scope: ocr-module-final-decision-track
decision_owner: agentops
decision_date: 2026-05-05
source_refs: docs/architecture/ocr-module-building-blocks-adr.md, docs/evals/ocr-module-final-candidate-comparison.md, docs/evals/results/ockp-ocr-module-final-candidate-comparison.md, docs/architecture/local-artifact-ocr-platform-policy-decision.md, docs/architecture/semantic-retrieval-building-blocks.md
---
# Decision: OCR Module Final

## Status

Accepted as the final decision for `oc-n286` and implementation track
`oc-6eqz`.

Promote one OCR implementation shape: a supported optional local
`tesseract-ocr` module for explicit read-only `artifact_candidate_plan` OCR
review. Do not add hidden core OCR, hidden scanned-PDF fallback, cloud/model
egress, OCR caches, parser truth, storage migrations, or durable writes from
OCR output.

Evidence:

- [`docs/architecture/ocr-module-building-blocks-adr.md`](ocr-module-building-blocks-adr.md)
- [`docs/evals/ocr-module-final-candidate-comparison.md`](../evals/ocr-module-final-candidate-comparison.md)
- [`docs/evals/results/ockp-ocr-module-final-candidate-comparison.md`](../evals/results/ockp-ocr-module-final-candidate-comparison.md)
- [`docs/architecture/local-artifact-ocr-platform-policy-decision.md`](local-artifact-ocr-platform-policy-decision.md)
- [`docs/architecture/semantic-retrieval-building-blocks.md`](semantic-retrieval-building-blocks.md)

## Decision

Select the optional local Tesseract/OCRmyPDF module for implementation.

OpenClerk now differentiates ordinary text-extractable local artifacts from
scan-only or unreliable artifacts:

- UTF-8 text, markdown, and text-bearing PDFs use the normal local artifact
  extraction path and do not need OCR.
- Common image files and PDFs can be explicitly routed through OCR by setting
  `artifact.text_extraction` to `ocr_review` and `artifact.ocr_provider` to
  `tesseract`.
- Explicit `ocr_review` is an override for PDFs as well as a path for scan-only
  PDFs, so a caller can bypass bad or partial embedded PDF text when review
  requires OCR-derived text.

The promoted module is local-first and manifest-verified:
`modules/tesseract-ocr/module.json` declares `tesseract` plus `ocrmypdf`, no
network, no durable writes, no SQLite/vault bypass, and candidate-only output.
The runner stores only module enabled state, manifest digest, command, command
args, and redacted provider config in `runtime_config`.

Other candidates are not promoted. Go OCR bindings are killed as a standalone
answer because they do not remove native dependency and language-data burden.
PaddleOCR, Ollama vision, OpenAI vision, Mistral OCR, Textract, Azure Document
Intelligence, and Google OCR remain reference or future-provider candidates
until credentials, egress, retention, audit, packaging, and fixture gates pass.

## Safety, Capability, UX

Safety pass: pass for the promoted local module shape. OCR runs only after the
module is installed, enabled, and manifest-verified. The runner exposes
extractor identity, versions, language, provenance, local-only privacy posture,
warnings, duplicate status, and `planned_no_write`. Durable writes still
require approved `create_document` or `ingest_source_url`.

Capability pass: pass for generic candidate text from common image files and
scanned or force-OCR PDFs. The implementation does not claim structured
receipt/invoice fields, layout truth, handwriting quality, or canonical text.

UX quality: pass for a local-first optional module. A normal user can keep
text-extractable documents on the simpler existing path and request OCR only
for scan-only or suspect PDFs/images. The surface is the existing
`artifact_candidate_plan` action rather than a new workflow or hidden parser.

## Compatibility

Compatibility:

- Default `artifact_candidate_plan` behavior remains read-only and supports
  explicit supplied content, UTF-8 text, markdown, and text-bearing PDF local
  artifacts without OCR.
- OCR is available only for explicit `artifact.text_extraction:
  "ocr_review"` with the installed `tesseract` OCR module.
- Image parsing and scan-only PDF parsing remain unsupported without that
  explicit OCR review request and verified module.
- Durable writes still require approved `create_document` or
  `ingest_source_url`.
- External OCR or agent-model reading outside OpenClerk remains supported only
  as reviewed supplied text unless a future provider module passes gates.

## Follow-Up

`oc-ab1h` is satisfied by `oc-6eqz`: the local OCR candidate is installed in
the evidence environment, fixture-tested for image and scanned-PDF extraction,
and implemented as an optional module. Future work should compare cloud,
vision, layout, or domain-field OCR only as separate provider tracks.
