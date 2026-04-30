# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `capture-save-this-note-candidate`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `15.25`
- Harness elapsed seconds: `129.40`
- Effective parallel speedup: `0.71x`
- Parallel efficiency: `0.71`
- Targeted acceptance: save-this-note capture rows report natural save intent, scripted candidate validation control, duplicate checks, low-confidence clarification, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification
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
| copy_repo | 0.25 |
| install_variant | 21.97 |
| warm_cache | 0.00 |
| seed_data | 0.01 |
| agent_run | 91.82 |
| parse_metrics | 0.00 |
| verify | 0.07 |
| total | 114.14 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `capture-save-this-note-natural-intent` | `failed` | 0 | 0 | 2 | 2863 | 20.31 | `<run-root>/production/capture-save-this-note-natural-intent/turn-1/events.jsonl` |
| `production` | `capture-save-this-note-scripted-control` | `completed` | 6 | 6 | 4 | 7506 | 16.43 | `<run-root>/production/capture-save-this-note-scripted-control/turn-1/events.jsonl` |
| `production` | `capture-save-this-note-duplicate-check` | `completed` | 10 | 10 | 5 | 8045 | 21.75 | `<run-root>/production/capture-save-this-note-duplicate-check/turn-1/events.jsonl` |
| `production` | `capture-save-this-note-low-confidence-ask` | `completed` | 0 | 0 | 1 | 2735 | 6.90 | `<run-root>/production/capture-save-this-note-low-confidence-ask/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2710 | 9.16 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2691 | 5.50 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2723 | 5.14 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2721 | 6.63 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `promote_save_this_note_capture_surface_design`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted evidence supports filing a separate implementation bead for the exact promoted save-this-note capture surface; no runner action, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `capture-save-this-note-natural-intent` | `failed` | `ergonomics_gap` | 0 | 0 | 2 | 20.31 | `natural-user-intent` | `answer_repair_needed` | `natural_prompt_sensitive` | 0 | 0 | `medium` | `high_if_natural_prompt_failed` | `none_observed` | `not_applicable` | natural save-this-note intent did not complete the safe current-primitives workflow |
| `production` | `capture-save-this-note-scripted-control` | `completed` | `none` | 6 | 6 | 4 | 16.43 | `scripted-control` | `completed` | `normal` | 0 | 6 | `medium` | `high_exact_request_shape` | `none_observed` | `not_applicable` | save-this-note capture preserved candidate faithfulness, duplicate checks, no-write boundary, approval-before-write, and bypass controls |
| `production` | `capture-save-this-note-duplicate-check` | `completed` | `none` | 10 | 10 | 5 | 21.75 | `scripted-control` | `completed` | `normal` | 0 | 10 | `medium` | `high_exact_request_shape` | `none_observed` | `not_applicable` | save-this-note capture preserved candidate faithfulness, duplicate checks, no-write boundary, approval-before-write, and bypass controls |
| `production` | `capture-save-this-note-low-confidence-ask` | `completed` | `none` | 0 | 0 | 1 | 6.90 | `natural-user-intent` | `completed` | `normal` | 0 | 0 | `low` | `low_natural_user_intent` | `none_observed` | `not_applicable` | save-this-note capture preserved candidate faithfulness, duplicate checks, no-write boundary, approval-before-write, and bypass controls |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 9.16 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 5.50 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 5.14 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 6.63 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
