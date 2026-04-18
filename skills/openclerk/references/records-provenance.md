# Records, Graph, And Provenance Recipes

OpenClerk keeps canonical Markdown as the source of truth. Graph links,
promoted records, provenance events, and projection states are derived views over
those documents.

## Links And Graph Neighborhoods

```go
links, err := client.GetDocumentLinks(ctx, docID)
if err != nil {
	return err
}
for _, link := range links.Outgoing {
	log.Printf("outgoing %s %s", link.DocID, link.Path)
}

neighborhood, err := client.GraphNeighborhood(ctx, local.GraphNeighborhoodOptions{
	DocID: docID,
	Limit: 20,
})
if err != nil {
	return err
}
log.Printf("nodes=%d edges=%d", len(neighborhood.Nodes), len(neighborhood.Edges))
```

## Promoted Records

Use records lookup only for promoted-domain questions. The current projection is
derived from record-shaped Markdown with `entity_*` frontmatter and a `Facts`
section.

```go
lookup, err := client.LookupRecords(ctx, local.RecordLookupOptions{
	Text:  "solenoid",
	Limit: 10,
})
if err != nil {
	return err
}
for _, entity := range lookup.Entities {
	log.Printf("%s %s facts=%d", entity.EntityID, entity.Name, len(entity.Facts))
}

entity, err := client.GetRecordEntity(ctx, lookup.Entities[0].EntityID)
if err != nil {
	return err
}
```

## Provenance And Freshness

Use provenance and projection freshness when maintaining source-linked
synthesis. A synthesis page should not hide whether it came from canonical docs,
promoted records, or a derived projection that may need refresh.

```go
events, err := client.ListProvenanceEvents(ctx, local.ProvenanceEventOptions{
	RefKind: "document",
	RefID:   docID,
	Limit:   20,
})
if err != nil {
	return err
}

states, err := client.ListProjectionStates(ctx, local.ProjectionStateOptions{
	RefKind: "document",
	RefID:   docID,
	Limit:   20,
})
if err != nil {
	return err
}
log.Printf("events=%d projections=%d", len(events.Events), len(states.Projections))
```
