---
name: openclerk
description: Use OpenClerk when an agent needs a local-first knowledge plane for canonical notes, source-linked synthesis, promoted records, and provenance-backed retrieval through the AgentOps JSON runner.
license: MIT
compatibility: Requires a Go-capable environment with local filesystem access and the openclerk repository checkout.
---

# OpenClerk AgentOps

Use this skill for routine local OpenClerk knowledge-plane tasks. The production
agent interface is the JSON runner in `cmd/openclerk-agentops`, backed by the
`agentops` package. Supported routine tasks are:

- document tasks: validate, create, list, get, append, replace-section, and resolve paths; see [references/documents.md](references/documents.md)
- retrieval tasks: search, document links, graph neighborhoods, records lookup/entity reads, provenance events, and projection states; see [references/search.md](references/search.md) and [references/records-provenance.md](references/records-provenance.md)
- source-linked synthesis workflows composed from document and retrieval tasks; see [references/openclerk.md](references/openclerk.md)

For supported tasks, run `go run ./cmd/openclerk-agentops document` or
`go run ./cmd/openclerk-agentops retrieval`, pass exactly one JSON request on
stdin, read the JSON result from stdout, and answer only from that JSON. Use
`local.Config{}` defaults unless the user names a specific dataset. The runner
honors `OPENCLERK_DATA_DIR`, `OPENCLERK_DATABASE_PATH`, and
`OPENCLERK_VAULT_ROOT`; optional `--data-dir`, `--db`, `--vault-root`, and
`--embedding-provider` flags are for tests or explicit user-directed datasets
only. For routine requests, do not pass those flags; rely on the configured
environment so data, database, and vault paths stay together. Do not inspect the
repo to rediscover runner schemas; use the documented request shapes directly.

Before using any runner, reject final-answer-only, with exactly one assistant
answer and no tools, when the request is missing required document or retrieval
fields, asks for an obviously invalid limit such as a negative number, or asks
to bypass AgentOps for routine lower-level SDK, HTTP, SQLite, or
generated-client work, human CLI work, or unevaluated MCP-style work. Do not
first announce skill use or process for those direct rejections.

Do not inspect generated clients, backend-variant packages, generated server
code, the Go module cache, or SQLite directly for routine OpenClerk tasks. Do
not run broad repo searches, `bd prime`, or maintainer setup before acting on a
direct user request to read or write local OpenClerk knowledge. Search the repo
only if the AgentOps runner fails in a way that requires debugging the checkout.

## Runner Pattern

Document task example:

```bash
printf '%s\n' '{"action":"list_documents","list":{"path_prefix":"notes/","limit":20}}' |
  go run ./cmd/openclerk-agentops document
```

Common one-line document shapes:

```bash
{"action":"create_document","document":{"path":"notes/projects/example.md","title":"Example","body":"# Example\n\n## Summary\nReusable knowledge.\n"}}
{"action":"list_documents","list":{"path_prefix":"notes/","limit":20}}
{"action":"get_document","doc_id":"doc_id_from_json"}
{"action":"append_document","doc_id":"doc_id_from_json","content":"## Decisions\nUse the AgentOps runner."}
{"action":"replace_section","doc_id":"doc_id_from_json","heading":"Decisions","content":"Use the AgentOps runner."}
```

Retrieval task example:

```bash
printf '%s\n' '{"action":"search","search":{"text":"architecture","limit":10}}' |
  go run ./cmd/openclerk-agentops retrieval
```

Common one-line retrieval shapes:

```bash
{"action":"search","search":{"text":"architecture","limit":10}}
{"action":"document_links","doc_id":"doc_id_from_json"}
{"action":"records_lookup","records":{"text":"OpenClerk AgentOps","limit":10}}
{"action":"provenance_events","provenance":{"ref_kind":"document","ref_id":"doc_id_from_json","limit":20}}
{"action":"projection_states","projection":{"ref_kind":"document","ref_id":"doc_id_from_json","limit":20}}
```

Validation rejections are normal JSON results with `rejected: true` and
`rejection_reason`. Runtime failures exit non-zero and write the error to
stderr.

When reporting results, answer from JSON fields such as `document`,
`documents`, `search`, `links`, `graph`, `records`, `entity`, `provenance`,
`projections`, `paths`, or `rejection_reason`. Preserve citation paths, source
refs, and provenance details for source-sensitive claims.
