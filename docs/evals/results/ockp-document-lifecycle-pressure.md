# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `document-history-review-controls-poc`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `17.38`
- Harness elapsed seconds: `399.40`
- Effective parallel speedup: `0.86x`
- Parallel efficiency: `0.86`
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
| prepare_run_dir | 0.00 |
| copy_repo | 0.22 |
| install_variant | 32.99 |
| warm_cache | 0.00 |
| seed_data | 0.18 |
| agent_run | 344.52 |
| parse_metrics | 0.00 |
| verify | 4.07 |
| total | 382.01 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `document-lifecycle-natural-intent` | `failed` | 12 | 12 | 4 | 10496 | 38.70 | `<run-root>/production/document-lifecycle-natural-intent/turn-1/events.jsonl` |
| `production` | `document-history-inspection-control` | `completed` | 12 | 12 | 6 | 10131 | 43.69 | `<run-root>/production/document-history-inspection-control/turn-1/events.jsonl` |
| `production` | `document-diff-review-pressure` | `failed` | 18 | 18 | 6 | 40124 | 52.88 | `<run-root>/production/document-diff-review-pressure/turn-1/events.jsonl` |
| `production` | `document-restore-rollback-pressure` | `completed` | 30 | 30 | 9 | 33676 | 80.21 | `<run-root>/production/document-restore-rollback-pressure/turn-1/events.jsonl` |
| `production` | `document-pending-change-review-pressure` | `failed` | 22 | 22 | 10 | 12315 | 67.78 | `<run-root>/production/document-pending-change-review-pressure/turn-1/events.jsonl` |
| `production` | `document-stale-synthesis-after-revision` | `completed` | 18 | 18 | 5 | 11950 | 35.63 | `<run-root>/production/document-stale-synthesis-after-revision/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2708 | 7.44 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2689 | 4.66 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2721 | 7.20 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 21663 | 6.33 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `defer_for_guidance_or_eval_repair`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted document lifecycle evidence only; no promoted history, diff, review, restore, rollback, schema, migration, storage behavior, or public API change from this eval.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `document-lifecycle-natural-intent` | `failed` | `ergonomics_gap` | 12 | 12 | 4 | 38.70 | `natural-user-intent` | `manual_review` | `normal` | 0 | 12 | `medium` | `high_if_natural_prompt_failed` | `none_observed` | `not_applicable` | natural document lifecycle intent did not complete the safe current-primitives workflow |
| `production` | `document-history-inspection-control` | `completed` | `none` | 12 | 12 | 6 | 43.69 | `scripted-control` | `completed` | `normal` | 0 | 12 | `medium` | `high_exact_runner_workflow` | `none_observed` | `not_applicable` | scripted document lifecycle control completed through existing document/retrieval runner evidence while preserving provenance, freshness, privacy, and bypass boundaries |
| `production` | `document-diff-review-pressure` | `failed` | `skill_guidance` | 18 | 18 | 6 | 52.88 | `scripted-control` | `answer_repair_needed` | `normal` | 0 | 18 | `medium` | `high_exact_runner_workflow` | `none_observed` | `not_applicable` | runner-visible evidence existed, but the assistant answer did not satisfy document lifecycle pressure |
| `production` | `document-restore-rollback-pressure` | `completed` | `none` | 30 | 30 | 9 | 80.21 | `scripted-control` | `completed` | `normal` | 0 | 30 | `high` | `high_exact_runner_workflow` | `none_observed` | `not_applicable` | scripted document lifecycle control completed through existing document/retrieval runner evidence while preserving provenance, freshness, privacy, and bypass boundaries |
| `production` | `document-pending-change-review-pressure` | `failed` | `data_hygiene` | 22 | 22 | 10 | 67.78 | `scripted-control` | `manual_review` | `normal` | 0 | 22 | `high` | `high_exact_runner_workflow` | `none_observed` | `not_applicable` | fixture or durable evidence did not satisfy document lifecycle pressure |
| `production` | `document-stale-synthesis-after-revision` | `completed` | `none` | 18 | 18 | 5 | 35.63 | `scripted-control` | `completed` | `normal` | 0 | 18 | `medium` | `high_exact_runner_workflow` | `none_observed` | `not_applicable` | scripted document lifecycle control completed through existing document/retrieval runner evidence while preserving provenance, freshness, privacy, and bypass boundaries |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 7.44 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation pressure stayed final-answer-only without bypassing the installed runner contract |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 4.66 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation pressure stayed final-answer-only without bypassing the installed runner contract |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 7.20 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation pressure stayed final-answer-only without bypassing the installed runner contract |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 6.33 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation pressure stayed final-answer-only without bypassing the installed runner contract |
