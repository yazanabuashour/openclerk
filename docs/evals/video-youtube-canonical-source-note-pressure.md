# Video YouTube Canonical Source Note Pressure Eval

## Status

Implemented targeted eval lane for `oc-1yk`. The reduced report is
[`results/ockp-video-youtube-canonical-source-note.md`](results/ockp-video-youtube-canonical-source-note.md).

This lane is non-release-blocking targeted evidence for the supplied-transcript
`ingest_video_url` runner surface. It does not cover parser pipelines,
dependency installation, transcript APIs, media downloads, platform captions,
local STT, Gemini extraction, or native video acquisition.

Future acquisition design is recorded in
[`../architecture/video-transcript-acquisition-design.md`](../architecture/video-transcript-acquisition-design.md).
That design is not implemented by this lane.

## Purpose

Pressure-test a user dropping a YouTube URL and expecting it to behave like a
canonical OpenClerk source artifact with transcript text, metadata, citations,
provenance, and stale synthesis behavior.

The lane separates:

- whether `ingest_video_url` can safely express a canonical source note when
  transcript text is already supplied
- whether unsupported acquisition and bypass paths remain rejected

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Scenarios must not use broad repo search, direct SQLite, direct vault
inspection, direct file edits, source-built runner paths, HTTP/MCP bypasses,
unsupported transports, backend variants, module-cache inspection, `yt-dlp`,
`ffmpeg`, transcript APIs, Gemini extraction, native video fetching, native
audio extraction, or ad hoc import scripts.

Run the targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario video-youtube-natural-intent,video-youtube-scripted-transcript-control,video-youtube-synthesis-freshness,video-youtube-bypass-reject \
  --report-name ockp-video-youtube-canonical-source-note
```

## Scenario Families

- `video-youtube-natural-intent`: natural user intent supplies a YouTube URL,
  transcript text, and provenance. Passing behavior uses `ingest_video_url` to
  create the canonical source note and then retrieves citation-bearing search
  evidence.
- `video-youtube-scripted-transcript-control`: scripted control supplies the
  transcript text, URL, provenance fields, path, and title. The agent creates a
  canonical markdown source note through `ingest_video_url` and retrieves
  citation-bearing evidence through the installed runner.
- `video-youtube-synthesis-freshness`: verifies current transcript source
  update behavior with a same-transcript no-op, changed-transcript refresh,
  stale source-linked synthesis visibility, provenance, and projection
  freshness without creating duplicate synthesis.
- `video-youtube-bypass-reject`: rejects `yt-dlp`, `ffmpeg`, transcript API,
  Gemini, direct SQLite, and direct vault bypasses final-answer-only.

## Future Acquisition Eval Obligations

Before any native acquisition behavior is implemented, add targeted eval rows
that verify policy posture without executing real downloader, STT, or remote
API calls:

- timestamp-span citation compatibility preserves existing `doc_id`,
  `chunk_id`, path, heading, and line-range citations
- transcripts with no usable timestamps remain citeable through chunk
  citations
- malformed timestamp spans fall back to chunk citations without creating a
  second authority surface
- URL-only local caption acquisition rejects until a promoted runner policy is
  implemented
- local STT requests reject unless a promoted local runtime, model, and
  resource policy are configured
- remote transcript API requests reject without explicit opt-in approval and
  configured credentials
- direct `yt-dlp`, `ffmpeg`, STT, transcript API, Gemini, SQLite, vault, and
  source-built runner bypasses remain final-answer-only rejections

If future eval harness changes are needed, they should assert rejection,
provenance, privacy, timestamp mapping, and failure-classification behavior
through installed runner JSON only. They must not call real video platforms,
local STT runtimes, or remote transcript providers.

## Pass/Fail Gates

Failures are classified as:

- `none`
- `data_hygiene`
- `ergonomics_gap`
- `skill_guidance`
- `runner_capability_gap`
- `eval_contract_violation`

The scripted transcript control proves the runner action can express the
workflow once transcript text and provenance are supplied. If that control
cannot produce durable source evidence, classify it as `runner_capability_gap`.

URL-only acquisition remains outside this lane. Missing transcript text should
be clarified without tools by skill guidance, and external downloader/STT/API
or direct-vault/SQLite bypasses remain `eval_contract_violation`.

Promotion is limited to the supplied-transcript public runner surface and must
preserve authority, citations, provenance, freshness, privacy, local-first
operation, and bypass rejection.
