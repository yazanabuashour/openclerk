package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestSynthesisProjectionIsFreshForCurrentSources(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	clock := testClock()
	store.now = func() time.Time { return clock }

	clock = clock.Add(time.Minute)
	source, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "notes/sources/current.md",
		Title: "Current Source",
		Body:  "# Current Source\n\n## Summary\nCurrent canonical evidence.\n",
	})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	clock = clock.Add(time.Minute)
	synthesis, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "notes/synthesis/current.md",
		Title: "Current Synthesis",
		Body:  synthesisBody("notes/sources/current.md", "Current canonical evidence."),
	})
	if err != nil {
		t.Fatalf("create synthesis: %v", err)
	}

	projection := requireSynthesisProjection(t, ctx, store, synthesis.DocID)
	if projection.Freshness != "fresh" {
		t.Fatalf("freshness = %q, want fresh: %+v", projection.Freshness, projection)
	}
	if projection.SourceRef != "doc:"+source.DocID {
		t.Fatalf("source_ref = %q, want doc ref for source", projection.SourceRef)
	}
	if projection.Details["current_source_refs"] != "notes/sources/current.md" ||
		projection.Details["source_refs"] != "notes/sources/current.md" ||
		projection.Details["freshness_reason"] != "sources current" {
		t.Fatalf("projection details = %+v", projection.Details)
	}
}

func TestSynthesisProjectionStaleAfterSourceUpdateAndFreshAfterRepair(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	clock := testClock()
	store.now = func() time.Time { return clock }

	clock = clock.Add(time.Minute)
	source, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "notes/sources/runner.md",
		Title: "Runner Source",
		Body:  "# Runner Source\n\n## Summary\nInitial source guidance.\n",
	})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	clock = clock.Add(time.Minute)
	synthesis, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "notes/synthesis/runner.md",
		Title: "Runner Synthesis",
		Body:  synthesisBody("notes/sources/runner.md", "Initial source guidance."),
	})
	if err != nil {
		t.Fatalf("create synthesis: %v", err)
	}
	if got := requireSynthesisProjection(t, ctx, store, synthesis.DocID); got.Freshness != "fresh" {
		t.Fatalf("initial projection freshness = %q, want fresh", got.Freshness)
	}

	clock = clock.Add(time.Minute)
	if _, err := store.ReplaceDocumentSection(ctx, source.DocID, domain.ReplaceSectionInput{
		Heading: "Summary",
		Content: "Updated source guidance.",
	}); err != nil {
		t.Fatalf("update source: %v", err)
	}
	stale := requireSynthesisProjection(t, ctx, store, synthesis.DocID)
	if stale.Freshness != "stale" ||
		stale.Details["stale_source_refs"] != "notes/sources/runner.md" ||
		!strings.Contains(stale.Details["freshness_reason"], "source newer than synthesis") {
		t.Fatalf("stale projection = %+v", stale)
	}
	invalidations, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind: "projection",
		RefID:   "synthesis:" + synthesis.DocID,
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("list invalidations: %v", err)
	}
	if !hasEventType(invalidations.Events, "projection_invalidated") {
		t.Fatalf("missing synthesis invalidation event: %+v", invalidations.Events)
	}
	sourceEvents, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind: "source",
		RefID:   source.DocID,
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("list source events: %v", err)
	}
	if !hasEventType(sourceEvents.Events, "source_created") || !hasEventType(sourceEvents.Events, "source_updated") {
		t.Fatalf("source events = %+v, want created and updated", sourceEvents.Events)
	}

	clock = clock.Add(time.Minute)
	if _, err := store.ReplaceDocumentSection(ctx, synthesis.DocID, domain.ReplaceSectionInput{
		Heading: "Freshness",
		Content: "Checked source: notes/sources/runner.md after the source update.",
	}); err != nil {
		t.Fatalf("repair synthesis: %v", err)
	}
	repaired := requireSynthesisProjection(t, ctx, store, synthesis.DocID)
	if repaired.Freshness != "fresh" || repaired.Details["stale_source_refs"] != "" {
		t.Fatalf("repaired projection = %+v", repaired)
	}
	events, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind: "projection",
		RefID:   "synthesis:" + synthesis.DocID,
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("list synthesis events: %v", err)
	}
	if !hasEventType(events.Events, "projection_refreshed") {
		t.Fatalf("missing synthesis refresh event: %+v", events.Events)
	}
}

func TestSynthesisProjectionReportsMissingAndSupersededSources(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	clock := testClock()
	store.now = func() time.Time { return clock }

	clock = clock.Add(time.Minute)
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "notes/sources/old.md",
		Title: "Old Source",
		Body: strings.TrimSpace(`---
status: superseded
superseded_by: notes/sources/current.md
---
# Old Source

## Summary
Old guidance.
`) + "\n",
	}); err != nil {
		t.Fatalf("create old source: %v", err)
	}
	clock = clock.Add(time.Minute)
	synthesis, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "notes/synthesis/missing.md",
		Title: "Missing Synthesis",
		Body:  synthesisBody("notes/sources/old.md, notes/sources/missing.md", "Old guidance."),
	})
	if err != nil {
		t.Fatalf("create synthesis: %v", err)
	}

	projection := requireSynthesisProjection(t, ctx, store, synthesis.DocID)
	if projection.Freshness != "stale" {
		t.Fatalf("freshness = %q, want stale", projection.Freshness)
	}
	if projection.Details["missing_source_refs"] != "notes/sources/missing.md" {
		t.Fatalf("missing source refs = %q", projection.Details["missing_source_refs"])
	}
	if projection.Details["superseded_source_refs"] != "notes/sources/old.md" {
		t.Fatalf("superseded source refs = %q", projection.Details["superseded_source_refs"])
	}
	if projection.Details["current_source_refs"] != "notes/sources/current.md" {
		t.Fatalf("current source refs = %q", projection.Details["current_source_refs"])
	}
	if !strings.Contains(projection.Details["freshness_reason"], "current replacement missing from source refs") {
		t.Fatalf("freshness reason = %q", projection.Details["freshness_reason"])
	}
}

func TestSynthesisProjectionFreshWithSupersedesAndSupersededByMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	clock := testClock()
	store.now = func() time.Time { return clock }

	clock = clock.Add(time.Minute)
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "notes/sources/old.md",
		Title: "Old Source",
		Body: strings.TrimSpace(`---
status: superseded
superseded_by: notes/sources/current.md
---
# Old Source

## Summary
Old guidance.
`) + "\n",
	}); err != nil {
		t.Fatalf("create old source: %v", err)
	}
	clock = clock.Add(time.Minute)
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "notes/sources/current.md",
		Title: "Current Source",
		Body: strings.TrimSpace(`---
supersedes: notes/sources/old.md
---
# Current Source

## Summary
Current guidance.
`) + "\n",
	}); err != nil {
		t.Fatalf("create current source: %v", err)
	}
	clock = clock.Add(time.Minute)
	synthesis, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "notes/synthesis/supersession.md",
		Title: "Supersession Synthesis",
		Body:  synthesisBody("notes/sources/current.md, notes/sources/old.md", "Current guidance supersedes old guidance."),
	})
	if err != nil {
		t.Fatalf("create synthesis: %v", err)
	}

	projection := requireSynthesisProjection(t, ctx, store, synthesis.DocID)
	if projection.Freshness != "fresh" {
		t.Fatalf("freshness = %q, want fresh: %+v", projection.Freshness, projection)
	}
	if projection.Details["current_source_refs"] != "notes/sources/current.md" {
		t.Fatalf("current source refs = %q", projection.Details["current_source_refs"])
	}
	if projection.Details["superseded_source_refs"] != "notes/sources/old.md" {
		t.Fatalf("superseded source refs = %q", projection.Details["superseded_source_refs"])
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

func testClock() time.Time {
	return time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
}

func synthesisBody(sourceRefs string, summary string) string {
	return strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: `+sourceRefs+`
---
# Synthesis

## Summary
`+summary+`

## Sources
- `+sourceRefs+`

## Freshness
Checked source refs.
`) + "\n"
}

func requireSynthesisProjection(t *testing.T, ctx context.Context, store *Store, docID string) domain.ProjectionState {
	t.Helper()

	result, err := store.ListProjectionStates(ctx, domain.ProjectionStateQuery{
		Projection: "synthesis",
		RefKind:    "document",
		RefID:      docID,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("list synthesis projection: %v", err)
	}
	if len(result.Projections) != 1 {
		t.Fatalf("projection count = %d, want 1: %+v", len(result.Projections), result.Projections)
	}
	return result.Projections[0]
}

func hasEventType(events []domain.ProvenanceEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}
