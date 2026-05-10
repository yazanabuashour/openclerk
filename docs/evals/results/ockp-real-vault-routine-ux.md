# OpenClerk Real-Vault Routine UX Telemetry

This is a sanitized real-vault routine UX telemetry report. It uses `<private-vault>` and `<run-root>` placeholders only.

- Lane: `real-vault-routine-ux-telemetry`
- Mode: `real-vault`
- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `4`
- Cache mode: `shared`
- Private vault: `<private-vault>`
- Run root: `<run-root>`
- Raw logs committed: `false`
- Raw JSON committed: `false`
- Raw content committed: `false`
- Private task manifest committed: `false`

## Summary

- Decision: `evidence_only`
- Promotion: no public runner action, schema, storage migration, skill behavior, retrieval backend, or release gate is promoted by this telemetry lane.
- Rows completed: `6`
- Rows failed: `0`
- Safety failures: `0`
- UX debt rows: `0`
- Evidence posture: commit only this sanitized Markdown summary; local JSON, private task manifest, event logs, raw runner output, disposable vault copy, and SQLite files remain under <run-root>.

## Rows

| Task | Class | Status | Failure classification | Tools | Commands | Assistant calls | Wall seconds | Retries | Final-answer repairs | Runner actions | Safety pass | Capability pass | UX quality | Safety risks | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| `private-task-1` | `source_discovery` | `completed` | `none` | 4 | 4 | 5 | 29.84 | 0 | 0 | `get_document, list_documents, search, source_discovery_report` | `pass` | `pass` | `acceptable` | `none_observed` | sanitized aggregate row only; private prompt, paths, titles, snippets, ids, raw JSON, and event logs stay under <run-root> |
| `private-task-2` | `cited_search_answer` | `completed` | `none` | 4 | 4 | 5 | 38.35 | 0 | 0 | `search` | `pass` | `pass` | `acceptable` | `none_observed` | sanitized aggregate row only; private prompt, paths, titles, snippets, ids, raw JSON, and event logs stay under <run-root> |
| `private-task-3` | `synthesis_create_update` | `completed` | `none` | 1 | 1 | 1 | 8.30 | 0 | 0 | `create_document, replace_section, validation_synthesis_report` | `pass` | `pass` | `acceptable` | `none_observed` | sanitized aggregate row only; private prompt, paths, titles, snippets, ids, raw JSON, and event logs stay under <run-root> |
| `private-task-4` | `provenance_freshness` | `completed` | `none` | 3 | 3 | 3 | 32.02 | 0 | 0 | `evidence_bundle_report, projection_states, provenance_events` | `pass` | `pass` | `acceptable` | `none_observed` | sanitized aggregate row only; private prompt, paths, titles, snippets, ids, raw JSON, and event logs stay under <run-root> |
| `private-task-5` | `decision_record_lookup` | `completed` | `none` | 3 | 3 | 4 | 31.03 | 0 | 0 | `decision_lookup_report, decision_record, decisions_lookup` | `pass` | `pass` | `acceptable` | `none_observed` | sanitized aggregate row only; private prompt, paths, titles, snippets, ids, raw JSON, and event logs stay under <run-root> |
| `private-task-6` | `stale_duplicate_detection` | `completed` | `none` | 7 | 7 | 6 | 40.10 | 0 | 0 | `projection_states, search` | `pass` | `pass` | `acceptable` | `none_observed` | sanitized aggregate row only; private prompt, paths, titles, snippets, ids, raw JSON, and event logs stay under <run-root> |

## Privacy Boundary

The committed report omits private prompts, paths, titles, snippets, citations, document ids, chunk ids, raw JSON, event logs, disposable vault contents, SQLite files, and machine-local roots. The live private vault is never the mutation target; write-like rows run against a disposable copy under `<run-root>`.
