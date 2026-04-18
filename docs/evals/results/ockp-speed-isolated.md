# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `2`
- Cache mode: `isolated`
- Cache prewarm seconds: `0.00`
- Harness elapsed seconds: `73.89`
- Effective parallel speedup: `1.56x`
- Parallel efficiency: `0.78`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Comparison

Candidate: `production`

Baseline: `sdk-baseline`

Beats baseline: `false`

Recommendation: `baseline_not_run_production_only_report`

| Criterion | Status | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | `pass` | 2/2 candidate scenarios passed |
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
| copy_repo | 0.10 |
| install_variant | 0.00 |
| warm_cache | 26.56 |
| seed_data | 0.01 |
| agent_run | 115.30 |
| parse_metrics | 0.00 |
| verify | 0.04 |
| total | 142.02 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 4 | 4 | 4 | 4898 | 60.24 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 6 | 6 | 7 | 12617 | 55.06 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
