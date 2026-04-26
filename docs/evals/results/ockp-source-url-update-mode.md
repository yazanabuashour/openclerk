# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `source-url-update`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `16.05`
- Harness elapsed seconds: `371.77`
- Effective parallel speedup: `0.92x`
- Parallel efficiency: `0.92`
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
| copy_repo | 0.05 |
| install_variant | 12.89 |
| warm_cache | 0.00 |
| seed_data | 0.10 |
| agent_run | 342.36 |
| parse_metrics | 0.01 |
| verify | 0.26 |
| total | 355.72 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `source-url-update-duplicate-create` | `completed` | 6 | 6 | 3 | 4380 | 18.16 | `<run-root>/production/source-url-update-duplicate-create/turn-1/events.jsonl` |
| `production` | `source-url-update-same-sha-noop` | `completed` | 26 | 26 | 8 | 10172 | 52.30 | `<run-root>/production/source-url-update-same-sha-noop/turn-1/events.jsonl` |
| `production` | `source-url-update-changed-pdf-stale` | `completed` | 58 | 58 | 12 | 34430 | 121.23 | `<run-root>/production/source-url-update-changed-pdf-stale/turn-1/events.jsonl` |
| `production` | `source-url-update-path-hint-conflict` | `completed` | 64 | 64 | 8 | 25689 | 150.67 | `<run-root>/production/source-url-update-path-hint-conflict/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `keep_existing_update_mode`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted AgentOps evidence for existing ingest_source_url source.mode update behavior; no new runner action, schema, storage API, or transport.

| Variant | Scenario | Status | Failure classification | Evidence posture |
| --- | --- | --- | --- | --- |
| `production` | `source-url-update-duplicate-create` | `completed` | `none` | installed document/retrieval runner evidence covered source URL update mode |
| `production` | `source-url-update-same-sha-noop` | `completed` | `none` | installed document/retrieval runner evidence covered source URL update mode |
| `production` | `source-url-update-changed-pdf-stale` | `completed` | `none` | installed document/retrieval runner evidence covered source URL update mode |
| `production` | `source-url-update-path-hint-conflict` | `completed` | `none` | installed document/retrieval runner evidence covered source URL update mode |
