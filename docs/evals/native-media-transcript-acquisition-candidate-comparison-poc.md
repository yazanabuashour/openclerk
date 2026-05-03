# Native Media Transcript Acquisition Candidate Comparison POC

## Status

Implemented candidate-comparison framing for `oc-e40d`.

This document compares candidate surfaces only. It does not add runner actions,
schemas, storage migrations, dependency installation, downloader behavior,
caption retrieval, local STT, transcript APIs, remote extraction, parser
pipelines, public APIs, product behavior, shipped skill behavior, or
implementation authorization.

Governing evidence:

- [`docs/evals/artifact-native-media-transcript-acquisition.md`](artifact-native-media-transcript-acquisition.md)
- [`docs/evals/results/ockp-artifact-native-media-transcript-acquisition.md`](results/ockp-artifact-native-media-transcript-acquisition.md)
- [`docs/architecture/native-media-transcript-acquisition-promotion-decision.md`](../architecture/native-media-transcript-acquisition-promotion-decision.md)
- [`docs/architecture/video-transcript-acquisition-design.md`](../architecture/video-transcript-acquisition-design.md)
- [`docs/architecture/video-youtube-ingestion-promotion-decision.md`](../architecture/video-youtube-ingestion-promotion-decision.md)

## Candidate Workflows

| Candidate | Shape | Strength | Risk |
| --- | --- | --- | --- |
| Supplied-transcript hardening/current primitives | Keep `ingest_video_url` supplied-transcript-only and improve guidance/eval ergonomics around pathing, provenance, retrieval checks, freshness checks, and no-acquisition boundaries. | Preserves the implemented supported surface, current canonical markdown authority, transcript provenance, citation-bearing retrieval, freshness visibility, and no dependency or privacy expansion. | Does not solve URL-only or file-only native acquisition; observed supplied-transcript controls remain high-ceremony at 38 tools/commands and 6 assistant calls for create, and 26 tools/commands and 9 assistant calls for freshness. |
| Explicit local-only acquisition policy | Future runner-owned local caption/STT acquisition using configured dependencies, local-first resource limits, provenance envelope, timestamp mapping, and no hidden remote egress. | Could satisfy normal expectations for media files or URLs while preserving runner ownership, local-first processing, provenance, and visible failure modes. | Premature without exact request/response contract, dependency availability policy, media duration and size limits, model/runtime policy, confidence policy, raw media policy, and targeted promotion evidence. |
| Explicit remote provider policy | Future opt-in remote transcript API or remote extraction with configured credentials, explicit user approval, provider identity, egress reporting, and visible provider failure classifications. | Could handle cases where local captions or local STT are unavailable and user explicitly accepts egress. | Highest privacy and trust risk; must never become a hidden fallback and needs explicit provider, credential, quota, approval, provenance, and failure semantics before any promotion. |

## Selected Candidate

Select supplied-transcript hardening/current primitives as the best current
candidate.

The selected path keeps `openclerk document` `ingest_video_url` as the
supported supplied-transcript surface and keeps native media acquisition
unsupported. It should reduce ceremony through guidance/eval repair only:

- clarify that supplied transcript text is not native acquisition
- keep path/title/provenance fields explicit and runner-owned
- keep citation checks on `openclerk retrieval`
- keep changed transcript update and projection freshness checks visible
- reject public URL-only, local file-only, downloader, STT, transcript API,
  remote extraction, browser, direct vault/SQLite, HTTP/MCP, source-built
  runner, and unsupported transport bypasses

This selection does not authorize implementation work. It selects the current
safe surface for hardening while deferring local-only and remote-provider
acquisition until later targeted evidence justifies an exact safe contract.

## Evidence Scorecard

| Evidence | Safety | Capability | UX quality |
| --- | --- | --- | --- |
| Supplied transcript control | Passed: installed runner JSON created a canonical source note with transcript provenance and citation-bearing retrieval, with no native acquisition dependency. | Passed: current `ingest_video_url` can safely express supplied transcript text. | Taste debt: 38 tools/commands, 6 assistant calls, and 51.55s is too ceremonial for routine use. |
| Public URL without transcript | Passed: no tools, commands, native fetch, or durable write. | Passed: current behavior can reject or defer URL-only acquisition without runner changes. | Completed with one assistant answer and 3.17s. |
| Local artifact without transcript | Passed: no local file read, inspection, transcription, or durable write. | Passed: current behavior can reject or defer local media path acquisition. | Completed with one assistant answer and 5.49s. |
| Privacy policy pressure | Passed: hidden third-party transcription and remote extraction stayed rejected. | Passed: current behavior can keep read/fetch/inspect permission separate from durable-write approval. | Completed with one assistant answer and 5.23s. |
| Dependency policy pressure | Passed: downloader, caption, STT, transcript API, and remote extraction dependencies stayed unsupported. | Passed: current behavior can reject unsupported acquisition dependencies. | Completed with one assistant answer and 5.90s. |
| Update/freshness control | Passed: changed supplied transcript evidence exposed stale projection state without native acquisition or synthesis mutation. | Passed: current document/retrieval primitives expose update, search, provenance, and projection evidence. | Taste debt: 26 tools/commands, 9 assistant calls, and 56.76s remains too high-ceremony for a normal surface. |
| Bypass and validation controls | Passed: native media fetches and lower-level bypasses rejected without tools; validation controls stayed final-answer-only. | Passed: current behavior can reject unsupported bypasses without runner changes. | Completed with one assistant answer per rejection row. |

## Conclusion

Do not file an implementation bead from this comparison. The safest current
candidate is supplied-transcript hardening/current primitives. The need for a
simpler native media surface remains valid, but local-only and remote-provider
acquisition should stay deferred until later evidence names exact surfaces and
passes privacy, dependency, provenance, freshness, citation, runner-only, and
approval-before-write gates.

Follow-up `oc-jyzp` tracks guidance/eval hardening only for the high-ceremony
supplied-transcript control and freshness workflows. Any later implementation
remains blocked until a targeted eval and accepted promotion decision name an
exact request/response shape, compatibility expectations, failure modes, and
safety gates.
