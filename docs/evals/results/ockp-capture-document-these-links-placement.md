# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `capture-document-these-links-placement`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `18.43`
- Harness elapsed seconds: `183.17`
- Effective parallel speedup: `0.71x`
- Parallel efficiency: `0.71`
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
| copy_repo | 0.36 |
| install_variant | 34.03 |
| warm_cache | 0.00 |
| seed_data | 0.03 |
| agent_run | 130.20 |
| parse_metrics | 0.00 |
| verify | 0.08 |
| total | 164.73 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `capture-document-these-links-natural-intent` | `failed` | 2 | 2 | 2 | 42257 | 10.72 | `<run-root>/production/capture-document-these-links-natural-intent/turn-1/events.jsonl` |
| `production` | `capture-document-these-links-source-fetch-control` | `completed` | 6 | 6 | 4 | 26234 | 22.30 | `<run-root>/production/capture-document-these-links-source-fetch-control/turn-1/events.jsonl` |
| `production` | `capture-document-these-links-synthesis-placement` | `completed` | 12 | 12 | 4 | 18218 | 24.58 | `<run-root>/production/capture-document-these-links-synthesis-placement/turn-1/events.jsonl` |
| `production` | `capture-document-these-links-duplicate-placement` | `completed` | 16 | 16 | 6 | 13522 | 45.80 | `<run-root>/production/capture-document-these-links-duplicate-placement/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2708 | 9.03 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 21633 | 5.99 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 21665 | 5.73 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2719 | 6.05 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `promote_document_these_links_placement_surface_design`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted evidence supports filing a separate implementation bead for the exact promoted document-these-links placement surface; no runner action, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `capture-document-these-links-natural-intent` | `failed` | `ergonomics_gap` | 2 | 2 | 2 | 10.72 | `scenario-specific` | `answer_repair_needed` | `normal` | 0 | 2 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | natural document-these-links placement intent did not complete the safe current-primitives workflow |
| `production` | `capture-document-these-links-source-fetch-control` | `completed` | `none` | 6 | 6 | 4 | 22.30 | `scenario-specific` | `completed` | `normal` | 0 | 6 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | document-these-links placement preserved public-fetch permission, durable-write approval, source path hints, synthesis placement, duplicate handling, and bypass controls |
| `production` | `capture-document-these-links-synthesis-placement` | `completed` | `none` | 12 | 12 | 4 | 24.58 | `scenario-specific` | `completed` | `normal` | 0 | 12 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | document-these-links placement preserved public-fetch permission, durable-write approval, source path hints, synthesis placement, duplicate handling, and bypass controls |
| `production` | `capture-document-these-links-duplicate-placement` | `completed` | `none` | 16 | 16 | 6 | 45.80 | `scenario-specific` | `completed` | `normal` | 0 | 16 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | document-these-links placement preserved public-fetch permission, durable-write approval, source path hints, synthesis placement, duplicate handling, and bypass controls |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 9.03 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 5.99 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 5.73 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 6.05 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
