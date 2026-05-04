---
decision_id: iteration-artifact-intake-autofiling-tags-fields
status: implemented
decision_scope: artifact-intake
source_refs: docs/architecture/artifact-intake-autofiling-tags-fields-promotion-decision.md, docs/evals/results/ockp-artifact-intake-autofiling-tags-fields.md
---

# Iteration: Artifact Intake, Auto-Filing, Tags, and Fields

## Required References

- docs/architecture/agent-knowledge-plane.md
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Implemented Shape

`artifact_candidate_plan` now gives agents one read-only candidate-planning
surface for common artifact intake:

- explicit content to faithful markdown preview
- path/title autofiling
- artifact kind and source type
- tags and metadata fields
- duplicate search and likely target
- existing source URL detection
- confidence and confidence reasons
- approved `create_document` or `ingest_source_url` handoff

## Iteration Notes

Tag taxonomy is conservative: explicit tags are preserved first, then
`artifact-intake` and the artifact kind are appended. Metadata fields are
frontmatter-shaped strings; explicit fields override inferred fields.

Duplicate evidence starts with runner lexical search scoped to the candidate
path prefix, plus source URL duplicate detection for public-source handoff. It
does not claim semantic duplicate certainty.

Domain templates are intentionally small:

- invoices -> `artifacts/invoices/`
- receipts -> `artifacts/receipts/`
- legal documents -> `artifacts/legal/`
- transcripts -> `artifacts/transcripts/`
- source summaries -> `sources/candidates/`
- note-like artifacts -> `notes/candidates/`

## Future Candidate Comparisons

Future work may compare:

- multi-tag frontmatter conventions versus current single `tag` authority
- richer invoice/receipt/legal/transcript field templates from explicit content
  only
- duplicate scoring explanations beyond first lexical hit

None of those are required to close this promoted surface.
