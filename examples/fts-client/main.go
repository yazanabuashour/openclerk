package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	fts "github.com/yazanabuashour/openclerk/client/fts"
	local "github.com/yazanabuashour/openclerk/client/local"
)

func main() {
	ctx := context.Background()
	client, runtime, err := local.OpenFTS(config())
	if err != nil {
		log.Fatal(err)
	}
	defer runtime.Close()

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

	fmt.Printf("backend=%s dataDir=%s doc=%s chunk=%s title=%q\n", fts.CapabilitiesBackendFts, runtime.Paths().DataDir, create.JSON201.DocId, search.JSON200.Hits[0].ChunkId, search.JSON200.Hits[0].Title)
}

func config() local.Config {
	if value := os.Getenv("OPENCLERK_DATA_DIR"); value != "" {
		return local.Config{DataDir: value}
	}
	return local.Config{}
}
