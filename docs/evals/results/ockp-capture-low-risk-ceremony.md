# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `capture-low-risk-ceremony`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `20.79`
- Harness elapsed seconds: `140.34`
- Effective parallel speedup: `0.65x`
- Parallel efficiency: `0.65`
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
| copy_repo | 0.26 |
| install_variant | 28.43 |
| warm_cache | 0.00 |
| seed_data | 0.01 |
| agent_run | 90.78 |
| parse_metrics | 0.00 |
| verify | 0.03 |
| total | 119.56 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `capture-low-risk-natural-intent` | `failed` | 6 | 6 | 3 | 48906 | 23.55 | `<run-root>/production/capture-low-risk-natural-intent/turn-1/events.jsonl` |
| `production` | `capture-low-risk-scripted-control` | `completed` | 4 | 4 | 3 | 6591 | 13.09 | `<run-root>/production/capture-low-risk-scripted-control/turn-1/events.jsonl` |
| `production` | `capture-low-risk-duplicate-check` | `completed` | 8 | 8 | 4 | 26610 | 26.42 | `<run-root>/production/capture-low-risk-duplicate-check/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2710 | 5.13 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2691 | 6.78 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2723 | 5.23 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2721 | 10.58 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `promote_low_risk_capture_surface_design`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted evidence supports filing a separate implementation bead for the exact promoted low-risk capture surface; no runner action, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `capture-low-risk-natural-intent` | `failed` | `ergonomics_gap` | 6 | 6 | 3 | 23.55 | `natural-user-intent` | `answer_repair_needed` | `natural_prompt_sensitive` | 0 | 6 | `medium` | `high_if_natural_prompt_failed` | `none_observed` | `not_applicable` | natural low-risk capture intent did not complete the safe current-primitives workflow |
| `production` | `capture-low-risk-scripted-control` | `completed` | `none` | 4 | 4 | 3 | 13.09 | `scripted-control` | `completed` | `normal` | 0 | 4 | `low` | `high_exact_request_shape` | `none_observed` | `not_applicable` | low-risk capture preserved candidate faithfulness, duplicate checks, no-write boundary, approval-before-write, and bypass controls |
| `production` | `capture-low-risk-duplicate-check` | `completed` | `none` | 8 | 8 | 4 | 26.42 | `scripted-control` | `completed` | `normal` | 0 | 8 | `medium` | `high_exact_request_shape` | `none_observed` | `not_applicable` | low-risk capture preserved candidate faithfulness, duplicate checks, no-write boundary, approval-before-write, and bypass controls |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 5.13 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 6.78 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 5.23 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 10.58 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
