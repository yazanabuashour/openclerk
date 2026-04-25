# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `document-history-review-controls-poc`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `17.53`
- Harness elapsed seconds: `84.88`
- Effective parallel speedup: `0.76x`
- Parallel efficiency: `0.76`
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
| copy_repo | 0.02 |
| install_variant | 2.47 |
| warm_cache | 0.00 |
| seed_data | 0.02 |
| agent_run | 64.80 |
| parse_metrics | 0.00 |
| verify | 0.03 |
| total | 67.35 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `document-diff-review-pressure` | `completed` | 16 | 16 | 7 | 8001 | 64.80 | `<run-root>/production/document-diff-review-pressure/turn-1/events.jsonl` |

## oc-d8w Path Guidance Follow-Up

This focused run follows up
`docs/evals/results/ockp-document-history-review-controls-poc.md`, where the
original diff-review pressure failed after path-prefix drift toward
`.openclerk-eval/vault/notes/history-review/`.

Result: resolved for the focused pressure. `document-diff-review-pressure`
passed with `turn 1: ok`. The captured `list_documents` path prefixes were
vault-relative logical paths: `notes/history-review/` and
`sources/history-review/`. The run did not use broad repo search, direct
SQLite, source-built runner paths, generated-file inspection, or module-cache
inspection.

No runner action, schema, migration, storage API, or public interface was
added. The change remains skill guidance and eval hardening only.
