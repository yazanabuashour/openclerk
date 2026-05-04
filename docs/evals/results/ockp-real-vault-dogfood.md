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
| sqlite_storage_bytes | 38966392 |
| documents | 80 |
| source_documents | 24 |
| synthesis_documents | 16 |
| decision_documents | 8 |
| duplicate_marked_documents | 8 |
| stale_marked_documents | 8 |
| tagged_documents | 80 |
| projection_state_sample_count | 16 |
| provenance_event_sample_count | 100 |

Counting policy: Counts are derived from runner-visible document summaries and metadata; reduced reports do not include document paths, titles, snippets, or private roots.

## Timings

| Probe | Seconds |
| --- | ---: |
| generate | 0.00 |
| import_sync | 5.90 |
| reopen_rebuild | 11.05 |
| list_latency | 0.00 |
| get_latency | 0.00 |
| projection_check | 0.00 |
| provenance_check | 0.00 |
| search_total | 0.04 |

## Read Probes

| Probe | Query reference | Status | Results | Seconds | Evidence posture |
| --- | --- | --- | ---: | ---: | --- |
| `list-documents` | `` | `completed` | 50 | 0.00 | runner-visible summaries counted without emitting paths or titles |
| `get-document` | `` | `completed` | 1 | 0.00 | document body was read for timing only and is excluded from reduced reports |
| `fts-search` | `private-query-1` | `completed` | 10 | 0.02 | hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids |
| `fts-search` | `private-query-2` | `completed` | 10 | 0.02 | hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids |
| `fts-search` | `private-query-3` | `completed` | 8 | 0.00 | hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids |
| `projection-synthesis-sample` | `` | `completed` | 16 | 0.00 | projection freshness count only; reduced report excludes projection refs |
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
