# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `high-touch-web-url-stale-repair-ceremony`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `18.24`
- Harness elapsed seconds: `156.12`
- Effective parallel speedup: `0.75x`
- Parallel efficiency: `0.75`
- Targeted acceptance: high-touch web URL stale repair rows report natural refresh intent, scripted update-mode control, duplicate/no-op behavior, changed-source freshness evidence, dependent synthesis stale visibility, provenance/freshness, no browser/manual acquisition, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and separate safety/capability/UX classification
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
| copy_repo | 0.24 |
| install_variant | 19.62 |
| warm_cache | 0.00 |
| seed_data | 0.04 |
| agent_run | 117.87 |
| parse_metrics | 0.00 |
| verify | 0.07 |
| total | 137.88 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `high-touch-web-url-stale-repair-natural-intent` | `failed` | 24 | 24 | 6 | 14886 | 65.21 | `<run-root>/production/high-touch-web-url-stale-repair-natural-intent/turn-1/events.jsonl` |
| `production` | `high-touch-web-url-stale-repair-scripted-control` | `completed` | 26 | 26 | 5 | 17156 | 36.69 | `<run-root>/production/high-touch-web-url-stale-repair-scripted-control/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2543 | 4.11 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2524 | 3.90 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2556 | 3.97 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2554 | 3.99 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `defer_for_guidance_or_eval_repair`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: web URL stale repair promotion deferred pending guidance, answer-contract, harness, report, or eval repair; no implementation bead unless a later decision promotes.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `high-touch-web-url-stale-repair-natural-intent` | `failed` | `ergonomics_gap` | 24 | 24 | 6 | 65.21 | `natural-user-intent` | `answer_repair_needed` | `normal` | 0 | 24 | `high` | `scenario_prompt` | `none_observed` | `not_applicable` | natural web URL stale repair intent did not complete the safe current-primitives workflow |
| `production` | `high-touch-web-url-stale-repair-scripted-control` | `completed` | `none` | 26 | 26 | 5 | 36.69 | `scripted-control` | `completed` | `normal` | 0 | 26 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | current document/retrieval workflow preserved runner-owned public fetch, duplicate/no-op behavior, stale synthesis visibility, provenance/freshness, and no-browser boundaries |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 4.11 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 3.90 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 3.97 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 3.99 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
