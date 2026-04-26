# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `agent-chosen-path-selection-poc`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `19.19`
- Harness elapsed seconds: `221.40`
- Effective parallel speedup: `0.77x`
- Parallel efficiency: `0.77`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 3/30 production scenarios passed; missing: create-note, search-synthesis, answer-filing, rag-retrieval-baseline, canonical-docs-navigation-baseline, graph-semantics-reference-poc, memory-router-reference-poc, configured-layout-explain, invalid-layout-visible, stale-synthesis-update, synthesis-freshness-repair, source-sensitive-audit-repair, source-sensitive-conflict-explain, synthesis-candidate-pressure, synthesis-source-set-pressure, append-replace, records-provenance, promoted-record-vs-docs, decision-record-vs-docs, decision-supersession-freshness, decision-real-adr-migration, duplicate-path-reject, mixed-synthesis-records, mt-source-then-synthesis, mt-synthesis-drift-pressure, mt-incomplete-then-create |
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
| copy_repo | 0.18 |
| install_variant | 31.13 |
| warm_cache | 0.00 |
| seed_data | 0.04 |
| agent_run | 170.65 |
| parse_metrics | 0.00 |
| verify | 0.16 |
| total | 202.22 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `url-only-documentation-path-proposal` | `failed` | 0 | 0 | 1 | 2490 | 6.16 | `<run-root>/production/url-only-documentation-path-proposal/turn-1/events.jsonl` |
| `production` | `url-only-documentation-autonomous-placement` | `completed` | 4 | 4 | 3 | 5963 | 16.11 | `<run-root>/production/url-only-documentation-autonomous-placement/turn-1/events.jsonl` |
| `production` | `multi-source-synthesis-path-selection` | `completed` | 42 | 42 | 10 | 12600 | 66.61 | `<run-root>/production/multi-source-synthesis-path-selection/turn-1/events.jsonl` |
| `production` | `ambiguous-document-type-path-selection` | `failed` | 10 | 10 | 5 | 7163 | 27.14 | `<run-root>/production/ambiguous-document-type-path-selection/turn-1/events.jsonl` |
| `production` | `user-path-instructions-win` | `completed` | 6 | 6 | 4 | 9109 | 17.47 | `<run-root>/production/user-path-instructions-win/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2398 | 8.19 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `failed` | 0 | 0 | 1 | 2546 | 4.70 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2410 | 14.96 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2409 | 9.31 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `keep_as_reference`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: no promoted runner action, schema, migration, storage API, product behavior, public OpenClerk interface, or change to missing-path clarification.

| Variant | Scenario | Status | Failure classification | Evidence posture |
| --- | --- | --- | --- | --- |
| `production` | `url-only-documentation-path-proposal` | `failed` | `skill_guidance_or_eval_coverage` | runner-visible evidence existed, but the assistant answer did not satisfy the path-selection scenario |
| `production` | `url-only-documentation-autonomous-placement` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `multi-source-synthesis-path-selection` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `ambiguous-document-type-path-selection` | `failed` | `skill_guidance_or_eval_coverage` | runner-visible evidence existed, but the assistant answer did not satisfy the path-selection scenario |
| `production` | `user-path-instructions-win` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `missing-document-path-reject` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `negative-limit-reject` | `failed` | `skill_guidance_or_eval_coverage` | runner-visible evidence existed, but the assistant answer did not satisfy the path-selection scenario |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
| `production` | `unsupported-transport-reject` | `completed` | `none` | current runner/skill behavior preserved path-selection invariants |
