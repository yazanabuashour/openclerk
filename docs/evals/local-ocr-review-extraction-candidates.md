# Local OCR Review Extraction Candidates

## Status

Implemented candidate-evidence framing for `oc-s3wg`.

This document does not add a runner action, OCR dependency, model provider,
parser pipeline, storage behavior, public API, product behavior, shipped skill
behavior, or implementation authorization.

Governing evidence:

- [`docs/evals/local-ocr-review-contract-dependency-policy.md`](local-ocr-review-contract-dependency-policy.md)
- [`docs/evals/results/ockp-local-ocr-review-contract-dependency-policy.md`](results/ockp-local-ocr-review-contract-dependency-policy.md)
- [`docs/evals/results/ockp-local-ocr-artifact-fixture-evidence.md`](results/ockp-local-ocr-artifact-fixture-evidence.md)
- [`docs/architecture/local-ocr-review-contract-dependency-policy-decision.md`](../architecture/local-ocr-review-contract-dependency-policy-decision.md)

## Baseline Probes

The current installed runner behavior remains intentionally non-OCR:

| Probe | Result | Evidence posture |
| --- | --- | --- |
| `artifact_candidate_plan` with draft `text_extraction: "ocr_review"` | Rejected at JSON decode as an unknown field. | Confirms the future OCR contract is not accidentally live. |
| `artifact_candidate_plan` with explicit local PNG artifact | Rejected: OCR/image parsing is unsupported. | Confirms current unsupported-file behavior remains intact. |
| Local dependency availability | Tesseract unavailable; OCRmyPDF unavailable; Poppler PDF helper present. | Poppler can help render or inspect PDFs but is not OCR by itself. |

No scanned-PDF OCR extraction, image OCR extraction, model call, local OCR
binary invocation, or durable write was run.

## Candidate Rows

| Candidate | Safety | Capability | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Model-assisted OCR review | Fails production promotion today. It is simplest for users, but hosted or harness model use needs a broader provider/egress policy before private local artifacts can be sent to a model. | Plausible, especially for mixed image/PDF layouts, but no runner-owned file-to-model path, confidence calibration, correction contract, or duplicate proof exists. | Best taste if explicitly configured and review-first. | Keep as reference pressure only. |
| Local OCR with Tesseract | Partial. Local-first in principle, but local runtime, language data, version reporting, timeout policy, and installation behavior are unresolved. | Plausible for images, but unavailable in the current local environment and unproven against fixtures. | Worse setup taste than model-assisted OCR. | Do not promote. |
| Scanned-PDF OCR with OCRmyPDF/Tesseract/Poppler | Partial. Best local-first scanned-PDF shape, but requires multiple external tools and page-rendering policy. | Plausible for scanned PDFs, but unavailable in the current local environment and unproven against fixtures. | Useful but heavy. | Do not promote. |
| Poppler-only PDF fallback | Passes local-first, but does not provide OCR. | Not sufficient; current text-bearing PDF extraction already covers the non-OCR case. | No simpler OCR surface. | Kill as OCR candidate. |
| Go-native OCR binding | Partial. A Go wrapper may still hide native libraries, CGO, language data, and platform-specific failures. | Unproven and not simpler than a declared local dependency. | Poor until dependency behavior is visible. | Do not promote. |
| Domain-specific receipt/invoice OCR | Fails promotion before generic OCR review exists. Field extraction would overstate parser or model authority. | Useful only after reviewed OCR text is safe. | Potential later convenience. | Keep future-only outside this path. |
| Current rejection plus supplied text | Passes. Keeps runner-only access, no hidden parser truth, no egress, duplicate behavior for supplied text, and approval-before-write. | Partial: does not recover OCR text, but safely supports pasted or explicitly supplied content and current text/PDF planning. | Acceptable as a durable safety boundary for this path. | Select durable non-promotion. |

## Required Gates Not Met

No candidate satisfied all product gates:

- exact promoted request/response fields
- runner-owned file-to-extractor path
- local-first default or explicit egress approval policy
- extractor identity and version/model provenance
- page/image refs and text-span provenance
- confidence, uncertainty, and correction status
- duplicate search from reviewed OCR text with next-create suppression
- unsupported-file and missing-dependency behavior
- approval-before-write through existing durable actions

## Outcome

Select durable non-promotion for this OCR/scanned-PDF artifact-candidate path.

Do not add product OCR, model OCR, scanned-PDF fallback, Tesseract, OCRmyPDF,
Poppler rendering, Go OCR bindings, receipt/invoice OCR fields, schema changes,
skill behavior, or public API changes from this path. Future OCR work should
open a new decision track only if a broader runner-owned media/provider policy
or a concrete local OCR runtime policy already exists.
