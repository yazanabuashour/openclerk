# OpenClerk Reference

The production agent path is the ergonomic local SDK facade in
`github.com/yazanabuashour/openclerk/client/local`.

## Agent Quick Start

- Use `local.OpenClient(local.Config{})` for live local data. It opens the
  default SQLite database and vault, syncs canonical Markdown, and calls the
  in-process service directly.
- Use `local.ResolvePaths(local.Config{})` only when you need to report or verify
  the default `DataDir`, `DatabasePath`, and `VaultRoot`.
- Use explicit `local.Config{DataDir: "..."}`, `local.Config{DatabasePath: "..."}`,
  or `local.Config{VaultRoot: "..."}` for tests, fixtures, and throwaway
  examples.
- Use generated OpenAPI methods only for endpoints not covered by the SDK facade
  or when the user explicitly needs raw API-contract behavior.

## Install In A Go Workspace

```bash
go get github.com/yazanabuashour/openclerk/client/local@main
```

## Minimal Flow

```go
client, err := local.OpenClient(local.Config{})
if err != nil {
	return err
}
defer client.Close()

document, err := client.CreateDocument(ctx, local.DocumentInput{
	Path:  "notes/architecture/knowledge-plane.md",
	Title: "Knowledge plane",
	Body:  "---\ntype: note\nstatus: active\n---\n# Knowledge plane\n\n## Summary\nCanonical architecture note.\n",
})
if err != nil {
	return err
}

results, err := client.Search(ctx, local.SearchOptions{
	Text:  "architecture",
	Limit: 10,
})
if err != nil {
	return err
}

events, err := client.ListProvenanceEvents(ctx, local.ProvenanceEventOptions{
	RefKind: "document",
	RefID:   document.DocID,
	Limit:   10,
})
if err != nil {
	return err
}
_ = results
_ = events
```
