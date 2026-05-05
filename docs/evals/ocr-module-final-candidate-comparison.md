# OCR Module Final Candidate Comparison

## Status

Implemented final candidate comparison for `oc-n286`, updated with the
`oc-6eqz` local OCR module evidence.

This document authorizes only the local optional Tesseract/OCRmyPDF module
shape. It does not authorize hidden core OCR, cloud or hosted model OCR,
storage changes, durable writes, OCR caches, or parser truth.

Governing references:

- [`docs/architecture/agent-knowledge-plane.md`](../architecture/agent-knowledge-plane.md)
- [`docs/architecture/semantic-retrieval-building-blocks.md`](../architecture/semantic-retrieval-building-blocks.md)
- [`docs/architecture/local-artifact-ocr-platform-policy-decision.md`](../architecture/local-artifact-ocr-platform-policy-decision.md)
- [`docs/architecture/ocr-module-building-blocks-adr.md`](../architecture/ocr-module-building-blocks-adr.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/images-vision>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

## Environment Probe

Current local dependency state:

| Probe | Result | Decision impact |
| --- | --- | --- |
| Tesseract | Present on `PATH`; `tesseract 5.5.2`. | Local image OCR can be fixture-tested and module-promoted. |
| Tesseract languages | `eng`, `osd`, `snum`. | The promoted module defaults to English OCR review. |
| OCRmyPDF | Present on `PATH`; `ocrmypdf 17.4.2`. | Scanned-PDF OCR can be fixture-tested and module-promoted. |
| Poppler helpers | Present on `PATH`. | Useful for text-bearing PDFs or rendering, but not OCR by itself. |
| Ollama | Present on `PATH`. | Local model hosting is plausible. |
| Ollama installed models | Embedding-only models present: `nomic-embed-text`, `bge-m3`, `mxbai-embed-large`, `embeddinggemma`. | No local vision OCR row can be promoted without pulling and proving a vision model. |
| Cloud OCR credentials | No OCR provider key configured for this decision. | Cloud OCR remains benchmark pressure, not implementation evidence. |

Synthetic fixture extraction was performed in a temporary directory, not
committed. `tesseract` extracted text from an image fixture containing
`Receipt ID OC-7781` and `Total paid 42 USD`. `ocrmypdf --sidecar
--force-ocr` extracted the same receipt text from a scanned PDF generated from
the image. No private file was sent to a model, no cloud call was made, and no
durable write occurred.

## Candidate Rows

| Candidate | Safety | Capability | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Tesseract image OCR module | Passes local-first, no-network, manifest verification, version reporting, no-write, and rejection-boundary gates. | Extracted the image fixture text through explicit OCR review. | Good as an optional module; too dependency-heavy for default core. | Promote as `modules/tesseract-ocr`. |
| OCRmyPDF/Tesseract scanned-PDF module | Passes local-first, no-network, manifest verification, version reporting, no-write, and sidecar-text gates. | Extracted the scanned-PDF fixture text through explicit OCR review. | Good for scan-only and force-OCR PDFs; too dependency-heavy for default core. | Promote as `modules/tesseract-ocr`. |
| Go OCR bindings | Does not remove native dependency risk; Go bindings still need Tesseract libraries and language data. | Packaging adapter only, not an independent OCR engine. | Worse than a manifest-declared binary dependency until proven. | Kill as standalone answer. |
| PaddleOCR module | Open-source and likely stronger for layout/document parsing, but model assets, Python/C++ runtime, packaging, resource profile, and fixture behavior are not proven. | Plausible; not locally installed or evaluated. | Potentially good capability, high module complexity. | Keep as reference pressure. |
| Ollama vision module | Local-first and matches existing building-block taste, but no vision model is installed and OCR reliability is unproven. | Plausible only after a vision model is pulled and tested. | Better setup story than Tesseract only if model quality is good. | Do not promote now. |
| OpenAI vision module | Clear hosted vision capability, but private-local-file egress, credential storage, retention posture, and approval policy are not accepted for OpenClerk OCR. | Plausible for quality and mixed layouts. | Best simplicity if explicitly configured, but surprising as default. | Keep as benchmark pressure. |
| Cloud document OCR module | Mature managed OCR options exist, including Mistral OCR, AWS Textract, Azure Document Intelligence, and Google Document AI/Vision. | Plausible for quality and structure, but no key, provider policy, or fixture evidence exists here. | Strong convenience, high egress/account setup burden. | Keep as benchmark pressure. |
| Agent model does OCR outside OpenClerk | Safe when the user supplies reviewed text back to OpenClerk. | Works outside the product boundary; OpenClerk cannot audit extractor identity or provenance. | Useful escape hatch. | Support as supplied text only. |
| Current rejection plus supplied text | Passes existing no-hidden-parser, no-egress, runner-only, duplicate, and approval-before-write boundaries. | Partial: no OCR recovery. | Useful fallback when no OCR module is installed. | Keep supported. |

## Gate Results

The promoted local module satisfies the implementation gates for generic
candidate text:

- installed and versioned extractor identity for `tesseract` and `ocrmypdf`
- fixture-proven image and scanned-PDF extraction
- explicit provider, module name, language, page count, provenance, privacy
  posture, warnings, and no-write handoff in `ocr_extraction`
- duplicate search and confidence calculation operate on reviewed OCR candidate
  text before any durable write
- missing module, disabled/unverified module, unsupported file, timeout,
  oversize, and extraction-empty cases reject through runner validation
- local-first default with no hidden remote fallback
- no committed OCR cache, model artifact, private extracted text, or storage
  migration

Known limits remain accepted for this module: OCR confidence is coarse,
page-span provenance is not yet text-span-level, handwriting/layout quality is
not claimed, and structured receipt or invoice fields are out of scope.

## Outcome

Select `modules/tesseract-ocr` for implementation as a supported optional
local OCR module.

Default local artifact planning remains text/markdown/text-bearing-PDF
extraction without OCR. Explicit `artifact.text_extraction: "ocr_review"` is
the OCR route for common images, scan-only PDFs, and PDFs whose embedded text
is bad or partial. OCR output is candidate evidence only until the user
approves an existing durable write action.
