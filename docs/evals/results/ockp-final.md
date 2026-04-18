# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `21.20`
- Harness elapsed seconds: `79.80`
- Effective parallel speedup: `2.35x`
- Parallel efficiency: `0.59`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Comparison

Candidate: `production`

Baseline: `sdk-baseline`

Beats baseline: `false`

Recommendation: `baseline_not_run_production_only_report`

| Criterion | Status | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | `pass` | 11/11 candidate scenarios passed |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect generated clients or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `pass` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_openclerk_cli_usage` | `pass` | production must not use the human OpenClerk CLI for AgentOps tasks |
| `no_direct_sqlite_access` | `pass` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `pass` | rule-covered validation scenarios used no tools, no command executions, and at most one assistant answer |
| `baseline_not_run` | `fail` | baseline comparison criteria skipped because this report selected only production |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.48 |
| install_variant | 0.00 |
| warm_cache | 0.00 |
| seed_data | 0.12 |
| agent_run | 187.79 |
| parse_metrics | 0.00 |
| verify | 0.09 |
| total | 188.50 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 12 | 12 | 5 | 9686 | 23.74 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `search-synthesis` | `completed` | 12 | 12 | 7 | 6045 | 25.62 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `append-replace` | `completed` | 6 | 6 | 5 | 5602 | 16.51 | `<run-root>/production/append-replace/turn-1/events.jsonl` |
| `production` | `records-provenance` | `completed` | 16 | 16 | 5 | 10651 | 34.62 | `<run-root>/production/records-provenance/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 3540 | 4.73 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 3529 | 4.81 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 3544 | 6.31 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `duplicate-path-reject` | `completed` | 2 | 2 | 2 | 3947 | 9.62 | `<run-root>/production/duplicate-path-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 8 | 8 | 5 | 5884 | 20.97 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 8 | 8 | 6 | 10106 | 28.40 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-incomplete-then-create` | `completed` | 2 | 2 | 3 | 7843 | 12.46 | `<run-root>/production/mt-incomplete-then-create/turn-2/events.jsonl` |
