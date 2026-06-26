package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
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

func TestSemanticRetrievalAdapterModuleVersionLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "source default", in: "0.1.0", want: "0.1.0"},
		{name: "release tag ldflag", in: "v0.1.1", want: "0.1.1"},
		{name: "blank fallback", in: " ", want: "0.1.0"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := adapterVersionLabel(tc.in); got != tc.want {
				t.Fatalf("adapterVersionLabel(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSemanticRetrievalAdapterOllamaEmbedsCorpusInBatches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	for idx := 0; idx < ollamaEmbedBatchSize+1; idx++ {
		createModuleDocument(t, ctx, dbPath, fmt.Sprintf("docs/batch/doc-%02d.md", idx), "Batch Document", "# Batch Document\n\n## Summary\nSemantic batch recall evidence stays citation-bearing.\n")
	}

	requests := 0
	maxInputs := 0
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
		if len(req.Input) > maxInputs {
			maxInputs = len(req.Input)
		}
		vectors := make([][]float64, 0, len(req.Input))
		for range req.Input {
			vectors = append(vectors, []float64{1, 0, 0})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": vectors})
	}))
	defer server.Close()

	response, err := executeSearch(ctx, runclient.Config{DatabasePath: dbPath}, searchRequest{
		Query:          "semantic batch recall",
		PathPrefix:     "docs/batch/",
		Limit:          3,
		Provider:       providerOllama,
		OllamaURL:      server.URL,
		EmbeddingModel: "embeddinggemma",
		CacheDir:       t.TempDir(),
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	expectedChunks := (ollamaEmbedBatchSize + 1) * 2
	expectedRequests := (expectedChunks+ollamaEmbedBatchSize-1)/ollamaEmbedBatchSize + 1
	if response.SearchStatus != "completed" || response.Cache.Status != "rebuilt" || response.Cache.RebuiltCount != expectedChunks {
		t.Fatalf("response = %+v", response)
	}
	if maxInputs > ollamaEmbedBatchSize {
		t.Fatalf("Ollama batch too large: maxInputs=%d batchSize=%d", maxInputs, ollamaEmbedBatchSize)
	}
	if requests != expectedRequests || response.Provider.RequestCount != expectedRequests {
		t.Fatalf("requests=%d provider=%+v", requests, response.Provider)
	}
}

func TestSemanticRetrievalAdapterSplitsLargeSectionsBeforeEmbedding(t *testing.T) {
	t.Parallel()

	longLines := strings.Repeat("Semantic recall evidence stays citation-bearing and local.\n", 120)
	longSingleLine := strings.Repeat("semantic ", semanticChunkTargetCharacters*2)
	chunks := chunksForDocument(domain.Document{
		DocID: "doc_long",
		Path:  "derived/text/long.md",
		Title: "Long Extracted Text",
		Body:  "# Long Extracted Text\n\n## Extracted Text\n" + longLines + "\n## Huge Line\n" + longSingleLine + "\n",
	})
	if len(chunks) < 5 {
		t.Fatalf("expected large document to split into several chunks, got %d", len(chunks))
	}
	for _, chunk := range chunks {
		if len([]rune(chunk.Content)) > semanticChunkTargetCharacters {
			t.Fatalf("chunk %s content length = %d, want <= %d", chunk.ChunkID, len([]rune(chunk.Content)), semanticChunkTargetCharacters)
		}
		if chunk.LineStart < 1 || chunk.LineEnd < chunk.LineStart {
			t.Fatalf("invalid citation lines: %+v", chunk)
		}
	}
}

func TestSemanticRetrievalAdapterRejectsOversizedChunkCorpus(t *testing.T) {
	t.Parallel()

	var body strings.Builder
	body.WriteString("# Oversized Corpus\n\n## Summary\n")
	for range maxSemanticChunks + 1 {
		body.WriteString(strings.Repeat("semantic ", semanticChunkTargetCharacters/len("semantic ")+1))
		body.WriteString("\n")
	}

	_, err := chunksForDocumentLimited(domain.Document{
		DocID: "doc_oversized",
		Path:  "docs/oversized.md",
		Title: "Oversized Corpus",
		Body:  body.String(),
	}, maxSemanticChunks)
	if err == nil || !strings.Contains(err.Error(), "semantic corpus exceeds maximum supported chunks") {
		t.Fatalf("oversized corpus error = %v", err)
	}
}

func TestChunksForDocumentLimitedRejectsHugeSingleDocumentBeforeFullAllocation(t *testing.T) {
	t.Parallel()

	_, err := chunksForDocumentLimited(domain.Document{
		DocID: "doc_huge",
		Path:  "derived/text/huge.md",
		Title: "Huge Extracted Text",
		Body:  strings.Repeat("semantic ", semanticChunkTargetCharacters*3),
	}, 2)
	if err == nil || !strings.Contains(err.Error(), "semantic corpus exceeds maximum supported chunks") {
		t.Fatalf("limited huge document error = %v", err)
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

func TestSemanticRetrievalAdapterCacheHitUsesCurrentCitationMetadata(t *testing.T) {
	t.Parallel()

	chunks := []semanticChunk{{
		ChunkID:   "chunk_current",
		DocID:     "doc_current",
		Path:      "docs/current.md",
		Title:     "Current",
		Heading:   "Summary",
		Content:   "Current citation metadata is authoritative.",
		LineStart: 3,
		LineEnd:   4,
		Hash:      "current-hash",
	}}
	expected := cacheFile{
		SchemaVersion: "semantic_retrieval_adapter_cache.v1",
		Provider:      providerOllama,
		Model:         "embeddinggemma",
		CorpusHash:    corpusHash(chunks),
	}
	forged := chunks[0]
	forged.DocID = "doc_forged"
	forged.Path = "notes/private.md"
	forged.Title = "Forged"
	forged.LineStart = 99
	forged.Vector = []float64{1, 2, 3}
	cache := expected
	cache.Chunks = []semanticChunk{forged}
	cachePath := filepath.Join(t.TempDir(), "cache.json")
	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatalf("marshal forged cache: %v", err)
	}
	if err := os.WriteFile(cachePath, data, 0o600); err != nil {
		t.Fatalf("write forged cache: %v", err)
	}

	hit, status := readCache(cachePath, expected, chunks)
	if status.Status != "hit" || len(hit) != 1 {
		t.Fatalf("readCache status=%+v hit=%+v", status, hit)
	}
	if hit[0].DocID != chunks[0].DocID || hit[0].Path != chunks[0].Path || hit[0].LineStart != chunks[0].LineStart {
		t.Fatalf("cache metadata was trusted: %+v want current %+v", hit[0], chunks[0])
	}
	if len(hit[0].Vector) != 3 {
		t.Fatalf("cache vector was not retained: %+v", hit[0].Vector)
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

	for _, prefix := range []string{"../private", "docs/..", "docs/../private"} {
		prefix := prefix
		t.Run(prefix, func(t *testing.T) {
			t.Parallel()

			_, err := executeSearch(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "openclerk.sqlite")}, searchRequest{
				Query:      "semantic",
				PathPrefix: prefix,
				Limit:      5,
			})
			if err == nil || !strings.Contains(err.Error(), "path_prefix") {
				t.Fatalf("expected unsafe prefix rejection, got %v", err)
			}
		})
	}
}

func TestSemanticRetrievalAdapterCachePathStaysUnderOpenClerkState(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	attackerCacheDir := filepath.Join(t.TempDir(), "attacker-cache")
	chunks := []semanticChunk{{
		ChunkID:      "chunk_cache",
		DocID:        "doc_cache",
		Path:         "docs/cache.md",
		Title:        "Cache",
		Heading:      "Summary",
		Content:      "cache boundary",
		TextForIndex: "Cache\ndocs/cache.md\nSummary\ncache boundary",
		Hash:         "cache-hash",
	}}

	_, cachePath, cacheRef := cacheForRequest(dbPath, searchRequest{
		Provider:       providerOllama,
		EmbeddingModel: "embeddinggemma",
		CacheDir:       attackerCacheDir,
	}, chunks)
	wantRoot := semanticCacheRoot(dbPath)
	if !strings.HasPrefix(cachePath, wantRoot+string(os.PathSeparator)) {
		t.Fatalf("cache path = %q, want under %q", cachePath, wantRoot)
	}
	if strings.HasPrefix(cachePath, attackerCacheDir+string(os.PathSeparator)) {
		t.Fatalf("cache path used caller cache_dir: %q", cachePath)
	}
	if !strings.HasPrefix(cacheRef, "openclerk_state:semantic-retrieval-adapter/") {
		t.Fatalf("cache ref = %q", cacheRef)
	}
}

func TestSemanticRetrievalAdapterRejectsUnsafeProviderSettingsBeforeDBAccess(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "missing", "openclerk.sqlite")
	cases := []struct {
		name    string
		request searchRequest
		want    string
	}{
		{
			name: "remote ollama",
			request: searchRequest{
				Query:     "semantic",
				Limit:     5,
				Provider:  providerOllama,
				OllamaURL: "https://embeddings.example.test",
			},
			want: "ollama_url must be a loopback HTTP URL",
		},
		{
			name: "non canonical gemini base",
			request: searchRequest{
				Query:         "semantic",
				Limit:         5,
				Provider:      providerGemini,
				GeminiAPIBase: "http://127.0.0.1:9999",
			},
			want: "gemini_api_base must be https://generativelanguage.googleapis.com/v1beta",
		},
		{
			name: "non default gemini config key",
			request: searchRequest{
				Query:           "semantic",
				Limit:           5,
				Provider:        providerGemini,
				GeminiConfigKey: "OTHER_RUNTIME_KEY",
			},
			want: "gemini_config_key must be GEMINI_API_KEY",
		},
		{
			name: "fallback gemini base",
			request: searchRequest{
				Query:            "semantic",
				Limit:            5,
				Provider:         providerOllama,
				OllamaURL:        "http://localhost:11434",
				FallbackProvider: providerGemini,
				GeminiAPIBase:    "https://attacker.example.test/v1beta",
			},
			want: "gemini_api_base must be https://generativelanguage.googleapis.com/v1beta",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := executeSearch(context.Background(), runclient.Config{DatabasePath: dbPath}, tc.request)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestSemanticRetrievalAdapterGeminiValidationDoesNotReadRuntimeConfigOrCallHTTP(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		http.Error(w, "should not be called", http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := executeSearch(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "missing.sqlite")}, searchRequest{
		Query:         "semantic",
		Limit:         5,
		Provider:      providerGemini,
		GeminiAPIBase: server.URL,
	})
	if err == nil || !strings.Contains(err.Error(), "gemini_api_base") {
		t.Fatalf("error = %v, want gemini_api_base rejection", err)
	}
	if requests != 0 {
		t.Fatalf("unsafe gemini request reached HTTP server %d time(s)", requests)
	}
}

func TestSemanticRetrievalAdapterDoesNotDefaultRemoteFallback(t *testing.T) {
	t.Parallel()

	request := normalizeRequest(searchRequest{Provider: providerOllama})
	if request.FallbackProvider != "" {
		t.Fatalf("unexpected implicit fallback provider %q", request.FallbackProvider)
	}
}

func TestGeminiClientRetriesAndBatches(t *testing.T) {
	t.Parallel()

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

	client := geminiClient{
		baseURL:    server.URL,
		apiKey:     "test-secret-gemini-key",
		httpClient: providerHTTPClient(45 * time.Second),
		sleep:      func(time.Duration) {},
	}
	vectors, stats, err := client.embed(context.Background(), "gemini-embedding-001", []string{"semantic recall", "citation evidence"}, "RETRIEVAL_DOCUMENT", 3)
	if err != nil {
		t.Fatalf("gemini search: %v", err)
	}
	if len(vectors) != 2 || stats.RetryCount != 1 || stats.RequestCount != 2 {
		t.Fatalf("vectors=%v stats=%+v", vectors, stats)
	}
}

func TestGeminiRetryBackoffHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Retry-After", "60")
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := geminiClient{
		baseURL:    server.URL,
		apiKey:     "test-secret-gemini-key",
		httpClient: providerHTTPClient(45 * time.Second),
		sleep: func(time.Duration) {
			cancel()
		},
	}
	_, stats, err := client.embed(ctx, "gemini-embedding-001", []string{"private corpus text"}, "RETRIEVAL_DOCUMENT", 3)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("error = %v, want context canceled", err)
	}
	if stats.RequestCount != 1 || stats.RetryCount != 1 || requests != 1 {
		t.Fatalf("retry stats=%+v requests=%d, want one request and one canceled retry", stats, requests)
	}
}

func TestProviderClientsDoNotFollowRedirects(t *testing.T) {
	t.Parallel()

	attackerRequests := 0
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attackerRequests++
		if r.Header.Get("x-goog-api-key") != "" {
			t.Fatalf("api key reached redirected server")
		}
	}))
	defer attacker.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, attacker.URL+"/steal", http.StatusTemporaryRedirect)
	}))
	defer redirector.Close()

	ollama := ollamaClient{baseURL: redirector.URL, client: providerHTTPClient(60 * time.Second)}
	if _, _, err := ollama.embed(context.Background(), "embeddinggemma", []string{"private corpus text"}); err == nil {
		t.Fatalf("expected ollama redirect to fail")
	}
	gemini := geminiClient{
		baseURL:    redirector.URL,
		apiKey:     "test-secret-gemini-key",
		httpClient: providerHTTPClient(45 * time.Second),
		sleep:      func(time.Duration) {},
	}
	if _, _, err := gemini.embed(context.Background(), "gemini-embedding-001", []string{"private corpus text"}, "RETRIEVAL_DOCUMENT", 3); err == nil {
		t.Fatalf("expected gemini redirect to fail")
	}
	if attackerRequests != 0 {
		t.Fatalf("redirected provider request reached attacker server %d time(s)", attackerRequests)
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
