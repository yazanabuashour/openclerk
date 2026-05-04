# OpenClerk Maturity Report

- Lane: `scale-ladder-validation`
- Mode: `scale-ladder`
- Tier: `100mb`
- Seed: `53`
- Harness: maintainer-only OpenClerk embedded runtime maturity harness; reduced reports only
- Run root: `<run-root>`
- Raw logs committed: `false`
- Raw content committed: `false`

## Corpus

| Metric | Value |
| --- | ---: |
| target_bytes | 104857600 |
| generated_bytes | 104975477 |
| sqlite_storage_bytes | 434479968 |
| documents | 800 |
| source_documents | 240 |
| synthesis_documents | 160 |
| decision_documents | 80 |
| duplicate_marked_documents | 80 |
| stale_marked_documents | 80 |
| tagged_documents | 800 |
| projection_state_sample_count | 100 |
| provenance_event_sample_count | 100 |

Counting policy: Counts are derived from runner-visible document summaries and metadata; reduced reports do not include document paths, titles, snippets, or private roots.

## Timings

| Probe | Seconds |
| --- | ---: |
| generate | 0.34 |
| import_sync | 19.38 |
| reopen_rebuild | 0.39 |
| list_latency | 0.00 |
| get_latency | 0.00 |
| projection_check | 0.00 |
| provenance_check | 0.01 |
| search_total | 0.28 |

## Sync Diagnostics

### `import_sync`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| last_phase | `completed` |
| paths_scanned | 800 |
| documents_created | 800 |
| documents_updated | 0 |
| documents_unchanged | 0 |
| documents_pruned | 0 |
| bytes_read | 104975477 |
| chunks_written | 1120 |
| fts_rows_written | 1120 |
| projection_bootstrap | false |
| projection_rebuild_skipped | false |
| scan_seconds | 0.00 |
| prune_seconds | 0.00 |
| document_read_parse_seconds | 0.00 |
| document_write_seconds | 16.67 |
| projection_rebuild_seconds | 1.43 |
| total_seconds | 19.37 |

| Projection | Seconds |
| --- | ---: |
| `graph` | 0.86 |
| `records` | 0.16 |
| `services` | 0.18 |
| `decisions` | 0.16 |
| `synthesis` | 0.07 |

### `reopen_sync`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| last_phase | `completed` |
| paths_scanned | 800 |
| documents_created | 0 |
| documents_updated | 0 |
| documents_unchanged | 800 |
| documents_pruned | 0 |
| bytes_read | 104975477 |
| chunks_written | 0 |
| fts_rows_written | 0 |
| projection_bootstrap | false |
| projection_rebuild_skipped | true |
| scan_seconds | 0.00 |
| prune_seconds | 0.00 |
| document_read_parse_seconds | 0.00 |
| document_write_seconds | 0.00 |
| projection_rebuild_seconds | 0.00 |
| total_seconds | 0.39 |

Projection rebuilds: none.

## Read Probes

| Probe | Query reference | Status | Results | Seconds | Evidence posture |
| --- | --- | --- | ---: | ---: | --- |
| `list-documents` | `` | `completed` | 50 | 0.00 | runner-visible summaries counted without emitting paths or titles |
| `get-document` | `` | `completed` | 1 | 0.00 | document body was read for timing only and is excluded from reduced reports |
| `fts-search` | `scale ladder authority marker seed 53` | `completed` | 10 | 0.16 | hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids |
| `fts-search` | `scale ladder synthesis freshness marker` | `completed` | 10 | 0.10 | hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids |
| `fts-search` | `scale ladder duplicate candidate marker` | `completed` | 10 | 0.02 | hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids |
| `projection-synthesis-sample` | `` | `completed` | 100 | 0.00 | projection freshness count only; reduced report excludes projection refs |
| `provenance-sample` | `` | `completed` | 100 | 0.01 | provenance event count only; reduced report excludes source refs and event ids |

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
