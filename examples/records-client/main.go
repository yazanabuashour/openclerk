package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	local "github.com/yazanabuashour/openclerk/client/local"
	records "github.com/yazanabuashour/openclerk/client/records"
)

func main() {
	ctx := context.Background()
	client, runtime, err := local.OpenRecords(config())
	if err != nil {
		log.Fatal(err)
	}
	defer runtime.Close()

	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	path := fmt.Sprintf("examples/records-%s.md", runID)
	title := fmt.Sprintf("Transmission solenoid %s", runID)
	entityID := fmt.Sprintf("transmission-solenoid-%s", runID)
	create, err := client.CreateDocumentWithResponse(ctx, records.CreateDocumentRequest{
		Path:  path,
		Title: title,
		Body: fmt.Sprintf(`---
entity_type: part
entity_name: %s
entity_id: %s
---
# %s

## Summary
Canonical part record.

## Facts
- sku: SOL-1
- vendor: OpenClerk Motors
`, title, entityID, title),
	})
	if err != nil {
		log.Fatal(err)
	}
	if create.JSON201 == nil {
		log.Fatalf("create document failed: %s", string(create.Body))
	}

	lookup, err := client.RecordsLookupWithResponse(ctx, records.RecordsLookupRequest{Text: runID})
	if err != nil {
		log.Fatal(err)
	}
	if lookup.JSON200 == nil {
		log.Fatalf("records lookup failed: %s", string(lookup.Body))
	}
	if len(lookup.JSON200.Entities) == 0 {
		log.Fatal("records lookup returned no entities")
	}

	entity, err := client.GetRecordEntityWithResponse(ctx, lookup.JSON200.Entities[0].EntityId)
	if err != nil {
		log.Fatal(err)
	}
	if entity.JSON200 == nil {
		log.Fatalf("get record entity failed: %s", string(entity.Body))
	}

	fmt.Printf("backend=%s dataDir=%s entity=%s facts=%d sourceDoc=%s\n", records.CapabilitiesBackendRecords, runtime.Paths().DataDir, entity.JSON200.EntityId, len(entity.JSON200.Facts), create.JSON201.DocId)
}

func config() local.Config {
	if value := os.Getenv("OPENCLERK_DATA_DIR"); value != "" {
		return local.Config{DataDir: value}
	}
	return local.Config{}
}
