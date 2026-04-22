# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `18.84`
- Harness elapsed seconds: `52.54`
- Effective parallel speedup: `0.60x`
- Parallel efficiency: `0.60`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 1/18 production scenarios passed; missing: create-note, search-synthesis, answer-filing, rag-retrieval-baseline, stale-synthesis-update, synthesis-freshness-repair, append-replace, records-provenance, promoted-record-vs-docs, missing-document-path-reject, negative-limit-reject, unsupported-lower-level-reject, unsupported-transport-reject, duplicate-path-reject, mixed-synthesis-records, mt-source-then-synthesis, mt-incomplete-then-create |
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
| copy_repo | 0.01 |
| install_variant | 2.16 |
| warm_cache | 0.00 |
| seed_data | 0.04 |
| agent_run | 31.41 |
| parse_metrics | 0.01 |
| verify | 0.07 |
| total | 33.70 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `canonical-docs-navigation-baseline` | `completed` | 12 | 12 | 3 | 13794 | 31.41 | `<run-root>/production/canonical-docs-navigation-baseline/turn-1/events.jsonl` |
