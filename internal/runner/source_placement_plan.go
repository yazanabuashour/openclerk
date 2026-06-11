package runner

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"unicode"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func runSourcePlacementPlan(ctx context.Context, client *runclient.Client, input SourceURLInput) (SourcePlacementPlan, error) {
	placement, err := planSourceURLPlacement(ctx, client, sourceURLPlacementInput{
		URL:           input.URL,
		Title:         input.Title,
		SourceType:    input.SourceType,
		PathHint:      input.PathHint,
		AssetPathHint: input.AssetPathHint,
	})
	if err != nil {
		return SourcePlacementPlan{}, err
	}

	validationBoundaries := sourcePlacementValidationBoundaries()
	authorityLimits := sourcePlacementAuthorityLimits()
	plan := SourcePlacementPlan{
		SourceURL:              placement.SourceURL,
		SourceType:             placement.SourceType,
		CandidateSourcePaths:   placement.CandidateSourcePaths,
		CandidateAssetPaths:    placement.CandidateAssetPaths,
		CandidateSynthesisPath: placement.CandidateSynthesisPath,
		ExistingSource:         placement.ExistingSource,
		DuplicateStatus:        placement.DuplicateStatus,
		FetchStatus:            "planned_no_fetch",
		WriteStatus:            "planned_no_write",
		ApprovalBoundary:       "public URL inspection intent is not durable-write approval; approve source fetch/write, synthesis creation, or existing-source update before mutating",
		ValidationBoundaries:   validationBoundaries,
		AuthorityLimits:        authorityLimits,
	}
	plan.AgentHandoff = sourcePlacementHandoff(plan)
	return plan, nil
}

type sourceURLPlacementInput struct {
	URL           string
	Title         string
	SourceType    string
	PathHint      string
	AssetPathHint string
}

type sourceURLPlacement struct {
	SourceURL              string
	SourceType             string
	Slug                   string
	CandidateSourcePaths   []string
	CandidateAssetPaths    []string
	CandidateSynthesisPath string
	ExistingSource         *DocumentSummary
	DuplicateStatus        string
}

func planSourceURLPlacement(ctx context.Context, client *runclient.Client, input sourceURLPlacementInput) (sourceURLPlacement, error) {
	sourceURL := normalizeSourcePlacementURL(input.URL)
	sourceType := sourcePlacementType(SourceURLInput{
		URL:        input.URL,
		SourceType: input.SourceType,
	})
	slug := sourcePlacementSlug(SourceURLInput{
		URL:   input.URL,
		Title: input.Title,
	}, sourceURL)
	existing, err := sourcePlacementExistingSource(ctx, client, sourceURL, input.URL)
	if err != nil {
		return sourceURLPlacement{}, err
	}

	placement := sourceURLPlacement{
		SourceURL:              sourceURL,
		SourceType:             sourceType,
		Slug:                   slug,
		CandidateSourcePaths:   sourcePlacementCandidatePaths(input.PathHint, sourceType, slug),
		CandidateAssetPaths:    sourcePlacementAssetPaths(input.AssetPathHint, sourceType, slug),
		CandidateSynthesisPath: "synthesis/" + slug + ".md",
		ExistingSource:         existing,
		DuplicateStatus:        "no_existing_source_url_found",
	}
	if existing != nil {
		placement.CandidateSynthesisPath = ""
		placement.DuplicateStatus = "existing_source_url_found_no_fetch_no_write"
	}
	return placement, nil
}

func (placement sourceURLPlacement) NextIngestSourceRequest(candidateStatus string) string {
	if candidateStatus != "public_candidate_requires_ingest_source_url_approval" {
		return ""
	}
	if placement.ExistingSource != nil {
		return fmt.Sprintf(`{"action":"ingest_source_url","source":{"url":%q,"mode":"update","source_type":%q}}`, placement.SourceURL, placement.SourceType)
	}
	if placement.SourceType == "pdf" {
		return fmt.Sprintf(`{"action":"ingest_source_url","source":{"url":%q,"path_hint":%q,"asset_path_hint":%q,"source_type":"pdf"}}`, placement.SourceURL, "sources/"+placement.Slug+".md", "assets/sources/"+placement.Slug+".pdf")
	}
	return fmt.Sprintf(`{"action":"ingest_source_url","source":{"url":%q,"path_hint":%q,"source_type":"web"}}`, placement.SourceURL, "sources/web/"+placement.Slug+".md")
}

func normalizeSourcePlacementURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimSpace(raw)
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	return normalizeGitHubPlacementSourceURL(parsed.String())
}

func sourcePlacementType(input SourceURLInput) string {
	if input.SourceType != "" {
		return input.SourceType
	}
	parsed, err := url.Parse(input.URL)
	if err == nil && strings.HasSuffix(strings.ToLower(parsed.Path), ".pdf") {
		return "pdf"
	}
	return "web"
}

func sourcePlacementSlug(input SourceURLInput, sourceURL string) string {
	label := input.Title
	if label == "" {
		label = githubPlacementLabel(input.URL)
	}
	if label == "" {
		label = githubPlacementLabel(sourceURL)
	}
	if label == "" {
		if parsed, err := url.Parse(sourceURL); err == nil {
			base := strings.TrimSuffix(path.Base(parsed.Path), path.Ext(parsed.Path))
			if base == "" || base == "." || base == "/" {
				base = parsed.Host
			}
			label = parsed.Host + " " + base
		}
	}
	if label == "" {
		label = "source"
	}
	return slugifyPlacementLabel(label)
}

func slugifyPlacementLabel(label string) string {
	var builder strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(label) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}
	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		return "source"
	}
	return slug
}

func normalizeGitHubPlacementSourceURL(sourceURL string) string {
	parsed, err := url.Parse(sourceURL)
	if err != nil {
		return sourceURL
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "github.com" {
		return sourceURL
	}
	segments, ok := placementURLPathSegments(parsed)
	if !ok || len(segments) < 2 {
		return sourceURL
	}
	owner := segments[0]
	repo := strings.TrimSuffix(segments[1], ".git")
	if owner == "" || repo == "" {
		return sourceURL
	}
	if len(segments) == 2 {
		return githubPlacementRawContentURL(owner, repo, "HEAD", []string{"README.md"})
	}
	if len(segments) >= 5 && (segments[2] == "blob" || segments[2] == "raw") && isMarkdownPlacementPathSegments(segments[4:]) {
		return githubPlacementRawContentURL(owner, repo, segments[3], segments[4:])
	}
	if len(segments) == 4 && segments[2] == "tree" {
		return githubPlacementRawContentURL(owner, repo, segments[3], []string{"README.md"})
	}
	return sourceURL
}

func githubPlacementLabel(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	host := strings.ToLower(parsed.Hostname())
	segments, ok := placementURLPathSegments(parsed)
	if !ok {
		return ""
	}
	switch host {
	case "github.com":
		if len(segments) < 2 {
			return ""
		}
		owner := segments[0]
		repo := strings.TrimSuffix(segments[1], ".git")
		if len(segments) == 2 || (len(segments) == 4 && segments[2] == "tree") {
			return owner + " " + repo
		}
		if len(segments) >= 5 && (segments[2] == "blob" || segments[2] == "raw") && isMarkdownPlacementPathSegments(segments[4:]) {
			return githubPlacementContentLabel(owner, repo, segments[4:])
		}
	case "raw.githubusercontent.com":
		if len(segments) >= 4 && isMarkdownPlacementPathSegments(segments[3:]) {
			return githubPlacementContentLabel(segments[0], segments[1], segments[3:])
		}
	}
	return ""
}

func githubPlacementContentLabel(owner string, repo string, filePath []string) string {
	label := owner + " " + repo
	base := ""
	if len(filePath) > 0 {
		base = strings.TrimSuffix(path.Base(filePath[len(filePath)-1]), path.Ext(filePath[len(filePath)-1]))
	}
	if base != "" {
		label += " " + base
	}
	return label
}

func placementURLPathSegments(parsed *url.URL) ([]string, bool) {
	trimmed := strings.Trim(parsed.EscapedPath(), "/")
	if trimmed == "" {
		return nil, false
	}
	rawSegments := strings.Split(trimmed, "/")
	segments := make([]string, 0, len(rawSegments))
	for _, rawSegment := range rawSegments {
		segment, err := url.PathUnescape(rawSegment)
		if err != nil || segment == "" {
			return nil, false
		}
		segments = append(segments, segment)
	}
	return segments, true
}

func githubPlacementRawContentURL(owner string, repo string, ref string, filePath []string) string {
	segments := []string{owner, repo, ref}
	segments = append(segments, filePath...)
	escaped := make([]string, 0, len(segments))
	for _, segment := range segments {
		escaped = append(escaped, url.PathEscape(segment))
	}
	return "https://raw.githubusercontent.com/" + strings.Join(escaped, "/")
}

func isMarkdownPlacementPathSegments(segments []string) bool {
	if len(segments) == 0 {
		return false
	}
	switch strings.ToLower(path.Ext(segments[len(segments)-1])) {
	case ".md", ".markdown", ".mdown", ".mkd":
		return true
	default:
		return false
	}
}

func sourcePlacementExistingSource(ctx context.Context, client *runclient.Client, normalizedURL string, rawURL string) (*DocumentSummary, error) {
	candidates := []string{normalizedURL}
	if trimmed := strings.TrimSpace(rawURL); trimmed != "" && trimmed != normalizedURL {
		candidates = append(candidates, trimmed)
	}
	for _, candidate := range candidates {
		list, err := client.ListDocuments(ctx, domain.DocumentListQuery{
			MetadataKey:   "source_url",
			MetadataValue: candidate,
			Limit:         10,
		})
		if err != nil {
			return nil, err
		}
		if len(list.Documents) == 0 {
			continue
		}
		summary := toDocumentSummaries(list.Documents[:1])[0]
		return &summary, nil
	}
	return nil, nil
}

func sourcePlacementCandidatePaths(pathHint string, sourceType string, slug string) []string {
	paths := []string{}
	if pathHint != "" {
		paths = appendUniqueString(paths, pathHint)
	}
	if sourceType == "web" {
		paths = appendUniqueString(paths, "sources/web/"+slug+".md")
	} else {
		paths = appendUniqueString(paths, "sources/"+slug+".md")
	}
	paths = appendUniqueString(paths, "sources/candidates/"+slug+".md")
	return paths
}

func sourcePlacementAssetPaths(assetPathHint string, sourceType string, slug string) []string {
	if sourceType != "pdf" {
		return nil
	}
	paths := []string{}
	if assetPathHint != "" {
		paths = appendUniqueString(paths, assetPathHint)
	}
	paths = appendUniqueString(paths, "assets/sources/"+slug+".pdf")
	paths = appendUniqueString(paths, "assets/sources/candidates/"+slug+".pdf")
	return paths
}

func sourcePlacementValidationBoundaries() string {
	return "read-only ingest_source_url plan mode; no fetch, no durable write, no source document create/update, no synthesis create/update, no browser or HTTP bypass, no direct vault inspection, no direct SQLite, no source-built runner, and no unsupported transport"
}

func sourcePlacementAuthorityLimits() string {
	return "placement plan proposes runner-owned source and synthesis paths only; canonical markdown sources, citations, provenance, and projection freshness remain authority after approved ingestion"
}

func sourcePlacementHandoff(plan SourcePlacementPlan) *AgentHandoff {
	evidence := []string{
		"source_url=" + plan.SourceURL,
		"source_type=" + plan.SourceType,
		"candidate_source_paths=" + strings.Join(plan.CandidateSourcePaths, ", "),
		"duplicate_status=" + plan.DuplicateStatus,
		"fetch_status=" + plan.FetchStatus,
		"write_status=" + plan.WriteStatus,
	}
	if plan.CandidateSynthesisPath != "" {
		evidence = append(evidence, "candidate_synthesis_path="+plan.CandidateSynthesisPath)
	}
	if plan.ExistingSource != nil {
		evidence = append(evidence, "existing_source="+plan.ExistingSource.Path)
	}
	return &AgentHandoff{
		AnswerSummary: fmt.Sprintf(
			"ingest_source_url plan mode inspected %s and proposed %d source path hint(s); %s; no fetch or write occurred",
			plan.SourceURL,
			len(plan.CandidateSourcePaths),
			plan.DuplicateStatus,
		),
		Evidence:                    evidence,
		ValidationBoundaries:        plan.ValidationBoundaries,
		AuthorityLimits:             plan.AuthorityLimits,
		FollowUpPrimitiveInspection: "not required for routine placement answer; after approval call ingest_source_url create/update or compile_synthesis as appropriate",
	}
}
