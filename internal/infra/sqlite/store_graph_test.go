package sqlite

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

func TestGraphNeighborhoodIncludesOutgoingLinksForChunk(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	target, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/reference.md",
		Title: "Reference",
		Body:  "# Reference\n\nCanonical supporting note.\n",
	})
	if err != nil {
		t.Fatalf("create target document: %v", err)
	}
	source, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/guide.md",
		Title: "Guide",
		Body: strings.TrimSpace(`
# Guide

## Overview
See the [reference](reference.md) for details.
`),
	})
	if err != nil {
		t.Fatalf("create source document: %v", err)
	}

	search, err := store.Search(context.Background(), domain.SearchQuery{Text: "reference", Limit: 10})
	if err != nil {
		t.Fatalf("search source chunk: %v", err)
	}
	var chunkID string
	for _, hit := range search.Hits {
		if hit.DocID == source.DocID {
			chunkID = hit.ChunkID
			break
		}
	}
	if chunkID == "" {
		t.Fatal("did not find source chunk in search results")
	}

	neighborhood, err := store.GraphNeighborhood(context.Background(), domain.GraphNeighborhoodInput{ChunkID: chunkID, Limit: 10})
	if err != nil {
		t.Fatalf("graph neighborhood by chunk: %v", err)
	}

	targetNodeID := "doc:" + target.DocID
	foundNode := false
	foundEdge := false
	for _, node := range neighborhood.Nodes {
		if node.NodeID == targetNodeID {
			foundNode = true
		}
	}
	for _, edge := range neighborhood.Edges {
		if edge.FromNodeID == "chunk:"+chunkID && edge.ToNodeID == targetNodeID && edge.Kind == "links_to" {
			foundEdge = true
		}
	}
	if !foundNode || !foundEdge {
		t.Fatalf("chunk neighborhood missing outgoing link: nodes=%v edges=%v", neighborhood.Nodes, neighborhood.Edges)
	}
}

func TestGraphProjectionIgnoresDuplicateMarkdownLinks(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	if _, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/reference.md",
		Title: "Reference",
		Body:  "# Reference\n\nCanonical supporting note.\n",
	}); err != nil {
		t.Fatalf("create target document: %v", err)
	}
	if _, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/guide.md",
		Title: "Guide",
		Body: strings.TrimSpace(`
# Guide

## Overview
See the [reference](reference.md) for details.
See the [reference](reference.md) again before writing synthesis.
`),
	}); err != nil {
		t.Fatalf("create source document with duplicate links: %v", err)
	}
}

func TestResolveLinkPathUsesVaultPathPolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		doc    string
		target string
		want   string
	}{
		{
			name:   "relative markdown path",
			doc:    "docs/guide.md",
			target: "reference",
			want:   "docs/reference.md",
		},
		{
			name:   "normalizes backslash target",
			doc:    "docs/guide.md",
			target: `references\runner.md`,
			want:   "docs/references/runner.md",
		},
		{
			name:   "rejects root absolute target",
			doc:    "docs/guide.md",
			target: "/reference.md",
			want:   "",
		},
		{
			name:   "rejects windows absolute target",
			doc:    "docs/guide.md",
			target: `C:\reference.md`,
			want:   "",
		},
		{
			name:   "rejects root escape",
			doc:    "guide.md",
			target: "../reference.md",
			want:   "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := resolveLinkPath(tt.doc, tt.target); got != tt.want {
				t.Fatalf("resolveLinkPath(%q, %q) = %q, want %q", tt.doc, tt.target, got, tt.want)
			}
		})
	}
}
