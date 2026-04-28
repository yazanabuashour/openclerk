---
decision_id: design-video-transcript-acquisition
decision_title: Video Transcript Acquisition Design
decision_status: accepted
decision_scope: video-youtube-ingestion
decision_owner: platform
---
# Design: Video Transcript Acquisition

## Status

Accepted as a coordinated design pass for future video transcript acquisition.
This note closes the current design questions for timestamp-span citations,
local downloader and platform caption retrieval, local STT, and remote
transcript APIs. It does not add runner behavior, media download, platform
caption retrieval, local STT, remote API calls, parser pipelines, dependency
installation, schema migrations, storage migrations, public API behavior, or
shipped skill behavior.

The implemented `ingest_video_url` v1 surface remains supplied-transcript only.
Future acquisition work requires a separate promotion decision and follow-up
implementation Beads before any behavior changes.

## Shared Provenance Envelope

Every future transcript acquisition path must produce one runner-visible
provenance envelope before it can create or update a canonical markdown source
note. The canonical source note remains the authority; raw media, platform
caption files, STT output, remote API responses, logs, and metadata sidecars
are supporting evidence only.

Required fields:

- `acquisition_policy`: one of `supplied`, `platform_caption_local`,
  `local_stt`, or `remote_transcript_api`
- `transcript_origin`: specific origin such as `user_supplied_transcript`,
  `platform_caption`, `local_stt`, or the configured remote provider name
- `source_url`: normalized video URL
- `captured_at`: transcript capture timestamp in RFC3339 format
- `transcript_sha256`: SHA256 of the transcript text stored in canonical
  markdown
- `language`: language code when known, omitted only when unavailable
- `tool`: downloader, caption extractor, STT runtime, or API client identity
  when a tool is used
- `model`: STT model or remote provider model when a model is used
- `provider`: remote provider identity for remote acquisition only
- `egress`: `none`, `video_platform`, or `remote_transcript_provider`
- `user_approval`: `not_required`, `required`, or `granted`
- `failure_classification`: `unsupported`, `dependency_unavailable`,
  `source_unavailable`, `no_captions`, `resource_limit`, `policy_rejected`,
  `provider_rejected`, `retryable_remote_failure`, or `internal_error`
- `retryable`: boolean value for the reported failure classification

Compatibility rules:

- Existing v1 `transcript.policy` and `transcript.origin` map into the envelope
  as `acquisition_policy: supplied` and `transcript_origin`.
- The envelope may be serialized into a future metadata sidecar, provenance
  event, and response payload, but it must not become independent authority.
- Update behavior must continue to report previous and new transcript hashes,
  same-hash no-op status, provenance events, and stale source-linked synthesis
  through existing projection freshness surfaces.
- Committed docs, reports, and examples must use repo-relative paths or neutral
  placeholders, not machine-absolute paths.

## Timestamp-Span Citations

Timestamp spans are an optional citation enrichment layered on top of existing
canonical markdown chunk citations. Existing `doc_id`, `chunk_id`, source path,
heading, and line-range citations remain valid and sufficient when timestamp
metadata is absent.

Timestamp parsing policy:

- Parse source timestamps only from transcript inputs that already expose
  explicit video-relative offsets, such as `00:00`, `00:00:15`,
  `00:00:15.250`, caption cue start/end offsets, or STT segment offsets.
- Normalize spans to video-relative offsets, not wall-clock capture time.
- Preserve the original timestamp token in supporting metadata when available,
  but index and compare normalized offsets.
- Reject negative, decreasing, malformed, or wall-clock-only timestamp spans
  for timestamp enrichment; keep the transcript citeable through normal chunk
  citations instead of rejecting the whole supplied transcript.

Citation span shape:

```json
{
  "doc_id": "source_doc_id",
  "chunk_id": "chunk_id",
  "path": "sources/video-youtube/example.md",
  "heading": "Transcript",
  "line_start": 24,
  "line_end": 28,
  "timestamp_start_ms": 15000,
  "timestamp_end_ms": 30250,
  "timestamp_label": "00:15-00:30.250"
}
```

Search and indexing behavior:

- Search continues to index canonical markdown chunks as the source of truth.
- When a chunk contains one or more normalized transcript spans, retrieval may
  include the best matching timestamp range alongside the chunk citation.
- Timestamp ranges must map back to canonical markdown lines and chunks. A
  timestamp span that cannot be mapped to a canonical chunk is ignored for
  citation enrichment.
- Timestamp ranges are not required to be unique across chunks, but a returned
  timestamp citation must identify exactly one canonical chunk and path.
- Source-sensitive final answers may mention timestamp ranges only together
  with the canonical source path, `doc_id`, `chunk_id`, or equivalent chunk
  citation evidence.

Targeted tests and evals before implementation:

- unit tests for supported timestamp token formats and malformed timestamp
  fallback
- retrieval tests proving timestamp citations preserve existing chunk citation
  fields
- update tests proving changed transcript hashes refresh timestamp mappings and
  stale synthesis state
- AgentOps eval rows for timestamp-compatible supplied transcripts, timestamp
  absence fallback, malformed timestamp fallback, and bypass rejection

## Local Downloader And Platform Captions

A future local caption acquisition mode may use a configured local downloader
to reach the video platform and retrieve metadata or platform captions. The
conceptual policy is `acquisition_policy: platform_caption_local`.

Supported conceptual modes:

- metadata probe: resolve platform metadata needed to identify a caption track
- caption list: inspect available caption tracks without selecting unsupported
  remote transcription
- caption fetch: retrieve the selected platform caption track and normalize it
  into canonical markdown transcript text

Dependency policy:

- Downloader tools such as `yt-dlp` are never invoked by routine agents outside
  the installed runner contract.
- The runner must report the downloader name and version in provenance when a
  downloader is used.
- If the configured downloader is missing, disabled, unsupported, or too old,
  the request must fail before writing canonical docs or assets with
  `failure_classification: dependency_unavailable`.
- Platform changes, missing captions, private video access failures, and
  unsupported caption formats must be visible failures, not hidden fallback to
  local STT or remote APIs.

Privacy and egress:

- Caption retrieval is local-first for processing, but it still contacts the
  video platform. Provenance must report `egress: video_platform`.
- Private URLs, cookies, credentials, and platform metadata must not be printed
  into committed docs or eval reports.
- Any future credential use must be configured explicitly and recorded as a
  policy choice without exposing secret values.

Targeted tests and evals before implementation:

- dependency-unavailable rejection without writes
- no-captions rejection without falling back to STT or remote APIs
- caption provenance envelope serialization without storing raw credentials
- timestamp mapping from caption cues to canonical chunks
- bypass rejection for direct `yt-dlp`, `ffmpeg`, or vault/SQLite workflows

## Local STT

A future local STT mode may transcribe locally available media after an
explicitly supported acquisition path has produced audio or media. The
conceptual policy is `acquisition_policy: local_stt`.

Dependency and resource policy:

- Local STT requires an explicitly configured runtime and model. The runner
  must report both in provenance through `tool` and `model`.
- Model weights must be installed through documented local configuration, not
  downloaded opportunistically during routine ingestion.
- Resource limits must be explicit before implementation, including maximum
  media duration, maximum asset size, timeout, concurrency, and CPU/GPU policy.
- When resource limits are exceeded, fail with
  `failure_classification: resource_limit` before writing canonical transcript
  docs.

Transcript and timestamp policy:

- STT output must preserve transcript text and, when available, segment
  timestamps mapped to canonical markdown chunks.
- Language detection may populate `language`, but unknown language must remain
  visible rather than guessed in final answers.
- Confidence values may be stored as supporting metadata, but low confidence
  cannot be hidden. A future implementation must define whether low confidence
  rejects ingestion or creates a source note with visible quality metadata.
- Media hashes, transcript hashes, model identity, and capture timestamp must
  be recorded in provenance. Raw media remains supporting evidence only.

Privacy policy:

- Local STT must not send media, transcript text, private URLs, or metadata to
  third-party services.
- If the configured STT runtime would perform network calls for model download,
  telemetry, or transcription, the request must be policy-rejected unless a
  future explicit remote policy covers that egress.

Targeted tests and evals before implementation:

- missing-runtime and missing-model rejection without writes
- resource-limit rejection without partial canonical docs
- provenance coverage for runtime, model, language, media hash, and transcript
  hash
- timestamp-segment mapping to chunk citations
- low-confidence policy behavior once the implementation decision chooses
  reject-or-visible-quality semantics

## Remote Transcript APIs

Remote transcript APIs and remote extraction are the least local-first option.
They may be considered only as explicit opt-in future policy, never as a hidden
fallback from local captions or local STT. The conceptual policy is
`acquisition_policy: remote_transcript_api`.

Approval and credential policy:

- Remote acquisition requires explicit user approval for the request and
  configured credentials for the provider.
- Missing approval or missing credentials must reject before any network call
  or write with `failure_classification: policy_rejected`.
- Provenance must record provider identity, provider API family when useful,
  model or extraction mode when used, request timestamp, egress posture, and
  transcript hash. It must never record secret credential values.
- Provider terms, quota failures, unsupported URLs, and moderation or access
  failures must be visible provider failures, not silent fallback to another
  provider.

Egress policy:

- Provenance must report `egress: remote_transcript_provider`.
- Final answers and committed eval reports may say that remote egress occurred
  and name the provider, but must not include private URLs, raw provider
  payloads, tokens, or machine-local paths.
- Remote APIs cannot become the default acquisition path for routine video URL
  ingestion.

Targeted tests and evals before implementation:

- no-approval rejection before network or writes
- missing-credentials rejection before network or writes
- provider rejection and retryable remote failure classification
- provenance coverage for provider, egress, approval, model, capture time, and
  transcript hash
- final-answer-only rejection for requests that ask agents to call remote APIs
  outside the installed runner contract

## Close Criteria For This Design Pass

This design pass is complete when documentation records:

- timestamp parsing, timestamp citation shape, chunk mapping, search behavior,
  backward compatibility, and test/eval obligations
- local downloader and platform caption modes, dependency policy, provenance,
  privacy gates, failure modes, and tests/evals
- local STT dependency policy, model/runtime policy, resource bounds, privacy,
  language/confidence metadata, failure modes, and tests/evals
- remote transcript API opt-in policy, credentials, provider identity, egress
  visibility, rejection behavior, provenance, and tests/evals

No acquisition implementation Bead should be filed from this note alone. A
future implementation Bead requires a promotion decision that names the exact
request shape, response shape, dependency setup, compatibility behavior,
failure modes, tests, eval gates, and AgentOps invariants.
