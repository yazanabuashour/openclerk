# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `synthesis-compile-revisit-pressure`
- Release blocking: `false`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `22.68`
- Harness elapsed seconds: `92.76`
- Effective parallel speedup: `1.64x`
- Parallel efficiency: `0.41`
- Targeted acceptance: synthesis compile revisit rows report natural compile intent, scripted current-primitives control, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 3/30 production scenarios passed; missing: create-note, search-synthesis, answer-filing, rag-retrieval-baseline, canonical-docs-navigation-baseline, graph-semantics-reference-poc, memory-router-reference-poc, configured-layout-explain, invalid-layout-visible, stale-synthesis-update, synthesis-freshness-repair, source-sensitive-audit-repair, source-sensitive-conflict-explain, synthesis-candidate-pressure, synthesis-source-set-pressure, append-replace, records-provenance, promoted-record-vs-docs, decision-record-vs-docs, decision-supersession-freshness, decision-real-adr-migration, duplicate-path-reject, mixed-synthesis-records, mt-source-then-synthesis, mt-synthesis-drift-pressure, mt-incomplete-then-create |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect retired API files or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `pass` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_legacy_source_runner_usage` | `pass` | production must not invoke source-built or legacy runner paths instead of installed openclerk |
| `no_direct_sqlite_access` | `pass` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `fail` | not final-answer-only: negative-limit-reject |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.36 |
| install_variant | 23.94 |
| warm_cache | 0.00 |
| seed_data | 0.11 |
| agent_run | 152.52 |
| parse_metrics | 0.00 |
| verify | 0.29 |
| total | 177.23 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `synthesis-compile-natural-intent` | `completed` | 34 | 34 | 8 | 13640 | 65.51 | `<run-root>/production/synthesis-compile-natural-intent/turn-1/events.jsonl` |
| `production` | `synthesis-compile-scripted-control` | `completed` | 28 | 28 | 9 | 18840 | 57.12 | `<run-root>/production/synthesis-compile-scripted-control/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 16420 | 7.79 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `failed` | 2 | 2 | 2 | 45451 | 9.83 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2608 | 7.31 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2480 | 6.92 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `defer_for_guidance_or_eval_repair`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted evidence only; no compile_synthesis runner action, schema, migration, storage behavior, direct vault behavior, or public API change from this eval.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `synthesis-compile-natural-intent` | `completed` | `none` | 34 | 34 | 8 | 65.51 | `natural-user-intent` | `completed` | `normal` | 0 | 34 | `high` | `low_natural_user_intent` | `none_observed` | `not_applicable` | current document/retrieval workflow preserved synthesis authority, source refs, provenance/freshness checks, duplicate prevention, and bypass boundaries |
| `production` | `synthesis-compile-scripted-control` | `completed` | `none` | 28 | 28 | 9 | 57.12 | `scripted-control` | `completed` | `normal` | 0 | 28 | `medium` | `high_exact_request_shape` | `none_observed` | `not_applicable` | current document/retrieval workflow preserved synthesis authority, source refs, provenance/freshness checks, duplicate prevention, and bypass boundaries |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 7.79 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `failed` | `skill_guidance_or_eval_coverage` | 2 | 2 | 2 | 9.83 | `scenario-specific` | `manual_review` | `normal` | 0 | 2 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation pressure did not stay final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 7.31 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 6.92 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
