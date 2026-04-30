# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `capture-low-risk-ceremony`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `20.13`
- Harness elapsed seconds: `122.07`
- Effective parallel speedup: `0.64x`
- Parallel efficiency: `0.64`
- Targeted acceptance: low-risk capture rows report natural low-risk save intent, scripted candidate validation control, duplicate checks, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification
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
| copy_repo | 0.20 |
| install_variant | 23.60 |
| warm_cache | 0.00 |
| seed_data | 0.01 |
| agent_run | 78.06 |
| parse_metrics | 0.00 |
| verify | 0.04 |
| total | 101.94 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `capture-low-risk-natural-intent` | `completed` | 6 | 6 | 4 | 6804 | 15.08 | `<run-root>/production/capture-low-risk-natural-intent/turn-1/events.jsonl` |
| `production` | `capture-low-risk-scripted-control` | `completed` | 8 | 8 | 4 | 9120 | 21.29 | `<run-root>/production/capture-low-risk-scripted-control/turn-1/events.jsonl` |
| `production` | `capture-low-risk-duplicate-check` | `completed` | 10 | 10 | 5 | 7977 | 20.29 | `<run-root>/production/capture-low-risk-duplicate-check/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2874 | 4.40 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2689 | 6.29 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2721 | 4.92 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2719 | 5.79 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `keep_as_reference`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: keep low-risk capture as reference evidence for product implementation; focused skill-policy guidance hardening was applied with no implementation bead, runner action, schema, storage, public API, direct-create, hidden-autofiling, or product behavior change.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `capture-low-risk-natural-intent` | `completed` | `none` | 6 | 6 | 4 | 15.08 | `natural-user-intent` | `completed` | `normal` | 0 | 6 | `medium` | `low_natural_user_intent` | `none_observed` | `not_applicable` | low-risk capture preserved candidate faithfulness, duplicate checks, no-write boundary, approval-before-write, and bypass controls |
| `production` | `capture-low-risk-scripted-control` | `completed` | `none` | 8 | 8 | 4 | 21.29 | `scripted-control` | `completed` | `normal` | 0 | 8 | `medium` | `high_exact_request_shape` | `none_observed` | `not_applicable` | low-risk capture preserved candidate faithfulness, duplicate checks, no-write boundary, approval-before-write, and bypass controls |
| `production` | `capture-low-risk-duplicate-check` | `completed` | `none` | 10 | 10 | 5 | 20.29 | `scripted-control` | `completed` | `normal` | 0 | 10 | `medium` | `high_exact_request_shape` | `none_observed` | `not_applicable` | low-risk capture preserved candidate faithfulness, duplicate checks, no-write boundary, approval-before-write, and bypass controls |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 4.40 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 6.29 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 4.92 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 5.79 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
