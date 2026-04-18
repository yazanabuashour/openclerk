# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `16.42`
- Harness elapsed seconds: `40.38`
- Effective parallel speedup: `1.59x`
- Parallel efficiency: `0.40`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Comparison

Candidate: `production`

Baseline: `sdk-baseline`

Beats baseline: `false`

Recommendation: `baseline_not_run_production_only_report`

| Criterion | Status | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | `pass` | 4/4 candidate scenarios passed |
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
| copy_repo | 0.16 |
| install_variant | 0.00 |
| warm_cache | 0.00 |
| seed_data | 0.07 |
| agent_run | 64.17 |
| parse_metrics | 0.00 |
| verify | 0.07 |
| total | 64.47 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `answer-filing` | `completed` | 4 | 4 | 3 | 4594 | 12.43 | `<run-root>/production/answer-filing/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 8 | 8 | 5 | 6254 | 23.85 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `promoted-record-vs-docs` | `completed` | 6 | 6 | 4 | 10211 | 22.35 | `<run-root>/production/promoted-record-vs-docs/turn-1/events.jsonl` |
| `production` | `unsupported-cli-mcp-reject` | `completed` | 0 | 0 | 1 | 3565 | 5.54 | `<run-root>/production/unsupported-cli-mcp-reject/turn-1/events.jsonl` |
