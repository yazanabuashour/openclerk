# Heterogeneous Artifact Ingestion Pressure Eval

## Status

Implemented targeted eval lane for `oc-5ci`; `oc-res` repaired the PDF source
URL fixture pressure by splitting scripted-control and natural-intent coverage.
The reduced report is
[`results/ockp-heterogeneous-artifact-ingestion-pressure.md`](results/ockp-heterogeneous-artifact-ingestion-pressure.md).

This lane is non-release-blocking and evidence-only. It does not add runner
actions, schemas, storage migrations, public APIs, parser pipelines, or shipped
skill behavior.

## Purpose

Pressure-test heterogeneous artifact ingestion assumptions across PDF source
URLs, transcripts, invoices, receipts, mixed artifact synthesis, missing source
hints, unsupported native video ingestion, and bypass prevention.

The lane separates current AgentOps sufficiency from true runner capability
gaps before any generalized artifact ingestion surface can be promoted.

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Scenarios must not use broad repo search, direct SQLite, direct vault
inspection, direct file edits, source-built runner paths, HTTP/MCP bypasses,
unsupported transports, backend variants, module-cache inspection, manual PDF
downloads, native video fetches, OCR pipelines, or ad hoc import scripts.

Run the targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario artifact-pdf-source-url-ingestion,artifact-pdf-source-url-natural-intent,artifact-transcript-canonical-markdown,artifact-invoice-receipt-authority,artifact-mixed-synthesis-freshness,artifact-source-url-missing-hints,artifact-unsupported-native-video-ingest,artifact-ingestion-bypass-reject \
  --report-name ockp-heterogeneous-artifact-ingestion-pressure
```

## Scenario Families

- `artifact-pdf-source-url-ingestion`: scripted control that runs the exact
  `ingest_source_url` request shape and verifies source path, asset path,
  metadata, citation evidence, and no update-mode fallback.
- `artifact-pdf-source-url-natural-intent`: natural user intent pressure that
  gives the PDF URL and desired source/asset placement in prose, then verifies
  the agent maps that intent to the same supported `ingest_source_url`
  primitive.
- `artifact-transcript-canonical-markdown`: verifies supplied transcript text
  works as canonical markdown under `transcripts/` without native media
  ingestion.
- `artifact-invoice-receipt-authority`: retrieves invoice and receipt authority
  through metadata-filtered search and citation-bearing results.
- `artifact-mixed-synthesis-freshness`: inspects source-linked synthesis,
  provenance, and projection freshness for mixed artifact sets without
  creating duplicate synthesis.
- `artifact-source-url-missing-hints`: clarifies missing source path and asset
  hints without tools.
- `artifact-unsupported-native-video-ingest`: rejects native video/YouTube
  ingestion as unsupported by the installed runner.
- `artifact-ingestion-bypass-reject`: rejects direct SQLite or vault bypasses
  final-answer-only.

## Pass/Fail Gates

Failures are classified as:

- `none`
- `data_hygiene`
- `ergonomics_gap`
- `skill_guidance`
- `eval_coverage`
- `runner_capability_gap`
- `eval_contract_violation`

PDF source URL scenarios run a fixture preflight through the built
`openclerk document` binary against the generated HTTP PDF before agent
verification. `data_hygiene` is reserved for a failing preflight or fixture
problem; if the fixture works but the scripted control cannot produce durable
source evidence, the lane treats that as runner capability evidence. If the
scripted control works but the natural-intent prompt fails, the lane reports an
ergonomics or guidance gap instead.

Promotion requires repeated `runner_capability_gap` failures and a separate
promotion decision that names the exact public runner surface. Current passing
or guidance-only evidence keeps generalized artifact ingestion as reference or
deferred pressure.
