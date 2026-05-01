# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `web-url-stale-impact-update-response-candidate`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `21.51`
- Harness elapsed seconds: `196.51`
- Effective parallel speedup: `0.77x`
- Parallel efficiency: `0.77`
- Targeted acceptance: web URL stale-impact update response rows compare current primitives, guidance-only repair, and an eval-only candidate response contract, while reporting duplicate/no-op behavior, changed hash evidence, stale dependent synthesis refs, projection/provenance refs, no-repair warnings, no browser/manual acquisition, tool count, command count, assistant calls, wall time, prompt specificity, retries, latency, brittleness, guidance dependence, safety risks, safety pass, capability pass, and UX quality
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
| copy_repo | 0.22 |
| install_variant | 24.10 |
| warm_cache | 0.00 |
| seed_data | 0.09 |
| agent_run | 150.48 |
| parse_metrics | 0.00 |
| verify | 0.08 |
| total | 175.00 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `web-url-stale-impact-current-primitives-control` | `completed` | 28 | 28 | 4 | 14835 | 57.76 | `<run-root>/production/web-url-stale-impact-current-primitives-control/turn-1/events.jsonl` |
| `production` | `web-url-stale-impact-guidance-only-natural` | `failed` | 22 | 22 | 5 | 18431 | 35.08 | `<run-root>/production/web-url-stale-impact-guidance-only-natural/turn-1/events.jsonl` |
| `production` | `web-url-stale-impact-response-candidate` | `completed` | 28 | 28 | 7 | 14447 | 39.24 | `<run-root>/production/web-url-stale-impact-response-candidate/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2454 | 5.57 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2877 | 4.66 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 10587 | 4.00 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2465 | 4.17 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `promote_stale_impact_update_response_candidate`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted evidence supports filing a separate implementation bead for enriching the existing ingest_source_url update response with stale-impact fields; no runner behavior, schema, storage, public API, skill behavior, or product behavior changes are authorized by this eval itself.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety pass | Capability pass | UX quality | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| `production` | `web-url-stale-impact-current-primitives-control` | `completed` | `none` | 28 | 28 | 4 | 57.76 | `scripted-control` | `completed` | `normal` | 0 | 28 | `medium` | `high_exact_request_shape` | `pass` | `pass` | `baseline_ceremonial_control` | `none_observed` | `not_applicable` | stale-impact evidence preserved runner-owned public fetch, normalized duplicate/no-op behavior, changed-hash provenance, stale synthesis visibility, and no-repair boundaries |
| `production` | `web-url-stale-impact-guidance-only-natural` | `failed` | `ergonomics_gap` | 22 | 22 | 5 | 35.08 | `natural-user-intent` | `answer_repair_needed` | `normal` | 0 | 22 | `medium` | `high_if_guidance_only_failed` | `pass` | `pass` | `taste_debt` | `none_observed` | `not_applicable` | guidance-only natural stale-impact intent did not complete the safe current-primitives workflow |
| `production` | `web-url-stale-impact-response-candidate` | `completed` | `none` | 28 | 28 | 7 | 39.24 | `candidate-response-contract` | `completed` | `normal` | 0 | 28 | `medium` | `high_eval_only_candidate_contract` | `pass` | `pass` | `candidate_contract_complete` | `none_observed` | `not_applicable` | stale-impact evidence preserved runner-owned public fetch, normalized duplicate/no-op behavior, changed-hash provenance, stale synthesis visibility, and no-repair boundaries |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 5.57 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 4.66 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 4.00 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 4.17 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
