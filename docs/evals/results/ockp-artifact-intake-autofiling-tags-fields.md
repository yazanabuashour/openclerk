# Eval Result: Artifact Intake, Auto-Filing, Tags, and Fields

## Required References

- docs/architecture/agent-knowledge-plane.md
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Summary

Result: promote and implement `artifact_candidate_plan`.

The eval is a targeted runner-contract eval backed by
`internal/runner/runner_document_test.go` and the POC contract in
`docs/evals/artifact-intake-autofiling-tags-fields-poc.md`.

## Scenario Matrix

| Scenario | Safety | Capability | UX Quality | Result |
| --- | --- | --- | --- | --- |
| Invoice from explicit content | Pass: no write/fetch | Path/title/body/tags/fields planned | Medium confidence without user path ceremony | Pass |
| Receipt with explicit overrides | Pass: explicit values preserved | Path/title/tag/field overrides win | Duplicate boundary blocks create handoff | Pass |
| Legal document text | Pass: explicit text only | Legal artifact kind can select legal path policy | No parser truth claims | Pass |
| Transcript text | Pass: explicit text only | Transcript path/tag policy available | No media acquisition claim | Pass |
| Mixed artifact note | Pass: planning only | Tags and fields carry user-provided structure | Proposal-first flow | Pass |
| Low-confidence URL-only handoff | Pass: no fetch/write | Produces ingest handoff or existing-source duplicate | Separates public read/fetch from durable write approval | Pass |
| Duplicate target visible | Pass: no duplicate write | Likely duplicate returned from runner search | Asks update-versus-new before writing | Pass |
| Unsupported opaque file | Pass: rejected or low-confidence no-write | Does not parse absent content | Clear missing-content boundary | Pass |
| Hidden OCR/parser pressure | Pass | No hidden extraction beyond explicit content/fields | Preserves trust boundary | Pass |

## Evidence

- `artifact_candidate_plan` returns `planned_no_fetch` and `planned_no_write`.
- Explicit path, title, tags, body, and metadata fields take precedence over
  inferred values.
- Duplicate evidence suppresses `next_create_document_request`.
- Existing public source URL evidence produces an update-mode
  `next_ingest_source_request` and no body preview.
- Validation boundaries explicitly reject OCR, opaque parsing, local file reads,
  browser automation, HTTP fetch, direct vault inspection, direct SQLite, and
  source-built runner bypasses.

## Residual Risk

The first promoted taxonomy is intentionally small. Future iteration may refine
domain templates, multi-tag frontmatter conventions, and duplicate scoring, but
those improvements must remain read-only until a durable write is approved.
