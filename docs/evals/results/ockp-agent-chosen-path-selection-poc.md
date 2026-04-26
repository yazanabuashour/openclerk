# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `agent-chosen-path-selection-poc`
- Release blocking: `false`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `18.64`
- Harness elapsed seconds: `115.74`
- Effective parallel speedup: `1.50x`
- Parallel efficiency: `0.38`
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
| copy_repo | 0.32 |
| install_variant | 29.02 |
| warm_cache | 0.00 |
| seed_data | 0.05 |
| agent_run | 173.53 |
| parse_metrics | 0.00 |
| verify | 0.15 |
| total | 203.12 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `url-only-documentation-path-proposal` | `completed` | 0 | 0 | 1 | 2710 | 5.22 | `<run-root>/production/url-only-documentation-path-proposal/turn-1/events.jsonl` |
| `production` | `url-only-documentation-autonomous-placement` | `completed` | 4 | 4 | 3 | 25059 | 13.62 | `<run-root>/production/url-only-documentation-autonomous-placement/turn-1/events.jsonl` |
| `production` | `multi-source-synthesis-path-selection` | `completed` | 46 | 46 | 14 | 16161 | 93.22 | `<run-root>/production/multi-source-synthesis-path-selection/turn-1/events.jsonl` |
| `production` | `ambiguous-document-type-path-selection` | `completed` | 8 | 8 | 3 | 35672 | 20.18 | `<run-root>/production/ambiguous-document-type-path-selection/turn-1/events.jsonl` |
| `production` | `user-path-instructions-win` | `completed` | 4 | 4 | 3 | 6099 | 12.41 | `<run-root>/production/user-path-instructions-win/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 16401 | 6.59 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2559 | 5.03 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2423 | 6.25 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 21366 | 11.01 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `keep_as_reference`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: no promoted runner action, schema, migration, storage API, product behavior, public OpenClerk interface, or change to missing-path clarification.

| Variant | Scenario | Status | Failure classification | Evidence posture |
| --- | --- | --- | --- | --- |
| `production` | `url-only-documentation-path-proposal` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `url-only-documentation-autonomous-placement` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `multi-source-synthesis-path-selection` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `ambiguous-document-type-path-selection` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `user-path-instructions-win` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `missing-document-path-reject` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `negative-limit-reject` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `unsupported-transport-reject` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
