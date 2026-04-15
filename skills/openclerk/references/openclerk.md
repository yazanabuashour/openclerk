# OpenClerk Reference

## Install in a Go workspace

Use one tagged install command:

```bash
go get github.com/yazanabuashour/openclerk/client/local@v0.1.0
```

Import [`../../../client/openclerk`](../../../client/openclerk) from the same module for generated types. Do not add a second `go get` step for that package.

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
