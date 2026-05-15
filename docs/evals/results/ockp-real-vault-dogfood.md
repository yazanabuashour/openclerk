# OpenClerk Maturity Report

- Lane: `representative-real-vault-dogfood`
- Mode: `real-vault`
- Harness: maintainer-only OpenClerk embedded runtime maturity harness; reduced reports only
- Run root: `<run-root>`
- Private vault: `<private-vault>`
- Raw logs committed: `false`
- Raw content committed: `false`

## Corpus

| Metric | Value |
| --- | ---: |
| sqlite_storage_bytes | 8245376 |
| documents | 133 |
| source_documents | 13 |
| synthesis_documents | 4 |
| decision_documents | 2 |
| duplicate_marked_documents | 0 |
| stale_marked_documents | 5 |
| tagged_documents | 126 |
| projection_state_sample_count | 3 |
| provenance_event_sample_count | 100 |

Counting policy: Counts are derived from runner-visible document summaries and metadata; reduced reports do not include document paths, titles, snippets, or private roots.

## Timings

| Probe | Seconds |
| --- | ---: |
| generate | 0.00 |
| import_sync | 0.19 |
| reopen_rebuild | 0.02 |
| list_latency | 0.00 |
| get_latency | 0.00 |
| projection_check | 0.00 |
| provenance_check | 0.00 |
| search_total | 0.29 |

## Sync Diagnostics

### `import_sync`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| last_phase | `completed` |
| paths_scanned | 133 |
| documents_created | 133 |
| documents_updated | 0 |
| documents_unchanged | 0 |
| documents_pruned | 0 |
| bytes_read | 831973 |
| chunks_written | 1004 |
| fts_rows_written | 1004 |
| fts_strategy | `bulk_rebuild` |
| fts_bootstrap | true |
| fts_rebuild_pending | false |
| fts_rebuild_skipped | false |
| projection_bootstrap | false |
| projection_rebuild_skipped | false |
| scan_seconds | 0.01 |
| prune_seconds | 0.00 |
| document_read_parse_seconds | 0.04 |
| document_write_seconds | 0.04 |
| document_record_write_seconds | 0.02 |
| chunk_write_seconds | 0.01 |
| provenance_write_seconds | 0.01 |
| incremental_fts_write_seconds | 0.00 |
| bulk_fts_rebuild_seconds | 0.03 |
| projection_rebuild_seconds | 0.06 |
| total_seconds | 0.18 |

| Projection | Seconds |
| --- | ---: |
| `graph` | 0.04 |
| `records` | 0.01 |
| `services` | 0.01 |
| `decisions` | 0.01 |
| `synthesis` | 0.00 |

### `reopen_sync`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| last_phase | `completed` |
| paths_scanned | 133 |
| documents_created | 0 |
| documents_updated | 0 |
| documents_unchanged | 133 |
| documents_pruned | 0 |
| bytes_read | 831973 |
| chunks_written | 0 |
| fts_rows_written | 0 |
| fts_strategy | `skipped_no_changes` |
| fts_bootstrap | false |
| fts_rebuild_pending | false |
| fts_rebuild_skipped | true |
| projection_bootstrap | false |
| projection_rebuild_skipped | true |
| scan_seconds | 0.00 |
| prune_seconds | 0.00 |
| document_read_parse_seconds | 0.01 |
| document_write_seconds | 0.00 |
| document_record_write_seconds | 0.00 |
| chunk_write_seconds | 0.00 |
| provenance_write_seconds | 0.00 |
| incremental_fts_write_seconds | 0.00 |
| bulk_fts_rebuild_seconds | 0.00 |
| projection_rebuild_seconds | 0.00 |
| total_seconds | 0.02 |

Projection rebuilds: none.

## Read Probes

| Probe | Query reference | Status | Results | Seconds | Evidence posture |
| --- | --- | --- | ---: | ---: | --- |
| `list-documents` | `` | `completed` | 50 | 0.00 | runner-visible summaries counted without emitting paths or titles |
| `get-document` | `` | `completed` | 1 | 0.00 | document body was read for timing only and is excluded from reduced reports |
| `fts-search` | `private-query-1` | `completed` | 10 | 0.00 | hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids |
| `fts-search` | `private-query-2` | `completed` | 10 | 0.15 | hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids |
| `fts-search` | `private-query-3` | `completed` | 10 | 0.14 | hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids |
| `projection-synthesis-sample` | `` | `completed` | 3 | 0.00 | projection freshness count only; reduced report excludes projection refs |
| `provenance-sample` | `` | `completed` | 100 | 0.00 | provenance event count only; reduced report excludes source refs and event ids |

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
