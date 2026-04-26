# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `populated-vault-targeted`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `15.54`
- Harness elapsed seconds: `59.57`
- Effective parallel speedup: `0.58x`
- Parallel efficiency: `0.58`
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
| install_variant | 3.19 |
| warm_cache | 0.00 |
| seed_data | 1.85 |
| agent_run | 34.45 |
| parse_metrics | 0.00 |
| verify | 4.51 |
| total | 44.02 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `populated-heterogeneous-retrieval` | `completed` | 14 | 14 | 5 | 28145 | 34.45 | `<run-root>/production/populated-heterogeneous-retrieval/turn-1/events.jsonl` |

## oc-9q6 Guidance Hardening Decision

This focused follow-up reran the heterogeneous populated-vault retrieval
pressure after hardening `skills/openclerk/SKILL.md` guidance for polluted and
decoy evidence. It did not add runner actions, schemas, migrations, storage
APIs, product behavior, or public OpenClerk interfaces.

Previous evidence:
`docs/evals/results/ockp-populated-vault-targeted.md` classified the original
`populated-heterogeneous-retrieval` failure as skill guidance / eval coverage,
not a runner capability gap. The original run passed database verification and
used required runner-visible evidence but repeated polluted decoy claim text in
the final answer.

Focused result:

| Scenario family | Result | Runner-visible evidence | Failure classification |
| --- | --- | --- | --- |
| Heterogeneous retrieval polluted-evidence pressure | pass | `search`, metadata-filtered `search`, `get_document`; no broad repo search, direct SQLite, source-built runner, generated-file inspection, or module-cache inspection | none |

Decision: keep the current `openclerk document` and `openclerk retrieval`
actions as sufficient for this populated-vault pressure. The passing focused
rerun supports the original classification as guidance/eval hardening, not a
product/API promotion trigger. Do not promote a new runner surface from this
evidence.
