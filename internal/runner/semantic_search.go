package runner

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	semanticSearchProviderOllama = "ollama"
	semanticSearchCacheVersion   = "openclerk_semantic_search_cache.v1"
)

type semanticSearchChunk struct {
	ChunkID      string    `json:"chunk_id"`
	DocID        string    `json:"doc_id"`
	Path         string    `json:"path"`
	Title        string    `json:"title"`
	Heading      string    `json:"heading"`
	Content      string    `json:"content"`
	LineStart    int       `json:"line_start"`
	LineEnd      int       `json:"line_end"`
	TextForIndex string    `json:"text_for_index"`
	Hash         string    `json:"hash"`
	Vector       []float64 `json:"vector,omitempty"`
}

type semanticSearchCacheFile struct {
	SchemaVersion string                `json:"schema_version"`
	Provider      string                `json:"provider"`
	Model         string                `json:"model"`
	CorpusHash    string                `json:"corpus_hash"`
	Chunks        []semanticSearchChunk `json:"chunks"`
}

type semanticOllamaEmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

func runSemanticSearch(ctx context.Context, client *runclient.Client, options SemanticSearchOptions) (SemanticSearchResult, error) {
	options = normalizeSemanticSearchOptions(options)
	chunks, err := semanticSearchLoadChunks(ctx, client, options)
	if err != nil {
		return SemanticSearchResult{}, err
	}
	base := baseSemanticSearchResult(options, len(chunks))
	if len(chunks) == 0 {
		base.Provider.Status = "empty_corpus"
		base.Cache.Status = "not_used"
		base.SearchStatus = "empty_corpus"
		return base, nil
	}

	cache, cachePath, cacheRef := semanticSearchCacheForOptions(options, chunks)
	cached := semanticSearchReadCache(cachePath, cache, chunks)
	cacheStatus := SemanticCacheStatus{Status: "hit", CacheRef: cacheRef, ChunkCount: len(chunks)}
	if len(cached) == 0 {
		embedded, provider, err := semanticSearchEmbedChunks(ctx, options, chunks)
		if err != nil {
			base.Provider = provider
			base.Provider.Status = "provider_blocked"
			base.Provider.ErrorSummary = semanticSearchErrorSummary(err)
			base.Cache = SemanticCacheStatus{Status: "provider_blocked", CacheRef: cacheRef, ChunkCount: len(chunks)}
			base.SearchStatus = "provider_blocked"
			base.AgentHandoff.AnswerSummary = "local Ollama embedding provider is blocked; start Ollama and pull the requested embedding model, then rerun semantic_search"
			return base, nil
		}
		chunks = embedded
		cache.Chunks = embedded
		_ = semanticSearchWriteCache(cachePath, cache)
		base.Provider = provider
		cacheStatus = SemanticCacheStatus{Status: "rebuilt", CacheRef: cacheRef, ChunkCount: len(chunks), RebuiltCount: len(chunks)}
	} else {
		chunks = cached
		base.Provider.EmbeddingDims = len(chunks[0].Vector)
	}

	queryVector, provider, err := semanticSearchEmbedQuery(ctx, options)
	if err != nil {
		base.Provider = provider
		base.Provider.Status = "provider_blocked"
		base.Provider.ErrorSummary = semanticSearchErrorSummary(err)
		base.Cache = cacheStatus
		base.SearchStatus = "provider_blocked"
		base.AgentHandoff.AnswerSummary = "local Ollama query embedding is blocked; start Ollama and pull the requested embedding model, then rerun semantic_search"
		return base, nil
	}
	if provider.EmbeddingDims > 0 {
		base.Provider.EmbeddingDims = provider.EmbeddingDims
	}
	base.Cache = cacheStatus
	lexicalRanks, _ := semanticSearchLexicalRanks(ctx, client, options)
	hits, duplicates := semanticSearchRank(chunks, queryVector, lexicalRanks, options.Limit)
	base.Hits = hits
	base.DuplicateChunks = duplicates
	base.SearchStatus = "completed"
	base.AgentHandoff.Evidence = semanticSearchTopPaths(hits, 3)
	return base, nil
}

func normalizeSemanticSearchOptions(options SemanticSearchOptions) SemanticSearchOptions {
	if options.Limit == 0 {
		options.Limit = 10
	}
	if options.OllamaURL == "" {
		options.OllamaURL = "http://localhost:11434"
	}
	if options.EmbeddingModel == "" {
		options.EmbeddingModel = "nomic-embed-text"
	}
	return options
}

func baseSemanticSearchResult(options SemanticSearchOptions, chunkCount int) SemanticSearchResult {
	return SemanticSearchResult{
		Query:         options.Query,
		PathPrefix:    options.PathPrefix,
		Tag:           options.Tag,
		MetadataKey:   options.MetadataKey,
		MetadataValue: options.MetadataValue,
		Provider: SemanticProviderStatus{
			Provider:  semanticSearchProviderOllama,
			Model:     options.EmbeddingModel,
			Status:    "completed",
			OllamaURL: options.OllamaURL,
		},
		Cache:                SemanticCacheStatus{Status: "miss", ChunkCount: chunkCount},
		Ranking:              "local_hybrid_rrf_vector_lexical",
		SearchStatus:         "completed",
		PrivacyDisclosure:    "local Ollama embeddings keep corpus/query text on this machine; no provider fallback is used",
		ValidationBoundaries: "explicit retrieval semantic_search mode; default search remains lexical; no provider config writes, no committed embedding cache, no durable document writes",
		AuthorityLimits:      "semantic similarity is retrieval evidence only; canonical markdown citations and approved OpenClerk runner writes remain authority",
		AgentHandoff: &AgentHandoff{
			AnswerSummary:               "semantic_search returned citation-bearing local hybrid results without changing default search ranking",
			ValidationBoundaries:        "explicit semantic mode; default lexical search is unchanged",
			AuthorityLimits:             "use canonical documents and citations for authority",
			FollowUpPrimitiveInspection: "use get_document, provenance_events, and projection_states for authority drill-down before durable writes",
		},
	}
}

func semanticSearchLoadChunks(ctx context.Context, client *runclient.Client, options SemanticSearchOptions) ([]semanticSearchChunk, error) {
	summaries := []domain.DocumentSummary{}
	cursor := ""
	for {
		page, err := client.ListDocuments(ctx, domain.DocumentListQuery{
			PathPrefix:    options.PathPrefix,
			MetadataKey:   options.MetadataKey,
			MetadataValue: options.MetadataValue,
			Tag:           options.Tag,
			Limit:         100,
			Cursor:        cursor,
		})
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, page.Documents...)
		if !page.PageInfo.HasMore {
			break
		}
		cursor = page.PageInfo.NextCursor
	}
	chunks := []semanticSearchChunk{}
	for _, summary := range summaries {
		doc, err := client.GetDocument(ctx, summary.DocID)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, semanticSearchChunksForDocument(doc)...)
	}
	return chunks, nil
}

func semanticSearchChunksForDocument(doc domain.Document) []semanticSearchChunk {
	lines := strings.Split(strings.ReplaceAll(doc.Body, "\r\n", "\n"), "\n")
	type section struct {
		heading string
		start   int
		lines   []string
	}
	sections := []section{}
	current := section{heading: doc.Title, start: 1}
	for idx, line := range lines {
		lineNo := idx + 1
		if strings.HasPrefix(line, "## ") {
			if strings.TrimSpace(strings.Join(current.lines, "\n")) != "" {
				sections = append(sections, current)
			}
			current = section{heading: strings.TrimSpace(strings.TrimPrefix(line, "## ")), start: lineNo, lines: []string{line}}
			continue
		}
		current.lines = append(current.lines, line)
	}
	if strings.TrimSpace(strings.Join(current.lines, "\n")) != "" {
		sections = append(sections, current)
	}
	chunks := make([]semanticSearchChunk, 0, len(sections))
	for _, section := range sections {
		content := strings.TrimSpace(strings.Join(section.lines, "\n"))
		lineEnd := section.start + len(section.lines) - 1
		indexText := strings.Join([]string{doc.Title, doc.Path, section.heading, content}, "\n")
		hash := semanticSearchHash(indexText)
		chunks = append(chunks, semanticSearchChunk{
			ChunkID:      "chunk_" + semanticSearchHash(doc.DocID + "\n" + section.heading + "\n" + hash)[:16],
			DocID:        doc.DocID,
			Path:         doc.Path,
			Title:        doc.Title,
			Heading:      section.heading,
			Content:      content,
			LineStart:    section.start,
			LineEnd:      lineEnd,
			TextForIndex: indexText,
			Hash:         hash,
		})
	}
	return chunks
}

func semanticSearchCacheForOptions(options SemanticSearchOptions, chunks []semanticSearchChunk) (semanticSearchCacheFile, string, string) {
	corpusHash := semanticSearchCorpusHash(chunks)
	key := semanticSearchHash(strings.Join([]string{
		semanticSearchProviderOllama,
		options.EmbeddingModel,
		options.PathPrefix,
		options.Tag,
		options.MetadataKey,
		options.MetadataValue,
		corpusHash,
	}, "\n"))
	cacheDir := options.CacheDir
	if cacheDir == "" {
		if userCache, err := os.UserCacheDir(); err == nil {
			cacheDir = filepath.Join(userCache, "openclerk", "semantic-search")
		} else {
			cacheDir = filepath.Join(os.TempDir(), "openclerk-semantic-search-cache")
		}
	}
	return semanticSearchCacheFile{
		SchemaVersion: semanticSearchCacheVersion,
		Provider:      semanticSearchProviderOllama,
		Model:         options.EmbeddingModel,
		CorpusHash:    corpusHash,
	}, filepath.Join(cacheDir, key+".json"), "user_cache:semantic-search/" + key + ".json"
}

func semanticSearchReadCache(cachePath string, expected semanticSearchCacheFile, chunks []semanticSearchChunk) []semanticSearchChunk {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil
	}
	var cached semanticSearchCacheFile
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil
	}
	if cached.SchemaVersion != expected.SchemaVersion || cached.Provider != expected.Provider || cached.Model != expected.Model || cached.CorpusHash != expected.CorpusHash || len(cached.Chunks) != len(chunks) {
		return nil
	}
	for idx, chunk := range chunks {
		if cached.Chunks[idx].Hash != chunk.Hash || len(cached.Chunks[idx].Vector) == 0 {
			return nil
		}
	}
	return cached.Chunks
}

func semanticSearchWriteCache(cachePath string, cache semanticSearchCacheFile) error {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath, data, 0o600)
}

func semanticSearchEmbedChunks(ctx context.Context, options SemanticSearchOptions, chunks []semanticSearchChunk) ([]semanticSearchChunk, SemanticProviderStatus, error) {
	inputs := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		inputs = append(inputs, chunk.TextForIndex)
	}
	vectors, status, err := semanticSearchEmbedTexts(ctx, options, inputs)
	if err != nil {
		return nil, status, err
	}
	for idx := range chunks {
		chunks[idx].Vector = vectors[idx]
	}
	return chunks, status, nil
}

func semanticSearchEmbedQuery(ctx context.Context, options SemanticSearchOptions) ([]float64, SemanticProviderStatus, error) {
	vectors, status, err := semanticSearchEmbedTexts(ctx, options, []string{options.Query})
	if err != nil {
		return nil, status, err
	}
	if len(vectors) == 0 {
		return nil, status, errors.New("ollama returned no query vector")
	}
	return vectors[0], status, nil
}

func semanticSearchEmbedTexts(ctx context.Context, options SemanticSearchOptions, inputs []string) ([][]float64, SemanticProviderStatus, error) {
	status := SemanticProviderStatus{Provider: semanticSearchProviderOllama, Model: options.EmbeddingModel, Status: "completed", OllamaURL: options.OllamaURL}
	body, err := json.Marshal(map[string]any{
		"model": options.EmbeddingModel,
		"input": inputs,
	})
	if err != nil {
		return nil, status, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(options.OllamaURL, "/")+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, status, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return nil, status, err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, status, fmt.Errorf("ollama returned HTTP %d", response.StatusCode)
	}
	var decoded semanticOllamaEmbedResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return nil, status, err
	}
	if len(decoded.Embeddings) != len(inputs) {
		return nil, status, fmt.Errorf("ollama returned %d embeddings for %d inputs", len(decoded.Embeddings), len(inputs))
	}
	if len(decoded.Embeddings) > 0 {
		status.EmbeddingDims = len(decoded.Embeddings[0])
	}
	return decoded.Embeddings, status, nil
}

func semanticSearchLexicalRanks(ctx context.Context, client *runclient.Client, options SemanticSearchOptions) (map[string]int, error) {
	search, err := client.Search(ctx, domain.SearchQuery{
		Text:          options.Query,
		PathPrefix:    options.PathPrefix,
		MetadataKey:   options.MetadataKey,
		MetadataValue: options.MetadataValue,
		Tag:           options.Tag,
		Limit:         100,
	})
	if err != nil {
		return nil, err
	}
	ranks := map[string]int{}
	for _, hit := range search.Hits {
		if _, exists := ranks[hit.DocID]; !exists {
			ranks[hit.DocID] = hit.Rank
		}
	}
	return ranks, nil
}

func semanticSearchRank(chunks []semanticSearchChunk, queryVector []float64, lexicalRanks map[string]int, limit int) ([]SearchHit, int) {
	type scored struct {
		chunk      semanticSearchChunk
		vectorRank int
		score      float64
	}
	vectorRanked := make([]scored, 0, len(chunks))
	for _, chunk := range chunks {
		vectorRanked = append(vectorRanked, scored{chunk: chunk, score: semanticSearchDot(queryVector, chunk.Vector)})
	}
	sort.SliceStable(vectorRanked, func(i, j int) bool {
		if vectorRanked[i].score == vectorRanked[j].score {
			return vectorRanked[i].chunk.ChunkID < vectorRanked[j].chunk.ChunkID
		}
		return vectorRanked[i].score > vectorRanked[j].score
	})
	bestByDoc := map[string]scored{}
	duplicates := 0
	for idx, item := range vectorRanked {
		item.vectorRank = idx + 1
		if _, exists := bestByDoc[item.chunk.DocID]; exists {
			duplicates++
			continue
		}
		bestByDoc[item.chunk.DocID] = item
	}
	docRanked := make([]scored, 0, len(bestByDoc))
	for _, item := range bestByDoc {
		item.score = 1 / float64(60+item.vectorRank)
		if lexicalRank, ok := lexicalRanks[item.chunk.DocID]; ok {
			item.score += 1 / float64(60+lexicalRank)
		}
		docRanked = append(docRanked, item)
	}
	sort.SliceStable(docRanked, func(i, j int) bool {
		if docRanked[i].score == docRanked[j].score {
			return docRanked[i].chunk.Path < docRanked[j].chunk.Path
		}
		return docRanked[i].score > docRanked[j].score
	})
	if limit > len(docRanked) {
		limit = len(docRanked)
	}
	hits := make([]SearchHit, 0, limit)
	for idx := 0; idx < limit; idx++ {
		chunk := docRanked[idx].chunk
		hits = append(hits, SearchHit{
			Rank:    idx + 1,
			Score:   math.Round(docRanked[idx].score*1000000) / 1000000,
			DocID:   chunk.DocID,
			ChunkID: chunk.ChunkID,
			Title:   chunk.Title,
			Snippet: semanticSearchSnippet(chunk.Content),
			Citations: []Citation{{
				DocID:     chunk.DocID,
				ChunkID:   chunk.ChunkID,
				Path:      chunk.Path,
				Heading:   chunk.Heading,
				LineStart: chunk.LineStart,
				LineEnd:   chunk.LineEnd,
			}},
		})
	}
	return hits, duplicates
}

func semanticSearchDot(left []float64, right []float64) float64 {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	var score float64
	for idx := 0; idx < limit; idx++ {
		score += left[idx] * right[idx]
	}
	return score
}

func semanticSearchHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func semanticSearchCorpusHash(chunks []semanticSearchChunk) string {
	parts := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		parts = append(parts, chunk.ChunkID+":"+chunk.Hash)
	}
	sort.Strings(parts)
	return semanticSearchHash(strings.Join(parts, "\n"))
}

func semanticSearchSnippet(content string) string {
	fields := strings.Fields(content)
	if len(fields) > 40 {
		fields = fields[:40]
	}
	return strings.Join(fields, " ")
}

func semanticSearchTopPaths(hits []SearchHit, limit int) []string {
	if limit > len(hits) {
		limit = len(hits)
	}
	paths := make([]string, 0, limit)
	for idx := 0; idx < limit; idx++ {
		if len(hits[idx].Citations) > 0 {
			paths = append(paths, hits[idx].Citations[0].Path)
		}
	}
	return paths
}

func semanticSearchErrorSummary(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) > 180 {
		return message[:180]
	}
	return message
}
