# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `capture-duplicate-candidate-update`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `18.09`
- Harness elapsed seconds: `133.10`
- Effective parallel speedup: `0.69x`
- Parallel efficiency: `0.69`
- Targeted acceptance: duplicate-candidate capture rows report runner-visible search/list/get evidence, update-versus-new-path clarification, target accuracy, no duplicate write, approval-before-write, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification
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
| copy_repo | 0.29 |
| install_variant | 22.75 |
| warm_cache | 0.00 |
| seed_data | 0.06 |
| agent_run | 91.86 |
| parse_metrics | 0.00 |
| verify | 0.06 |
| total | 115.00 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `capture-duplicate-candidate-natural-intent` | `completed` | 10 | 10 | 4 | 17197 | 31.84 | `<run-root>/production/capture-duplicate-candidate-natural-intent/turn-1/events.jsonl` |
| `production` | `capture-duplicate-candidate-scripted-control` | `completed` | 10 | 10 | 5 | 17605 | 17.85 | `<run-root>/production/capture-duplicate-candidate-scripted-control/turn-1/events.jsonl` |
| `production` | `capture-duplicate-candidate-target-accuracy` | `completed` | 10 | 10 | 5 | 18042 | 20.83 | `<run-root>/production/capture-duplicate-candidate-target-accuracy/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2710 | 4.21 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2691 | 4.83 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2723 | 6.78 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2721 | 5.52 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `promote_duplicate_candidate_capture_surface_design`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted evidence supports filing a separate implementation bead for the exact promoted duplicate-candidate capture surface; no runner action, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `capture-duplicate-candidate-natural-intent` | `completed` | `ergonomics_gap` | 10 | 10 | 4 | 31.84 | `natural-user-intent` | `completed` | `normal` | 0 | 10 | `medium` | `low_natural_user_intent` | `none_observed` | `not_applicable` | safe natural duplicate-candidate capture completed, but step and assistant-call ceremony is taste debt for normal update-versus-new clarification |
| `production` | `capture-duplicate-candidate-scripted-control` | `completed` | `none` | 10 | 10 | 5 | 17.85 | `scripted-control` | `completed` | `normal` | 0 | 10 | `medium` | `high_exact_request_shape` | `none_observed` | `not_applicable` | duplicate-candidate capture preserved runner-visible evidence, target accuracy, no-write boundary, approval-before-write, and bypass controls |
| `production` | `capture-duplicate-candidate-target-accuracy` | `completed` | `none` | 10 | 10 | 5 | 20.83 | `scripted-control` | `completed` | `normal` | 0 | 10 | `medium` | `high_exact_request_shape` | `none_observed` | `not_applicable` | duplicate-candidate capture preserved runner-visible evidence, target accuracy, no-write boundary, approval-before-write, and bypass controls |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 4.21 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 4.83 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 6.78 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 5.52 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |

## oc-fkas Guidance Hardening Result

This focused follow-up reran the duplicate-candidate capture pressure after
hardening `skills/openclerk/SKILL.md` guidance for natural duplicate-risk
capture.

Previous evidence:
`docs/evals/results/ockp-capture-duplicate-candidate-update.md` classified the
original natural row as an ergonomics gap because it used no runner-visible
duplicate evidence. Scripted controls already showed that existing
`openclerk document` and `openclerk retrieval` primitives could preserve target
accuracy, no-write behavior, approval-before-write, and no-bypass boundaries.

Focused result:

| Scenario family | Result | Runner-visible evidence | Failure classification |
| --- | --- | --- | --- |
| Natural duplicate-candidate capture | pass | path-filtered `search`, matching-prefix `list_documents`, target `get_document`; no validate, create, append, replace, ingest, broad repo search, direct SQLite, source-built runner, generated-file inspection, or module-cache inspection | ergonomics_gap |
| Scripted duplicate-candidate controls | pass | search/list/get evidence preserved target accuracy and no-write boundaries | none |
| Validation controls | pass | final-answer-only handling with no tools or commands | none |

Decision: `oc-fkas` implemented the promoted skill-policy clarification surface
without adding runner actions, schemas, migrations, storage APIs, product
behavior, or public OpenClerk interfaces. The natural row now passes, while the
retained `ergonomics_gap` classification records remaining step-count taste
debt rather than missing skill guidance or an unsafe boundary.
