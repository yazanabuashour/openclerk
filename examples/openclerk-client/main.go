package main

import (
	"context"
	"fmt"
	"log"
	"os"

	local "github.com/yazanabuashour/openclerk/client/local"
	openclerk "github.com/yazanabuashour/openclerk/client/openclerk"
)

func main() {
	ctx := context.Background()
	client, runtime, err := local.Open(config())
	if err != nil {
		log.Fatal(err)
	}
	defer runtime.Close()

	architecture, err := client.CreateDocumentWithResponse(ctx, openclerk.CreateDocumentRequest{
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
	if architecture.JSON201 == nil {
		log.Fatalf("create architecture note failed: %s", string(architecture.Body))
	}

	roadmap, err := client.CreateDocumentWithResponse(ctx, openclerk.CreateDocumentRequest{
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
	if roadmap.JSON201 == nil {
		log.Fatalf("create roadmap note failed: %s", string(roadmap.Body))
	}

	record, err := client.CreateDocumentWithResponse(ctx, openclerk.CreateDocumentRequest{
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
	if record.JSON201 == nil {
		log.Fatalf("create record note failed: %s", string(record.Body))
	}

	pathPrefix := "notes/"
	list, err := client.ListDocumentsWithResponse(ctx, &openclerk.ListDocumentsParams{PathPrefix: &pathPrefix})
	if err != nil {
		log.Fatal(err)
	}
	if list.JSON200 == nil {
		log.Fatalf("list documents failed: %s", string(list.Body))
	}

	links, err := client.GetDocumentLinksWithResponse(ctx, roadmap.JSON201.DocId)
	if err != nil {
		log.Fatal(err)
	}
	if links.JSON200 == nil {
		log.Fatalf("get document links failed: %s", string(links.Body))
	}

	lookup, err := client.RecordsLookupWithResponse(ctx, openclerk.RecordsLookupRequest{Text: "solenoid"})
	if err != nil {
		log.Fatal(err)
	}
	if lookup.JSON200 == nil || len(lookup.JSON200.Entities) == 0 {
		log.Fatalf("records lookup failed: %s", string(lookup.Body))
	}

	refKind := "document"
	events, err := client.ListProvenanceEventsWithResponse(ctx, &openclerk.ListProvenanceEventsParams{
		RefKind: &refKind,
		RefId:   &roadmap.JSON201.DocId,
	})
	if err != nil {
		log.Fatal(err)
	}
	if events.JSON200 == nil {
		log.Fatalf("list provenance events failed: %s", string(events.Body))
	}

	fmt.Printf(
		"backend=%s dataDir=%s docs=%d links=%d entity=%s events=%d\n",
		openclerk.CapabilitiesBackendOpenclerk,
		runtime.Paths().DataDir,
		len(list.JSON200.Documents),
		len(links.JSON200.Outgoing),
		lookup.JSON200.Entities[0].EntityId,
		len(events.JSON200.Events),
	)
}

func config() local.Config {
	if value := os.Getenv("OPENCLERK_DATA_DIR"); value != "" {
		return local.Config{DataDir: value}
	}
	return local.Config{}
}
