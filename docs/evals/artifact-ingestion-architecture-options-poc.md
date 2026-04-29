# Artifact Ingestion Architecture Options POC

## Status

Implemented POC framing for `oc-ihz`. This document records architecture
options and eval obligations only. It does not add runner actions, schemas,
storage migrations, public API behavior, parser pipelines, or shipped skill
behavior.

The governing ADR is
[`../architecture/generalized-artifact-ingestion-adr.md`](../architecture/generalized-artifact-ingestion-adr.md).
The targeted reduced report is
[`results/ockp-heterogeneous-artifact-ingestion-pressure.md`](results/ockp-heterogeneous-artifact-ingestion-pressure.md).

## Candidate Surfaces

Current PDF URL ingestion remains the compatibility baseline:

```json
{"action":"ingest_source_url","source":{"url":"https://example.test/source.pdf","path_hint":"sources/example.md","asset_path_hint":"assets/sources/example.pdf","title":"Example Source"}}
```

The existing action returns source path, asset path, hash, MIME type, page
count, capture timestamp, PDF metadata, and citation-bearing source evidence.
Missing `source.mode` means `create`; duplicate creates reject; explicit
`source.mode: "update"` targets an existing normalized `source.url`; mismatched
path or asset hints conflict without writing.

Artifact-specific actions could look like:

```json
{"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=demo","path_hint":"transcripts/demo.md","asset_path_hint":"assets/videos/demo.json","title":"Demo Video"}}
{"action":"ingest_receipt","receipt":{"uri":"<receipt-artifact-uri>","path_hint":"receipts/vendor-2026-04.md","asset_path_hint":"assets/receipts/vendor-2026-04.pdf","metadata":{"vendor":"Vendor","total_usd":"86.40"}}}
```

This option may keep validation and failure modes clear, but it risks a growing
set of narrow actions and duplicated provenance behavior.

A generalized action could look like:

```json
{"action":"ingest_artifact","artifact":{"kind":"video","uri":"https://youtube.example.test/watch?v=demo","path_hint":"transcripts/demo.md","asset_path_hint":"assets/videos/demo.json","mode":"create","metadata":{"source_platform":"youtube"}}}
```

This option could centralize duplicate handling, asset references, provenance,
and update semantics across artifact kinds. It also risks becoming too broad:
`kind`, URI semantics, parser availability, metadata validation, partial
success behavior, and citation mapping would all need exact contracts before
promotion.

## Mapping Expectations

- PDFs map to canonical `sources/*.md` notes plus `assets/**/*.pdf` assets
  through the existing `ingest_source_url` action.
- Pasted or preexisting transcript text maps to canonical markdown under
  `transcripts/` and is searchable with citations through existing retrieval.
- Invoices and receipts map to canonical markdown under `invoices/` and
  `receipts/`; typed extraction is not promoted by this POC.
- Mixed artifact sets map to source-linked synthesis under `synthesis/`, with
  `source_refs`, `## Sources`, `## Freshness`, provenance events, and
  projection-state inspection.
- Videos and YouTube links remain unsupported as native ingestion unless a
  later decision promotes an exact media/transcript surface.

## Compatibility And Failure Modes

Any promoted future surface must preserve these compatibility rules:

- `ingest_source_url` request and response behavior remains unchanged.
- Missing mode defaults to create; update semantics and duplicate/conflict
  behavior stay compatible with existing PDF source URL ingestion.
- Existing document/retrieval actions remain sufficient for canonical markdown
  transcripts, invoice notes, receipt notes, and synthesis maintenance.
- New native artifact ingestion must not require routine direct SQLite, direct
  vault inspection, broad repo search, source-built runner paths, HTTP/MCP
  bypasses, unsupported transports, backend variants, module-cache inspection,
  or ad hoc import scripts.

Failure modes to classify in targeted evals:

- missing `source.path_hint`, `source.asset_path_hint`, path, title, body, or
  artifact kind
- unsupported artifact kind or unsupported URI transport
- parser, OCR, media, transcript, or metadata extraction failure
- duplicate source URL, duplicate asset, duplicate transcript, or duplicate
  synthesis candidate
- partial mixed-artifact success where some sources ingest and others fail
- stale source-linked synthesis after artifact refresh
- conflicting current sources with no runner-visible supersession authority
- provenance, citation, or freshness gaps

## POC Conclusion

Existing document and retrieval actions remain structurally sufficient for the
passing markdown-transcribed artifact cases: transcripts, invoices, receipts,
and explicit unsupported-workflow rejection can be represented or handled
without a new runner surface.

The targeted lane did not justify promotion. It classified the remaining
failure as data hygiene rather than repeated `runner_capability_gap` evidence.
Native artifact handling may still need a future runner surface for artifact
bytes that require OpenClerk-managed acquisition or parsing, such as
video/YouTube transcription, OCR-heavy receipts, local file import, or richer
media metadata. That promotion remains deferred until targeted eval evidence
shows repeated runner capability gaps and the exact surface is named in a
promotion decision.
