package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func TestSemanticRetrievalAdapterOllamaSearchUsesCacheAndCitations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	createModuleDocument(t, ctx, dbPath, "docs/architecture/hybrid.md", "Hybrid Retrieval", "# Hybrid Retrieval\n\n## Summary\nSemantic recall vector ranking evidence should preserve citations.\n")
	createModuleDocument(t, ctx, dbPath, "docs/architecture/lexical.md", "Lexical Search", "# Lexical Search\n\n## Summary\nExact lexical lookup evidence stays local.\n")

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			http.NotFound(w, r)
			return
		}
		requests++
		var req struct {
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		vectors := make([][]float64, 0, len(req.Input))
		for _, input := range req.Input {
			if strings.Contains(strings.ToLower(input), "semantic recall") {
				vectors = append(vectors, []float64{1, 0, 0})
			} else {
				vectors = append(vectors, []float64{0, 1, 0})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": vectors})
	}))
	defer server.Close()

	req := searchRequest{
		Query:          "semantic recall",
		PathPrefix:     "docs/architecture/",
		Limit:          5,
		Provider:       providerOllama,
		OllamaURL:      server.URL,
		EmbeddingModel: "embeddinggemma",
		CacheDir:       t.TempDir(),
	}
	first, err := executeSearch(ctx, runclient.Config{DatabasePath: dbPath}, req)
	if err != nil {
		t.Fatalf("first search: %v", err)
	}
	if first.SearchStatus != "completed" || first.Provider.Provider != providerOllama || first.Cache.Status != "rebuilt" ||
		len(first.Results) == 0 || first.Results[0].Citations[0].Path != "docs/architecture/hybrid.md" {
		t.Fatalf("first response = %+v", first)
	}
	second, err := executeSearch(ctx, runclient.Config{DatabasePath: dbPath}, req)
	if err != nil {
		t.Fatalf("second search: %v", err)
	}
	if second.Cache.Status != "hit" || requests != 3 {
		t.Fatalf("cache did not avoid document re-embedding, second=%+v requests=%d", second.Cache, requests)
	}
}

func TestSemanticRetrievalAdapterPathPrefixAndStaleCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	targetDocID := createModuleDocument(t, ctx, dbPath, "docs/architecture/semantic.md", "Semantic Retrieval", "# Semantic Retrieval\n\n## Summary\nSemantic recall citations stay local.\n")
	createModuleDocument(t, ctx, dbPath, "archive/semantic.md", "Archived Semantic Retrieval", "# Archived Semantic Retrieval\n\n## Summary\nArchived semantic recall must stay out of scoped module results.\n")

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			http.NotFound(w, r)
			return
		}
		requests++
		var req struct {
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		vectors := make([][]float64, 0, len(req.Input))
		for _, input := range req.Input {
			if strings.Contains(strings.ToLower(input), "semantic") {
				vectors = append(vectors, []float64{1, 0, 0})
			} else {
				vectors = append(vectors, []float64{0, 1, 0})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": vectors})
	}))
	defer server.Close()

	req := searchRequest{
		Query:          "semantic recall",
		PathPrefix:     "docs/architecture/",
		Limit:          5,
		Provider:       providerOllama,
		OllamaURL:      server.URL,
		EmbeddingModel: "embeddinggemma",
		CacheDir:       t.TempDir(),
	}
	first, err := executeSearch(ctx, runclient.Config{DatabasePath: dbPath}, req)
	if err != nil {
		t.Fatalf("first search: %v", err)
	}
	if first.Cache.Status != "rebuilt" || len(first.Results) == 0 || first.Results[0].Citations[0].Path != "docs/architecture/semantic.md" {
		t.Fatalf("first response = %+v", first)
	}
	for _, hit := range first.Results {
		if strings.HasPrefix(hit.Citations[0].Path, "archive/") {
			t.Fatalf("path prefix leaked archive hit: %+v", first.Results)
		}
	}

	cached, err := executeSearch(ctx, runclient.Config{DatabasePath: dbPath}, req)
	if err != nil {
		t.Fatalf("cached search: %v", err)
	}
	if cached.Cache.Status != "hit" {
		t.Fatalf("expected cache hit, got %+v", cached.Cache)
	}

	appendModuleDocument(t, ctx, dbPath, targetDocID, "Fresh corpus update changes the semantic cache hash.")
	stale, err := executeSearch(ctx, runclient.Config{DatabasePath: dbPath}, req)
	if err != nil {
		t.Fatalf("stale search: %v", err)
	}
	if stale.Cache.Status != "rebuilt" || stale.Cache.RebuiltCount == 0 || requests < 5 {
		t.Fatalf("expected stale cache rebuild, response=%+v requests=%d", stale.Cache, requests)
	}
}

func TestSemanticRetrievalAdapterTagAndMetadataFilters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	createModuleDocument(t, ctx, dbPath, "docs/architecture/semantic.md", "Semantic Retrieval", strings.TrimSpace(`---
tag: semantic-local
owner: architecture
---
# Semantic Retrieval

## Summary
Semantic recall citations stay local.
`)+"\n")
	createModuleDocument(t, ctx, dbPath, "docs/architecture/archive.md", "Archived Semantic Retrieval", strings.TrimSpace(`---
tag: semantic-local
owner: archive
---
# Archived Semantic Retrieval

## Summary
Archived semantic recall must stay out of owner-scoped module results.
`)+"\n")
	createModuleDocument(t, ctx, dbPath, "docs/architecture/lexical.md", "Lexical Retrieval", strings.TrimSpace(`---
tag: lexical-local
owner: architecture
---
# Lexical Retrieval

## Summary
Lexical retrieval should stay out of semantic tag-scoped module results.
`)+"\n")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			http.NotFound(w, r)
			return
		}
		var req struct {
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		vectors := make([][]float64, 0, len(req.Input))
		for _, input := range req.Input {
			lower := strings.ToLower(input)
			switch {
			case strings.Contains(lower, "semantic recall"):
				vectors = append(vectors, []float64{1, 0, 0})
			case strings.Contains(lower, "lexical retrieval"):
				vectors = append(vectors, []float64{0, 1, 0})
			default:
				vectors = append(vectors, []float64{0, 0, 1})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": vectors})
	}))
	defer server.Close()

	tagged, err := executeSearch(ctx, runclient.Config{DatabasePath: dbPath}, searchRequest{
		Query:          "semantic recall",
		PathPrefix:     "docs/architecture/",
		Tag:            "semantic-local",
		Limit:          5,
		Provider:       providerOllama,
		OllamaURL:      server.URL,
		EmbeddingModel: "embeddinggemma",
		CacheDir:       t.TempDir(),
	})
	if err != nil {
		t.Fatalf("tag search: %v", err)
	}
	if tagged.Tag != "semantic-local" || len(tagged.Results) != 2 {
		t.Fatalf("tagged response = %+v", tagged)
	}
	for _, hit := range tagged.Results {
		if hit.Citations[0].Path == "docs/architecture/lexical.md" {
			t.Fatalf("tag filter leaked lexical hit: %+v", tagged.Results)
		}
	}

	metadata, err := executeSearch(ctx, runclient.Config{DatabasePath: dbPath}, searchRequest{
		Query:          "semantic recall",
		PathPrefix:     "docs/architecture/",
		MetadataKey:    "owner",
		MetadataValue:  "architecture",
		Limit:          5,
		Provider:       providerOllama,
		OllamaURL:      server.URL,
		EmbeddingModel: "embeddinggemma",
		CacheDir:       t.TempDir(),
	})
	if err != nil {
		t.Fatalf("metadata search: %v", err)
	}
	if metadata.MetadataKey != "owner" || metadata.MetadataValue != "architecture" || len(metadata.Results) != 2 {
		t.Fatalf("metadata response = %+v", metadata)
	}
	for _, hit := range metadata.Results {
		if hit.Citations[0].Path == "docs/architecture/archive.md" {
			t.Fatalf("metadata filter leaked archive hit: %+v", metadata.Results)
		}
	}
}

func TestSemanticRetrievalAdapterFilterValidation(t *testing.T) {
	t.Parallel()

	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "openclerk.sqlite")}
	cases := []struct {
		name    string
		request searchRequest
		want    string
	}{
		{
			name:    "empty tag",
			request: searchRequest{Query: "semantic", Tag: " ", tagProvided: true, Limit: 5},
			want:    "tag must be non-empty",
		},
		{
			name:    "tag with metadata",
			request: searchRequest{Query: "semantic", Tag: "semantic-local", MetadataKey: "owner", MetadataValue: "architecture", Limit: 5},
			want:    "tag cannot be combined with metadata_key or metadata_value",
		},
		{
			name:    "metadata missing value",
			request: searchRequest{Query: "semantic", MetadataKey: "owner", Limit: 5},
			want:    "metadata_key and metadata_value must be provided together",
		},
		{
			name:    "metadata missing key",
			request: searchRequest{Query: "semantic", MetadataValue: "architecture", Limit: 5},
			want:    "metadata_key and metadata_value must be provided together",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := executeSearch(context.Background(), config, tc.request)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, err)
			}
		})
	}
}

func TestSemanticRetrievalAdapterProviderBlockedWithoutFallback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	createModuleDocument(t, ctx, dbPath, "docs/architecture/semantic.md", "Semantic Retrieval", "# Semantic Retrieval\n\n## Summary\nSemantic recall citations stay local.\n")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing model", http.StatusNotFound)
	}))
	defer server.Close()

	response, err := executeSearch(ctx, runclient.Config{DatabasePath: dbPath}, searchRequest{
		Query:          "semantic recall",
		PathPrefix:     "docs/architecture/",
		Limit:          5,
		Provider:       providerOllama,
		OllamaURL:      server.URL,
		EmbeddingModel: "embeddinggemma",
		CacheDir:       t.TempDir(),
	})
	if err != nil {
		t.Fatalf("blocked search: %v", err)
	}
	if response.SearchStatus != "provider_blocked" ||
		response.Provider.Provider != providerOllama ||
		response.Provider.FallbackProvider != "" ||
		!strings.Contains(response.AgentHandoff.ApprovalOrConfigurationNeeded, "runtime_config:GEMINI_API_KEY") {
		t.Fatalf("provider blocked response = %+v", response)
	}
}

func TestSemanticRetrievalAdapterRejectsUnsafePrefix(t *testing.T) {
	t.Parallel()

	_, err := executeSearch(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "openclerk.sqlite")}, searchRequest{
		Query:      "semantic",
		PathPrefix: "../private",
		Limit:      5,
	})
	if err == nil || !strings.Contains(err.Error(), "path_prefix") {
		t.Fatalf("expected unsafe prefix rejection, got %v", err)
	}
}

func TestSemanticRetrievalAdapterDoesNotDefaultRemoteFallback(t *testing.T) {
	t.Parallel()

	request := normalizeRequest(searchRequest{Provider: providerOllama})
	if request.FallbackProvider != "" {
		t.Fatalf("unexpected implicit fallback provider %q", request.FallbackProvider)
	}
}

func TestSemanticRetrievalAdapterGeminiRetriesAndRedactsCredential(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	createModuleDocument(t, ctx, dbPath, "docs/architecture/semantic.md", "Semantic Retrieval", "# Semantic Retrieval\n\n## Summary\nSemantic recall evidence keeps citations.\n")
	writeModuleRuntimeConfig(t, dbPath, "GEMINI_API_KEY", "test-secret-gemini-key")
	cacheDir := t.TempDir()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models/gemini-embedding-001:batchEmbedContents" {
			http.NotFound(w, r)
			return
		}
		requests++
		if requests == 1 {
			w.Header().Set("Retry-After", "0.001")
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		var req struct {
			Requests []json.RawMessage `json:"requests"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode gemini request: %v", err)
		}
		embeddings := make([]map[string]any, 0, len(req.Requests))
		for range req.Requests {
			embeddings = append(embeddings, map[string]any{"values": []float64{1, 0, 0}})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": embeddings})
	}))
	defer server.Close()

	response, err := executeSearch(ctx, runclient.Config{DatabasePath: dbPath}, searchRequest{
		Query:                     "semantic recall",
		Limit:                     5,
		Provider:                  providerGemini,
		EmbeddingModel:            "gemini-embedding-001",
		GeminiAPIBase:             server.URL,
		EmbeddingOutputDimensions: 3,
		CacheDir:                  cacheDir,
	})
	if err != nil {
		t.Fatalf("gemini search: %v", err)
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if response.SearchStatus != "completed" ||
		response.Provider.Provider != providerGemini ||
		response.Provider.CredentialRef != "runtime_config:GEMINI_API_KEY" ||
		response.Provider.RetryCount == 0 ||
		strings.Contains(string(encoded), "test-secret-gemini-key") ||
		len(response.Results) == 0 ||
		response.Results[0].Citations[0].Path != "docs/architecture/semantic.md" {
		t.Fatalf("gemini response = %s", encoded)
	}

	cached, err := executeSearch(ctx, runclient.Config{DatabasePath: dbPath}, searchRequest{
		Query:                     "semantic recall",
		Limit:                     5,
		Provider:                  providerGemini,
		EmbeddingModel:            "gemini-embedding-001",
		GeminiAPIBase:             server.URL,
		EmbeddingOutputDimensions: 3,
		CacheDir:                  cacheDir,
	})
	if err != nil {
		t.Fatalf("cached gemini search: %v", err)
	}
	if cached.Cache.Status != "hit" || cached.Provider.CredentialRef != "runtime_config:GEMINI_API_KEY" {
		t.Fatalf("cached gemini response = %+v", cached)
	}
}

func createModuleDocument(t *testing.T, ctx context.Context, dbPath string, path string, title string, body string) string {
	t.Helper()
	if _, err := runclient.InitializePaths(runclient.Config{DatabasePath: dbPath}, filepath.Join(filepath.Dir(dbPath), "vault")); err != nil {
		t.Fatalf("initialize paths: %v", err)
	}
	result, err := runner.RunDocumentTask(ctx, runclient.Config{DatabasePath: dbPath}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  path,
			Title: title,
			Body:  body,
		},
	})
	if err != nil {
		t.Fatalf("create document: %v", err)
	}
	if result.Rejected {
		t.Fatalf("create rejected: %+v", result)
	}
	if result.Document == nil {
		t.Fatalf("create missing document: %+v", result)
	}
	return result.Document.DocID
}

func writeModuleRuntimeConfig(t *testing.T, dbPath string, key string, value string) {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()
	_, err = db.Exec(`INSERT OR REPLACE INTO runtime_config (key_name, value_text, updated_at) VALUES (?, ?, ?)`, key, value, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("write runtime config: %v", err)
	}
}

func appendModuleDocument(t *testing.T, ctx context.Context, dbPath string, docID string, content string) {
	t.Helper()
	result, err := runner.RunDocumentTask(ctx, runclient.Config{DatabasePath: dbPath}, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionAppend,
		DocID:   docID,
		Content: content,
	})
	if err != nil {
		t.Fatalf("append document: %v", err)
	}
	if result.Rejected {
		t.Fatalf("append rejected: %+v", result)
	}
}
