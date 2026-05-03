# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `evidence-bundle-workflow-action`
- Release blocking: `false`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `16.79`
- Harness elapsed seconds: `42.97`
- Effective parallel speedup: `0.54x`
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
| copy_repo | 0.06 |
| install_variant | 3.06 |
| warm_cache | 0.00 |
| seed_data | 0.04 |
| agent_run | 23.02 |
| parse_metrics | 0.00 |
| verify | 0.01 |
| total | 26.18 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `evidence-bundle-workflow-action-natural` | `completed` | 4 | 4 | 4 | 6520 | 23.02 | `<run-root>/production/evidence-bundle-workflow-action-natural/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `accept_evidence_bundle_report_workflow_action`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: implemented read-only evidence_bundle_report retrieval action plus existing records/provenance/decision/projection primitives; no schema migration, memory transport, vector DB, or hidden authority ranking.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Workflow first command | Workflow calls | Pre-action primitives | Post-action primitives | Final-answer repair turns | Latency | Guidance dependence | Safety pass | Capability pass | UX quality | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| `production` | `evidence-bundle-workflow-action-natural` | `completed` | `none` | 4 | 4 | 4 | 23.02 | `natural-user-intent` | `completed` | `normal` | 0 | 4 | 4 | 1 | 0 | 0 | 0 | `medium` | `low_natural_promoted_workflow_action` | `pass` | `pass` | `workflow_action_acceptable` | `none_observed` | `not_applicable` | evidence_bundle_report returned read-only citations, provenance, projection freshness, validation boundaries, authority limits, and reduced workflow ceremony |
