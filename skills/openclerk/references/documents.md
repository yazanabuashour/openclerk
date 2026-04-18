# Document Task Recipes

Use these snippets after opening the local runtime:

```go
client, err := local.OpenClient(local.Config{})
if err != nil {
	return err
}
defer client.Close()
ctx := context.Background()
```

## Create A Canonical Document

```go
document, err := client.CreateDocument(ctx, local.DocumentInput{
	Path:  "notes/projects/openclerk-roadmap.md",
	Title: "Roadmap",
	Body:  "---\ntype: project\nstatus: active\n---\n# Roadmap\n\n## Summary\nCanonical project note.\n",
})
if err != nil {
	return err
}
log.Printf("created %s at %s", document.DocID, document.Path)
```

`CreateDocument` fails with a conflict when the path already exists. For
incremental updates, use `AppendDocument` or `ReplaceSection`; do not overwrite
the whole Markdown body unless the user explicitly asks for that behavior.

## Create Source-Linked Synthesis

Search before creating synthesis. If a relevant topic, entity, comparison, or
overview page already exists, update it with `AppendDocument` or
`ReplaceSection` instead of creating a duplicate.

Use synthesis pages for durable compiled knowledge that should survive beyond
the current chat. Include source refs or citations in the body/frontmatter for
claims that depend on source evidence.

```go
document, err := client.CreateDocument(ctx, local.DocumentInput{
	Path:  "notes/synthesis/openclerk-knowledge-plane.md",
	Title: "OpenClerk knowledge plane synthesis",
	Body:  "---\ntype: synthesis\nstatus: active\nfreshness: fresh\nsource_refs:\n  - notes/architecture/knowledge-plane.md\n---\n# OpenClerk knowledge plane synthesis\n\n## Summary\nSource-linked synthesis of the current architecture.\n\n## Sources\n- notes/architecture/knowledge-plane.md\n",
})
if err != nil {
	return err
}
```

## List And Read Documents

```go
docs, err := client.ListDocuments(ctx, local.DocumentListOptions{
	PathPrefix:    "notes/",
	MetadataKey:   "status",
	MetadataValue: "active",
	Limit:         20,
})
if err != nil {
	return err
}
for _, doc := range docs.Documents {
	log.Printf("%s %s", doc.DocID, doc.Path)
}

document, err := client.GetDocument(ctx, docs.Documents[0].DocID)
if err != nil {
	return err
}
log.Printf("%s headings=%v", document.Title, document.Headings)
```

## Append Or Replace A Section

```go
updated, err := client.AppendDocument(ctx, document.DocID, "## Decisions\nUse the code-first local SDK.")
if err != nil {
	return err
}

updated, err = client.ReplaceSection(ctx, updated.DocID, "Decisions", "Use `local.OpenClient` for routine agent workflows.")
if err != nil {
	return err
}
log.Printf("updated %s", updated.Path)
```
