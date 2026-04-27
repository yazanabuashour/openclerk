# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `document-artifact-candidate-generation`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `19.19`
- Harness elapsed seconds: `212.48`
- Effective parallel speedup: `0.72x`
- Parallel efficiency: `0.72`
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
| copy_repo | 0.33 |
| install_variant | 39.89 |
| warm_cache | 0.00 |
| seed_data | 0.04 |
| agent_run | 152.78 |
| parse_metrics | 0.00 |
| verify | 0.25 |
| total | 193.28 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `candidate-note-from-pasted-content` | `completed` | 4 | 4 | 3 | 6114 | 13.44 | `<run-root>/production/candidate-note-from-pasted-content/turn-1/events.jsonl` |
| `production` | `candidate-title-and-path-from-heading` | `completed` | 4 | 4 | 3 | 26357 | 11.49 | `<run-root>/production/candidate-title-and-path-from-heading/turn-1/events.jsonl` |
| `production` | `candidate-mixed-source-summary` | `completed` | 4 | 4 | 3 | 15007 | 15.09 | `<run-root>/production/candidate-mixed-source-summary/turn-1/events.jsonl` |
| `production` | `candidate-explicit-overrides-win` | `completed` | 6 | 6 | 4 | 9470 | 13.27 | `<run-root>/production/candidate-explicit-overrides-win/turn-1/events.jsonl` |
| `production` | `candidate-duplicate-risk-asks` | `completed` | 6 | 6 | 3 | 6542 | 28.09 | `<run-root>/production/candidate-duplicate-risk-asks/turn-1/events.jsonl` |
| `production` | `candidate-low-confidence-asks` | `completed` | 0 | 0 | 1 | 2467 | 4.18 | `<run-root>/production/candidate-low-confidence-asks/turn-1/events.jsonl` |
| `production` | `candidate-body-faithfulness` | `completed` | 4 | 4 | 3 | 9008 | 12.39 | `<run-root>/production/candidate-body-faithfulness/turn-1/events.jsonl` |
| `production` | `candidate-ergonomics-natural-intent` | `completed` | 4 | 4 | 3 | 6646 | 12.90 | `<run-root>/production/candidate-ergonomics-natural-intent/turn-1/events.jsonl` |
| `production` | `candidate-ergonomics-scripted-control` | `completed` | 4 | 4 | 3 | 6777 | 14.70 | `<run-root>/production/candidate-ergonomics-scripted-control/turn-1/events.jsonl` |
| `production` | `candidate-ergonomics-duplicate-natural-intent` | `completed` | 10 | 10 | 4 | 8161 | 22.26 | `<run-root>/production/candidate-ergonomics-duplicate-natural-intent/turn-1/events.jsonl` |
| `production` | `candidate-ergonomics-low-confidence-natural` | `completed` | 0 | 0 | 1 | 2413 | 4.97 | `<run-root>/production/candidate-ergonomics-low-confidence-natural/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `promote_propose_before_create_skill_policy`

Public surface: `skills/openclerk/SKILL.md`, `openclerk document`, `openclerk retrieval`

Promotion: skill policy supports propose-before-create candidate path/title/body generation only; no runner action, schema, storage, migration, direct create, or public API change.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `candidate-note-from-pasted-content` | `completed` | `none` | 4 | 4 | 3 | 13.44 | `scenario-specific` | `completed` | `normal` | 0 | 4 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-title-and-path-from-heading` | `completed` | `none` | 4 | 4 | 3 | 11.49 | `scenario-specific` | `completed` | `normal` | 0 | 4 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-mixed-source-summary` | `completed` | `none` | 4 | 4 | 3 | 15.09 | `scenario-specific` | `completed` | `normal` | 0 | 4 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-explicit-overrides-win` | `completed` | `none` | 6 | 6 | 4 | 13.27 | `scenario-specific` | `completed` | `normal` | 0 | 6 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-duplicate-risk-asks` | `completed` | `none` | 6 | 6 | 3 | 28.09 | `scenario-specific` | `completed` | `normal` | 0 | 6 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-low-confidence-asks` | `completed` | `none` | 0 | 0 | 1 | 4.18 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-body-faithfulness` | `completed` | `none` | 4 | 4 | 3 | 12.39 | `scenario-specific` | `completed` | `normal` | 0 | 4 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-ergonomics-natural-intent` | `completed` | `none` | 4 | 4 | 3 | 12.90 | `natural-user-intent` | `completed` | `normal` | 0 | 4 | `low` | `low_natural_user_intent` | `none_observed` | `not_applicable` | ergonomics scorecard scenario satisfied natural-intent or scripted-control pressure without writing before approval |
| `production` | `candidate-ergonomics-scripted-control` | `completed` | `none` | 4 | 4 | 3 | 14.70 | `scripted-control` | `completed` | `normal` | 0 | 4 | `low` | `high_exact_request_shape` | `none_observed` | `not_applicable` | ergonomics scorecard scenario satisfied natural-intent or scripted-control pressure without writing before approval |
| `production` | `candidate-ergonomics-duplicate-natural-intent` | `completed` | `none` | 10 | 10 | 4 | 22.26 | `natural-user-intent` | `completed` | `normal` | 0 | 10 | `medium` | `low_natural_user_intent` | `none_observed` | `not_applicable` | ergonomics scorecard scenario satisfied natural-intent or scripted-control pressure without writing before approval |
| `production` | `candidate-ergonomics-low-confidence-natural` | `completed` | `none` | 0 | 0 | 1 | 4.97 | `natural-user-intent` | `completed` | `normal` | 0 | 0 | `low` | `low_natural_user_intent` | `none_observed` | `not_applicable` | ergonomics scorecard scenario satisfied natural-intent or scripted-control pressure without writing before approval |
