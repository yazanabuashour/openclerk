# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `16.65`
- Harness elapsed seconds: `88.19`
- Effective parallel speedup: `2.80x`
- Parallel efficiency: `0.70`
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
| copy_repo | 0.25 |
| install_variant | 11.60 |
| warm_cache | 0.00 |
| seed_data | 0.20 |
| agent_run | 247.14 |
| parse_metrics | 0.00 |
| verify | 0.22 |
| total | 259.45 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 6 | 6 | 4 | 4827 | 24.40 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `search-synthesis` | `completed` | 10 | 10 | 6 | 5765 | 26.46 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `answer-filing` | `completed` | 4 | 4 | 3 | 4922 | 16.14 | `<run-root>/production/answer-filing/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 8 | 8 | 5 | 5946 | 23.68 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `append-replace` | `completed` | 6 | 6 | 3 | 4962 | 22.04 | `<run-root>/production/append-replace/turn-1/events.jsonl` |
| `production` | `records-provenance` | `completed` | 10 | 10 | 4 | 5891 | 26.81 | `<run-root>/production/records-provenance/turn-1/events.jsonl` |
| `production` | `promoted-record-vs-docs` | `completed` | 6 | 6 | 3 | 5946 | 17.92 | `<run-root>/production/promoted-record-vs-docs/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 3650 | 4.46 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 3638 | 4.80 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 3656 | 5.81 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-cli-mcp-reject` | `completed` | 0 | 0 | 1 | 3156 | 3.90 | `<run-root>/production/unsupported-cli-mcp-reject/turn-1/events.jsonl` |
| `production` | `duplicate-path-reject` | `completed` | 2 | 2 | 2 | 4166 | 10.12 | `<run-root>/production/duplicate-path-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 8 | 8 | 4 | 5709 | 21.12 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 8 | 8 | 6 | 9384 | 26.16 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-incomplete-then-create` | `completed` | 2 | 2 | 3 | 7901 | 13.32 | `<run-root>/production/mt-incomplete-then-create/turn-2/events.jsonl` |
