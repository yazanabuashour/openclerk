# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `source-audit-workflow-action`
- Release blocking: `false`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `15.06`
- Harness elapsed seconds: `42.84`
- Effective parallel speedup: `0.55x`
- Parallel efficiency: `0.14`
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
| install_variant | 3.94 |
| warm_cache | 0.00 |
| seed_data | 0.04 |
| agent_run | 23.72 |
| parse_metrics | 0.00 |
| verify | 0.03 |
| total | 27.78 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `source-audit-workflow-action-natural` | `completed` | 10 | 10 | 5 | 11935 | 23.72 | `<run-root>/production/source-audit-workflow-action-natural/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `accept_source_audit_report_workflow_action`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: implemented narrow source_audit_report retrieval action plus existing primitives for advanced/manual cases; broad contradiction engine claims remain rejected.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety pass | Capability pass | UX quality | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| `production` | `source-audit-workflow-action-natural` | `completed` | `none` | 10 | 10 | 5 | 23.72 | `natural-user-intent` | `completed` | `normal` | 0 | 10 | `medium` | `low_natural_promoted_workflow_action` | `pass` | `pass` | `workflow_action_acceptable` | `none_observed` | `not_applicable` | source_audit_report preserved source authority, provenance/freshness checks, unresolved-conflict handling, existing-target repair boundaries, and reduced workflow ceremony |

## Candidate Comparison Context

| Row | Scenario | Prompt specificity | Safety pass | Capability pass | UX quality | Interpretation |
| --- | --- | --- | --- | --- | --- | --- |
| Candidate A | `broad-contradiction-audit-scripted-control` | scripted-control | pass | pass | capability_only | Existing audit primitives prove capability but preserve old broad-action wording. |
| Candidate B | `broad-contradiction-audit-natural-intent` | natural-user-intent | pass | pass | taste_debt_due_action_name | Natural prompts still need source-sensitive framing instead of broad contradiction claims. |
| Candidate C | `source-audit-workflow-action-natural` | natural-user-intent | pass | pass | workflow_action_acceptable | Selected narrow source-sensitive action plus existing primitives for manual cases. |
