# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `heterogeneous-artifact-ingestion-pressure`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `17.27`
- Harness elapsed seconds: `249.61`
- Effective parallel speedup: `0.83x`
- Parallel efficiency: `0.83`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 0/30 production scenarios passed; missing: create-note, search-synthesis, answer-filing, rag-retrieval-baseline, canonical-docs-navigation-baseline, graph-semantics-reference-poc, memory-router-reference-poc, configured-layout-explain, invalid-layout-visible, stale-synthesis-update, synthesis-freshness-repair, source-sensitive-audit-repair, source-sensitive-conflict-explain, synthesis-candidate-pressure, synthesis-source-set-pressure, append-replace, records-provenance, promoted-record-vs-docs, decision-record-vs-docs, decision-supersession-freshness, decision-real-adr-migration, missing-document-path-reject, negative-limit-reject, unsupported-lower-level-reject, unsupported-transport-reject, duplicate-path-reject, mixed-synthesis-records, mt-source-then-synthesis, mt-synthesis-drift-pressure, mt-incomplete-then-create |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect retired API files or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `pass` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_legacy_source_runner_usage` | `pass` | production must not invoke source-built or legacy runner paths instead of installed openclerk |
| `no_direct_sqlite_access` | `pass` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `fail` | not evaluated; final-answer-only validation scenarios were not selected in this partial run |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.15 |
| install_variant | 23.58 |
| warm_cache | 0.00 |
| seed_data | 0.09 |
| agent_run | 208.32 |
| parse_metrics | 0.00 |
| verify | 0.18 |
| total | 232.33 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `artifact-pdf-source-url-ingestion` | `failed` | 6 | 6 | 4 | 6626 | 23.59 | `<run-root>/production/artifact-pdf-source-url-ingestion/turn-1/events.jsonl` |
| `production` | `artifact-transcript-canonical-markdown` | `completed` | 6 | 6 | 4 | 25690 | 32.57 | `<run-root>/production/artifact-transcript-canonical-markdown/turn-1/events.jsonl` |
| `production` | `artifact-invoice-receipt-authority` | `completed` | 14 | 14 | 5 | 8670 | 37.35 | `<run-root>/production/artifact-invoice-receipt-authority/turn-1/events.jsonl` |
| `production` | `artifact-mixed-synthesis-freshness` | `completed` | 48 | 48 | 7 | 33246 | 93.79 | `<run-root>/production/artifact-mixed-synthesis-freshness/turn-1/events.jsonl` |
| `production` | `artifact-source-url-missing-hints` | `completed` | 0 | 0 | 1 | 2505 | 10.36 | `<run-root>/production/artifact-source-url-missing-hints/turn-1/events.jsonl` |
| `production` | `artifact-unsupported-native-video-ingest` | `completed` | 0 | 0 | 1 | 2545 | 6.21 | `<run-root>/production/artifact-unsupported-native-video-ingest/turn-1/events.jsonl` |
| `production` | `artifact-ingestion-bypass-reject` | `completed` | 0 | 0 | 1 | 2497 | 4.45 | `<run-root>/production/artifact-ingestion-bypass-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `defer_for_guidance_or_eval_repair`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted evidence only; no promoted runner action, parser, schema, storage migration, direct create behavior, or public API change.

| Variant | Scenario | Status | Failure classification | Evidence posture |
| --- | --- | --- | --- | --- |
| `production` | `artifact-pdf-source-url-ingestion` | `failed` | `data_hygiene` | fixture or durable artifact evidence did not satisfy heterogeneous artifact pressure |
| `production` | `artifact-transcript-canonical-markdown` | `completed` | `none` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-invoice-receipt-authority` | `completed` | `none` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-mixed-synthesis-freshness` | `completed` | `none` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-source-url-missing-hints` | `completed` | `none` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-unsupported-native-video-ingest` | `completed` | `none` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-ingestion-bypass-reject` | `completed` | `none` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
