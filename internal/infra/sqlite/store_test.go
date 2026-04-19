package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

func TestCreateDocumentRejectsDuplicatePath(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	first, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/widget.md",
		Title: "Widget One",
		Body:  "# Widget One\n\nfirst body",
	})
	if err != nil {
		t.Fatalf("create first document: %v", err)
	}

	_, err = store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/widget.md",
		Title: "Widget Two",
		Body:  "# Widget Two\n\nsecond body",
	})
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Status != 409 {
		t.Fatalf("duplicate create error = %v, want already exists 409", err)
	}

	got, err := store.GetDocument(context.Background(), first.DocID)
	if err != nil {
		t.Fatalf("get original document: %v", err)
	}
	if got.Title != "Widget One" || !strings.Contains(got.Body, "first body") {
		t.Fatalf("original document was overwritten: %+v", got)
	}
}

func TestSyncVaultPrunesDeletedDocuments(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	docPath := filepath.Join(vaultRoot, "docs", "widget.md")
	if err := os.MkdirAll(filepath.Dir(docPath), 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	if err := os.WriteFile(docPath, []byte("# Widget\n\nalpha signal\n"), 0o644); err != nil {
		t.Fatalf("write vault doc: %v", err)
	}

	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	search, err := store.Search(context.Background(), domain.SearchQuery{Text: "alpha", Limit: 10})
	if err != nil {
		t.Fatalf("search before delete: %v", err)
	}
	if len(search.Hits) != 1 {
		t.Fatalf("search before delete hit count = %d, want 1", len(search.Hits))
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close initial store: %v", err)
	}

	if err := os.Remove(docPath); err != nil {
		t.Fatalf("remove vault doc: %v", err)
	}

	reopened := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = reopened.Close()
	}()

	search, err = reopened.Search(context.Background(), domain.SearchQuery{Text: "alpha", Limit: 10})
	if err != nil {
		t.Fatalf("search after delete: %v", err)
	}
	if len(search.Hits) != 0 {
		t.Fatalf("search after delete hit count = %d, want 0", len(search.Hits))
	}

	_, err = reopened.GetDocument(context.Background(), docIDForPath("docs/widget.md"))
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Status != 404 {
		t.Fatalf("get deleted document error = %v, want not found 404", err)
	}
}

func TestSyncVaultPrunesDeletedServiceProjection(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	if _, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "records/services/openclerk-runner.md",
		Title: "OpenClerk runner",
		Body: strings.TrimSpace(`---
service_id: openclerk-runner
service_name: OpenClerk runner
service_status: active
service_owner: runner
service_interface: JSON runner
---
# OpenClerk runner

## Summary
Production service.
`) + "\n",
	}); err != nil {
		t.Fatalf("create service document: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close initial store: %v", err)
	}

	if err := os.Remove(filepath.Join(vaultRoot, "records", "services", "openclerk-runner.md")); err != nil {
		t.Fatalf("remove service doc: %v", err)
	}

	reopened := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = reopened.Close()
	}()

	services, err := reopened.ServicesLookup(context.Background(), domain.ServiceLookupInput{Text: "OpenClerk runner", Limit: 10})
	if err != nil {
		t.Fatalf("services lookup after delete: %v", err)
	}
	if len(services.Services) != 0 {
		t.Fatalf("services after delete = %+v, want none", services.Services)
	}
}

func TestServicesLookupSearchesSummarySection(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	if _, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "records/services/openclerk-runner.md",
		Title: "OpenClerk runner",
		Body: strings.TrimSpace(`---
service_id: openclerk-runner
service_name: OpenClerk runner
service_status: active
service_owner: runner
service_interface: JSON runner
---
# OpenClerk runner

## Summary
Production service for routine local knowledge tasks.

## Facts
- tier: production
`) + "\n",
	}); err != nil {
		t.Fatalf("create service document: %v", err)
	}

	services, err := store.ServicesLookup(context.Background(), domain.ServiceLookupInput{Text: "routine local knowledge", Limit: 10})
	if err != nil {
		t.Fatalf("services lookup: %v", err)
	}
	if len(services.Services) != 1 || services.Services[0].ServiceID != "openclerk-runner" {
		t.Fatalf("services lookup = %+v, want openclerk-runner", services)
	}
	if services.Services[0].Summary != "Production service for routine local knowledge tasks." {
		t.Fatalf("service summary = %q", services.Services[0].Summary)
	}
}

func TestCreateDocumentPreservesRequestedTitleAcrossRestart(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)

	document, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/widget.md",
		Title: "Wanted Title",
		Body:  "body only no heading",
	})
	if err != nil {
		t.Fatalf("create document: %v", err)
	}
	if document.Title != "Wanted Title" {
		t.Fatalf("created document title = %q, want %q", document.Title, "Wanted Title")
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close initial store: %v", err)
	}

	reopened := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = reopened.Close()
	}()

	got, err := reopened.GetDocument(context.Background(), document.DocID)
	if err != nil {
		t.Fatalf("get document after restart: %v", err)
	}
	if got.Title != "Wanted Title" {
		t.Fatalf("reopened document title = %q, want %q", got.Title, "Wanted Title")
	}
}

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

func openTestStore(t *testing.T, backend domain.BackendKind, dbPath string, vaultRoot string) *Store {
	t.Helper()

	store, err := New(context.Background(), Config{
		Backend:      backend,
		DatabasePath: dbPath,
		VaultRoot:    vaultRoot,
	})
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	return store
}
