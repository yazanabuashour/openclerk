package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	graph "github.com/yazanabuashour/openclerk/client/graph"
)

func main() {
	ctx := context.Background()
	client, err := graph.NewClientWithResponses(serverURL())
	if err != nil {
		log.Fatal(err)
	}

	prefix := fmt.Sprintf("examples/graph-%d", time.Now().UnixNano())
	targetPath := prefix + "/reference.md"
	sourcePath := prefix + "/guide.md"

	target, err := client.CreateDocumentWithResponse(ctx, graph.CreateDocumentRequest{
		Path:  targetPath,
		Title: "Reference",
		Body: `# Reference

## Overview
Canonical supporting note.
`,
	})
	if err != nil {
		log.Fatal(err)
	}
	if target.JSON201 == nil {
		log.Fatalf("create target failed: %s", string(target.Body))
	}

	source, err := client.CreateDocumentWithResponse(ctx, graph.CreateDocumentRequest{
		Path:  sourcePath,
		Title: "Guide",
		Body: `# Guide

## Overview
See the [reference](reference.md) for details.
`,
	})
	if err != nil {
		log.Fatal(err)
	}
	if source.JSON201 == nil {
		log.Fatalf("create source failed: %s", string(source.Body))
	}

	limit := 8
	neighborhood, err := client.GraphNeighborhoodWithResponse(ctx, graph.GraphNeighborhoodRequest{
		DocId: &source.JSON201.DocId,
		Limit: &limit,
	})
	if err != nil {
		log.Fatal(err)
	}
	if neighborhood.JSON200 == nil {
		log.Fatalf("graph neighborhood failed: %s", string(neighborhood.Body))
	}

	fmt.Printf("backend=%s nodes=%d edges=%d source=%s target=%s\n", graph.CapabilitiesBackendGraph, len(neighborhood.JSON200.Nodes), len(neighborhood.JSON200.Edges), source.JSON201.DocId, target.JSON201.DocId)
}

func serverURL() string {
	if value := os.Getenv("OPENCLERK_SERVER"); value != "" {
		return value
	}
	return "http://127.0.0.1:8080"
}
