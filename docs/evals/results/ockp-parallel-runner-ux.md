# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `parallel-runner-ux`
- Release blocking: `false`
- Configured parallelism: `2`
- Cache mode: `shared`
- Cache prewarm seconds: `26.64`
- Harness elapsed seconds: `53.43`
- Effective parallel speedup: `0.59x`
- Parallel efficiency: `0.30`
- Targeted acceptance: parallel runner rows report fresh startup and safe-read command UX, tool count, command count, assistant calls, wall time, guidance dependence, safety risks, and raw SQLite/runtime_config/upsert failure absence
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
| copy_repo | 0.08 |
| install_variant | 8.19 |
| warm_cache | 0.00 |
| seed_data | 0.02 |
| agent_run | 31.43 |
| parse_metrics | 0.00 |
| verify | 0.01 |
| total | 39.75 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `parallel-runner-safe-startup` | `completed` | 10 | 10 | 3 | 44334 | 22.81 | `<run-root>/production/parallel-runner-safe-startup/turn-1/events.jsonl` |
| `production` | `parallel-runner-safe-reads` | `failed` | 0 | 0 | 1 | 8761 | 8.62 | `<run-root>/production/parallel-runner-safe-reads/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `relax_skill_guidance_for_safe_parallel_reads`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted parallel runner UX evidence for documented safe read/startup workflows; no public JSON schema, storage schema, or write-concurrency expansion.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `parallel-runner-safe-startup` | `completed` | `none` | 10 | 10 | 3 | 22.81 | `scenario-specific` | `completed` | `normal` | 0 | 10 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | parallel startup/read workflow completed through installed runner commands without raw SQLite/runtime_config/upsert failures |
| `production` | `parallel-runner-safe-reads` | `failed` | `skill_guidance_or_eval_coverage` | 0 | 0 | 1 | 8.62 | `scenario-specific` | `manual_review` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | runner-visible parallel evidence existed, but the assistant answer or required runner steps did not satisfy the scenario |
