package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	structuredStoreValidationBoundaries = "read-only report; uses installed OpenClerk retrieval JSON, promoted record projections, and projection freshness only; does not create independent canonical tables, metrics stores, time-series stores, external connectors, direct SQLite reads, direct vault inspection, HTTP/MCP bypasses, source-built runners, durable writes, or hidden authority ranking"
	structuredStoreAuthorityLimits      = "canonical markdown remains authority for structured facts; record, service, and decision projections are schema-backed derived evidence with citations, provenance, and freshness, not independent truth stores"
)

func runStructuredStoreReport(ctx context.Context, client *runclient.Client, options StructuredStoreOptions) (StructuredStoreReport, error) {
	limit := options.Limit
	if limit == 0 {
		limit = 10
	}
	if limit < 1 || limit > 100 {
		return StructuredStoreReport{}, domain.ValidationError("structured_store.limit must be between 1 and 100", map[string]any{"limit": limit})
	}

	domainName := options.Domain
	if domainName == "" {
		domainName = "records"
	}
	projectionName := structuredStoreProjection(domainName)
	evidenceInspected := structuredStoreEvidencePrefix(options, domainName)

	report := StructuredStoreReport{
		Domain:               domainName,
		Query:                options.Query,
		EntityType:           options.EntityType,
		Status:               options.Status,
		Owner:                options.Owner,
		Interface:            options.Interface,
		Scope:                options.Scope,
		CandidateSurfaces:    structuredStoreCandidates(),
		Recommendation:       "promote structured_store_report as the read-only structured-store decision surface; keep canonical facts in markdown-backed records, services, and decisions, and require separate evidence before any independent non-document canonical store",
		SafetyPass:           "passes: report is read-only, local-first, runner-only, and keeps durable writes and independent canonical storage out of scope",
		CapabilityPass:       "passes for current structured evidence: packages schema-backed record, service, or decision projections with projection freshness and candidate-store boundaries",
		UXQuality:            "improves structured-data review ergonomics by replacing records/projection/candidate-policy choreography with one retrieval action and agent_handoff",
		EvidencePosture:      "evidence comes from promoted record projections and projection freshness; no raw database, spreadsheet, metrics, health, finance, inventory, or external-store artifact is treated as authority by this report",
		ValidationBoundaries: structuredStoreValidationBoundaries,
		AuthorityLimits:      structuredStoreAuthorityLimits,
	}

	switch domainName {
	case "records":
		records, err := client.LookupRecords(ctx, domain.RecordLookupInput{
			Text:       options.Query,
			EntityType: options.EntityType,
			Limit:      limit,
		})
		if err != nil {
			return StructuredStoreReport{}, err
		}
		converted := toRecordLookupResult(records)
		report.Records = &converted
		evidenceInspected = append(evidenceInspected, fmt.Sprintf("records:%d", len(converted.Entities)))
	case "services":
		services, err := client.LookupServices(ctx, domain.ServiceLookupInput{
			Text:      options.Query,
			Status:    options.Status,
			Owner:     options.Owner,
			Interface: options.Interface,
			Limit:     limit,
		})
		if err != nil {
			return StructuredStoreReport{}, err
		}
		converted := toServiceLookupResult(services)
		report.Services = &converted
		evidenceInspected = append(evidenceInspected, fmt.Sprintf("services:%d", len(converted.Services)))
	case "decisions":
		decisions, err := client.LookupDecisions(ctx, domain.DecisionLookupInput{
			Text:   options.Query,
			Status: options.Status,
			Scope:  options.Scope,
			Owner:  options.Owner,
			Limit:  limit,
		})
		if err != nil {
			return StructuredStoreReport{}, err
		}
		converted := toDecisionLookupResult(decisions)
		report.Decisions = &converted
		evidenceInspected = append(evidenceInspected, fmt.Sprintf("decisions:%d", len(converted.Decisions)))
	default:
		return StructuredStoreReport{}, domain.ValidationError("structured_store.domain must be records, services, or decisions", map[string]any{"domain": domainName})
	}

	projections, err := client.ListProjectionStates(ctx, domain.ProjectionStateQuery{
		Projection: projectionName,
		Limit:      limit,
	})
	if err != nil {
		return StructuredStoreReport{}, err
	}
	convertedProjections := toProjectionStateList(projections)
	report.Projections = &convertedProjections
	evidenceInspected = append(evidenceInspected, "projection:"+projectionName, fmt.Sprintf("projection_states:%d", len(convertedProjections.Projections)))
	report.EvidenceInspected = evidenceInspected
	report.AgentHandoff = &AgentHandoff{
		AnswerSummary:               report.Recommendation,
		Evidence:                    evidenceInspected,
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "Use records_lookup, services_lookup, decisions_lookup, provenance_events, and projection_states directly only for drill-down after this report; do not infer independent non-document canonical-store authority from this report.",
	}

	return report, nil
}

func structuredStoreProjection(domainName string) string {
	switch domainName {
	case "services":
		return "services"
	case "decisions":
		return "decisions"
	default:
		return "records"
	}
}

func structuredStoreEvidencePrefix(options StructuredStoreOptions, domainName string) []string {
	evidence := []string{"domain:" + domainName}
	if options.Query != "" {
		evidence = append(evidence, "query:"+options.Query)
	}
	if options.EntityType != "" {
		evidence = append(evidence, "entity_type:"+options.EntityType)
	}
	if options.Status != "" {
		evidence = append(evidence, "status:"+options.Status)
	}
	if options.Owner != "" {
		evidence = append(evidence, "owner:"+options.Owner)
	}
	if options.Interface != "" {
		evidence = append(evidence, "interface:"+options.Interface)
	}
	if options.Scope != "" {
		evidence = append(evidence, "scope:"+options.Scope)
	}
	return evidence
}

func structuredStoreFilterRejection(options StructuredStoreOptions) string {
	switch options.Domain {
	case "records":
		if options.Status != "" || options.Owner != "" || options.Interface != "" || options.Scope != "" {
			return "structured_store records domain supports query and entity_type only"
		}
		if strings.TrimSpace(options.Query) == "" {
			return "structured_store.query is required for records domain"
		}
	case "services":
		if options.EntityType != "" || options.Scope != "" {
			return "structured_store services domain supports query, status, owner, and interface only"
		}
		if !structuredStoreHasSupportedFilter(options.Query, options.Status, options.Owner, options.Interface) {
			return "structured_store query or domain filter is required"
		}
	case "decisions":
		if options.EntityType != "" || options.Interface != "" {
			return "structured_store decisions domain supports query, status, owner, and scope only"
		}
		if !structuredStoreHasSupportedFilter(options.Query, options.Status, options.Owner, options.Scope) {
			return "structured_store query or domain filter is required"
		}
	}
	return ""
}

func structuredStoreHasSupportedFilter(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func structuredStoreCandidates() []StructuredStoreCandidate {
	return []StructuredStoreCandidate{
		{
			Surface:    "canonical_markdown_with_promoted_record_projections",
			Status:     "promoted_current",
			Safety:     "passes because canonical markdown owns record identity and facts while projections expose citations, provenance, and freshness",
			Capability: "passes for selective records, services, and decisions; not a general metrics, lab, finance, or inventory database",
			UXQuality:  "acceptable for routine lookup through records/services/decisions actions, but decision work benefits from a packaged report",
			Implementation: []string{
				"existing records_lookup and record_entity",
				"existing services_lookup and service_record",
				"existing decisions_lookup and decision_record",
				"existing projection_states and provenance_events",
			},
		},
		{
			Surface:    "structured_store_report",
			Status:     "promoted_read_only",
			Safety:     "passes because it packages existing runner evidence and declares non-document canonical stores out of scope",
			Capability: "passes for structured-store candidate comparison and current projection evidence packaging",
			UXQuality:  "one retrieval action replaces repeated records/projection/candidate-boundary choreography",
			Implementation: []string{
				"runner JSON action under openclerk retrieval",
				"help and skill action index",
				"unit tests and reduced eval report",
			},
		},
		{
			Surface:    "domain_specific_typed_actions",
			Status:     "selectively_available",
			Safety:     "passes for services and decisions because they remain derived from markdown; unproven for new domains",
			Capability: "strong when a domain has stable schema and repeated lookup pressure",
			UXQuality:  "good for mature domains, but premature typed surfaces create product clutter without repeated evidence",
			Implementation: []string{
				"requires domain-specific schema and projection lifecycle",
				"requires citations, provenance, freshness, duplicate handling, and approval boundaries",
			},
		},
		{
			Surface:    "independent_sqlite_canonical_tables",
			Status:     "not_promoted",
			Safety:     "fails current authority posture unless writes, provenance, freshness, correction, and markdown reconciliation are promoted",
			Capability: "could help dense metrics or time-series later, but requires domain-specific semantics and lifecycle evidence",
			UXQuality:  "would surprise users if hidden tables outranked visible canonical records",
			Implementation: []string{
				"requires write contract and migration design",
				"requires audit trail and correction workflow",
				"requires projection-to-markdown reconciliation policy",
			},
		},
		{
			Surface:    "external_domain_store_connectors",
			Status:     "not_promoted",
			Safety:     "does not fit routine local-first boundaries without approval, sync, privacy, and source-authority design",
			Capability: "useful as future import/reference evidence, not current OpenClerk authority",
			UXQuality:  "adds provider and sync ceremony before the local product surface is proven",
			Implementation: []string{
				"treat as future import/adapter candidate only",
				"do not bypass installed runner",
			},
		},
	}
}
