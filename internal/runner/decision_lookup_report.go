package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	decisionLookupValidationBoundaries = "read-only decision lookup report; uses installed OpenClerk decisions, records, search, provenance, and projection reads only; no writes, broad repo search, direct vault inspection, direct SQLite, source-built runners, HTTP/MCP bypasses, unsupported transports, hidden authority ranking, or projection mutation"
	decisionLookupAuthorityLimits      = "canonical markdown remains authority; decision and record projections are derived evidence with citations, provenance, and freshness, and this report degrades gracefully when decision-like evidence is not a formal decision record"
)

func runDecisionLookupReport(ctx context.Context, client *runclient.Client, options DecisionLookupReportOptions) (DecisionLookupReport, error) {
	limit, err := boundedRunnerLimit(options.Limit, 10, 50, "decision_lookup")
	if err != nil {
		return DecisionLookupReport{}, err
	}
	query := strings.TrimSpace(options.Query)
	decisionID := strings.TrimSpace(options.DecisionID)
	report := DecisionLookupReport{
		Query:                query,
		DecisionID:           decisionID,
		ValidationBoundaries: decisionLookupValidationBoundaries,
		AuthorityLimits:      decisionLookupAuthorityLimits,
	}

	if query != "" || decisionID == "" {
		decisions, err := client.LookupDecisions(ctx, domain.DecisionLookupInput{
			Text:   query,
			Status: strings.TrimSpace(options.Status),
			Scope:  strings.TrimSpace(options.Scope),
			Owner:  strings.TrimSpace(options.Owner),
			Limit:  limit,
		})
		if err != nil {
			return DecisionLookupReport{}, err
		}
		converted := toDecisionLookupResult(decisions)
		report.Decisions = &converted
		report.Citations = append(report.Citations, citationsFromDecisions(converted.Decisions)...)
		if decisionID == "" && len(converted.Decisions) > 0 {
			decisionID = converted.Decisions[0].DecisionID
			report.DecisionID = decisionID
		}
	}

	if decisionID != "" {
		decision, err := client.GetDecisionRecord(ctx, decisionID)
		if err == nil {
			converted := toDecisionRecord(decision)
			report.Decision = &converted
			report.Citations = append(report.Citations, converted.Citations...)
		}
	}

	if query != "" {
		records, err := client.LookupRecords(ctx, domain.RecordLookupInput{Text: query, Limit: limit})
		if err != nil {
			return DecisionLookupReport{}, err
		}
		convertedRecords := toRecordLookupResult(records)
		report.Records = &convertedRecords
		report.Citations = append(report.Citations, citationsFromRecords(convertedRecords.Entities)...)

		search, err := client.Search(ctx, domain.SearchQuery{Text: query, Limit: limit})
		if err != nil {
			return DecisionLookupReport{}, err
		}
		convertedSearch := toSearchResult(search)
		report.Search = &convertedSearch
		report.Citations = append(report.Citations, citationsFromSearch(convertedSearch)...)
	}

	refKind, refID := decisionLookupRef(report)
	if refKind != "" && refID != "" {
		provenance, err := client.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
			RefKind: refKind,
			RefID:   refID,
			Limit:   limit,
		})
		if err != nil {
			return DecisionLookupReport{}, err
		}
		convertedProvenance := toProvenanceEventList(provenance)
		report.Provenance = &convertedProvenance

		projections, err := client.ListProjectionStates(ctx, domain.ProjectionStateQuery{
			Projection: "decisions",
			RefKind:    refKind,
			RefID:      refID,
			Limit:      limit,
		})
		if err != nil {
			return DecisionLookupReport{}, err
		}
		convertedProjections := toProjectionStateList(projections)
		report.Projections = &convertedProjections
	}

	report.Citations = dedupeCitations(report.Citations)
	report.LookupStatus = decisionLookupStatus(report)
	report.AgentHandoff = decisionLookupHandoff(report)
	return report, nil
}

func decisionLookupRef(report DecisionLookupReport) (string, string) {
	if report.Decision != nil && report.Decision.DecisionID != "" {
		return "decision", report.Decision.DecisionID
	}
	return "", ""
}

func decisionLookupStatus(report DecisionLookupReport) string {
	switch {
	case report.Decision != nil:
		return "formal_decision_record_found"
	case report.Decisions != nil && len(report.Decisions.Decisions) > 0:
		return "decision_projection_candidates_found"
	case report.Records != nil && len(report.Records.Entities) > 0:
		return "decision_like_record_evidence_found"
	case report.Search != nil && len(report.Search.Hits) > 0:
		return "decision_like_search_evidence_found"
	default:
		return "no_decision_like_evidence_found"
	}
}

func decisionLookupHandoff(report DecisionLookupReport) *AgentHandoff {
	decisionCount := 0
	if report.Decisions != nil {
		decisionCount = len(report.Decisions.Decisions)
	}
	recordCount := 0
	if report.Records != nil {
		recordCount = len(report.Records.Entities)
	}
	searchHits := 0
	if report.Search != nil {
		searchHits = len(report.Search.Hits)
	}
	provenanceCount := 0
	if report.Provenance != nil {
		provenanceCount = len(report.Provenance.Events)
	}
	evidence := []string{
		"lookup_status=" + report.LookupStatus,
		fmt.Sprintf("decision_candidates=%d", decisionCount),
		fmt.Sprintf("record_candidates=%d", recordCount),
		fmt.Sprintf("search_hits=%d", searchHits),
		fmt.Sprintf("citations=%d", len(report.Citations)),
		fmt.Sprintf("provenance_events=%d", provenanceCount),
		"projection_freshness=" + projectionListFreshnessSummary(report.Projections),
		"read_only=true",
	}
	if report.Decision != nil {
		evidence = append(evidence, "decision_id="+report.Decision.DecisionID)
	}
	return &AgentHandoff{
		AnswerSummary: fmt.Sprintf(
			"decision_lookup_report returned %s with %d formal decision candidates, %d record candidates, %d search hits, %d citations, %d provenance events, %s, read-only behavior, validation boundaries, and authority limits",
			report.LookupStatus,
			decisionCount,
			recordCount,
			searchHits,
			len(report.Citations),
			provenanceCount,
			projectionListFreshnessSummary(report.Projections),
		),
		Evidence:                    evidence,
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "not required for routine decision-like lookup; use decisions_lookup, decision_record, records_lookup, search, provenance_events, and projection_states only for explicit drill-down or runner rejection repair",
	}
}
