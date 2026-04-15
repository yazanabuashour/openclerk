package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	fts "github.com/yazanabuashour/openclerk/client/fts"
)

func main() {
	ctx := context.Background()
	client, err := fts.NewClientWithResponses(serverURL())
	if err != nil {
		log.Fatal(err)
	}

	path := fmt.Sprintf("examples/fts-%d.md", time.Now().UnixNano())
	create, err := client.CreateDocumentWithResponse(ctx, fts.CreateDocumentRequest{
		Path:  path,
		Title: "FTS example",
		Body: `# FTS example

## Overview
Transmission solenoid exact-match note.

## Facts
- sku: SOL-1
`,
	})
	if err != nil {
		log.Fatal(err)
	}
	if create.JSON201 == nil {
		log.Fatalf("create document failed: %s", string(create.Body))
	}

	limit := 3
	search, err := client.SearchQueryWithResponse(ctx, fts.SearchQuery{
		Text:  "solenoid",
		Limit: &limit,
	})
	if err != nil {
		log.Fatal(err)
	}
	if search.JSON200 == nil {
		log.Fatalf("search failed: %s", string(search.Body))
	}
	if len(search.JSON200.Hits) == 0 {
		log.Fatal("search returned no hits")
	}

	fmt.Printf("backend=%s doc=%s chunk=%s title=%q\n", fts.CapabilitiesBackendFts, create.JSON201.DocId, search.JSON200.Hits[0].ChunkId, search.JSON200.Hits[0].Title)
}

func serverURL() string {
	if value := os.Getenv("OPENCLERK_SERVER"); value != "" {
		return value
	}
	return "http://127.0.0.1:8080"
}
