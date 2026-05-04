# OpenClerk Maturity Report

- Lane: `scale-ladder-validation`
- Mode: `scale-ladder`
- Tier: `1gb`
- Seed: `53`
- Harness: maintainer-only OpenClerk embedded runtime maturity harness; reduced reports only
- Run root: `<run-root>`
- Raw logs committed: `false`
- Raw content committed: `false`

## Corpus

| Metric | Value |
| --- | ---: |
| target_bytes | 1073741824 |
| generated_bytes | 1073813879 |
| sqlite_storage_bytes | 4709521184 |
| documents | 8183 |
| source_documents | 2455 |
| synthesis_documents | 1637 |
| decision_documents | 819 |
| duplicate_marked_documents | 818 |
| stale_marked_documents | 818 |
| tagged_documents | 8183 |
| projection_state_sample_count | 100 |
| provenance_event_sample_count | 100 |

Counting policy: Counts are derived from runner-visible document summaries and metadata; reduced reports do not include document paths, titles, snippets, or private roots.

## Timings

| Probe | Seconds |
| --- | ---: |
| generate | 3.88 |
| import_sync | 68.06 |
| reopen_rebuild | 5.97 |
| list_latency | 0.00 |
| get_latency | 0.00 |
| projection_check | 0.00 |
| provenance_check | 0.13 |
| search_total | 4.40 |

## Sync Diagnostics

### `import_sync`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| last_phase | `completed` |
| paths_scanned | 8183 |
| documents_created | 8183 |
| documents_updated | 0 |
| documents_unchanged | 0 |
| documents_pruned | 0 |
| bytes_read | 1073813879 |
| chunks_written | 11457 |
| fts_rows_written | 11457 |
| fts_strategy | `bulk_rebuild` |
| fts_bootstrap | true |
| fts_rebuild_pending | false |
| fts_rebuild_skipped | false |
| projection_bootstrap | false |
| projection_rebuild_skipped | false |
| scan_seconds | 0.01 |
| prune_seconds | 0.00 |
| document_read_parse_seconds | 4.05 |
| document_write_seconds | 8.26 |
| document_record_write_seconds | 3.04 |
| chunk_write_seconds | 2.42 |
| provenance_write_seconds | 2.80 |
| incremental_fts_write_seconds | 0.00 |
| bulk_fts_rebuild_seconds | 17.61 |
| projection_rebuild_seconds | 31.70 |
| total_seconds | 68.05 |

| Projection | Seconds |
| --- | ---: |
| `graph` | 10.62 |
| `records` | 9.54 |
| `services` | 4.99 |
| `decisions` | 5.45 |
| `synthesis` | 1.10 |

### `reopen_sync`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| last_phase | `completed` |
| paths_scanned | 8183 |
| documents_created | 0 |
| documents_updated | 0 |
| documents_unchanged | 8183 |
| documents_pruned | 0 |
| bytes_read | 1073813879 |
| chunks_written | 0 |
| fts_rows_written | 0 |
| fts_strategy | `skipped_no_changes` |
| fts_bootstrap | false |
| fts_rebuild_pending | false |
| fts_rebuild_skipped | true |
| projection_bootstrap | false |
| projection_rebuild_skipped | true |
| scan_seconds | 0.03 |
| prune_seconds | 0.01 |
| document_read_parse_seconds | 3.54 |
| document_write_seconds | 0.00 |
| document_record_write_seconds | 0.00 |
| chunk_write_seconds | 0.00 |
| provenance_write_seconds | 0.00 |
| incremental_fts_write_seconds | 0.00 |
| bulk_fts_rebuild_seconds | 0.00 |
| projection_rebuild_seconds | 0.00 |
| total_seconds | 5.97 |

Projection rebuilds: none.

## Read Probes

| Probe | Query reference | Status | Results | Seconds | Evidence posture |
| --- | --- | --- | ---: | ---: | --- |
| `list-documents` | `` | `completed` | 50 | 0.00 | runner-visible summaries counted without emitting paths or titles |
| `get-document` | `` | `completed` | 1 | 0.00 | document body was read for timing only and is excluded from reduced reports |
| `fts-search` | `scale ladder authority marker seed 53` | `completed` | 10 | 3.31 | hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids |
| `fts-search` | `scale ladder synthesis freshness marker` | `completed` | 10 | 0.96 | hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids |
| `fts-search` | `scale ladder duplicate candidate marker` | `completed` | 10 | 0.13 | hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids |
| `projection-synthesis-sample` | `` | `completed` | 100 | 0.00 | projection freshness count only; reduced report excludes projection refs |
| `provenance-sample` | `` | `completed` | 100 | 0.13 | provenance event count only; reduced report excludes source refs and event ids |

## Checks

| Check | Value |
| --- | --- |
| reduced_report_only | `true` |
| raw_logs_committed | `false` |
| raw_content_committed | `false` |
| machine_absolute_artifact_refs | `false` |
| routine_agent_bypass_events_available | `false` |
| boundary | This harness validates local runtime behavior and reduced-report hygiene; routine-agent bypass checks require an agent eval row with event logs. |

## Outcomes

| Name | Status | Safety pass | Capability pass | UX quality | Performance | Evidence posture | Details |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `reduced-report-boundary` | `completed` | `pass` | `pass` | `not_agent_ux_evidence` | `not_applicable` | repo-relative or neutral artifact references only; raw content and raw logs are not committed | Report intentionally excludes document paths, titles, snippets, private vault roots, and machine-absolute run roots. |
| `runtime-maturity-readiness` | `completed` | `pass` | `pass` | `not_agent_ux_evidence` | `recorded` | SQLite FTS, list/get, projection, and provenance probes executed through embedded OpenClerk runtime APIs. | Use these numbers as decision input only; promotion decisions still need safety, capability, UX, performance, and evidence posture recorded separately. |
