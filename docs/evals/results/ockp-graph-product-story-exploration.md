# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `graph-product-story-exploration`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `17.89`
- Harness elapsed seconds: `125.33`
- Effective parallel speedup: `0.50x`
- Parallel efficiency: `0.50`
- Targeted acceptance: graph product story exploration rows compare existing primitives/baseline, graph_context_report, narrow read-only reports, approval-before-write maintenance plans, durable semantic graph/storage options, and no-new-surface across graph explanation, path finding, direct-vs-inferred reporting, typed candidates, stale/contradictory/orphaned audits, maintenance plans, and durable storage candidates, while recording concrete outcomes, safety pass, capability pass, UX quality, authority model, provenance/freshness posture, validation boundaries, workflow impact, and negative controls
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
| copy_repo | 0.37 |
| install_variant | 44.27 |
| warm_cache | 0.00 |
| seed_data | 0.03 |
| agent_run | 62.74 |
| parse_metrics | 0.00 |
| verify | 0.01 |
| total | 107.42 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `graph-product-story-exploration-control` | `completed` | 3 | 3 | 4 | 11555 | 34.66 | `<run-root>/production/graph-product-story-exploration-control/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2889 | 6.83 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2871 | 7.19 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2901 | 7.10 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2900 | 6.96 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `promote_graph_context_report_defer_adjacent_graph_stories`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: promote graph_context_report as the only current graph product story; defer adjacent read-only reports and approval-before-write maintenance-plan candidates to follow-up comparisons; kill durable semantic graph/storage candidates without auditability, rollback, provenance, freshness, and failure-mode evidence.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Workflow first command | Workflow calls | Setup discovery | Pre-action setup discovery | Pre-action primitives | Post-action primitives | Final-answer repair turns | Latency | Guidance dependence | Safety pass | Capability pass | UX quality | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| `production` | `graph-product-story-exploration-control` | `completed` | `none` | 3 | 3 | 4 | 34.66 | `candidate-surface-comparison` | `completed` | `normal` | 0 | 3 | 2 | 2 | 0 | 0 | 0 | 0 | 0 | `medium` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | graph product story control preserved graph_context_report as the promoted read-only baseline and assigned concrete outcomes to adjacent story surfaces |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 6.83 | `scenario-specific` | `completed` | `normal` | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 7.19 | `scenario-specific` | `completed` | `normal` | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 7.10 | `scenario-specific` | `completed` | `normal` | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 6.96 | `scenario-specific` | `completed` | `normal` | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
