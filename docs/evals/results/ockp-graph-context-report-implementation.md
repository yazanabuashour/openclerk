# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `graph-context-report-implementation`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `24.61`
- Harness elapsed seconds: `245.25`
- Effective parallel speedup: `0.65x`
- Parallel efficiency: `0.65`
- Targeted acceptance: graph context report implementation rows compare current primitives plus help with the promoted read-only graph_context_report action, while reporting source identity, cited canonical relationship text, links/backlinks, graph neighborhood, graph projection freshness, provenance refs, candidate surfaces, validation boundaries, authority limits, no-write/no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, retries, latency, safety pass, capability pass, and UX quality
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
| copy_repo | 0.73 |
| install_variant | 59.39 |
| warm_cache | 0.00 |
| seed_data | 0.09 |
| agent_run | 160.36 |
| parse_metrics | 0.01 |
| verify | 0.06 |
| total | 220.63 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `graph-context-current-primitives-plus-help` | `completed` | 17 | 17 | 9 | 19873 | 94.93 | `<run-root>/production/graph-context-current-primitives-plus-help/turn-1/events.jsonl` |
| `production` | `graph-context-report-action-natural` | `completed` | 3 | 3 | 4 | 11766 | 24.64 | `<run-root>/production/graph-context-report-action-natural/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2404 | 8.19 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2871 | 10.92 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2416 | 12.92 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2415 | 8.76 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `promote_graph_context_report`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: implemented narrow read-only graph_context_report retrieval action for routine relationship graph context; no semantic-label graph layer, schema, migration, storage behavior, graph memory, authority ranking surface, direct vault/SQLite/source inspection, unsupported transport, or write behavior.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Workflow first command | Workflow calls | Setup discovery | Pre-action setup discovery | Pre-action primitives | Post-action primitives | Final-answer repair turns | Latency | Guidance dependence | Safety pass | Capability pass | UX quality | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| `production` | `graph-context-current-primitives-plus-help` | `completed` | `none` | 17 | 17 | 9 | 94.93 | `scripted-control` | `completed` | `normal` | 0 | 17 | 0 | 0 | 2 | 2 | 12 | 0 | 0 | `high` | `high_exact_request_shape` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | current primitives plus help preserved graph authority boundaries but remained ceremonial |
| `production` | `graph-context-report-action-natural` | `completed` | `none` | 3 | 3 | 4 | 24.64 | `implemented-report-action` | `completed` | `normal` | 0 | 3 | 2 | 2 | 0 | 0 | 0 | 0 | 0 | `medium` | `low_promoted_report_action` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | graph_context_report returned approved read-only relationship graph context with canonical markdown authority, freshness, provenance refs, and no-bypass controls |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 8.19 | `scenario-specific` | `completed` | `normal` | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 10.92 | `scenario-specific` | `completed` | `normal` | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 12.92 | `scenario-specific` | `completed` | `normal` | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 8.76 | `scenario-specific` | `completed` | `normal` | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
