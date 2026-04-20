# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `19.47`
- Harness elapsed seconds: `151.02`
- Effective parallel speedup: `2.95x`
- Parallel efficiency: `0.74`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `true`

Recommendation: `use_agentops_runner_for_routine_openclerk_operations`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `pass` | 15/15 production scenarios passed |
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
| copy_repo | 0.20 |
| install_variant | 8.88 |
| warm_cache | 0.00 |
| seed_data | 0.13 |
| agent_run | 445.81 |
| parse_metrics | 0.00 |
| verify | 0.25 |
| total | 455.32 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 6 | 6 | 4 | 5565 | 20.31 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `search-synthesis` | `completed` | 10 | 10 | 4 | 5723 | 25.40 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `answer-filing` | `completed` | 36 | 36 | 6 | 10740 | 82.12 | `<run-root>/production/answer-filing/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 10 | 10 | 4 | 7334 | 29.14 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `append-replace` | `completed` | 12 | 12 | 4 | 6476 | 25.51 | `<run-root>/production/append-replace/turn-1/events.jsonl` |
| `production` | `records-provenance` | `completed` | 10 | 10 | 4 | 5939 | 26.15 | `<run-root>/production/records-provenance/turn-1/events.jsonl` |
| `production` | `promoted-record-vs-docs` | `completed` | 14 | 14 | 5 | 6687 | 65.78 | `<run-root>/production/promoted-record-vs-docs/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 3829 | 7.93 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 3817 | 6.45 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 3835 | 3.83 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-cli-mcp-reject` | `completed` | 0 | 0 | 1 | 3847 | 9.35 | `<run-root>/production/unsupported-cli-mcp-reject/turn-1/events.jsonl` |
| `production` | `duplicate-path-reject` | `completed` | 2 | 2 | 2 | 4021 | 9.36 | `<run-root>/production/duplicate-path-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 22 | 22 | 6 | 9181 | 61.05 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 12 | 12 | 8 | 10589 | 46.17 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-incomplete-then-create` | `completed` | 2 | 2 | 3 | 8536 | 27.26 | `<run-root>/production/mt-incomplete-then-create/turn-2/events.jsonl` |
