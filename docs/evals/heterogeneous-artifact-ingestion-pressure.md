# Heterogeneous Artifact Ingestion Pressure Eval

## Status

Implemented targeted eval lane for `oc-5ci`. The reduced report is
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
  --scenario artifact-pdf-source-url-ingestion,artifact-transcript-canonical-markdown,artifact-invoice-receipt-authority,artifact-mixed-synthesis-freshness,artifact-source-url-missing-hints,artifact-unsupported-native-video-ingest,artifact-ingestion-bypass-reject \
  --report-name ockp-heterogeneous-artifact-ingestion-pressure
```

## Scenario Families

- `artifact-pdf-source-url-ingestion`: ingests a PDF source URL through the
  existing `ingest_source_url` action and verifies source path, asset path, and
  citation evidence.
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
- `skill_guidance`
- `eval_coverage`
- `runner_capability_gap`
- `eval_contract_violation`

Promotion requires repeated `runner_capability_gap` failures and a separate
promotion decision that names the exact public runner surface. Current passing
or guidance-only evidence keeps generalized artifact ingestion as reference or
deferred pressure.
