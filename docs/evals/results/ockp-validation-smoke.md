# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `3`
- Cache mode: `shared`
- Cache prewarm seconds: `21.14`
- Harness elapsed seconds: `26.99`
- Effective parallel speedup: `0.58x`
- Parallel efficiency: `0.19`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Comparison

Candidate: `production`

Baseline: `sdk-baseline`

Beats baseline: `false`

Recommendation: `baseline_not_run_production_only_report`

| Criterion | Status | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | `pass` | 3/3 candidate scenarios passed |
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
| copy_repo | 0.15 |
| install_variant | 0.00 |
| warm_cache | 0.00 |
| seed_data | 0.00 |
| agent_run | 15.72 |
| parse_metrics | 0.00 |
| verify | 0.00 |
| total | 15.88 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 3540 | 5.01 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 3529 | 4.92 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 3544 | 5.79 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
