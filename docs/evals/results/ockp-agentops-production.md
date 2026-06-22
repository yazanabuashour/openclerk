# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `agentops-production`
- Release blocking: `true`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `9.31`
- Harness elapsed seconds: `271.30`
- Effective parallel speedup: `3.20x`
- Parallel efficiency: `0.80`
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
| prepare_run_dir | 0.00 |
| copy_repo | 0.04 |
| install_variant | 73.92 |
| warm_cache | 0.00 |
| seed_data | 0.14 |
| agent_run | 867.34 |
| parse_metrics | 0.00 |
| verify | 0.17 |
| total | 941.75 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 3 | 3 | 4 | 25689 | 37.97 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `search-synthesis` | `completed` | 4 | 4 | 4 | 48112 | 20.10 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `answer-filing` | `completed` | 4 | 4 | 5 | 30698 | 55.40 | `<run-root>/production/answer-filing/turn-1/events.jsonl` |
| `production` | `rag-retrieval-baseline` | `completed` | 7 | 7 | 5 | 88670 | 31.65 | `<run-root>/production/rag-retrieval-baseline/turn-2/events.jsonl` |
| `production` | `canonical-docs-navigation-baseline` | `completed` | 5 | 5 | 3 | 25798 | 29.43 | `<run-root>/production/canonical-docs-navigation-baseline/turn-1/events.jsonl` |
| `production` | `graph-semantics-reference-poc` | `completed` | 8 | 8 | 4 | 54466 | 48.01 | `<run-root>/production/graph-semantics-reference-poc/turn-1/events.jsonl` |
| `production` | `memory-router-reference-poc` | `completed` | 12 | 12 | 8 | 104038 | 57.44 | `<run-root>/production/memory-router-reference-poc/turn-2/events.jsonl` |
| `production` | `configured-layout-explain` | `completed` | 2 | 2 | 3 | 30336 | 15.60 | `<run-root>/production/configured-layout-explain/turn-1/events.jsonl` |
| `production` | `invalid-layout-visible` | `completed` | 2 | 2 | 3 | 8568 | 16.11 | `<run-root>/production/invalid-layout-visible/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 8 | 8 | 7 | 33830 | 61.43 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `synthesis-freshness-repair` | `completed` | 8 | 8 | 4 | 51727 | 34.81 | `<run-root>/production/synthesis-freshness-repair/turn-1/events.jsonl` |
| `production` | `source-sensitive-audit-repair` | `completed` | 8 | 8 | 5 | 15968 | 37.64 | `<run-root>/production/source-sensitive-audit-repair/turn-1/events.jsonl` |
| `production` | `source-sensitive-conflict-explain` | `completed` | 4 | 4 | 4 | 12610 | 24.81 | `<run-root>/production/source-sensitive-conflict-explain/turn-1/events.jsonl` |
| `production` | `synthesis-candidate-pressure` | `completed` | 7 | 7 | 6 | 32419 | 31.77 | `<run-root>/production/synthesis-candidate-pressure/turn-1/events.jsonl` |
| `production` | `synthesis-source-set-pressure` | `completed` | 3 | 3 | 4 | 5864 | 21.31 | `<run-root>/production/synthesis-source-set-pressure/turn-1/events.jsonl` |
| `production` | `append-replace` | `completed` | 4 | 4 | 4 | 26165 | 19.32 | `<run-root>/production/append-replace/turn-1/events.jsonl` |
| `production` | `records-provenance` | `completed` | 4 | 4 | 3 | 31856 | 24.50 | `<run-root>/production/records-provenance/turn-1/events.jsonl` |
| `production` | `promoted-record-vs-docs` | `completed` | 3 | 3 | 3 | 8559 | 26.80 | `<run-root>/production/promoted-record-vs-docs/turn-1/events.jsonl` |
| `production` | `decision-record-vs-docs` | `completed` | 4 | 4 | 3 | 30656 | 26.53 | `<run-root>/production/decision-record-vs-docs/turn-1/events.jsonl` |
| `production` | `decision-supersession-freshness` | `completed` | 6 | 6 | 3 | 10063 | 26.63 | `<run-root>/production/decision-supersession-freshness/turn-1/events.jsonl` |
| `production` | `decision-real-adr-migration` | `completed` | 6 | 6 | 4 | 50496 | 39.41 | `<run-root>/production/decision-real-adr-migration/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 3534 | 7.16 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 21424 | 6.74 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 3336 | 7.04 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 21255 | 6.31 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |
| `production` | `duplicate-path-reject` | `completed` | 2 | 2 | 3 | 11716 | 15.01 | `<run-root>/production/duplicate-path-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 6 | 6 | 4 | 11380 | 25.23 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 4 | 4 | 6 | 66238 | 25.24 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-synthesis-drift-pressure` | `completed` | 12 | 12 | 8 | 37428 | 55.42 | `<run-root>/production/mt-synthesis-drift-pressure/turn-2/events.jsonl` |
| `production` | `mt-incomplete-then-create` | `completed` | 2 | 2 | 4 | 68666 | 32.52 | `<run-root>/production/mt-incomplete-then-create/turn-2/events.jsonl` |
