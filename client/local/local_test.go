package local_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	ftsclient "github.com/yazanabuashour/openclerk/client/fts"
	graphclient "github.com/yazanabuashour/openclerk/client/graph"
	hybridclient "github.com/yazanabuashour/openclerk/client/hybrid"
	localclient "github.com/yazanabuashour/openclerk/client/local"
	recordsclient "github.com/yazanabuashour/openclerk/client/records"
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

type localDriver struct {
	name           string
	embedding      string
	open           func(t *testing.T, dataDir, embedding string) any
	capabilities   func(t *testing.T, client any) capabilitiesInfo
	createDocument func(t *testing.T, client any, path, title, body string) documentInfo
	search         func(t *testing.T, client any, text string, limit int, cursor string) searchInfo
	getDocument    func(t *testing.T, client any, docID string) documentInfo
	append         func(t *testing.T, client any, docID string, content string) documentInfo
	replaceSection func(t *testing.T, client any, docID string, heading string, content string) documentInfo
	getChunk       func(t *testing.T, client any, chunkID string) searchHitInfo
}

func TestOpenFTSDefaultStorage(t *testing.T) {
	xdgDataHome := filepath.Join(t.TempDir(), "xdg")
	t.Setenv("XDG_DATA_HOME", xdgDataHome)

	paths, err := localclient.ResolvePaths(localclient.Config{})
	if err != nil {
		t.Fatalf("resolve default paths: %v", err)
	}
	wantDataDir := filepath.Join(xdgDataHome, "openclerk")
	if paths.DataDir != wantDataDir {
		t.Fatalf("data dir = %q, want %q", paths.DataDir, wantDataDir)
	}
	if paths.DatabasePath != filepath.Join(wantDataDir, "openclerk.sqlite") {
		t.Fatalf("database path = %q", paths.DatabasePath)
	}
	if paths.VaultRoot != filepath.Join(wantDataDir, "vault") {
		t.Fatalf("vault root = %q", paths.VaultRoot)
	}

	client, runtime, err := localclient.OpenFTS(localclient.Config{})
	if err != nil {
		t.Fatalf("open fts: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })

	if runtime.Paths() != paths {
		t.Fatalf("runtime paths = %+v, want %+v", runtime.Paths(), paths)
	}

	document := ftsCreateDocument(t, client, "health/default-storage.md", "Default storage", strings.TrimSpace(`
# Default storage

## Summary
Health data stored under the XDG data directory.
`))
	if _, err := os.Stat(paths.DatabasePath); err != nil {
		t.Fatalf("stat sqlite database: %v", err)
	}
	if _, err := os.Stat(filepath.Join(paths.VaultRoot, filepath.FromSlash(document.path))); err != nil {
		t.Fatalf("stat canonical document: %v", err)
	}
}

func TestCoreBackends(t *testing.T) {
	t.Parallel()

	drivers := []localDriver{
		ftsDriver(),
		hybridDriver(""),
		graphDriver(),
		recordsDriver(),
	}

	for _, driver := range drivers {
		driver := driver
		t.Run(driver.name, func(t *testing.T) {
			t.Parallel()

			dataDir := filepath.Join(t.TempDir(), "data")
			client := driver.open(t, dataDir, driver.embedding)

			capabilities := driver.capabilities(t, client)
			if capabilities.backend != driver.name {
				t.Fatalf("capabilities backend = %q, want %q", capabilities.backend, driver.name)
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
		name      string
		embedding string
		want      []string
	}{
		{name: "lexical-only", embedding: "", want: []string{"lexical"}},
		{name: "local-provider", embedding: "local", want: []string{"lexical", "vector", "hybrid"}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, runtime, err := localclient.OpenHybrid(localclient.Config{
				DataDir:           filepath.Join(t.TempDir(), "data"),
				EmbeddingProvider: tc.embedding,
			})
			if err != nil {
				t.Fatalf("open hybrid client: %v", err)
			}
			t.Cleanup(func() { _ = runtime.Close() })

			got := hybridCapabilities(t, client)
			if !slices.Equal(got.searchMode, tc.want) {
				t.Fatalf("search modes = %v, want %v", got.searchMode, tc.want)
			}
		})
	}
}

func TestGraphExtension(t *testing.T) {
	t.Parallel()

	client, runtime, err := localclient.OpenGraph(localclient.Config{DataDir: filepath.Join(t.TempDir(), "data")})
	if err != nil {
		t.Fatalf("open graph client: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })

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

	limit := 8
	response, err := client.GraphNeighborhoodWithResponse(context.Background(), graphclient.GraphNeighborhoodRequest{DocId: &source.docID, Limit: &limit})
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

	client, runtime, err := localclient.OpenRecords(localclient.Config{DataDir: filepath.Join(t.TempDir(), "data")})
	if err != nil {
		t.Fatalf("open records client: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })

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

func ftsDriver() localDriver {
	return localDriver{
		name: "fts",
		open: func(t *testing.T, dataDir, _ string) any {
			t.Helper()
			client, runtime, err := localclient.OpenFTS(localclient.Config{DataDir: dataDir})
			if err != nil {
				t.Fatalf("open fts client: %v", err)
			}
			t.Cleanup(func() { _ = runtime.Close() })
			return client
		},
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

func hybridDriver(embedding string) localDriver {
	return localDriver{
		name:      "hybrid",
		embedding: embedding,
		open: func(t *testing.T, dataDir, embedding string) any {
			t.Helper()
			client, runtime, err := localclient.OpenHybrid(localclient.Config{
				DataDir:           dataDir,
				EmbeddingProvider: embedding,
			})
			if err != nil {
				t.Fatalf("open hybrid client: %v", err)
			}
			t.Cleanup(func() { _ = runtime.Close() })
			return client
		},
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

func graphDriver() localDriver {
	return localDriver{
		name: "graph",
		open: func(t *testing.T, dataDir, _ string) any {
			t.Helper()
			client, runtime, err := localclient.OpenGraph(localclient.Config{DataDir: dataDir})
			if err != nil {
				t.Fatalf("open graph client: %v", err)
			}
			t.Cleanup(func() { _ = runtime.Close() })
			return client
		},
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

func recordsDriver() localDriver {
	return localDriver{
		name: "records",
		open: func(t *testing.T, dataDir, _ string) any {
			t.Helper()
			client, runtime, err := localclient.OpenRecords(localclient.Config{DataDir: dataDir})
			if err != nil {
				t.Fatalf("open records client: %v", err)
			}
			t.Cleanup(func() { _ = runtime.Close() })
			return client
		},
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

func ftsCapabilities(t *testing.T, client *ftsclient.ClientWithResponses) capabilitiesInfo {
	t.Helper()
	response, err := client.GetCapabilitiesWithResponse(context.Background())
	if err != nil {
		t.Fatalf("fts capabilities: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("fts capabilities error: %s", string(response.Body))
	}
	return capabilitiesInfo{
		backend:    string(response.JSON200.Backend),
		searchMode: enumStrings(response.JSON200.SearchModes),
		extensions: enumStrings(response.JSON200.Extensions),
	}
}

func ftsCreateDocument(t *testing.T, client *ftsclient.ClientWithResponses, path, title, body string) documentInfo {
	t.Helper()
	response, err := client.CreateDocumentWithResponse(context.Background(), ftsclient.CreateDocumentRequest{
		Path:  path,
		Title: title,
		Body:  body,
	})
	if err != nil {
		t.Fatalf("fts create document: %v", err)
	}
	if response.JSON201 == nil {
		t.Fatalf("fts create document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON201.DocId, path: response.JSON201.Path}
}

func ftsSearch(t *testing.T, client *ftsclient.ClientWithResponses, text string, limit int, cursor string) searchInfo {
	t.Helper()
	request := ftsclient.SearchQuery{Text: text}
	if limit > 0 {
		request.Limit = &limit
	}
	if cursor != "" {
		request.Cursor = &cursor
	}
	response, err := client.SearchQueryWithResponse(context.Background(), request)
	if err != nil {
		t.Fatalf("fts search query: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("fts search query error: %s", string(response.Body))
	}
	return ftsSearchResponse(*response.JSON200)
}

func ftsGetDocument(t *testing.T, client *ftsclient.ClientWithResponses, docID string) documentInfo {
	t.Helper()
	response, err := client.GetDocumentWithResponse(context.Background(), docID)
	if err != nil {
		t.Fatalf("fts get document: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("fts get document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func ftsAppend(t *testing.T, client *ftsclient.ClientWithResponses, docID string, content string) documentInfo {
	t.Helper()
	response, err := client.AppendDocumentWithResponse(context.Background(), docID, ftsclient.AppendDocumentRequest{Content: content})
	if err != nil {
		t.Fatalf("fts append document: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("fts append document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func ftsReplaceSection(t *testing.T, client *ftsclient.ClientWithResponses, docID string, heading string, content string) documentInfo {
	t.Helper()
	response, err := client.ReplaceDocumentSectionWithResponse(context.Background(), docID, ftsclient.ReplaceSectionRequest{
		Heading: heading,
		Content: content,
	})
	if err != nil {
		t.Fatalf("fts replace section: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("fts replace section error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func ftsGetChunk(t *testing.T, client *ftsclient.ClientWithResponses, chunkID string) searchHitInfo {
	t.Helper()
	response, err := client.GetChunkWithResponse(context.Background(), chunkID)
	if err != nil {
		t.Fatalf("fts get chunk: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("fts get chunk error: %s", string(response.Body))
	}
	return searchHitInfo{chunkID: response.JSON200.ChunkId, docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func hybridCapabilities(t *testing.T, client *hybridclient.ClientWithResponses) capabilitiesInfo {
	t.Helper()
	response, err := client.GetCapabilitiesWithResponse(context.Background())
	if err != nil {
		t.Fatalf("hybrid capabilities: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("hybrid capabilities error: %s", string(response.Body))
	}
	return capabilitiesInfo{
		backend:    string(response.JSON200.Backend),
		searchMode: enumStrings(response.JSON200.SearchModes),
		extensions: enumStrings(response.JSON200.Extensions),
	}
}

func hybridCreateDocument(t *testing.T, client *hybridclient.ClientWithResponses, path, title, body string) documentInfo {
	t.Helper()
	response, err := client.CreateDocumentWithResponse(context.Background(), hybridclient.CreateDocumentRequest{
		Path:  path,
		Title: title,
		Body:  body,
	})
	if err != nil {
		t.Fatalf("hybrid create document: %v", err)
	}
	if response.JSON201 == nil {
		t.Fatalf("hybrid create document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON201.DocId, path: response.JSON201.Path}
}

func hybridSearch(t *testing.T, client *hybridclient.ClientWithResponses, text string, limit int, cursor string) searchInfo {
	t.Helper()
	request := hybridclient.SearchQuery{Text: text}
	if limit > 0 {
		request.Limit = &limit
	}
	if cursor != "" {
		request.Cursor = &cursor
	}
	response, err := client.SearchQueryWithResponse(context.Background(), request)
	if err != nil {
		t.Fatalf("hybrid search query: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("hybrid search query error: %s", string(response.Body))
	}
	return hybridSearchResponse(*response.JSON200)
}

func hybridGetDocument(t *testing.T, client *hybridclient.ClientWithResponses, docID string) documentInfo {
	t.Helper()
	response, err := client.GetDocumentWithResponse(context.Background(), docID)
	if err != nil {
		t.Fatalf("hybrid get document: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("hybrid get document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func hybridAppend(t *testing.T, client *hybridclient.ClientWithResponses, docID string, content string) documentInfo {
	t.Helper()
	response, err := client.AppendDocumentWithResponse(context.Background(), docID, hybridclient.AppendDocumentRequest{Content: content})
	if err != nil {
		t.Fatalf("hybrid append document: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("hybrid append document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func hybridReplaceSection(t *testing.T, client *hybridclient.ClientWithResponses, docID string, heading string, content string) documentInfo {
	t.Helper()
	response, err := client.ReplaceDocumentSectionWithResponse(context.Background(), docID, hybridclient.ReplaceSectionRequest{
		Heading: heading,
		Content: content,
	})
	if err != nil {
		t.Fatalf("hybrid replace section: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("hybrid replace section error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func hybridGetChunk(t *testing.T, client *hybridclient.ClientWithResponses, chunkID string) searchHitInfo {
	t.Helper()
	response, err := client.GetChunkWithResponse(context.Background(), chunkID)
	if err != nil {
		t.Fatalf("hybrid get chunk: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("hybrid get chunk error: %s", string(response.Body))
	}
	return searchHitInfo{chunkID: response.JSON200.ChunkId, docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func graphCapabilities(t *testing.T, client *graphclient.ClientWithResponses) capabilitiesInfo {
	t.Helper()
	response, err := client.GetCapabilitiesWithResponse(context.Background())
	if err != nil {
		t.Fatalf("graph capabilities: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("graph capabilities error: %s", string(response.Body))
	}
	return capabilitiesInfo{
		backend:    string(response.JSON200.Backend),
		searchMode: enumStrings(response.JSON200.SearchModes),
		extensions: enumStrings(response.JSON200.Extensions),
	}
}

func graphCreateDocument(t *testing.T, client *graphclient.ClientWithResponses, path, title, body string) documentInfo {
	t.Helper()
	response, err := client.CreateDocumentWithResponse(context.Background(), graphclient.CreateDocumentRequest{
		Path:  path,
		Title: title,
		Body:  body,
	})
	if err != nil {
		t.Fatalf("graph create document: %v", err)
	}
	if response.JSON201 == nil {
		t.Fatalf("graph create document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON201.DocId, path: response.JSON201.Path}
}

func graphSearch(t *testing.T, client *graphclient.ClientWithResponses, text string, limit int, cursor string) searchInfo {
	t.Helper()
	request := graphclient.SearchQuery{Text: text}
	if limit > 0 {
		request.Limit = &limit
	}
	if cursor != "" {
		request.Cursor = &cursor
	}
	response, err := client.SearchQueryWithResponse(context.Background(), request)
	if err != nil {
		t.Fatalf("graph search query: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("graph search query error: %s", string(response.Body))
	}
	return graphSearchResponse(*response.JSON200)
}

func graphGetDocument(t *testing.T, client *graphclient.ClientWithResponses, docID string) documentInfo {
	t.Helper()
	response, err := client.GetDocumentWithResponse(context.Background(), docID)
	if err != nil {
		t.Fatalf("graph get document: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("graph get document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func graphAppend(t *testing.T, client *graphclient.ClientWithResponses, docID string, content string) documentInfo {
	t.Helper()
	response, err := client.AppendDocumentWithResponse(context.Background(), docID, graphclient.AppendDocumentRequest{Content: content})
	if err != nil {
		t.Fatalf("graph append document: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("graph append document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func graphReplaceSection(t *testing.T, client *graphclient.ClientWithResponses, docID string, heading string, content string) documentInfo {
	t.Helper()
	response, err := client.ReplaceDocumentSectionWithResponse(context.Background(), docID, graphclient.ReplaceSectionRequest{
		Heading: heading,
		Content: content,
	})
	if err != nil {
		t.Fatalf("graph replace section: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("graph replace section error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func graphGetChunk(t *testing.T, client *graphclient.ClientWithResponses, chunkID string) searchHitInfo {
	t.Helper()
	response, err := client.GetChunkWithResponse(context.Background(), chunkID)
	if err != nil {
		t.Fatalf("graph get chunk: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("graph get chunk error: %s", string(response.Body))
	}
	return searchHitInfo{chunkID: response.JSON200.ChunkId, docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func recordsCapabilities(t *testing.T, client *recordsclient.ClientWithResponses) capabilitiesInfo {
	t.Helper()
	response, err := client.GetCapabilitiesWithResponse(context.Background())
	if err != nil {
		t.Fatalf("records capabilities: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("records capabilities error: %s", string(response.Body))
	}
	return capabilitiesInfo{
		backend:    string(response.JSON200.Backend),
		searchMode: enumStrings(response.JSON200.SearchModes),
		extensions: enumStrings(response.JSON200.Extensions),
	}
}

func recordsCreateDocument(t *testing.T, client *recordsclient.ClientWithResponses, path, title, body string) documentInfo {
	t.Helper()
	response, err := client.CreateDocumentWithResponse(context.Background(), recordsclient.CreateDocumentRequest{
		Path:  path,
		Title: title,
		Body:  body,
	})
	if err != nil {
		t.Fatalf("records create document: %v", err)
	}
	if response.JSON201 == nil {
		t.Fatalf("records create document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON201.DocId, path: response.JSON201.Path}
}

func recordsSearch(t *testing.T, client *recordsclient.ClientWithResponses, text string, limit int, cursor string) searchInfo {
	t.Helper()
	request := recordsclient.SearchQuery{Text: text}
	if limit > 0 {
		request.Limit = &limit
	}
	if cursor != "" {
		request.Cursor = &cursor
	}
	response, err := client.SearchQueryWithResponse(context.Background(), request)
	if err != nil {
		t.Fatalf("records search query: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("records search query error: %s", string(response.Body))
	}
	return recordsSearchResponse(*response.JSON200)
}

func recordsGetDocument(t *testing.T, client *recordsclient.ClientWithResponses, docID string) documentInfo {
	t.Helper()
	response, err := client.GetDocumentWithResponse(context.Background(), docID)
	if err != nil {
		t.Fatalf("records get document: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("records get document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func recordsAppend(t *testing.T, client *recordsclient.ClientWithResponses, docID string, content string) documentInfo {
	t.Helper()
	response, err := client.AppendDocumentWithResponse(context.Background(), docID, recordsclient.AppendDocumentRequest{Content: content})
	if err != nil {
		t.Fatalf("records append document: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("records append document error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func recordsReplaceSection(t *testing.T, client *recordsclient.ClientWithResponses, docID string, heading string, content string) documentInfo {
	t.Helper()
	response, err := client.ReplaceDocumentSectionWithResponse(context.Background(), docID, recordsclient.ReplaceSectionRequest{
		Heading: heading,
		Content: content,
	})
	if err != nil {
		t.Fatalf("records replace section: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("records replace section error: %s", string(response.Body))
	}
	return documentInfo{docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func recordsGetChunk(t *testing.T, client *recordsclient.ClientWithResponses, chunkID string) searchHitInfo {
	t.Helper()
	response, err := client.GetChunkWithResponse(context.Background(), chunkID)
	if err != nil {
		t.Fatalf("records get chunk: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("records get chunk error: %s", string(response.Body))
	}
	return searchHitInfo{chunkID: response.JSON200.ChunkId, docID: response.JSON200.DocId, path: response.JSON200.Path}
}

func enumStrings[T ~string](values []T) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value))
	}
	return result
}

func ftsSearchResponse(response ftsclient.SearchResponse) searchInfo {
	result := searchInfo{hasMore: response.PageInfo.HasMore}
	if response.PageInfo.NextCursor != nil {
		result.nextCursor = *response.PageInfo.NextCursor
	}
	for _, hit := range response.Hits {
		path := ""
		if len(hit.Citations) > 0 {
			path = hit.Citations[0].Path
		}
		result.hits = append(result.hits, searchHitInfo{chunkID: hit.ChunkId, docID: hit.DocId, path: path})
	}
	return result
}

func hybridSearchResponse(response hybridclient.SearchResponse) searchInfo {
	result := searchInfo{hasMore: response.PageInfo.HasMore}
	if response.PageInfo.NextCursor != nil {
		result.nextCursor = *response.PageInfo.NextCursor
	}
	for _, hit := range response.Hits {
		path := ""
		if len(hit.Citations) > 0 {
			path = hit.Citations[0].Path
		}
		result.hits = append(result.hits, searchHitInfo{chunkID: hit.ChunkId, docID: hit.DocId, path: path})
	}
	return result
}

func graphSearchResponse(response graphclient.SearchResponse) searchInfo {
	result := searchInfo{hasMore: response.PageInfo.HasMore}
	if response.PageInfo.NextCursor != nil {
		result.nextCursor = *response.PageInfo.NextCursor
	}
	for _, hit := range response.Hits {
		path := ""
		if len(hit.Citations) > 0 {
			path = hit.Citations[0].Path
		}
		result.hits = append(result.hits, searchHitInfo{chunkID: hit.ChunkId, docID: hit.DocId, path: path})
	}
	return result
}

func recordsSearchResponse(response recordsclient.SearchResponse) searchInfo {
	result := searchInfo{hasMore: response.PageInfo.HasMore}
	if response.PageInfo.NextCursor != nil {
		result.nextCursor = *response.PageInfo.NextCursor
	}
	for _, hit := range response.Hits {
		path := ""
		if len(hit.Citations) > 0 {
			path = hit.Citations[0].Path
		}
		result.hits = append(result.hits, searchHitInfo{chunkID: hit.ChunkId, docID: hit.DocId, path: path})
	}
	return result
}
