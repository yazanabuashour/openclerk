# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `agentops-production`
- Release blocking: `true`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `16.81`
- Harness elapsed seconds: `278.35`
- Effective parallel speedup: `3.15x`
- Parallel efficiency: `0.79`
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
| copy_repo | 2.60 |
| install_variant | 85.34 |
| warm_cache | 0.00 |
| seed_data | 0.61 |
| agent_run | 876.10 |
| parse_metrics | 0.00 |
| verify | 0.78 |
| total | 965.46 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `create-note` | `completed` | 3 | 3 | 4 | 8799 | 35.82 | `<run-root>/production/create-note/turn-1/events.jsonl` |
| `production` | `search-synthesis` | `completed` | 4 | 4 | 4 | 28975 | 27.97 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `answer-filing` | `completed` | 3 | 3 | 4 | 32240 | 19.88 | `<run-root>/production/answer-filing/turn-1/events.jsonl` |
| `production` | `rag-retrieval-baseline` | `completed` | 7 | 7 | 5 | 78227 | 35.69 | `<run-root>/production/rag-retrieval-baseline/turn-2/events.jsonl` |
| `production` | `canonical-docs-navigation-baseline` | `completed` | 6 | 6 | 4 | 31400 | 30.35 | `<run-root>/production/canonical-docs-navigation-baseline/turn-1/events.jsonl` |
| `production` | `graph-semantics-reference-poc` | `completed` | 7 | 7 | 4 | 32781 | 51.16 | `<run-root>/production/graph-semantics-reference-poc/turn-1/events.jsonl` |
| `production` | `memory-router-reference-poc` | `completed` | 12 | 12 | 8 | 70059 | 53.95 | `<run-root>/production/memory-router-reference-poc/turn-2/events.jsonl` |
| `production` | `configured-layout-explain` | `completed` | 1 | 1 | 2 | 5358 | 21.64 | `<run-root>/production/configured-layout-explain/turn-1/events.jsonl` |
| `production` | `invalid-layout-visible` | `completed` | 2 | 2 | 3 | 9507 | 18.66 | `<run-root>/production/invalid-layout-visible/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 8 | 8 | 5 | 15469 | 51.46 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `synthesis-freshness-repair` | `completed` | 10 | 10 | 5 | 34787 | 56.41 | `<run-root>/production/synthesis-freshness-repair/turn-1/events.jsonl` |
| `production` | `source-sensitive-audit-repair` | `completed` | 7 | 7 | 4 | 27791 | 36.26 | `<run-root>/production/source-sensitive-audit-repair/turn-1/events.jsonl` |
| `production` | `source-sensitive-conflict-explain` | `completed` | 4 | 4 | 4 | 28756 | 23.00 | `<run-root>/production/source-sensitive-conflict-explain/turn-1/events.jsonl` |
| `production` | `synthesis-candidate-pressure` | `completed` | 7 | 7 | 5 | 11407 | 34.93 | `<run-root>/production/synthesis-candidate-pressure/turn-1/events.jsonl` |
| `production` | `synthesis-source-set-pressure` | `completed` | 4 | 4 | 3 | 33169 | 42.87 | `<run-root>/production/synthesis-source-set-pressure/turn-1/events.jsonl` |
| `production` | `append-replace` | `completed` | 3 | 3 | 3 | 9216 | 20.28 | `<run-root>/production/append-replace/turn-1/events.jsonl` |
| `production` | `records-provenance` | `completed` | 3 | 3 | 4 | 10640 | 28.44 | `<run-root>/production/records-provenance/turn-1/events.jsonl` |
| `production` | `promoted-record-vs-docs` | `completed` | 3 | 3 | 3 | 13007 | 22.33 | `<run-root>/production/promoted-record-vs-docs/turn-1/events.jsonl` |
| `production` | `decision-record-vs-docs` | `completed` | 4 | 4 | 3 | 28378 | 20.61 | `<run-root>/production/decision-record-vs-docs/turn-1/events.jsonl` |
| `production` | `decision-supersession-freshness` | `completed` | 6 | 6 | 3 | 28776 | 26.88 | `<run-root>/production/decision-supersession-freshness/turn-1/events.jsonl` |
| `production` | `decision-real-adr-migration` | `completed` | 6 | 6 | 3 | 34470 | 28.76 | `<run-root>/production/decision-real-adr-migration/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 3671 | 5.60 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 18253 | 4.76 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 3693 | 6.30 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 3682 | 4.86 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |
| `production` | `duplicate-path-reject` | `completed` | 2 | 2 | 3 | 12080 | 22.56 | `<run-root>/production/duplicate-path-reject/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 6 | 6 | 4 | 15085 | 35.25 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 4 | 4 | 6 | 61248 | 27.01 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
| `production` | `mt-synthesis-drift-pressure` | `completed` | 12 | 12 | 8 | 97681 | 58.10 | `<run-root>/production/mt-synthesis-drift-pressure/turn-2/events.jsonl` |
| `production` | `mt-incomplete-then-create` | `completed` | 2 | 2 | 4 | 18152 | 24.31 | `<run-root>/production/mt-incomplete-then-create/turn-2/events.jsonl` |
