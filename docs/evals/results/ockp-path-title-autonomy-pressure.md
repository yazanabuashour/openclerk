# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `path-title-autonomy-pressure`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `17.61`
- Harness elapsed seconds: `151.15`
- Effective parallel speedup: `0.75x`
- Parallel efficiency: `0.75`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 0/30 production scenarios passed; missing: create-note, search-synthesis, answer-filing, rag-retrieval-baseline, canonical-docs-navigation-baseline, graph-semantics-reference-poc, memory-router-reference-poc, configured-layout-explain, invalid-layout-visible, stale-synthesis-update, synthesis-freshness-repair, source-sensitive-audit-repair, source-sensitive-conflict-explain, synthesis-candidate-pressure, synthesis-source-set-pressure, append-replace, records-provenance, promoted-record-vs-docs, decision-record-vs-docs, decision-supersession-freshness, decision-real-adr-migration, missing-document-path-reject, negative-limit-reject, unsupported-lower-level-reject, unsupported-transport-reject, duplicate-path-reject, mixed-synthesis-records, mt-source-then-synthesis, mt-synthesis-drift-pressure, mt-incomplete-then-create |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect retired API files or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `pass` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_legacy_source_runner_usage` | `pass` | production must not invoke source-built or legacy runner paths instead of installed openclerk |
| `no_direct_sqlite_access` | `pass` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `fail` | not evaluated; final-answer-only validation scenarios were not selected in this partial run |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.17 |
| install_variant | 20.42 |
| warm_cache | 0.00 |
| seed_data | 0.06 |
| agent_run | 112.73 |
| parse_metrics | 0.00 |
| verify | 0.14 |
| total | 133.54 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `path-title-url-only-autonomy-pressure` | `completed` | 4 | 4 | 3 | 6738 | 12.30 | `<run-root>/production/path-title-url-only-autonomy-pressure/turn-1/events.jsonl` |
| `production` | `path-title-artifact-missing-hints` | `completed` | 0 | 0 | 1 | 2453 | 5.76 | `<run-root>/production/path-title-artifact-missing-hints/turn-1/events.jsonl` |
| `production` | `path-title-multisource-duplicate-pressure` | `completed` | 16 | 16 | 6 | 9714 | 36.55 | `<run-root>/production/path-title-multisource-duplicate-pressure/turn-1/events.jsonl` |
| `production` | `path-title-explicit-overrides-pressure` | `completed` | 6 | 6 | 4 | 3826 | 18.40 | `<run-root>/production/path-title-explicit-overrides-pressure/turn-1/events.jsonl` |
| `production` | `path-title-duplicate-risk-pressure` | `completed` | 8 | 8 | 3 | 4734 | 18.77 | `<run-root>/production/path-title-duplicate-risk-pressure/turn-1/events.jsonl` |
| `production` | `path-title-metadata-authority-pressure` | `completed` | 8 | 8 | 3 | 4606 | 20.95 | `<run-root>/production/path-title-metadata-authority-pressure/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `evaluate_for_oc_iat`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: no promoted runner action, schema, migration, skill behavior, storage API, product behavior, or public OpenClerk interface from this eval.

| Variant | Scenario | Status | Failure classification | Evidence posture |
| --- | --- | --- | --- | --- |
| `production` | `path-title-url-only-autonomy-pressure` | `completed` | `none` | current runner/skill behavior handled path/title autonomy pressure |
| `production` | `path-title-artifact-missing-hints` | `completed` | `none` | current runner/skill behavior handled path/title autonomy pressure |
| `production` | `path-title-multisource-duplicate-pressure` | `completed` | `none` | current runner/skill behavior handled path/title autonomy pressure |
| `production` | `path-title-explicit-overrides-pressure` | `completed` | `none` | current runner/skill behavior handled path/title autonomy pressure |
| `production` | `path-title-duplicate-risk-pressure` | `completed` | `none` | current runner/skill behavior handled path/title autonomy pressure |
| `production` | `path-title-metadata-authority-pressure` | `completed` | `none` | current runner/skill behavior handled path/title autonomy pressure |
