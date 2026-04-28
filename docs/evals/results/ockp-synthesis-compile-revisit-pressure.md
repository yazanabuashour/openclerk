# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `synthesis-compile-revisit-pressure`
- Release blocking: `false`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `19.66`
- Harness elapsed seconds: `129.33`
- Effective parallel speedup: `1.27x`
- Parallel efficiency: `0.32`
- Targeted acceptance: synthesis compile revisit rows report natural compile intent, scripted current-primitives control, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification
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
| install_variant | 22.29 |
| warm_cache | 0.00 |
| seed_data | 0.13 |
| agent_run | 164.39 |
| parse_metrics | 0.00 |
| verify | 0.24 |
| total | 187.28 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `synthesis-compile-natural-intent` | `completed` | 34 | 34 | 12 | 17487 | 105.24 | `<run-root>/production/synthesis-compile-natural-intent/turn-1/events.jsonl` |
| `production` | `synthesis-compile-scripted-control` | `completed` | 18 | 18 | 5 | 13891 | 34.55 | `<run-root>/production/synthesis-compile-scripted-control/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 8494 | 4.89 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 21621 | 6.82 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2709 | 6.92 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 21651 | 5.97 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `defer_compile_synthesis`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted evidence only; no compile_synthesis runner action, schema, migration, storage behavior, direct vault behavior, or public API change from this eval.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `synthesis-compile-natural-intent` | `completed` | `none` | 34 | 34 | 12 | 105.24 | `natural-user-intent` | `completed` | `normal` | 0 | 34 | `high` | `low_natural_user_intent` | `none_observed` | `not_applicable` | current document/retrieval workflow preserved synthesis authority, source refs, provenance/freshness checks, duplicate prevention, and bypass boundaries |
| `production` | `synthesis-compile-scripted-control` | `completed` | `none` | 18 | 18 | 5 | 34.55 | `scripted-control` | `completed` | `normal` | 0 | 18 | `medium` | `high_exact_request_shape` | `none_observed` | `not_applicable` | current document/retrieval workflow preserved synthesis authority, source refs, provenance/freshness checks, duplicate prevention, and bypass boundaries |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 4.89 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 6.82 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 6.92 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 5.97 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
