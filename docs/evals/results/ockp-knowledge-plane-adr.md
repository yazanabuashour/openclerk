# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `15.34`
- Harness elapsed seconds: `265.18`
- Effective parallel speedup: `3.53x`
- Parallel efficiency: `0.88`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Comparison

Candidate: `production`

Baseline: `sdk-baseline`

Beats baseline: `true`

Recommendation: `use_production_agentops_for_routine_openclerk_operations`

| Criterion | Status | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | `pass` | 11/11 candidate scenarios passed |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect generated clients or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `pass` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_openclerk_cli_usage` | `pass` | production must not use the human OpenClerk CLI for AgentOps tasks |
| `no_direct_sqlite_access` | `pass` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `pass` | rule-covered validation scenarios used no tools, no command executions, and at most one assistant answer |
| `total_tools_less_than_or_equal_baseline` | `pass` | production tools 96 vs baseline tools 347; missing baseline: ok |
| `minimum_scenarios_at_or_below_baseline` | `pass` | 11 scenarios at or below baseline tools; required 9 of 11 |
| `non_cached_token_majority` | `pass` | 11 scenarios with lower non-cached input tokens; required 6 of 11; missing usage: ok |
| `non_cached_token_total_less_than_or_equal_baseline` | `pass` | production non-cached input tokens 79767 vs baseline 306322; missing usage: ok |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.53 |
| install_variant | 0.00 |
| warm_cache | 0.00 |
| seed_data | 0.16 |
| agent_run | 937.38 |
| parse_metrics | 0.00 |
| verify | 0.20 |
| total | 938.35 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 8 | 8 | 5 | 4790 | 18.90 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `search-synthesis` | `completed` | 10 | 10 | 5 | 10545 | 20.16 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `append-replace` | `completed` | 6 | 6 | 3 | 4845 | 13.69 | `<run-root>/production/append-replace/turn-1/events.jsonl` |
| `production` | `records-provenance` | `completed` | 30 | 30 | 7 | 19599 | 65.86 | `<run-root>/production/records-provenance/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 3540 | 4.94 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 3529 | 4.83 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 3544 | 5.35 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `duplicate-path-reject` | `completed` | 2 | 2 | 2 | 3854 | 9.08 | `<run-root>/production/duplicate-path-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 32 | 32 | 7 | 8145 | 54.67 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 6 | 6 | 5 | 9546 | 21.67 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-incomplete-then-create` | `completed` | 2 | 2 | 3 | 7830 | 14.11 | `<run-root>/production/mt-incomplete-then-create/turn-2/events.jsonl` |
| `sdk-baseline` | `create-note` | `completed` | 32 | 32 | 9 | 23492 | 43.24 | `<run-root>/sdk-baseline/create-note/turn-1/events.jsonl` |
| `sdk-baseline` | `search-synthesis` | `completed` | 26 | 26 | 10 | 33372 | 81.69 | `<run-root>/sdk-baseline/search-synthesis/turn-1/events.jsonl` |
| `sdk-baseline` | `append-replace` | `completed` | 20 | 20 | 7 | 7775 | 21.29 | `<run-root>/sdk-baseline/append-replace/turn-1/events.jsonl` |
| `sdk-baseline` | `records-provenance` | `completed` | 54 | 54 | 10 | 50238 | 140.88 | `<run-root>/sdk-baseline/records-provenance/turn-1/events.jsonl` |
| `sdk-baseline` | `missing-document-path-reject` | `failed` | 22 | 22 | 5 | 24155 | 21.37 | `<run-root>/sdk-baseline/missing-document-path-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `negative-limit-reject` | `failed` | 30 | 30 | 7 | 25927 | 32.62 | `<run-root>/sdk-baseline/negative-limit-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `unsupported-lower-level-reject` | `failed` | 39 | 39 | 11 | 27998 | 93.19 | `<run-root>/sdk-baseline/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `duplicate-path-reject` | `completed` | 28 | 28 | 7 | 19780 | 37.78 | `<run-root>/sdk-baseline/duplicate-path-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `mixed-synthesis-records` | `completed` | 34 | 34 | 10 | 25412 | 96.87 | `<run-root>/sdk-baseline/mixed-synthesis-records/turn-1/events.jsonl` |
| `sdk-baseline` | `mt-source-then-synthesis` | `completed` | 40 | 40 | 16 | 48641 | 86.79 | `<run-root>/sdk-baseline/mt-source-then-synthesis/turn-2/events.jsonl` |
| `sdk-baseline` | `mt-incomplete-then-create` | `failed` | 22 | 22 | 7 | 19532 | 48.40 | `<run-root>/sdk-baseline/mt-incomplete-then-create/turn-2/events.jsonl` |
