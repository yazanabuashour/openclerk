package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/client/local"
)

func RunRetrievalTask(ctx context.Context, config local.Config, request RetrievalTaskRequest) (RetrievalTaskResult, error) {
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

	client, err := local.OpenClient(config)
	if err != nil {
		return RetrievalTaskResult{}, err
	}
	defer func() {
		_ = client.Close()
	}()

	switch normalized.Action {
	case RetrievalTaskActionSearch:
		search, err := client.Search(ctx, local.SearchOptions(normalized.Search))
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
		graph, err := client.GraphNeighborhood(ctx, local.GraphNeighborhoodOptions{
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
		records, err := client.LookupRecords(ctx, local.RecordLookupOptions(normalized.Records))
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
		services, err := client.LookupServices(ctx, local.ServiceLookupOptions(normalized.Services))
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
	case RetrievalTaskActionProvenanceEvents:
		events, err := client.ListProvenanceEvents(ctx, local.ProvenanceEventOptions(normalized.Provenance))
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toProvenanceEventList(events)
		return RetrievalTaskResult{
			Provenance: &converted,
			Summary:    fmt.Sprintf("returned %d provenance events", len(converted.Events)),
		}, nil
	case RetrievalTaskActionProjectionStates:
		projections, err := client.ListProjectionStates(ctx, local.ProjectionStateOptions(normalized.Projection))
		if err != nil {
			return RetrievalTaskResult{}, err
		}
		converted := toProjectionStateList(projections)
		return RetrievalTaskResult{
			Projections: &converted,
			Summary:     fmt.Sprintf("returned %d projection states", len(converted.Projections)),
		}, nil
	default:
		return RetrievalTaskResult{}, fmt.Errorf("unsupported retrieval task action %q", normalized.Action)
	}
}

type normalizedRetrievalTaskRequest struct {
	Action     string
	Search     SearchOptions
	DocID      string
	ChunkID    string
	NodeID     string
	EntityID   string
	ServiceID  string
	Records    RecordLookupOptions
	Services   ServiceLookupOptions
	Provenance ProvenanceEventOptions
	Projection ProjectionStateOptions
	Limit      int
}

func normalizeRetrievalTaskRequest(request RetrievalTaskRequest) (normalizedRetrievalTaskRequest, string) {
	action := strings.TrimSpace(request.Action)
	if action == "" {
		action = RetrievalTaskActionValidate
	}
	normalized := normalizedRetrievalTaskRequest{
		Action:     action,
		Search:     request.Search,
		DocID:      strings.TrimSpace(request.DocID),
		ChunkID:    strings.TrimSpace(request.ChunkID),
		NodeID:     strings.TrimSpace(request.NodeID),
		EntityID:   strings.TrimSpace(request.EntityID),
		ServiceID:  strings.TrimSpace(request.ServiceID),
		Records:    request.Records,
		Services:   request.Services,
		Provenance: request.Provenance,
		Projection: request.Projection,
		Limit:      request.Limit,
	}

	if request.Limit < 0 ||
		request.Search.Limit < 0 ||
		request.Records.Limit < 0 ||
		request.Services.Limit < 0 ||
		request.Provenance.Limit < 0 ||
		request.Projection.Limit < 0 {
		return normalizedRetrievalTaskRequest{}, "limit must be greater than or equal to 0"
	}

	switch action {
	case RetrievalTaskActionValidate:
		return normalized, ""
	case RetrievalTaskActionSearch:
		if strings.TrimSpace(request.Search.Text) == "" {
			return normalizedRetrievalTaskRequest{}, "search.text is required"
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
	case RetrievalTaskActionProvenanceEvents, RetrievalTaskActionProjectionStates:
		return normalized, ""
	default:
		return normalizedRetrievalTaskRequest{}, fmt.Sprintf("unsupported retrieval task action %q", action)
	}
}
