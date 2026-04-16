package main

import (
	"context"
	"fmt"
	"log"
	"os"

	local "github.com/yazanabuashour/openclerk/client/local"
)

func main() {
	ctx := context.Background()
	client, err := local.OpenClient(config())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	architecture, err := client.CreateDocument(ctx, local.DocumentInput{
		Path:  "notes/architecture/knowledge-plane.md",
		Title: "Knowledge plane",
		Body: `---
type: note
status: active
---
# Knowledge plane

## Summary
Canonical agent-facing architecture note.
`,
	})
	if err != nil {
		log.Fatal(err)
	}

	roadmap, err := client.CreateDocument(ctx, local.DocumentInput{
		Path:  "notes/projects/openclerk-roadmap.md",
		Title: "Roadmap",
		Body: `---
type: project
status: active
---
# Roadmap

## Summary
See the [knowledge plane](../architecture/knowledge-plane.md) architecture note.
`,
	})
	if err != nil {
		log.Fatal(err)
	}

	record, err := client.CreateDocument(ctx, local.DocumentInput{
		Path:  "records/assets/transmission-solenoid.md",
		Title: "Transmission solenoid",
		Body: `---
entity_type: part
entity_name: Transmission solenoid
entity_id: transmission-solenoid
type: record
status: active
---
# Transmission solenoid

## Summary
Canonical promoted-domain baseline.

## Facts
- sku: SOL-1
- vendor: OpenClerk Motors
`,
	})
	if err != nil {
		log.Fatal(err)
	}

	list, err := client.ListDocuments(ctx, local.DocumentListOptions{PathPrefix: "notes/"})
	if err != nil {
		log.Fatal(err)
	}

	links, err := client.GetDocumentLinks(ctx, roadmap.DocID)
	if err != nil {
		log.Fatal(err)
	}

	lookup, err := client.LookupRecords(ctx, local.RecordLookupOptions{Text: "solenoid"})
	if err != nil {
		log.Fatal(err)
	}
	if len(lookup.Entities) == 0 {
		log.Fatal("records lookup returned no entities")
	}

	events, err := client.ListProvenanceEvents(ctx, local.ProvenanceEventOptions{
		RefKind: "document",
		RefID:   roadmap.DocID,
	})
	if err != nil {
		log.Fatal(err)
	}
	if architecture.DocID == "" || record.DocID == "" {
		log.Fatal("created documents returned empty ids")
	}

	fmt.Printf(
		"backend=%s dataDir=%s docs=%d links=%d entity=%s events=%d\n",
		"openclerk",
		client.Paths().DataDir,
		len(list.Documents),
		len(links.Outgoing),
		lookup.Entities[0].EntityID,
		len(events.Events),
	)
}

func config() local.Config {
	if value := os.Getenv("OPENCLERK_DATA_DIR"); value != "" {
		return local.Config{DataDir: value}
	}
	return local.Config{}
}
