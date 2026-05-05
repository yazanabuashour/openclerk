# OCR Module Final Candidate Comparison Result

## Summary

`oc-n286` compared OCR as optional OpenClerk building blocks rather than core
artifact parsing. The final implementation outcome is `none viable yet`.

This result keeps the prior OpenClerk-owned OCR extraction kill decision intact
for core behavior while recording the only acceptable future shape: explicit
`artifact_candidate_plan` OCR review routed through installed, enabled, and
manifest-verified optional OCR modules.

## Local Probe Result

| Check | Result |
| --- | --- |
| Tesseract | unavailable |
| OCRmyPDF | unavailable |
| Poppler PDF helpers | available |
| Ollama | available |
| Ollama vision model | unavailable; installed models are embedding-only |
| Cloud OCR key | unavailable for this run |

No extraction, model call, cloud egress, private artifact upload, durable write,
module registration, direct SQLite access, vault mutation, or module-cache
inspection was performed.

## Decision Result

| Candidate family | Result |
| --- | --- |
| Tesseract/OCRmyPDF local OCR | Do not promote. Local-first posture is good, but dependencies and fixture quality are unproven here. |
| Go OCR bindings | Kill as standalone answer. Bindings do not remove native dependency and language-data burden. |
| PaddleOCR/open-source model OCR | Keep as reference pressure. Capability is plausible but packaging/runtime evidence is missing. |
| Ollama vision OCR | Do not promote now. Ollama exists, but no installed vision model or fixture evidence exists. |
| OpenAI or agent-model vision OCR | Keep as benchmark pressure unless routed through a future accepted provider/egress policy. |
| Cloud document OCR | Keep as benchmark pressure until credentials, egress, audit, and fixture proof exist. |
| External OCR followed by supplied text | Supported by existing supplied-content workflow. |

## Gate Results

Safety pass: pass for non-promotion. Current production behavior preserves
runner-only access, local-first default behavior, no hidden local parser truth,
no hidden cloud egress, duplicate checks for supplied text, unsupported-file
rejection, and approval-before-write.

Capability pass: partial. OpenClerk still does not recover text from scanned
images or scanned PDFs. It can safely handle user-supplied reviewed text,
markdown, UTF-8 text, and text-bearing PDF candidates.

UX quality: acceptable as a final boundary for current evidence. A normal user
would expect a simpler OCR surface, and optional OCR modules remain the best
taste fit, but no candidate currently passes enough dependency, provenance,
egress, confidence, duplicate, and fixture gates to justify implementation.

## Implementation Authorization

Do not implement:

- OCR module manifests
- OCR provider adapters
- `artifact_candidate_plan.text_extraction`
- scanned-PDF OCR fallback
- Tesseract, OCRmyPDF, Poppler rendering, PaddleOCR, Ollama vision, OpenAI
  vision, Mistral OCR, Textract, Azure, or Google OCR integrations
- storage migrations or OCR caches
- shipped skill behavior
- public API changes

Future OCR work must start from a new evidence track where at least one module
candidate is installed or credentialed, fixture-tested, and accepted by a
promotion decision before product code is written.
