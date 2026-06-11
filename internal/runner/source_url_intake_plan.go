package runner

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	sourceURLIntakeValidationBoundaries = "read-only ingest_source_url inspect mode; runner-owned public fetch of the supplied URL only, bounded link extraction, no recursive crawl, no durable write, no source document create/update, no synthesis create/update, no browser automation, no direct vault inspection, no direct SQLite, no source-built runner, and no unsupported transport"
	sourceURLIntakeAuthorityLimits      = "inspection preview and discovered links are candidate evidence only; canonical markdown sources, citations, provenance, and projection freshness begin after approved ingest_source_url writes"
)

func runSourceURLIntakePlan(ctx context.Context, client *runclient.Client, input SourceURLInput) (SourceURLIntakePlan, error) {
	inspection, err := client.InspectSourceURL(ctx, domain.SourceURLInput{
		URL:           input.URL,
		PathHint:      input.PathHint,
		AssetPathHint: input.AssetPathHint,
		Title:         input.Title,
		SourceType:    input.SourceType,
	})
	if err != nil {
		return SourceURLIntakePlan{}, err
	}
	primaryPlacement, err := planSourceURLPlacement(ctx, client, sourceURLPlacementInput{
		URL:           inspection.SourceURL,
		Title:         firstNonEmpty(input.Title, inspection.Title),
		SourceType:    inspection.SourceType,
		PathHint:      input.PathHint,
		AssetPathHint: input.AssetPathHint,
	})
	if err != nil {
		return SourceURLIntakePlan{}, err
	}
	primary := sourceURLIntakeCandidateFromPlacement(primaryPlacement, inspection.SourceURL, firstNonEmpty(input.Title, inspection.Title), "", "primary", 1)
	limit := cappedRunnerLimit(input.Limit, 8, 20)
	related, err := sourceURLIntakeRelatedCandidates(ctx, client, inspection, limit)
	if err != nil {
		return SourceURLIntakePlan{}, err
	}
	plan := SourceURLIntakePlan{
		SourceURL:            inspection.SourceURL,
		SourceType:           inspection.SourceType,
		Title:                inspection.Title,
		MIMEType:             inspection.MIMEType,
		SizeBytes:            inspection.SizeBytes,
		SHA256:               inspection.SHA256,
		PageCount:            inspection.PageCount,
		TextPreview:          inspection.TextPreview,
		DiscoveredLinkCount:  len(inspection.Links),
		PrimaryCandidate:     &primary,
		RelatedCandidates:    related,
		FetchStatus:          "inspected_public_source",
		WriteStatus:          "planned_no_write",
		ApprovalBoundary:     "public fetch and link inspection are not durable-write approval; approve returned ingest_source_url requests before creating, updating, or refreshing source notes",
		ValidationBoundaries: sourceURLIntakeValidationBoundaries,
		AuthorityLimits:      sourceURLIntakeAuthorityLimits,
	}
	plan.AgentHandoff = sourceURLIntakeHandoff(plan)
	return plan, nil
}

func sourceURLIntakeRelatedCandidates(ctx context.Context, client *runclient.Client, inspection domain.SourceURLInspection, limit int) ([]SourceURLIntakeCandidate, error) {
	candidates := []SourceURLIntakeCandidate{}
	if limit <= 0 {
		return candidates, nil
	}
	for _, link := range inspection.Links {
		if len(candidates) >= limit {
			break
		}
		if !sourceURLIntakeLikelySourceLink(inspection.SourceURL, link.URL) {
			continue
		}
		parsed, rejection := validateOptionalRunnerHTTPURL(link.URL, "source.link.url")
		if rejection != "" || parsed == nil {
			continue
		}
		placement, err := planSourceURLPlacement(ctx, client, sourceURLPlacementInput{
			URL:   link.URL,
			Title: link.Text,
		})
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, sourceURLIntakeCandidateFromPlacement(placement, link.URL, "", link.Text, "related_link", len(candidates)+1))
	}
	return candidates, nil
}

func sourceURLIntakeCandidateFromPlacement(placement sourceURLPlacement, rawURL string, title string, linkText string, relation string, rank int) SourceURLIntakeCandidate {
	return SourceURLIntakeCandidate{
		Rank:                    rank,
		Relation:                relation,
		URL:                     rawURL,
		NormalizedURL:           placement.SourceURL,
		Title:                   title,
		LinkText:                linkText,
		SourceType:              placement.SourceType,
		CandidateStatus:         "public_candidate_requires_ingest_source_url_approval",
		DuplicateStatus:         placement.DuplicateStatus,
		CandidateSourcePaths:    placement.CandidateSourcePaths,
		CandidateAssetPaths:     placement.CandidateAssetPaths,
		CandidateSynthesisPath:  placement.CandidateSynthesisPath,
		ExistingSource:          placement.ExistingSource,
		NextIngestSourceRequest: placement.NextIngestSourceRequest("public_candidate_requires_ingest_source_url_approval"),
	}
}

func sourceURLIntakeLikelySourceLink(primaryURL string, linkURL string) bool {
	primary, err := url.Parse(primaryURL)
	if err != nil {
		return false
	}
	link, err := url.Parse(linkURL)
	if err != nil || link.Hostname() == "" {
		return false
	}
	if sameGitHubContentRoot(primary, link) {
		return sourceURLIntakeSupportedPath(link.Path)
	}
	if strings.EqualFold(primary.Hostname(), link.Hostname()) {
		return sourceURLIntakeSupportedPath(link.Path)
	}
	return false
}

func sameGitHubContentRoot(primary *url.URL, link *url.URL) bool {
	primaryOwner, primaryRepo, primaryOK := githubContentRoot(primary)
	linkOwner, linkRepo, linkOK := githubContentRoot(link)
	return primaryOK && linkOK &&
		strings.EqualFold(primaryOwner, linkOwner) &&
		strings.EqualFold(primaryRepo, linkRepo)
}

func githubContentRoot(parsed *url.URL) (string, string, bool) {
	segments := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	if len(segments) < 2 {
		return "", "", false
	}
	owner, err := url.PathUnescape(segments[0])
	if err != nil {
		return "", "", false
	}
	repo, err := url.PathUnescape(segments[1])
	if err != nil {
		return "", "", false
	}
	switch strings.ToLower(parsed.Hostname()) {
	case "github.com":
		return owner, strings.TrimSuffix(repo, ".git"), true
	case "raw.githubusercontent.com":
		return owner, repo, true
	default:
		return "", "", false
	}
}

func sourceURLIntakeSupportedPath(rawPath string) bool {
	ext := strings.ToLower(path.Ext(rawPath))
	switch ext {
	case ".md", ".markdown", ".mdown", ".mkd", ".html", ".htm", ".pdf":
		return true
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".ico", ".zip", ".gz", ".tgz", ".tar":
		return false
	}
	base := strings.ToLower(path.Base(rawPath))
	if base == "" || base == "." || base == "/" {
		return false
	}
	lowerPath := strings.ToLower(rawPath)
	for _, marker := range []string{"readme", "docs", "doc", "guide", "setup", "install", "skill", "claude", "manual", "reference", "api"} {
		if strings.Contains(base, marker) || strings.Contains(lowerPath, "/"+marker) {
			return true
		}
	}
	return false
}

func sourceURLIntakeHandoff(plan SourceURLIntakePlan) *AgentHandoff {
	evidence := []string{
		"source_url=" + plan.SourceURL,
		"source_type=" + plan.SourceType,
		"mime_type=" + plan.MIMEType,
		fmt.Sprintf("related_candidate_count=%d", len(plan.RelatedCandidates)),
		"fetch_status=" + plan.FetchStatus,
		"write_status=" + plan.WriteStatus,
	}
	if plan.PrimaryCandidate != nil {
		evidence = append(evidence,
			"primary_duplicate_status="+plan.PrimaryCandidate.DuplicateStatus,
			"primary_candidate_paths="+strings.Join(plan.PrimaryCandidate.CandidateSourcePaths, ", "),
		)
	}
	if len(plan.RelatedCandidates) > 0 {
		evidence = append(evidence, "top_related_url="+plan.RelatedCandidates[0].NormalizedURL)
	}
	return &AgentHandoff{
		AnswerSummary: fmt.Sprintf(
			"ingest_source_url inspect mode fetched %s and proposed one primary source plus %d related source candidate(s); no durable write occurred",
			plan.SourceURL,
			len(plan.RelatedCandidates),
		),
		Evidence:                    evidence,
		ValidationBoundaries:        plan.ValidationBoundaries,
		AuthorityLimits:             plan.AuthorityLimits,
		FollowUpPrimitiveInspection: "review the primary and related next_ingest_source_request values; after approval call ingest_source_url create/update for the selected sources, then compile_synthesis only after source evidence exists",
	}
}

func sourceURLIntakePlanSummary(plan SourceURLIntakePlan) string {
	return fmt.Sprintf("inspected source URL and returned %d related candidates", len(plan.RelatedCandidates))
}
