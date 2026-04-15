package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	records "github.com/yazanabuashour/openclerk/client/records"
)

func main() {
	ctx := context.Background()
	client, err := records.NewClientWithResponses(serverURL())
	if err != nil {
		log.Fatal(err)
	}

	path := fmt.Sprintf("examples/records-%d.md", time.Now().UnixNano())
	create, err := client.CreateDocumentWithResponse(ctx, records.CreateDocumentRequest{
		Path:  path,
		Title: "Transmission solenoid",
		Body: `---
entity_type: part
entity_name: Transmission solenoid
entity_id: transmission-solenoid
---
# Transmission solenoid

## Summary
Canonical part record.

## Facts
- sku: SOL-1
- vendor: OpenClerk Motors
`,
	})
	if err != nil {
		log.Fatal(err)
	}
	if create.JSON201 == nil {
		log.Fatalf("create document failed: %s", string(create.Body))
	}

	lookup, err := client.RecordsLookupWithResponse(ctx, records.RecordsLookupRequest{Text: "solenoid"})
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

	fmt.Printf("backend=%s entity=%s facts=%d sourceDoc=%s\n", records.CapabilitiesBackendRecords, entity.JSON200.EntityId, len(entity.JSON200.Facts), create.JSON201.DocId)
}

func serverURL() string {
	if value := os.Getenv("OPENCLERK_SERVER"); value != "" {
		return value
	}
	return "http://127.0.0.1:8080"
}
