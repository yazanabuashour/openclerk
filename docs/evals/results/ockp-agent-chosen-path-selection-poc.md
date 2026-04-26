# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `agent-chosen-path-selection-poc`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `17.35`
- Harness elapsed seconds: `195.65`
- Effective parallel speedup: `0.74x`
- Parallel efficiency: `0.74`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 4/30 production scenarios passed; missing: create-note, search-synthesis, answer-filing, rag-retrieval-baseline, canonical-docs-navigation-baseline, graph-semantics-reference-poc, memory-router-reference-poc, configured-layout-explain, invalid-layout-visible, stale-synthesis-update, synthesis-freshness-repair, source-sensitive-audit-repair, source-sensitive-conflict-explain, synthesis-candidate-pressure, synthesis-source-set-pressure, append-replace, records-provenance, promoted-record-vs-docs, decision-record-vs-docs, decision-supersession-freshness, decision-real-adr-migration, duplicate-path-reject, mixed-synthesis-records, mt-source-then-synthesis, mt-synthesis-drift-pressure, mt-incomplete-then-create |
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
| copy_repo | 0.18 |
| install_variant | 32.70 |
| warm_cache | 0.00 |
| seed_data | 0.03 |
| agent_run | 145.16 |
| parse_metrics | 0.00 |
| verify | 0.19 |
| total | 178.29 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `explicit-fields-path-title-type` | `completed` | 6 | 6 | 4 | 3583 | 22.50 | `<run-root>/production/explicit-fields-path-title-type/turn-1/events.jsonl` |
| `production` | `missing-path-title-type-reject` | `failed` | 0 | 0 | 1 | 2437 | 6.57 | `<run-root>/production/missing-path-title-type-reject/turn-1/events.jsonl` |
| `production` | `url-only-documentation-path-proposal` | `completed` | 0 | 0 | 1 | 2550 | 6.97 | `<run-root>/production/url-only-documentation-path-proposal/turn-1/events.jsonl` |
| `production` | `url-only-documentation-autonomous-placement` | `failed` | 0 | 0 | 2 | 2596 | 31.26 | `<run-root>/production/url-only-documentation-autonomous-placement/turn-1/events.jsonl` |
| `production` | `multi-source-synthesis-path-selection` | `completed` | 10 | 10 | 4 | 8364 | 27.18 | `<run-root>/production/multi-source-synthesis-path-selection/turn-1/events.jsonl` |
| `production` | `ambiguous-document-type-path-selection` | `failed` | 0 | 0 | 2 | 2688 | 14.98 | `<run-root>/production/ambiguous-document-type-path-selection/turn-1/events.jsonl` |
| `production` | `user-path-instructions-win` | `completed` | 4 | 4 | 3 | 3488 | 13.07 | `<run-root>/production/user-path-instructions-win/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 8215 | 4.65 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 8197 | 6.13 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2429 | 6.99 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2428 | 4.86 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `keep_as_reference`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: no promoted runner action, schema, migration, storage API, product behavior, public OpenClerk interface, or change to missing-path clarification.

| Variant | Scenario | Status | Failure classification | Evidence posture |
| --- | --- | --- | --- | --- |
| `production` | `explicit-fields-path-title-type` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `missing-path-title-type-reject` | `failed` | `skill_guidance_or_eval_coverage` | runner-visible evidence existed, but the assistant answer did not satisfy the path-selection scenario |
| `production` | `url-only-documentation-path-proposal` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `url-only-documentation-autonomous-placement` | `failed` | `data_hygiene_or_fixture_gap` | fixture or durable document evidence did not satisfy the path-selection contract |
| `production` | `multi-source-synthesis-path-selection` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `ambiguous-document-type-path-selection` | `failed` | `data_hygiene_or_fixture_gap` | fixture or durable document evidence did not satisfy the path-selection contract |
| `production` | `user-path-instructions-win` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `missing-document-path-reject` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `negative-limit-reject` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `unsupported-transport-reject` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
