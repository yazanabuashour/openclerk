# OCR Module Final Candidate Comparison

## Status

Implemented final candidate comparison for `oc-n286`.

This document does not add runner actions, OCR dependencies, provider calls,
module manifests, storage behavior, public APIs, product behavior, shipped
skill behavior, or implementation authorization.

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
| Tesseract | Not present on `PATH`. | Local image OCR cannot be promoted from this environment. |
| OCRmyPDF | Not present on `PATH`. | Scanned-PDF OCR cannot be promoted from this environment. |
| Poppler helpers | Present on `PATH`. | Useful for text-bearing PDFs or rendering, but not OCR by itself. |
| Ollama | Present on `PATH`. | Local model hosting is plausible. |
| Ollama installed models | Embedding-only models present: `nomic-embed-text`, `bge-m3`, `mxbai-embed-large`, `embeddinggemma`. | No local vision OCR row can be promoted without pulling and proving a vision model. |
| Cloud OCR credentials | No OCR provider key configured for this decision. | Cloud OCR remains benchmark pressure, not implementation evidence. |

No local artifact text was extracted, no private file was sent to a model, no
cloud call was made, and no durable write occurred.

## Candidate Rows

| Candidate | Safety | Capability | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Tesseract image OCR module | Strong local-first posture in principle, but dependency, language data, version reporting, confidence, and packaging are unproven here. | Plausible for images, but not installed and not fixture-proven. | Familiar and offline, but setup burden is visible. | Do not promote. |
| OCRmyPDF/Tesseract scanned-PDF module | Strong local-first posture in principle, but heavier dependency chain and page-rendering policy are unresolved. | Plausible for scanned PDFs, but not installed and not fixture-proven. | Good for searchable PDFs, too heavy for default core. | Do not promote. |
| Go OCR bindings | Does not remove native dependency risk; Go bindings still need Tesseract libraries and language data. | Packaging adapter only, not an independent OCR engine. | Worse than a manifest-declared binary dependency until proven. | Kill as standalone answer. |
| PaddleOCR module | Open-source and likely stronger for layout/document parsing, but model assets, Python/C++ runtime, packaging, resource profile, and fixture behavior are not proven. | Plausible; not locally installed or evaluated. | Potentially good capability, high module complexity. | Keep as reference pressure. |
| Ollama vision module | Local-first and matches existing building-block taste, but no vision model is installed and OCR reliability is unproven. | Plausible only after a vision model is pulled and tested. | Better setup story than Tesseract only if model quality is good. | Do not promote now. |
| OpenAI vision module | Clear hosted vision capability, but private-local-file egress, credential storage, retention posture, and approval policy are not accepted for OpenClerk OCR. | Plausible for quality and mixed layouts. | Best simplicity if explicitly configured, but surprising as default. | Keep as benchmark pressure. |
| Cloud document OCR module | Mature managed OCR options exist, including Mistral OCR, AWS Textract, Azure Document Intelligence, and Google Document AI/Vision. | Plausible for quality and structure, but no key, provider policy, or fixture evidence exists here. | Strong convenience, high egress/account setup burden. | Keep as benchmark pressure. |
| Agent model does OCR outside OpenClerk | Safe when the user supplies reviewed text back to OpenClerk. | Works outside the product boundary; OpenClerk cannot audit extractor identity or provenance. | Useful escape hatch. | Support as supplied text only. |
| Current rejection plus supplied text | Passes existing no-hidden-parser, no-egress, runner-only, duplicate, and approval-before-write boundaries. | Partial: no OCR recovery. | Acceptable only as a final product boundary for current evidence. | Select. |

## Required Gates Not Met

No OCR module candidate currently satisfies all gates:

- installed and versioned extractor identity
- fixture-proven image and scanned-PDF extraction
- page/image references and text-span provenance
- confidence, uncertainty, warning, and correction behavior
- duplicate search from reviewed OCR text
- explicit no-write handoff through existing durable actions only
- missing dependency, unsupported file, timeout, oversize, credential, and
  cloud-egress rejection behavior
- local-first default with no hidden remote fallback
- no committed OCR cache, model artifact, or private extracted text
- acceptable natural-prompt UX without skill bloat or exact prompt choreography

## Outcome

Select `none viable yet` for OCR module implementation.

The user need remains valid, and the best future shape remains optional OCR
modules behind `artifact_candidate_plan` OCR review. Current evidence does not
authorize module manifests, runner schema changes, provider implementation,
OCR dependencies, model calls, storage changes, or skill behavior.
