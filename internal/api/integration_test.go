package api_test

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	ftsclient "github.com/yazanabuashour/openclerk/client/fts"
	graphclient "github.com/yazanabuashour/openclerk/client/graph"
	hybridclient "github.com/yazanabuashour/openclerk/client/hybrid"
	openclerkclient "github.com/yazanabuashour/openclerk/client/openclerk"
	recordsclient "github.com/yazanabuashour/openclerk/client/records"
	"github.com/yazanabuashour/openclerk/internal/api"
	"github.com/yazanabuashour/openclerk/internal/app"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/infra/sqlite"
)

type capabilitiesInfo struct {
	backend    string
	searchMode []string
	extensions []string
}

type documentInfo struct {
	docID string
	path  string
}

type searchHitInfo struct {
	chunkID string
	docID   string
	path    string
}

type searchInfo struct {
	hits       []searchHitInfo
	hasMore    bool
	nextCursor string
}

type coreDriver struct {
	name           string
	backend        domain.BackendKind
	provider       string
	newClient      func(t *testing.T, serverURL string) any
	capabilities   func(t *testing.T, client any) capabilitiesInfo
	createDocument func(t *testing.T, client any, path, title, body string) documentInfo
	search         func(t *testing.T, client any, text string, limit int, cursor string) searchInfo
	getDocument    func(t *testing.T, client any, docID string) documentInfo
	append         func(t *testing.T, client any, docID string, content string) documentInfo
	replaceSection func(t *testing.T, client any, docID string, heading string, content string) documentInfo
	getChunk       func(t *testing.T, client any, chunkID string) searchHitInfo
}

func TestCoreBackends(t *testing.T) {
	t.Parallel()

	drivers := []coreDriver{
		ftsDriver(),
		hybridDriver(""),
		graphDriver(),
		recordsDriver(),
	}

	for _, driver := range drivers {
		driver := driver
		t.Run(driver.name, func(t *testing.T) {
			t.Parallel()

			serverURL := newTestServer(t, driver.backend, driver.provider)
			client := driver.newClient(t, serverURL)

			capabilities := driver.capabilities(t, client)
			if capabilities.backend != string(driver.backend) {
				t.Fatalf("capabilities backend = %q, want %q", capabilities.backend, driver.backend)
			}

			document := driver.createDocument(t, client, "docs/transmission.md", "Transmission solenoid", strings.TrimSpace(`
# Transmission solenoid

## Overview
Alpha exact lookup term.

## Facts
- sku: SOL-1

## Appendix
Stable beta anchor.
`))

			if document.docID == "" {
				t.Fatal("create document returned empty docID")
			}

			alphaResult := driver.search(t, client, "alpha", 1, "")
			if len(alphaResult.hits) != 1 {
				t.Fatalf("alpha search hit count = %d, want 1", len(alphaResult.hits))
			}
			if alphaResult.hits[0].docID != document.docID {
				t.Fatalf("alpha search docID = %q, want %q", alphaResult.hits[0].docID, document.docID)
			}

			betaResult := driver.search(t, client, "beta", 1, "")
			if len(betaResult.hits) != 1 {
				t.Fatalf("beta search hit count = %d, want 1", len(betaResult.hits))
			}
			stableChunkID := betaResult.hits[0].chunkID

			gotDocument := driver.getDocument(t, client, document.docID)
			if gotDocument.docID != document.docID || gotDocument.path != document.path {
				t.Fatalf("get document = %+v, want docID=%q path=%q", gotDocument, document.docID, document.path)
			}

			chunk := driver.getChunk(t, client, stableChunkID)
			if chunk.chunkID != stableChunkID || chunk.path != document.path {
				t.Fatalf("get chunk = %+v, want chunkID=%q path=%q", chunk, stableChunkID, document.path)
			}

			driver.replaceSection(t, client, document.docID, "Overview", "Gamma updated overview.\n")
			betaAfterReplace := driver.search(t, client, "beta", 1, "")
			if len(betaAfterReplace.hits) != 1 {
				t.Fatalf("beta search after replace hit count = %d, want 1", len(betaAfterReplace.hits))
			}
			if betaAfterReplace.hits[0].chunkID != stableChunkID {
				t.Fatalf("appendix chunk changed after unrelated section replacement: got %q want %q", betaAfterReplace.hits[0].chunkID, stableChunkID)
			}

			driver.append(t, client, document.docID, "## Notes\nDelta appended detail.\n")
			deltaResult := driver.search(t, client, "delta", 1, "")
			if len(deltaResult.hits) != 1 {
				t.Fatalf("delta search hit count = %d, want 1", len(deltaResult.hits))
			}
			betaAfterAppend := driver.search(t, client, "beta", 1, "")
			if len(betaAfterAppend.hits) != 1 {
				t.Fatalf("beta search after append hit count = %d, want 1", len(betaAfterAppend.hits))
			}
			if betaAfterAppend.hits[0].chunkID != stableChunkID {
				t.Fatalf("appendix chunk changed after append: got %q want %q", betaAfterAppend.hits[0].chunkID, stableChunkID)
			}
		})
	}
}

func TestHybridCapabilitiesFallbackAndLocalProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		want     []string
	}{
		{name: "lexical-only", provider: "", want: []string{"lexical"}},
		{name: "local-provider", provider: "local", want: []string{"lexical", "vector", "hybrid"}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			serverURL := newTestServer(t, domain.BackendHybrid, tc.provider)
			client := hybridclientClient(t, serverURL)
			got := hybridCapabilities(t, client)
			if !slices.Equal(got.searchMode, tc.want) {
				t.Fatalf("search modes = %v, want %v", got.searchMode, tc.want)
			}
		})
	}
}

func TestGraphExtension(t *testing.T) {
	t.Parallel()

	serverURL := newTestServer(t, domain.BackendGraph, "")
	client := graphclientClient(t, serverURL)
	source := graphCreateDocument(t, client, "docs/guide.md", "Guide", strings.TrimSpace(`
# Guide

## Overview
See the [reference](reference.md) for details.
`))
	target := graphCreateDocument(t, client, "docs/reference.md", "Reference", strings.TrimSpace(`
# Reference

## Overview
Canonical supporting note.
`))

	response, err := client.GraphNeighborhoodWithResponse(context.Background(), graphclient.GraphNeighborhoodRequest{DocId: &source.docID})
	if err != nil {
		t.Fatalf("graph neighborhood: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("graph neighborhood default response: %s", string(response.Body))
	}
	foundTarget := false
	foundLink := false
	for _, node := range response.JSON200.Nodes {
		if node.NodeId == "doc:"+target.docID {
			foundTarget = true
		}
	}
	for _, edge := range response.JSON200.Edges {
		if edge.Kind == graphclient.LinksTo && edge.ToNodeId == "doc:"+target.docID {
			foundLink = true
		}
	}
	if !foundTarget || !foundLink {
		t.Fatalf("graph neighborhood missing target node or link edge: nodes=%v edges=%v", response.JSON200.Nodes, response.JSON200.Edges)
	}
}

func TestRecordsExtensionInvalidatesOnDocumentUpdate(t *testing.T) {
	t.Parallel()

	serverURL := newTestServer(t, domain.BackendRecords, "")
	client := recordsclientClient(t, serverURL)
	document := recordsCreateDocument(t, client, "records/solenoid.md", "Transmission solenoid", strings.TrimSpace(`
---
entity_type: part
entity_name: Transmission solenoid
entity_id: transmission-solenoid
---
# Transmission solenoid

## Summary
Canonical part record.

## Facts
- sku: SOL-1
- vendor: ACME
`))

	lookup, err := client.RecordsLookupWithResponse(context.Background(), recordsclient.RecordsLookupRequest{Text: "solenoid"})
	if err != nil {
		t.Fatalf("records lookup: %v", err)
	}
	if lookup.JSON200 == nil || len(lookup.JSON200.Entities) != 1 {
		t.Fatalf("records lookup response = %#v", lookup.JSON200)
	}
	entityID := lookup.JSON200.Entities[0].EntityId

	recordsReplaceSection(t, client, document.docID, "Facts", "- sku: SOL-2\n- vendor: OpenClerk Motors\n")
	entity, err := client.GetRecordEntityWithResponse(context.Background(), entityID)
	if err != nil {
		t.Fatalf("get record entity: %v", err)
	}
	if entity.JSON200 == nil {
		t.Fatalf("get record entity default response: %s", string(entity.Body))
	}
	facts := map[string]string{}
	for _, fact := range entity.JSON200.Facts {
		facts[fact.Key] = fact.Value
	}
	if facts["sku"] != "SOL-2" || facts["vendor"] != "OpenClerk Motors" {
		t.Fatalf("record facts after update = %v", facts)
	}
}

func TestOpenClerkUnifiedSurface(t *testing.T) {
	t.Parallel()

	serverURL := newTestServer(t, domain.BackendOpenClerk, "local")
	client := openclerkclientClient(t, serverURL)

	source := openclerkCreateDocument(t, client, "notes/architecture/knowledge-plane.md", "Knowledge plane", strings.TrimSpace(`
---
type: note
status: active
---
# Knowledge plane

## Summary
Canonical architecture note.
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

	links, err := client.GetDocumentLinksWithResponse(context.Background(), target.docID)
	if err != nil {
		t.Fatalf("document links: %v", err)
	}
	if links.JSON200 == nil || len(links.JSON200.Outgoing) != 1 || links.JSON200.Outgoing[0].DocId != source.docID {
		t.Fatalf("document links response = %#v", links.JSON200)
	}

	lookup, err := client.RecordsLookupWithResponse(context.Background(), openclerkclient.RecordsLookupRequest{Text: "solenoid"})
	if err != nil {
		t.Fatalf("records lookup: %v", err)
	}
	if lookup.JSON200 == nil || len(lookup.JSON200.Entities) != 1 {
		t.Fatalf("records lookup response = %#v", lookup.JSON200)
	}

	projection := "records"
	projections, err := client.ListProjectionStatesWithResponse(context.Background(), &openclerkclient.ListProjectionStatesParams{Projection: &projection})
	if err != nil {
		t.Fatalf("projection states: %v", err)
	}
	if projections.JSON200 == nil || len(projections.JSON200.Projections) == 0 {
		t.Fatalf("projection states response = %#v", projections.JSON200)
	}
}

func newTestServer(t *testing.T, backend domain.BackendKind, provider string) string {
	t.Helper()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store, err := sqlite.New(context.Background(), sqlite.Config{
		Backend:           backend,
		DatabasePath:      dbPath,
		VaultRoot:         vaultRoot,
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

func ftsDriver() coreDriver {
	return coreDriver{
		name:      "fts",
		backend:   domain.BackendFTS,
		newClient: func(t *testing.T, serverURL string) any { return ftsclientClient(t, serverURL) },
		capabilities: func(t *testing.T, client any) capabilitiesInfo {
			return ftsCapabilities(t, client.(*ftsclient.ClientWithResponses))
		},
		createDocument: func(t *testing.T, client any, path, title, body string) documentInfo {
			return ftsCreateDocument(t, client.(*ftsclient.ClientWithResponses), path, title, body)
		},
		search: func(t *testing.T, client any, text string, limit int, cursor string) searchInfo {
			return ftsSearch(t, client.(*ftsclient.ClientWithResponses), text, limit, cursor)
		},
		getDocument: func(t *testing.T, client any, docID string) documentInfo {
			return ftsGetDocument(t, client.(*ftsclient.ClientWithResponses), docID)
		},
		append: func(t *testing.T, client any, docID string, content string) documentInfo {
			return ftsAppend(t, client.(*ftsclient.ClientWithResponses), docID, content)
		},
		replaceSection: func(t *testing.T, client any, docID string, heading string, content string) documentInfo {
			return ftsReplaceSection(t, client.(*ftsclient.ClientWithResponses), docID, heading, content)
		},
		getChunk: func(t *testing.T, client any, chunkID string) searchHitInfo {
			return ftsGetChunk(t, client.(*ftsclient.ClientWithResponses), chunkID)
		},
	}
}

func hybridDriver(provider string) coreDriver {
	return coreDriver{
		name:      "hybrid",
		backend:   domain.BackendHybrid,
		provider:  provider,
		newClient: func(t *testing.T, serverURL string) any { return hybridclientClient(t, serverURL) },
		capabilities: func(t *testing.T, client any) capabilitiesInfo {
			return hybridCapabilities(t, client.(*hybridclient.ClientWithResponses))
		},
		createDocument: func(t *testing.T, client any, path, title, body string) documentInfo {
			return hybridCreateDocument(t, client.(*hybridclient.ClientWithResponses), path, title, body)
		},
		search: func(t *testing.T, client any, text string, limit int, cursor string) searchInfo {
			return hybridSearch(t, client.(*hybridclient.ClientWithResponses), text, limit, cursor)
		},
		getDocument: func(t *testing.T, client any, docID string) documentInfo {
			return hybridGetDocument(t, client.(*hybridclient.ClientWithResponses), docID)
		},
		append: func(t *testing.T, client any, docID string, content string) documentInfo {
			return hybridAppend(t, client.(*hybridclient.ClientWithResponses), docID, content)
		},
		replaceSection: func(t *testing.T, client any, docID string, heading string, content string) documentInfo {
			return hybridReplaceSection(t, client.(*hybridclient.ClientWithResponses), docID, heading, content)
		},
		getChunk: func(t *testing.T, client any, chunkID string) searchHitInfo {
			return hybridGetChunk(t, client.(*hybridclient.ClientWithResponses), chunkID)
		},
	}
}

func openclerkclientClient(t *testing.T, serverURL string) *openclerkclient.ClientWithResponses {
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

func graphDriver() coreDriver {
	return coreDriver{
		name:      "graph",
		backend:   domain.BackendGraph,
		newClient: func(t *testing.T, serverURL string) any { return graphclientClient(t, serverURL) },
		capabilities: func(t *testing.T, client any) capabilitiesInfo {
			return graphCapabilities(t, client.(*graphclient.ClientWithResponses))
		},
		createDocument: func(t *testing.T, client any, path, title, body string) documentInfo {
			return graphCreateDocument(t, client.(*graphclient.ClientWithResponses), path, title, body)
		},
		search: func(t *testing.T, client any, text string, limit int, cursor string) searchInfo {
			return graphSearch(t, client.(*graphclient.ClientWithResponses), text, limit, cursor)
		},
		getDocument: func(t *testing.T, client any, docID string) documentInfo {
			return graphGetDocument(t, client.(*graphclient.ClientWithResponses), docID)
		},
		append: func(t *testing.T, client any, docID string, content string) documentInfo {
			return graphAppend(t, client.(*graphclient.ClientWithResponses), docID, content)
		},
		replaceSection: func(t *testing.T, client any, docID string, heading string, content string) documentInfo {
			return graphReplaceSection(t, client.(*graphclient.ClientWithResponses), docID, heading, content)
		},
		getChunk: func(t *testing.T, client any, chunkID string) searchHitInfo {
			return graphGetChunk(t, client.(*graphclient.ClientWithResponses), chunkID)
		},
	}
}

func recordsDriver() coreDriver {
	return coreDriver{
		name:      "records",
		backend:   domain.BackendRecords,
		newClient: func(t *testing.T, serverURL string) any { return recordsclientClient(t, serverURL) },
		capabilities: func(t *testing.T, client any) capabilitiesInfo {
			return recordsCapabilities(t, client.(*recordsclient.ClientWithResponses))
		},
		createDocument: func(t *testing.T, client any, path, title, body string) documentInfo {
			return recordsCreateDocument(t, client.(*recordsclient.ClientWithResponses), path, title, body)
		},
		search: func(t *testing.T, client any, text string, limit int, cursor string) searchInfo {
			return recordsSearch(t, client.(*recordsclient.ClientWithResponses), text, limit, cursor)
		},
		getDocument: func(t *testing.T, client any, docID string) documentInfo {
			return recordsGetDocument(t, client.(*recordsclient.ClientWithResponses), docID)
		},
		append: func(t *testing.T, client any, docID string, content string) documentInfo {
			return recordsAppend(t, client.(*recordsclient.ClientWithResponses), docID, content)
		},
		replaceSection: func(t *testing.T, client any, docID string, heading string, content string) documentInfo {
			return recordsReplaceSection(t, client.(*recordsclient.ClientWithResponses), docID, heading, content)
		},
		getChunk: func(t *testing.T, client any, chunkID string) searchHitInfo {
			return recordsGetChunk(t, client.(*recordsclient.ClientWithResponses), chunkID)
		},
	}
}
