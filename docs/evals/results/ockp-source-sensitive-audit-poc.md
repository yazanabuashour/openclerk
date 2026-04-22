# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `18.33`
- Harness elapsed seconds: `216.67`
- Effective parallel speedup: `0.83x`
- Parallel efficiency: `0.83`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 8/27 production scenarios passed; missing: create-note, search-synthesis, answer-filing, rag-retrieval-baseline, canonical-docs-navigation-baseline, configured-layout-explain, invalid-layout-visible, synthesis-candidate-pressure, synthesis-source-set-pressure, append-replace, records-provenance, promoted-record-vs-docs, decision-record-vs-docs, decision-supersession-freshness, duplicate-path-reject, mixed-synthesis-records, mt-source-then-synthesis, mt-synthesis-drift-pressure, mt-incomplete-then-create |
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
| copy_repo | 0.15 |
| install_variant | 16.82 |
| warm_cache | 0.00 |
| seed_data | 0.18 |
| agent_run | 180.91 |
| parse_metrics | 0.00 |
| verify | 0.29 |
| total | 198.33 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `stale-synthesis-update` | `completed` | 12 | 12 | 4 | 13482 | 28.23 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `synthesis-freshness-repair` | `completed` | 18 | 18 | 6 | 10970 | 40.68 | `<run-root>/production/synthesis-freshness-repair/turn-1/events.jsonl` |
| `production` | `source-sensitive-audit-repair` | `completed` | 38 | 38 | 10 | 14436 | 57.64 | `<run-root>/production/source-sensitive-audit-repair/turn-1/events.jsonl` |
| `production` | `source-sensitive-conflict-explain` | `completed` | 12 | 12 | 5 | 7502 | 31.99 | `<run-root>/production/source-sensitive-conflict-explain/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2825 | 5.98 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2813 | 7.05 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2831 | 4.95 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2836 | 4.39 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |
