# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `web-url-intake-pressure`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `19.40`
- Harness elapsed seconds: `222.78`
- Effective parallel speedup: `0.82x`
- Parallel efficiency: `0.82`
- Targeted acceptance: web URL intake rows report missing path-hint handling, web create, duplicate URL rejection, no-op update, changed-source stale synthesis evidence, unsupported acquisition rejection, and final classification
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
| copy_repo | 0.19 |
| install_variant | 19.85 |
| warm_cache | 0.00 |
| seed_data | 0.05 |
| agent_run | 183.24 |
| parse_metrics | 0.00 |
| verify | 0.05 |
| total | 203.38 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `web-url-missing-path-hint` | `completed` | 0 | 0 | 1 | 2772 | 7.14 | `<run-root>/production/web-url-missing-path-hint/turn-1/events.jsonl` |
| `production` | `web-url-create` | `completed` | 6 | 6 | 4 | 7036 | 17.17 | `<run-root>/production/web-url-create/turn-1/events.jsonl` |
| `production` | `web-url-duplicate-normalized-url` | `completed` | 8 | 8 | 5 | 7664 | 17.97 | `<run-root>/production/web-url-duplicate-normalized-url/turn-1/events.jsonl` |
| `production` | `web-url-same-hash-noop` | `completed` | 14 | 14 | 5 | 9414 | 56.08 | `<run-root>/production/web-url-same-hash-noop/turn-1/events.jsonl` |
| `production` | `web-url-changed-stale` | `completed` | 56 | 56 | 11 | 8926 | 73.38 | `<run-root>/production/web-url-changed-stale/turn-1/events.jsonl` |
| `production` | `web-url-unsupported-acquisition` | `completed` | 4 | 4 | 3 | 6660 | 11.50 | `<run-root>/production/web-url-unsupported-acquisition/turn-1/events.jsonl` |
