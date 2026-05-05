# OCR Module Final Candidate Comparison Result

## Summary

`oc-n286` compared OCR as optional OpenClerk building blocks rather than core
artifact parsing. The final implementation outcome is a promoted optional
local `tesseract-ocr` module.

This result keeps the prior OpenClerk-owned hidden OCR extraction kill decision
intact for default core behavior while implementing explicit
`artifact_candidate_plan` OCR review through an installed, enabled, and
manifest-verified optional module.

## Local Probe Result

| Check | Result |
| --- | --- |
| Tesseract | available; `tesseract 5.5.2` |
| Tesseract languages | `eng`, `osd`, `snum` |
| OCRmyPDF | available; `ocrmypdf 17.4.2` |
| Poppler PDF helpers | available |
| Ollama | available |
| Ollama vision model | unavailable; installed models are embedding-only |
| Cloud OCR key | unavailable for this run |

Temporary synthetic fixture extraction was performed for an image and scanned
PDF containing `Receipt ID OC-7781` and `Total paid 42 USD`. The text was
recovered through Tesseract image OCR and OCRmyPDF sidecar text. No model call,
cloud egress, private artifact upload, durable write, direct SQLite access,
vault mutation, or module-cache inspection was performed.

## Decision Result

| Candidate family | Result |
| --- | --- |
| Tesseract/OCRmyPDF local OCR | Promote as `modules/tesseract-ocr` for explicit local OCR review of common images and PDFs. |
| Go OCR bindings | Kill as standalone answer. Bindings do not remove native dependency and language-data burden. |
| PaddleOCR/open-source model OCR | Keep as reference pressure. Capability is plausible but packaging/runtime evidence is missing. |
| Ollama vision OCR | Do not promote now. Ollama exists, but no installed vision model or fixture evidence exists. |
| OpenAI or agent-model vision OCR | Keep as benchmark pressure unless routed through a future accepted provider/egress policy. |
| Cloud document OCR | Keep as benchmark pressure until credentials, egress, audit, and fixture proof exist. |
| External OCR followed by supplied text | Supported by existing supplied-content workflow. |

## Gate Results

Safety pass: pass. Current default behavior preserves runner-only access,
local-first default behavior, no hidden local parser truth, no hidden cloud
egress, unsupported-file rejection, duplicate checks, and
approval-before-write. OCR adds only explicit local module execution with
`planned_no_write`.

Capability pass: pass for generic OCR candidate text from common image files
and scanned or force-OCR PDFs. Text-extractable documents continue to use the
normal non-OCR path.

UX quality: pass. A normal user gets a simpler surface: no OCR for documents
with embedded text, explicit OCR review for scan-only images/PDFs or bad PDF
text, and no new durable-write ceremony.

## Implementation Authorization

Implement only:

- `modules/tesseract-ocr/module.json`
- optional module install/config/list/remove support for `kind:
  "ocr_provider"` and provider `tesseract`
- `artifact_candidate_plan` fields `text_extraction: "ocr_review"` and
  `ocr_provider: "tesseract"`
- local Tesseract image OCR and OCRmyPDF scanned-PDF or force-PDF OCR
- returned `ocr_extraction` provenance, versions, language, warnings, privacy
  posture, and `planned_no_write`

Do not implement cloud OCR, hosted model OCR, Ollama vision OCR, PaddleOCR,
structured OCR fields, storage migrations, or OCR caches from this result.
