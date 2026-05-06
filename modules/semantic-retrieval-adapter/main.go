package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	_ "modernc.org/sqlite"
)

const (
	providerOllama = "ollama"
	providerGemini = "gemini"

	geminiAPIBase = "https://generativelanguage.googleapis.com/v1beta"
	geminiKeyName = "GEMINI_API_KEY"

	ollamaEmbedBatchSize          = 4
	semanticChunkTargetCharacters = 1800
)

var wordPattern = regexp.MustCompile(`[a-z0-9]+`)

type searchRequest struct {
	Query                     string `json:"query,omitempty"`
	PathPrefix                string `json:"path_prefix,omitempty"`
	Tag                       string `json:"tag,omitempty"`
	MetadataKey               string `json:"metadata_key,omitempty"`
	MetadataValue             string `json:"metadata_value,omitempty"`
	Limit                     int    `json:"limit,omitempty"`
	Provider                  string `json:"provider,omitempty"`
	FallbackProvider          string `json:"fallback_provider,omitempty"`
	OllamaURL                 string `json:"ollama_url,omitempty"`
	EmbeddingModel            string `json:"embedding_model,omitempty"`
	GeminiAPIBase             string `json:"gemini_api_base,omitempty"`
	GeminiConfigKey           string `json:"gemini_config_key,omitempty"`
	EmbeddingOutputDimensions int    `json:"embedding_output_dimensions,omitempty"`
	CacheDir                  string `json:"cache_dir,omitempty"`

	tagProvided bool
}

type searchResponse struct {
	SchemaVersion        string         `json:"schema_version"`
	Module               moduleMetadata `json:"module"`
	Query                string         `json:"query"`
	PathPrefix           string         `json:"path_prefix,omitempty"`
	Tag                  string         `json:"tag,omitempty"`
	MetadataKey          string         `json:"metadata_key,omitempty"`
	MetadataValue        string         `json:"metadata_value,omitempty"`
	Provider             providerStatus `json:"provider"`
	Cache                cacheStatus    `json:"cache"`
	Results              []semanticHit  `json:"results,omitempty"`
	HitCount             int            `json:"hit_count"`
	DuplicateChunks      int            `json:"duplicate_chunks"`
	Ranking              string         `json:"ranking"`
	SearchStatus         string         `json:"search_status"`
	PrivacyDisclosure    string         `json:"privacy_disclosure"`
	ValidationBoundaries string         `json:"validation_boundaries"`
	AuthorityLimits      string         `json:"authority_limits"`
	SafetyPass           string         `json:"safety_pass"`
	CapabilityPass       string         `json:"capability_pass"`
	UXQuality            string         `json:"ux_quality"`
	AgentHandoff         agentHandoff   `json:"agent_handoff"`
}

type moduleMetadata struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type providerStatus struct {
	Provider         string  `json:"provider"`
	Model            string  `json:"model"`
	Status           string  `json:"status"`
	CredentialRef    string  `json:"credential_ref,omitempty"`
	EmbeddingDims    int     `json:"embedding_dimensions,omitempty"`
	RequestCount     int     `json:"request_count,omitempty"`
	RetryCount       int     `json:"retry_count,omitempty"`
	BackoffSeconds   float64 `json:"backoff_seconds,omitempty"`
	ErrorSummary     string  `json:"error_summary,omitempty"`
	FallbackProvider string  `json:"fallback_provider,omitempty"`
}

type cacheStatus struct {
	Status       string `json:"status"`
	CacheRef     string `json:"cache_ref"`
	ChunkCount   int    `json:"chunk_count"`
	RebuiltCount int    `json:"rebuilt_count,omitempty"`
}

type semanticHit struct {
	Rank      int        `json:"rank"`
	Score     float64    `json:"score"`
	DocID     string     `json:"doc_id"`
	ChunkID   string     `json:"chunk_id"`
	Title     string     `json:"title"`
	Snippet   string     `json:"snippet,omitempty"`
	Citations []citation `json:"citations,omitempty"`
}

type citation struct {
	DocID     string `json:"doc_id"`
	ChunkID   string `json:"chunk_id"`
	Path      string `json:"path"`
	Heading   string `json:"heading,omitempty"`
	LineStart int    `json:"line_start"`
	LineEnd   int    `json:"line_end"`
}

type agentHandoff struct {
	Summary                       string   `json:"summary"`
	EvidenceInspected             []string `json:"evidence_inspected,omitempty"`
	FollowUpPrimitiveInspection   string   `json:"follow_up_primitive_inspection"`
	ApprovalOrConfigurationNeeded string   `json:"approval_or_configuration_needed,omitempty"`
}

type semanticChunk struct {
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

type sectionPart struct {
	lineStart int
	lineEnd   int
	lines     []string
}

type cacheFile struct {
	SchemaVersion string          `json:"schema_version"`
	Provider      string          `json:"provider"`
	Model         string          `json:"model"`
	Dimensions    int             `json:"dimensions"`
	CorpusHash    string          `json:"corpus_hash"`
	Chunks        []semanticChunk `json:"chunks"`
}

type ollamaClient struct {
	baseURL string
	client  *http.Client
}

type ollamaEmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

type geminiClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	sleep      func(time.Duration)
}

type geminiEmbeddingStats struct {
	RequestCount   int
	RetryCount     int
	BackoffSeconds float64
}

type geminiBatchEmbedResponse struct {
	Embeddings []struct {
		Values []float64 `json:"values"`
	} `json:"embeddings"`
}

type geminiHTTPError struct {
	StatusCode int
	Body       string
	RetryAfter time.Duration
}

func (e geminiHTTPError) Error() string {
	return fmt.Sprintf("gemini returned HTTP %d: %s", e.StatusCode, e.Body)
}

func (request *searchRequest) UnmarshalJSON(data []byte) error {
	type searchRequestAlias searchRequest
	var decoded searchRequestAlias
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*request = searchRequest(decoded)
	_, request.tagProvided = raw["tag"]
	return nil
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "search" {
		_, _ = fmt.Fprintln(stderr, "usage: semantic-retrieval-adapter search [--db path] < request.json")
		return 2
	}
	fs := flag.NewFlagSet("semantic-retrieval-adapter search", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dbPath := fs.String("db", "", "OpenClerk runtime database path")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", fs.Args())
		return 2
	}
	var request searchRequest
	decoder := json.NewDecoder(os.Stdin)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		_, _ = fmt.Fprintf(stderr, "decode request: %v\n", err)
		return 1
	}
	response, err := executeSearch(context.Background(), runclient.Config{DatabasePath: *dbPath}, request)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode response: %v\n", err)
		return 1
	}
	return 0
}

func executeSearch(ctx context.Context, config runclient.Config, request searchRequest) (searchResponse, error) {
	request = normalizeRequest(request)
	if request.Query == "" {
		return searchResponse{}, errors.New("query is required")
	}
	if request.Limit < 1 || request.Limit > 100 {
		return searchResponse{}, errors.New("limit must be between 1 and 100")
	}
	if unsafePrefix(request.PathPrefix) {
		return searchResponse{}, errors.New("path_prefix must stay inside the vault root")
	}
	if err := validateProviderSettings(request); err != nil {
		return searchResponse{}, err
	}
	if request.tagProvided && request.Tag == "" {
		return searchResponse{}, errors.New("tag must be non-empty")
	}
	if request.Tag != "" && (request.MetadataKey != "" || request.MetadataValue != "") {
		return searchResponse{}, errors.New("tag cannot be combined with metadata_key or metadata_value")
	}
	if (request.MetadataKey == "") != (request.MetadataValue == "") {
		return searchResponse{}, errors.New("metadata_key and metadata_value must be provided together")
	}
	paths, err := runclient.ResolvePaths(config)
	if err != nil {
		return searchResponse{}, err
	}
	client, err := runclient.OpenReadOnly(config)
	if err != nil {
		return searchResponse{}, err
	}
	defer func() {
		_ = client.Close()
	}()
	chunks, err := loadChunks(ctx, client, request)
	if err != nil {
		return searchResponse{}, err
	}
	if len(chunks) == 0 {
		return baseResponse(request, providerStatus{Provider: request.Provider, Model: request.EmbeddingModel, Status: "empty_corpus"}, cacheStatus{Status: "not_used"}), nil
	}
	provider := providerStatus{Provider: request.Provider, Model: request.EmbeddingModel, Status: "completed"}
	cache, cachePath, cacheRef := cacheForRequest(request, chunks)
	cached, cacheReadStatus := readCache(cachePath, cache, chunks)
	cacheStatusValue := cacheStatus{Status: cacheReadStatus.Status, CacheRef: cacheRef, ChunkCount: len(chunks)}
	if len(cached) == 0 {
		vectorChunks, status, err := embedChunks(ctx, request, paths.DatabasePath, chunks)
		cacheHitAfterFallback := false
		provider = status
		if err != nil && request.FallbackProvider != "" && request.FallbackProvider != request.Provider {
			originalFallbackProvider := request.FallbackProvider
			fallbackRequest := request
			fallbackRequest.Provider = request.FallbackProvider
			fallbackRequest.EmbeddingModel = defaultModel(fallbackRequest.Provider, "")
			request = fallbackRequest
			cache, cachePath, cacheRef = cacheForRequest(request, chunks)
			if fallbackCached, _ := readCache(cachePath, cache, chunks); len(fallbackCached) > 0 {
				vectorChunks = fallbackCached
				status = providerStatus{
					Provider:         request.Provider,
					Model:            request.EmbeddingModel,
					Status:           "completed",
					EmbeddingDims:    len(fallbackCached[0].Vector),
					FallbackProvider: originalFallbackProvider,
				}
				err = nil
				cacheStatusValue = cacheStatus{Status: "hit", CacheRef: cacheRef, ChunkCount: len(fallbackCached)}
				cacheHitAfterFallback = true
			} else {
				vectorChunks, status, err = embedChunks(ctx, request, paths.DatabasePath, chunks)
				status.FallbackProvider = originalFallbackProvider
			}
			provider = status
		}
		if err != nil {
			response := baseResponse(request, provider, cacheStatus{Status: "provider_blocked", CacheRef: cacheRef, ChunkCount: len(chunks)})
			response.SearchStatus = "provider_blocked"
			response.AgentHandoff.ApprovalOrConfigurationNeeded = "configure local Ollama or runtime_config:GEMINI_API_KEY, then rerun semantic-retrieval-adapter search"
			return response, nil
		}
		chunks = vectorChunks
		if !cacheHitAfterFallback {
			cache.Dimensions = provider.EmbeddingDims
			cache.Chunks = chunks
			_ = writeCache(cachePath, cache)
			cacheStatusValue = cacheStatus{Status: "rebuilt", CacheRef: cacheRef, ChunkCount: len(chunks), RebuiltCount: len(chunks)}
		}
	} else {
		chunks = cached
		cacheStatusValue = cacheStatus{Status: "hit", CacheRef: cacheRef, ChunkCount: len(chunks)}
		provider.EmbeddingDims = len(chunks[0].Vector)
	}
	queryVector, status, err := embedQuery(ctx, request, paths.DatabasePath, request.Query)
	provider.RequestCount += status.RequestCount
	provider.RetryCount += status.RetryCount
	provider.BackoffSeconds = math.Round((provider.BackoffSeconds+status.BackoffSeconds)*100) / 100
	if status.EmbeddingDims > 0 {
		provider.EmbeddingDims = status.EmbeddingDims
	}
	if status.CredentialRef != "" {
		provider.CredentialRef = status.CredentialRef
	}
	if err != nil {
		provider.Status = "provider_blocked"
		provider.ErrorSummary = errorSummary(err)
		response := baseResponse(request, provider, cacheStatusValue)
		response.SearchStatus = "provider_blocked"
		return response, nil
	}
	hits, duplicates := rankChunks(chunks, queryVector, request)
	response := baseResponse(request, provider, cacheStatusValue)
	response.Results = hits
	response.HitCount = len(hits)
	response.DuplicateChunks = duplicates
	response.SearchStatus = "completed"
	response.AgentHandoff.EvidenceInspected = topPaths(hits, 3)
	return response, nil
}

func normalizeRequest(request searchRequest) searchRequest {
	request.Query = strings.TrimSpace(request.Query)
	request.PathPrefix = normalizePrefix(request.PathPrefix)
	request.Tag = strings.TrimSpace(request.Tag)
	request.MetadataKey = strings.TrimSpace(request.MetadataKey)
	request.MetadataValue = strings.TrimSpace(request.MetadataValue)
	request.Provider = strings.ToLower(strings.TrimSpace(request.Provider))
	if request.Provider == "" {
		request.Provider = providerOllama
	}
	request.FallbackProvider = strings.ToLower(strings.TrimSpace(request.FallbackProvider))
	request.OllamaURL = strings.TrimRight(strings.TrimSpace(request.OllamaURL), "/")
	if request.OllamaURL == "" {
		request.OllamaURL = "http://localhost:11434"
	}
	request.GeminiAPIBase = strings.TrimRight(strings.TrimSpace(request.GeminiAPIBase), "/")
	if request.GeminiAPIBase == "" {
		request.GeminiAPIBase = geminiAPIBase
	}
	request.GeminiConfigKey = strings.TrimSpace(request.GeminiConfigKey)
	if request.GeminiConfigKey == "" {
		request.GeminiConfigKey = geminiKeyName
	}
	request.EmbeddingModel = defaultModel(request.Provider, strings.TrimSpace(request.EmbeddingModel))
	if request.EmbeddingOutputDimensions == 0 {
		request.EmbeddingOutputDimensions = 3072
	}
	if request.Limit == 0 {
		request.Limit = 10
	}
	return request
}

func validateProviderSettings(request searchRequest) error {
	if err := validateProviderName(request.Provider); err != nil {
		return err
	}
	if request.FallbackProvider != "" {
		if err := validateProviderName(request.FallbackProvider); err != nil {
			return fmt.Errorf("fallback_provider must be %q or %q", providerOllama, providerGemini)
		}
	}
	for _, provider := range activeProviders(request) {
		switch provider {
		case providerOllama:
			if err := validateOllamaBaseURL(request.OllamaURL); err != nil {
				return err
			}
		case providerGemini:
			if err := validateGeminiSettings(request.GeminiAPIBase, request.GeminiConfigKey); err != nil {
				return err
			}
		}
	}
	return nil
}

func activeProviders(request searchRequest) []string {
	providers := []string{request.Provider}
	if request.FallbackProvider != "" && request.FallbackProvider != request.Provider {
		providers = append(providers, request.FallbackProvider)
	}
	return providers
}

func validateProviderName(provider string) error {
	if provider != providerOllama && provider != providerGemini {
		return fmt.Errorf("provider must be %q or %q", providerOllama, providerGemini)
	}
	return nil
}

func validateOllamaBaseURL(raw string) error {
	parsed, err := parseProviderBaseURL(raw)
	if err != nil {
		return fmt.Errorf("invalid ollama_url: %w", err)
	}
	if !isLoopbackHost(parsed.Hostname()) {
		return errors.New("ollama_url must be a loopback HTTP URL")
	}
	return nil
}

func validateGeminiSettings(rawBase string, keyName string) error {
	parsed, err := parseProviderBaseURL(rawBase)
	if err != nil {
		return fmt.Errorf("invalid gemini_api_base: %w", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "generativelanguage.googleapis.com" || parsed.Path != "/v1beta" {
		return errors.New("gemini_api_base must be https://generativelanguage.googleapis.com/v1beta")
	}
	if keyName != geminiKeyName {
		return errors.New("gemini_config_key must be GEMINI_API_KEY")
	}
	return nil
}

func parseProviderBaseURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("scheme must be http or https")
	}
	if parsed.Host == "" {
		return nil, errors.New("host is required")
	}
	if parsed.User != nil {
		return nil, errors.New("userinfo is not allowed")
	}
	if parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return nil, errors.New("query and fragment are not allowed")
	}
	return parsed, nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func defaultModel(provider string, model string) string {
	if model != "" {
		return model
	}
	if provider == providerGemini {
		return "gemini-embedding-001"
	}
	return "embeddinggemma"
}

func unsafePrefix(prefix string) bool {
	return filepath.IsAbs(prefix) || strings.HasPrefix(prefix, "/") || prefix == "." || prefix == ".." || strings.HasPrefix(prefix, "../")
}

func normalizePrefix(raw string) string {
	trimmed := strings.TrimSpace(filepath.ToSlash(raw))
	if trimmed == "" || unsafePrefix(trimmed) {
		return trimmed
	}
	trailingSlash := strings.HasSuffix(trimmed, "/")
	clean := path.Clean(trimmed)
	if clean == "." {
		return ""
	}
	if trailingSlash && !strings.HasSuffix(clean, "/") {
		clean += "/"
	}
	return clean
}

func baseResponse(request searchRequest, provider providerStatus, cache cacheStatus) searchResponse {
	privacy := "local Ollama embeddings keep corpus/query text on this machine"
	if provider.Provider == providerGemini {
		privacy = "Gemini provider embeddings send corpus/query text to a remote provider using redacted runtime_config credentials"
	}
	return searchResponse{
		SchemaVersion: "openclerk_semantic_retrieval.v1",
		Module: moduleMetadata{
			Name:    "semantic-retrieval-adapter",
			Version: "0.1.0",
		},
		Query:                request.Query,
		PathPrefix:           request.PathPrefix,
		Tag:                  request.Tag,
		MetadataKey:          request.MetadataKey,
		MetadataValue:        request.MetadataValue,
		Provider:             provider,
		Cache:                cache,
		Ranking:              "hybrid_rrf_vector_lexical",
		SearchStatus:         "completed",
		PrivacyDisclosure:    privacy,
		ValidationBoundaries: "optional OpenClerk module; read-only runner access; no core search default change, no durable OpenClerk schema migration, no committed cache, no provider key output",
		AuthorityLimits:      "semantic similarity is retrieval evidence only; canonical markdown citations and approved OpenClerk runner writes remain authority",
		SafetyPass:           "yes",
		CapabilityPass:       "recorded",
		UXQuality:            "building_block_optional_module",
		AgentHandoff: agentHandoff{
			Summary:                     "semantic retrieval adapter returned citation-bearing hybrid results without changing openclerk retrieval search",
			FollowUpPrimitiveInspection: "use openclerk retrieval search, get_document, provenance_events, and projection_states for authority drill-down before durable writes",
		},
	}
}

func loadChunks(ctx context.Context, client *runclient.Client, request searchRequest) ([]semanticChunk, error) {
	documents := []domain.DocumentSummary{}
	cursor := ""
	for {
		result, err := client.ListDocuments(ctx, domain.DocumentListQuery{
			PathPrefix:    request.PathPrefix,
			MetadataKey:   request.MetadataKey,
			MetadataValue: request.MetadataValue,
			Tag:           request.Tag,
			Limit:         100,
			Cursor:        cursor,
		})
		if err != nil {
			return nil, err
		}
		documents = append(documents, result.Documents...)
		if !result.PageInfo.HasMore {
			break
		}
		cursor = result.PageInfo.NextCursor
	}
	chunks := []semanticChunk{}
	for _, summary := range documents {
		doc, err := client.GetDocument(ctx, summary.DocID)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, chunksForDocument(doc)...)
	}
	return chunks, nil
}

func chunksForDocument(doc domain.Document) []semanticChunk {
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
	chunks := []semanticChunk{}
	for _, section := range sections {
		chunks = append(chunks, chunksForSection(doc, section.heading, section.start, section.lines)...)
	}
	return chunks
}

func chunksForSection(doc domain.Document, heading string, start int, lines []string) []semanticChunk {
	parts := splitSectionParts(start, lines)
	chunks := make([]semanticChunk, 0, len(parts))
	for idx, part := range parts {
		content := strings.TrimSpace(strings.Join(part.lines, "\n"))
		if content == "" {
			continue
		}
		indexText := strings.Join([]string{doc.Title, doc.Path, heading, content}, "\n")
		hash := hashString(indexText)
		chunkIDMaterial := doc.DocID + "\n" + heading + "\n" + hash
		if len(parts) > 1 {
			chunkIDMaterial = fmt.Sprintf("%s\n%s\n%d\n%s", doc.DocID, heading, idx, hash)
		}
		chunks = append(chunks, semanticChunk{
			ChunkID:      "chunk_" + hashString(chunkIDMaterial)[:16],
			DocID:        doc.DocID,
			Path:         doc.Path,
			Title:        doc.Title,
			Heading:      heading,
			Content:      content,
			LineStart:    part.lineStart,
			LineEnd:      part.lineEnd,
			TextForIndex: indexText,
			Hash:         hash,
		})
	}
	return chunks
}

func splitSectionParts(start int, lines []string) []sectionPart {
	parts := []sectionPart{}
	current := []string{}
	currentStart := start
	currentLength := 0
	flush := func(end int) {
		if strings.TrimSpace(strings.Join(current, "\n")) == "" {
			current = nil
			currentLength = 0
			return
		}
		copied := append([]string(nil), current...)
		parts = append(parts, sectionPart{lineStart: currentStart, lineEnd: end, lines: copied})
		current = nil
		currentLength = 0
	}
	for idx, line := range lines {
		lineNo := start + idx
		if len([]rune(line)) > semanticChunkTargetCharacters {
			if len(current) > 0 {
				flush(lineNo - 1)
			}
			for _, segment := range splitLongLine(line, semanticChunkTargetCharacters) {
				parts = append(parts, sectionPart{lineStart: lineNo, lineEnd: lineNo, lines: []string{segment}})
			}
			currentStart = lineNo + 1
			continue
		}
		lineLength := len([]rune(line)) + 1
		if len(current) > 0 && currentLength+lineLength > semanticChunkTargetCharacters {
			flush(lineNo - 1)
			currentStart = lineNo
		}
		if len(current) == 0 {
			currentStart = lineNo
		}
		current = append(current, line)
		currentLength += lineLength
	}
	if len(current) > 0 {
		flush(start + len(lines) - 1)
	}
	return parts
}

func splitLongLine(line string, limit int) []string {
	if limit <= 0 {
		return []string{line}
	}
	runes := []rune(line)
	parts := []string{}
	for len(runes) > limit {
		parts = append(parts, string(runes[:limit]))
		runes = runes[limit:]
	}
	if len(runes) > 0 || len(parts) == 0 {
		parts = append(parts, string(runes))
	}
	return parts
}

func cacheForRequest(request searchRequest, chunks []semanticChunk) (cacheFile, string, string) {
	corpusHash := corpusHash(chunks)
	key := hashString(strings.Join([]string{
		request.Provider,
		request.EmbeddingModel,
		fmt.Sprint(request.EmbeddingOutputDimensions),
		request.PathPrefix,
		request.Tag,
		request.MetadataKey,
		request.MetadataValue,
		corpusHash,
	}, "\n"))
	cacheDir := strings.TrimSpace(request.CacheDir)
	if cacheDir == "" {
		if userCache, err := os.UserCacheDir(); err == nil {
			cacheDir = filepath.Join(userCache, "openclerk", "semantic-retrieval-adapter")
		} else {
			cacheDir = filepath.Join(os.TempDir(), "openclerk-semantic-retrieval-adapter-cache")
		}
	}
	cache := cacheFile{
		SchemaVersion: "semantic_retrieval_adapter_cache.v1",
		Provider:      request.Provider,
		Model:         request.EmbeddingModel,
		CorpusHash:    corpusHash,
	}
	return cache, filepath.Join(cacheDir, key+".json"), "user_cache:semantic-retrieval-adapter/" + key + ".json"
}

func readCache(cachePath string, expected cacheFile, chunks []semanticChunk) ([]semanticChunk, cacheStatus) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, cacheStatus{Status: "miss"}
	}
	var cached cacheFile
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, cacheStatus{Status: "stale"}
	}
	if cached.SchemaVersion != expected.SchemaVersion || cached.Provider != expected.Provider || cached.Model != expected.Model || cached.CorpusHash != expected.CorpusHash || len(cached.Chunks) != len(chunks) {
		return nil, cacheStatus{Status: "stale"}
	}
	for idx, chunk := range chunks {
		if cached.Chunks[idx].Hash != chunk.Hash || len(cached.Chunks[idx].Vector) == 0 {
			return nil, cacheStatus{Status: "stale"}
		}
	}
	return cached.Chunks, cacheStatus{Status: "hit"}
}

func writeCache(cachePath string, cache cacheFile) error {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath, data, 0o600)
}

func embedChunks(ctx context.Context, request searchRequest, dbPath string, chunks []semanticChunk) ([]semanticChunk, providerStatus, error) {
	inputs := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		inputs = append(inputs, chunk.TextForIndex)
	}
	vectors, status, err := embedTexts(ctx, request, dbPath, inputs, true)
	if err != nil {
		return nil, status, err
	}
	for idx := range chunks {
		chunks[idx].Vector = vectors[idx]
	}
	return chunks, status, nil
}

func embedQuery(ctx context.Context, request searchRequest, dbPath string, query string) ([]float64, providerStatus, error) {
	vectors, status, err := embedTexts(ctx, request, dbPath, []string{query}, false)
	if err != nil {
		return nil, status, err
	}
	if len(vectors) == 0 {
		return nil, status, errors.New("provider returned no query vector")
	}
	return vectors[0], status, nil
}

func embedTexts(ctx context.Context, request searchRequest, dbPath string, inputs []string, documents bool) ([][]float64, providerStatus, error) {
	status := providerStatus{Provider: request.Provider, Model: request.EmbeddingModel, Status: "completed"}
	switch request.Provider {
	case providerOllama:
		client := ollamaClient{baseURL: request.OllamaURL, client: providerHTTPClient(60 * time.Second)}
		vectors, requestCount, err := client.embed(ctx, request.EmbeddingModel, inputs)
		status.RequestCount = requestCount
		if err != nil {
			status.Status = "provider_blocked"
			status.ErrorSummary = errorSummary(err)
			return nil, status, err
		}
		if len(vectors) > 0 {
			status.EmbeddingDims = len(vectors[0])
		}
		return vectors, status, nil
	case providerGemini:
		key, ref, err := readGeminiAPIKey(ctx, dbPath, request.GeminiConfigKey)
		status.CredentialRef = ref
		if err != nil {
			status.Status = "provider_blocked"
			status.ErrorSummary = errorSummary(err)
			return nil, status, err
		}
		client := geminiClient{baseURL: request.GeminiAPIBase, apiKey: key, httpClient: providerHTTPClient(45 * time.Second), sleep: time.Sleep}
		taskType := "RETRIEVAL_QUERY"
		if documents {
			taskType = "RETRIEVAL_DOCUMENT"
		}
		vectors, stats, err := client.embed(ctx, request.EmbeddingModel, inputs, taskType, request.EmbeddingOutputDimensions)
		status.RequestCount = stats.RequestCount
		status.RetryCount = stats.RetryCount
		status.BackoffSeconds = stats.BackoffSeconds
		if err != nil {
			status.Status = "provider_blocked"
			status.ErrorSummary = errorSummary(err)
			return nil, status, err
		}
		if len(vectors) > 0 {
			status.EmbeddingDims = len(vectors[0])
		}
		return vectors, status, nil
	default:
		return nil, status, fmt.Errorf("provider must be %q or %q", providerOllama, providerGemini)
	}
}

func providerHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func rankChunks(chunks []semanticChunk, queryVector []float64, request searchRequest) ([]semanticHit, int) {
	type scored struct {
		chunk semanticChunk
		score float64
	}
	vectorRanked := make([]scored, 0, len(chunks))
	for _, chunk := range chunks {
		vectorRanked = append(vectorRanked, scored{chunk: chunk, score: dot(queryVector, chunk.Vector)})
	}
	sort.SliceStable(vectorRanked, func(i, j int) bool {
		if vectorRanked[i].score == vectorRanked[j].score {
			return vectorRanked[i].chunk.ChunkID < vectorRanked[j].chunk.ChunkID
		}
		return vectorRanked[i].score > vectorRanked[j].score
	})
	vectorRanks := map[string]int{}
	for idx, item := range vectorRanked {
		vectorRanks[item.chunk.ChunkID] = idx + 1
	}
	lexicalRanks := lexicalChunkRanks(chunks, request.Query)
	scoredChunks := make([]scored, 0, len(chunks))
	for _, chunk := range chunks {
		score := rrfScore(vectorRanks[chunk.ChunkID])
		if rank, ok := lexicalRanks[chunk.ChunkID]; ok {
			score += rrfScore(rank)
		}
		scoredChunks = append(scoredChunks, scored{chunk: chunk, score: score})
	}
	sort.SliceStable(scoredChunks, func(i, j int) bool {
		if scoredChunks[i].score == scoredChunks[j].score {
			return scoredChunks[i].chunk.ChunkID < scoredChunks[j].chunk.ChunkID
		}
		return scoredChunks[i].score > scoredChunks[j].score
	})
	results := []semanticHit{}
	seen := map[string]struct{}{}
	duplicates := 0
	for _, item := range scoredChunks {
		if _, ok := seen[item.chunk.Path]; ok {
			duplicates++
			continue
		}
		seen[item.chunk.Path] = struct{}{}
		results = append(results, semanticHit{
			Rank:    len(results) + 1,
			Score:   math.Round(item.score*1000000) / 1000000,
			DocID:   item.chunk.DocID,
			ChunkID: item.chunk.ChunkID,
			Title:   item.chunk.Title,
			Snippet: snippet(item.chunk.Content, request.Query),
			Citations: []citation{{
				DocID:     item.chunk.DocID,
				ChunkID:   item.chunk.ChunkID,
				Path:      item.chunk.Path,
				Heading:   item.chunk.Heading,
				LineStart: item.chunk.LineStart,
				LineEnd:   item.chunk.LineEnd,
			}},
		})
		if len(results) >= request.Limit {
			break
		}
	}
	return results, duplicates
}

func lexicalChunkRanks(chunks []semanticChunk, query string) map[string]int {
	tokens := wordPattern.FindAllString(strings.ToLower(query), -1)
	if len(tokens) == 0 {
		return map[string]int{}
	}
	type scored struct {
		chunkID string
		score   float64
	}
	scoredChunks := []scored{}
	for _, chunk := range chunks {
		text := strings.ToLower(strings.Join([]string{chunk.Title, chunk.Path, chunk.Heading, chunk.Content}, "\n"))
		score := 0.0
		for _, token := range tokens {
			if strings.Contains(text, token) {
				score++
			}
		}
		if score > 0 {
			scoredChunks = append(scoredChunks, scored{chunkID: chunk.ChunkID, score: score})
		}
	}
	sort.SliceStable(scoredChunks, func(i, j int) bool {
		if scoredChunks[i].score == scoredChunks[j].score {
			return scoredChunks[i].chunkID < scoredChunks[j].chunkID
		}
		return scoredChunks[i].score > scoredChunks[j].score
	})
	ranks := map[string]int{}
	for idx, item := range scoredChunks {
		ranks[item.chunkID] = idx + 1
	}
	return ranks
}

func rrfScore(rank int) float64 {
	if rank <= 0 {
		return 0
	}
	return 1 / float64(60+rank)
}

func topPaths(hits []semanticHit, limit int) []string {
	paths := []string{}
	for idx, hit := range hits {
		if idx >= limit {
			break
		}
		if len(hit.Citations) > 0 {
			paths = append(paths, hit.Citations[0].Path)
		}
	}
	return paths
}

func (c ollamaClient) embed(ctx context.Context, model string, input []string) ([][]float64, int, error) {
	if len(input) == 0 {
		return nil, 0, nil
	}
	vectors := make([][]float64, 0, len(input))
	requestCount := 0
	for start := 0; start < len(input); start += ollamaEmbedBatchSize {
		end := start + ollamaEmbedBatchSize
		if end > len(input) {
			end = len(input)
		}
		batch := input[start:end]
		var result ollamaEmbedResponse
		requestCount++
		if err := c.postJSON(ctx, "/api/embed", map[string]any{"model": model, "input": batch}, &result); err != nil {
			return nil, requestCount, err
		}
		if len(result.Embeddings) != len(batch) {
			return nil, requestCount, fmt.Errorf("ollama returned %d embeddings for %d inputs", len(result.Embeddings), len(batch))
		}
		vectors = append(vectors, result.Embeddings...)
	}
	return vectors, requestCount, nil
}

func (c ollamaClient) postJSON(ctx context.Context, endpoint string, payload any, result any) error {
	return postJSON(ctx, c.client, c.baseURL+endpoint, payload, result, nil, func(statusCode int, body string, _ http.Header) error {
		return fmt.Errorf("ollama %s returned HTTP %d: %s", endpoint, statusCode, body)
	})
}

func readGeminiAPIKey(ctx context.Context, dbPath string, keyName string) (string, string, error) {
	ref := "runtime_config:" + keyName
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return "", ref, errors.New("open Gemini runtime config database")
	}
	defer func() {
		_ = db.Close()
	}()
	var value string
	err = db.QueryRowContext(ctx, `SELECT value_text FROM runtime_config WHERE key_name = ?`, keyName).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ref, fmt.Errorf("runtime config key %q not found", keyName)
	}
	if err != nil {
		return "", ref, fmt.Errorf("read runtime config key %q", keyName)
	}
	if strings.TrimSpace(value) == "" {
		return "", ref, fmt.Errorf("runtime config key %q is empty", keyName)
	}
	return strings.TrimSpace(value), ref, nil
}

func (c geminiClient) embed(ctx context.Context, model string, input []string, taskType string, outputDimensions int) ([][]float64, geminiEmbeddingStats, error) {
	if len(input) == 0 {
		return nil, geminiEmbeddingStats{}, nil
	}
	const batchSize = 4
	vectors := make([][]float64, 0, len(input))
	stats := geminiEmbeddingStats{}
	for start := 0; start < len(input); start += batchSize {
		end := start + batchSize
		if end > len(input) {
			end = len(input)
		}
		batchVectors, batchStats, err := c.embedBatch(ctx, model, input[start:end], taskType, outputDimensions)
		stats.RequestCount += batchStats.RequestCount
		stats.RetryCount += batchStats.RetryCount
		stats.BackoffSeconds += batchStats.BackoffSeconds
		if err != nil {
			return nil, stats, err
		}
		vectors = append(vectors, batchVectors...)
	}
	for idx, vector := range vectors {
		if len(vector) != outputDimensions {
			return nil, stats, fmt.Errorf("gemini embedding %d has %d dimensions, want %d", idx, len(vector), outputDimensions)
		}
	}
	stats.BackoffSeconds = math.Round(stats.BackoffSeconds*100) / 100
	return vectors, stats, nil
}

func (c geminiClient) embedBatch(ctx context.Context, model string, input []string, taskType string, outputDimensions int) ([][]float64, geminiEmbeddingStats, error) {
	requests := make([]map[string]any, 0, len(input))
	modelName := geminiModelName(model)
	for _, text := range input {
		requests = append(requests, map[string]any{
			"model": modelName,
			"content": map[string]any{
				"parts": []map[string]string{{"text": text}},
			},
			"taskType":             taskType,
			"outputDimensionality": outputDimensions,
		})
	}
	var result geminiBatchEmbedResponse
	stats, err := c.postJSONWithRetry(ctx, "/"+modelName+":batchEmbedContents", map[string]any{"requests": requests}, &result)
	if err != nil {
		return nil, stats, err
	}
	vectors := make([][]float64, 0, len(result.Embeddings))
	for _, embedding := range result.Embeddings {
		vectors = append(vectors, embedding.Values)
	}
	if len(vectors) != len(input) {
		return nil, stats, fmt.Errorf("gemini returned %d embeddings for %d inputs", len(vectors), len(input))
	}
	return vectors, stats, nil
}

func geminiModelName(model string) string {
	if strings.HasPrefix(model, "models/") {
		return model
	}
	return "models/" + strings.TrimSpace(model)
}

func (c geminiClient) postJSONWithRetry(ctx context.Context, endpoint string, payload any, result any) (geminiEmbeddingStats, error) {
	stats := geminiEmbeddingStats{}
	const maxAttempts = 7
	for attempt := 0; attempt < maxAttempts; attempt++ {
		stats.RequestCount++
		err := c.postJSON(ctx, endpoint, payload, result)
		if err == nil {
			return stats, nil
		}
		if attempt == maxAttempts-1 || !geminiRetryable(err) {
			return stats, err
		}
		stats.RetryCount++
		delay := geminiRetryDelay(err, attempt)
		stats.BackoffSeconds += delay.Seconds()
		sleep := c.sleep
		if sleep == nil {
			sleep = time.Sleep
		}
		sleep(delay)
	}
	return stats, errors.New("gemini retry loop exhausted")
}

func (c geminiClient) postJSON(ctx context.Context, endpoint string, payload any, result any) error {
	return postJSON(ctx, c.httpClient, c.baseURL+endpoint, payload, result, map[string]string{"x-goog-api-key": c.apiKey}, func(statusCode int, body string, header http.Header) error {
		return geminiHTTPError{StatusCode: statusCode, Body: body, RetryAfter: parseRetryAfter(header.Get("Retry-After"))}
	})
}

func postJSON(ctx context.Context, client *http.Client, requestURL string, payload any, result any, headers map[string]string, errorResponse func(statusCode int, body string, header http.Header) error) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return errorResponse(resp.StatusCode, strings.TrimSpace(string(data)), resp.Header)
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

func geminiRetryable(err error) bool {
	var httpErr geminiHTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == http.StatusTooManyRequests || httpErr.StatusCode == http.StatusInternalServerError || httpErr.StatusCode == http.StatusBadGateway || httpErr.StatusCode == http.StatusServiceUnavailable || httpErr.StatusCode == http.StatusGatewayTimeout
	}
	return true
}

func geminiRetryDelay(err error, attempt int) time.Duration {
	var httpErr geminiHTTPError
	if errors.As(err, &httpErr) && httpErr.RetryAfter > 0 {
		return httpErr.RetryAfter
	}
	base := time.Second << attempt
	if base > 20*time.Second {
		base = 20 * time.Second
	}
	return base + time.Duration((attempt*137)%250)*time.Millisecond
}

func parseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := time.ParseDuration(value + "s"); err == nil {
		return seconds
	}
	if when, err := http.ParseTime(value); err == nil {
		if delay := time.Until(when); delay > 0 {
			return delay
		}
	}
	return 0
}

func corpusHash(chunks []semanticChunk) string {
	parts := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		parts = append(parts, chunk.Path+"\n"+chunk.ChunkID+"\n"+chunk.Hash)
	}
	return hashString(strings.Join(parts, "\n"))
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func dot(a []float64, b []float64) float64 {
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	total := 0.0
	for idx := 0; idx < limit; idx++ {
		total += a[idx] * b[idx]
	}
	return total
}

func snippet(content string, query string) string {
	lower := strings.ToLower(content)
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return firstN(content, 180)
	}
	index := strings.Index(lower, needle)
	if index == -1 {
		return firstN(strings.TrimSpace(content), 180)
	}
	start := index - 60
	if start < 0 {
		start = 0
	}
	end := index + len(needle) + 80
	if end > len(content) {
		end = len(content)
	}
	return strings.TrimSpace(content[start:end])
}

func firstN(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func errorSummary(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) > 240 {
		message = message[:240]
	}
	return message
}
