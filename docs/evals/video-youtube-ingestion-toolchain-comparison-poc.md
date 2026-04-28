# Video And YouTube Ingestion Toolchain Comparison POC

## Status

Implemented POC framing for `oc-sxu`. This document compares candidate
toolchains and evidence obligations only. It does not add runner actions,
schemas, migrations, parser pipelines, dependency installation, public API
behavior, or shipped skill behavior.

The governing ADR is
[`../architecture/video-youtube-source-ingestion-adr.md`](../architecture/video-youtube-source-ingestion-adr.md).
Follow-up acquisition design is recorded in
[`../architecture/video-transcript-acquisition-design.md`](../architecture/video-transcript-acquisition-design.md).
The targeted reduced report is
[`results/ockp-video-youtube-canonical-source-note.md`](results/ockp-video-youtube-canonical-source-note.md).

## Candidate Pipelines

| Pipeline | Reliability | Local-first behavior | Dependency and security implications | Transcript quality | Provenance and citation fit | Privacy posture |
| --- | --- | --- | --- | --- | --- | --- |
| Current document/retrieval primitives with supplied transcript | High when transcript text is supplied and path/title/body are explicit | Fully local through installed runner | No new dependencies | Depends on supplied transcript | Good after canonical markdown creation; citations come from indexed chunks | Strong; user supplies text already intended for OpenClerk |
| `yt-dlp` metadata/transcript extraction | Medium; platform changes can break extraction | Local command, but reaches platform URL | Adds fast-moving downloader and site-specific behavior | Good when captions exist; absent captions need another step | Needs explicit capture metadata, transcript hash, and URL-to-chunk mapping | URL and metadata are fetched from video platform |
| `ffmpeg` media extraction | High for media transforms after acquisition | Local command after media exists | Adds large binary surface and codec handling | No transcript by itself | Useful only as a supporting media step | Strong after local acquisition, but acquisition remains unresolved |
| Local Whisper or STT model | Medium-high for audio speech; varies by language and audio quality | Local if model and media are local | Adds model weights, CPU/GPU cost, and version drift | Often acceptable but imperfect; timestamps available | Needs model identity, language, confidence limits, and timestamp-to-chunk mapping | Strong if media stays local |
| Transcript APIs | Medium; depends on provider and URL support | Not local-first | Adds credentials, network egress, provider policy, and quota failure | Often good when provider supports source | Needs provider identity, capture timestamp, and external response provenance | Weak for private media or internal URLs unless explicitly approved |
| Gemini CLI or Gemini harness extraction | Medium; useful for exploratory extraction, not a stable runner contract | Usually not local-first unless configured to local assets only | Adds LLM dependency, prompt sensitivity, model drift, and possible data egress | Can summarize or transform, but transcript faithfulness is harder to prove | Weak unless raw transcript spans are preserved before generation | Weak by default for private media or proprietary transcript text |

## Ergonomics Scorecard

| Workflow | Candidate promoted surface | Tool or command count | Assistant calls | Wall time | Prompt specificity required | Failure classification | Authority/provenance/freshness risk |
| --- | --- | ---: | ---: | --- | --- | --- | --- |
| Current URL-only user intent | None; agent must reject unsupported native ingestion | 0 | 1 | Low | Natural user intent | `ergonomics_gap` | Safe rejection preserves invariants, but the expected source note is not created |
| Current supplied transcript control | Existing `create_document`, `search`, optional `list_documents` | 2-3 | 1 | Low-medium | Scripted control with explicit path, title, body, and provenance fields | `none` if transcript text is supplied | Low after canonical markdown creation; user/tool provenance must be written faithfully |
| Current stale synthesis inspection | Existing `search`, `list_documents`, `get_document`, `projection_states`, `provenance_events` | 5+ | 1 | Medium-high | Scenario-specific freshness choreography | `none` if agent inspects freshness | Low, but UX is too procedural for routine URL drops |
| `yt-dlp` plus runner create | `ingest_video_url` could wrap fetch, transcript capture, canonical note creation | 1 runner call if promoted; 3+ commands today | 1-2 | Medium | Moderate with dependency policy | `ergonomics_gap` if routine users must choreograph tools | Medium; platform extraction changes and transcript provenance must be explicit |
| `yt-dlp` plus `ffmpeg` plus local Whisper | `ingest_video_url` with local transcription policy | 1 runner call if promoted; 4+ commands today | 1-2 | High | High today; low-medium if promoted | `ergonomics_gap` for routine use | Medium; model/version/media hash/timestamps must be recorded |
| Transcript API plus runner create | `ingest_video_url` with explicit remote transcript policy | 1 runner call if promoted; 2+ commands/API calls today | 1-2 | Medium | High due privacy and credentials | `ergonomics_gap` only if user opts into remote policy | High unless egress, credentials, and provider provenance are visible |
| Gemini extraction plus runner create | Not recommended as authority surface | 2+ commands/calls | 1-2 | Medium-high | High and prompt-sensitive | `skill_guidance` or `eval_contract_violation` for routine ingestion | High; generated extraction can hide transcript span authority |

## Technical Expressibility

Current primitives can safely express a canonical video source note only after
transcript text and metadata are already available. The agent can create a
markdown source with fields such as:

```json
{"action":"create_document","document":{"path":"sources/video-youtube/demo.md","title":"Demo Video Transcript","body":"---\ntype: source\nstatus: active\nsource_type: video_transcript\nsource_url: https://youtube.example.test/watch?v=demo\ntranscript_origin: user_supplied\ncaptured_at: 2026-04-27T00:00:00Z\n---\n# Demo Video Transcript\n\n## Transcript\n..."}}
```

After creation, retrieval can cite indexed chunks with `doc_id`, `chunk_id`,
path, heading, and line ranges. Existing synthesis freshness behavior can also
show stale source-linked synthesis when a newer canonical transcript source
supersedes an older one.

Current primitives do not acquire the transcript from a video URL. That missing
acquisition step is not merely a JSON shape problem; it includes dependency
policy, privacy policy, transcript provenance, timestamp mapping, update
semantics, and failure classification.

## UX Acceptability

The current URL-only workflow is safe but not acceptable for the routine user
intent of "treat this YouTube URL like a source artifact." A correct agent must
reject native video ingestion rather than fetch media or call external tools.
That preserves AgentOps invariants, but it leaves the user to gather the
transcript, metadata, provenance, and citation mapping manually.

The supplied-transcript control remains important because it proves current
OpenClerk primitives can preserve authority once canonical text exists. It does
not solve the ergonomics gap for URL-only ingestion.

## POC Conclusion

Promote no production implementation from this POC alone. Use the targeted eval
lane to decide whether `ingest_video_url` should be promoted as a follow-up
surface.

If promoted later, the surface must be narrow and policy-explicit: no default
remote transcript API, no hidden LLM extraction, no asset-as-authority behavior,
and no weakening of citations, provenance, freshness, local-first operation, or
bypass rejection.
