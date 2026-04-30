# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `capture-document-these-links-placement`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `21.63`
- Harness elapsed seconds: `202.10`
- Effective parallel speedup: `0.75x`
- Parallel efficiency: `0.75`
- Targeted acceptance: document-these-links placement rows report natural public-link placement intent, approved source fetch control, synthesis placement proposal, duplicate source/synthesis handling, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification
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
| install_variant | 27.77 |
| warm_cache | 0.00 |
| seed_data | 0.04 |
| agent_run | 152.33 |
| parse_metrics | 0.00 |
| verify | 0.08 |
| total | 180.47 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `capture-document-these-links-natural-intent` | `failed` | 6 | 6 | 4 | 27890 | 19.57 | `<run-root>/production/capture-document-these-links-natural-intent/turn-1/events.jsonl` |
| `production` | `capture-document-these-links-source-fetch-control` | `completed` | 6 | 6 | 4 | 7112 | 23.78 | `<run-root>/production/capture-document-these-links-source-fetch-control/turn-1/events.jsonl` |
| `production` | `capture-document-these-links-synthesis-placement` | `completed` | 20 | 20 | 6 | 11654 | 30.83 | `<run-root>/production/capture-document-these-links-synthesis-placement/turn-1/events.jsonl` |
| `production` | `capture-document-these-links-duplicate-placement` | `completed` | 18 | 18 | 7 | 27428 | 46.58 | `<run-root>/production/capture-document-these-links-duplicate-placement/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2718 | 9.58 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2699 | 9.36 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 8529 | 4.47 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2729 | 8.16 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `promote_document_these_links_placement_surface_design`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted evidence supports filing a separate implementation bead for the exact promoted document-these-links placement surface; no runner action, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `capture-document-these-links-natural-intent` | `failed` | `ergonomics_gap` | 6 | 6 | 4 | 19.57 | `scenario-specific` | `answer_repair_needed` | `normal` | 0 | 6 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | natural document-these-links placement intent did not complete the safe current-primitives workflow |
| `production` | `capture-document-these-links-source-fetch-control` | `completed` | `none` | 6 | 6 | 4 | 23.78 | `scenario-specific` | `completed` | `normal` | 0 | 6 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | document-these-links placement preserved public-fetch permission, durable-write approval, source path hints, synthesis placement, duplicate handling, and bypass controls |
| `production` | `capture-document-these-links-synthesis-placement` | `completed` | `none` | 20 | 20 | 6 | 30.83 | `scenario-specific` | `completed` | `normal` | 0 | 20 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | document-these-links placement preserved public-fetch permission, durable-write approval, source path hints, synthesis placement, duplicate handling, and bypass controls |
| `production` | `capture-document-these-links-duplicate-placement` | `completed` | `none` | 18 | 18 | 7 | 46.58 | `scenario-specific` | `completed` | `normal` | 0 | 18 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | document-these-links placement preserved public-fetch permission, durable-write approval, source path hints, synthesis placement, duplicate handling, and bypass controls |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 9.58 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 9.36 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 4.47 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 8.16 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
