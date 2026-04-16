package local

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

func TestOpenClientKnowledgePlaneFacade(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client, err := OpenClient(Config{DataDir: filepath.Join(t.TempDir(), "data")})
	if err != nil {
		t.Fatalf("open client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if client.Generated() == nil {
		t.Fatal("generated fallback client is nil")
	}
	if client.Paths().DataDir == "" || client.Paths().DatabasePath == "" || client.Paths().VaultRoot == "" {
		t.Fatalf("resolved paths incomplete: %+v", client.Paths())
	}

	capabilities, err := client.Capabilities(ctx)
	if err != nil {
		t.Fatalf("capabilities: %v", err)
	}
	if capabilities.Backend != string(domain.BackendOpenClerk) {
		t.Fatalf("backend = %q, want %q", capabilities.Backend, domain.BackendOpenClerk)
	}
	if !contains(capabilities.Extensions, "graph") || !contains(capabilities.Extensions, "records") || !contains(capabilities.Extensions, "provenance") {
		t.Fatalf("extensions = %v, want graph, records, provenance", capabilities.Extensions)
	}

	architecture, err := client.CreateDocument(ctx, DocumentInput{
		Path:  "notes/architecture/knowledge-plane.md",
		Title: "Knowledge plane",
		Body: strings.TrimSpace(`
---
type: note
status: active
---
# Knowledge plane

## Summary
Canonical agent-facing architecture note.
`) + "\n",
	})
	if err != nil {
		t.Fatalf("create architecture note: %v", err)
	}

	roadmap, err := client.CreateDocument(ctx, DocumentInput{
		Path:  "notes/projects/openclerk-roadmap.md",
		Title: "Roadmap",
		Body: strings.TrimSpace(`
---
type: project
status: active
---
# Roadmap

## Summary
See the [knowledge plane](../architecture/knowledge-plane.md) architecture note.
`) + "\n",
	})
	if err != nil {
		t.Fatalf("create roadmap note: %v", err)
	}

	record, err := client.CreateDocument(ctx, DocumentInput{
		Path:  "records/assets/transmission-solenoid.md",
		Title: "Transmission solenoid",
		Body: strings.TrimSpace(`
---
entity_type: part
entity_name: Transmission solenoid
entity_id: transmission-solenoid
type: record
status: active
---
# Transmission solenoid

## Summary
Canonical promoted-domain baseline.

## Facts
- sku: SOL-1
- vendor: OpenClerk Motors
`) + "\n",
	})
	if err != nil {
		t.Fatalf("create record note: %v", err)
	}

	gotRoadmap, err := client.GetDocument(ctx, roadmap.DocID)
	if err != nil {
		t.Fatalf("get roadmap: %v", err)
	}
	if gotRoadmap.Path != roadmap.Path || gotRoadmap.Metadata["type"] != "project" {
		t.Fatalf("roadmap document = %+v", gotRoadmap)
	}

	list, err := client.ListDocuments(ctx, DocumentListOptions{PathPrefix: "notes/", Limit: 10})
	if err != nil {
		t.Fatalf("list documents: %v", err)
	}
	if len(list.Documents) != 2 {
		t.Fatalf("listed documents = %d, want 2", len(list.Documents))
	}

	search, err := client.Search(ctx, SearchOptions{
		Text:          "roadmap",
		MetadataKey:   "type",
		MetadataValue: "project",
		Limit:         5,
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(search.Hits) == 0 || search.Hits[0].DocID != roadmap.DocID || len(search.Hits[0].Citations) == 0 {
		t.Fatalf("search result = %+v, want roadmap hit with citation", search)
	}

	appended, err := client.AppendDocument(ctx, roadmap.DocID, "## Decisions\nUse the local SDK facade for agent tasks.")
	if err != nil {
		t.Fatalf("append document: %v", err)
	}
	if !strings.Contains(appended.Body, "local SDK facade") {
		t.Fatalf("appended body missing content: %q", appended.Body)
	}

	replaced, err := client.ReplaceSection(ctx, roadmap.DocID, "Decisions", "Use `OpenClient` for routine local workflows.")
	if err != nil {
		t.Fatalf("replace section: %v", err)
	}
	if !strings.Contains(replaced.Body, "Use `OpenClient`") || strings.Contains(replaced.Body, "local SDK facade") {
		t.Fatalf("replaced body = %q", replaced.Body)
	}

	links, err := client.GetDocumentLinks(ctx, roadmap.DocID)
	if err != nil {
		t.Fatalf("document links: %v", err)
	}
	if len(links.Outgoing) != 1 || links.Outgoing[0].DocID != architecture.DocID {
		t.Fatalf("links = %+v, want outgoing architecture link", links)
	}

	neighborhood, err := client.GraphNeighborhood(ctx, GraphNeighborhoodOptions{DocID: roadmap.DocID, Limit: 10})
	if err != nil {
		t.Fatalf("graph neighborhood: %v", err)
	}
	if len(neighborhood.Nodes) == 0 || len(neighborhood.Edges) == 0 {
		t.Fatalf("graph neighborhood = %+v, want nodes and edges", neighborhood)
	}

	lookup, err := client.LookupRecords(ctx, RecordLookupOptions{Text: "solenoid", Limit: 10})
	if err != nil {
		t.Fatalf("records lookup: %v", err)
	}
	if len(lookup.Entities) != 1 || lookup.Entities[0].EntityID != "transmission-solenoid" {
		t.Fatalf("records lookup = %+v, want transmission-solenoid", lookup)
	}

	entity, err := client.GetRecordEntity(ctx, lookup.Entities[0].EntityID)
	if err != nil {
		t.Fatalf("get record entity: %v", err)
	}
	if entity.EntityID != "transmission-solenoid" || len(entity.Facts) != 2 || len(entity.Citations) == 0 {
		t.Fatalf("record entity = %+v", entity)
	}

	events, err := client.ListProvenanceEvents(ctx, ProvenanceEventOptions{
		RefKind: "document",
		RefID:   roadmap.DocID,
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("list provenance events: %v", err)
	}
	if len(events.Events) == 0 {
		t.Fatal("expected roadmap provenance events")
	}

	projections, err := client.ListProjectionStates(ctx, ProjectionStateOptions{
		Projection: "graph",
		RefKind:    "document",
		RefID:      roadmap.DocID,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("list projection states: %v", err)
	}
	if len(projections.Projections) != 1 || projections.Projections[0].Freshness != "fresh" {
		t.Fatalf("graph projections = %+v, want one fresh projection", projections)
	}

	if record.DocID == "" {
		t.Fatal("record document id is empty")
	}
}

func TestOpenClientErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client, err := OpenClient(Config{DataDir: filepath.Join(t.TempDir(), "data")})
	if err != nil {
		t.Fatalf("open client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	_, err = client.Search(ctx, SearchOptions{})
	assertLocalError(t, err, "validation_error", 400)

	_, err = client.GetDocument(ctx, "missing")
	assertLocalError(t, err, "not_found", 404)

	_, err = client.CreateDocument(ctx, DocumentInput{
		Path:  "docs/duplicate.md",
		Title: "Duplicate",
		Body:  "# Duplicate\n",
	})
	if err != nil {
		t.Fatalf("create duplicate baseline: %v", err)
	}
	_, err = client.CreateDocument(ctx, DocumentInput{
		Path:  "docs/duplicate.md",
		Title: "Duplicate Again",
		Body:  "# Duplicate Again\n",
	})
	assertLocalError(t, err, "already_exists", 409)

	ftsRuntime, err := newRuntime(domain.BackendFTS, Config{DataDir: filepath.Join(t.TempDir(), "fts")})
	if err != nil {
		t.Fatalf("open fts runtime: %v", err)
	}
	t.Cleanup(func() { _ = ftsRuntime.Close() })
	ftsClient := &Client{runtime: ftsRuntime}
	_, err = ftsClient.GraphNeighborhood(ctx, GraphNeighborhoodOptions{DocID: "doc"})
	assertLocalError(t, err, "unsupported", 404)
}

func assertLocalError(t *testing.T, err error, code string, status int) {
	t.Helper()

	var localErr *Error
	if !errors.As(err, &localErr) {
		t.Fatalf("error = %v, want *local.Error", err)
	}
	if localErr.Code != code || localErr.Status != status {
		t.Fatalf("local error = %+v, want code=%s status=%d", localErr, code, status)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
