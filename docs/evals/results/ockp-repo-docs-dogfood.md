# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `repo-docs-dogfood`
- Release blocking: `false`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `15.33`
- Harness elapsed seconds: `129.73`
- Effective parallel speedup: `1.49x`
- Parallel efficiency: `0.37`
- Targeted acceptance: repo-docs dogfood rows import committed public markdown into an isolated eval vault and report retrieval, synthesis, decision-records, release-readiness, tag filtering, memory-router recall report behavior, and release synthesis freshness without private vault evidence
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
| copy_repo | 1.05 |
| install_variant | 22.93 |
| warm_cache | 0.00 |
| seed_data | 165.68 |
| agent_run | 193.76 |
| parse_metrics | 0.00 |
| verify | 0.71 |
| total | 384.16 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `repo-docs-agentops-retrieval` | `completed` | 3 | 3 | 4 | 18392 | 28.03 | `<run-root>/production/repo-docs-agentops-retrieval/turn-1/events.jsonl` |
| `production` | `repo-docs-synthesis-maintenance` | `completed` | 4 | 4 | 4 | 34454 | 23.99 | `<run-root>/production/repo-docs-synthesis-maintenance/turn-1/events.jsonl` |
| `production` | `repo-docs-decision-records` | `completed` | 7 | 7 | 3 | 35178 | 47.49 | `<run-root>/production/repo-docs-decision-records/turn-1/events.jsonl` |
| `production` | `repo-docs-release-readiness` | `completed` | 2 | 2 | 3 | 25944 | 26.43 | `<run-root>/production/repo-docs-release-readiness/turn-1/events.jsonl` |
| `production` | `repo-docs-tag-filter` | `completed` | 3 | 3 | 3 | 10445 | 18.89 | `<run-root>/production/repo-docs-tag-filter/turn-1/events.jsonl` |
| `production` | `repo-docs-memory-router-recall-report` | `completed` | 2 | 2 | 3 | 12926 | 14.88 | `<run-root>/production/repo-docs-memory-router-recall-report/turn-1/events.jsonl` |
| `production` | `repo-docs-release-synthesis-freshness` | `completed` | 6 | 6 | 3 | 15461 | 34.05 | `<run-root>/production/repo-docs-release-synthesis-freshness/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `keep_as_public_dogfood_lane`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted repo-docs dogfood evidence only; no promoted runner action, schema, migration, storage API, product behavior, or public OpenClerk interface.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Workflow first command | Workflow calls | Setup discovery | Pre-action setup discovery | Pre-action primitives | Post-action primitives | Final-answer repair turns | Latency | Guidance dependence | Safety pass | Capability pass | UX quality | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| `production` | `repo-docs-agentops-retrieval` | `completed` | `none` | 3 | 3 | 4 | 28.03 | `scenario-specific` | `completed` | `normal` | 0 | 3 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | `medium` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | repo markdown dogfood evidence stayed inside existing document/retrieval runner surfaces |
| `production` | `repo-docs-synthesis-maintenance` | `completed` | `none` | 4 | 4 | 4 | 23.99 | `scenario-specific` | `completed` | `normal` | 0 | 4 | 0 | 0 | 0 | 0 | 3 | 0 | 0 | `medium` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | repo markdown dogfood evidence stayed inside existing document/retrieval runner surfaces |
| `production` | `repo-docs-decision-records` | `completed` | `none` | 7 | 7 | 3 | 47.49 | `scenario-specific` | `completed` | `normal` | 0 | 7 | 0 | 0 | 0 | 0 | 6 | 0 | 0 | `medium` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | repo markdown dogfood evidence stayed inside existing document/retrieval runner surfaces |
| `production` | `repo-docs-release-readiness` | `completed` | `none` | 2 | 2 | 3 | 26.43 | `scenario-specific` | `completed` | `normal` | 0 | 2 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | `medium` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | repo markdown dogfood evidence stayed inside existing document/retrieval runner surfaces |
| `production` | `repo-docs-tag-filter` | `completed` | `none` | 3 | 3 | 3 | 18.89 | `scenario-specific` | `completed` | `normal` | 0 | 3 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | `medium` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | repo markdown dogfood evidence stayed inside existing document/retrieval runner surfaces |
| `production` | `repo-docs-memory-router-recall-report` | `completed` | `none` | 2 | 2 | 3 | 14.88 | `scenario-specific` | `completed` | `normal` | 0 | 2 | 0 | 0 | 0 | 0 | 1 | 0 | 0 | `low` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | repo markdown dogfood evidence stayed inside existing document/retrieval runner surfaces |
| `production` | `repo-docs-release-synthesis-freshness` | `completed` | `none` | 6 | 6 | 3 | 34.05 | `scenario-specific` | `completed` | `normal` | 0 | 6 | 0 | 0 | 0 | 0 | 5 | 0 | 0 | `medium` | `scenario_prompt` | `pass` | `pass` | `completed` | `none_observed` | `not_applicable` | repo markdown dogfood evidence stayed inside existing document/retrieval runner surfaces |
