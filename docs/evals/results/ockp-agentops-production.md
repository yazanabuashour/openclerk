# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `agentops-production`
- Release blocking: `true`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `59.36`
- Harness elapsed seconds: `303.93`
- Effective parallel speedup: `2.25x`
- Parallel efficiency: `0.56`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `true`

Recommendation: `use_agentops_runner_for_routine_openclerk_operations`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `pass` | 30/30 production scenarios passed |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect retired API files or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `pass` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_legacy_source_runner_usage` | `pass` | production must not invoke source-built or legacy runner paths instead of installed openclerk |
| `no_direct_sqlite_access` | `pass` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `pass` | rule-covered validation scenarios used no tools, no command executions, and at most one assistant answer |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.01 |
| copy_repo | 2.53 |
| install_variant | 261.66 |
| warm_cache | 0.00 |
| seed_data | 0.46 |
| agent_run | 684.10 |
| parse_metrics | 0.01 |
| verify | 0.41 |
| total | 949.19 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 7 | 7 | 6 | 13950 | 29.68 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `search-synthesis` | `completed` | 3 | 3 | 4 | 9740 | 13.82 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `answer-filing` | `completed` | 2 | 2 | 3 | 6345 | 13.68 | `<run-root>/production/answer-filing/turn-1/events.jsonl` |
| `production` | `rag-retrieval-baseline` | `completed` | 7 | 7 | 5 | 33861 | 28.99 | `<run-root>/production/rag-retrieval-baseline/turn-2/events.jsonl` |
| `production` | `canonical-docs-navigation-baseline` | `completed` | 5 | 5 | 4 | 13219 | 22.29 | `<run-root>/production/canonical-docs-navigation-baseline/turn-1/events.jsonl` |
| `production` | `graph-semantics-reference-poc` | `completed` | 7 | 7 | 3 | 12689 | 32.50 | `<run-root>/production/graph-semantics-reference-poc/turn-1/events.jsonl` |
| `production` | `memory-router-reference-poc` | `completed` | 12 | 12 | 8 | 22795 | 38.83 | `<run-root>/production/memory-router-reference-poc/turn-2/events.jsonl` |
| `production` | `configured-layout-explain` | `completed` | 1 | 1 | 2 | 16755 | 10.87 | `<run-root>/production/configured-layout-explain/turn-1/events.jsonl` |
| `production` | `invalid-layout-visible` | `completed` | 2 | 2 | 3 | 9792 | 12.95 | `<run-root>/production/invalid-layout-visible/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 19 | 19 | 10 | 20247 | 86.27 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `synthesis-freshness-repair` | `completed` | 8 | 8 | 5 | 13694 | 25.17 | `<run-root>/production/synthesis-freshness-repair/turn-1/events.jsonl` |
| `production` | `source-sensitive-audit-repair` | `completed` | 7 | 7 | 4 | 10273 | 31.11 | `<run-root>/production/source-sensitive-audit-repair/turn-1/events.jsonl` |
| `production` | `source-sensitive-conflict-explain` | `completed` | 3 | 3 | 3 | 6794 | 15.58 | `<run-root>/production/source-sensitive-conflict-explain/turn-1/events.jsonl` |
| `production` | `synthesis-candidate-pressure` | `completed` | 7 | 7 | 6 | 12008 | 22.38 | `<run-root>/production/synthesis-candidate-pressure/turn-1/events.jsonl` |
| `production` | `synthesis-source-set-pressure` | `completed` | 3 | 3 | 2 | 4253 | 18.53 | `<run-root>/production/synthesis-source-set-pressure/turn-1/events.jsonl` |
| `production` | `append-replace` | `completed` | 3 | 3 | 3 | 4278 | 17.88 | `<run-root>/production/append-replace/turn-1/events.jsonl` |
| `production` | `records-provenance` | `completed` | 5 | 5 | 4 | 11555 | 19.85 | `<run-root>/production/records-provenance/turn-1/events.jsonl` |
| `production` | `promoted-record-vs-docs` | `completed` | 2 | 2 | 3 | 6337 | 17.12 | `<run-root>/production/promoted-record-vs-docs/turn-1/events.jsonl` |
| `production` | `decision-record-vs-docs` | `completed` | 1 | 1 | 2 | 6378 | 13.42 | `<run-root>/production/decision-record-vs-docs/turn-1/events.jsonl` |
| `production` | `decision-supersession-freshness` | `completed` | 6 | 6 | 3 | 10531 | 18.52 | `<run-root>/production/decision-supersession-freshness/turn-1/events.jsonl` |
| `production` | `decision-real-adr-migration` | `completed` | 5 | 5 | 3 | 5671 | 23.67 | `<run-root>/production/decision-real-adr-migration/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2434 | 6.87 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2443 | 7.89 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2473 | 7.95 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2445 | 9.81 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |
| `production` | `duplicate-path-reject` | `completed` | 2 | 2 | 3 | 8960 | 17.69 | `<run-root>/production/duplicate-path-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 6 | 6 | 4 | 9803 | 24.78 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 5 | 5 | 7 | 19978 | 35.41 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-synthesis-drift-pressure` | `completed` | 11 | 11 | 10 | 19474 | 36.11 | `<run-root>/production/mt-synthesis-drift-pressure/turn-2/events.jsonl` |
| `production` | `mt-incomplete-then-create` | `completed` | 2 | 2 | 4 | 9604 | 24.48 | `<run-root>/production/mt-incomplete-then-create/turn-2/events.jsonl` |
