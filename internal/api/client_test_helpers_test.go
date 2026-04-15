package api_test

import (
	"context"
	"testing"

	ftsclient "github.com/yazanabuashour/openclerk/client/fts"
	graphclient "github.com/yazanabuashour/openclerk/client/graph"
	hybridclient "github.com/yazanabuashour/openclerk/client/hybrid"
	recordsclient "github.com/yazanabuashour/openclerk/client/records"
)

func ftsclientClient(t *testing.T, serverURL string) *ftsclient.ClientWithResponses {
	t.Helper()
	client, err := ftsclient.NewClientWithResponses(serverURL)
	if err != nil {
		t.Fatalf("new fts client: %v", err)
	}
	return client
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

func hybridclientClient(t *testing.T, serverURL string) *hybridclient.ClientWithResponses {
	t.Helper()
	client, err := hybridclient.NewClientWithResponses(serverURL)
	if err != nil {
		t.Fatalf("new hybrid client: %v", err)
	}
	return client
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

func graphclientClient(t *testing.T, serverURL string) *graphclient.ClientWithResponses {
	t.Helper()
	client, err := graphclient.NewClientWithResponses(serverURL)
	if err != nil {
		t.Fatalf("new graph client: %v", err)
	}
	return client
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

func recordsclientClient(t *testing.T, serverURL string) *recordsclient.ClientWithResponses {
	t.Helper()
	client, err := recordsclient.NewClientWithResponses(serverURL)
	if err != nil {
		t.Fatalf("new records client: %v", err)
	}
	return client
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
