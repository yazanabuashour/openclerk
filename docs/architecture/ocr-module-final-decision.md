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

Accepted as the final decision for `oc-n286`.

Do not add OCR module manifests, OCR provider adapters, runtime OCR, multimodal
model calls, scanned-PDF OCR fallback, image parsing, Tesseract, OCRmyPDF,
PaddleOCR, Ollama vision, OpenAI vision, Mistral OCR, Textract, Azure Document
Intelligence, Google Cloud OCR, Go OCR bindings, parser pipelines, storage
migrations, public APIs, shipped skill behavior, or product implementation work
from this decision.

Evidence:

- [`docs/architecture/ocr-module-building-blocks-adr.md`](ocr-module-building-blocks-adr.md)
- [`docs/evals/ocr-module-final-candidate-comparison.md`](../evals/ocr-module-final-candidate-comparison.md)
- [`docs/evals/results/ockp-ocr-module-final-candidate-comparison.md`](../evals/results/ockp-ocr-module-final-candidate-comparison.md)
- [`docs/architecture/local-artifact-ocr-platform-policy-decision.md`](local-artifact-ocr-platform-policy-decision.md)
- [`docs/architecture/semantic-retrieval-building-blocks.md`](semantic-retrieval-building-blocks.md)

## Decision

Select `none viable yet` for OCR implementation.

The evaluated shape is better than the old core-OCR path: if OCR is ever
implemented, it should be an optional building-block module family routed
through explicit `artifact_candidate_plan` OCR review. That future shape would
preserve OpenClerk's markdown authority, runner-only operation, local-first
default, citations or source refs, provenance, duplicate handling, and
approval-before-write.

No current candidate passes enough gates to authorize implementation. Local
OCR dependencies are not installed and not fixture-proven. Go OCR bindings do
not remove the native OCR dependency burden. Ollama is installed, but only
embedding models are present. Hosted and cloud OCR options are plausible, but
credentials, egress approval, retention, audit, and provider provenance are
not accepted for private local artifacts in this track.

## Safety, Capability, UX

Safety pass: pass for non-promotion. Existing production behavior preserves
runner-only access, no hidden parser truth, no hidden cloud egress,
unsupported-file rejection, duplicate checks for supplied text, and
approval-before-write.

Capability pass: partial. OpenClerk still does not recover OCR text from
scanned images or scanned PDFs. It safely supports reviewed text supplied by
the user, UTF-8 text, markdown, and text-bearing PDF local artifact planning.

UX quality: acceptable for a final boundary from current evidence. A normal
user would reasonably expect OCR review to be simpler than manual external OCR,
and optional modules remain the right taste direction. Implementation is still
not justified without a passing provider, dependency, egress, provenance,
confidence, duplicate, correction, and fixture story.

## Compatibility

Existing behavior remains unchanged:

- `artifact_candidate_plan` remains read-only and supports only explicit
  UTF-8 text, markdown, and text-bearing PDF local artifacts.
- `artifact.text_extraction` remains unrecognized by the current runner.
- OCR/image parsing, scanned-PDF OCR fallback, and opaque binary parsing remain
  unsupported.
- Durable writes still require approved `create_document` or
  `ingest_source_url`.
- External OCR or agent-model reading outside OpenClerk may produce reviewed
  text that the user supplies through existing supported surfaces.

## Follow-Up

No product implementation Bead is filed.

Searches before closing:

- `bd search "OCR module"`: found only `oc-n286`.
- `bd search "artifact_candidate_plan OCR review"`: no existing follow-up.

Created gated follow-up:

- `oc-ab1h`: reopen OCR only with installed or credentialed candidate fixture
  evidence.

The need remains valid, but the current evidence does not justify
implementation. `oc-ab1h` is not product implementation authorization. It is a
guarded reopen condition for a future evidence track with an installed local
OCR candidate or configured cloud/vision provider, fixture results, provenance,
confidence, duplicate handling, egress posture, and approval-before-write.
