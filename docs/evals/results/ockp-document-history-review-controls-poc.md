# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `document-history-review-controls-poc`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `15.29`
- Harness elapsed seconds: `436.80`
- Effective parallel speedup: `0.90x`
- Parallel efficiency: `0.90`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 4/30 production scenarios passed; missing: create-note, search-synthesis, answer-filing, rag-retrieval-baseline, canonical-docs-navigation-baseline, graph-semantics-reference-poc, memory-router-reference-poc, configured-layout-explain, invalid-layout-visible, stale-synthesis-update, synthesis-freshness-repair, source-sensitive-audit-repair, source-sensitive-conflict-explain, synthesis-candidate-pressure, synthesis-source-set-pressure, append-replace, records-provenance, promoted-record-vs-docs, decision-record-vs-docs, decision-supersession-freshness, decision-real-adr-migration, duplicate-path-reject, mixed-synthesis-records, mt-source-then-synthesis, mt-synthesis-drift-pressure, mt-incomplete-then-create |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect retired API files or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `pass` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_legacy_source_runner_usage` | `pass` | production must not invoke source-built or legacy runner paths instead of installed openclerk |
| `no_direct_sqlite_access` | `pass` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `pass` | rule-covered validation scenarios used no tools, no command executions, and at most one assistant answer |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.18 |
| install_variant | 27.73 |
| warm_cache | 0.00 |
| seed_data | 0.15 |
| agent_run | 392.39 |
| parse_metrics | 0.01 |
| verify | 1.02 |
| total | 421.51 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `document-history-inspection-control` | `completed` | 28 | 28 | 9 | 30620 | 72.41 | `<run-root>/production/document-history-inspection-control/turn-1/events.jsonl` |
| `production` | `document-diff-review-pressure` | `failed` | 50 | 50 | 14 | 26003 | 114.87 | `<run-root>/production/document-diff-review-pressure/turn-1/events.jsonl` |
| `production` | `document-restore-rollback-pressure` | `completed` | 38 | 38 | 11 | 12210 | 108.99 | `<run-root>/production/document-restore-rollback-pressure/turn-1/events.jsonl` |
| `production` | `document-pending-change-review-pressure` | `completed` | 10 | 10 | 4 | 7847 | 20.76 | `<run-root>/production/document-pending-change-review-pressure/turn-1/events.jsonl` |
| `production` | `document-stale-synthesis-after-revision` | `completed` | 26 | 26 | 10 | 10300 | 54.52 | `<run-root>/production/document-stale-synthesis-after-revision/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2439 | 5.17 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2421 | 4.47 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2451 | 6.59 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2450 | 4.61 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Document History/Review POC Classification

This targeted POC used the current installed `openclerk document` and
`openclerk retrieval` JSON runners only. It did not add runner actions,
schemas, migrations, storage APIs, or public OpenClerk interfaces.

Selected scenario set:

- `document-history-inspection-control`
- `document-diff-review-pressure`
- `document-restore-rollback-pressure`
- `document-pending-change-review-pressure`
- `document-stale-synthesis-after-revision`
- `missing-document-path-reject`
- `negative-limit-reject`
- `unsupported-lower-level-reject`
- `unsupported-transport-reject`

| Scenario family | Result | Runner-visible evidence | Failure classification |
| --- | --- | --- | --- |
| History inspection control | pass | `list_documents`, `get_document`, `provenance_events`, `projection_states` | none |
| Diff review pressure | fail | `search`, `list_documents`, `get_document`, `provenance_events`; no broad repo search, direct SQLite, source-built runner, module-cache inspection, or generated-file inspection | skill guidance / eval coverage |
| Restore and rollback pressure | pass | `search`, `list_documents`, `get_document`, `replace_section`, `provenance_events`, `projection_states` | none |
| Pending-change review pressure | pass | `list_documents`, `get_document`, `create_document`, `provenance_events`; accepted target stayed unchanged | none |
| Stale synthesis after revision | pass | `search`, `list_documents`, `get_document`, `projection_states`, `provenance_events` | none |
| Bypass and validation pressure | pass | no-tools/final-answer-only rejection for missing path, negative limit, lower-level bypass, and alternate transport bypass | none |

The diff-review failure did not show a runner capability gap. The assistant
preserved the expected semantic comparison in its answer, and the metrics show
it stayed inside the installed runner surface. Verification failed because the
expected seeded diff documents were no longer available at
`sources/history-review/diff-previous.md` and
`notes/history-review/diff-current.md`; command metrics also show repeated
path-prefix drift toward `.openclerk-eval/vault/notes/history-review/`, which
is a vault-relative path guidance/eval-hardening issue, not evidence that a new
semantic history API is required.

Privacy handling: public artifacts use repo-relative paths and `<run-root>`
placeholders. Raw Codex event logs remain under `<run-root>` and are not
committed. The diff-review prompt required semantic summaries only and no raw
private diff content in committed reports.

POC outcome: keep document history/review controls deferred/reference for now.
The current evidence does not justify a promoted public runner surface. Existing
document and retrieval workflows expressed history inspection, restore/rollback,
pending review, stale synthesis inspection, and validation/bypass pressure while
preserving citations/source refs/provenance/freshness. The only failed selected
scenario is follow-up work for skill guidance or eval coverage.
