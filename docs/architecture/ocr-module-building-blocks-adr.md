---
decision_id: adr-ocr-module-building-blocks
decision_title: OCR Module Building Blocks
decision_status: accepted
decision_scope: ocr-module-final-decision-track
decision_owner: agentops
decision_date: 2026-05-05
source_refs: docs/architecture/agent-knowledge-plane.md, docs/architecture/semantic-retrieval-building-blocks.md, docs/architecture/local-artifact-ocr-platform-policy-decision.md, docs/evals/ocr-module-final-candidate-comparison.md, https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md, https://mitchellh.com/writing/building-block-economy, https://developers.openai.com/api/docs/guides/prompt-guidance, https://openai.com/index/harness-engineering/, https://developers.openai.com/api/docs/guides/images-vision, https://developers.openai.com/api/docs/guides/retrieval, https://docs.mem0.ai/open-source/overview
---
# ADR: OCR Module Building Blocks

## Status

Accepted as the final comparison frame for `oc-n286`.

This ADR does not add OCR runtime behavior, module manifests, runner fields,
storage migrations, public APIs, product code, shipped skill behavior, model
egress, local artifact parsing, or durable writes.

## Decision

If OCR is ever implemented in OpenClerk, it must use the optional building-block
module pattern rather than core-owned hidden extraction.

The natural product surface remains the existing read-only
`artifact_candidate_plan` local artifact planning action with an explicit OCR
review mode:

```json
{"action":"artifact_candidate_plan","artifact":{"local_path":"<explicit-user-local-file>","artifact_kind":"receipt","text_extraction":"ocr_review","limit":5}}
```

That shape stays unpromoted until a later decision proves an OCR provider can
meet all gates. OCR output would be candidate evidence only. Durable writes
would still require approved `create_document` or `ingest_source_url`.

The shared module contract would be `openclerk_artifact_ocr.v1`:

- extracted text or markdown preview
- artifact kind and page/image references
- extractor identity, version, model, provider, and local/cloud posture
- confidence, uncertainty, and warnings where available
- duplicate-search text used for candidate suppression
- explicit no-write handoff
- rejection details for unsupported files, missing dependencies, timeouts,
  oversize inputs, missing credentials, and cloud-egress denial

The module authority boundary would mirror semantic retrieval modules:

- installed, enabled, and manifest-verified before use
- read-only authority with durable writes forbidden
- redacted provider config and credential references only
- no hidden local or remote fallback
- no committed OCR cache or extracted private artifact text
- no direct SQLite, vault, module-cache, source-built runner, or ad hoc parser
  bypass for routine agents

## Candidate Frame

Evaluate OCR as separately installable parts, not as a core parser stack:

| Candidate | Role | Initial outcome |
| --- | --- | --- |
| Tesseract image OCR module | Local-first image text extraction. | Compare, but do not promote without installed dependency, language data, version reporting, fixtures, and confidence behavior. |
| OCRmyPDF/Tesseract scanned-PDF module | Local-first searchable PDF and scanned-PDF text layer path. | Compare, but do not promote without dependency and page-rendering policy. |
| Go OCR bindings such as gosseract | Go adapter over native OCR. | Compare as packaging shape only; it does not remove the Tesseract dependency. |
| PaddleOCR module | Open-source OCR/layout model path. | Compare for capability; do not promote without packaging, runtime, model asset, and fixture proof. |
| Ollama vision module | Local multimodal OCR through installed Ollama. | Compare only when a vision model is installed and fixture quality is proven. |
| OpenAI vision module | Explicit hosted multimodal OCR. | Compare only with configured credentials, model identity, retention/egress posture, and private-local-file approval policy. |
| Cloud document OCR module | Managed OCR such as Mistral OCR, Textract, Azure Document Intelligence, or Google Document AI/Vision. | Keep as benchmark pressure until credentials, egress, audit, and provider-specific provenance are proven. |
| Agent-model OCR outside OpenClerk | User or agent obtains reviewed text elsewhere. | Supported as external supplied text only; not an OpenClerk OCR module. |

## Rationale

The Agent Knowledge Plane keeps canonical markdown, citations, provenance, and
freshness as the authority layer. OCR is an acquisition aid, not a new truth
store.

Karpathy's LLM Wiki pattern supports durable source-linked knowledge that
compounds over time, but does not require the wiki host to own every extraction
engine. The building-block economy framing fits OCR especially well: Tesseract,
OCRmyPDF, local multimodal models, hosted vision models, and cloud document OCR
are different blocks with different dependency, cost, privacy, and maintenance
profiles.

OpenAI prompt and harness guidance points toward explicit, testable contracts,
not prompt recipes hidden in skill prose. OpenAI vision guidance confirms
models can analyze image inputs, while retrieval guidance reinforces that
derived representations and search tools are supporting infrastructure, not
authority. Mem0 remains a useful reference for modular agent memory and
multimodal capability, but OpenClerk should not adopt an opaque memory or media
truth layer for OCR.

## Compatibility

Existing behavior remains unchanged:

- `artifact_candidate_plan` supports explicit supplied text, markdown, and
  text-bearing PDF planning.
- `text_extraction` remains unrecognized by the runner.
- OCR/image parsing, scanned-PDF OCR fallback, and opaque binary parsing remain
  unsupported.
- No OCR module manifests are supported or installable from this ADR.
