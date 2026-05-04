---
decision_id: adr-generalized-artifact-ingestion
decision_title: Generalized Artifact Ingestion Direction
decision_status: accepted
decision_scope: artifact-ingestion
decision_owner: platform
---
# ADR: Generalized Artifact Ingestion Direction

## Status

Accepted as an evidence-gathering direction only. This ADR does not promote a
new runner action, parser, storage schema, public API, migration, or artifact
pipeline.

Supporting evidence:

- [`../evals/artifact-ingestion-architecture-options-poc.md`](../evals/artifact-ingestion-architecture-options-poc.md)
- [`../evals/results/ockp-heterogeneous-artifact-ingestion-pressure.md`](../evals/results/ockp-heterogeneous-artifact-ingestion-pressure.md)
- [`generalized-artifact-ingestion-promotion-decision.md`](generalized-artifact-ingestion-promotion-decision.md)

Required references:

- [`agent-knowledge-plane.md`](agent-knowledge-plane.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/embeddings>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

## Context

OpenClerk already supports local-first AgentOps workflows through the installed
`openclerk document` and `openclerk retrieval` JSON runners. The only native
source ingestion action today is `ingest_source_url`, which downloads an
HTTP/HTTPS PDF source URL into a canonical source note and a vault-relative
asset path, and after `oc-v1ed` can also fetch public HTML/web pages into
canonical markdown source notes.

The next artifact question is broader: whether OpenClerk should generalize
source ingestion beyond PDF URLs to videos and YouTube links, transcripts,
invoices, receipts, and mixed artifact sets while preserving the existing
knowledge-plane invariants.

Candidate artifact classes:

- PDF source URLs with canonical `sources/*.md` notes and `assets/**/*.pdf`
  assets
- videos and YouTube links that may require transcript extraction, metadata,
  media assets, and source citation mapping
- transcript text already supplied as canonical markdown under
  `transcripts/`
- invoices and receipts stored as canonical markdown documents, with optional
  asset references when raw artifacts exist
- mixed artifact sets that combine sources, transcripts, invoices, receipts,
  and synthesis
- unknown future artifacts such as images, slide decks, emails, exported chats,
  or forms

## Decision

Keep canonical markdown and promoted canonical records as the authority model.
Native artifact bytes, parsed metadata, citations, provenance events,
projection states, and source-linked synthesis remain supporting or derived
evidence. No parser output, media metadata, OCR result, transcript extraction,
or generalized artifact index may outrank canonical markdown or promoted
records unless a later decision explicitly promotes a typed domain.

Use the following runner-surface options as the comparison frame for POC and
eval evidence:

- **Explicit user-provided content:** accept user-supplied title/body or
  transcript text through existing document actions. This is enough when the
  user has already extracted faithful content and can approve a durable
  markdown write.
- **Local artifact registry:** consider a runner-visible asset registry only if
  artifact identity, duplicate detection, and asset provenance cannot be
  represented by current markdown plus asset paths.
- **Parser/OCR candidate extraction:** consider local-first parser or OCR
  candidates only as proposed extracted text with source provenance, confidence
  posture, unsupported-file behavior, and approval before any canonical record
  write.
- **Keep `ingest_source_url` only:** preserve the PDF and public web URL
  ingestion contract and model other non-PDF artifacts as canonical markdown
  or source-linked synthesis.
- **Add artifact-specific actions:** consider actions such as
  `ingest_video_url`, `ingest_transcript`, or `ingest_receipt` only if targeted
  evidence shows artifact-specific validation, parsing, provenance, or citation
  semantics cannot be expressed through existing runners.
- **Add generalized `ingest_artifact`:** consider a single action with
  artifact `kind`, URI or local reference, path hints, asset hints, mode, and
  metadata only if repeated evidence shows a shared ingestion abstraction is
  safer than action-specific surfaces.
- **Keep as skill/eval reference:** use artifact pressure to harden guidance
  and eval coverage when existing document/retrieval workflows are sufficient.

Public read/fetch/inspect permission is not durable-write approval. A public
URL can be fetched through the existing runner after approval, and explicit
user-provided text can be inspected as candidate content, but parser/OCR output,
local artifact metadata, and extracted records remain candidates until the user
approves the canonical markdown or promoted-record write.

Promotion requires targeted AgentOps eval evidence showing that existing
`openclerk document` and `openclerk retrieval` workflows are structurally
insufficient, not merely verbose, underdocumented, or missing fixture data.

## Invariants

- Routine agents use only the installed `openclerk document` and
  `openclerk retrieval` JSON runners.
- Canonical markdown source docs and promoted records remain authority.
- Source-sensitive claims preserve citations, source paths, `doc_id`,
  `chunk_id`, source refs, or equivalent stable identifiers.
- Parser/OCR output must carry source provenance and cannot become canonical
  without approval.
- Unsupported file kinds must reject or plan explicitly; they must not fall
  back to direct local file reads, OCR bypasses, or opaque artifact parsing.
- Duplicate handling must compare source URLs, asset hints, canonical paths,
  and runner-visible metadata before creating new records.
- Provenance and projection freshness remain inspectable for source updates,
  derived records, and source-linked synthesis.
- Missing required source-ingestion fields clarify without tools; invalid
  limits and bypass requests reject final-answer-only.
- Committed reports and artifact references use repo-relative paths or neutral
  placeholders such as `<run-root>`.

## Non-Goals

This ADR does not:

- commit OpenClerk to generalized artifact ingestion
- add `ingest_artifact` or any artifact-specific runner action
- add OCR, video, YouTube, audio transcription, invoice parsing, receipt
  parsing, or media download pipelines
- add migrations, new storage schemas, background workers, queues, or asset
  registries
- relax `ingest_source_url` validation or update semantics
- authorize routine direct SQLite, direct vault inspection, broad repo search,
  source-built runner paths, HTTP/MCP bypasses, unsupported transports,
  backend variants, module-cache inspection, or ad hoc import scripts
- use machine-absolute paths in committed docs, reports, or artifact references

## Promotion Gate

Use the deferred capability rubric in
[`deferred-capability-promotion-gates.md`](deferred-capability-promotion-gates.md).
Promote only after the POC and targeted eval record repeated
`runner_capability_gap` failures and the promotion decision names the exact
surface, request/response shape, compatibility rules, failure modes, and
follow-up implementation Beads.

Safety, capability, and UX quality remain separate gates:

- Safety pass requires runner-only access, source provenance for extracted
  text, duplicate handling, unsupported-file behavior, local-first parsing, and
  approval before durable records are written.
- Capability pass requires repeated proof that explicit user-provided content,
  current source URL ingestion, and canonical markdown workflows cannot express
  the artifact need.
- UX quality pass requires reducing real artifact workflow ceremony without
  hiding parser uncertainty, provenance, or durable-write approval.

Remaining work is represented by linked beads:

- `oc-tnnw.5.2` POC for artifact/OCR candidate evidence.
- `oc-tnnw.5.3` eval for safety, capability, and UX quality.
- `oc-tnnw.5.4` promotion decision.
- `oc-tnnw.5.5` conditional implementation only if promoted.
- `oc-tnnw.5.6` iteration and follow-up bead creation.
