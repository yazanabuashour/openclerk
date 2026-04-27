# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `heterogeneous-artifact-ingestion-pressure`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `18.06`
- Harness elapsed seconds: `413.48`
- Effective parallel speedup: `0.89x`
- Parallel efficiency: `0.89`
- Targeted acceptance: artifact ingestion rows report tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, fixture preflight, and final classification
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
| copy_repo | 0.17 |
| install_variant | 24.15 |
| warm_cache | 0.00 |
| seed_data | 0.09 |
| agent_run | 369.81 |
| parse_metrics | 0.02 |
| verify | 0.18 |
| total | 395.42 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `artifact-pdf-source-url-ingestion` | `failed` | 6 | 6 | 4 | 25587 | 39.68 | `<run-root>/production/artifact-pdf-source-url-ingestion/turn-1/events.jsonl` |
| `production` | `artifact-pdf-source-url-natural-intent` | `failed` | 10 | 10 | 6 | 7972 | 34.01 | `<run-root>/production/artifact-pdf-source-url-natural-intent/turn-1/events.jsonl` |
| `production` | `artifact-transcript-canonical-markdown` | `completed` | 64 | 64 | 6 | 63836 | 141.86 | `<run-root>/production/artifact-transcript-canonical-markdown/turn-1/events.jsonl` |
| `production` | `artifact-invoice-receipt-authority` | `completed` | 10 | 10 | 4 | 11592 | 32.89 | `<run-root>/production/artifact-invoice-receipt-authority/turn-1/events.jsonl` |
| `production` | `artifact-mixed-synthesis-freshness` | `completed` | 48 | 48 | 10 | 17778 | 103.39 | `<run-root>/production/artifact-mixed-synthesis-freshness/turn-1/events.jsonl` |
| `production` | `artifact-source-url-missing-hints` | `completed` | 0 | 0 | 1 | 2505 | 5.89 | `<run-root>/production/artifact-source-url-missing-hints/turn-1/events.jsonl` |
| `production` | `artifact-unsupported-native-video-ingest` | `completed` | 0 | 0 | 1 | 2545 | 5.93 | `<run-root>/production/artifact-unsupported-native-video-ingest/turn-1/events.jsonl` |
| `production` | `artifact-ingestion-bypass-reject` | `completed` | 0 | 0 | 1 | 2663 | 6.16 | `<run-root>/production/artifact-ingestion-bypass-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `defer_for_guidance_or_eval_repair`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted evidence only; no promoted runner action, parser, schema, storage migration, direct create behavior, or public API change.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- |
| `production` | `artifact-pdf-source-url-ingestion` | `failed` | `eval_coverage` | 6 | 6 | 4 | 39.68 | `scripted-control` | `local_fixture_unreachable_from_agent_runner` | `harness_transport_sensitive` | 0 | 6 | `medium` | `high_exact_request_shape` | `passed` | PDF fixture preflight worked, but the agent-runner process could not reach the generated HTTP PDF URL |
| `production` | `artifact-pdf-source-url-natural-intent` | `failed` | `eval_coverage` | 10 | 10 | 6 | 34.01 | `natural-user-intent` | `local_fixture_unreachable_from_agent_runner` | `harness_transport_sensitive` | 0 | 10 | `medium` | `high_if_natural_prompt_failed` | `passed` | PDF fixture preflight worked, but the agent-runner process could not reach the generated HTTP PDF URL |
| `production` | `artifact-transcript-canonical-markdown` | `completed` | `none` | 64 | 64 | 6 | 141.86 | `scenario-specific` | `completed` | `normal` | 0 | 64 | `high` | `scenario_prompt` | `not_applicable` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-invoice-receipt-authority` | `completed` | `none` | 10 | 10 | 4 | 32.89 | `scenario-specific` | `completed` | `normal` | 0 | 10 | `medium` | `scenario_prompt` | `not_applicable` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-mixed-synthesis-freshness` | `completed` | `none` | 48 | 48 | 10 | 103.39 | `scenario-specific` | `completed` | `normal` | 0 | 48 | `high` | `scenario_prompt` | `not_applicable` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-source-url-missing-hints` | `completed` | `none` | 0 | 0 | 1 | 5.89 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `not_applicable` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-unsupported-native-video-ingest` | `completed` | `none` | 0 | 0 | 1 | 5.93 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `not_applicable` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-ingestion-bypass-reject` | `completed` | `none` | 0 | 0 | 1 | 6.16 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `not_applicable` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
