# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `agentops-production`
- Release blocking: `true`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `15.98`
- Harness elapsed seconds: `181.11`
- Effective parallel speedup: `2.88x`
- Parallel efficiency: `0.72`
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
| copy_repo | 1.25 |
| install_variant | 82.22 |
| warm_cache | 0.00 |
| seed_data | 0.44 |
| agent_run | 521.69 |
| parse_metrics | 0.01 |
| verify | 0.42 |
| total | 606.06 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 6 | 6 | 3 | 9869 | 21.37 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `search-synthesis` | `completed` | 6 | 6 | 4 | 7234 | 11.84 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `answer-filing` | `completed` | 4 | 4 | 3 | 6719 | 7.65 | `<run-root>/production/answer-filing/turn-1/events.jsonl` |
| `production` | `rag-retrieval-baseline` | `completed` | 14 | 14 | 5 | 22447 | 24.01 | `<run-root>/production/rag-retrieval-baseline/turn-2/events.jsonl` |
| `production` | `canonical-docs-navigation-baseline` | `completed` | 10 | 10 | 4 | 10936 | 15.32 | `<run-root>/production/canonical-docs-navigation-baseline/turn-1/events.jsonl` |
| `production` | `graph-semantics-reference-poc` | `completed` | 16 | 16 | 4 | 14924 | 29.97 | `<run-root>/production/graph-semantics-reference-poc/turn-1/events.jsonl` |
| `production` | `memory-router-reference-poc` | `completed` | 24 | 24 | 8 | 29800 | 40.03 | `<run-root>/production/memory-router-reference-poc/turn-2/events.jsonl` |
| `production` | `configured-layout-explain` | `completed` | 2 | 2 | 2 | 4399 | 10.41 | `<run-root>/production/configured-layout-explain/turn-1/events.jsonl` |
| `production` | `invalid-layout-visible` | `completed` | 2 | 2 | 3 | 5360 | 10.37 | `<run-root>/production/invalid-layout-visible/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 20 | 20 | 5 | 20339 | 33.75 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `synthesis-freshness-repair` | `completed` | 34 | 34 | 10 | 40661 | 47.50 | `<run-root>/production/synthesis-freshness-repair/turn-1/events.jsonl` |
| `production` | `source-sensitive-audit-repair` | `completed` | 14 | 14 | 7 | 9946 | 21.92 | `<run-root>/production/source-sensitive-audit-repair/turn-1/events.jsonl` |
| `production` | `source-sensitive-conflict-explain` | `completed` | 6 | 6 | 3 | 8826 | 14.38 | `<run-root>/production/source-sensitive-conflict-explain/turn-1/events.jsonl` |
| `production` | `synthesis-candidate-pressure` | `completed` | 6 | 6 | 4 | 8141 | 17.52 | `<run-root>/production/synthesis-candidate-pressure/turn-1/events.jsonl` |
| `production` | `synthesis-source-set-pressure` | `completed` | 6 | 6 | 4 | 7682 | 13.85 | `<run-root>/production/synthesis-source-set-pressure/turn-1/events.jsonl` |
| `production` | `append-replace` | `completed` | 8 | 8 | 6 | 8988 | 16.85 | `<run-root>/production/append-replace/turn-1/events.jsonl` |
| `production` | `records-provenance` | `completed` | 6 | 6 | 2 | 10055 | 17.89 | `<run-root>/production/records-provenance/turn-1/events.jsonl` |
| `production` | `promoted-record-vs-docs` | `completed` | 4 | 4 | 3 | 6995 | 12.63 | `<run-root>/production/promoted-record-vs-docs/turn-1/events.jsonl` |
| `production` | `decision-record-vs-docs` | `completed` | 6 | 6 | 2 | 7919 | 15.94 | `<run-root>/production/decision-record-vs-docs/turn-1/events.jsonl` |
| `production` | `decision-supersession-freshness` | `completed` | 8 | 8 | 4 | 9390 | 20.41 | `<run-root>/production/decision-supersession-freshness/turn-1/events.jsonl` |
| `production` | `decision-real-adr-migration` | `completed` | 10 | 10 | 4 | 9121 | 16.78 | `<run-root>/production/decision-real-adr-migration/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2472 | 3.88 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2454 | 4.39 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2484 | 3.97 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2483 | 4.36 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |
| `production` | `duplicate-path-reject` | `completed` | 2 | 2 | 2 | 3517 | 9.80 | `<run-root>/production/duplicate-path-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 10 | 10 | 4 | 8880 | 16.11 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 6 | 6 | 5 | 14543 | 14.56 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-synthesis-drift-pressure` | `completed` | 24 | 24 | 6 | 26489 | 33.95 | `<run-root>/production/mt-synthesis-drift-pressure/turn-2/events.jsonl` |
| `production` | `mt-incomplete-then-create` | `completed` | 2 | 2 | 3 | 7096 | 10.28 | `<run-root>/production/mt-incomplete-then-create/turn-2/events.jsonl` |
