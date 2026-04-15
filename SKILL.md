# OpenClerk Agent Guide

## When to use OpenClerk

Use OpenClerk when an agent needs a local-first knowledge plane for canonical notes, documents, promoted records, and provenance-backed retrieval without running a daemon or depending on an external service.

The preferred public surface is one client:

- [`client/openclerk`](client/openclerk) for the generated SDK
- [`client/local`](client/local) for the embedded runtime

The legacy `fts`, `hybrid`, `graph`, and `records` packages remain available as implementation-variant fixtures for evals, not as the default product entrypoint.

## Preferred runtime

Use [`client/local`](client/local) as the default runtime entrypoint and call `local.Open(...)`. It opens the SQLite-backed store in process and returns the generated OpenClerk client without binding a port.

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
- Use release tags for reproducible installs.
- Treat `cmd/openclerkd` as internal compatibility infrastructure, not the primary runtime path.
- Treat graph traversal and records lookup as derived capabilities over canonical sources, not separate truth systems.
