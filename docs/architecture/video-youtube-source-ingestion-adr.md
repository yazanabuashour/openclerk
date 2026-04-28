---
decision_id: adr-video-youtube-source-ingestion
decision_title: Video And YouTube Source Ingestion
decision_status: accepted
decision_scope: video-youtube-ingestion
decision_owner: platform
---
# ADR: Video And YouTube Source Ingestion

## Status

Accepted as an evidence-gathering direction for video and YouTube source
ingestion. This ADR does not add a runner action, parser, downloader,
transcriber, storage schema, migration, public API, or shipped skill behavior.

Supporting evidence:

- [`../evals/video-youtube-ingestion-toolchain-comparison-poc.md`](../evals/video-youtube-ingestion-toolchain-comparison-poc.md)
- [`../evals/video-youtube-canonical-source-note-pressure.md`](../evals/video-youtube-canonical-source-note-pressure.md)
- [`../evals/results/ockp-video-youtube-canonical-source-note.md`](../evals/results/ockp-video-youtube-canonical-source-note.md)
- [`video-youtube-ingestion-promotion-decision.md`](video-youtube-ingestion-promotion-decision.md)

## Context

OpenClerk already supports local-first AgentOps workflows through the installed
`openclerk document` and `openclerk retrieval` JSON runners. The promoted
native source ingestion action is `ingest_source_url`, limited to HTTP/HTTPS
PDF source URLs.

Video and YouTube links differ from PDF source URLs because the canonical text
usually depends on a transcript acquisition step. That step may come from a
platform transcript, a local media download plus STT model, a supplied
transcript, or a separate extraction tool. Each option affects source
authority, citation granularity, privacy, dependency surface, repeatability,
and freshness.

This track evaluates both promotion paths from the deferred capability gates:

- **Capability gap:** whether current document and retrieval actions cannot
  safely express canonical video source notes even when transcript text is
  supplied.
- **Ergonomics gap:** whether current actions can express the workflow only
  with too many tools, too much prompt choreography, unacceptable latency, or
  unsafe dependency choices for routine users who drop a video URL.

## Decision

Keep canonical markdown and promoted canonical records as authority. A video
or YouTube transcript becomes OpenClerk authority only after it is represented
as a canonical markdown source note with provenance fields that identify where
the transcript came from and how it was captured. Raw media, transcript API
responses, downloaded metadata, model output, and parser traces are supporting
evidence only.

Current public surface remains:

- `openclerk document`
- `openclerk retrieval`
- existing `ingest_source_url` for PDF source URLs

Until a later promotion decision adds a native video surface, routine agents
must not fetch video media, run `yt-dlp`, run `ffmpeg`, call transcript APIs,
invoke Gemini extraction, use source-built runners, inspect SQLite directly,
or inspect vault files directly for video ingestion. They may use existing
document and retrieval actions for transcript text already supplied by the user
or already present as canonical markdown.

## Required Source Model

Any future video ingestion surface must preserve these properties:

- **Authority:** canonical markdown source notes outrank raw transcript assets,
  tool output, media metadata, and synthesis pages.
- **Transcript provenance:** the source note records the video URL, transcript
  origin, capture method, capture timestamp, language when known, tool or model
  identity when used, and whether transcript text was user-supplied,
  platform-supplied, or locally transcribed.
- **Citation shape:** source-sensitive answers cite stable `doc_id`,
  `chunk_id`, source paths, headings, line ranges, timestamp ranges, or an
  equivalent stable mapping from transcript spans to canonical markdown chunks.
- **Asset storage:** any raw media, sidecar transcript, metadata JSON, or model
  output must be optional supporting evidence under vault-relative asset paths;
  assets cannot become a second truth surface.
- **Privacy:** routine ingestion must be local-first by default and must not
  send media, transcript text, or private URLs to third-party transcript or LLM
  APIs unless the user explicitly chooses that policy in a future promoted
  surface.
- **Freshness:** update behavior must expose changed transcript hashes,
  provenance events, and stale source-linked synthesis through
  `projection_states` and `provenance_events`.

## Non-Goals

This ADR does not:

- promote `ingest_video_url`, `ingest_transcript`, or `ingest_artifact`
- add media downloads, audio extraction, OCR, STT, transcript API calls, Gemini
  extraction, background workers, queues, or asset registries
- define production dependency installation for `yt-dlp`, `ffmpeg`, Whisper,
  Gemini, or transcript APIs
- relax `ingest_source_url` validation or PDF update behavior
- authorize routine direct SQLite, direct vault inspection, broad repo search,
  source-built command paths, HTTP/MCP bypasses, unsupported transports,
  backend variants, module-cache inspection, or ad hoc runtime programs
- use machine-absolute paths in committed docs, reports, or artifact references

## Promotion Gate

Use the deferred capability rubric in
[`deferred-capability-promotion-gates.md`](deferred-capability-promotion-gates.md).
Promote only after targeted POC and eval evidence shows repeated capability-gap
or ergonomics-gap pressure and a promotion decision names the exact request
shape, response shape, dependency policy, privacy model, transcript provenance
contract, citation mapping, update behavior, failure modes, compatibility
rules, and follow-up implementation Beads.
