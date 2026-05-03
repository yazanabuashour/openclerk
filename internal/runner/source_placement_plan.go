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
	sourceURL := normalizeSourcePlacementURL(input.URL)
	sourceType := sourcePlacementType(input)
	slug := sourcePlacementSlug(input, sourceURL)
	existing, err := sourcePlacementExistingSource(ctx, client, sourceURL, input.URL)
	if err != nil {
		return SourcePlacementPlan{}, err
	}

	candidateSourcePaths := sourcePlacementCandidatePaths(input.PathHint, sourceType, slug)
	candidateAssetPaths := sourcePlacementAssetPaths(input.AssetPathHint, sourceType, slug)
	candidateSynthesisPath := "synthesis/" + slug + ".md"
	duplicateStatus := "no_existing_source_url_found"
	if existing != nil {
		duplicateStatus = "existing_source_url_found_no_fetch_no_write"
		candidateSynthesisPath = ""
	}
	validationBoundaries := sourcePlacementValidationBoundaries()
	authorityLimits := sourcePlacementAuthorityLimits()
	plan := SourcePlacementPlan{
		SourceURL:              sourceURL,
		SourceType:             sourceType,
		CandidateSourcePaths:   candidateSourcePaths,
		CandidateAssetPaths:    candidateAssetPaths,
		CandidateSynthesisPath: candidateSynthesisPath,
		ExistingSource:         existing,
		DuplicateStatus:        duplicateStatus,
		FetchStatus:            "planned_no_fetch",
		WriteStatus:            "planned_no_write",
		ApprovalBoundary:       "public URL inspection intent is not durable-write approval; approve source fetch/write, synthesis creation, or existing-source update before mutating",
		ValidationBoundaries:   validationBoundaries,
		AuthorityLimits:        authorityLimits,
	}
	plan.AgentHandoff = sourcePlacementHandoff(plan)
	return plan, nil
}

func normalizeSourcePlacementURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimSpace(raw)
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	return parsed.String()
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
