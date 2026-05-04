package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	semanticRecallModeLocalHybrid     = "local-hybrid"
	semanticRecallModeLexicalFallback = "lexical-fallback"
	semanticRecallModeAll             = "all"

	semanticRecallLaneName = "semantic-recall"
)

var semanticRecallWordPattern = regexp.MustCompile(`[a-z0-9]+`)

type semanticRecallConfig struct {
	Mode           string
	RunRoot        string
	ReportDir      string
	ReportName     string
	OllamaURL      string
	EmbeddingModel string
}

type semanticRecallReport struct {
	Metadata       semanticRecallMetadata       `json:"metadata"`
	Corpus         semanticRecallCorpus         `json:"corpus"`
	Ollama         semanticRecallOllama         `json:"ollama,omitempty"`
	Methods        []semanticRecallMethodReport `json:"methods"`
	FreshnessProbe semanticRecallFreshnessProbe `json:"freshness_probe"`
	Checks         semanticRecallChecks         `json:"checks"`
	Outcomes       []maturityOutcome            `json:"outcomes"`
}

type semanticRecallMetadata struct {
	Lane                     string    `json:"lane"`
	Mode                     string    `json:"mode"`
	GeneratedAt              time.Time `json:"generated_at"`
	Harness                  string    `json:"harness"`
	RunRootArtifactReference string    `json:"run_root_artifact_reference"`
	ReportName               string    `json:"report_name"`
	RawLogsCommitted         bool      `json:"raw_logs_committed"`
	RawContentCommitted      bool      `json:"raw_content_committed"`
}

type semanticRecallCorpus struct {
	Documents        int      `json:"documents"`
	Chunks           int      `json:"chunks"`
	QueryRows        int      `json:"query_rows"`
	DocumentPaths    []string `json:"document_paths,omitempty"`
	ChunkingPolicy   string   `json:"chunking_policy"`
	CitationPolicy   string   `json:"citation_policy"`
	SourceReferences []string `json:"source_references,omitempty"`
}

type semanticRecallOllama struct {
	URL              string         `json:"url"`
	Model            string         `json:"model"`
	Status           string         `json:"status"`
	Capabilities     []string       `json:"capabilities,omitempty"`
	Details          map[string]any `json:"details,omitempty"`
	EmbeddingDims    int            `json:"embedding_dimensions,omitempty"`
	ModelInfoSummary map[string]any `json:"model_info_summary,omitempty"`
	ErrorSummary     string         `json:"error_summary,omitempty"`
}

type semanticRecallMethodReport struct {
	Method             string              `json:"method"`
	Status             string              `json:"status"`
	Description        string              `json:"description"`
	HitAt3             int                 `json:"hit_at_3"`
	MRR                float64             `json:"mrr"`
	QueryCount         int                 `json:"query_count"`
	RawDuplicateHits   int                 `json:"raw_duplicate_hits"`
	TotalSeconds       float64             `json:"total_seconds"`
	Rows               []semanticRecallRow `json:"rows,omitempty"`
	EvidencePosture    string              `json:"evidence_posture"`
	ValidationBoundary string              `json:"validation_boundary"`
	CandidateOnly      bool                `json:"candidate_only"`
	EnvironmentBlocked bool                `json:"environment_blocked,omitempty"`
}

type semanticRecallRow struct {
	QueryID      string                   `json:"query_id"`
	Kind         string                   `json:"kind"`
	ExpectedPath string                   `json:"expected_path"`
	Rank         int                      `json:"rank,omitempty"`
	TopCitations []semanticRecallCitation `json:"top_citations,omitempty"`
	Hit          bool                     `json:"hit"`
}

type semanticRecallCitation struct {
	Path      string `json:"path"`
	Heading   string `json:"heading,omitempty"`
	LineStart int    `json:"line_start,omitempty"`
	LineEnd   int    `json:"line_end,omitempty"`
}

type semanticRecallFreshnessProbe struct {
	Status             string  `json:"status"`
	ChangedPath        string  `json:"changed_path,omitempty"`
	StaleChunks        int     `json:"stale_chunks,omitempty"`
	RebuiltChunks      int     `json:"rebuilt_chunks,omitempty"`
	Seconds            float64 `json:"seconds,omitempty"`
	EvidencePosture    string  `json:"evidence_posture"`
	ValidationBoundary string  `json:"validation_boundary"`
}

type semanticRecallChecks struct {
	ReducedReportOnly              bool   `json:"reduced_report_only"`
	RawLogsCommitted               bool   `json:"raw_logs_committed"`
	RawContentCommitted            bool   `json:"raw_content_committed"`
	MachineAbsoluteArtifactRefs    bool   `json:"machine_absolute_artifact_refs"`
	ProductionSearchDefaultChanged bool   `json:"production_search_default_changed"`
	Boundary                       string `json:"boundary"`
}

type semanticRecallQuery struct {
	ID           string
	Kind         string
	Text         string
	ExpectedPath string
	Aliases      []string
}

type semanticRecallChunk struct {
	ChunkID      string
	DocID        string
	Path         string
	Title        string
	Heading      string
	Content      string
	LineStart    int
	LineEnd      int
	TextForIndex string
	Hash         string
}

type semanticRecallHit struct {
	Chunk semanticRecallChunk
	Score float64
}

type semanticRecallRankedDoc struct {
	Path     string
	Score    float64
	Citation semanticRecallCitation
	ChunkID  string
	RawRank  int
}

type ollamaClient struct {
	baseURL string
	client  *http.Client
}

type ollamaShowResponse struct {
	Capabilities []string       `json:"capabilities"`
	Details      map[string]any `json:"details"`
	ModelInfo    map[string]any `json:"model_info"`
}

type ollamaEmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

func parseSemanticRecallConfig(args []string, stderr io.Writer) (semanticRecallConfig, error) {
	fs := flag.NewFlagSet("ockp semantic-recall", flag.ContinueOnError)
	fs.SetOutput(stderr)
	config := semanticRecallConfig{
		Mode:           semanticRecallModeAll,
		ReportDir:      filepath.Join("docs", "evals", "results"),
		OllamaURL:      "http://localhost:11434",
		EmbeddingModel: "embeddinggemma",
	}
	fs.StringVar(&config.Mode, "mode", config.Mode, "semantic recall mode: local-hybrid, lexical-fallback, or all")
	fs.StringVar(&config.RunRoot, "run-root", "", "directory for local generated/private semantic-recall artifacts")
	fs.StringVar(&config.ReportDir, "report-dir", config.ReportDir, "directory for reduced reports")
	fs.StringVar(&config.ReportName, "report-name", "", "base filename for reduced reports, without extension")
	fs.StringVar(&config.OllamaURL, "ollama-url", config.OllamaURL, "Ollama base URL for local embedding POC")
	fs.StringVar(&config.EmbeddingModel, "embedding-model", config.EmbeddingModel, "Ollama embedding model")
	if err := fs.Parse(args); err != nil {
		return semanticRecallConfig{}, err
	}
	if fs.NArg() != 0 {
		return semanticRecallConfig{}, fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	config.Mode = strings.ToLower(strings.TrimSpace(config.Mode))
	switch config.Mode {
	case semanticRecallModeLocalHybrid, semanticRecallModeLexicalFallback, semanticRecallModeAll:
	default:
		return semanticRecallConfig{}, fmt.Errorf("--mode must be %q, %q, or %q", semanticRecallModeLocalHybrid, semanticRecallModeLexicalFallback, semanticRecallModeAll)
	}
	if strings.TrimSpace(config.RunRoot) == "" {
		config.RunRoot = filepath.Join(os.TempDir(), fmt.Sprintf("openclerk-ockp-semantic-recall-%d", time.Now().UnixNano()))
	}
	if strings.TrimSpace(config.ReportName) == "" {
		config.ReportName = defaultSemanticRecallReportName(config.Mode)
	}
	config.ReportName = strings.TrimSpace(config.ReportName)
	if !isSafeSemanticRecallReportName(config.ReportName) {
		return semanticRecallConfig{}, errors.New("--report-name must be a safe base filename without path components")
	}
	config.OllamaURL = strings.TrimRight(strings.TrimSpace(config.OllamaURL), "/")
	config.EmbeddingModel = strings.TrimSpace(config.EmbeddingModel)
	if config.OllamaURL == "" {
		return semanticRecallConfig{}, errors.New("--ollama-url is required")
	}
	if config.EmbeddingModel == "" {
		return semanticRecallConfig{}, errors.New("--embedding-model is required")
	}
	return config, nil
}

func isSafeSemanticRecallReportName(name string) bool {
	if name == "" || name == "." || name == ".." || filepath.IsAbs(name) {
		return false
	}
	return filepath.Clean(name) == name && filepath.Base(name) == name
}

func defaultSemanticRecallReportName(mode string) string {
	switch mode {
	case semanticRecallModeLocalHybrid:
		return "ockp-semantic-recall-local-hybrid"
	case semanticRecallModeLexicalFallback:
		return "ockp-semantic-recall-lexical-fallback"
	default:
		return "ockp-semantic-recall"
	}
}

func executeSemanticRecall(ctx context.Context, config semanticRecallConfig, stdout io.Writer) error {
	start := time.Now()
	workRoot := filepath.Join(config.RunRoot, config.ReportName)
	vaultRoot := filepath.Join(workRoot, "vault")
	dbPath := filepath.Join(workRoot, "openclerk.sqlite")
	if err := os.RemoveAll(workRoot); err != nil {
		return fmt.Errorf("reset semantic recall work root: %w", err)
	}
	if err := copySemanticRecallCorpus(vaultRoot); err != nil {
		return err
	}
	if _, err := runclient.InitializePaths(runclient.Config{DatabasePath: dbPath}, vaultRoot); err != nil {
		return fmt.Errorf("initialize semantic recall runtime: %w", err)
	}
	client, err := runclient.Open(runclient.Config{DatabasePath: dbPath})
	if err != nil {
		return fmt.Errorf("open semantic recall runtime: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()
	chunks, err := buildSemanticRecallChunks(vaultRoot)
	if err != nil {
		return err
	}

	queries := semanticRecallQueries()
	report := semanticRecallReport{
		Metadata: semanticRecallMetadata{
			Lane:                     semanticRecallLaneName,
			Mode:                     config.Mode,
			GeneratedAt:              time.Now().UTC(),
			Harness:                  "scripts/agent-eval/ockp semantic-recall",
			RunRootArtifactReference: "<run-root>",
			ReportName:               config.ReportName,
			RawLogsCommitted:         false,
			RawContentCommitted:      false,
		},
		Corpus: semanticRecallCorpus{
			Documents:      len(semanticRecallCorpusPaths()),
			Chunks:         len(chunks),
			QueryRows:      len(queries),
			DocumentPaths:  semanticRecallCorpusPaths(),
			ChunkingPolicy: "eval-only heading-section chunks parsed from committed docs copied into <run-root>; index text includes title, repo-relative path, heading, and section body",
			CitationPolicy: "reports reduced repo-relative path, heading, and line span citations; canonical markdown remains authority",
			SourceReferences: []string{
				"docs/evals/results/ockp-semantic-recall-hybrid-vector-prototype.md",
				"docs/architecture/local-first-hybrid-retrieval-implementation-candidate-decision.md",
			},
		},
		Checks: semanticRecallChecks{
			ReducedReportOnly:              true,
			RawLogsCommitted:               false,
			RawContentCommitted:            false,
			MachineAbsoluteArtifactRefs:    false,
			ProductionSearchDefaultChanged: false,
			Boundary:                       "eval-only maintainer harness; no openclerk document/retrieval JSON schema change, no durable embedding store, no provider embedding default, no production search ranking change",
		},
	}

	if config.Mode == semanticRecallModeLocalHybrid || config.Mode == semanticRecallModeLexicalFallback || config.Mode == semanticRecallModeAll {
		baseline, err := runSemanticRecallLexicalBaseline(ctx, client, queries)
		if err != nil {
			return err
		}
		report.Methods = append(report.Methods, baseline)
	}
	if config.Mode == semanticRecallModeLexicalFallback || config.Mode == semanticRecallModeAll {
		report.Methods = append(report.Methods, runSemanticRecallTokenFallback(chunks, queries, false))
		report.Methods = append(report.Methods, runSemanticRecallTokenFallback(chunks, queries, true))
	}
	if config.Mode == semanticRecallModeLocalHybrid || config.Mode == semanticRecallModeAll {
		vectorReports, ollama := runSemanticRecallLocalHybrid(ctx, config, chunks, queries)
		report.Ollama = ollama
		report.Methods = append(report.Methods, vectorReports...)
		report.FreshnessProbe = runSemanticRecallFreshnessProbe(vaultRoot, chunks)
	} else {
		report.FreshnessProbe = semanticRecallFreshnessProbe{
			Status:             "not_run",
			EvidencePosture:    "freshness probe belongs to local-hybrid mode",
			ValidationBoundary: "no document mutation outside <run-root>",
		}
	}
	report.Outcomes = semanticRecallOutcomes(report)

	if err := os.MkdirAll(config.ReportDir, 0o755); err != nil {
		return fmt.Errorf("create semantic recall report dir: %w", err)
	}
	jsonPath := filepath.Join(config.ReportDir, config.ReportName+".json")
	mdPath := filepath.Join(config.ReportDir, config.ReportName+".md")
	if err := writeJSON(jsonPath, report); err != nil {
		return err
	}
	if err := writeSemanticRecallMarkdownReport(mdPath, report); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "wrote %s and %s in %.2fs\n", jsonPath, mdPath, time.Since(start).Seconds())
	return nil
}

func semanticRecallCorpusPaths() []string {
	return []string{
		"docs/architecture/agent-knowledge-plane.md",
		"docs/architecture/artifact-intake-autofiling-tags-fields-adr.md",
		"docs/architecture/eval-backed-knowledge-plane-adr.md",
		"docs/architecture/generalized-artifact-ingestion-adr.md",
		"docs/architecture/git-lifecycle-version-control-adr.md",
		"docs/architecture/harness-owned-web-search-fetch-adr.md",
		"docs/architecture/hybrid-retrieval-adr.md",
		"docs/architecture/hybrid-retrieval-promotion-decision.md",
		"docs/architecture/knowledge-configuration-v1-adr.md",
		"docs/architecture/local-first-hybrid-retrieval-implementation-candidate-decision.md",
		"docs/architecture/memory-architecture-recall-adr.md",
		"docs/architecture/structured-data-canonical-stores-adr.md",
	}
}

func semanticRecallQueries() []semanticRecallQuery {
	return []semanticRecallQuery{
		{ID: "wiki_synthesis", Kind: "concept-recall", Text: "durable wiki style knowledge plane for agents where synthesized pages reduce repeated retrieval work", ExpectedPath: "docs/architecture/agent-knowledge-plane.md", Aliases: []string{"llm wiki", "agent knowledge plane", "synthesis"}},
		{ID: "semantic_retrieval_gap", Kind: "paraphrase", Text: "semantic search should find architecture notes even when the user does not use exact source words", ExpectedPath: "docs/architecture/hybrid-retrieval-adr.md", Aliases: []string{"hybrid retrieval", "vector ranking", "semantic recall"}},
		{ID: "structured_rows_vs_notes", Kind: "synonym-drift", Text: "when should structured rows become canonical instead of keeping ordinary markdown notes", ExpectedPath: "docs/architecture/structured-data-canonical-stores-adr.md", Aliases: []string{"structured data", "canonical stores", "records"}},
		{ID: "checkpoint_not_restore", Kind: "indirect-source", Text: "a saved version checkpoint is storage history and should not be confused with semantic restore authority", ExpectedPath: "docs/architecture/git-lifecycle-version-control-adr.md", Aliases: []string{"git lifecycle", "checkpoint", "restore"}},
		{ID: "search_then_ingest", Kind: "indirect-source", Text: "public web discovery should rank candidate links before approved source ingestion writes anything durable", ExpectedPath: "docs/architecture/harness-owned-web-search-fetch-adr.md", Aliases: []string{"web search", "ingest source url", "approval boundary"}},
		{ID: "ocr_uncertain_artifact", Kind: "concept-recall", Text: "uncertain OCR or extracted artifact content should not quietly become trusted canonical knowledge", ExpectedPath: "docs/architecture/generalized-artifact-ingestion-adr.md", Aliases: []string{"artifact ingestion", "ocr", "confidence policy"}},
		{ID: "memory_no_hidden_truth", Kind: "paraphrase", Text: "memory recall must not create a hidden truth store that outranks visible canonical documents", ExpectedPath: "docs/architecture/memory-architecture-recall-adr.md", Aliases: []string{"memory architecture", "hidden authority", "canonical docs"}},
		{ID: "plan_filename_tags", Kind: "synonym-drift", Text: "the capture plan should infer useful filenames and tags from artifact content without unsafe autonomy", ExpectedPath: "docs/architecture/artifact-intake-autofiling-tags-fields-adr.md", Aliases: []string{"autofiling", "tags", "path title"}},
	}
}

func copySemanticRecallCorpus(vaultRoot string) error {
	repoRoot, err := semanticRecallRepoRoot()
	if err != nil {
		return err
	}
	for _, rel := range semanticRecallCorpusPaths() {
		content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(rel)))
		if err != nil {
			return fmt.Errorf("read semantic recall corpus doc %s: %w", rel, err)
		}
		target := filepath.Join(vaultRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create corpus doc dir: %w", err)
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			return fmt.Errorf("write corpus doc %s: %w", rel, err)
		}
	}
	return nil
}

func semanticRecallRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", errors.New("find repo root: go.mod not found")
		}
		wd = parent
	}
}

func buildSemanticRecallChunks(vaultRoot string) ([]semanticRecallChunk, error) {
	chunks := []semanticRecallChunk{}
	for _, rel := range semanticRecallCorpusPaths() {
		contentBytes, err := os.ReadFile(filepath.Join(vaultRoot, filepath.FromSlash(rel)))
		if err != nil {
			return nil, fmt.Errorf("read copied corpus doc %s: %w", rel, err)
		}
		chunks = append(chunks, semanticRecallChunksForDocument(rel, string(contentBytes))...)
	}
	return chunks, nil
}

func semanticRecallChunksForDocument(relPath string, body string) []semanticRecallChunk {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	title := strings.TrimSuffix(path.Base(relPath), path.Ext(relPath))
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			break
		}
	}
	type section struct {
		heading string
		start   int
		lines   []string
	}
	sections := []section{}
	current := section{heading: title, start: 1}
	for idx, line := range lines {
		lineNo := idx + 1
		if strings.HasPrefix(line, "## ") {
			if len(strings.TrimSpace(strings.Join(current.lines, "\n"))) > 0 {
				sections = append(sections, current)
			}
			current = section{heading: strings.TrimSpace(strings.TrimPrefix(line, "## ")), start: lineNo, lines: []string{line}}
			continue
		}
		current.lines = append(current.lines, line)
	}
	if len(strings.TrimSpace(strings.Join(current.lines, "\n"))) > 0 {
		sections = append(sections, current)
	}
	chunks := make([]semanticRecallChunk, 0, len(sections))
	docID := semanticRecallStableID("doc", relPath)
	for _, section := range sections {
		content := strings.TrimSpace(strings.Join(section.lines, "\n"))
		if content == "" {
			continue
		}
		lineEnd := section.start + len(section.lines) - 1
		indexText := strings.Join([]string{title, relPath, section.heading, content}, "\n")
		hash := semanticRecallHash(indexText)
		chunks = append(chunks, semanticRecallChunk{
			ChunkID:      semanticRecallStableID("chunk", relPath+"\n"+section.heading+"\n"+hash),
			DocID:        docID,
			Path:         relPath,
			Title:        title,
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

func runSemanticRecallLexicalBaseline(ctx context.Context, client *runclient.Client, queries []semanticRecallQuery) (semanticRecallMethodReport, error) {
	start := time.Now()
	rows := []semanticRecallRow{}
	duplicateHits := 0
	for _, query := range queries {
		search, err := client.Search(ctx, domain.SearchQuery{
			Text:       query.Text,
			PathPrefix: "docs/architecture/",
			Limit:      25,
		})
		if err != nil {
			return semanticRecallMethodReport{}, fmt.Errorf("run lexical baseline for %s: %w", query.ID, err)
		}
		ranked, duplicates := collapseDomainSearchHits(search.Hits)
		duplicateHits += duplicates
		rows = append(rows, semanticRecallRow{
			QueryID:      query.ID,
			Kind:         query.Kind,
			ExpectedPath: query.ExpectedPath,
			Rank:         semanticRecallRankOfPath(ranked, query.ExpectedPath),
			TopCitations: semanticRecallTopCitations(ranked, 3),
			Hit:          semanticRecallRankOfPath(ranked, query.ExpectedPath) > 0 && semanticRecallRankOfPath(ranked, query.ExpectedPath) <= 3,
		})
	}
	hitAt3, mrr := semanticRecallMetrics(rows)
	return semanticRecallMethodReport{
		Method:             "current_lexical_fts",
		Status:             "completed",
		Description:        "Installed OpenClerk SQLite FTS through current Search; no ranking or schema change.",
		HitAt3:             hitAt3,
		MRR:                mrr,
		QueryCount:         len(queries),
		RawDuplicateHits:   duplicateHits,
		TotalSeconds:       roundSeconds(time.Since(start).Seconds()),
		Rows:               rows,
		EvidencePosture:    "citation-bearing current lexical baseline; no vector evidence claimed",
		ValidationBoundary: "uses embedded OpenClerk runtime only; no direct SQLite reads, no raw vault inspection beyond copied eval corpus setup, no default ranking change",
	}, nil
}

func collapseDomainSearchHits(hits []domain.SearchHit) ([]semanticRecallRankedDoc, int) {
	seen := map[string]bool{}
	ranked := []semanticRecallRankedDoc{}
	duplicates := 0
	for _, hit := range hits {
		citation := semanticRecallCitation{}
		if len(hit.Citations) > 0 {
			citation = semanticRecallCitation{
				Path:      hit.Citations[0].Path,
				Heading:   hit.Citations[0].Heading,
				LineStart: hit.Citations[0].LineStart,
				LineEnd:   hit.Citations[0].LineEnd,
			}
		}
		if seen[citation.Path] {
			duplicates++
			continue
		}
		seen[citation.Path] = true
		ranked = append(ranked, semanticRecallRankedDoc{
			Path:     citation.Path,
			Score:    hit.Score,
			Citation: citation,
			ChunkID:  hit.ChunkID,
			RawRank:  hit.Rank,
		})
	}
	return ranked, duplicates
}

func runSemanticRecallTokenFallback(chunks []semanticRecallChunk, queries []semanticRecallQuery, includeAliases bool) semanticRecallMethodReport {
	start := time.Now()
	methodName := "lexical_token_overlap_fallback"
	description := "Eval-only stopword-trimmed token-overlap fallback with title/path/heading weighting."
	if includeAliases {
		methodName = "lexical_alias_overlap_fallback"
		description = "Eval-only token-overlap fallback plus documented domain aliases for each semantic-recall query row."
	}
	rows := []semanticRecallRow{}
	duplicateHits := 0
	for _, query := range queries {
		tokens := semanticRecallTokens(query.Text)
		if includeAliases {
			tokens = append(tokens, semanticRecallTokens(strings.Join(query.Aliases, " "))...)
		}
		hits := scoreSemanticRecallChunks(chunks, tokens)
		ranked, duplicates := collapseSemanticRecallHits(hits)
		duplicateHits += duplicates
		rank := semanticRecallRankOfPath(ranked, query.ExpectedPath)
		rows = append(rows, semanticRecallRow{
			QueryID:      query.ID,
			Kind:         query.Kind,
			ExpectedPath: query.ExpectedPath,
			Rank:         rank,
			TopCitations: semanticRecallTopCitations(ranked, 3),
			Hit:          rank > 0 && rank <= 3,
		})
	}
	hitAt3, mrr := semanticRecallMetrics(rows)
	return semanticRecallMethodReport{
		Method:             methodName,
		Status:             "completed",
		Description:        description,
		HitAt3:             hitAt3,
		MRR:                mrr,
		QueryCount:         len(queries),
		RawDuplicateHits:   duplicateHits,
		TotalSeconds:       roundSeconds(time.Since(start).Seconds()),
		Rows:               rows,
		EvidencePosture:    "eval-only no-vector lexical fallback; does not change production Search",
		ValidationBoundary: "candidate scoring runs inside maintainer harness only; no openclerk retrieval JSON contract change",
		CandidateOnly:      true,
	}
}

func scoreSemanticRecallChunks(chunks []semanticRecallChunk, tokens []string) []semanticRecallHit {
	tokenSet := stringSetFromSlice(tokens)
	hits := []semanticRecallHit{}
	for _, chunk := range chunks {
		title := stringSetFromSlice(semanticRecallTokens(chunk.Title))
		heading := stringSetFromSlice(semanticRecallTokens(chunk.Heading))
		pathTokens := stringSetFromSlice(semanticRecallTokens(chunk.Path))
		content := stringSetFromSlice(semanticRecallTokens(chunk.Content))
		score := 0.0
		for token := range tokenSet {
			if _, ok := title[token]; ok {
				score += 3
			}
			if _, ok := heading[token]; ok {
				score += 2
			}
			if _, ok := pathTokens[token]; ok {
				score += 1.5
			}
			if _, ok := content[token]; ok {
				score += 1
			}
		}
		if score > 0 {
			hits = append(hits, semanticRecallHit{Chunk: chunk, Score: score})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			return hits[i].Chunk.ChunkID < hits[j].Chunk.ChunkID
		}
		return hits[i].Score > hits[j].Score
	})
	return hits
}

func runSemanticRecallLocalHybrid(ctx context.Context, config semanticRecallConfig, chunks []semanticRecallChunk, queries []semanticRecallQuery) ([]semanticRecallMethodReport, semanticRecallOllama) {
	start := time.Now()
	oc := ollamaClient{baseURL: config.OllamaURL, client: &http.Client{Timeout: 30 * time.Second}}
	show, err := oc.show(ctx, config.EmbeddingModel)
	ollama := semanticRecallOllama{
		URL:    config.OllamaURL,
		Model:  config.EmbeddingModel,
		Status: "completed",
	}
	if err != nil {
		ollama.Status = "environment_blocked"
		ollama.ErrorSummary = semanticRecallErrorSummary(err)
		return []semanticRecallMethodReport{
			semanticRecallBlockedMethod("local_vector_only", len(queries), time.Since(start), err),
			semanticRecallBlockedMethod("local_hybrid_rrf", len(queries), time.Since(start), err),
		}, ollama
	}
	ollama.Capabilities = show.Capabilities
	ollama.Details = show.Details
	ollama.ModelInfoSummary = semanticRecallModelInfoSummary(show.ModelInfo)

	chunkInputs := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		chunkInputs = append(chunkInputs, chunk.TextForIndex)
	}
	chunkVectors, err := oc.embed(ctx, config.EmbeddingModel, chunkInputs)
	if err != nil {
		ollama.Status = "environment_blocked"
		ollama.ErrorSummary = semanticRecallErrorSummary(err)
		return []semanticRecallMethodReport{
			semanticRecallBlockedMethod("local_vector_only", len(queries), time.Since(start), err),
			semanticRecallBlockedMethod("local_hybrid_rrf", len(queries), time.Since(start), err),
		}, ollama
	}
	if len(chunkVectors) > 0 {
		ollama.EmbeddingDims = len(chunkVectors[0])
	}
	queryInputs := make([]string, 0, len(queries))
	for _, query := range queries {
		queryInputs = append(queryInputs, query.Text)
	}
	queryVectors, err := oc.embed(ctx, config.EmbeddingModel, queryInputs)
	if err != nil {
		ollama.Status = "environment_blocked"
		ollama.ErrorSummary = semanticRecallErrorSummary(err)
		return []semanticRecallMethodReport{
			semanticRecallBlockedMethod("local_vector_only", len(queries), time.Since(start), err),
			semanticRecallBlockedMethod("local_hybrid_rrf", len(queries), time.Since(start), err),
		}, ollama
	}
	vector := buildVectorMethodReport("local_vector_only", "Ollama local embedding vector-only chunk ranking.", chunks, queries, chunkVectors, queryVectors, nil)
	hybrid := buildVectorMethodReport("local_hybrid_rrf", "RRF fusion over eval current lexical-token score and Ollama local vector chunk ranks.", chunks, queries, chunkVectors, queryVectors, func(query semanticRecallQuery) []semanticRecallHit {
		return scoreSemanticRecallChunks(chunks, semanticRecallTokens(query.Text))
	})
	vector.TotalSeconds = roundSeconds(time.Since(start).Seconds())
	hybrid.TotalSeconds = vector.TotalSeconds
	return []semanticRecallMethodReport{vector, hybrid}, ollama
}

func buildVectorMethodReport(method string, description string, chunks []semanticRecallChunk, queries []semanticRecallQuery, chunkVectors [][]float64, queryVectors [][]float64, lexical func(semanticRecallQuery) []semanticRecallHit) semanticRecallMethodReport {
	rows := []semanticRecallRow{}
	duplicatesTotal := 0
	for idx, query := range queries {
		vectorHits := vectorSemanticRecallHits(chunks, chunkVectors, queryVectors[idx])
		hits := vectorHits
		if lexical != nil {
			hits = rrfSemanticRecallHits(vectorHits, lexical(query))
		}
		ranked, duplicates := collapseSemanticRecallHits(hits)
		duplicatesTotal += duplicates
		rank := semanticRecallRankOfPath(ranked, query.ExpectedPath)
		rows = append(rows, semanticRecallRow{
			QueryID:      query.ID,
			Kind:         query.Kind,
			ExpectedPath: query.ExpectedPath,
			Rank:         rank,
			TopCitations: semanticRecallTopCitations(ranked, 3),
			Hit:          rank > 0 && rank <= 3,
		})
	}
	hitAt3, mrr := semanticRecallMetrics(rows)
	return semanticRecallMethodReport{
		Method:             method,
		Status:             "completed",
		Description:        description,
		HitAt3:             hitAt3,
		MRR:                mrr,
		QueryCount:         len(queries),
		RawDuplicateHits:   duplicatesTotal,
		Rows:               rows,
		EvidencePosture:    "real local/offline embedding evidence if Ollama is local and model metadata is recorded",
		ValidationBoundary: "eval-only in-memory vectors; no durable embedding store, provider call, or production ranking change",
		CandidateOnly:      true,
	}
}

func vectorSemanticRecallHits(chunks []semanticRecallChunk, chunkVectors [][]float64, queryVector []float64) []semanticRecallHit {
	hits := make([]semanticRecallHit, 0, len(chunks))
	for idx, chunk := range chunks {
		if idx >= len(chunkVectors) {
			break
		}
		hits = append(hits, semanticRecallHit{Chunk: chunk, Score: dotProduct(queryVector, chunkVectors[idx])})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			return hits[i].Chunk.ChunkID < hits[j].Chunk.ChunkID
		}
		return hits[i].Score > hits[j].Score
	})
	return hits
}

func rrfSemanticRecallHits(vectorHits []semanticRecallHit, lexicalHits []semanticRecallHit) []semanticRecallHit {
	scores := map[string]semanticRecallHit{}
	for idx, hit := range vectorHits {
		hit.Score = 1.0 / float64(60+idx+1)
		scores[hit.Chunk.ChunkID] = hit
	}
	for idx, hit := range lexicalHits {
		score := 1.0 / float64(60+idx+1)
		existing, ok := scores[hit.Chunk.ChunkID]
		if ok {
			existing.Score += score
			scores[hit.Chunk.ChunkID] = existing
			continue
		}
		hit.Score = score
		scores[hit.Chunk.ChunkID] = hit
	}
	combined := make([]semanticRecallHit, 0, len(scores))
	for _, hit := range scores {
		combined = append(combined, hit)
	}
	sort.SliceStable(combined, func(i, j int) bool {
		if combined[i].Score == combined[j].Score {
			return combined[i].Chunk.ChunkID < combined[j].Chunk.ChunkID
		}
		return combined[i].Score > combined[j].Score
	})
	return combined
}

func semanticRecallBlockedMethod(method string, queryCount int, elapsed time.Duration, err error) semanticRecallMethodReport {
	return semanticRecallMethodReport{
		Method:             method,
		Status:             "environment_blocked",
		Description:        "Local Ollama embedding runtime/model was unavailable; semantic evidence is intentionally not faked.",
		QueryCount:         queryCount,
		TotalSeconds:       roundSeconds(elapsed.Seconds()),
		EvidencePosture:    "environment-blocked; rerun with local Ollama and embedding model to produce vector/hybrid evidence",
		ValidationBoundary: "no provider fallback, no fake vectors, no durable embedding store, no production ranking change; error: " + semanticRecallErrorSummary(err),
		CandidateOnly:      true,
		EnvironmentBlocked: true,
	}
}

func collapseSemanticRecallHits(hits []semanticRecallHit) ([]semanticRecallRankedDoc, int) {
	seen := map[string]bool{}
	ranked := []semanticRecallRankedDoc{}
	duplicates := 0
	for idx, hit := range hits {
		citation := semanticRecallCitation{
			Path:      hit.Chunk.Path,
			Heading:   hit.Chunk.Heading,
			LineStart: hit.Chunk.LineStart,
			LineEnd:   hit.Chunk.LineEnd,
		}
		if seen[hit.Chunk.Path] {
			duplicates++
			continue
		}
		seen[hit.Chunk.Path] = true
		ranked = append(ranked, semanticRecallRankedDoc{
			Path:     hit.Chunk.Path,
			Score:    hit.Score,
			Citation: citation,
			ChunkID:  hit.Chunk.ChunkID,
			RawRank:  idx + 1,
		})
	}
	return ranked, duplicates
}

func semanticRecallMetrics(rows []semanticRecallRow) (int, float64) {
	hitAt3 := 0
	reciprocal := 0.0
	for _, row := range rows {
		if row.Rank > 0 {
			reciprocal += 1.0 / float64(row.Rank)
			if row.Rank <= 3 {
				hitAt3++
			}
		}
	}
	if len(rows) == 0 {
		return 0, 0
	}
	return hitAt3, math.Round((reciprocal/float64(len(rows)))*1000) / 1000
}

func semanticRecallRankOfPath(ranked []semanticRecallRankedDoc, expected string) int {
	for idx, hit := range ranked {
		if hit.Path == expected {
			return idx + 1
		}
	}
	return 0
}

func semanticRecallTopCitations(ranked []semanticRecallRankedDoc, limit int) []semanticRecallCitation {
	citations := []semanticRecallCitation{}
	for idx, hit := range ranked {
		if idx >= limit {
			break
		}
		citations = append(citations, hit.Citation)
	}
	return citations
}

func runSemanticRecallFreshnessProbe(vaultRoot string, original []semanticRecallChunk) semanticRecallFreshnessProbe {
	start := time.Now()
	changedPath := "docs/architecture/hybrid-retrieval-adr.md"
	target := filepath.Join(vaultRoot, filepath.FromSlash(changedPath))
	content, err := os.ReadFile(target)
	if err != nil {
		return semanticRecallFreshnessProbe{Status: "error", ChangedPath: changedPath, EvidencePosture: "could not read copied document", ValidationBoundary: semanticRecallErrorSummary(err)}
	}
	if err := os.WriteFile(target, append(content, []byte("\n\n<!-- semantic recall stale-index probe -->\n")...), 0o644); err != nil {
		return semanticRecallFreshnessProbe{Status: "error", ChangedPath: changedPath, EvidencePosture: "could not update copied document", ValidationBoundary: semanticRecallErrorSummary(err)}
	}
	updatedContent, err := os.ReadFile(target)
	if err != nil {
		return semanticRecallFreshnessProbe{Status: "error", ChangedPath: changedPath, EvidencePosture: "could not reread copied document", ValidationBoundary: semanticRecallErrorSummary(err)}
	}
	updated := semanticRecallChunksForDocument(changedPath, string(updatedContent))
	originalByID := map[string]string{}
	for _, chunk := range original {
		if chunk.Path == changedPath {
			originalByID[chunk.ChunkID] = chunk.Hash
		}
	}
	stale := 0
	for _, chunk := range updated {
		if prior, ok := originalByID[chunk.ChunkID]; !ok || prior != chunk.Hash {
			stale++
		}
	}
	return semanticRecallFreshnessProbe{
		Status:             "completed",
		ChangedPath:        changedPath,
		StaleChunks:        stale,
		RebuiltChunks:      len(updated),
		Seconds:            roundSeconds(time.Since(start).Seconds()),
		EvidencePosture:    "content-hash mismatch on copied <run-root> corpus detects stale local index rows and identifies affected chunks for rebuild",
		ValidationBoundary: "probe mutates only copied eval corpus under <run-root>; no production documents or durable indexes are changed",
	}
}

func semanticRecallOutcomes(report semanticRecallReport) []maturityOutcome {
	vectorBlocked := false
	vectorCompleted := false
	lexicalCompleted := false
	lexicalHitAt3 := 0
	for _, method := range report.Methods {
		if strings.HasPrefix(method.Method, "local_") && method.EnvironmentBlocked {
			vectorBlocked = true
		}
		if strings.HasPrefix(method.Method, "local_") && method.Status == "completed" {
			vectorCompleted = true
		}
		if strings.HasPrefix(method.Method, "lexical_") && method.Status == "completed" {
			lexicalCompleted = true
			if method.HitAt3 > lexicalHitAt3 {
				lexicalHitAt3 = method.HitAt3
			}
		}
	}
	outcomes := []maturityOutcome{}
	if lexicalCompleted {
		capability := "partial"
		details := "lexical fallback produced reduced recall evidence without embeddings; promotion still requires source-sensitive regression review before default ranking changes"
		if lexicalHitAt3 == 0 {
			capability = "fail"
			details = "lexical fallback did not recover expected docs on the semantic-recall pressure set"
		}
		outcomes = append(outcomes, maturityOutcome{
			Name:            "lexical-fallback-eval",
			Status:          "recorded",
			SafetyPass:      "pass",
			CapabilityPass:  capability,
			UXQuality:       "pass_if_invisible_in_search",
			Performance:     "low_cost_eval_only",
			EvidencePosture: "reduced query-row metrics; no vector or provider calls",
			Details:         details,
		})
	}
	if vectorCompleted {
		outcomes = append(outcomes, maturityOutcome{
			Name:            "local-hybrid-poc",
			Status:          "recorded",
			SafetyPass:      "partial",
			CapabilityPass:  "recorded",
			UXQuality:       "pass_if_hidden_behind_search",
			Performance:     "recorded",
			EvidencePosture: "real local Ollama embedding evidence with citations, duplicate counts, and freshness probe",
			Details:         "local/offline hybrid evidence is available for promotion/defer decision",
		})
	} else if vectorBlocked {
		outcomes = append(outcomes, maturityOutcome{
			Name:            "local-hybrid-poc",
			Status:          "environment_blocked",
			SafetyPass:      "pass",
			CapabilityPass:  "not_recorded",
			UXQuality:       "not_recorded",
			Performance:     "not_recorded",
			EvidencePosture: "Ollama local embedding runtime/model unavailable; no fake vectors produced",
			Details:         "rerun with local Ollama and embedding model to satisfy oc-bq8c vector evidence",
		})
	}
	return outcomes
}

func writeSemanticRecallMarkdownReport(path string, rep semanticRecallReport) error {
	var b strings.Builder
	b.WriteString("# OpenClerk Semantic Recall Report\n\n")
	fmt.Fprintf(&b, "- Lane: `%s`\n", rep.Metadata.Lane)
	fmt.Fprintf(&b, "- Mode: `%s`\n", rep.Metadata.Mode)
	fmt.Fprintf(&b, "- Harness: %s\n", rep.Metadata.Harness)
	fmt.Fprintf(&b, "- Run root: `%s`\n", rep.Metadata.RunRootArtifactReference)
	fmt.Fprintf(&b, "- Raw logs committed: `%t`\n", rep.Metadata.RawLogsCommitted)
	fmt.Fprintf(&b, "- Raw content committed: `%t`\n\n", rep.Metadata.RawContentCommitted)

	b.WriteString("## Corpus\n\n")
	b.WriteString("| Metric | Value |\n| --- | ---: |\n")
	fmt.Fprintf(&b, "| documents | %d |\n", rep.Corpus.Documents)
	fmt.Fprintf(&b, "| chunks | %d |\n", rep.Corpus.Chunks)
	fmt.Fprintf(&b, "| query_rows | %d |\n\n", rep.Corpus.QueryRows)
	fmt.Fprintf(&b, "Chunking policy: %s\n\n", rep.Corpus.ChunkingPolicy)
	fmt.Fprintf(&b, "Citation policy: %s\n\n", rep.Corpus.CitationPolicy)

	if rep.Ollama.Status != "" {
		b.WriteString("## Ollama\n\n")
		b.WriteString("| Field | Value |\n| --- | --- |\n")
		fmt.Fprintf(&b, "| url | `%s` |\n", markdownCell(rep.Ollama.URL))
		fmt.Fprintf(&b, "| model | `%s` |\n", markdownCell(rep.Ollama.Model))
		fmt.Fprintf(&b, "| status | `%s` |\n", rep.Ollama.Status)
		fmt.Fprintf(&b, "| embedding_dimensions | %d |\n", rep.Ollama.EmbeddingDims)
		if rep.Ollama.ErrorSummary != "" {
			fmt.Fprintf(&b, "| error_summary | %s |\n", markdownCell(rep.Ollama.ErrorSummary))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Methods\n\n")
	for _, method := range rep.Methods {
		fmt.Fprintf(&b, "### `%s`\n\n", method.Method)
		b.WriteString("| Metric | Value |\n| --- | ---: |\n")
		fmt.Fprintf(&b, "| status | `%s` |\n", method.Status)
		fmt.Fprintf(&b, "| hit_at_3 | %d |\n", method.HitAt3)
		fmt.Fprintf(&b, "| query_count | %d |\n", method.QueryCount)
		fmt.Fprintf(&b, "| mrr | %.3f |\n", method.MRR)
		fmt.Fprintf(&b, "| raw_duplicate_hits | %d |\n", method.RawDuplicateHits)
		fmt.Fprintf(&b, "| total_seconds | %.2f |\n\n", method.TotalSeconds)
		fmt.Fprintf(&b, "Description: %s\n\n", method.Description)
		fmt.Fprintf(&b, "Evidence posture: %s\n\n", method.EvidencePosture)
		fmt.Fprintf(&b, "Validation boundary: %s\n\n", method.ValidationBoundary)
		if len(method.Rows) > 0 {
			b.WriteString("| Query | Kind | Expected | Rank | Hit | Top citations |\n| --- | --- | --- | ---: | --- | --- |\n")
			for _, row := range method.Rows {
				fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | %d | `%t` | %s |\n", row.QueryID, row.Kind, row.ExpectedPath, row.Rank, row.Hit, markdownCell(formatSemanticRecallCitations(row.TopCitations)))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("## Freshness Probe\n\n")
	b.WriteString("| Field | Value |\n| --- | --- |\n")
	fmt.Fprintf(&b, "| status | `%s` |\n", rep.FreshnessProbe.Status)
	fmt.Fprintf(&b, "| changed_path | `%s` |\n", rep.FreshnessProbe.ChangedPath)
	fmt.Fprintf(&b, "| stale_chunks | %d |\n", rep.FreshnessProbe.StaleChunks)
	fmt.Fprintf(&b, "| rebuilt_chunks | %d |\n", rep.FreshnessProbe.RebuiltChunks)
	fmt.Fprintf(&b, "| seconds | %.2f |\n", rep.FreshnessProbe.Seconds)
	fmt.Fprintf(&b, "| evidence_posture | %s |\n", markdownCell(rep.FreshnessProbe.EvidencePosture))
	fmt.Fprintf(&b, "| validation_boundary | %s |\n\n", markdownCell(rep.FreshnessProbe.ValidationBoundary))

	b.WriteString("## Checks\n\n")
	b.WriteString("| Check | Value |\n| --- | --- |\n")
	fmt.Fprintf(&b, "| reduced_report_only | `%t` |\n", rep.Checks.ReducedReportOnly)
	fmt.Fprintf(&b, "| raw_logs_committed | `%t` |\n", rep.Checks.RawLogsCommitted)
	fmt.Fprintf(&b, "| raw_content_committed | `%t` |\n", rep.Checks.RawContentCommitted)
	fmt.Fprintf(&b, "| machine_absolute_artifact_refs | `%t` |\n", rep.Checks.MachineAbsoluteArtifactRefs)
	fmt.Fprintf(&b, "| production_search_default_changed | `%t` |\n", rep.Checks.ProductionSearchDefaultChanged)
	fmt.Fprintf(&b, "| boundary | %s |\n\n", markdownCell(rep.Checks.Boundary))

	b.WriteString("## Outcomes\n\n")
	b.WriteString("| Name | Status | Safety pass | Capability pass | UX quality | Performance | Evidence posture | Details |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, outcome := range rep.Outcomes {
		fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | `%s` | `%s` | `%s` | %s | %s |\n",
			outcome.Name,
			outcome.Status,
			outcome.SafetyPass,
			outcome.CapabilityPass,
			outcome.UXQuality,
			outcome.Performance,
			markdownCell(outcome.EvidencePosture),
			markdownCell(outcome.Details),
		)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write semantic recall Markdown report: %w", err)
	}
	return nil
}

func formatSemanticRecallCitations(citations []semanticRecallCitation) string {
	parts := []string{}
	for _, citation := range citations {
		part := citation.Path
		if citation.Heading != "" {
			part += " / " + citation.Heading
		}
		if citation.LineStart > 0 {
			part += fmt.Sprintf(" lines %d-%d", citation.LineStart, citation.LineEnd)
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, "; ")
}

func (c ollamaClient) show(ctx context.Context, model string) (ollamaShowResponse, error) {
	var result ollamaShowResponse
	err := c.postJSON(ctx, "/api/show", map[string]any{"model": model}, &result)
	return result, err
}

func (c ollamaClient) embed(ctx context.Context, model string, input []string) ([][]float64, error) {
	var result ollamaEmbedResponse
	if err := c.postJSON(ctx, "/api/embed", map[string]any{"model": model, "input": input}, &result); err != nil {
		return nil, err
	}
	if len(result.Embeddings) != len(input) {
		return nil, fmt.Errorf("ollama returned %d embeddings for %d inputs", len(result.Embeddings), len(input))
	}
	return result.Embeddings, nil
}

func (c ollamaClient) postJSON(ctx context.Context, endpoint string, payload any, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("ollama %s returned HTTP %d: %s", endpoint, resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}
	return nil
}

func semanticRecallModelInfoSummary(info map[string]any) map[string]any {
	if len(info) == 0 {
		return nil
	}
	keys := []string{"general.architecture", "general.parameter_count", "embeddinggemma.embedding_length", "nomic-bert.embedding_length", "gemma3.embedding_length"}
	result := map[string]any{}
	for _, key := range keys {
		if value, ok := info[key]; ok {
			result[key] = value
		}
	}
	return result
}

func semanticRecallTokens(text string) []string {
	raw := semanticRecallWordPattern.FindAllString(strings.ToLower(text), -1)
	tokens := []string{}
	for _, token := range raw {
		if _, stop := semanticRecallStopwords()[token]; stop {
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func semanticRecallStopwords() map[string]struct{} {
	return map[string]struct{}{
		"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {}, "before": {}, "between": {}, "by": {}, "for": {}, "from": {}, "in": {}, "into": {}, "is": {}, "it": {}, "no": {}, "not": {}, "of": {}, "on": {}, "or": {}, "should": {}, "that": {}, "the": {}, "then": {}, "through": {}, "to": {}, "use": {}, "when": {}, "where": {}, "with": {}, "without": {},
	}
}

func stringSetFromSlice(values []string) map[string]struct{} {
	result := map[string]struct{}{}
	for _, value := range values {
		if value != "" {
			result[value] = struct{}{}
		}
	}
	return result
}

func dotProduct(a []float64, b []float64) float64 {
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	sum := 0.0
	for idx := 0; idx < limit; idx++ {
		sum += a[idx] * b[idx]
	}
	return sum
}

func semanticRecallHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func semanticRecallStableID(prefix string, value string) string {
	return prefix + "_" + semanticRecallHash(value)[:16]
}

func semanticRecallErrorSummary(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) > 240 {
		msg = msg[:240]
	}
	return msg
}
