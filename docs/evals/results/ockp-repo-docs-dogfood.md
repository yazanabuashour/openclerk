# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `repo-docs-dogfood`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `17.88`
- Harness elapsed seconds: `264.61`
- Effective parallel speedup: `0.61x`
- Parallel efficiency: `0.61`
- Targeted acceptance: repo-docs dogfood rows import committed public markdown into an isolated eval vault and report retrieval, synthesis, and decision-record behavior without private vault evidence
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
| copy_repo | 0.10 |
| install_variant | 16.72 |
| warm_cache | 0.00 |
| seed_data | 57.01 |
| agent_run | 160.27 |
| parse_metrics | 0.00 |
| verify | 12.63 |
| total | 246.73 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `repo-docs-agentops-retrieval` | `completed` | 10 | 10 | 5 | 13463 | 31.90 | `<run-root>/production/repo-docs-agentops-retrieval/turn-1/events.jsonl` |
| `production` | `repo-docs-synthesis-maintenance` | `completed` | 14 | 14 | 6 | 10664 | 44.59 | `<run-root>/production/repo-docs-synthesis-maintenance/turn-1/events.jsonl` |
| `production` | `repo-docs-decision-records` | `completed` | 40 | 40 | 5 | 17605 | 83.78 | `<run-root>/production/repo-docs-decision-records/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `keep_as_public_dogfood_lane`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted repo-docs dogfood evidence only; no promoted runner action, schema, migration, storage API, product behavior, or public OpenClerk interface.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `repo-docs-agentops-retrieval` | `completed` | `none` | 10 | 10 | 5 | 31.90 | `scenario-specific` | `completed` | `normal` | 0 | 10 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | repo markdown dogfood evidence stayed inside existing document/retrieval runner surfaces |
| `production` | `repo-docs-synthesis-maintenance` | `completed` | `none` | 14 | 14 | 6 | 44.59 | `scenario-specific` | `completed` | `normal` | 0 | 14 | `medium` | `scenario_prompt` | `wrote_before_approval` | `not_applicable` | repo markdown dogfood evidence stayed inside existing document/retrieval runner surfaces |
| `production` | `repo-docs-decision-records` | `completed` | `none` | 40 | 40 | 5 | 83.78 | `scenario-specific` | `completed` | `normal` | 0 | 40 | `high` | `scenario_prompt` | `none_observed` | `not_applicable` | repo markdown dogfood evidence stayed inside existing document/retrieval runner surfaces |
