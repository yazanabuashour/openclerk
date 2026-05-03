---
decision_id: decision-native-media-transcript-acquisition
decision_title: Native Media Transcript Acquisition
decision_status: accepted
decision_scope: artifact-native-media-transcript-acquisition
decision_owner: platform
---
# Decision: Native Media Transcript Acquisition

## Status

Accepted as reference evidence for `oc-69h3`: keep native media transcript
acquisition unpromoted. The current supported surface remains supplied
transcript text through `openclerk document` `ingest_video_url`; native media
download, caption retrieval, local STT, transcript APIs, Gemini or remote
extraction, parser pipelines, dependency installation, schema changes, storage
changes, public API changes, and skill behavior changes remain unsupported.

Evidence:

- [`../evals/artifact-native-media-transcript-acquisition.md`](../evals/artifact-native-media-transcript-acquisition.md)
- [`../evals/results/ockp-artifact-native-media-transcript-acquisition.md`](../evals/results/ockp-artifact-native-media-transcript-acquisition.md)
- [`video-transcript-acquisition-design.md`](video-transcript-acquisition-design.md)
- [`video-youtube-ingestion-promotion-decision.md`](video-youtube-ingestion-promotion-decision.md)

## Decision

Keep this lane as reference evidence. Do not file an implementation bead from
`oc-69h3.4`.

Safety pass: the targeted run preserved runner-only access, no direct SQLite or
vault inspection, no browser automation, no native media fetch, no downloader
or STT dependency, no hidden remote extraction, transcript provenance,
citation-bearing supplied transcript evidence, update/freshness visibility, and
approval-before-write boundaries.

Capability pass: current `openclerk document` and `openclerk retrieval`
primitives can express the supported supplied-transcript control. They do not
provide native media transcript acquisition, and the eval does not authorize a
new runner action or dependency policy.

UX quality: the lane completed, but the supplied-transcript and freshness
controls were still highly ceremonial in practice. A normal user may reasonably
expect a simpler surface for media URLs or recordings than "bring transcript
text, pathing, provenance, update mode, retrieval checks, projection checks, and
provenance checks yourself."

Outcome category: need exists, candidate comparison required. The evaluated
shape passed as reference evidence, but native acquisition remains a real UX
need with unresolved privacy, dependency, provenance, citation mapping,
freshness, and local-first tradeoffs.

## Follow-Up

`bd search` found no existing candidate-comparison follow-up for native media
transcript acquisition beyond the `oc-69h3` epic and children. Created
`oc-e40d`, "Compare native media transcript acquisition candidate surfaces", to
compare 2-3 candidate surfaces and choose, combine, defer, kill, or record
`none viable yet`.

## Non-Authorization

This decision does not authorize implementation work. Any future promotion must
name the exact public surface, request and response shape, compatibility
expectations, failure modes, dependency and privacy gates, provenance and
citation behavior, freshness semantics, and approval-before-write boundary.
