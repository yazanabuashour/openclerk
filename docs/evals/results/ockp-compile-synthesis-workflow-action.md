# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `compile-synthesis-workflow-action`
- Release blocking: `false`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `16.00`
- Harness elapsed seconds: `106.65`
- Effective parallel speedup: `0.82x`
- Parallel efficiency: `0.21`
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
| copy_repo | 0.05 |
| install_variant | 3.42 |
| warm_cache | 0.00 |
| seed_data | 0.02 |
| agent_run | 87.12 |
| parse_metrics | 0.01 |
| verify | 0.02 |
| total | 90.65 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `compile-synthesis-workflow-action-natural` | `completed` | 72 | 72 | 19 | 51342 | 87.12 | `<run-root>/production/compile-synthesis-workflow-action-natural/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `accept_compile_synthesis_workflow_action`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: implemented narrow compile_synthesis document action plus existing primitives for advanced/manual cases; no schema migration, direct vault behavior, broad synthesis engine, or source authority change.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety pass | Capability pass | UX quality | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| `production` | `compile-synthesis-workflow-action-natural` | `completed` | `workflow_choreography_gap` | 72 | 72 | 19 | 87.12 | `natural-user-intent` | `completed` | `normal` | 0 | 72 | `high` | `high_ceremony_promoted_workflow_action` | `pass` | `pass` | `taste_debt` | `none_observed` | `not_applicable` | compile_synthesis preserved source authority and passed, but natural workflow-action use still required more commands or assistant turns than the low-ceremony UX threshold |

## Candidate Comparison Context

| Row | Scenario | Prompt specificity | Safety pass | Capability pass | UX quality | Interpretation |
| --- | --- | --- | --- | --- | --- | --- |
| Candidate A | `compile-synthesis-current-primitives-control` | scripted-control | pass | pass | baseline_ceremonial_control | Existing primitives prove capability but require workflow choreography. |
| Candidate B | `compile-synthesis-response-candidate` | candidate-response-contract | pass | pass | candidate_contract_complete | Candidate fields are useful, but should be runner-owned instead of skill choreography. |
| Candidate C | `compile-synthesis-workflow-action-natural` | natural-user-intent | pass | pass | taste_debt | Selected narrow action plus existing primitives; residual command ceremony is tracked by `oc-nj5h`. |
