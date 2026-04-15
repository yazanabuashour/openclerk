package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	hybrid "github.com/yazanabuashour/openclerk/client/hybrid"
)

func main() {
	ctx := context.Background()
	client, err := hybrid.NewClientWithResponses(serverURL())
	if err != nil {
		log.Fatal(err)
	}

	path := fmt.Sprintf("examples/hybrid-%d.md", time.Now().UnixNano())
	create, err := client.CreateDocumentWithResponse(ctx, hybrid.CreateDocumentRequest{
		Path:  path,
		Title: "Hybrid example",
		Body: `# Hybrid example

## Overview
Transmission solenoid similarity anchor.

## Notes
Magnetic valve actuator for drivetrain control.
`,
	})
	if err != nil {
		log.Fatal(err)
	}
	if create.JSON201 == nil {
		log.Fatalf("create document failed: %s", string(create.Body))
	}

	capabilities, err := client.GetCapabilitiesWithResponse(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if capabilities.JSON200 == nil {
		log.Fatalf("capabilities failed: %s", string(capabilities.Body))
	}

	limit := 3
	search, err := client.SearchQueryWithResponse(ctx, hybrid.SearchQuery{
		Text:  "drivetrain actuator",
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

	modes := make([]string, 0, len(capabilities.JSON200.SearchModes))
	for _, mode := range capabilities.JSON200.SearchModes {
		modes = append(modes, string(mode))
	}

	fmt.Printf("backend=%s modes=%s topChunk=%s\n", hybrid.CapabilitiesBackendHybrid, strings.Join(modes, ","), search.JSON200.Hits[0].ChunkId)
}

func serverURL() string {
	if value := os.Getenv("OPENCLERK_SERVER"); value != "" {
		return value
	}
	return "http://127.0.0.1:8080"
}
