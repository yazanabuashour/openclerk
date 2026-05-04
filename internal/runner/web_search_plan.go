package runner

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	webSearchPlanValidationBoundaries = "read-only harness-supplied web search planning; no live web search provider call, no browser automation, no HTTP fetch, no durable source write, no synthesis write, no direct vault inspection, no direct SQLite, no source-built runner, and no unsupported transport"
	webSearchPlanAuthorityLimits      = "web search titles and snippets are discovery hints only; durable source evidence, citations, provenance, and projection freshness begin only after approved ingest_source_url through the installed runner"
)

func runWebSearchPlan(ctx context.Context, client *runclient.Client, options WebSearchPlanOptions) (WebSearchPlan, error) {
	limit := options.Limit
	if limit == 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}
	if limit > len(options.Results) {
		limit = len(options.Results)
	}
	plan := WebSearchPlan{
		Query:                options.Query,
		Candidates:           make([]WebSearchCandidate, 0, limit),
		FetchStatus:          "planned_no_fetch",
		WriteStatus:          "planned_no_write",
		ApprovalBoundary:     "public search/read intent is not durable-write approval; approve ingest_source_url fetch/write or synthesis creation before mutating",
		ValidationBoundaries: webSearchPlanValidationBoundaries,
		AuthorityLimits:      webSearchPlanAuthorityLimits,
	}
	for index, result := range options.Results[:limit] {
		candidate, err := webSearchCandidate(ctx, client, result, index+1)
		if err != nil {
			return WebSearchPlan{}, err
		}
		plan.Candidates = append(plan.Candidates, candidate)
	}
	plan.AgentHandoff = webSearchPlanHandoff(plan)
	return plan, nil
}

func webSearchCandidate(ctx context.Context, client *runclient.Client, result WebSearchResultInput, rank int) (WebSearchCandidate, error) {
	normalizedURL := normalizeSourcePlacementURL(result.URL)
	sourceType := result.SourceType
	if sourceType == "" {
		sourceType = inferredWebSearchSourceType(result.URL)
	}
	accessStatus := result.AccessStatus
	if accessStatus == "" {
		accessStatus = "public"
	}
	sourceInput := SourceURLInput{
		URL:        result.URL,
		Title:      result.Title,
		SourceType: sourceType,
	}
	slug := sourcePlacementSlug(sourceInput, normalizedURL)
	existing, err := sourcePlacementExistingSource(ctx, client, normalizedURL, result.URL)
	if err != nil {
		return WebSearchCandidate{}, err
	}
	duplicateStatus := "no_existing_source_url_found"
	candidateSynthesisPath := "synthesis/" + slug + ".md"
	if existing != nil {
		duplicateStatus = "existing_source_url_found_no_fetch_no_write"
		candidateSynthesisPath = ""
	}
	candidateStatus := webSearchCandidateStatus(accessStatus)
	candidate := WebSearchCandidate{
		Rank:                    rank,
		URL:                     result.URL,
		NormalizedURL:           normalizedURL,
		Title:                   result.Title,
		Snippet:                 result.Snippet,
		SourceType:              sourceType,
		AccessStatus:            accessStatus,
		CandidateStatus:         candidateStatus,
		DuplicateStatus:         duplicateStatus,
		CandidateSourcePaths:    sourcePlacementCandidatePaths("", sourceType, slug),
		CandidateAssetPaths:     sourcePlacementAssetPaths("", sourceType, slug),
		CandidateSynthesisPath:  candidateSynthesisPath,
		ExistingSource:          existing,
		NextIngestSourceRequest: webSearchNextIngestRequest(normalizedURL, sourceType, slug, existing != nil, candidateStatus),
	}
	return candidate, nil
}

func inferredWebSearchSourceType(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err == nil && strings.HasSuffix(strings.ToLower(parsed.Path), ".pdf") {
		return "pdf"
	}
	return "web"
}

func webSearchCandidateStatus(accessStatus string) string {
	switch accessStatus {
	case "blocked":
		return "blocked_no_fetch"
	case "authenticated", "private":
		return "unsupported_private_or_authenticated_no_fetch"
	case "unknown":
		return "unknown_access_requires_review_no_fetch"
	default:
		return "public_candidate_requires_ingest_source_url_approval"
	}
}

func webSearchNextIngestRequest(sourceURL string, sourceType string, slug string, existing bool, candidateStatus string) string {
	if candidateStatus != "public_candidate_requires_ingest_source_url_approval" {
		return ""
	}
	if existing {
		return fmt.Sprintf(`{"action":"ingest_source_url","source":{"url":%q,"mode":"update","source_type":%q}}`, sourceURL, sourceType)
	}
	if sourceType == "pdf" {
		return fmt.Sprintf(`{"action":"ingest_source_url","source":{"url":%q,"path_hint":%q,"asset_path_hint":%q,"source_type":"pdf"}}`, sourceURL, "sources/"+slug+".md", "assets/sources/"+slug+".pdf")
	}
	return fmt.Sprintf(`{"action":"ingest_source_url","source":{"url":%q,"path_hint":%q,"source_type":"web"}}`, sourceURL, "sources/web/"+slug+".md")
}

func webSearchPlanHandoff(plan WebSearchPlan) *AgentHandoff {
	evidence := []string{
		"query=" + plan.Query,
		fmt.Sprintf("candidate_count=%d", len(plan.Candidates)),
		"fetch_status=" + plan.FetchStatus,
		"write_status=" + plan.WriteStatus,
	}
	if len(plan.Candidates) > 0 {
		top := plan.Candidates[0]
		evidence = append(evidence,
			"top_url="+top.NormalizedURL,
			"top_source_type="+top.SourceType,
			"top_access_status="+top.AccessStatus,
			"top_candidate_status="+top.CandidateStatus,
			"top_duplicate_status="+top.DuplicateStatus,
		)
	}
	followUp := "no public fetch candidate selected; do not call ingest_source_url for blocked, private, authenticated, or unknown-access results"
	if planHasPublicIngestCandidate(plan) {
		followUp = "after approval, call ingest_source_url with the selected public candidate request; use compile_synthesis only after source evidence exists and synthesis placement is approved"
	}
	return &AgentHandoff{
		AnswerSummary: fmt.Sprintf(
			"web_search_plan ranked %d harness-supplied URL candidate(s) for %q; no search provider call, fetch, or write occurred",
			len(plan.Candidates),
			plan.Query,
		),
		Evidence:                    evidence,
		ValidationBoundaries:        plan.ValidationBoundaries,
		AuthorityLimits:             plan.AuthorityLimits,
		FollowUpPrimitiveInspection: followUp,
	}
}

func planHasPublicIngestCandidate(plan WebSearchPlan) bool {
	for _, candidate := range plan.Candidates {
		if candidate.NextIngestSourceRequest != "" {
			return true
		}
	}
	return false
}

func webSearchPlanSummary(plan WebSearchPlan) string {
	return fmt.Sprintf("returned %d web search candidates", len(plan.Candidates))
}
