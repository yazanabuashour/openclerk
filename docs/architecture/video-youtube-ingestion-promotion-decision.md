---
decision_id: decision-video-youtube-ingestion-promotion
decision_title: Video And YouTube Ingestion Promotion
decision_status: accepted
decision_scope: video-youtube-ingestion
decision_owner: platform
---
# Decision: Video And YouTube Ingestion Promotion

## Status

Accepted and partially implemented: promote `openclerk document`
`ingest_video_url` for v1 supplied-transcript video and YouTube source
ingestion. The implemented v1 runner action creates and updates canonical
source notes only when transcript text is explicitly supplied. It does not
implement media download, platform caption retrieval, local STT, transcript
APIs, Gemini extraction, parser pipelines, dependency installation, storage
migrations, or native video acquisition.

Evidence:

- [`video-youtube-source-ingestion-adr.md`](video-youtube-source-ingestion-adr.md)
- [`video-transcript-acquisition-design.md`](video-transcript-acquisition-design.md)
- [`../evals/video-youtube-ingestion-toolchain-comparison-poc.md`](../evals/video-youtube-ingestion-toolchain-comparison-poc.md)
- [`../evals/video-youtube-canonical-source-note-pressure.md`](../evals/video-youtube-canonical-source-note-pressure.md)
- [`../evals/results/ockp-video-youtube-canonical-source-note.md`](../evals/results/ockp-video-youtube-canonical-source-note.md)

## Decision

Promote and implement a narrow `openclerk document` action named
`ingest_video_url` for supplied transcript text.

Capability path: current primitives can safely express canonical video source
notes only when transcript text and provenance are already supplied. The
scripted transcript control passed by creating a canonical markdown source note
and retrieving citation-bearing evidence through existing runners, so the
current document/retrieval model remains the authority baseline.

Ergonomics path: promote only the supplied-transcript runner surface for now.
The refreshed natural scenario can create a canonical source note through
`ingest_video_url` when transcript text is supplied. URL-only transcript
acquisition remains deferred because it still requires explicit downloader,
caption, STT, remote API, privacy, and provenance design.

Do not promote generalized `ingest_artifact`, arbitrary media import, hidden
transcript APIs, Gemini extraction as authority, OCR, local file import, or any
direct-vault/SQLite bypass from this decision.

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
      "text": "Supplied transcript text.",
      "policy": "supplied",
      "origin": "user_supplied_transcript",
      "language": "en",
      "captured_at": "2026-04-27T00:00:00Z"
    },
    "asset_path_hint": "assets/video-youtube/demo.json"
  }
}
```

Response shape includes:

- `doc_id`, `source_path`, `source_url`, `citations`, and captured transcript
  hash
- transcript provenance: origin, capture method, captured timestamp, language,
  tool or model identity, and whether transcript text was user-supplied,
  platform-supplied, or locally transcribed
- optional supporting metadata sidecar path for provenance metadata
- update result fields equivalent to source URL update behavior: previous hash,
  new hash, no-op same-hash status, and freshness/provenance effects

Compatibility rules:

- Existing `openclerk document`, `openclerk retrieval`, and PDF
  `ingest_source_url` behavior remain unchanged.
- Missing `video.mode` defaults to `create`; duplicate creates reject; update
  mode targets the normalized `video.url`; mismatched path or asset hints
  conflict without writing extra documents or assets.
- Canonical markdown source notes remain authority. Metadata JSON sidecars are
  supporting evidence only.
- Citation mapping exposes stable `doc_id`, `chunk_id`, source path, heading,
  and line range through the indexed canonical markdown note.

## Dependency And Privacy Policy

The v1 implementation is local-first by construction because transcript text is
supplied by the caller and no acquisition dependencies are invoked. Any future
dependency on `yt-dlp`, `ffmpeg`, local STT models, transcript APIs, or
Gemini-style extraction must be explicitly configured and reported in
provenance. No default path may send private URLs, media, transcript text, or
metadata to a third-party service.

Remote transcript APIs or LLM extraction may be considered only as explicit
future policy options with visible egress, credentials, provider identity,
failure modes, and user approval. They must never become hidden routine
fallbacks.

## Deferred Gates

The following gates remain deferred for later acquisition work:

- local downloader and caption retrieval policy
- local STT dependency and model policy
- remote transcript API and remote extraction policy
- richer timestamp-span citation mapping
- raw media storage policy, if ever needed

The coordinated design for the first four deferred gates is recorded in
[`video-transcript-acquisition-design.md`](video-transcript-acquisition-design.md).
That design preserves the current v1 boundary: supplied transcripts are
implemented, while native acquisition remains unsupported until a separate
promotion decision and implementation Beads name an exact surface.

If these gates cannot preserve authority, citations, provenance, freshness,
privacy, local-first operation, and the no-bypass contract, defer or kill the
implementation rather than shipping a weaker surface.
