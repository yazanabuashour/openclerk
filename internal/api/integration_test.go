package api_test

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	openclerkclient "github.com/yazanabuashour/openclerk/client/openclerk"
	"github.com/yazanabuashour/openclerk/internal/api"
	"github.com/yazanabuashour/openclerk/internal/app"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/infra/sqlite"
)

func TestOpenClerkHTTPUnifiedSurface(t *testing.T) {
	t.Parallel()

	client := openclerkClient(t, newTestServer(t, "local"))
	capabilities, err := client.GetCapabilitiesWithResponse(context.Background())
	if err != nil {
		t.Fatalf("get capabilities: %v", err)
	}
	if capabilities.JSON200 == nil {
		t.Fatalf("capabilities error: %s", string(capabilities.Body))
	}
	if capabilities.JSON200.Backend != openclerkclient.Openclerk {
		t.Fatalf("backend = %q", capabilities.JSON200.Backend)
	}
	if !slices.Equal(enumStrings(capabilities.JSON200.SearchModes), []string{"lexical", "vector", "hybrid"}) {
		t.Fatalf("search modes = %v", capabilities.JSON200.SearchModes)
	}
	if !slices.Equal(enumStrings(capabilities.JSON200.Extensions), []string{"provenance", "graph", "records"}) {
		t.Fatalf("extensions = %v", capabilities.JSON200.Extensions)
	}

	source := openclerkCreateDocument(t, client, "notes/architecture/knowledge-plane.md", "Knowledge plane", strings.TrimSpace(`
---
type: note
status: active
---
# Knowledge plane

## Summary
Canonical agent-facing architecture note.
`))
	target := openclerkCreateDocument(t, client, "notes/projects/openclerk-roadmap.md", "Roadmap", strings.TrimSpace(`
---
type: project
status: active
---
# Roadmap

## Summary
See the [knowledge plane](../architecture/knowledge-plane.md).
`))
	openclerkCreateDocument(t, client, "records/assets/transmission-solenoid.md", "Transmission solenoid", strings.TrimSpace(`
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
`))

	pathPrefix := "notes/"
	list, err := client.ListDocumentsWithResponse(context.Background(), &openclerkclient.ListDocumentsParams{PathPrefix: &pathPrefix})
	if err != nil {
		t.Fatalf("list documents: %v", err)
	}
	if list.JSON200 == nil || len(list.JSON200.Documents) != 2 {
		t.Fatalf("list documents response = %#v", list.JSON200)
	}

	search, err := client.SearchQueryWithResponse(context.Background(), openclerkclient.SearchQuery{Text: "roadmap"})
	if err != nil {
		t.Fatalf("search query: %v", err)
	}
	if search.JSON200 == nil || len(search.JSON200.Hits) == 0 || search.JSON200.Hits[0].DocId != target.docID {
		t.Fatalf("search response = %#v", search.JSON200)
	}

	links, err := client.GetDocumentLinksWithResponse(context.Background(), target.docID)
	if err != nil {
		t.Fatalf("document links: %v", err)
	}
	if links.JSON200 == nil || len(links.JSON200.Outgoing) != 1 || links.JSON200.Outgoing[0].DocId != source.docID {
		t.Fatalf("document links response = %#v", links.JSON200)
	}

	neighborhood, err := client.GraphNeighborhoodWithResponse(context.Background(), openclerkclient.GraphNeighborhoodRequest{DocId: &target.docID})
	if err != nil {
		t.Fatalf("graph neighborhood: %v", err)
	}
	if neighborhood.JSON200 == nil || len(neighborhood.JSON200.Edges) == 0 {
		t.Fatalf("graph neighborhood response = %#v", neighborhood.JSON200)
	}

	lookup, err := client.RecordsLookupWithResponse(context.Background(), openclerkclient.RecordsLookupRequest{Text: "solenoid"})
	if err != nil {
		t.Fatalf("records lookup: %v", err)
	}
	if lookup.JSON200 == nil || len(lookup.JSON200.Entities) != 1 || lookup.JSON200.Entities[0].EntityId != "transmission-solenoid" {
		t.Fatalf("records lookup response = %#v", lookup.JSON200)
	}

	events, err := client.ListProvenanceEventsWithResponse(context.Background(), &openclerkclient.ListProvenanceEventsParams{
		RefKind: ptr("document"),
		RefId:   &target.docID,
	})
	if err != nil {
		t.Fatalf("list provenance events: %v", err)
	}
	if events.JSON200 == nil || len(events.JSON200.Events) == 0 {
		t.Fatalf("provenance events response = %#v", events.JSON200)
	}

	projections, err := client.ListProjectionStatesWithResponse(context.Background(), &openclerkclient.ListProjectionStatesParams{
		Projection: ptr("graph"),
		RefKind:    ptr("document"),
		RefId:      &target.docID,
	})
	if err != nil {
		t.Fatalf("list projection states: %v", err)
	}
	if projections.JSON200 == nil || len(projections.JSON200.Projections) != 1 || projections.JSON200.Projections[0].Freshness != openclerkclient.Fresh {
		t.Fatalf("projection states response = %#v", projections.JSON200)
	}
}

func TestOpenClerkHTTPLexicalFallbackWithoutEmbeddingProvider(t *testing.T) {
	t.Parallel()

	client := openclerkClient(t, newTestServer(t, ""))
	capabilities, err := client.GetCapabilitiesWithResponse(context.Background())
	if err != nil {
		t.Fatalf("get capabilities: %v", err)
	}
	if capabilities.JSON200 == nil {
		t.Fatalf("capabilities error: %s", string(capabilities.Body))
	}
	if !slices.Equal(enumStrings(capabilities.JSON200.SearchModes), []string{"lexical"}) {
		t.Fatalf("search modes = %v", capabilities.JSON200.SearchModes)
	}
}

func newTestServer(t *testing.T, provider string) string {
	t.Helper()

	store, err := sqlite.New(context.Background(), sqlite.Config{
		Backend:           domain.BackendOpenClerk,
		DatabasePath:      filepath.Join(t.TempDir(), "openclerk.sqlite"),
		VaultRoot:         t.TempDir(),
		EmbeddingProvider: provider,
	})
	if err != nil {
		t.Fatalf("create sqlite store: %v", err)
	}
	service := app.New(store)
	server := httptest.NewServer(api.NewHandler(service))
	t.Cleanup(func() {
		server.Close()
		_ = service.Close()
	})
	return server.URL
}

func openclerkClient(t *testing.T, serverURL string) *openclerkclient.ClientWithResponses {
	t.Helper()
	client, err := openclerkclient.NewClientWithResponses(serverURL)
	if err != nil {
		t.Fatalf("new openclerk client: %v", err)
	}
	return client
}

func openclerkCreateDocument(t *testing.T, client *openclerkclient.ClientWithResponses, path, title, body string) documentInfo {
	t.Helper()
	response, err := client.CreateDocumentWithResponse(context.Background(), openclerkclient.CreateDocumentRequest{
		Path:  path,
		Title: title,
		Body:  body + "\n",
	})
	if err != nil {
		t.Fatalf("openclerk create document: %v", err)
	}
	if response.JSON201 == nil {
		t.Fatalf("openclerk create document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON201.DocId, path: response.JSON201.Path}
}

type documentInfo struct {
	docID string
	path  string
}

func ptr(value string) *string {
	return &value
}

func enumStrings[T ~string](values []T) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value))
	}
	return result
}
