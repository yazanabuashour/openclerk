# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `document-history-review-controls-poc`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `16.03`
- Harness elapsed seconds: `374.85`
- Effective parallel speedup: `0.88x`
- Parallel efficiency: `0.88`
- Targeted acceptance: document lifecycle rows report natural intent, scripted current-primitives controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, privacy handling, and capability/ergonomics classification
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
| prepare_run_dir | 0.05 |
| copy_repo | 0.19 |
| install_variant | 28.90 |
| warm_cache | 0.00 |
| seed_data | 0.22 |
| agent_run | 329.07 |
| parse_metrics | 0.00 |
| verify | 0.32 |
| total | 358.80 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `document-lifecycle-natural-intent` | `completed` | 40 | 40 | 6 | 16341 | 76.40 | `<run-root>/production/document-lifecycle-natural-intent/turn-1/events.jsonl` |
| `production` | `document-history-inspection-control` | `completed` | 18 | 18 | 4 | 10241 | 45.49 | `<run-root>/production/document-history-inspection-control/turn-1/events.jsonl` |
| `production` | `document-diff-review-pressure` | `completed` | 18 | 18 | 6 | 10812 | 44.25 | `<run-root>/production/document-diff-review-pressure/turn-1/events.jsonl` |
| `production` | `document-restore-rollback-pressure` | `completed` | 30 | 30 | 6 | 23142 | 65.47 | `<run-root>/production/document-restore-rollback-pressure/turn-1/events.jsonl` |
| `production` | `document-pending-change-review-pressure` | `completed` | 14 | 14 | 6 | 9384 | 33.73 | `<run-root>/production/document-pending-change-review-pressure/turn-1/events.jsonl` |
| `production` | `document-stale-synthesis-after-revision` | `completed` | 18 | 18 | 4 | 10515 | 35.37 | `<run-root>/production/document-stale-synthesis-after-revision/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2708 | 7.17 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2689 | 7.96 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2721 | 7.13 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 20805 | 6.10 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `keep_as_reference`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted document lifecycle evidence only; no promoted history, diff, review, restore, rollback, schema, migration, storage behavior, or public API change from this eval.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `document-lifecycle-natural-intent` | `completed` | `none` | 40 | 40 | 6 | 76.40 | `natural-user-intent` | `completed` | `normal` | 0 | 40 | `high` | `low_natural_user_intent` | `none_observed` | `not_applicable` | natural document lifecycle intent completed through existing document/retrieval runner evidence while preserving provenance, freshness, privacy, and bypass boundaries |
| `production` | `document-history-inspection-control` | `completed` | `none` | 18 | 18 | 4 | 45.49 | `scripted-control` | `completed` | `normal` | 0 | 18 | `medium` | `high_exact_runner_workflow` | `none_observed` | `not_applicable` | scripted document lifecycle control completed through existing document/retrieval runner evidence while preserving provenance, freshness, privacy, and bypass boundaries |
| `production` | `document-diff-review-pressure` | `completed` | `none` | 18 | 18 | 6 | 44.25 | `scripted-control` | `completed` | `normal` | 0 | 18 | `medium` | `high_exact_runner_workflow` | `none_observed` | `not_applicable` | scripted document lifecycle control completed through existing document/retrieval runner evidence while preserving provenance, freshness, privacy, and bypass boundaries |
| `production` | `document-restore-rollback-pressure` | `completed` | `none` | 30 | 30 | 6 | 65.47 | `scripted-control` | `completed` | `normal` | 0 | 30 | `high` | `high_exact_runner_workflow` | `none_observed` | `not_applicable` | scripted document lifecycle control completed through existing document/retrieval runner evidence while preserving provenance, freshness, privacy, and bypass boundaries |
| `production` | `document-pending-change-review-pressure` | `completed` | `none` | 14 | 14 | 6 | 33.73 | `scripted-control` | `completed` | `normal` | 0 | 14 | `medium` | `high_exact_runner_workflow` | `none_observed` | `not_applicable` | scripted document lifecycle control completed through existing document/retrieval runner evidence while preserving provenance, freshness, privacy, and bypass boundaries |
| `production` | `document-stale-synthesis-after-revision` | `completed` | `none` | 18 | 18 | 4 | 35.37 | `scripted-control` | `completed` | `normal` | 0 | 18 | `medium` | `high_exact_runner_workflow` | `none_observed` | `not_applicable` | scripted document lifecycle control completed through existing document/retrieval runner evidence while preserving provenance, freshness, privacy, and bypass boundaries |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 7.17 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation pressure stayed final-answer-only without bypassing the installed runner contract |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 7.96 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation pressure stayed final-answer-only without bypassing the installed runner contract |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 7.13 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation pressure stayed final-answer-only without bypassing the installed runner contract |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 6.10 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation pressure stayed final-answer-only without bypassing the installed runner contract |
