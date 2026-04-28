---
decision_id: decision-video-youtube-ingestion-promotion
decision_title: Video And YouTube Ingestion Promotion
decision_status: accepted
decision_scope: video-youtube-ingestion
decision_owner: platform
---
# Decision: Video And YouTube Ingestion Promotion

## Status

Accepted: promote follow-up surface design for native video and YouTube source
ingestion through `ingest_video_url`. This decision does not implement the
runner action, parser, downloader, transcriber, dependency installation,
schema, storage migration, or public API behavior.

Evidence:

- [`video-youtube-source-ingestion-adr.md`](video-youtube-source-ingestion-adr.md)
- [`../evals/video-youtube-ingestion-toolchain-comparison-poc.md`](../evals/video-youtube-ingestion-toolchain-comparison-poc.md)
- [`../evals/video-youtube-canonical-source-note-pressure.md`](../evals/video-youtube-canonical-source-note-pressure.md)
- [`../evals/results/ockp-video-youtube-canonical-source-note.md`](../evals/results/ockp-video-youtube-canonical-source-note.md)

## Decision

Promote a follow-up implementation design for a narrow `openclerk document`
action named `ingest_video_url`.

Capability path: current primitives can safely express canonical video source
notes only when transcript text and provenance are already supplied. The
scripted transcript control passed by creating a canonical markdown source note
and retrieving citation-bearing evidence through existing runners, so the
current document/retrieval model remains the authority baseline.

Ergonomics path: promote follow-up design. The natural URL-only scenario
correctly rejected native video ingestion with no tools, but the targeted lane
classified that safe rejection as `ergonomics_gap`: a user dropping a YouTube
URL still cannot get the expected canonical source note without manually
orchestrating transcript acquisition, provenance capture, citation mapping, and
freshness checks. The scripted and freshness controls show the desired
authority and freshness model, but they are too procedural for routine URL-only
source ingestion.

The promoted implementation follow-up is unblocked only for the exact surface
below. Do not promote generalized `ingest_artifact`, arbitrary media import,
hidden transcript APIs, Gemini extraction as authority, OCR, local file import,
or any direct-vault/SQLite bypass from this decision.

## Promoted Surface Contract

Candidate request shape:

```json
{
  "action": "ingest_video_url",
  "video": {
    "url": "https://youtube.example.test/watch?v=demo",
    "path_hint": "sources/video-youtube/demo.md",
    "title": "Demo Video Transcript",
    "mode": "create",
    "transcript": {
      "policy": "local_first",
      "origin": "platform_caption_or_local_transcription",
      "language": "en"
    },
    "asset_path_hint": "assets/video-youtube/demo.json"
  }
}
```

Candidate response shape must include:

- `doc_id`, `source_path`, `source_url`, `citations`, and captured transcript
  hash
- transcript provenance: origin, capture method, captured timestamp, language,
  tool or model identity, and whether transcript text was user-supplied,
  platform-supplied, or locally transcribed
- optional supporting asset path and asset hash for metadata, sidecar
  transcript, or media-derived evidence
- update result fields equivalent to source URL update behavior: previous hash,
  new hash, no-op same-hash status, and freshness/provenance effects

Compatibility rules:

- Existing `openclerk document`, `openclerk retrieval`, and PDF
  `ingest_source_url` behavior remain unchanged.
- Missing `video.mode` defaults to `create`; duplicate creates reject; update
  mode targets the normalized `video.url`; mismatched path or asset hints
  conflict without writing extra documents or assets.
- Canonical markdown source notes remain authority. Raw media, transcript API
  responses, model output, metadata JSON, and downloaded assets are supporting
  evidence only.
- Citation mapping must expose stable `doc_id`, `chunk_id`, source path,
  heading, line range, timestamp range, or an equivalent stable span.

## Dependency And Privacy Policy

The implementation must be local-first by default. Any dependency on `yt-dlp`,
`ffmpeg`, local STT models, transcript APIs, or Gemini-style extraction must be
explicitly configured and reported in provenance. No default path may send
private URLs, media, transcript text, or metadata to a third-party service.

Remote transcript APIs or LLM extraction may be considered only as explicit
future policy options with visible egress, credentials, provider identity,
failure modes, and user approval. They must never become hidden routine
fallbacks.

## Implementation Gates

The follow-up implementation must add:

- runner request/response types and validation for `ingest_video_url`
- transcript acquisition policy that rejects unsupported dependency modes
  clearly
- canonical markdown source-note creation with transcript provenance fields
- duplicate, update, same-hash, changed-transcript, missing transcript,
  unsupported URL, dependency failure, parser failure, and partial-success
  behavior
- citation mapping from transcript spans to indexed chunks
- provenance events and projection invalidation for changed transcripts
- targeted tests and eval scenarios proving natural URL-only ingestion,
  scripted controls, update/freshness behavior, privacy rejection, and bypass
  rejection

If these gates cannot preserve authority, citations, provenance, freshness,
privacy, local-first operation, and the no-bypass contract, defer or kill the
implementation rather than shipping a weaker surface.
