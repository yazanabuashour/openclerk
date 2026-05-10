package runner

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	sourceDiscoveryValidationBoundaries = "read-only source discovery report; uses installed OpenClerk retrieval search plus runner-visible document listing/get inspection only; no writes, broad repo search, direct vault inspection, direct SQLite, source-built runners, HTTP/MCP bypasses, unsupported transports, or raw private content exposure"
	sourceDiscoveryAuthorityLimits      = "canonical markdown sources remain authority; this report summarizes runner-visible categories and representative refs without creating a new authority source or hidden ranking"
)

func runSourceDiscoveryReport(ctx context.Context, client *runclient.Client, options SourceDiscoveryOptions) (SourceDiscoveryReport, error) {
	limit, err := boundedRunnerLimit(options.Limit, 10, 50, "source_discovery")
	if err != nil {
		return SourceDiscoveryReport{}, err
	}
	query := strings.TrimSpace(options.Query)
	pathPrefix := strings.TrimSpace(options.PathPrefix)

	search, err := client.Search(ctx, domain.SearchQuery{
		Text:       query,
		PathPrefix: pathPrefix,
		Limit:      limit,
	})
	if err != nil {
		return SourceDiscoveryReport{}, err
	}
	convertedSearch := toSearchResult(search)

	sourceDocs, err := client.ListDocuments(ctx, domain.DocumentListQuery{
		PathPrefix: firstNonEmpty(pathPrefix, "sources/"),
		Limit:      limit,
	})
	if err != nil {
		return SourceDiscoveryReport{}, err
	}
	representatives := toDocumentSummaries(sourceDocs.Documents)
	citations := dedupeCitations(citationsFromSearch(convertedSearch))
	categories := sourceDiscoveryCategories(convertedSearch, representatives)
	summary := sourceDiscoverySanitizedSummary(query, categories, len(citations))

	report := SourceDiscoveryReport{
		QueryPresent:         query != "",
		PathPrefix:           sourceDiscoverySafePathPrefix(firstNonEmpty(pathPrefix, "sources/")),
		SearchHitCount:       len(convertedSearch.Hits),
		RepresentativeCount:  len(representatives),
		CitationCount:        len(citations),
		SourceCategories:     categories,
		SanitizedSummary:     summary,
		ValidationBoundaries: sourceDiscoveryValidationBoundaries,
		AuthorityLimits:      sourceDiscoveryAuthorityLimits,
	}
	report.AgentHandoff = &AgentHandoff{
		AnswerSummary: summary,
		Evidence: append([]string{
			"query_present=" + fmt.Sprint(query != ""),
			fmt.Sprintf("search_hits=%d", len(convertedSearch.Hits)),
			fmt.Sprintf("representative_docs=%d", len(representatives)),
			fmt.Sprintf("citations=%d", len(citations)),
			"read_only=true",
		}, sourceDiscoveryCategoryEvidence(categories)...),
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "not required for routine source discovery; use search/list_documents/get_document only for explicit drill-down or runner rejection repair",
	}
	return report, nil
}

func sourceDiscoveryCategories(search SearchResult, docs []DocumentSummary) []SourceDiscoveryCategory {
	byCategory := map[string]*SourceDiscoveryCategory{}
	ensure := func(path string) *SourceDiscoveryCategory {
		category, prefix := sourceDiscoveryCategory(path)
		if existing, ok := byCategory[category]; ok {
			return existing
		}
		created := &SourceDiscoveryCategory{Category: category, PathPrefix: prefix}
		byCategory[category] = created
		return created
	}
	for _, hit := range search.Hits {
		for _, citation := range hit.Citations {
			category := ensure(citation.Path)
			category.SearchHits++
		}
	}
	for _, doc := range docs {
		category := ensure(doc.Path)
		category.ListedDocuments++
	}
	result := make([]SourceDiscoveryCategory, 0, len(byCategory))
	for _, category := range byCategory {
		result = append(result, *category)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SearchHits+result[i].ListedDocuments == result[j].SearchHits+result[j].ListedDocuments {
			return result[i].Category < result[j].Category
		}
		return result[i].SearchHits+result[i].ListedDocuments > result[j].SearchHits+result[j].ListedDocuments
	})
	return result
}

func sourceDiscoveryCategory(path string) (string, string) {
	switch {
	case strings.HasPrefix(path, "sources/"):
		return "canonical_sources", "sources/"
	case strings.HasPrefix(path, "records/decisions/"), strings.Contains(path, "/decision"):
		return "decision_records", "records/decisions/"
	case strings.HasPrefix(path, "records/"):
		return "promoted_records", "records/"
	case strings.HasPrefix(path, "synthesis/"):
		return "derived_synthesis", "synthesis/"
	case strings.HasPrefix(path, "docs/architecture/"):
		return "architecture_decisions", "docs/architecture/"
	default:
		return "other_runner_visible_documents", ""
	}
}

func sourceDiscoverySafePathPrefix(path string) string {
	_, prefix := sourceDiscoveryCategory(path)
	return prefix
}

func sourceDiscoverySanitizedSummary(query string, categories []SourceDiscoveryCategory, citationCount int) string {
	names := make([]string, 0, len(categories))
	for _, category := range categories {
		names = append(names, fmt.Sprintf("%s:%d", category.Category, category.SearchHits+category.ListedDocuments))
	}
	if len(names) == 0 {
		return "source_discovery_report found no representative runner-visible source categories; read-only behavior preserved"
	}
	queryStatus := "query supplied"
	if strings.TrimSpace(query) == "" {
		queryStatus = "no query supplied"
	}
	return fmt.Sprintf("source_discovery_report inspected representative runner-visible sources with %s, %d citations, categories %s, and read-only behavior", queryStatus, citationCount, strings.Join(names, ", "))
}

func sourceDiscoveryCategoryEvidence(categories []SourceDiscoveryCategory) []string {
	evidence := make([]string, 0, len(categories))
	for _, category := range categories {
		evidence = append(evidence, fmt.Sprintf("category:%s search_hits=%d listed_documents=%d", category.Category, category.SearchHits, category.ListedDocuments))
	}
	return evidence
}
