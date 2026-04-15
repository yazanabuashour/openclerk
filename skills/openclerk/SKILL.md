---
name: openclerk
description: Use OpenClerk when an agent needs a local-first knowledge plane for canonical notes, documents, promoted records, and provenance-backed retrieval in a Go workspace. Prefer this skill when working with the embedded runtime, generated OpenClerk client, local SQLite-backed storage, or HTTP compatibility and debugging flows.
license: MIT
---

# OpenClerk

## When to use this skill

Use OpenClerk when the task needs a local-first knowledge plane without running a daemon or depending on an external service.

Prefer this skill when the work involves:

- embedding OpenClerk in a Go process
- using generated request and response types from the same module
- storing canonical notes, documents, or promoted records in the local SQLite-backed store
- debugging or compatibility-testing the HTTP surface intentionally

The default public surface is one embedded runtime plus one generated client package:

- [`../../client/local`](../../client/local) for the embedded runtime
- [`../../client/openclerk`](../../client/openclerk) for generated request and response types

The legacy `fts`, `hybrid`, `graph`, and `records` packages remain available as implementation-variant fixtures for evals, not as the default product entrypoint.

## Runtime guidance

Use [`../../client/local`](../../client/local) as the default runtime entrypoint and call `local.Open(...)`. It opens the SQLite-backed store in process and returns the generated OpenClerk client without binding a port.

Use [`../../client/openclerk`](../../client/openclerk) with `openclerk.NewClientWithResponses(baseURL)` only when you intentionally start [`../../cmd/openclerkd`](../../cmd/openclerkd) for HTTP debugging or compatibility work.

Treat [`../../openapi/v1/openclerk.yaml`](../../openapi/v1/openclerk.yaml) as the source of truth for operations, schemas, and generated request and response types.

## Storage defaults

Unless the caller overrides paths in [`../../client/local/local.go`](../../client/local/local.go), OpenClerk stores data under `${XDG_DATA_HOME:-~/.local/share}/openclerk` and creates:

- `openclerk.sqlite`
- `vault/`

Use explicit `DataDir` overrides in tests and demos to avoid polluting the default XDG location.

## Practical defaults

- Use tagged installs such as `v0.1.0` and later for reproducible setups.
- Treat [`../../cmd/openclerkd`](../../cmd/openclerkd) as intentional debug infrastructure, not the primary runtime path.
- Treat graph traversal and records lookup as derived capabilities over canonical sources, not separate truth systems.

For the tagged install command and a minimal end-to-end Go example, read [the reference guide](references/openclerk.md).
