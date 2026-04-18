# Search Task Recipes

Use `Search` for questions like "what do I know about X?" Results include
citations that point back to canonical Markdown documents and chunks.

For source-linked synthesis work, search first. Prefer updating an existing
topic/entity/comparison page when results show one already exists, and use the
returned citations as the evidence trail for any source-sensitive answer.

```go
client, err := local.OpenClient(local.Config{})
if err != nil {
	return err
}
defer client.Close()

results, err := client.Search(ctx, local.SearchOptions{
	Text:  "architecture",
	Limit: 10,
})
if err != nil {
	return err
}
for _, hit := range results.Hits {
	log.Printf("%s %s", hit.DocID, hit.Snippet)
	for _, citation := range hit.Citations {
		log.Printf("source %s lines %d-%d", citation.Path, citation.LineStart, citation.LineEnd)
	}
}
```

## Scoped Search

Use path and metadata filters when the user narrows the request to a folder,
notebook, type, status, or other frontmatter value.

```go
results, err := client.Search(ctx, local.SearchOptions{
	Text:          "roadmap",
	PathPrefix:    "notes/projects/",
	MetadataKey:   "status",
	MetadataValue: "active",
	Limit:         10,
})
if err != nil {
	return err
}
```

If the user asks where the data came from, report citation paths and line ranges
from the returned hits. If results are empty, report the resolved `VaultRoot` so
the user can tell which local dataset was checked.
