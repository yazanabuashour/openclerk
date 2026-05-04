package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func RunRetrievalTask(ctx context.Context, config runclient.Config, request RetrievalTaskRequest) (RetrievalTaskResult, error) {
	normalized, rejection := normalizeRetrievalTaskRequest(request)
	if rejection != "" {
		return RetrievalTaskResult{
			Rejected:        true,
			RejectionReason: rejection,
			Summary:         rejection,
		}, nil
	}

	if normalized.Action == RetrievalTaskActionValidate {
		return RetrievalTaskResult{Summary: "valid"}, nil
	}

	if normalized.Action == RetrievalTaskActionWorkflowGuide {
		report := runWorkflowGuideReport(normalized.WorkflowGuide)
		return RetrievalTaskResult{
			WorkflowGuide: &report,
			Summary:       "returned workflow guide report",
		}, nil
	}

	if isMutatingRetrievalAction(normalized) {
		var result RetrievalTaskResult
		err := runclient.WithWriteLock(ctx, config, func() error {
			client, err := runclient.OpenForWrite(config)
			if err != nil {
				return err
			}
			defer func() {
				_ = client.Close()
			}()
			result, err = runRetrievalTaskWithClient(ctx, client, normalized)
			return err
		})
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		return result, nil
	}

	client, err := runclient.OpenReadOnly(config)
	if err != nil {
		return RetrievalTaskResult{}, err
	}
	defer func() {
		_ = client.Close()
	}()

	return runRetrievalTaskWithClient(ctx, client, normalized)
}

func isMutatingRetrievalAction(normalized normalizedRetrievalTaskRequest) bool {
	return (normalized.Action == RetrievalTaskActionAuditContradictions && normalized.Audit.Mode == "repair_existing") ||
		(normalized.Action == RetrievalTaskActionSourceAuditReport && normalized.SourceAudit.Mode == "repair_existing")
}

func runRetrievalTaskWithClient(ctx context.Context, client *runclient.Client, normalized normalizedRetrievalTaskRequest) (RetrievalTaskResult, error) {
	switch normalized.Action {
	case RetrievalTaskActionSearch:
		search, err := client.Search(ctx, domain.SearchQuery{
			Text:          normalized.Search.Text,
			PathPrefix:    normalized.Search.PathPrefix,
			MetadataKey:   normalized.Search.MetadataKey,
			MetadataValue: normalized.Search.MetadataValue,
			Tag:           normalized.Search.Tag,
			Limit:         normalized.Search.Limit,
			Cursor:        normalized.Search.Cursor,
		})
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toSearchResult(search)
		return RetrievalTaskResult{
			Search:  &converted,
			Summary: fmt.Sprintf("returned %d search hits", len(converted.Hits)),
		}, nil
	case RetrievalTaskActionDocumentLinks:
		links, err := client.GetDocumentLinks(ctx, normalized.DocID)
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toDocumentLinksResult(links)
		return RetrievalTaskResult{
			Links:   &converted,
			Summary: fmt.Sprintf("returned links for document %s", normalized.DocID),
		}, nil
	case RetrievalTaskActionGraph:
		graph, err := client.GraphNeighborhood(ctx, domain.GraphNeighborhoodInput{
			DocID:   normalized.DocID,
			ChunkID: normalized.ChunkID,
			NodeID:  normalized.NodeID,
			Limit:   normalized.Limit,
		})
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toGraphNeighborhood(graph)
		return RetrievalTaskResult{
			Graph:   &converted,
			Summary: fmt.Sprintf("returned %d graph nodes and %d edges", len(converted.Nodes), len(converted.Edges)),
		}, nil
	case RetrievalTaskActionRecordsLookup:
		records, err := client.LookupRecords(ctx, domain.RecordLookupInput{
			Text:       normalized.Records.Text,
			EntityType: normalized.Records.EntityType,
			Limit:      normalized.Records.Limit,
			Cursor:     normalized.Records.Cursor,
		})
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toRecordLookupResult(records)
		return RetrievalTaskResult{
			Records: &converted,
			Summary: fmt.Sprintf("returned %d record entities", len(converted.Entities)),
		}, nil
	case RetrievalTaskActionRecordEntity:
		entity, err := client.GetRecordEntity(ctx, normalized.EntityID)
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toRecordEntity(entity)
		return RetrievalTaskResult{
			Entity:  &converted,
			Summary: fmt.Sprintf("returned record entity %s", converted.EntityID),
		}, nil
	case RetrievalTaskActionServicesLookup:
		services, err := client.LookupServices(ctx, domain.ServiceLookupInput{
			Text:      normalized.Services.Text,
			Status:    normalized.Services.Status,
			Owner:     normalized.Services.Owner,
			Interface: normalized.Services.Interface,
			Limit:     normalized.Services.Limit,
			Cursor:    normalized.Services.Cursor,
		})
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toServiceLookupResult(services)
		return RetrievalTaskResult{
			Services: &converted,
			Summary:  fmt.Sprintf("returned %d services", len(converted.Services)),
		}, nil
	case RetrievalTaskActionServiceRecord:
		service, err := client.GetServiceRecord(ctx, normalized.ServiceID)
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toServiceRecord(service)
		return RetrievalTaskResult{
			Service: &converted,
			Summary: fmt.Sprintf("returned service %s", converted.ServiceID),
		}, nil
	case RetrievalTaskActionDecisionsLookup:
		decisions, err := client.LookupDecisions(ctx, domain.DecisionLookupInput{
			Text:   normalized.Decisions.Text,
			Status: normalized.Decisions.Status,
			Scope:  normalized.Decisions.Scope,
			Owner:  normalized.Decisions.Owner,
			Limit:  normalized.Decisions.Limit,
			Cursor: normalized.Decisions.Cursor,
		})
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toDecisionLookupResult(decisions)
		return RetrievalTaskResult{
			Decisions: &converted,
			Summary:   fmt.Sprintf("returned %d decisions", len(converted.Decisions)),
		}, nil
	case RetrievalTaskActionDecisionRecord:
		decision, err := client.GetDecisionRecord(ctx, normalized.DecisionID)
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toDecisionRecord(decision)
		return RetrievalTaskResult{
			Decision: &converted,
			Summary:  fmt.Sprintf("returned decision %s", converted.DecisionID),
		}, nil
	case RetrievalTaskActionProvenanceEvents:
		events, err := client.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
			RefKind:   normalized.Provenance.RefKind,
			RefID:     normalized.Provenance.RefID,
			SourceRef: normalized.Provenance.SourceRef,
			Limit:     normalized.Provenance.Limit,
			Cursor:    normalized.Provenance.Cursor,
		})
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toProvenanceEventList(events)
		return RetrievalTaskResult{
			Provenance: &converted,
			Summary:    fmt.Sprintf("returned %d provenance events", len(converted.Events)),
		}, nil
	case RetrievalTaskActionProjectionStates:
		projections, err := client.ListProjectionStates(ctx, domain.ProjectionStateQuery{
			Projection: normalized.Projection.Projection,
			RefKind:    normalized.Projection.RefKind,
			RefID:      normalized.Projection.RefID,
			Limit:      normalized.Projection.Limit,
			Cursor:     normalized.Projection.Cursor,
		})
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toProjectionStateList(projections)
		return RetrievalTaskResult{
			Projections: &converted,
			Summary:     fmt.Sprintf("returned %d projection states", len(converted.Projections)),
		}, nil
	case RetrievalTaskActionAuditContradictions:
		audit, err := runAuditContradictions(ctx, client, normalized.Audit)
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		return RetrievalTaskResult{
			Audit:   &audit,
			Summary: auditContradictionsSummary(audit),
		}, nil
	case RetrievalTaskActionMemoryRouterRecall:
		report, err := runMemoryRouterRecallReport(ctx, client, normalized.MemoryRouterRecall)
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		return RetrievalTaskResult{
			MemoryRouterRecall: &report,
			Summary:            "returned memory/router recall report",
		}, nil
	case RetrievalTaskActionSourceAuditReport:
		report, err := runSourceAuditReport(ctx, client, normalized.SourceAudit)
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		return RetrievalTaskResult{
			SourceAudit: &report,
			Summary:     sourceAuditReportSummary(report),
		}, nil
	case RetrievalTaskActionEvidenceBundle:
		report, err := runEvidenceBundleReport(ctx, client, normalized.EvidenceBundle)
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		return RetrievalTaskResult{
			EvidenceBundle: &report,
			Summary:        "returned evidence bundle report",
		}, nil
	case RetrievalTaskActionDuplicateCandidate:
		report, err := runDuplicateCandidateReport(ctx, client, normalized.DuplicateCandidate)
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		return RetrievalTaskResult{
			DuplicateCandidate: &report,
			Summary:            "returned duplicate candidate report",
		}, nil
	case RetrievalTaskActionWorkflowGuide:
		report := runWorkflowGuideReport(normalized.WorkflowGuide)
		return RetrievalTaskResult{
			WorkflowGuide: &report,
			Summary:       "returned workflow guide report",
		}, nil
	case RetrievalTaskActionStructuredStore:
		report, err := runStructuredStoreReport(ctx, client, normalized.StructuredStore)
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		return RetrievalTaskResult{
			StructuredStore: &report,
			Summary:         "returned structured store report",
		}, nil
	case RetrievalTaskActionHybridRetrieval:
		report, err := runHybridRetrievalReport(ctx, client, normalized.HybridRetrieval)
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		return RetrievalTaskResult{
			HybridRetrieval: &report,
			Summary:         "returned hybrid retrieval report",
		}, nil
	default:
		return RetrievalTaskResult{}, fmt.Errorf("unsupported retrieval task action %q", normalized.Action)
	}
}

type normalizedRetrievalTaskRequest struct {
	Action             string
	Search             SearchOptions
	DocID              string
	ChunkID            string
	NodeID             string
	EntityID           string
	ServiceID          string
	DecisionID         string
	Records            RecordLookupOptions
	Services           ServiceLookupOptions
	Decisions          DecisionLookupOptions
	Provenance         ProvenanceEventOptions
	Projection         ProjectionStateOptions
	Audit              AuditContradictionsOptions
	MemoryRouterRecall MemoryRouterRecallOptions
	SourceAudit        SourceAuditReportOptions
	EvidenceBundle     EvidenceBundleOptions
	DuplicateCandidate DuplicateCandidateOptions
	WorkflowGuide      WorkflowGuideOptions
	StructuredStore    StructuredStoreOptions
	HybridRetrieval    HybridRetrievalOptions
	Limit              int
}

func normalizeRetrievalTaskRequest(request RetrievalTaskRequest) (normalizedRetrievalTaskRequest, string) {
	action := strings.TrimSpace(request.Action)
	if action == "" {
		action = RetrievalTaskActionValidate
	}
	normalized := normalizedRetrievalTaskRequest{
		Action:             action,
		Search:             request.Search,
		DocID:              strings.TrimSpace(request.DocID),
		ChunkID:            strings.TrimSpace(request.ChunkID),
		NodeID:             strings.TrimSpace(request.NodeID),
		EntityID:           strings.TrimSpace(request.EntityID),
		ServiceID:          strings.TrimSpace(request.ServiceID),
		DecisionID:         strings.TrimSpace(request.DecisionID),
		Records:            request.Records,
		Services:           request.Services,
		Decisions:          request.Decisions,
		Provenance:         request.Provenance,
		Projection:         request.Projection,
		Audit:              request.Audit,
		MemoryRouterRecall: request.MemoryRouterRecall,
		SourceAudit:        request.SourceAudit,
		EvidenceBundle:     request.EvidenceBundle,
		DuplicateCandidate: request.DuplicateCandidate,
		WorkflowGuide:      request.WorkflowGuide,
		StructuredStore:    request.StructuredStore,
		HybridRetrieval:    request.HybridRetrieval,
		Limit:              request.Limit,
	}

	if request.Limit < 0 ||
		request.Search.Limit < 0 ||
		request.Records.Limit < 0 ||
		request.Services.Limit < 0 ||
		request.Decisions.Limit < 0 ||
		request.Provenance.Limit < 0 ||
		request.Projection.Limit < 0 ||
		request.Audit.Limit < 0 ||
		request.MemoryRouterRecall.Limit < 0 ||
		request.SourceAudit.Limit < 0 ||
		request.EvidenceBundle.Limit < 0 ||
		request.DuplicateCandidate.Limit < 0 ||
		request.StructuredStore.Limit < 0 ||
		request.HybridRetrieval.Limit < 0 {
		return normalizedRetrievalTaskRequest{}, "limit must be greater than or equal to 0"
	}

	switch action {
	case RetrievalTaskActionValidate:
		return normalized, ""
	case RetrievalTaskActionSearch:
		if strings.TrimSpace(request.Search.Text) == "" {
			return normalizedRetrievalTaskRequest{}, "search.text is required"
		}
		if rejection := normalizeSearchTagFilter(&normalized.Search); rejection != "" {
			return normalizedRetrievalTaskRequest{}, rejection
		}
		return normalized, ""
	case RetrievalTaskActionDocumentLinks:
		if normalized.DocID == "" {
			return normalizedRetrievalTaskRequest{}, "doc_id is required"
		}
		return normalized, ""
	case RetrievalTaskActionGraph:
		if normalized.DocID == "" && normalized.ChunkID == "" && normalized.NodeID == "" {
			return normalizedRetrievalTaskRequest{}, "doc_id, chunk_id, or node_id is required"
		}
		return normalized, ""
	case RetrievalTaskActionRecordsLookup:
		if strings.TrimSpace(request.Records.Text) == "" {
			return normalizedRetrievalTaskRequest{}, "records.text is required"
		}
		return normalized, ""
	case RetrievalTaskActionRecordEntity:
		if normalized.EntityID == "" {
			return normalizedRetrievalTaskRequest{}, "entity_id is required"
		}
		return normalized, ""
	case RetrievalTaskActionServicesLookup:
		return normalized, ""
	case RetrievalTaskActionServiceRecord:
		if normalized.ServiceID == "" {
			return normalizedRetrievalTaskRequest{}, "service_id is required"
		}
		return normalized, ""
	case RetrievalTaskActionDecisionsLookup:
		return normalized, ""
	case RetrievalTaskActionDecisionRecord:
		if normalized.DecisionID == "" {
			return normalizedRetrievalTaskRequest{}, "decision_id is required"
		}
		return normalized, ""
	case RetrievalTaskActionProvenanceEvents, RetrievalTaskActionProjectionStates:
		return normalized, ""
	case RetrievalTaskActionAuditContradictions:
		normalized.Audit.Query = strings.TrimSpace(request.Audit.Query)
		normalized.Audit.TargetPath = strings.TrimSpace(request.Audit.TargetPath)
		normalized.Audit.Mode = strings.TrimSpace(request.Audit.Mode)
		normalized.Audit.ConflictQuery = strings.TrimSpace(request.Audit.ConflictQuery)
		if normalized.Audit.Mode == "" {
			normalized.Audit.Mode = "plan_only"
		}
		if normalized.Audit.Query == "" {
			return normalizedRetrievalTaskRequest{}, "audit.query is required"
		}
		if normalized.Audit.TargetPath == "" {
			return normalizedRetrievalTaskRequest{}, "audit.target_path is required"
		}
		if normalized.Audit.Mode != "plan_only" && normalized.Audit.Mode != "repair_existing" {
			return normalizedRetrievalTaskRequest{}, "audit.mode must be plan_only or repair_existing"
		}
		return normalized, ""
	case RetrievalTaskActionMemoryRouterRecall:
		normalized.MemoryRouterRecall.Query = strings.TrimSpace(request.MemoryRouterRecall.Query)
		return normalized, ""
	case RetrievalTaskActionSourceAuditReport:
		normalized.SourceAudit.Query = strings.TrimSpace(request.SourceAudit.Query)
		normalized.SourceAudit.TargetPath = strings.TrimSpace(request.SourceAudit.TargetPath)
		normalized.SourceAudit.Mode = strings.TrimSpace(request.SourceAudit.Mode)
		normalized.SourceAudit.ConflictQuery = strings.TrimSpace(request.SourceAudit.ConflictQuery)
		if normalized.SourceAudit.Mode == "" {
			normalized.SourceAudit.Mode = "explain"
		}
		if normalized.SourceAudit.Query == "" {
			return normalizedRetrievalTaskRequest{}, "source_audit.query is required"
		}
		if normalized.SourceAudit.TargetPath == "" {
			return normalizedRetrievalTaskRequest{}, "source_audit.target_path is required"
		}
		if normalized.SourceAudit.Mode != "explain" && normalized.SourceAudit.Mode != "repair_existing" {
			return normalizedRetrievalTaskRequest{}, "source_audit.mode must be explain or repair_existing"
		}
		return normalized, ""
	case RetrievalTaskActionEvidenceBundle:
		normalized.EvidenceBundle.Query = strings.TrimSpace(request.EvidenceBundle.Query)
		normalized.EvidenceBundle.EntityID = strings.TrimSpace(request.EvidenceBundle.EntityID)
		normalized.EvidenceBundle.DecisionID = strings.TrimSpace(request.EvidenceBundle.DecisionID)
		normalized.EvidenceBundle.RefKind = strings.TrimSpace(request.EvidenceBundle.RefKind)
		normalized.EvidenceBundle.RefID = strings.TrimSpace(request.EvidenceBundle.RefID)
		normalized.EvidenceBundle.Projection = strings.TrimSpace(request.EvidenceBundle.Projection)
		if normalized.EvidenceBundle.Query == "" &&
			normalized.EvidenceBundle.EntityID == "" &&
			normalized.EvidenceBundle.DecisionID == "" &&
			(normalized.EvidenceBundle.RefKind == "" || normalized.EvidenceBundle.RefID == "") &&
			normalized.EvidenceBundle.Projection == "" {
			return normalizedRetrievalTaskRequest{}, "evidence_bundle query, entity_id, decision_id, ref_kind/ref_id, or projection is required"
		}
		if (normalized.EvidenceBundle.RefKind == "") != (normalized.EvidenceBundle.RefID == "") {
			return normalizedRetrievalTaskRequest{}, "evidence_bundle.ref_kind and evidence_bundle.ref_id must be provided together"
		}
		return normalized, ""
	case RetrievalTaskActionDuplicateCandidate:
		normalized.DuplicateCandidate.Query = strings.TrimSpace(request.DuplicateCandidate.Query)
		normalized.DuplicateCandidate.PathPrefix = strings.TrimSpace(request.DuplicateCandidate.PathPrefix)
		if normalized.DuplicateCandidate.Query == "" {
			return normalizedRetrievalTaskRequest{}, "duplicate_candidate.query is required"
		}
		return normalized, ""
	case RetrievalTaskActionWorkflowGuide:
		normalized.WorkflowGuide.Intent = strings.TrimSpace(request.WorkflowGuide.Intent)
		if normalized.WorkflowGuide.Intent == "" {
			return normalizedRetrievalTaskRequest{}, "workflow_guide.intent is required"
		}
		return normalized, ""
	case RetrievalTaskActionStructuredStore:
		normalized.StructuredStore.Domain = strings.TrimSpace(request.StructuredStore.Domain)
		normalized.StructuredStore.Query = strings.TrimSpace(request.StructuredStore.Query)
		normalized.StructuredStore.EntityType = strings.TrimSpace(request.StructuredStore.EntityType)
		normalized.StructuredStore.Status = strings.TrimSpace(request.StructuredStore.Status)
		normalized.StructuredStore.Owner = strings.TrimSpace(request.StructuredStore.Owner)
		normalized.StructuredStore.Interface = strings.TrimSpace(request.StructuredStore.Interface)
		normalized.StructuredStore.Scope = strings.TrimSpace(request.StructuredStore.Scope)
		if normalized.StructuredStore.Domain == "" {
			normalized.StructuredStore.Domain = "records"
		}
		if normalized.StructuredStore.Domain != "records" &&
			normalized.StructuredStore.Domain != "services" &&
			normalized.StructuredStore.Domain != "decisions" {
			return normalizedRetrievalTaskRequest{}, "structured_store.domain must be records, services, or decisions"
		}
		if rejection := structuredStoreFilterRejection(normalized.StructuredStore); rejection != "" {
			return normalizedRetrievalTaskRequest{}, rejection
		}
		return normalized, ""
	case RetrievalTaskActionHybridRetrieval:
		normalized.HybridRetrieval.Query = strings.TrimSpace(request.HybridRetrieval.Query)
		normalized.HybridRetrieval.PathPrefix = strings.TrimSpace(request.HybridRetrieval.PathPrefix)
		if normalized.HybridRetrieval.Query == "" {
			return normalizedRetrievalTaskRequest{}, "hybrid_retrieval.query is required"
		}
		return normalized, ""
	default:
		return normalizedRetrievalTaskRequest{}, fmt.Sprintf("unsupported retrieval task action %q", action)
	}
}

func normalizeSearchTagFilter(search *SearchOptions) string {
	return normalizeTagFilter("search", search.Tag, search.tagProvided, &search.MetadataKey, &search.MetadataValue, &search.Tag)
}

func normalizeTagFilter(fieldPrefix string, tag string, tagProvided bool, metadataKey *string, metadataValue *string, normalizedTag *string) string {
	trimmedTag := strings.TrimSpace(tag)
	if trimmedTag == "" && (tagProvided || tag != "") {
		return fieldPrefix + ".tag must be non-empty"
	}
	if trimmedTag == "" {
		*normalizedTag = ""
		return ""
	}
	if strings.TrimSpace(*metadataKey) != "" || strings.TrimSpace(*metadataValue) != "" {
		return fieldPrefix + ".tag cannot be combined with metadata_key or metadata_value"
	}
	*metadataKey = "tag"
	*metadataValue = trimmedTag
	*normalizedTag = ""
	return ""
}
