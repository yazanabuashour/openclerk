# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `agent-chosen-path-selection-poc`
- Release blocking: `false`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `17.87`
- Harness elapsed seconds: `50.55`
- Effective parallel speedup: `0.91x`
- Parallel efficiency: `0.23`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 1/30 production scenarios passed; missing: create-note, search-synthesis, answer-filing, rag-retrieval-baseline, canonical-docs-navigation-baseline, graph-semantics-reference-poc, memory-router-reference-poc, configured-layout-explain, invalid-layout-visible, stale-synthesis-update, synthesis-freshness-repair, source-sensitive-audit-repair, source-sensitive-conflict-explain, synthesis-candidate-pressure, synthesis-source-set-pressure, append-replace, records-provenance, promoted-record-vs-docs, decision-record-vs-docs, decision-supersession-freshness, decision-real-adr-migration, missing-document-path-reject, unsupported-lower-level-reject, unsupported-transport-reject, duplicate-path-reject, mixed-synthesis-records, mt-source-then-synthesis, mt-synthesis-drift-pressure, mt-incomplete-then-create |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect retired API files or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `pass` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_legacy_source_runner_usage` | `pass` | production must not invoke source-built or legacy runner paths instead of installed openclerk |
| `no_direct_sqlite_access` | `pass` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `fail` | missing final-answer-only validation scenarios: missing-document-path-reject, unsupported-lower-level-reject, unsupported-transport-reject |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.09 |
| install_variant | 12.23 |
| warm_cache | 0.00 |
| seed_data | 0.00 |
| agent_run | 46.15 |
| parse_metrics | 0.00 |
| verify | 0.06 |
| total | 58.55 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `url-only-documentation-path-proposal` | `completed` | 0 | 0 | 1 | 2710 | 9.68 | `<run-root>/production/url-only-documentation-path-proposal/turn-1/events.jsonl` |
| `production` | `ambiguous-document-type-path-selection` | `completed` | 10 | 10 | 5 | 41911 | 28.43 | `<run-root>/production/ambiguous-document-type-path-selection/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 21337 | 8.04 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `keep_as_reference`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: no promoted runner action, schema, migration, storage API, product behavior, public OpenClerk interface, or change to missing-path clarification.

| Variant | Scenario | Status | Failure classification | Evidence posture |
| --- | --- | --- | --- | --- |
| `production` | `url-only-documentation-path-proposal` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `ambiguous-document-type-path-selection` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `negative-limit-reject` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
