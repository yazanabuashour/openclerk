---
decision_id: decision-native-media-transcript-acquisition-candidate-comparison
decision_title: Native Media Transcript Acquisition Candidate Comparison
decision_status: accepted
decision_scope: artifact-native-media-transcript-acquisition
decision_owner: platform
source_refs: docs/evals/native-media-transcript-acquisition-candidate-comparison-poc.md, docs/evals/results/ockp-artifact-native-media-transcript-acquisition.md, docs/architecture/native-media-transcript-acquisition-promotion-decision.md, docs/architecture/video-transcript-acquisition-design.md, docs/architecture/video-youtube-ingestion-promotion-decision.md
---
# Decision: Native Media Transcript Acquisition Candidate Comparison

## Status

Accepted: select supplied-transcript hardening/current primitives as the best
current candidate for native media transcript acquisition.

This decision does not add a runner action, downloader dependency, caption
retrieval, local STT, transcript API, remote extraction, parser, schema,
migration, storage behavior, public API, public OpenClerk interface, product
behavior, shipped skill behavior, or implementation work. It does not
authorize an implementation bead.

Evidence:

- [`docs/evals/native-media-transcript-acquisition-candidate-comparison-poc.md`](../evals/native-media-transcript-acquisition-candidate-comparison-poc.md)
- [`docs/evals/results/ockp-artifact-native-media-transcript-acquisition.md`](../evals/results/ockp-artifact-native-media-transcript-acquisition.md)
- [`docs/architecture/native-media-transcript-acquisition-promotion-decision.md`](native-media-transcript-acquisition-promotion-decision.md)
- [`docs/architecture/video-transcript-acquisition-design.md`](video-transcript-acquisition-design.md)
- [`docs/architecture/video-youtube-ingestion-promotion-decision.md`](video-youtube-ingestion-promotion-decision.md)

## Decision

Select supplied-transcript hardening/current primitives:

- keep `openclerk document` `ingest_video_url` supplied-transcript-only
- keep canonical markdown source notes as authority
- preserve transcript provenance, source URL, transcript hash, capture time,
  language, citation-bearing retrieval, and source-linked freshness behavior
- improve guidance and eval ergonomics for supplied-transcript create and
  freshness workflows
- keep URL-only and local-file-only native acquisition unsupported
- defer local-only acquisition and remote-provider acquisition until later
  evidence names an exact safe surface

Outcome category: candidate selected for future guidance/eval hardening, not
implementation promotion.

## Rejected Alternatives

Do not select explicit local-only acquisition yet. A runner-owned caption or
local STT surface may eventually be viable, but it needs exact request/response
shape, dependency availability rules, model/runtime policy, resource limits,
confidence policy, timestamp mapping, raw media policy, and targeted promotion
evidence.

Do not select explicit remote-provider acquisition yet. Remote transcript APIs
or remote extraction have the highest privacy and trust risk. They require
explicit credentials, user approval, egress reporting, provider/model identity,
quota and provider-failure semantics, and proof that they never become hidden
fallbacks.

Do not kill the track. The underlying user need remains real: normal users may
expect OpenClerk to work from media URLs or recordings without separately
supplying transcript text. The current safe answer remains supplied transcript
text plus visible provenance, while native acquisition needs later surface
evidence.

## Safety, Capability, UX

Safety pass: pass. The `oc-69h3` eval preserved runner-only access, no direct
SQLite or vault inspection, no browser automation, no local file reads, no
manual media fetch, no downloader or STT dependency, no transcript API, no
remote extraction, no hidden third-party egress, no source-built runner usage,
no unsupported transports, transcript provenance, citation-bearing retrieval,
freshness visibility, and approval-before-write boundaries.

Capability pass: pass for current primitives. Existing `openclerk document`
and `openclerk retrieval` behavior can create and update supplied-transcript
canonical source notes, retrieve citation-bearing transcript evidence, expose
changed-transcript freshness/projection evidence, and reject unsupported native
acquisition or bypass requests without new runner behavior.

UX quality: hardening required, not implementation. Rejection and validation
controls completed cheaply. The supplied-transcript create control required 38
tools/commands, 6 assistant calls, and 51.55s; the freshness control required
26 tools/commands, 9 assistant calls, and 56.76s. That is taste debt in
guidance/eval ergonomics, but it does not justify native acquisition
implementation without stronger safety and surface evidence.

## Follow-Up

`oc-jyzp` tracks guidance/eval hardening for supplied-transcript create and
freshness ceremony. The follow-up must reduce prompt choreography and step
count while preserving:

- supplied-transcript-only `ingest_video_url`
- runner-only `openclerk document` and `openclerk retrieval`
- transcript provenance and citation-bearing retrieval
- update/no-op and stale projection visibility
- no native media fetch, downloader, STT, transcript API, remote extraction,
  browser, direct vault/SQLite, HTTP/MCP, source-built runner, or unsupported
  transport bypasses

No implementation bead should be filed from `oc-e40d`.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public OpenClerk
  surfaces for this track.
- Supplied transcript text remains supported through `ingest_video_url`.
- Public media URLs and local media paths without transcript text remain
  unsupported for routine durable intake.
- Local-only and remote-provider acquisition remain deferred until a later
  accepted promotion decision names an exact safe surface.
- Committed docs, reports, and examples must continue to use repo-relative
  paths or neutral placeholders such as `<run-root>`.
