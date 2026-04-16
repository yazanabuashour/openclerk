# OpenClerk Reference

## Install in a Go workspace

Use one tagged install command:

```bash
go get github.com/yazanabuashour/openclerk/client/local@v0.1.0
```

Import [`../../../client/openclerk`](../../../client/openclerk) from the same module for generated types. Do not add a second `go get` step for that package.

## Agent quick start

Use `local.Config{}` for live local state:

```go
paths, err := local.ResolvePaths(local.Config{})
if err != nil {
	return err
}
fmt.Printf("dataDir=%s db=%s vault=%s\n", paths.DataDir, paths.DatabasePath, paths.VaultRoot)

client, runtime, err := local.Open(local.Config{})
if err != nil {
	return err
}
defer runtime.Close()
```

Use explicit storage overrides for tests and throwaway examples:

```go
client, runtime, err := local.Open(local.Config{
	DataDir: t.TempDir(),
})
if err != nil {
	return err
}
defer runtime.Close()
```

`openclerk.NewClientWithResponses(baseURL)` is for intentional [`../../../cmd/openclerkd`](../../../cmd/openclerkd) HTTP debugging. Prefer embedded `local.Open` for normal local questions.

List documents:

```go
limit := 20
pathPrefix := "notes/"
docs, err := client.ListDocumentsWithResponse(ctx, &openclerk.ListDocumentsParams{
	PathPrefix: &pathPrefix,
	Limit:      &limit,
})
if err != nil {
	return err
}
if docs.JSON200 == nil {
	return fmt.Errorf("list documents failed: %s", string(docs.Body))
}
```

Search what OpenClerk knows:

```go
limit := 10
search, err := client.SearchQueryWithResponse(ctx, openclerk.SearchQuery{
	Text:  "architecture",
	Limit: &limit,
})
if err != nil {
	return err
}
if search.JSON200 == nil {
	return fmt.Errorf("search failed: %s", string(search.Body))
}
```

Look up promoted records:

```go
limit := 10
lookup, err := client.RecordsLookupWithResponse(ctx, openclerk.RecordsLookupRequest{
	Text:  "solenoid",
	Limit: &limit,
})
if err != nil {
	return err
}
if lookup.JSON200 == nil {
	return fmt.Errorf("records lookup failed: %s", string(lookup.Body))
}
```

Inspect links and provenance:

```go
links, err := client.GetDocumentLinksWithResponse(ctx, docID)
if err != nil {
	return err
}
if links.JSON200 == nil {
	return fmt.Errorf("links failed: %s", string(links.Body))
}

refKind := "document"
events, err := client.ListProvenanceEventsWithResponse(ctx, &openclerk.ListProvenanceEventsParams{
	RefKind: &refKind,
	RefId:   &docID,
	Limit:   &limit,
})
if err != nil {
	return err
}
if events.JSON200 == nil {
	return fmt.Errorf("provenance failed: %s", string(events.Body))
}
```

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
