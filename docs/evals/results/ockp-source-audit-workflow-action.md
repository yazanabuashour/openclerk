# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `source-audit-workflow-action`
- Release blocking: `false`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `15.14`
- Harness elapsed seconds: `48.91`
- Effective parallel speedup: `0.61x`
- Parallel efficiency: `0.15`
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
| copy_repo | 0.05 |
| install_variant | 3.91 |
| warm_cache | 0.00 |
| seed_data | 0.04 |
| agent_run | 29.73 |
| parse_metrics | 0.00 |
| verify | 0.04 |
| total | 33.77 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `source-audit-workflow-action-natural` | `completed` | 14 | 14 | 5 | 33210 | 29.73 | `<run-root>/production/source-audit-workflow-action-natural/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `accept_source_audit_report_workflow_action`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: implemented narrow source_audit_report retrieval action plus existing primitives for advanced/manual cases; broad contradiction engine claims remain rejected.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety pass | Capability pass | UX quality | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| `production` | `source-audit-workflow-action-natural` | `completed` | `workflow_choreography_gap` | 14 | 14 | 5 | 29.73 | `natural-user-intent` | `completed` | `normal` | 0 | 14 | `medium` | `high_ceremony_promoted_workflow_action` | `pass` | `pass` | `taste_debt` | `none_observed` | `not_applicable` | source_audit_report preserved source authority and passed, but natural workflow-action use still required more commands or assistant turns than the low-ceremony UX threshold |

## Candidate Comparison Context

| Row | Scenario | Prompt specificity | Safety pass | Capability pass | UX quality | Interpretation |
| --- | --- | --- | --- | --- | --- | --- |
| Candidate A | `broad-contradiction-audit-scripted-control` | scripted-control | pass | pass | capability_only | Existing audit primitives prove capability but preserve old broad-action wording. |
| Candidate B | `broad-contradiction-audit-natural-intent` | natural-user-intent | pass | pass | taste_debt_due_action_name | Natural prompts still need source-sensitive framing instead of broad contradiction claims. |
| Candidate C | `source-audit-workflow-action-natural` | natural-user-intent | pass | pass | taste_debt | Selected narrow source-sensitive action plus existing primitives; residual command ceremony is tracked by `oc-nj5h`. |
