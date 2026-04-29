# OpenClerk Runner Concurrency Contract

OpenClerk runner commands are process-safe for the concurrency classes below.
The contract applies to the installed `openclerk` JSON runner surface, not to
direct SQLite, direct vault inspection, source-built runner paths, or other
unsupported transports.

## Supported Parallel Workflows

Agents may run these read-only workflows in parallel against the same configured
OpenClerk database:

- `openclerk document` `resolve_paths`
- `openclerk document` `list_documents`
- `openclerk document` `get_document`
- `openclerk document` `inspect_layout`
- `openclerk retrieval` `search`
- `openclerk retrieval` `document_links`
- `openclerk retrieval` `graph_neighborhood`
- `openclerk retrieval` `records_lookup`
- `openclerk retrieval` `record_entity`
- `openclerk retrieval` `services_lookup`
- `openclerk retrieval` `service_record`
- `openclerk retrieval` `decisions_lookup`
- `openclerk retrieval` `decision_record`
- `openclerk retrieval` `provenance_events`
- `openclerk retrieval` `projection_states`
- `openclerk retrieval` `audit_contradictions` with `audit.mode` set to
  `plan_only`

Parallel first use is supported. A fresh database may briefly initialize SQLite
schema, runtime configuration, and default vault paths, but supported parallel
read/startup commands must not expose raw SQLite, `runtime_config`, or
`upsert` failures as user-facing errors.

## Serialized Workflows

Mutating workflows remain single-writer. Agents must sequence these commands
unless a future contract explicitly expands write concurrency:

- `openclerk init`
- `openclerk document` `create_document`
- `openclerk document` `ingest_source_url`
- `openclerk document` `ingest_video_url`
- `openclerk document` `append_document`
- `openclerk document` `replace_section`
- `openclerk retrieval` `audit_contradictions` with `audit.mode` set to
  `repair_existing`

The runner serializes these operations with a database-adjacent write lock so
separate commands do not interleave vault file writes with SQLite registry,
provenance, or projection updates. Conflicting writes should return clear
runner errors such as `already_exists` or `conflict`; they should not leak raw
SQLite lock or schema internals.

## Implementation Notes

SQLite connections use a shared setup path with `busy_timeout`, foreign keys,
and WAL for initializing or mutating opens. Established read-only opens avoid
schema and runtime-config writes, while fresh read-only startup can still
initialize the minimum required runtime state.

Skill guidance may permit parallel runner use only for the supported read-only
workflows above. Write guidance remains conservative until targeted tests and
eval evidence prove a broader contract.
