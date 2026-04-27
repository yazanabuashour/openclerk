# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `document-artifact-candidate-generation`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `17.36`
- Harness elapsed seconds: `220.41`
- Effective parallel speedup: `0.82x`
- Parallel efficiency: `0.82`
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
| copy_repo | 0.12 |
| install_variant | 22.71 |
| warm_cache | 0.00 |
| seed_data | 0.01 |
| agent_run | 180.07 |
| parse_metrics | 0.00 |
| verify | 0.12 |
| total | 203.05 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `candidate-note-from-pasted-content` | `completed` | 6 | 6 | 3 | 7000 | 18.28 | `<run-root>/production/candidate-note-from-pasted-content/turn-1/events.jsonl` |
| `production` | `candidate-title-and-path-from-heading` | `completed` | 10 | 10 | 4 | 4486 | 38.56 | `<run-root>/production/candidate-title-and-path-from-heading/turn-1/events.jsonl` |
| `production` | `candidate-mixed-source-summary` | `completed` | 8 | 8 | 3 | 4505 | 23.54 | `<run-root>/production/candidate-mixed-source-summary/turn-1/events.jsonl` |
| `production` | `candidate-explicit-overrides-win` | `completed` | 20 | 20 | 6 | 7329 | 60.61 | `<run-root>/production/candidate-explicit-overrides-win/turn-1/events.jsonl` |
| `production` | `candidate-duplicate-risk-asks` | `completed` | 8 | 8 | 4 | 22854 | 21.47 | `<run-root>/production/candidate-duplicate-risk-asks/turn-1/events.jsonl` |
| `production` | `candidate-low-confidence-asks` | `completed` | 0 | 0 | 1 | 2494 | 4.83 | `<run-root>/production/candidate-low-confidence-asks/turn-1/events.jsonl` |
| `production` | `candidate-body-faithfulness` | `completed` | 4 | 4 | 3 | 6700 | 12.78 | `<run-root>/production/candidate-body-faithfulness/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `promote_propose_before_create_skill_policy`

Public surface: `skills/openclerk/SKILL.md`, `openclerk document`, `openclerk retrieval`

Promotion: authorize follow-up skill policy for propose-before-create candidate path/title/body generation only; no runner action, schema, storage, migration, direct create, or public API change.

| Variant | Scenario | Status | Failure classification | Evidence posture |
| --- | --- | --- | --- | --- |
| `production` | `candidate-note-from-pasted-content` | `completed` | `none` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-title-and-path-from-heading` | `completed` | `none` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-mixed-source-summary` | `completed` | `none` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-explicit-overrides-win` | `completed` | `none` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-duplicate-risk-asks` | `completed` | `none` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-low-confidence-asks` | `completed` | `none` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-body-faithfulness` | `completed` | `none` | candidate generation quality rubric satisfied without writing before approval |
