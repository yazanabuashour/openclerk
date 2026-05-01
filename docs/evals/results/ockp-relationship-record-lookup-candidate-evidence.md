# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `relationship-record-lookup-candidate-evidence`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `18.09`
- Harness elapsed seconds: `230.72`
- Effective parallel speedup: `0.84x`
- Parallel efficiency: `0.84`
- Targeted acceptance: relationship-record lookup candidate rows compare current primitives, guidance-only repair, and an eval-only candidate response contract, while reporting query summary, relationship evidence, link/backlink evidence, graph freshness, record lookup/entity evidence, citation refs, provenance refs, records freshness, validation/no-bypass boundaries, authority limits, tool count, command count, assistant calls, wall time, prompt specificity, retries, latency, brittleness, guidance dependence, safety risks, safety pass, capability pass, and UX quality
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
| copy_repo | 0.33 |
| install_variant | 17.65 |
| warm_cache | 0.00 |
| seed_data | 0.12 |
| agent_run | 194.39 |
| parse_metrics | 0.02 |
| verify | 0.13 |
| total | 212.63 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `relationship-record-current-primitives-control` | `completed` | 26 | 26 | 5 | 17053 | 43.78 | `<run-root>/production/relationship-record-current-primitives-control/turn-1/events.jsonl` |
| `production` | `relationship-record-guidance-only-natural` | `completed` | 56 | 56 | 8 | 39488 | 68.10 | `<run-root>/production/relationship-record-guidance-only-natural/turn-1/events.jsonl` |
| `production` | `relationship-record-response-candidate` | `completed` | 30 | 30 | 7 | 21374 | 55.99 | `<run-root>/production/relationship-record-response-candidate/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2724 | 5.49 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2706 | 6.26 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2736 | 8.97 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2735 | 5.80 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `defer_guidance_only_current_primitives_sufficient`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: guidance-only current primitives satisfied this targeted pressure, so the relationship-record lookup candidate is deferred pending stronger repeated ergonomics or answer-contract evidence.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety pass | Capability pass | UX quality | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| `production` | `relationship-record-current-primitives-control` | `completed` | `none` | 26 | 26 | 5 | 43.78 | `scripted-control` | `completed` | `normal` | 0 | 26 | `medium` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | relationship-record candidate evidence preserved canonical relationship authority, links/backlinks, graph freshness, canonical record authority, citations, provenance, records freshness, eval-only response boundaries, and no-bypass controls |
| `production` | `relationship-record-guidance-only-natural` | `completed` | `none` | 56 | 56 | 8 | 68.10 | `natural-user-intent` | `completed` | `normal` | 0 | 56 | `high` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | relationship-record candidate evidence preserved canonical relationship authority, links/backlinks, graph freshness, canonical record authority, citations, provenance, records freshness, eval-only response boundaries, and no-bypass controls |
| `production` | `relationship-record-response-candidate` | `completed` | `none` | 30 | 30 | 7 | 55.99 | `candidate-response-contract` | `completed` | `normal` | 0 | 30 | `medium` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | relationship-record candidate evidence preserved canonical relationship authority, links/backlinks, graph freshness, canonical record authority, citations, provenance, records freshness, eval-only response boundaries, and no-bypass controls |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 5.49 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 6.26 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 8.97 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 5.80 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
