package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func runEvidenceBundleReport(ctx context.Context, client *runclient.Client, options EvidenceBundleOptions) (EvidenceBundleReport, error) {
	limit := options.Limit
	if limit == 0 {
		limit = 10
	}
	report := EvidenceBundleReport{
		QuerySummary:         evidenceBundleQuerySummary(options),
		ValidationBoundaries: evidenceBundleValidationBoundaries(),
		AuthorityLimits:      evidenceBundleAuthorityLimits(),
	}

	if options.Query != "" {
		search, err := client.Search(ctx, domain.SearchQuery{Text: options.Query, Limit: limit})
		if err != nil {
			return EvidenceBundleReport{}, err
		}
		converted := toSearchResult(search)
		report.Search = &converted
		report.Citations = append(report.Citations, citationsFromSearch(converted)...)

		records, err := client.LookupRecords(ctx, domain.RecordLookupInput{Text: options.Query, Limit: limit})
		if err != nil {
			return EvidenceBundleReport{}, err
		}
		convertedRecords := toRecordLookupResult(records)
		report.Records = &convertedRecords
		report.Citations = append(report.Citations, citationsFromRecords(convertedRecords.Entities)...)

		decisions, err := client.LookupDecisions(ctx, domain.DecisionLookupInput{Text: options.Query, Limit: limit})
		if err != nil {
			return EvidenceBundleReport{}, err
		}
		convertedDecisions := toDecisionLookupResult(decisions)
		report.Decisions = &convertedDecisions
		report.Citations = append(report.Citations, citationsFromDecisions(convertedDecisions.Decisions)...)
	}

	if options.EntityID != "" {
		entity, err := client.GetRecordEntity(ctx, options.EntityID)
		if err != nil {
			return EvidenceBundleReport{}, err
		}
		converted := toRecordEntity(entity)
		report.Entity = &converted
		report.Citations = append(report.Citations, converted.Citations...)
	}

	if options.DecisionID != "" {
		decision, err := client.GetDecisionRecord(ctx, options.DecisionID)
		if err != nil {
			return EvidenceBundleReport{}, err
		}
		converted := toDecisionRecord(decision)
		report.Decision = &converted
		report.Citations = append(report.Citations, converted.Citations...)
	}

	refKind := options.RefKind
	refID := options.RefID
	if refKind == "" && refID == "" {
		switch {
		case options.EntityID != "":
			refKind = "entity"
			refID = options.EntityID
		case options.DecisionID != "":
			refKind = "decision"
			refID = options.DecisionID
		}
	}

	if refKind != "" && refID != "" {
		provenance, err := client.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
			RefKind: refKind,
			RefID:   refID,
			Limit:   limit,
		})
		if err != nil {
			return EvidenceBundleReport{}, err
		}
		converted := toProvenanceEventList(provenance)
		report.Provenance = &converted
	}

	if options.Projection != "" || (refKind != "" && refID != "") {
		projections, err := client.ListProjectionStates(ctx, domain.ProjectionStateQuery{
			Projection: options.Projection,
			RefKind:    refKind,
			RefID:      refID,
			Limit:      limit,
		})
		if err != nil {
			return EvidenceBundleReport{}, err
		}
		converted := toProjectionStateList(projections)
		report.Projections = &converted
	}

	report.Citations = dedupeCitations(report.Citations)
	report.AgentHandoff = evidenceBundleHandoff(report)
	return report, nil
}

func evidenceBundleQuerySummary(options EvidenceBundleOptions) string {
	parts := []string{}
	if options.Query != "" {
		parts = append(parts, fmt.Sprintf("query=%q", options.Query))
	}
	if options.EntityID != "" {
		parts = append(parts, "entity_id="+options.EntityID)
	}
	if options.DecisionID != "" {
		parts = append(parts, "decision_id="+options.DecisionID)
	}
	if options.RefKind != "" && options.RefID != "" {
		parts = append(parts, "ref="+options.RefKind+":"+options.RefID)
	}
	if options.Projection != "" {
		parts = append(parts, "projection="+options.Projection)
	}
	return "evidence bundle for " + strings.Join(parts, ", ")
}

func citationsFromSearch(result SearchResult) []Citation {
	var citations []Citation
	for _, hit := range result.Hits {
		citations = append(citations, hit.Citations...)
	}
	return citations
}

func citationsFromRecords(records []RecordEntity) []Citation {
	var citations []Citation
	for _, record := range records {
		citations = append(citations, record.Citations...)
	}
	return citations
}

func citationsFromDecisions(decisions []DecisionRecord) []Citation {
	var citations []Citation
	for _, decision := range decisions {
		citations = append(citations, decision.Citations...)
	}
	return citations
}

func dedupeCitations(citations []Citation) []Citation {
	result := []Citation{}
	seen := map[string]struct{}{}
	for _, citation := range citations {
		key := citation.DocID + "|" + citation.ChunkID + "|" + citation.Path + "|" + citation.Heading
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, citation)
	}
	return result
}

func evidenceBundleValidationBoundaries() string {
	return "read-only openclerk retrieval report; no writes, no broad repo search, direct vault inspection, direct file edits, direct SQLite, source-built runners, HTTP/MCP bypasses, unsupported transports, hidden authority ranking, or projection mutation"
}

func evidenceBundleAuthorityLimits() string {
	return "canonical markdown, promoted records, decision records, provenance, and projection freshness remain inspectable evidence; the bundle packages evidence but does not create a new authority source"
}

func evidenceBundleHandoff(report EvidenceBundleReport) *AgentHandoff {
	recordCount := evidenceBundleRecordCount(report)
	decisionCount := evidenceBundleDecisionCount(report)
	provenanceCount := 0
	if report.Provenance != nil {
		provenanceCount = len(report.Provenance.Events)
	}
	evidence := []string{
		"query_summary=" + report.QuerySummary,
		fmt.Sprintf("records=%d", recordCount),
		fmt.Sprintf("decisions=%d", decisionCount),
		"citations=" + citationPathSummary(report.Citations),
		fmt.Sprintf("provenance_events=%d", provenanceCount),
		"projection_freshness=" + projectionListFreshnessSummary(report.Projections),
		"read_only=true",
	}
	if report.Entity != nil {
		evidence = append(evidence, "entity_id="+report.Entity.EntityID)
	}
	if report.Decision != nil {
		evidence = append(evidence, "decision_id="+report.Decision.DecisionID)
	}
	return &AgentHandoff{
		AnswerSummary: fmt.Sprintf(
			"evidence_bundle_report returned %s with %d records, %d decisions, %d citations, %d provenance events, and %s",
			report.QuerySummary,
			recordCount,
			decisionCount,
			len(report.Citations),
			provenanceCount,
			projectionListFreshnessSummary(report.Projections),
		),
		Evidence:                    evidence,
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "not required for routine answer; use primitives only for explicit drill-down or runner rejection repair",
	}
}

func evidenceBundleRecordCount(report EvidenceBundleReport) int {
	seen := map[string]struct{}{}
	if report.Records != nil {
		for _, record := range report.Records.Entities {
			seen[record.EntityID] = struct{}{}
		}
	}
	if report.Entity != nil {
		seen[report.Entity.EntityID] = struct{}{}
	}
	delete(seen, "")
	return len(seen)
}

func evidenceBundleDecisionCount(report EvidenceBundleReport) int {
	seen := map[string]struct{}{}
	if report.Decisions != nil {
		for _, decision := range report.Decisions.Decisions {
			seen[decision.DecisionID] = struct{}{}
		}
	}
	if report.Decision != nil {
		seen[report.Decision.DecisionID] = struct{}{}
	}
	delete(seen, "")
	return len(seen)
}
