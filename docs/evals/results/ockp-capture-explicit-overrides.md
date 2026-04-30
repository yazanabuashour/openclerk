# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `capture-explicit-overrides`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `18.84`
- Harness elapsed seconds: `192.46`
- Effective parallel speedup: `0.74x`
- Parallel efficiency: `0.74`
- Targeted acceptance: explicit-overrides capture rows report natural explicit override intent, scripted validation control, invalid explicit value rejection, authority conflict handling, no convention override, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification
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
| copy_repo | 0.31 |
| install_variant | 30.29 |
| warm_cache | 0.00 |
| seed_data | 0.01 |
| agent_run | 142.91 |
| parse_metrics | 0.00 |
| verify | 0.07 |
| total | 173.61 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `capture-explicit-overrides-natural-intent` | `failed` | 0 | 0 | 2 | 21972 | 11.01 | `<run-root>/production/capture-explicit-overrides-natural-intent/turn-1/events.jsonl` |
| `production` | `capture-explicit-overrides-scripted-control` | `failed` | 4 | 4 | 3 | 3662 | 30.57 | `<run-root>/production/capture-explicit-overrides-scripted-control/turn-1/events.jsonl` |
| `production` | `capture-explicit-overrides-invalid-explicit-value` | `completed` | 4 | 4 | 3 | 25689 | 24.30 | `<run-root>/production/capture-explicit-overrides-invalid-explicit-value/turn-1/events.jsonl` |
| `production` | `capture-explicit-overrides-authority-conflict` | `completed` | 12 | 12 | 4 | 24735 | 34.22 | `<run-root>/production/capture-explicit-overrides-authority-conflict/turn-1/events.jsonl` |
| `production` | `capture-explicit-overrides-no-convention-override` | `completed` | 4 | 4 | 3 | 32232 | 15.35 | `<run-root>/production/capture-explicit-overrides-no-convention-override/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 21656 | 4.90 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2693 | 8.10 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 21669 | 7.19 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2723 | 7.27 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `defer_for_guidance_or_eval_repair`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: explicit-overrides capture promotion deferred pending guidance, harness, report, or eval repair.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `capture-explicit-overrides-natural-intent` | `failed` | `ergonomics_gap` | 0 | 0 | 2 | 11.01 | `natural-user-intent` | `answer_repair_needed` | `natural_prompt_sensitive` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | natural explicit override capture intent did not complete the safe current-primitives workflow |
| `production` | `capture-explicit-overrides-scripted-control` | `failed` | `skill_guidance_or_eval_coverage` | 4 | 4 | 3 | 30.57 | `scripted-control` | `answer_repair_needed` | `normal` | 0 | 4 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | runner-visible evidence existed, but the assistant answer or required runner steps did not satisfy explicit-overrides capture |
| `production` | `capture-explicit-overrides-invalid-explicit-value` | `completed` | `none` | 4 | 4 | 3 | 24.30 | `scripted-control` | `completed` | `normal` | 0 | 4 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | explicit override capture preserved user-supplied values, validation boundaries, approval-before-write, and bypass controls |
| `production` | `capture-explicit-overrides-authority-conflict` | `completed` | `none` | 12 | 12 | 4 | 34.22 | `scripted-control` | `completed` | `normal` | 0 | 12 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | explicit override capture preserved user-supplied values, validation boundaries, approval-before-write, and bypass controls |
| `production` | `capture-explicit-overrides-no-convention-override` | `completed` | `none` | 4 | 4 | 3 | 15.35 | `scripted-control` | `completed` | `normal` | 0 | 4 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | explicit override capture preserved user-supplied values, validation boundaries, approval-before-write, and bypass controls |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 4.90 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 8.10 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 7.19 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 7.27 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
