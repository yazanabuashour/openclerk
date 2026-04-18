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
`--embedding-provider` flags are for tests or explicit user-directed datasets.

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

Retrieval task example:

```bash
printf '%s\n' '{"action":"search","search":{"text":"architecture","limit":10}}' |
  go run ./cmd/openclerk-agentops retrieval
```

Validation rejections are normal JSON results with `rejected: true` and
`rejection_reason`. Runtime failures exit non-zero and write the error to
stderr.

When reporting results, answer from JSON fields such as `document`,
`documents`, `search`, `links`, `graph`, `records`, `entity`, `provenance`,
`projections`, `paths`, or `rejection_reason`. Preserve citation paths, source
refs, and provenance details for source-sensitive claims.
