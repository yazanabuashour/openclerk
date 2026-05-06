package runner_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func TestRetrievalTaskSemanticSearchRequiresVerifiedModule(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "docs/semantic.md", "Semantic", "# Semantic\n\n## Summary\nSemantic recall citations stay local.\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSemanticSearch,
		SemanticSearch: runner.SemanticSearchOptions{
			Query: "semantic recall",
			Limit: 5,
		},
	})
	if err != nil {
		t.Fatalf("semantic search: %v", err)
	}
	if result.SemanticSearch == nil ||
		result.SemanticSearch.SearchStatus != "provider_blocked" ||
		result.SemanticSearch.Provider.Provider != "ollama" ||
		!strings.Contains(result.SemanticSearch.Provider.ErrorSummary, "not installed") ||
		!strings.Contains(result.SemanticSearch.ValidationBoundaries, "default search remains lexical") {
		t.Fatalf("result = %+v", result.SemanticSearch)
	}
}

func TestRetrievalTaskSemanticSearchDispatchesInstalledModule(t *testing.T) {
	installSemanticModuleHelper(t)

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "docs/architecture/semantic-core.md", "Semantic Core", strings.TrimSpace(`---
tag: semantic-local
owner: architecture
---
# Semantic Core

## Summary
Semantic recall citations stay local and support explicit module semantic search.
`)+"\n")
	createDocument(t, ctx, config, "docs/architecture/lexical-core.md", "Lexical Core", strings.TrimSpace(`---
tag: lexical-local
owner: architecture
---
# Lexical Core

## Summary
Lexical search defaults remain separate from semantic retrieval.
`)+"\n")
	manifestPath := writeSemanticModuleManifest(t, t.TempDir(), "ollama")
	if _, err := runclient.InstallSemanticModule(ctx, config, runclient.SemanticModuleInstallInput{
		Provider:     "ollama",
		ManifestPath: manifestPath,
		Command:      "semantic-retrieval-adapter",
		ProviderConfig: map[string]string{
			"embedding_model": "embeddinggemma",
			"ollama_url":      "http://localhost:11434",
		},
	}); err != nil {
		t.Fatalf("install module: %v", err)
	}

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSemanticSearch,
		SemanticSearch: runner.SemanticSearchOptions{
			Query:         "semantic recall",
			PathPrefix:    "docs/architecture/",
			MetadataKey:   "owner",
			MetadataValue: "architecture",
			Limit:         5,
		},
	})
	if err != nil {
		t.Fatalf("semantic search: %v", err)
	}
	if result.SemanticSearch == nil ||
		result.SemanticSearch.SearchStatus != "completed" ||
		result.SemanticSearch.Provider.Provider != "ollama" ||
		result.SemanticSearch.Provider.Model != "embeddinggemma" ||
		len(result.SemanticSearch.Hits) != 1 ||
		result.SemanticSearch.Hits[0].Citations[0].Path != "docs/architecture/semantic-core.md" {
		t.Fatalf("semantic result = %+v", result.SemanticSearch)
	}
}

func TestRetrievalTaskSemanticSearchRejectsModuleHitsWithoutCitations(t *testing.T) {
	installSemanticModuleHelper(t)
	t.Setenv("OPENCLERK_SEMANTIC_MODULE_HELPER_NO_CITATIONS", "1")

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "docs/semantic.md", "Semantic", "# Semantic\n\n## Summary\nSemantic recall citations stay local.\n")
	manifestPath := writeSemanticModuleManifest(t, t.TempDir(), "ollama")
	if _, err := runclient.InstallSemanticModule(ctx, config, runclient.SemanticModuleInstallInput{
		Provider:     "ollama",
		ManifestPath: manifestPath,
		Command:      "semantic-retrieval-adapter",
	}); err != nil {
		t.Fatalf("install module: %v", err)
	}

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:         runner.RetrievalTaskActionSemanticSearch,
		SemanticSearch: runner.SemanticSearchOptions{Query: "semantic recall", Limit: 5},
	})
	if err != nil {
		t.Fatalf("semantic search: %v", err)
	}
	if result.SemanticSearch == nil ||
		result.SemanticSearch.SearchStatus != "provider_blocked" ||
		!strings.Contains(result.SemanticSearch.Provider.ErrorSummary, "without citations") {
		t.Fatalf("semantic result = %+v", result.SemanticSearch)
	}
}

func TestRetrievalTaskSemanticSearchRejectsConfiguredRemoteOllamaURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "docs/semantic.md", "Semantic", "# Semantic\n\n## Summary\nSemantic recall citations stay local.\n")
	manifestPath := writeSemanticModuleManifest(t, t.TempDir(), "ollama")
	if _, err := runclient.InstallSemanticModule(ctx, config, runclient.SemanticModuleInstallInput{
		Provider:     "ollama",
		ManifestPath: manifestPath,
		Command:      "semantic-retrieval-adapter",
		ProviderConfig: map[string]string{
			"ollama_url": "https://embeddings.example.test",
		},
	}); err != nil {
		t.Fatalf("install module: %v", err)
	}

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:         runner.RetrievalTaskActionSemanticSearch,
		SemanticSearch: runner.SemanticSearchOptions{Query: "semantic recall", Limit: 5},
	})
	if err != nil {
		t.Fatalf("semantic search: %v", err)
	}
	if result.SemanticSearch == nil ||
		result.SemanticSearch.SearchStatus != "provider_blocked" ||
		!strings.Contains(result.SemanticSearch.Provider.ErrorSummary, "loopback HTTP URL") {
		t.Fatalf("semantic result = %+v", result.SemanticSearch)
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

	unknownProvider, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSemanticSearch,
		SemanticSearch: runner.SemanticSearchOptions{
			Query:    "semantic",
			Provider: "remote",
		},
	})
	if err != nil {
		t.Fatalf("provider validation: %v", err)
	}
	if !unknownProvider.Rejected || unknownProvider.RejectionReason != "semantic_search.provider must be ollama or gemini" {
		t.Fatalf("unknown provider result = %+v", unknownProvider)
	}
}

func installSemanticModuleHelper(t *testing.T) {
	t.Helper()

	t.Setenv("OPENCLERK_SEMANTIC_MODULE_HELPER", "1")
	helperDir := t.TempDir()
	helperPath := filepath.Join(helperDir, "semantic-retrieval-adapter")
	testBinary, err := filepath.Abs(os.Args[0])
	if err != nil {
		t.Fatalf("resolve test binary path: %v", err)
	}
	script := fmt.Sprintf("#!/bin/sh\nexec %q -test.run=TestSemanticSearchModuleHelper -- \"$@\"\n", testBinary)
	if err := os.WriteFile(helperPath, []byte(script), 0o700); err != nil {
		t.Fatalf("write semantic module helper: %v", err)
	}
	t.Setenv("PATH", helperDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestSemanticSearchModuleHelper(t *testing.T) {
	if os.Getenv("OPENCLERK_SEMANTIC_MODULE_HELPER") != "1" {
		return
	}
	args := os.Args
	separator := -1
	for idx, arg := range args {
		if arg == "--" {
			separator = idx
			break
		}
	}
	if separator == -1 || len(args[separator+1:]) < 3 || args[separator+1] != "search" || args[separator+2] != "--db" {
		t.Fatalf("unexpected helper args: %v", args)
	}
	var request struct {
		Query          string `json:"query"`
		PathPrefix     string `json:"path_prefix"`
		MetadataKey    string `json:"metadata_key"`
		MetadataValue  string `json:"metadata_value"`
		Provider       string `json:"provider"`
		EmbeddingModel string `json:"embedding_model"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	citations := []map[string]any{{
		"doc_id":     "doc_semantic_core",
		"chunk_id":   "chunk_summary",
		"path":       "docs/architecture/semantic-core.md",
		"heading":    "Summary",
		"line_start": 6,
		"line_end":   7,
	}}
	if os.Getenv("OPENCLERK_SEMANTIC_MODULE_HELPER_NO_CITATIONS") == "1" {
		citations = nil
	}
	response := map[string]any{
		"schema_version":        "openclerk_semantic_retrieval.v1",
		"query":                 request.Query,
		"path_prefix":           request.PathPrefix,
		"metadata_key":          request.MetadataKey,
		"metadata_value":        request.MetadataValue,
		"provider":              map[string]any{"provider": request.Provider, "model": request.EmbeddingModel, "status": "completed", "embedding_dimensions": 3},
		"cache":                 map[string]any{"status": "hit", "cache_ref": "user_cache:semantic-test/cache.json", "chunk_count": 1},
		"ranking":               "hybrid_rrf_vector_lexical",
		"search_status":         "completed",
		"privacy_disclosure":    "local Ollama embeddings keep corpus/query text on this machine",
		"validation_boundaries": "optional OpenClerk module; read-only runner access; no core search default change",
		"authority_limits":      "semantic similarity is retrieval evidence only; canonical markdown citations remain authority",
		"results": []map[string]any{{
			"rank":      1,
			"score":     0.5,
			"doc_id":    "doc_semantic_core",
			"chunk_id":  "chunk_summary",
			"title":     "Semantic Core",
			"snippet":   "Semantic recall citations stay local.",
			"citations": citations,
		}},
		"agent_handoff": map[string]any{
			"summary":                        "semantic module returned test result",
			"evidence_inspected":             []string{"docs/architecture/semantic-core.md"},
			"follow_up_primitive_inspection": "use get_document for authority drill-down",
		},
	}
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		t.Fatalf("encode response: %v", err)
	}
	os.Exit(0)
}

func writeSemanticModuleManifest(t *testing.T, dir string, provider string) string {
	t.Helper()
	path := filepath.Join(dir, "module.json")
	manifest := map[string]any{
		"schema_version": "openclerk-module.v1",
		"module": map[string]any{
			"name":    provider + "-embeddings",
			"version": "0.1.0",
			"kind":    "embedding_provider",
		},
		"provides": []map[string]any{{
			"type": "command",
			"name": "semantic-retrieval-adapter search",
		}},
		"authority": map[string]any{
			"default":        "read_only",
			"durable_writes": "forbidden",
			"forbidden":      []string{"write_documents", "change_openclerk_search_default"},
		},
		"release": map[string]any{
			"status": "supported_optional_module",
		},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}
