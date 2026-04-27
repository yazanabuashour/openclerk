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

## Context

OpenClerk already supports local-first AgentOps workflows through the installed
`openclerk document` and `openclerk retrieval` JSON runners. The only native
source ingestion action today is `ingest_source_url`, which downloads an
HTTP/HTTPS PDF source URL into a canonical source note and a vault-relative
asset path.

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

- **Keep `ingest_source_url` only:** preserve the current PDF URL ingestion
  contract and model non-PDF artifacts as canonical markdown or source-linked
  synthesis.
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

Promotion requires targeted AgentOps eval evidence showing that existing
`openclerk document` and `openclerk retrieval` workflows are structurally
insufficient, not merely verbose, underdocumented, or missing fixture data.

## Invariants

- Routine agents use only the installed `openclerk document` and
  `openclerk retrieval` JSON runners.
- Canonical markdown source docs and promoted records remain authority.
- Source-sensitive claims preserve citations, source paths, `doc_id`,
  `chunk_id`, source refs, or equivalent stable identifiers.
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
