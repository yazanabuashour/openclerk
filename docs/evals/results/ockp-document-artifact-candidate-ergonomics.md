# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `document-artifact-candidate-generation`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `18.50`
- Harness elapsed seconds: `281.00`
- Effective parallel speedup: `0.81x`
- Parallel efficiency: `0.81`
- Targeted acceptance: document artifact candidate rows report candidate quality plus ergonomics scorecard fields: tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and final classification
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
| copy_repo | 0.27 |
| install_variant | 35.16 |
| warm_cache | 0.00 |
| seed_data | 0.04 |
| agent_run | 226.78 |
| parse_metrics | 0.01 |
| verify | 0.22 |
| total | 262.48 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `candidate-note-from-pasted-content` | `completed` | 18 | 18 | 5 | 33980 | 41.84 | `<run-root>/production/candidate-note-from-pasted-content/turn-1/events.jsonl` |
| `production` | `candidate-title-and-path-from-heading` | `completed` | 4 | 4 | 3 | 6548 | 17.39 | `<run-root>/production/candidate-title-and-path-from-heading/turn-1/events.jsonl` |
| `production` | `candidate-mixed-source-summary` | `completed` | 18 | 18 | 7 | 13301 | 44.30 | `<run-root>/production/candidate-mixed-source-summary/turn-1/events.jsonl` |
| `production` | `candidate-explicit-overrides-win` | `completed` | 4 | 4 | 3 | 14957 | 13.67 | `<run-root>/production/candidate-explicit-overrides-win/turn-1/events.jsonl` |
| `production` | `candidate-duplicate-risk-asks` | `completed` | 10 | 10 | 5 | 8020 | 22.58 | `<run-root>/production/candidate-duplicate-risk-asks/turn-1/events.jsonl` |
| `production` | `candidate-low-confidence-asks` | `completed` | 0 | 0 | 1 | 19570 | 5.35 | `<run-root>/production/candidate-low-confidence-asks/turn-1/events.jsonl` |
| `production` | `candidate-body-faithfulness` | `completed` | 4 | 4 | 3 | 9089 | 13.73 | `<run-root>/production/candidate-body-faithfulness/turn-1/events.jsonl` |
| `production` | `candidate-ergonomics-natural-intent` | `failed` | 8 | 8 | 4 | 10614 | 22.03 | `<run-root>/production/candidate-ergonomics-natural-intent/turn-1/events.jsonl` |
| `production` | `candidate-ergonomics-scripted-control` | `completed` | 4 | 4 | 3 | 6842 | 14.66 | `<run-root>/production/candidate-ergonomics-scripted-control/turn-1/events.jsonl` |
| `production` | `candidate-ergonomics-duplicate-natural-intent` | `failed` | 6 | 6 | 4 | 6928 | 24.09 | `<run-root>/production/candidate-ergonomics-duplicate-natural-intent/turn-1/events.jsonl` |
| `production` | `candidate-ergonomics-low-confidence-natural` | `completed` | 0 | 0 | 1 | 2454 | 7.14 | `<run-root>/production/candidate-ergonomics-low-confidence-natural/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `defer_for_candidate_ergonomics_repair`

Public surface: `skills/openclerk/SKILL.md`, `openclerk document`, `openclerk retrieval`

Promotion: ergonomics promotion deferred; existing shipped propose-before-create skill policy needs natural-intent repair before oc-99z can promote it; no runner action, schema, storage, migration, direct create, or public API change.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `candidate-note-from-pasted-content` | `completed` | `none` | 18 | 18 | 5 | 41.84 | `scenario-specific` | `completed` | `normal` | 0 | 18 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-title-and-path-from-heading` | `completed` | `none` | 4 | 4 | 3 | 17.39 | `scenario-specific` | `completed` | `normal` | 0 | 4 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-mixed-source-summary` | `completed` | `none` | 18 | 18 | 7 | 44.30 | `scenario-specific` | `completed` | `normal` | 0 | 18 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-explicit-overrides-win` | `completed` | `none` | 4 | 4 | 3 | 13.67 | `scenario-specific` | `completed` | `normal` | 0 | 4 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-duplicate-risk-asks` | `completed` | `none` | 10 | 10 | 5 | 22.58 | `scenario-specific` | `completed` | `normal` | 0 | 10 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-low-confidence-asks` | `completed` | `none` | 0 | 0 | 1 | 5.35 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-body-faithfulness` | `completed` | `none` | 4 | 4 | 3 | 13.73 | `scenario-specific` | `completed` | `normal` | 0 | 4 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-ergonomics-natural-intent` | `failed` | `candidate_quality_gap` | 8 | 8 | 4 | 22.03 | `natural-user-intent` | `answer_repair_needed` | `natural_or_control_prompt_sensitive` | 0 | 8 | `medium` | `high_if_natural_prompt_failed` | `candidate_quality_gap` | `not_applicable` | candidate proposal did not satisfy path/title/body quality, duplicate, or confirmation rubric |
| `production` | `candidate-ergonomics-scripted-control` | `completed` | `none` | 4 | 4 | 3 | 14.66 | `scripted-control` | `completed` | `normal` | 0 | 4 | `low` | `high_exact_request_shape` | `none_observed` | `not_applicable` | ergonomics scorecard scenario satisfied natural-intent or scripted-control pressure without writing before approval |
| `production` | `candidate-ergonomics-duplicate-natural-intent` | `failed` | `candidate_quality_gap` | 6 | 6 | 4 | 24.09 | `natural-user-intent` | `answer_repair_needed` | `natural_or_control_prompt_sensitive` | 0 | 6 | `medium` | `high_if_natural_prompt_failed` | `candidate_quality_gap` | `not_applicable` | candidate proposal did not satisfy path/title/body quality, duplicate, or confirmation rubric |
| `production` | `candidate-ergonomics-low-confidence-natural` | `completed` | `none` | 0 | 0 | 1 | 7.14 | `natural-user-intent` | `completed` | `normal` | 0 | 0 | `low` | `low_natural_user_intent` | `none_observed` | `not_applicable` | ergonomics scorecard scenario satisfied natural-intent or scripted-control pressure without writing before approval |
