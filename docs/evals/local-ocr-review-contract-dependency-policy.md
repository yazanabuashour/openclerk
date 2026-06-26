# Local OCR Review Contract And Dependency Policy

## Status

Implemented contract/dependency framing for `oc-i8yk`.

This document does not add a runner action, OCR dependency, model provider,
parser pipeline, storage behavior, public API, product behavior, shipped skill
behavior, or implementation authorization.

Governing evidence:

- [`docs/evals/results/ockp-local-ocr-artifact-fixture-evidence.md`](results/ockp-local-ocr-artifact-fixture-evidence.md)
- [`docs/evals/local-ocr-artifact-candidate-planning-poc.md`](local-ocr-artifact-candidate-planning-poc.md)
- [`docs/architecture/local-ocr-artifact-candidate-planning-decision.md`](../architecture/local-ocr-artifact-candidate-planning-decision.md)
- [`docs/architecture/local-artifact-candidate-plan-promotion-decision.md`](../architecture/local-artifact-candidate-plan-promotion-decision.md)

## Baseline

`artifact_candidate_plan` is the promoted local artifact planning surface for
explicit UTF-8 text, markdown, and text-bearing PDF local files. Image OCR,
scanned-PDF OCR fallback, opaque parsing, hidden file inspection, direct vault
or SQLite access, and durable writes before approval remain unsupported.

The simplest future surface is still an explicit review mode on the existing
read-only action:

```json
{"action":"artifact_candidate_plan","artifact":{"local_path":"<explicit-user-local-file>","artifact_kind":"receipt","text_extraction":"ocr_review","limit":5}}
```

OCR output must remain candidate evidence. It is not canonical knowledge until
a user reviews or corrects it and approves an existing durable write action.

## Candidate Workflows

| Candidate | Shape | Strength | Risk |
| --- | --- | --- | --- |
| Model-assisted OCR review | Use a configured multimodal model as the OCR reviewer inside `artifact_candidate_plan`, returning extractor identity, page/image refs, extracted text preview, confidence, uncertainty, duplicate evidence, and no-write handoff. | Simplest product shape because the current eval harness already records multimodal-capable agent models such as `gpt-5.4-mini`, so a later pass can test whether a configured model is enough before adding OCR binaries. | Hosted models may violate local-first expectations unless egress is explicit and approved. Model output can be overconfident, non-deterministic, or hard to calibrate without review fields and fixture gates. |
| Local OCR engine or library | Use a configured local dependency such as Tesseract, OCRmyPDF plus Tesseract/Poppler, or a Go binding around a local OCR engine. | Best default for strict local-first operation. Tesseract is established for image OCR; OCRmyPDF is a strong scanned-PDF wrapper; Poppler is useful for PDF rendering/text helpers. | Adds installation, version, page rendering, language-pack, CGO or shell dependency, resource-limit, and platform support burden. Poppler alone is not OCR. |
| Scanned-PDF OCR fallback | Keep PDF handling inside `artifact_candidate_plan`; when current PDF text extraction returns no text, allow explicit OCR review to render pages and extract candidate text. | Targets the clearest current gap after text-bearing PDF promotion. | Still needs page refs, rendering policy, confidence, correction, duplicate behavior, and missing-dependency errors. |
| Domain-specific receipt/invoice OCR | Add receipt/invoice field extraction after a generic OCR review envelope exists. | Could produce high-value vendor/date/total fields later. | Premature now; field extraction can hide parser/model authority and should not precede generic review and correction behavior. |
| `none viable yet` | Keep OCR unsupported until a target dependency and contract pass promotion evidence. | Safest current production behavior. | Leaves valid user pressure unresolved; normal users expect scanned receipts and PDFs to become reviewable candidates. |

## Future Contract

A later promotion decision may name these fields only after targeted evidence
proves they are sufficient:

- request: `artifact.text_extraction: "ocr_review"` as explicit opt-in
- request: optional configured extraction mode such as `model_multimodal` or
  `local_ocr`, never a hidden fallback
- response: `ocr_review` envelope with extractor kind, extractor identity,
  version or model id, egress posture, page/image refs, text spans, confidence,
  uncertainty notes, correction status, and unsupported reason when rejected
- response: duplicate search derived from reviewed candidate text, with
  `next_create_document_request` suppressed on likely duplicate or unreviewed
  low-confidence OCR
- response: existing `planned_no_fetch`, `planned_no_write`,
  approval-before-write, validation-boundary, authority-limit, and
  `agent_handoff` conventions preserved

The correction workflow should be simple: low-confidence or disputed OCR text
returns no next-create handoff until the user supplies corrected text or
explicitly confirms the reviewed candidate body. Durable writes still use
`create_document` or `ingest_source_url`.

## Dependency Policy

Preferred next evidence candidate: model-assisted OCR review, because it is the
simplest surface to evaluate and avoids committing OpenClerk to a heavy OCR
binary before proving the review contract.

Local-first default candidate: configured local OCR engine. If strict offline
operation is required, the implementation should prefer local dependencies and
must report missing runtime, missing language data, unsupported file type, page
limit, timeout, and extraction failure without writing.

Hosted or harness model candidate: allowed only as an explicit configured mode
with visible egress posture, model identity, and no durable write from model
output before review. The runner must not silently send private local artifacts
to a remote provider.

## Required Evidence

Before any product implementation work item is filed, a later eval must include:

- natural and scripted prompts for receipt image, scanned receipt PDF, and
  low-confidence OCR review
- duplicate-risk fixture proving next-create suppression before update-versus-new
  approval
- unsupported-file, missing-dependency, oversized-file, and bypass rejection
  rows
- safety pass, capability pass, and UX quality recorded separately
- extractor identity, provenance refs, confidence, uncertainty, correction
  status, duplicate status, write status, and approval boundary in the final
  report

## Outcome

Select `artifact_candidate_plan` OCR review with model-assisted extraction as
the first candidate for future promotion evidence, not implementation. Keep
local OCR dependencies as the required local-first comparison.

Runtime OCR remains unsupported until a later accepted promotion decision
proves the exact request/response surface and dependency policy.
