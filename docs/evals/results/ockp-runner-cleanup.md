# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `15.54`
- Harness elapsed seconds: `375.71`
- Effective parallel speedup: `3.54x`
- Parallel efficiency: `0.89`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Comparison

Candidate: `production`

Baseline: `sdk-baseline`

Beats baseline: `true`

Recommendation: `use_production_runner_for_routine_openclerk_operations`

| Criterion | Status | Details |
| --- | --- | --- |
| `candidate_passes_all_scenarios` | `pass` | 15/15 candidate scenarios passed |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect retired API files or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `pass` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_legacy_source_runner_usage` | `pass` | production must not invoke source-built or legacy runner paths instead of installed openclerk |
| `no_direct_sqlite_access` | `pass` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `pass` | rule-covered validation scenarios used no tools, no command executions, and at most one assistant answer |
| `total_tools_less_than_or_equal_baseline` | `pass` | production tools 114 vs baseline tools 550; missing baseline: ok |
| `minimum_scenarios_at_or_below_baseline` | `pass` | 15 scenarios at or below baseline tools; required 12 of 15 |
| `non_cached_token_majority` | `pass` | 15 scenarios with lower non-cached input tokens; required 8 of 15; missing usage: ok |
| `non_cached_token_total_less_than_or_equal_baseline` | `pass` | production non-cached input tokens 94401 vs baseline 454369; missing usage: ok |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.34 |
| install_variant | 17.31 |
| warm_cache | 0.00 |
| seed_data | 0.23 |
| agent_run | 1329.54 |
| parse_metrics | 0.02 |
| verify | 0.33 |
| total | 1347.84 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 6 | 6 | 4 | 4704 | 18.02 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `search-synthesis` | `completed` | 8 | 8 | 4 | 5409 | 23.23 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `answer-filing` | `completed` | 14 | 14 | 5 | 11938 | 31.86 | `<run-root>/production/answer-filing/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 20 | 20 | 6 | 8989 | 58.92 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `append-replace` | `completed` | 6 | 6 | 3 | 4808 | 16.37 | `<run-root>/production/append-replace/turn-1/events.jsonl` |
| `production` | `records-provenance` | `completed` | 8 | 8 | 3 | 5336 | 28.48 | `<run-root>/production/records-provenance/turn-1/events.jsonl` |
| `production` | `promoted-record-vs-docs` | `completed` | 12 | 12 | 4 | 6752 | 28.64 | `<run-root>/production/promoted-record-vs-docs/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 3575 | 4.00 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 3563 | 5.67 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 3581 | 5.86 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-cli-mcp-reject` | `completed` | 0 | 0 | 1 | 3593 | 15.91 | `<run-root>/production/unsupported-cli-mcp-reject/turn-1/events.jsonl` |
| `production` | `duplicate-path-reject` | `completed` | 2 | 2 | 2 | 3956 | 10.04 | `<run-root>/production/duplicate-path-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 22 | 22 | 6 | 8664 | 41.84 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 14 | 14 | 7 | 11526 | 39.79 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-incomplete-then-create` | `completed` | 2 | 2 | 3 | 8007 | 14.79 | `<run-root>/production/mt-incomplete-then-create/turn-2/events.jsonl` |
| `sdk-baseline` | `create-note` | `completed` | 30 | 30 | 8 | 27003 | 67.66 | `<run-root>/sdk-baseline/create-note/turn-1/events.jsonl` |
| `sdk-baseline` | `search-synthesis` | `completed` | 34 | 34 | 12 | 33059 | 85.23 | `<run-root>/sdk-baseline/search-synthesis/turn-1/events.jsonl` |
| `sdk-baseline` | `answer-filing` | `completed` | 30 | 30 | 8 | 13567 | 43.48 | `<run-root>/sdk-baseline/answer-filing/turn-1/events.jsonl` |
| `sdk-baseline` | `stale-synthesis-update` | `completed` | 28 | 28 | 7 | 9930 | 34.71 | `<run-root>/sdk-baseline/stale-synthesis-update/turn-1/events.jsonl` |
| `sdk-baseline` | `append-replace` | `completed` | 20 | 20 | 7 | 7544 | 27.46 | `<run-root>/sdk-baseline/append-replace/turn-1/events.jsonl` |
| `sdk-baseline` | `records-provenance` | `failed` | 78 | 78 | 14 | 51249 | 221.89 | `<run-root>/sdk-baseline/records-provenance/turn-1/events.jsonl` |
| `sdk-baseline` | `promoted-record-vs-docs` | `completed` | 30 | 30 | 8 | 59287 | 59.79 | `<run-root>/sdk-baseline/promoted-record-vs-docs/turn-1/events.jsonl` |
| `sdk-baseline` | `missing-document-path-reject` | `failed` | 36 | 36 | 5 | 16342 | 43.58 | `<run-root>/sdk-baseline/missing-document-path-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `negative-limit-reject` | `failed` | 36 | 36 | 8 | 30480 | 39.20 | `<run-root>/sdk-baseline/negative-limit-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `unsupported-lower-level-reject` | `failed` | 50 | 50 | 10 | 20620 | 58.47 | `<run-root>/sdk-baseline/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `unsupported-cli-mcp-reject` | `failed` | 30 | 30 | 7 | 19605 | 36.08 | `<run-root>/sdk-baseline/unsupported-cli-mcp-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `duplicate-path-reject` | `completed` | 24 | 24 | 6 | 18302 | 39.39 | `<run-root>/sdk-baseline/duplicate-path-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `mixed-synthesis-records` | `failed` | 60 | 60 | 15 | 35517 | 111.63 | `<run-root>/sdk-baseline/mixed-synthesis-records/turn-1/events.jsonl` |
| `sdk-baseline` | `mt-source-then-synthesis` | `completed` | 46 | 46 | 13 | 96159 | 89.55 | `<run-root>/sdk-baseline/mt-source-then-synthesis/turn-2/events.jsonl` |
| `sdk-baseline` | `mt-incomplete-then-create` | `completed` | 18 | 18 | 5 | 15705 | 28.00 | `<run-root>/sdk-baseline/mt-incomplete-then-create/turn-2/events.jsonl` |
