package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestParseSemanticRecallConfigDefaults(t *testing.T) {
	config, err := parseSemanticRecallConfig([]string{"--mode", "LOCAL-HYBRID", "--run-root", "run"}, io.Discard)
	if err != nil {
		t.Fatalf("parse semantic recall config: %v", err)
	}
	if config.Mode != semanticRecallModeLocalHybrid ||
		config.ReportName != "ockp-semantic-recall-local-hybrid" ||
		config.EmbeddingProvider != semanticRecallProviderOllama ||
		config.OllamaURL != "http://localhost:11434" ||
		config.EmbeddingModel != "embeddinggemma" {
		t.Fatalf("config = %+v", config)
	}
}

func TestParseSemanticRecallConfigGeminiDefaults(t *testing.T) {
	config, err := parseSemanticRecallConfig([]string{"--embedding-provider", "gemini", "--run-root", "run"}, io.Discard)
	if err != nil {
		t.Fatalf("parse semantic recall config: %v", err)
	}
	if config.EmbeddingProvider != semanticRecallProviderGemini ||
		config.EmbeddingModel != "gemini-embedding-001" ||
		config.EmbeddingOutputDimensions != 3072 ||
		config.GeminiConfigKey != "GEMINI_API_KEY" {
		t.Fatalf("config = %+v", config)
	}
}

func TestParseSemanticRecallConfigGeminiConfigDBFollowsRuntimePath(t *testing.T) {
	t.Setenv("OPENCLERK_DATABASE_PATH", filepath.Join("custom", "openclerk.sqlite"))
	config, err := parseSemanticRecallConfig([]string{"--embedding-provider", "gemini", "--run-root", "run"}, io.Discard)
	if err != nil {
		t.Fatalf("parse semantic recall config: %v", err)
	}
	if config.GeminiConfigDB != filepath.Join("custom", "openclerk.sqlite") {
		t.Fatalf("gemini config db = %q", config.GeminiConfigDB)
	}

	t.Setenv("OPENCLERK_DATABASE_PATH", "")
	xdg := filepath.Join(t.TempDir(), "xdg")
	t.Setenv("XDG_DATA_HOME", xdg)
	config, err = parseSemanticRecallConfig([]string{"--embedding-provider", "gemini", "--run-root", "run"}, io.Discard)
	if err != nil {
		t.Fatalf("parse semantic recall config with XDG: %v", err)
	}
	if config.GeminiConfigDB != filepath.Join(xdg, "openclerk", "openclerk.sqlite") {
		t.Fatalf("xdg gemini config db = %q", config.GeminiConfigDB)
	}
}

func TestParseSemanticRecallConfigRejectsUnsafeReportNames(t *testing.T) {
	for _, name := range []string{"../erase", "nested/report", ".", ".."} {
		t.Run(name, func(t *testing.T) {
			_, err := parseSemanticRecallConfig([]string{"--run-root", "run", "--report-name", name}, io.Discard)
			if err == nil {
				t.Fatalf("expected unsafe report name %q to fail", name)
			}
		})
	}
}

func TestReadGeminiAPIKeyFromRuntimeConfigRedactsSource(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE runtime_config (key_name TEXT PRIMARY KEY, value_text TEXT NOT NULL, updated_at TEXT NOT NULL);
		INSERT INTO runtime_config (key_name, value_text, updated_at) VALUES ('GEMINI_API_KEY', 'secret-value', '2026-05-04T00:00:00Z');`)
	if err != nil {
		t.Fatalf("seed runtime config: %v", err)
	}
	_ = db.Close()

	secret, ref, err := readGeminiAPIKeyFromRuntimeConfig(context.Background(), dbPath, "GEMINI_API_KEY")
	if err != nil {
		t.Fatalf("read Gemini key: %v", err)
	}
	if secret != "secret-value" || ref != "runtime_config:GEMINI_API_KEY" {
		t.Fatalf("secret/ref = %q/%q", secret, ref)
	}
	_, _, err = readGeminiAPIKeyFromRuntimeConfig(context.Background(), dbPath, "MISSING")
	if err == nil || strings.Contains(err.Error(), "secret-value") {
		t.Fatalf("missing key error leaked secret or did not fail: %v", err)
	}
}

func TestSemanticRecallChunksCarryStableCitations(t *testing.T) {
	body := "# Test Doc\n\nIntro text.\n\n## First\n\nAlpha beta.\n\n## Second\n\nGamma delta.\n"
	chunks := semanticRecallChunksForDocument("docs/architecture/test.md", body)
	if len(chunks) != 3 {
		t.Fatalf("chunks = %d, want 3: %+v", len(chunks), chunks)
	}
	if chunks[1].Path != "docs/architecture/test.md" ||
		chunks[1].Heading != "First" ||
		chunks[1].LineStart != 5 ||
		chunks[1].LineEnd != 8 ||
		!strings.HasPrefix(chunks[1].ChunkID, "chunk_") {
		t.Fatalf("chunk citation = %+v", chunks[1])
	}
	if chunks[1].ChunkID != semanticRecallChunksForDocument("docs/architecture/test.md", body)[1].ChunkID {
		t.Fatal("chunk id is not stable")
	}
}

func TestSemanticRecallMetricsCollapseAndRRF(t *testing.T) {
	chunks := []semanticRecallChunk{
		{ChunkID: "chunk_a1", Path: "docs/architecture/a.md", Heading: "A", LineStart: 1, LineEnd: 2},
		{ChunkID: "chunk_a2", Path: "docs/architecture/a.md", Heading: "A2", LineStart: 3, LineEnd: 4},
		{ChunkID: "chunk_b1", Path: "docs/architecture/b.md", Heading: "B", LineStart: 5, LineEnd: 6},
	}
	ranked, duplicates := collapseSemanticRecallHits([]semanticRecallHit{
		{Chunk: chunks[0], Score: 9},
		{Chunk: chunks[1], Score: 8},
		{Chunk: chunks[2], Score: 7},
	})
	if duplicates != 1 || len(ranked) != 2 {
		t.Fatalf("collapse ranked=%+v duplicates=%d", ranked, duplicates)
	}
	rows := []semanticRecallRow{
		{Rank: 1},
		{Rank: 4},
		{Rank: 0},
	}
	hitAt3, mrr := semanticRecallMetrics(rows)
	if hitAt3 != 1 || mrr != 0.417 {
		t.Fatalf("metrics hit@3=%d mrr=%.3f", hitAt3, mrr)
	}
	rrf := rrfSemanticRecallHits(
		[]semanticRecallHit{{Chunk: chunks[2], Score: 10}, {Chunk: chunks[0], Score: 9}},
		[]semanticRecallHit{{Chunk: chunks[0], Score: 10}},
	)
	if rrf[0].Chunk.Path != "docs/architecture/a.md" {
		t.Fatalf("RRF did not lift shared lexical/vector hit: %+v", rrf)
	}
}

func TestSemanticRecallLexicalFallbackRegressionPressure(t *testing.T) {
	chunks := []semanticRecallChunk{
		{ChunkID: "target_a", Path: "notes/security-renewal.md", Title: "Security Renewal", Heading: "Tags", Content: "account renewal source-sensitive exact lookup marker"},
		{ChunkID: "target_b", Path: "notes/security-renewal.md", Title: "Security Renewal", Heading: "Details", Content: "duplicate chunk for the same document marker"},
		{ChunkID: "archive", Path: "archive/security-renewal.md", Title: "Archived Security Renewal", Heading: "Tags", Content: "account renewal old marker"},
	}
	for i := 0; i < 30; i++ {
		chunks = append(chunks, semanticRecallChunk{
			ChunkID: fmt.Sprintf("decoy_%02d", i),
			Path:    fmt.Sprintf("notes/decoy-%02d.md", i),
			Title:   "Renewal Decoy",
			Heading: "Tags",
			Content: "adjacent renewal text without the source-sensitive exact lookup marker",
		})
	}
	hits := scoreSemanticRecallChunks(chunks, semanticRecallTokens("security renewal source-sensitive exact lookup marker"))
	ranked, duplicates := collapseSemanticRecallHits(hits)
	if duplicates == 0 {
		t.Fatalf("expected duplicate pressure to be visible")
	}
	if semanticRecallRankOfPath(ranked, "notes/security-renewal.md") != 1 {
		t.Fatalf("target did not survive exact/title/path/heading pressure: %+v", ranked[:3])
	}
	if semanticRecallRankOfPath(ranked, "archive/security-renewal.md") == 0 {
		t.Fatalf("archive path pressure should remain visible for production path-prefix regression coverage")
	}
}

func TestGeminiClientEmbedBatchAndRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "test-key" {
			t.Fatalf("missing API key header")
		}
		if r.URL.Path != "/models/gemini-embedding-001:batchEmbedContents" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		var request struct {
			Requests []struct {
				TaskType             string `json:"taskType"`
				OutputDimensionality int    `json:"outputDimensionality"`
			} `json:"requests"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(request.Requests) != 2 || request.Requests[0].TaskType != "RETRIEVAL_DOCUMENT" || request.Requests[0].OutputDimensionality != 3 {
			t.Fatalf("request = %+v", request)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"embeddings": []map[string]any{
				{"values": []float64{1, 0, 0}},
				{"values": []float64{0, 1, 0}},
			},
		})
	}))
	defer server.Close()

	client := geminiClient{baseURL: server.URL, apiKey: "test-key", httpClient: server.Client(), sleep: func(time.Duration) {}}
	vectors, stats, err := client.embed(context.Background(), "gemini-embedding-001", []string{"a", "b"}, "RETRIEVAL_DOCUMENT", 3)
	if err != nil {
		t.Fatalf("embed: %v", err)
	}
	if len(vectors) != 2 || len(vectors[0]) != 3 || stats.RequestCount != 2 || stats.RetryCount != 1 || stats.BackoffSeconds != 1 {
		t.Fatalf("vectors=%+v stats=%+v", vectors, stats)
	}
}

func TestGeminiClientValidatesEmbeddingDimensions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": []map[string]any{{"values": []float64{1, 0}}}})
	}))
	defer server.Close()

	client := geminiClient{baseURL: server.URL, apiKey: "test-key", httpClient: server.Client(), sleep: func(time.Duration) {}}
	_, _, err := client.embed(context.Background(), "gemini-embedding-001", []string{"a"}, "RETRIEVAL_QUERY", 3)
	if err == nil || !strings.Contains(err.Error(), "dimensions") {
		t.Fatalf("expected dimension error, got %v", err)
	}
}

func TestOllamaClientShowAndEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/show":
			_, _ = w.Write([]byte(`{"capabilities":["embedding"],"details":{"family":"test"},"model_info":{"embeddinggemma.embedding_length":3}}`))
		case "/api/embed":
			var request struct {
				Input []string `json:"input"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode embed request: %v", err)
			}
			embeddings := make([][]float64, 0, len(request.Input))
			for idx := range request.Input {
				embeddings = append(embeddings, []float64{float64(idx + 1), 0, 0})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": embeddings})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := ollamaClient{baseURL: server.URL, client: server.Client()}
	show, err := client.show(context.Background(), "embeddinggemma")
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if len(show.Capabilities) != 1 || show.ModelInfo["embeddinggemma.embedding_length"] == nil {
		t.Fatalf("show = %+v", show)
	}
	embeddings, err := client.embed(context.Background(), "embeddinggemma", []string{"a", "b"})
	if err != nil {
		t.Fatalf("embed: %v", err)
	}
	if len(embeddings) != 2 || len(embeddings[0]) != 3 {
		t.Fatalf("embeddings = %+v", embeddings)
	}
}

func TestSemanticRecallGeminiHybridBlocksWithoutConfigKey(t *testing.T) {
	reports, provider := runSemanticRecallGeminiHybrid(context.Background(), semanticRecallConfig{
		GeminiConfigDB:            filepath.Join(t.TempDir(), "missing.sqlite"),
		GeminiConfigKey:           "GEMINI_API_KEY",
		GeminiAPIBase:             "https://generativelanguage.googleapis.com/v1beta",
		EmbeddingModel:            "gemini-embedding-001",
		EmbeddingOutputDimensions: 3,
	}, []semanticRecallChunk{{ChunkID: "chunk_a", Path: "docs/architecture/a.md", TextForIndex: "alpha"}}, semanticRecallQueries()[:1])
	if provider.Status != "provider_blocked" || len(reports) != 2 || !reports[0].ProviderBlocked ||
		strings.Contains(provider.ErrorSummary, "/") {
		t.Fatalf("blocked reports=%+v provider=%+v", reports, provider)
	}
}

func TestSemanticRecallLocalHybridBlocksWithoutOllama(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model unavailable", http.StatusNotFound)
	}))
	defer server.Close()

	reports, ollama := runSemanticRecallLocalHybrid(context.Background(), semanticRecallConfig{
		OllamaURL:      server.URL,
		EmbeddingModel: "embeddinggemma",
	}, []semanticRecallChunk{{ChunkID: "chunk_a", Path: "docs/architecture/a.md", TextForIndex: "alpha"}}, semanticRecallQueries()[:1])
	if ollama.Status != "environment_blocked" || len(reports) != 2 || !reports[0].EnvironmentBlocked {
		t.Fatalf("blocked reports=%+v ollama=%+v", reports, ollama)
	}
}

func TestExecuteSemanticRecallLocalHybridIncludesBaselineWhenBlocked(t *testing.T) {
	runRoot := t.TempDir()
	reportDir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model unavailable", http.StatusNotFound)
	}))
	defer server.Close()

	config := semanticRecallConfig{
		Mode:           semanticRecallModeLocalHybrid,
		RunRoot:        runRoot,
		ReportDir:      reportDir,
		ReportName:     "test-semantic-recall-local",
		OllamaURL:      server.URL,
		EmbeddingModel: "embeddinggemma",
	}
	var stdout bytes.Buffer
	if err := executeSemanticRecall(context.Background(), config, &stdout); err != nil {
		t.Fatalf("execute semantic recall: %v", err)
	}
	jsonContent := string(readReportForTest(t, filepath.Join(reportDir, "test-semantic-recall-local.json")))
	if !strings.Contains(jsonContent, `"method": "current_lexical_fts"`) ||
		!strings.Contains(jsonContent, `"method": "local_hybrid_rrf"`) ||
		!strings.Contains(jsonContent, `"status": "environment_blocked"`) {
		t.Fatalf("local-hybrid report missing baseline or blocked hybrid evidence: %s", jsonContent)
	}
}

func TestExecuteSemanticRecallLexicalFallbackWritesReducedReports(t *testing.T) {
	runRoot := t.TempDir()
	reportDir := t.TempDir()
	config := semanticRecallConfig{
		Mode:           semanticRecallModeLexicalFallback,
		RunRoot:        runRoot,
		ReportDir:      reportDir,
		ReportName:     "test-semantic-recall",
		OllamaURL:      "http://localhost:11434",
		EmbeddingModel: "embeddinggemma",
	}
	var stdout bytes.Buffer
	if err := executeSemanticRecall(context.Background(), config, &stdout); err != nil {
		t.Fatalf("execute semantic recall: %v", err)
	}
	jsonContent := string(readReportForTest(t, filepath.Join(reportDir, "test-semantic-recall.json")))
	markdownContent := string(readReportForTest(t, filepath.Join(reportDir, "test-semantic-recall.md")))
	for _, content := range []string{jsonContent, markdownContent} {
		assertReducedReportForTest(t, content, runRoot)
		if !strings.Contains(content, "lexical_token_overlap_fallback") ||
			strings.Contains(content, `"production_search_default_changed": true`) {
			t.Fatalf("semantic recall report missing lexical fallback or changed default: %s", content)
		}
	}
}
