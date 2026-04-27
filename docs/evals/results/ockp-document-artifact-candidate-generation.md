# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `document-artifact-candidate-generation`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `16.99`
- Harness elapsed seconds: `229.77`
- Effective parallel speedup: `0.84x`
- Parallel efficiency: `0.84`
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
| copy_repo | 0.18 |
| install_variant | 19.82 |
| warm_cache | 0.00 |
| seed_data | 0.02 |
| agent_run | 192.61 |
| parse_metrics | 0.00 |
| verify | 0.15 |
| total | 212.77 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `candidate-note-from-pasted-content` | `completed` | 8 | 8 | 3 | 3713 | 38.53 | `<run-root>/production/candidate-note-from-pasted-content/turn-1/events.jsonl` |
| `production` | `candidate-title-and-path-from-heading` | `failed` | 0 | 0 | 2 | 2619 | 13.80 | `<run-root>/production/candidate-title-and-path-from-heading/turn-1/events.jsonl` |
| `production` | `candidate-mixed-source-summary` | `failed` | 4 | 4 | 4 | 14732 | 20.95 | `<run-root>/production/candidate-mixed-source-summary/turn-1/events.jsonl` |
| `production` | `candidate-explicit-overrides-win` | `failed` | 12 | 12 | 7 | 9233 | 42.91 | `<run-root>/production/candidate-explicit-overrides-win/turn-1/events.jsonl` |
| `production` | `candidate-duplicate-risk-asks` | `completed` | 10 | 10 | 5 | 4616 | 22.49 | `<run-root>/production/candidate-duplicate-risk-asks/turn-1/events.jsonl` |
| `production` | `candidate-low-confidence-asks` | `failed` | 2 | 2 | 2 | 5402 | 8.88 | `<run-root>/production/candidate-low-confidence-asks/turn-1/events.jsonl` |
| `production` | `candidate-body-faithfulness` | `failed` | 18 | 18 | 6 | 30120 | 45.05 | `<run-root>/production/candidate-body-faithfulness/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `defer_for_candidate_quality_repair`

Public surface: `skills/openclerk/SKILL.md`, `openclerk document`, `openclerk retrieval`

Promotion: no promoted skill policy yet; repair candidate quality gaps before any propose-before-create skill behavior change.

| Variant | Scenario | Status | Failure classification | Evidence posture |
| --- | --- | --- | --- | --- |
| `production` | `candidate-note-from-pasted-content` | `completed` | `none` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-title-and-path-from-heading` | `failed` | `candidate_quality_gap` | candidate proposal did not satisfy path/title/body quality, duplicate, or confirmation rubric |
| `production` | `candidate-mixed-source-summary` | `failed` | `candidate_quality_gap` | candidate proposal did not satisfy path/title/body quality, duplicate, or confirmation rubric |
| `production` | `candidate-explicit-overrides-win` | `failed` | `candidate_quality_gap` | candidate proposal did not satisfy path/title/body quality, duplicate, or confirmation rubric |
| `production` | `candidate-duplicate-risk-asks` | `completed` | `none` | candidate generation quality rubric satisfied without writing before approval |
| `production` | `candidate-low-confidence-asks` | `failed` | `skill_guidance_or_eval_coverage` | low-confidence candidate pressure did not stay no-tools |
| `production` | `candidate-body-faithfulness` | `failed` | `candidate_quality_gap` | candidate proposal did not satisfy path/title/body quality, duplicate, or confirmation rubric |
