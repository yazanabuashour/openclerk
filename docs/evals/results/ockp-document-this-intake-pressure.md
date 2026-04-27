# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `document-this-intake-pressure`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `18.56`
- Harness elapsed seconds: `301.08`
- Effective parallel speedup: `0.85x`
- Parallel efficiency: `0.85`
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
| install_variant | 24.84 |
| warm_cache | 0.00 |
| seed_data | 0.12 |
| agent_run | 257.26 |
| parse_metrics | 0.01 |
| verify | 0.15 |
| total | 282.53 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `document-this-missing-fields` | `completed` | 0 | 0 | 1 | 2416 | 4.82 | `<run-root>/production/document-this-missing-fields/turn-1/events.jsonl` |
| `production` | `document-this-explicit-create` | `completed` | 4 | 4 | 3 | 3311 | 15.04 | `<run-root>/production/document-this-explicit-create/turn-1/events.jsonl` |
| `production` | `document-this-source-url-missing-hints` | `completed` | 0 | 0 | 1 | 8250 | 5.47 | `<run-root>/production/document-this-source-url-missing-hints/turn-1/events.jsonl` |
| `production` | `document-this-explicit-overrides` | `completed` | 4 | 4 | 3 | 24909 | 18.88 | `<run-root>/production/document-this-explicit-overrides/turn-1/events.jsonl` |
| `production` | `document-this-duplicate-candidate` | `completed` | 8 | 8 | 5 | 8407 | 22.61 | `<run-root>/production/document-this-duplicate-candidate/turn-1/events.jsonl` |
| `production` | `document-this-existing-update` | `completed` | 30 | 30 | 7 | 12908 | 70.58 | `<run-root>/production/document-this-existing-update/turn-1/events.jsonl` |
| `production` | `document-this-synthesis-freshness` | `completed` | 50 | 50 | 9 | 34012 | 119.86 | `<run-root>/production/document-this-synthesis-freshness/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `evaluate_for_oc_99z`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: no promoted runner action, schema, migration, skill behavior, storage API, product behavior, or public OpenClerk interface from this eval.

| Variant | Scenario | Status | Failure classification | Evidence posture |
| --- | --- | --- | --- | --- |
| `production` | `document-this-missing-fields` | `completed` | `none` | current document/retrieval runner behavior handled document-this intake pressure |
| `production` | `document-this-explicit-create` | `completed` | `none` | current document/retrieval runner behavior handled document-this intake pressure |
| `production` | `document-this-source-url-missing-hints` | `completed` | `none` | current document/retrieval runner behavior handled document-this intake pressure |
| `production` | `document-this-explicit-overrides` | `completed` | `none` | current document/retrieval runner behavior handled document-this intake pressure |
| `production` | `document-this-duplicate-candidate` | `completed` | `none` | current document/retrieval runner behavior handled document-this intake pressure |
| `production` | `document-this-existing-update` | `completed` | `none` | current document/retrieval runner behavior handled document-this intake pressure |
| `production` | `document-this-synthesis-freshness` | `completed` | `none` | current document/retrieval runner behavior handled document-this intake pressure |
