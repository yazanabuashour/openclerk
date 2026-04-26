# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `populated-vault-targeted`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `14.78`
- Harness elapsed seconds: `249.41`
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
| copy_repo | 0.06 |
| install_variant | 9.68 |
| warm_cache | 0.00 |
| seed_data | 5.88 |
| agent_run | 210.57 |
| parse_metrics | 0.00 |
| verify | 8.41 |
| total | 234.62 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `populated-heterogeneous-retrieval` | `failed` | 10 | 10 | 4 | 26574 | 25.18 | `<run-root>/production/populated-heterogeneous-retrieval/turn-1/events.jsonl` |
| `production` | `populated-freshness-conflict` | `completed` | 54 | 54 | 11 | 31257 | 105.62 | `<run-root>/production/populated-freshness-conflict/turn-1/events.jsonl` |
| `production` | `populated-synthesis-update-over-duplicate` | `completed` | 52 | 52 | 13 | 48631 | 79.77 | `<run-root>/production/populated-synthesis-update-over-duplicate/turn-1/events.jsonl` |

## Populated-Vault Targeted Decision

This targeted lane used the installed `openclerk document` and
`openclerk retrieval` JSON runners against the harness-generated synthetic
populated vault. It did not add runner actions, schemas, migrations, storage
APIs, product behavior, or public OpenClerk interfaces.

Selected scenario set:

- `populated-heterogeneous-retrieval`
- `populated-freshness-conflict`
- `populated-synthesis-update-over-duplicate`

| Scenario family | Result | Runner-visible evidence | Failure classification |
| --- | --- | --- | --- |
| Heterogeneous retrieval | fail | `search`, metadata-filtered `search`, `get_document`; no broad repo search, direct SQLite, source-built runner, generated-file inspection, or module-cache inspection | skill guidance / eval coverage |
| Freshness and conflict inspection | pass | `search`, `list_documents`, `get_document`, `projection_states`, `provenance_events`; conflict sources and synthesis remained unchanged | none |
| Synthesis update over duplicate | pass | `search`, `list_documents`, `get_document`, `projection_states`, `provenance_events`, and existing synthesis repair without duplicate synthesis creation | none |

The heterogeneous retrieval failure did not show a runner capability gap. The
database verification passed, the agent used the required runner-visible search
workflow including the `populated_role=authority` metadata filter, and the run
did not use any prohibited bypass path. Verification failed because the final
answer repeated polluted decoy claims after locating the populated authority
source, which is a skill guidance or eval-hardening issue around rejecting
polluted evidence in messy populated vaults.

Decision: keep the current `openclerk document` and `openclerk retrieval`
actions as the sufficient public surface for this targeted lane for now. Do not
promote a new product/API surface from this evidence. Any future promotion must
come from repeated targeted failures that show the existing runner actions are
structurally insufficient, not from a single assistant-answer handling failure.
