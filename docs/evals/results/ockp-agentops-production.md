# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `agentops-production`
- Release blocking: `true`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `16.93`
- Harness elapsed seconds: `431.44`
- Effective parallel speedup: `2.99x`
- Parallel efficiency: `0.75`
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
| copy_repo | 0.75 |
| install_variant | 82.63 |
| warm_cache | 0.00 |
| seed_data | 0.83 |
| agent_run | 1290.93 |
| parse_metrics | 0.02 |
| verify | 1.53 |
| total | 1376.80 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 6 | 6 | 4 | 7228 | 17.09 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `search-synthesis` | `completed` | 26 | 26 | 8 | 12586 | 53.08 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `answer-filing` | `completed` | 10 | 10 | 5 | 11919 | 23.52 | `<run-root>/production/answer-filing/turn-1/events.jsonl` |
| `production` | `rag-retrieval-baseline` | `completed` | 16 | 16 | 6 | 64190 | 30.31 | `<run-root>/production/rag-retrieval-baseline/turn-2/events.jsonl` |
| `production` | `canonical-docs-navigation-baseline` | `completed` | 78 | 78 | 10 | 55744 | 177.58 | `<run-root>/production/canonical-docs-navigation-baseline/turn-1/events.jsonl` |
| `production` | `graph-semantics-reference-poc` | `completed` | 16 | 16 | 3 | 14153 | 35.63 | `<run-root>/production/graph-semantics-reference-poc/turn-1/events.jsonl` |
| `production` | `memory-router-reference-poc` | `completed` | 28 | 28 | 8 | 22873 | 64.29 | `<run-root>/production/memory-router-reference-poc/turn-2/events.jsonl` |
| `production` | `configured-layout-explain` | `completed` | 4 | 4 | 3 | 7576 | 15.84 | `<run-root>/production/configured-layout-explain/turn-1/events.jsonl` |
| `production` | `invalid-layout-visible` | `completed` | 144 | 144 | 14 | 16626 | 163.30 | `<run-root>/production/invalid-layout-visible/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 16 | 16 | 5 | 9403 | 49.03 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `synthesis-freshness-repair` | `completed` | 18 | 18 | 5 | 12577 | 39.77 | `<run-root>/production/synthesis-freshness-repair/turn-1/events.jsonl` |
| `production` | `source-sensitive-audit-repair` | `completed` | 20 | 20 | 6 | 13233 | 55.12 | `<run-root>/production/source-sensitive-audit-repair/turn-1/events.jsonl` |
| `production` | `source-sensitive-conflict-explain` | `completed` | 8 | 8 | 4 | 4644 | 18.36 | `<run-root>/production/source-sensitive-conflict-explain/turn-1/events.jsonl` |
| `production` | `synthesis-candidate-pressure` | `completed` | 22 | 22 | 7 | 12133 | 56.39 | `<run-root>/production/synthesis-candidate-pressure/turn-1/events.jsonl` |
| `production` | `synthesis-source-set-pressure` | `completed` | 24 | 24 | 7 | 13151 | 49.02 | `<run-root>/production/synthesis-source-set-pressure/turn-1/events.jsonl` |
| `production` | `append-replace` | `completed` | 8 | 8 | 4 | 7400 | 17.39 | `<run-root>/production/append-replace/turn-1/events.jsonl` |
| `production` | `records-provenance` | `completed` | 20 | 20 | 4 | 21397 | 39.35 | `<run-root>/production/records-provenance/turn-1/events.jsonl` |
| `production` | `promoted-record-vs-docs` | `completed` | 10 | 10 | 6 | 10201 | 27.85 | `<run-root>/production/promoted-record-vs-docs/turn-1/events.jsonl` |
| `production` | `decision-record-vs-docs` | `completed` | 8 | 8 | 4 | 9342 | 27.73 | `<run-root>/production/decision-record-vs-docs/turn-1/events.jsonl` |
| `production` | `decision-supersession-freshness` | `completed` | 14 | 14 | 5 | 11000 | 27.89 | `<run-root>/production/decision-supersession-freshness/turn-1/events.jsonl` |
| `production` | `decision-real-adr-migration` | `completed` | 18 | 18 | 7 | 11849 | 33.73 | `<run-root>/production/decision-real-adr-migration/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2708 | 9.96 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2689 | 3.82 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2887 | 4.31 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2719 | 4.58 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |
| `production` | `duplicate-path-reject` | `completed` | 4 | 4 | 3 | 6473 | 10.51 | `<run-root>/production/duplicate-path-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 26 | 26 | 8 | 17384 | 47.73 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 14 | 14 | 7 | 17860 | 40.14 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-synthesis-drift-pressure` | `completed` | 62 | 62 | 11 | 33507 | 126.54 | `<run-root>/production/mt-synthesis-drift-pressure/turn-2/events.jsonl` |
| `production` | `mt-incomplete-then-create` | `completed` | 6 | 6 | 5 | 9918 | 21.07 | `<run-root>/production/mt-incomplete-then-create/turn-2/events.jsonl` |
