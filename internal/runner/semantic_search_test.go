package runner_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func TestRetrievalTaskSemanticSearchLocalCacheFiltersAndCitations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	target := createDocument(t, ctx, config, "docs/architecture/semantic-core.md", "Semantic Core", strings.TrimSpace(`---
tag: semantic-local
owner: architecture
---
# Semantic Core

## Summary
Semantic recall citations stay local and support explicit core semantic search.
`)+"\n")
	createDocument(t, ctx, config, "docs/architecture/semantic-archive.md", "Semantic Archive", strings.TrimSpace(`---
tag: semantic-local
owner: archive
---
# Semantic Archive

## Summary
Archived semantic recall should stay out of owner-scoped semantic search.
`)+"\n")
	createDocument(t, ctx, config, "docs/architecture/lexical-core.md", "Lexical Core", strings.TrimSpace(`---
tag: lexical-local
owner: architecture
---
# Lexical Core

## Summary
Lexical search defaults remain separate from semantic retrieval.
`)+"\n")

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
			lower := strings.ToLower(input)
			switch {
			case strings.Contains(lower, "semantic recall"):
				vectors = append(vectors, []float64{1, 0, 0})
			case strings.Contains(lower, "lexical search"):
				vectors = append(vectors, []float64{0, 1, 0})
			default:
				vectors = append(vectors, []float64{0, 0, 1})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": vectors})
	}))
	defer server.Close()

	req := runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSemanticSearch,
		SemanticSearch: runner.SemanticSearchOptions{
			Query:          "semantic recall",
			PathPrefix:     "docs/architecture/",
			MetadataKey:    "owner",
			MetadataValue:  "architecture",
			Limit:          5,
			OllamaURL:      server.URL,
			EmbeddingModel: "nomic-embed-text",
			CacheDir:       t.TempDir(),
		},
	}
	first, err := runner.RunRetrievalTask(ctx, config, req)
	if err != nil {
		t.Fatalf("semantic search: %v", err)
	}
	if first.SemanticSearch == nil ||
		first.SemanticSearch.SearchStatus != "completed" ||
		first.SemanticSearch.Cache.Status != "rebuilt" ||
		first.SemanticSearch.Provider.Provider != "ollama" ||
		len(first.SemanticSearch.Hits) != 2 ||
		first.SemanticSearch.Hits[0].Citations[0].Path != "docs/architecture/semantic-core.md" {
		t.Fatalf("first semantic result = %+v", first.SemanticSearch)
	}
	for _, hit := range first.SemanticSearch.Hits {
		if hit.Citations[0].Path == "docs/architecture/semantic-archive.md" {
			t.Fatalf("metadata filter leaked archive hit: %+v", first.SemanticSearch.Hits)
		}
	}

	second, err := runner.RunRetrievalTask(ctx, config, req)
	if err != nil {
		t.Fatalf("cached semantic search: %v", err)
	}
	if second.SemanticSearch == nil || second.SemanticSearch.Cache.Status != "hit" || requests != 3 {
		t.Fatalf("expected cache hit and query-only re-embed, result=%+v requests=%d", second.SemanticSearch, requests)
	}

	_, err = runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionAppend,
		DocID:   target.DocID,
		Content: "Fresh local semantic content changes the cache hash.",
	})
	if err != nil {
		t.Fatalf("append target: %v", err)
	}
	stale, err := runner.RunRetrievalTask(ctx, config, req)
	if err != nil {
		t.Fatalf("stale semantic search: %v", err)
	}
	if stale.SemanticSearch == nil || stale.SemanticSearch.Cache.Status != "rebuilt" || stale.SemanticSearch.Cache.RebuiltCount == 0 {
		t.Fatalf("expected stale cache rebuild, got %+v", stale.SemanticSearch)
	}
}

func TestRetrievalTaskSemanticSearchTagFilterAndValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "docs/semantic.md", "Semantic", strings.TrimSpace(`---
tag: semantic-local
---
# Semantic

## Summary
Semantic recall citations stay local.
`)+"\n")
	createDocument(t, ctx, config, "docs/lexical.md", "Lexical", strings.TrimSpace(`---
tag: lexical-local
---
# Lexical

## Summary
Lexical recall stays separate.
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
		for range req.Input {
			vectors = append(vectors, []float64{1, 0, 0})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": vectors})
	}))
	defer server.Close()

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSemanticSearch,
		SemanticSearch: runner.SemanticSearchOptions{
			Query:          "semantic recall",
			PathPrefix:     "docs/",
			Tag:            "semantic-local",
			Limit:          5,
			OllamaURL:      server.URL,
			EmbeddingModel: "nomic-embed-text",
			CacheDir:       t.TempDir(),
		},
	})
	if err != nil {
		t.Fatalf("tag semantic search: %v", err)
	}
	if result.SemanticSearch == nil || result.SemanticSearch.Tag != "" || result.SemanticSearch.MetadataKey != "tag" || len(result.SemanticSearch.Hits) != 1 || result.SemanticSearch.Hits[0].Citations[0].Path != "docs/semantic.md" {
		t.Fatalf("tag result = %+v", result.SemanticSearch)
	}

	mixed, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSemanticSearch,
		SemanticSearch: runner.SemanticSearchOptions{
			Query:         "semantic",
			Tag:           "semantic-local",
			MetadataKey:   "owner",
			MetadataValue: "architecture",
		},
	})
	if err != nil {
		t.Fatalf("mixed validation: %v", err)
	}
	if !mixed.Rejected || mixed.RejectionReason != "semantic_search.tag cannot be combined with metadata_key or metadata_value" {
		t.Fatalf("mixed result = %+v", mixed)
	}

	empty, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSemanticSearch,
		SemanticSearch: runner.SemanticSearchOptions{
			Query: "semantic",
			Tag:   " ",
		},
	})
	if err != nil {
		t.Fatalf("empty tag validation: %v", err)
	}
	if !empty.Rejected || empty.RejectionReason != "semantic_search.tag must be non-empty" {
		t.Fatalf("empty tag result = %+v", empty)
	}

	incompleteMetadata, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSemanticSearch,
		SemanticSearch: runner.SemanticSearchOptions{
			Query:       "semantic",
			MetadataKey: "owner",
		},
	})
	if err != nil {
		t.Fatalf("metadata validation: %v", err)
	}
	if !incompleteMetadata.Rejected || incompleteMetadata.RejectionReason != "semantic_search.metadata_key and metadata_value must be provided together" {
		t.Fatalf("incomplete metadata result = %+v", incompleteMetadata)
	}

	remoteURL, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSemanticSearch,
		SemanticSearch: runner.SemanticSearchOptions{
			Query:     "semantic",
			OllamaURL: "https://embeddings.example.test",
		},
	})
	if err != nil {
		t.Fatalf("ollama url validation: %v", err)
	}
	if !remoteURL.Rejected || remoteURL.RejectionReason != "semantic_search.ollama_url must be a loopback HTTP URL" {
		t.Fatalf("remote url result = %+v", remoteURL)
	}
}

func TestRetrievalTaskSemanticSearchProviderBlocked(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "docs/semantic.md", "Semantic", "# Semantic\n\n## Summary\nSemantic recall citations stay local.\n")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing model", http.StatusNotFound)
	}))
	defer server.Close()

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSemanticSearch,
		SemanticSearch: runner.SemanticSearchOptions{
			Query:          "semantic recall",
			Limit:          5,
			OllamaURL:      server.URL,
			EmbeddingModel: "nomic-embed-text",
			CacheDir:       t.TempDir(),
		},
	})
	if err != nil {
		t.Fatalf("provider blocked semantic search: %v", err)
	}
	if result.SemanticSearch == nil ||
		result.SemanticSearch.SearchStatus != "provider_blocked" ||
		result.SemanticSearch.Provider.Provider != "ollama" ||
		strings.Contains(result.SemanticSearch.PrivacyDisclosure, "Gemini") {
		t.Fatalf("blocked result = %+v", result.SemanticSearch)
	}
}
