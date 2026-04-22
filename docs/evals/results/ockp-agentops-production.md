# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `24.65`
- Harness elapsed seconds: `168.25`
- Effective parallel speedup: `2.78x`
- Parallel efficiency: `0.70`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 14/18 production scenarios passed |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect retired API files or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `fail` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_legacy_source_runner_usage` | `pass` | production must not invoke source-built or legacy runner paths instead of installed openclerk |
| `no_direct_sqlite_access` | `fail` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `fail` | not final-answer-only: missing-document-path-reject, negative-limit-reject, unsupported-lower-level-reject, unsupported-transport-reject |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.34 |
| install_variant | 42.08 |
| warm_cache | 0.00 |
| seed_data | 0.25 |
| agent_run | 468.13 |
| parse_metrics | 0.00 |
| verify | 0.58 |
| total | 511.36 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 12 | 12 | 5 | 7974 | 32.28 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `search-synthesis` | `completed` | 10 | 10 | 4 | 25372 | 24.92 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `answer-filing` | `completed` | 8 | 8 | 4 | 6763 | 20.57 | `<run-root>/production/answer-filing/turn-1/events.jsonl` |
| `production` | `rag-retrieval-baseline` | `completed` | 18 | 18 | 6 | 25240 | 36.08 | `<run-root>/production/rag-retrieval-baseline/turn-2/events.jsonl` |
| `production` | `canonical-docs-navigation-baseline` | `completed` | 12 | 12 | 4 | 13599 | 37.58 | `<run-root>/production/canonical-docs-navigation-baseline/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 10 | 10 | 5 | 7888 | 22.06 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `synthesis-freshness-repair` | `completed` | 18 | 18 | 6 | 9657 | 54.89 | `<run-root>/production/synthesis-freshness-repair/turn-1/events.jsonl` |
| `production` | `append-replace` | `completed` | 12 | 12 | 4 | 7012 | 22.22 | `<run-root>/production/append-replace/turn-1/events.jsonl` |
| `production` | `records-provenance` | `completed` | 18 | 18 | 6 | 12432 | 42.64 | `<run-root>/production/records-provenance/turn-1/events.jsonl` |
| `production` | `promoted-record-vs-docs` | `completed` | 6 | 6 | 3 | 6718 | 21.13 | `<run-root>/production/promoted-record-vs-docs/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `failed` | 2 | 2 | 2 | 4931 | 13.86 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `failed` | 2 | 2 | 2 | 4866 | 8.12 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `failed` | 4 | 4 | 3 | 5174 | 13.20 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `failed` | 2 | 2 | 2 | 4909 | 12.48 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |
| `production` | `duplicate-path-reject` | `completed` | 4 | 4 | 3 | 5120 | 14.08 | `<run-root>/production/duplicate-path-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 18 | 18 | 5 | 8614 | 43.44 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 10 | 10 | 6 | 13051 | 31.80 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-incomplete-then-create` | `completed` | 4 | 4 | 4 | 10714 | 16.78 | `<run-root>/production/mt-incomplete-then-create/turn-2/events.jsonl` |
