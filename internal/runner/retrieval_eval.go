package runner

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	retrievalEvalSchemaVersion        = "openclerk-retrieval-eval.v1"
	retrievalEvalValidationBoundaries = "explicit opt-in retrieval eval artifact only; local JSONL capture is off by default and stores sanitized query, action, filters, result ids/paths, module/provider status, latency, and timestamp without document bodies, snippets, document writes, raw vault content, direct SQLite reads, background jobs, or default ranking changes"
	retrievalEvalAuthorityLimits      = "retrieval eval rows are regression evidence for ranking and latency changes only; canonical markdown, citations, provenance, projection freshness, and approved runner writes remain authority"
)

func runRetrievalEvalCapture(ctx context.Context, client *runclient.Client, options RetrievalEvalOptions) (RetrievalEvalCaptureReport, error) {
	capturePath, err := resolveRetrievalEvalPath(client.Paths(), options.CapturePath)
	if err != nil {
		return RetrievalEvalCaptureReport{}, err
	}
	started := time.Now()
	capturedAt := started.UTC()
	var refs []RetrievalEvalResultRef
	var provider RetrievalEvalProviderStatus
	query := ""
	var filters RetrievalEvalFilters

	switch options.Action {
	case RetrievalTaskActionSearch:
		query = sanitizeRetrievalEvalQuery(options.Search.Text)
		limit := defaultRunnerLimit(options.Search.Limit, 10)
		search, err := client.Search(ctx, domain.SearchQuery{
			Text:          query,
			PathPrefix:    options.Search.PathPrefix,
			MetadataKey:   options.Search.MetadataKey,
			MetadataValue: options.Search.MetadataValue,
			Tag:           options.Search.Tag,
			Limit:         limit,
		})
		if err != nil {
			return RetrievalEvalCaptureReport{}, err
		}
		refs = evalRefsFromSearchHits(toSearchResult(search).Hits)
		filters = RetrievalEvalFilters{
			PathPrefix:    options.Search.PathPrefix,
			MetadataKey:   options.Search.MetadataKey,
			MetadataValue: options.Search.MetadataValue,
			Tag:           options.Search.Tag,
			Limit:         limit,
		}
		provider = RetrievalEvalProviderStatus{
			Mode:         "search",
			Provider:     "core_lexical",
			Status:       "completed",
			SearchStatus: "lexical_completed",
		}
	case RetrievalTaskActionSemanticSearch:
		query = sanitizeRetrievalEvalQuery(options.SemanticSearch.Query)
		semanticOptions := options.SemanticSearch
		semanticOptions.Query = query
		result, err := runSemanticSearch(ctx, client, semanticOptions)
		if err != nil {
			return RetrievalEvalCaptureReport{}, err
		}
		refs = evalRefsFromSearchHits(result.Hits)
		filters = RetrievalEvalFilters{
			PathPrefix:    options.SemanticSearch.PathPrefix,
			MetadataKey:   options.SemanticSearch.MetadataKey,
			MetadataValue: options.SemanticSearch.MetadataValue,
			Tag:           options.SemanticSearch.Tag,
			Limit:         defaultRunnerLimit(options.SemanticSearch.Limit, 10),
		}
		provider = RetrievalEvalProviderStatus{
			Mode:         "semantic_search",
			Provider:     result.Provider.Provider,
			Status:       result.Provider.Status,
			Model:        result.Provider.Model,
			CacheStatus:  result.Cache.Status,
			SearchStatus: result.SearchStatus,
		}
	default:
		return RetrievalEvalCaptureReport{}, domain.ValidationError("retrieval_eval.action must be search or semantic_search", nil)
	}

	latencyMS := float64(time.Since(started).Microseconds()) / 1000
	row := RetrievalEvalCase{
		SchemaVersion: retrievalEvalSchemaVersion,
		CaseID:        retrievalEvalCaseID(capturedAt, options.Action, query, filters),
		Action:        options.Action,
		Query:         query,
		Filters:       filters,
		Results:       refs,
		Provider:      provider,
		LatencyMS:     latencyMS,
		CapturedAt:    capturedAt,
	}
	if err := appendRetrievalEvalCase(capturePath, row); err != nil {
		return RetrievalEvalCaptureReport{}, err
	}
	report := RetrievalEvalCaptureReport{
		Case:                 row,
		CapturePath:          capturePath,
		WriteStatus:          "local_eval_artifact_appended",
		ValidationBoundaries: retrievalEvalValidationBoundaries,
		AuthorityLimits:      retrievalEvalAuthorityLimits,
	}
	report.AgentHandoff = &AgentHandoff{
		AnswerSummary:               fmt.Sprintf("captured retrieval eval case %s with %d sanitized result refs", row.CaseID, len(row.Results)),
		Evidence:                    []string{"case_id=" + row.CaseID, "action=" + row.Action, fmt.Sprintf("result_refs=%d", len(row.Results)), "write_status=" + report.WriteStatus},
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: `run retrieval_eval_replay against the same capture_path after retrieval changes; use search or semantic_search directly for source-sensitive answers`,
	}
	return report, nil
}

func runRetrievalEvalReplay(ctx context.Context, client *runclient.Client, options RetrievalReplayOptions) (RetrievalEvalReplayReport, error) {
	capturePath, err := resolveRetrievalEvalPath(client.Paths(), options.CapturePath)
	if err != nil {
		return RetrievalEvalReplayReport{}, err
	}
	limit := cappedRunnerLimit(options.Limit, 100, 1000)
	cases, err := readRetrievalEvalCases(capturePath, limit)
	if err != nil {
		return RetrievalEvalReplayReport{}, err
	}
	report := RetrievalEvalReplayReport{
		CapturePath:          capturePath,
		CapturedCases:        len(cases),
		ValidationBoundaries: retrievalEvalValidationBoundaries,
		AuthorityLimits:      retrievalEvalAuthorityLimits,
	}
	for _, captured := range cases {
		replayed, replayLatency, status, err := replayRetrievalEvalCase(ctx, client, captured)
		if err != nil {
			return RetrievalEvalReplayReport{}, err
		}
		comparison := compareRetrievalEvalCase(captured, replayed, replayLatency, status)
		report.Cases = append(report.Cases, comparison)
	}
	report.ComparedCases = len(report.Cases)
	report.AverageJaccard, report.Top1MatchRate, report.AverageCapturedLatencyMS, report.AverageReplayLatencyMS = aggregateRetrievalEvalReplay(report.Cases)
	report.AgentHandoff = &AgentHandoff{
		AnswerSummary: fmt.Sprintf(
			"replayed %d retrieval eval cases; average_jaccard=%.2f top1_match_rate=%.2f",
			report.ComparedCases,
			report.AverageJaccard,
			report.Top1MatchRate,
		),
		Evidence: []string{
			fmt.Sprintf("captured_cases=%d", report.CapturedCases),
			fmt.Sprintf("compared_cases=%d", report.ComparedCases),
			fmt.Sprintf("average_jaccard=%.4f", report.AverageJaccard),
			fmt.Sprintf("top1_match_rate=%.4f", report.Top1MatchRate),
		},
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "inspect cases with low jaccard or top1 mismatch using search, semantic_search, provenance_events, and projection_states before changing defaults",
	}
	return report, nil
}

func resolveRetrievalEvalPath(paths runclient.Paths, raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return filepath.Join(filepath.Dir(paths.DatabasePath), "retrieval-eval-capture.jsonl"), nil
	}
	cleaned := filepath.Clean(strings.TrimSpace(raw))
	if !filepath.IsAbs(cleaned) {
		if cleaned == "." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
			return "", domain.ValidationError("retrieval eval capture_path must stay within the OpenClerk data directory when relative", nil)
		}
		cleaned = filepath.Join(filepath.Dir(paths.DatabasePath), cleaned)
	}
	if insidePath(paths.VaultRoot, cleaned) {
		return "", domain.ValidationError("retrieval eval capture_path must not be inside the vault root", nil)
	}
	return cleaned, nil
}

func insidePath(root string, candidate string) bool {
	if strings.TrimSpace(root) == "" || strings.TrimSpace(candidate) == "" {
		return false
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	absCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absRoot, absCandidate)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..")
}

func appendRetrievalEvalCase(path string, row RetrievalEvalCase) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return domain.InternalError("create retrieval eval directory", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return domain.InternalError("open retrieval eval capture", err)
	}
	defer func() {
		_ = file.Close()
	}()
	encoded, err := json.Marshal(row)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		return domain.InternalError("write retrieval eval capture", err)
	}
	return nil
}

func readRetrievalEvalCases(path string, limit int) ([]RetrievalEvalCase, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ValidationError("retrieval eval capture_path does not exist", map[string]any{"capture_path": path})
		}
		return nil, domain.InternalError("open retrieval eval capture", err)
	}
	defer func() {
		_ = file.Close()
	}()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	cases := []RetrievalEvalCase{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var row RetrievalEvalCase
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, domain.ValidationError("retrieval eval capture row is invalid JSON", nil)
		}
		if row.SchemaVersion != retrievalEvalSchemaVersion {
			return nil, domain.ValidationError("retrieval eval capture row has unsupported schema_version", map[string]any{"schema_version": row.SchemaVersion})
		}
		cases = append(cases, row)
		if len(cases) >= limit {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, domain.InternalError("read retrieval eval capture", err)
	}
	return cases, nil
}

func replayRetrievalEvalCase(ctx context.Context, client *runclient.Client, captured RetrievalEvalCase) ([]RetrievalEvalResultRef, float64, string, error) {
	started := time.Now()
	switch captured.Action {
	case RetrievalTaskActionSearch:
		search, err := client.Search(ctx, domain.SearchQuery{
			Text:          captured.Query,
			PathPrefix:    captured.Filters.PathPrefix,
			MetadataKey:   captured.Filters.MetadataKey,
			MetadataValue: captured.Filters.MetadataValue,
			Tag:           captured.Filters.Tag,
			Limit:         defaultRunnerLimit(captured.Filters.Limit, 10),
		})
		if err != nil {
			return nil, 0, "", err
		}
		return evalRefsFromSearchHits(toSearchResult(search).Hits), float64(time.Since(started).Microseconds()) / 1000, "completed", nil
	case RetrievalTaskActionSemanticSearch:
		options := SemanticSearchOptions{
			Query:          captured.Query,
			PathPrefix:     captured.Filters.PathPrefix,
			MetadataKey:    captured.Filters.MetadataKey,
			MetadataValue:  captured.Filters.MetadataValue,
			Tag:            captured.Filters.Tag,
			Limit:          defaultRunnerLimit(captured.Filters.Limit, 10),
			Provider:       captured.Provider.Provider,
			EmbeddingModel: captured.Provider.Model,
		}
		result, err := runSemanticSearch(ctx, client, options)
		if err != nil {
			return nil, 0, "", err
		}
		return evalRefsFromSearchHits(result.Hits), float64(time.Since(started).Microseconds()) / 1000, result.SearchStatus, nil
	default:
		return nil, 0, "unsupported_action", nil
	}
}

func compareRetrievalEvalCase(captured RetrievalEvalCase, current []RetrievalEvalResultRef, replayLatency float64, status string) RetrievalEvalReplayCase {
	return RetrievalEvalReplayCase{
		CaseID:            captured.CaseID,
		Action:            captured.Action,
		Query:             captured.Query,
		CapturedAt:        captured.CapturedAt,
		CapturedResults:   captured.Results,
		CurrentResults:    current,
		Jaccard:           retrievalEvalJaccard(captured.Results, current),
		Top1Match:         retrievalEvalTop1Match(captured.Results, current),
		CapturedLatencyMS: captured.LatencyMS,
		ReplayLatencyMS:   replayLatency,
		Status:            firstNonEmpty(status, "completed"),
	}
}

func evalRefsFromSearchHits(hits []SearchHit) []RetrievalEvalResultRef {
	refs := make([]RetrievalEvalResultRef, 0, len(hits))
	for _, hit := range hits {
		path := ""
		if len(hit.Citations) > 0 {
			path = hit.Citations[0].Path
		}
		refs = append(refs, RetrievalEvalResultRef{
			Rank:    hit.Rank,
			DocID:   hit.DocID,
			ChunkID: hit.ChunkID,
			Path:    path,
		})
	}
	return refs
}

func retrievalEvalJaccard(left []RetrievalEvalResultRef, right []RetrievalEvalResultRef) float64 {
	leftSet := retrievalEvalRefSet(left)
	rightSet := retrievalEvalRefSet(right)
	union := map[string]struct{}{}
	for key := range leftSet {
		union[key] = struct{}{}
	}
	for key := range rightSet {
		union[key] = struct{}{}
	}
	if len(union) == 0 {
		return 1
	}
	intersection := 0
	for key := range leftSet {
		if _, ok := rightSet[key]; ok {
			intersection++
		}
	}
	return float64(intersection) / float64(len(union))
}

func retrievalEvalTop1Match(left []RetrievalEvalResultRef, right []RetrievalEvalResultRef) bool {
	if len(left) == 0 || len(right) == 0 {
		return len(left) == 0 && len(right) == 0
	}
	return retrievalEvalRefKey(left[0]) == retrievalEvalRefKey(right[0])
}

func retrievalEvalRefSet(refs []RetrievalEvalResultRef) map[string]struct{} {
	set := map[string]struct{}{}
	for _, ref := range refs {
		key := retrievalEvalRefKey(ref)
		if key == "" {
			continue
		}
		set[key] = struct{}{}
	}
	return set
}

func retrievalEvalRefKey(ref RetrievalEvalResultRef) string {
	if ref.DocID == "" && ref.ChunkID == "" && ref.Path == "" {
		return ""
	}
	return ref.DocID + "\x00" + ref.ChunkID + "\x00" + ref.Path
}

func aggregateRetrievalEvalReplay(cases []RetrievalEvalReplayCase) (float64, float64, float64, float64) {
	if len(cases) == 0 {
		return 0, 0, 0, 0
	}
	var jaccardSum, top1Sum, capturedLatencySum, replayLatencySum float64
	for _, row := range cases {
		jaccardSum += row.Jaccard
		if row.Top1Match {
			top1Sum++
		}
		capturedLatencySum += row.CapturedLatencyMS
		replayLatencySum += row.ReplayLatencyMS
	}
	denominator := float64(len(cases))
	return roundMetric(jaccardSum / denominator), roundMetric(top1Sum / denominator), roundMetric(capturedLatencySum / denominator), roundMetric(replayLatencySum / denominator)
}

func roundMetric(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func sanitizeRetrievalEvalQuery(query string) string {
	normalized := strings.Join(strings.Fields(query), " ")
	runes := []rune(normalized)
	if len(runes) <= 240 {
		return normalized
	}
	return string(runes[:240])
}

func retrievalEvalCaseID(capturedAt time.Time, action string, query string, filters RetrievalEvalFilters) string {
	payload := fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%d", action, query, filters.PathPrefix, filters.MetadataKey, filters.MetadataValue, filters.Tag, filters.Limit)
	sum := sha256.Sum256([]byte(payload))
	return capturedAt.Format("20060102T150405.000000000Z") + "-" + hex.EncodeToString(sum[:])[:12]
}
