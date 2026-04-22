# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `20.61`
- Harness elapsed seconds: `489.73`
- Effective parallel speedup: `0.90x`
- Parallel efficiency: `0.90`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 12/23 production scenarios passed; missing: create-note, answer-filing, rag-retrieval-baseline, canonical-docs-navigation-baseline, configured-layout-explain, invalid-layout-visible, append-replace, records-provenance, promoted-record-vs-docs, duplicate-path-reject, mt-incomplete-then-create |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect retired API files or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `pass` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_legacy_source_runner_usage` | `pass` | production must not invoke source-built or legacy runner paths instead of installed openclerk |
| `no_direct_sqlite_access` | `pass` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `pass` | rule-covered validation scenarios used no tools, no command executions, and at most one assistant answer |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.21 |
| install_variant | 27.15 |
| warm_cache | 0.00 |
| seed_data | 0.21 |
| agent_run | 441.07 |
| parse_metrics | 0.00 |
| verify | 0.50 |
| total | 469.12 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `search-synthesis` | `completed` | 16 | 16 | 6 | 6878 | 37.91 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 12 | 12 | 5 | 12525 | 42.33 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `synthesis-freshness-repair` | `completed` | 18 | 18 | 4 | 10674 | 41.90 | `<run-root>/production/synthesis-freshness-repair/turn-1/events.jsonl` |
| `production` | `synthesis-candidate-pressure` | `completed` | 30 | 30 | 10 | 38160 | 61.47 | `<run-root>/production/synthesis-candidate-pressure/turn-1/events.jsonl` |
| `production` | `synthesis-source-set-pressure` | `completed` | 16 | 16 | 6 | 8748 | 46.04 | `<run-root>/production/synthesis-source-set-pressure/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 6409 | 7.27 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 6397 | 7.47 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 6415 | 6.61 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2836 | 5.70 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 14 | 14 | 5 | 26116 | 39.13 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 10 | 10 | 6 | 13378 | 33.61 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-synthesis-drift-pressure` | `completed` | 40 | 40 | 11 | 48799 | 111.63 | `<run-root>/production/mt-synthesis-drift-pressure/turn-2/events.jsonl` |
