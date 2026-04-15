package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	hybrid "github.com/yazanabuashour/openclerk/client/hybrid"
	local "github.com/yazanabuashour/openclerk/client/local"
)

func main() {
	ctx := context.Background()
	cfg := config()
	cfg.EmbeddingProvider = "local"
	client, runtime, err := local.OpenHybrid(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer runtime.Close()

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

	fmt.Printf("backend=%s dataDir=%s modes=%s topChunk=%s\n", hybrid.CapabilitiesBackendHybrid, runtime.Paths().DataDir, strings.Join(modes, ","), search.JSON200.Hits[0].ChunkId)
}

func config() local.Config {
	if value := os.Getenv("OPENCLERK_DATA_DIR"); value != "" {
		return local.Config{DataDir: value}
	}
	return local.Config{}
}
