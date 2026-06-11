# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `compile-synthesis-workflow-action`
- Release blocking: `false`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `29.73`
- Harness elapsed seconds: `61.51`
- Effective parallel speedup: `0.46x`
- Parallel efficiency: `0.12`
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
| copy_repo | 0.16 |
| install_variant | 3.45 |
| warm_cache | 0.00 |
| seed_data | 0.03 |
| agent_run | 28.08 |
| parse_metrics | 0.00 |
| verify | 0.05 |
| total | 31.78 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `compile-synthesis-workflow-action-natural` | `completed` | 2 | 2 | 3 | 10487 | 28.08 | `<run-root>/production/compile-synthesis-workflow-action-natural/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `retain_compile_synthesis_surface_with_taste_debt_followup`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: implemented narrow compile_synthesis document action plus existing primitives for advanced/manual cases; no schema migration, direct vault behavior, broad synthesis engine, or source authority change; refreshed natural-row UX is taste debt, so acceptable-UX release claims are deferred to follow-up surface comparison.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Workflow first command | Workflow calls | Setup discovery | Pre-action setup discovery | Pre-action primitives | Post-action primitives | Final-answer repair turns | Latency | Guidance dependence | Safety pass | Capability pass | UX quality | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| `production` | `compile-synthesis-workflow-action-natural` | `completed` | `workflow_choreography_gap` | 2 | 2 | 3 | 28.08 | `natural-user-intent` | `completed` | `normal` | 0 | 2 | 1 | 2 | 0 | 0 | 0 | 0 | 0 | `medium` | `high_ceremony_promoted_workflow_action` | `pass` | `pass` | `taste_debt` | `none_observed` | `not_applicable` | compile_synthesis preserved source authority and passed, but natural workflow-action use still required more commands or assistant turns than the low-ceremony UX threshold |
