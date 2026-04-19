# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `24.14`
- Harness elapsed seconds: `387.33`
- Effective parallel speedup: `3.47x`
- Parallel efficiency: `0.87`
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
| `total_tools_less_than_or_equal_baseline` | `pass` | production tools 112 vs baseline tools 576; missing baseline: ok |
| `minimum_scenarios_at_or_below_baseline` | `pass` | 15 scenarios at or below baseline tools; required 12 of 15 |
| `non_cached_token_majority` | `pass` | 15 scenarios with lower non-cached input tokens; required 8 of 15; missing usage: ok |
| `non_cached_token_total_less_than_or_equal_baseline` | `pass` | production non-cached input tokens 108416 vs baseline 421057; missing usage: ok |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.48 |
| install_variant | 21.80 |
| warm_cache | 0.00 |
| seed_data | 0.36 |
| agent_run | 1343.01 |
| parse_metrics | 0.01 |
| verify | 0.49 |
| total | 1366.17 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 6 | 6 | 4 | 5046 | 21.76 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `search-synthesis` | `completed` | 22 | 22 | 6 | 14034 | 47.27 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `answer-filing` | `completed` | 18 | 18 | 5 | 12882 | 36.70 | `<run-root>/production/answer-filing/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 8 | 8 | 5 | 6097 | 26.87 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `append-replace` | `completed` | 6 | 6 | 3 | 4808 | 17.05 | `<run-root>/production/append-replace/turn-1/events.jsonl` |
| `production` | `records-provenance` | `completed` | 10 | 10 | 3 | 6128 | 23.40 | `<run-root>/production/records-provenance/turn-1/events.jsonl` |
| `production` | `promoted-record-vs-docs` | `completed` | 6 | 6 | 2 | 5546 | 18.95 | `<run-root>/production/promoted-record-vs-docs/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 3650 | 4.68 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 3638 | 5.14 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 3656 | 4.10 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-cli-mcp-reject` | `completed` | 0 | 0 | 1 | 3668 | 4.39 | `<run-root>/production/unsupported-cli-mcp-reject/turn-1/events.jsonl` |
| `production` | `duplicate-path-reject` | `completed` | 2 | 2 | 2 | 4057 | 10.03 | `<run-root>/production/duplicate-path-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 24 | 24 | 6 | 17951 | 71.81 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 8 | 8 | 6 | 9459 | 25.61 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-incomplete-then-create` | `completed` | 2 | 2 | 3 | 7796 | 16.24 | `<run-root>/production/mt-incomplete-then-create/turn-2/events.jsonl` |
| `sdk-baseline` | `create-note` | `completed` | 38 | 38 | 11 | 49707 | 73.50 | `<run-root>/sdk-baseline/create-note/turn-1/events.jsonl` |
| `sdk-baseline` | `search-synthesis` | `completed` | 60 | 60 | 13 | 35520 | 109.12 | `<run-root>/sdk-baseline/search-synthesis/turn-1/events.jsonl` |
| `sdk-baseline` | `answer-filing` | `completed` | 36 | 36 | 9 | 15930 | 58.27 | `<run-root>/sdk-baseline/answer-filing/turn-1/events.jsonl` |
| `sdk-baseline` | `stale-synthesis-update` | `completed` | 42 | 42 | 9 | 28217 | 128.61 | `<run-root>/sdk-baseline/stale-synthesis-update/turn-1/events.jsonl` |
| `sdk-baseline` | `append-replace` | `completed` | 42 | 42 | 10 | 24122 | 62.57 | `<run-root>/sdk-baseline/append-replace/turn-1/events.jsonl` |
| `sdk-baseline` | `records-provenance` | `completed` | 60 | 60 | 12 | 45201 | 112.14 | `<run-root>/sdk-baseline/records-provenance/turn-1/events.jsonl` |
| `sdk-baseline` | `promoted-record-vs-docs` | `completed` | 30 | 30 | 6 | 22574 | 46.67 | `<run-root>/sdk-baseline/promoted-record-vs-docs/turn-1/events.jsonl` |
| `sdk-baseline` | `missing-document-path-reject` | `failed` | 18 | 18 | 5 | 21858 | 27.07 | `<run-root>/sdk-baseline/missing-document-path-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `negative-limit-reject` | `failed` | 22 | 22 | 4 | 19602 | 27.85 | `<run-root>/sdk-baseline/negative-limit-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `unsupported-lower-level-reject` | `failed` | 44 | 44 | 9 | 21451 | 59.90 | `<run-root>/sdk-baseline/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `unsupported-cli-mcp-reject` | `failed` | 16 | 16 | 5 | 12178 | 26.48 | `<run-root>/sdk-baseline/unsupported-cli-mcp-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `duplicate-path-reject` | `completed` | 32 | 32 | 7 | 20753 | 59.14 | `<run-root>/sdk-baseline/duplicate-path-reject/turn-1/events.jsonl` |
| `sdk-baseline` | `mixed-synthesis-records` | `failed` | 50 | 50 | 9 | 30526 | 105.77 | `<run-root>/sdk-baseline/mixed-synthesis-records/turn-1/events.jsonl` |
| `sdk-baseline` | `mt-source-then-synthesis` | `completed` | 54 | 54 | 13 | 45662 | 69.09 | `<run-root>/sdk-baseline/mt-source-then-synthesis/turn-2/events.jsonl` |
| `sdk-baseline` | `mt-incomplete-then-create` | `completed` | 32 | 32 | 8 | 27756 | 42.83 | `<run-root>/sdk-baseline/mt-incomplete-then-create/turn-2/events.jsonl` |
