package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func runDuplicateCandidateReport(ctx context.Context, client *runclient.Client, options DuplicateCandidateOptions) (DuplicateCandidateReport, error) {
	limit := options.Limit
	if limit == 0 {
		limit = 10
	}
	search, err := client.Search(ctx, domain.SearchQuery{
		Text:       options.Query,
		PathPrefix: options.PathPrefix,
		Limit:      limit,
	})
	if err != nil {
		return DuplicateCandidateReport{}, err
	}
	convertedSearch := toSearchResult(search)

	var documents []DocumentSummary
	if options.PathPrefix != "" {
		list, err := client.ListDocuments(ctx, domain.DocumentListQuery{
			PathPrefix: options.PathPrefix,
			Limit:      limit,
		})
		if err != nil {
			return DuplicateCandidateReport{}, err
		}
		documents = toDocumentSummaries(list.Documents)
	}

	likelyTarget, evidenceInspected, err := duplicateCandidateLikelyTarget(ctx, client, convertedSearch, documents, options)
	if err != nil {
		return DuplicateCandidateReport{}, err
	}
	duplicateStatus := "no_runner_visible_duplicate_found"
	if likelyTarget != nil {
		duplicateStatus = "likely_duplicate_found"
	}
	validationBoundaries := duplicateCandidateValidationBoundaries()
	authorityLimits := duplicateCandidateAuthorityLimits()
	report := DuplicateCandidateReport{
		Query:                options.Query,
		PathPrefix:           options.PathPrefix,
		Search:               &convertedSearch,
		Documents:            documents,
		LikelyTarget:         likelyTarget,
		EvidenceInspected:    evidenceInspected,
		DuplicateStatus:      duplicateStatus,
		WriteStatus:          "read_only_no_document_created_or_updated",
		ApprovalBoundary:     "ask whether to update the likely existing target or create a new confirmed path before any durable write",
		ValidationBoundaries: validationBoundaries,
		AuthorityLimits:      authorityLimits,
	}
	report.AgentHandoff = duplicateCandidateHandoff(report)
	return report, nil
}

func duplicateCandidateLikelyTarget(ctx context.Context, client *runclient.Client, search SearchResult, documents []DocumentSummary, options DuplicateCandidateOptions) (*DocumentSummary, []string, error) {
	evidence := []string{"search:" + options.Query}
	if options.PathPrefix != "" {
		evidence = append(evidence, "list_documents:"+options.PathPrefix)
	}
	if len(search.Hits) > 0 {
		hit := search.Hits[0]
		document, err := client.GetDocument(ctx, hit.DocID)
		if err != nil {
			return nil, nil, err
		}
		summary := DocumentSummary{
			DocID:     document.DocID,
			Path:      document.Path,
			Title:     document.Title,
			Metadata:  cloneStringMap(document.Metadata),
			UpdatedAt: document.UpdatedAt,
		}
		evidence = append(evidence, "get_document:"+summary.Path)
		return &summary, evidence, nil
	}
	if len(documents) > 0 {
		summary := documents[0]
		evidence = append(evidence, "candidate_from_list:"+summary.Path)
		return &summary, evidence, nil
	}
	return nil, evidence, nil
}

func duplicateCandidateValidationBoundaries() string {
	return "read-only duplicate_candidate_report; no validate, create_document, append_document, replace_section, ingest_source_url, durable write, direct vault inspection, direct SQLite, broad repo search, source-built runner, HTTP/MCP bypass, or unsupported transport"
}

func duplicateCandidateAuthorityLimits() string {
	return "runner-visible search/list/get evidence identifies likely duplicates but does not choose update versus new path; canonical markdown remains authority"
}

func duplicateCandidateHandoff(report DuplicateCandidateReport) *AgentHandoff {
	evidence := []string{
		"query=" + report.Query,
		"duplicate_status=" + report.DuplicateStatus,
		"write_status=" + report.WriteStatus,
		"evidence_inspected=" + strings.Join(report.EvidenceInspected, ", "),
	}
	targetSummary := "no likely duplicate target found"
	if report.LikelyTarget != nil {
		targetSummary = report.LikelyTarget.Path + " (" + report.LikelyTarget.Title + ")"
		evidence = append(evidence, "likely_target="+report.LikelyTarget.Path)
	}
	return &AgentHandoff{
		AnswerSummary: fmt.Sprintf(
			"duplicate_candidate_report found %s for %q; %s",
			targetSummary,
			report.Query,
			report.WriteStatus,
		),
		Evidence:                    evidence,
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "not required for routine update-versus-new clarification; use primitives only for explicit drill-down or runner rejection repair",
	}
}
