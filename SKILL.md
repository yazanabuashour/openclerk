# OpenClerk Agent Guide

## When to use OpenClerk

Use OpenClerk when an agent needs a local-first knowledge plane for canonical notes, documents, promoted records, and provenance-backed retrieval without running a daemon or depending on an external service.

The preferred public surface is one embedded runtime plus one generated client package from the same module:

- [`client/local`](client/local) for the embedded runtime
- [`client/openclerk`](client/openclerk) for the generated request and response types

The legacy `fts`, `hybrid`, `graph`, and `records` packages remain available as implementation-variant fixtures for evals, not as the default product entrypoint.

## Install in a Go workspace

Use one tagged install command:

```bash
go get github.com/yazanabuashour/openclerk/client/local@v0.1.0
```

Import [`client/openclerk`](client/openclerk) from the same module for generated types. Do not add a second `go get` step for that package.

## Preferred runtime

Use [`client/local`](client/local) as the default runtime entrypoint and call `local.Open(...)`. It opens the SQLite-backed store in process and returns the generated OpenClerk client without binding a port.

Use [`client/openclerk`](client/openclerk) with `openclerk.NewClientWithResponses(baseURL)` only when you intentionally start [`cmd/openclerkd`](cmd/openclerkd) for HTTP debugging or compatibility work.

The OpenAPI contract in [`openapi/v1/openclerk.yaml`](openapi/v1/openclerk.yaml) remains the source of truth for operations, schemas, and generated request and response types.

## Default storage

Unless the caller overrides the paths in [`client/local.Config`](client/local/local.go), data is stored under:

```text
${XDG_DATA_HOME:-~/.local/share}/openclerk
```

The embedded runtime creates:

- `openclerk.sqlite`
- `vault/`

## Minimal flow

```go
client, runtime, err := local.Open(local.Config{})
if err != nil {
	return err
}
defer runtime.Close()

create, err := client.CreateDocumentWithResponse(ctx, openclerk.CreateDocumentRequest{
	Path:  "notes/architecture/knowledge-plane.md",
	Title: "Knowledge plane",
	Body:  "---\ntype: note\nstatus: active\n---\n# Knowledge plane\n\n## Summary\nCanonical architecture note.\n",
})
if err != nil {
	return err
}
if create.JSON201 == nil {
	return fmt.Errorf("create failed: %s", string(create.Body))
}

links, err := client.GetDocumentLinksWithResponse(ctx, create.JSON201.DocId)
if err != nil {
	return err
}
if links.JSON200 == nil {
	return fmt.Errorf("links failed: %s", string(links.Body))
}

events, err := client.ListProvenanceEventsWithResponse(ctx, nil)
if err != nil {
	return err
}
if events.JSON200 == nil {
	return fmt.Errorf("events failed: %s", string(events.Body))
}
```

## Practical defaults

- Use explicit `DataDir` overrides in tests and demos to avoid polluting the default XDG location.
- Use `v0.1.0` and later release tags for reproducible installs.
- Treat `cmd/openclerkd` as intentional debug and compatibility infrastructure, not the primary runtime path.
- Treat graph traversal and records lookup as derived capabilities over canonical sources, not separate truth systems.
