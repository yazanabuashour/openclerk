package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/app"
	"github.com/yazanabuashour/openclerk/internal/domain"
)

type Server struct {
	service *app.Service
}

func NewHandler(service *app.Service) http.Handler {
	return &Server{service: service}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/v1/capabilities":
		resp, err := s.GetCapabilities(r.Context(), GetCapabilitiesRequestObject{})
		writeVisitResponse(w, err, func() error { return resp.VisitGetCapabilitiesResponse(w) })
	case r.Method == http.MethodPost && r.URL.Path == "/v1/search/query":
		var body SearchQuery
		if !decodeJSONBody(w, r, &body) {
			return
		}
		resp, err := s.SearchQuery(r.Context(), SearchQueryRequestObject{Body: &body})
		writeVisitResponse(w, err, func() error { return resp.VisitSearchQueryResponse(w) })
	case r.Method == http.MethodGet && r.URL.Path == "/v1/documents":
		params := ListDocumentsParams{
			PathPrefix:    optionalString(r.URL.Query().Get("pathPrefix")),
			MetadataKey:   optionalString(r.URL.Query().Get("metadataKey")),
			MetadataValue: optionalString(r.URL.Query().Get("metadataValue")),
			Limit:         optionalInt(r.URL.Query().Get("limit")),
			Cursor:        optionalString(r.URL.Query().Get("cursor")),
		}
		resp, err := s.ListDocuments(r.Context(), ListDocumentsRequestObject{Params: params})
		writeVisitResponse(w, err, func() error { return resp.VisitListDocumentsResponse(w) })
	case r.Method == http.MethodPost && r.URL.Path == "/v1/documents":
		var body CreateDocumentRequest
		if !decodeJSONBody(w, r, &body) {
			return
		}
		resp, err := s.CreateDocument(r.Context(), CreateDocumentRequestObject{Body: &body})
		writeVisitResponse(w, err, func() error { return resp.VisitCreateDocumentResponse(w) })
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/documents/") && strings.HasSuffix(r.URL.Path, "/links"):
		docID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/documents/"), "/links")
		resp, err := s.GetDocumentLinks(r.Context(), GetDocumentLinksRequestObject{DocId: DocId(docID)})
		writeVisitResponse(w, err, func() error { return resp.VisitGetDocumentLinksResponse(w) })
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/documents/") && !strings.Contains(r.URL.Path, ":"):
		docID := strings.TrimPrefix(r.URL.Path, "/v1/documents/")
		resp, err := s.GetDocument(r.Context(), GetDocumentRequestObject{DocId: DocId(docID)})
		writeVisitResponse(w, err, func() error { return resp.VisitGetDocumentResponse(w) })
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/documents/") && strings.HasSuffix(r.URL.Path, ":append"):
		docID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/documents/"), ":append")
		var body AppendDocumentRequest
		if !decodeJSONBody(w, r, &body) {
			return
		}
		resp, err := s.AppendDocument(r.Context(), AppendDocumentRequestObject{DocId: DocId(docID), Body: &body})
		writeVisitResponse(w, err, func() error { return resp.VisitAppendDocumentResponse(w) })
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/documents/") && strings.HasSuffix(r.URL.Path, ":replace-section"):
		docID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/documents/"), ":replace-section")
		var body ReplaceSectionRequest
		if !decodeJSONBody(w, r, &body) {
			return
		}
		resp, err := s.ReplaceDocumentSection(r.Context(), ReplaceDocumentSectionRequestObject{DocId: DocId(docID), Body: &body})
		writeVisitResponse(w, err, func() error { return resp.VisitReplaceDocumentSectionResponse(w) })
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/chunks/"):
		chunkID := strings.TrimPrefix(r.URL.Path, "/v1/chunks/")
		resp, err := s.GetChunk(r.Context(), GetChunkRequestObject{ChunkId: ChunkId(chunkID)})
		writeVisitResponse(w, err, func() error { return resp.VisitGetChunkResponse(w) })
	case r.Method == http.MethodPost && r.URL.Path == "/v1/extensions/graph/neighborhood":
		var body GraphNeighborhoodRequest
		if !decodeJSONBody(w, r, &body) {
			return
		}
		resp, err := s.GraphNeighborhood(r.Context(), GraphNeighborhoodRequestObject{Body: &body})
		writeVisitResponse(w, err, func() error { return resp.VisitGraphNeighborhoodResponse(w) })
	case r.Method == http.MethodPost && r.URL.Path == "/v1/extensions/records/lookup":
		var body RecordsLookupRequest
		if !decodeJSONBody(w, r, &body) {
			return
		}
		resp, err := s.RecordsLookup(r.Context(), RecordsLookupRequestObject{Body: &body})
		writeVisitResponse(w, err, func() error { return resp.VisitRecordsLookupResponse(w) })
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/extensions/records/entities/"):
		entityID := strings.TrimPrefix(r.URL.Path, "/v1/extensions/records/entities/")
		resp, err := s.GetRecordEntity(r.Context(), GetRecordEntityRequestObject{EntityId: EntityId(entityID)})
		writeVisitResponse(w, err, func() error { return resp.VisitGetRecordEntityResponse(w) })
	case r.Method == http.MethodGet && r.URL.Path == "/v1/provenance/events":
		params := ListProvenanceEventsParams{
			RefKind:   optionalString(r.URL.Query().Get("refKind")),
			RefId:     optionalString(r.URL.Query().Get("refId")),
			SourceRef: optionalString(r.URL.Query().Get("sourceRef")),
			Limit:     optionalInt(r.URL.Query().Get("limit")),
			Cursor:    optionalString(r.URL.Query().Get("cursor")),
		}
		resp, err := s.ListProvenanceEvents(r.Context(), ListProvenanceEventsRequestObject{Params: params})
		writeVisitResponse(w, err, func() error { return resp.VisitListProvenanceEventsResponse(w) })
	case r.Method == http.MethodGet && r.URL.Path == "/v1/provenance/projections":
		params := ListProjectionStatesParams{
			Projection: optionalString(r.URL.Query().Get("projection")),
			RefKind:    optionalString(r.URL.Query().Get("refKind")),
			RefId:      optionalString(r.URL.Query().Get("refId")),
			Limit:      optionalInt(r.URL.Query().Get("limit")),
			Cursor:     optionalString(r.URL.Query().Get("cursor")),
		}
		resp, err := s.ListProjectionStates(r.Context(), ListProjectionStatesRequestObject{Params: params})
		writeVisitResponse(w, err, func() error { return resp.VisitListProjectionStatesResponse(w) })
	default:
		writeJSONError(w, http.StatusNotFound, ErrorEnvelope{
			Code:      "not_found",
			Message:   "route not found",
			Retryable: false,
			RequestId: requestID(),
		})
	}
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, target any) bool {
	defer func() {
		_ = r.Body.Close()
	}()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrorEnvelope{
			Code:      "bad_request",
			Message:   err.Error(),
			Retryable: false,
			RequestId: requestID(),
		})
		return false
	}
	return true
}

func writeVisitResponse(w http.ResponseWriter, err error, visit func() error) {
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrorEnvelope{
			Code:      "internal_error",
			Message:   err.Error(),
			Retryable: true,
			RequestId: requestID(),
		})
		return
	}
	if err := visit(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrorEnvelope{
			Code:      "internal_error",
			Message:   err.Error(),
			Retryable: true,
			RequestId: requestID(),
		})
	}
}

func writeJSONError(w http.ResponseWriter, status int, envelope ErrorEnvelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope)
}

func (s *Server) GetCapabilities(ctx context.Context, _ GetCapabilitiesRequestObject) (GetCapabilitiesResponseObject, error) {
	capabilities, err := s.service.Capabilities(ctx)
	if err != nil {
		return defaultCapabilitiesResponse(err), nil
	}
	return GetCapabilities200JSONResponse{
		AuthMode:    CapabilitiesAuthMode(capabilities.AuthMode),
		Backend:     CapabilitiesBackend(capabilities.Backend),
		Extensions:  toCapabilityExtensions(capabilities.Extensions),
		SearchModes: toCapabilityModes(capabilities.SearchModes),
	}, nil
}

func (s *Server) SearchQuery(ctx context.Context, request SearchQueryRequestObject) (SearchQueryResponseObject, error) {
	if request.Body == nil {
		return SearchQuerydefaultJSONResponse{Body: errorEnvelope(domain.ValidationError("request body is required", nil)), StatusCode: 400}, nil
	}
	result, err := s.service.Search(ctx, domain.SearchQuery{
		Text:          request.Body.Text,
		Limit:         intValue(request.Body.Limit),
		Cursor:        stringValue(request.Body.Cursor),
		PathPrefix:    stringValue(request.Body.PathPrefix),
		MetadataKey:   stringValue(request.Body.MetadataKey),
		MetadataValue: stringValue(request.Body.MetadataValue),
	})
	if err != nil {
		return SearchQuerydefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return SearchQuery200JSONResponse(toSearchResponse(result)), nil
}

func (s *Server) ListDocuments(ctx context.Context, request ListDocumentsRequestObject) (ListDocumentsResponseObject, error) {
	result, err := s.service.ListDocuments(ctx, domain.DocumentListQuery{
		PathPrefix:    stringValue(request.Params.PathPrefix),
		MetadataKey:   stringValue(request.Params.MetadataKey),
		MetadataValue: stringValue(request.Params.MetadataValue),
		Limit:         intValue(request.Params.Limit),
		Cursor:        stringValue(request.Params.Cursor),
	})
	if err != nil {
		return ListDocumentsdefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return ListDocuments200JSONResponse(toDocumentListResponse(result)), nil
}

func (s *Server) CreateDocument(ctx context.Context, request CreateDocumentRequestObject) (CreateDocumentResponseObject, error) {
	if request.Body == nil {
		return CreateDocumentdefaultJSONResponse{Body: errorEnvelope(domain.ValidationError("request body is required", nil)), StatusCode: 400}, nil
	}
	document, err := s.service.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  request.Body.Path,
		Title: request.Body.Title,
		Body:  request.Body.Body,
	})
	if err != nil {
		return CreateDocumentdefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return CreateDocument201JSONResponse(toDocument(document)), nil
}

func (s *Server) GetDocument(ctx context.Context, request GetDocumentRequestObject) (GetDocumentResponseObject, error) {
	document, err := s.service.GetDocument(ctx, string(request.DocId))
	if err != nil {
		return GetDocumentdefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return GetDocument200JSONResponse(toDocument(document)), nil
}

func (s *Server) GetDocumentLinks(ctx context.Context, request GetDocumentLinksRequestObject) (GetDocumentLinksResponseObject, error) {
	links, err := s.service.GetDocumentLinks(ctx, string(request.DocId))
	if err != nil {
		return GetDocumentLinksdefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return GetDocumentLinks200JSONResponse(toDocumentLinksResponse(links)), nil
}

func (s *Server) AppendDocument(ctx context.Context, request AppendDocumentRequestObject) (AppendDocumentResponseObject, error) {
	if request.Body == nil {
		return AppendDocumentdefaultJSONResponse{Body: errorEnvelope(domain.ValidationError("request body is required", nil)), StatusCode: 400}, nil
	}
	document, err := s.service.AppendDocument(ctx, string(request.DocId), domain.AppendDocumentInput{
		Content: request.Body.Content,
	})
	if err != nil {
		return AppendDocumentdefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return AppendDocument200JSONResponse(toDocument(document)), nil
}

func (s *Server) ReplaceDocumentSection(ctx context.Context, request ReplaceDocumentSectionRequestObject) (ReplaceDocumentSectionResponseObject, error) {
	if request.Body == nil {
		return ReplaceDocumentSectiondefaultJSONResponse{Body: errorEnvelope(domain.ValidationError("request body is required", nil)), StatusCode: 400}, nil
	}
	document, err := s.service.ReplaceDocumentSection(ctx, string(request.DocId), domain.ReplaceSectionInput{
		Heading: request.Body.Heading,
		Content: request.Body.Content,
	})
	if err != nil {
		return ReplaceDocumentSectiondefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return ReplaceDocumentSection200JSONResponse(toDocument(document)), nil
}

func (s *Server) GetChunk(ctx context.Context, request GetChunkRequestObject) (GetChunkResponseObject, error) {
	chunk, err := s.service.GetChunk(ctx, string(request.ChunkId))
	if err != nil {
		return GetChunkdefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return GetChunk200JSONResponse(toChunk(chunk)), nil
}

func (s *Server) GraphNeighborhood(ctx context.Context, request GraphNeighborhoodRequestObject) (GraphNeighborhoodResponseObject, error) {
	if request.Body == nil {
		return GraphNeighborhooddefaultJSONResponse{Body: errorEnvelope(domain.ValidationError("request body is required", nil)), StatusCode: 400}, nil
	}
	result, err := s.service.GraphNeighborhood(ctx, domain.GraphNeighborhoodInput{
		DocID:   stringValue(request.Body.DocId),
		ChunkID: stringValue(request.Body.ChunkId),
		NodeID:  stringValue(request.Body.NodeId),
		Limit:   intValue(request.Body.Limit),
	})
	if err != nil {
		return GraphNeighborhooddefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return GraphNeighborhood200JSONResponse(toGraphNeighborhood(result)), nil
}

func (s *Server) RecordsLookup(ctx context.Context, request RecordsLookupRequestObject) (RecordsLookupResponseObject, error) {
	if request.Body == nil {
		return RecordsLookupdefaultJSONResponse{Body: errorEnvelope(domain.ValidationError("request body is required", nil)), StatusCode: 400}, nil
	}
	result, err := s.service.RecordsLookup(ctx, domain.RecordLookupInput{
		Text:       request.Body.Text,
		EntityType: stringValue(request.Body.EntityType),
		Limit:      intValue(request.Body.Limit),
		Cursor:     stringValue(request.Body.Cursor),
	})
	if err != nil {
		return RecordsLookupdefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return RecordsLookup200JSONResponse(toRecordsLookupResponse(result)), nil
}

func (s *Server) GetRecordEntity(ctx context.Context, request GetRecordEntityRequestObject) (GetRecordEntityResponseObject, error) {
	entity, err := s.service.GetRecordEntity(ctx, string(request.EntityId))
	if err != nil {
		return GetRecordEntitydefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return GetRecordEntity200JSONResponse(toRecordEntity(entity)), nil
}

func (s *Server) ListProvenanceEvents(ctx context.Context, request ListProvenanceEventsRequestObject) (ListProvenanceEventsResponseObject, error) {
	result, err := s.service.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind:   stringValue(request.Params.RefKind),
		RefID:     stringValue(request.Params.RefId),
		SourceRef: stringValue(request.Params.SourceRef),
		Limit:     intValue(request.Params.Limit),
		Cursor:    stringValue(request.Params.Cursor),
	})
	if err != nil {
		return ListProvenanceEventsdefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return ListProvenanceEvents200JSONResponse(toProvenanceEventsResponse(result)), nil
}

func (s *Server) ListProjectionStates(ctx context.Context, request ListProjectionStatesRequestObject) (ListProjectionStatesResponseObject, error) {
	result, err := s.service.ListProjectionStates(ctx, domain.ProjectionStateQuery{
		Projection: stringValue(request.Params.Projection),
		RefKind:    stringValue(request.Params.RefKind),
		RefID:      stringValue(request.Params.RefId),
		Limit:      intValue(request.Params.Limit),
		Cursor:     stringValue(request.Params.Cursor),
	})
	if err != nil {
		return ListProjectionStatesdefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}, nil
	}
	return ListProjectionStates200JSONResponse(toProjectionStatesResponse(result)), nil
}

func defaultCapabilitiesResponse(err error) GetCapabilitiesdefaultJSONResponse {
	return GetCapabilitiesdefaultJSONResponse{Body: errorEnvelope(err), StatusCode: statusCode(err)}
}

func errorEnvelope(err error) ErrorEnvelope {
	var appErr *domain.Error
	if !errors.As(err, &appErr) {
		appErr = domain.InternalError("unexpected server error", err)
	}
	envelope := ErrorEnvelope{
		Code:      appErr.Code,
		Message:   appErr.Message,
		Retryable: appErr.Retryable,
		RequestId: requestID(),
	}
	if len(appErr.Details) > 0 {
		details := make(map[string]interface{}, len(appErr.Details))
		for key, value := range appErr.Details {
			details[key] = value
		}
		envelope.Details = &details
	}
	return envelope
}

func statusCode(err error) int {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		return appErr.Status
	}
	return http.StatusInternalServerError
}

func toDocument(doc domain.Document) Document {
	return Document{
		Body:      doc.Body,
		CreatedAt: doc.CreatedAt,
		DocId:     doc.DocID,
		Headings:  doc.Headings,
		Metadata:  doc.Metadata,
		Path:      doc.Path,
		Title:     doc.Title,
		UpdatedAt: doc.UpdatedAt,
	}
}

func toDocumentListResponse(result domain.DocumentListResult) DocumentListResponse {
	documents := make([]DocumentSummary, 0, len(result.Documents))
	for _, document := range result.Documents {
		documents = append(documents, DocumentSummary{
			DocId:     document.DocID,
			Metadata:  document.Metadata,
			Path:      document.Path,
			Title:     document.Title,
			UpdatedAt: document.UpdatedAt,
		})
	}
	pageInfo := PageInfo{HasMore: result.PageInfo.HasMore}
	if result.PageInfo.NextCursor != "" {
		pageInfo.NextCursor = &result.PageInfo.NextCursor
	}
	return DocumentListResponse{Documents: documents, PageInfo: pageInfo}
}

func toDocumentLinksResponse(result domain.DocumentLinks) DocumentLinksResponse {
	return DocumentLinksResponse{
		DocId:    result.DocID,
		Incoming: toDocumentLinks(result.Incoming),
		Outgoing: toDocumentLinks(result.Outgoing),
	}
}

func toDocumentLinks(links []domain.DocumentLink) []DocumentLink {
	result := make([]DocumentLink, 0, len(links))
	for _, link := range links {
		result = append(result, DocumentLink{
			Citations: toCitations(link.Citations),
			DocId:     link.DocID,
			Path:      link.Path,
			Title:     link.Title,
		})
	}
	return result
}

func toChunk(chunk domain.Chunk) Chunk {
	return Chunk{
		ChunkId:   chunk.ChunkID,
		Content:   chunk.Content,
		DocId:     chunk.DocID,
		Heading:   chunk.Heading,
		LineEnd:   chunk.LineEnd,
		LineStart: chunk.LineStart,
		Path:      chunk.Path,
	}
}

func toSearchResponse(result domain.SearchResult) SearchResponse {
	hits := make([]SearchHit, 0, len(result.Hits))
	for _, hit := range result.Hits {
		hits = append(hits, SearchHit{
			Citations: toCitations(hit.Citations),
			ChunkId:   hit.ChunkID,
			DocId:     hit.DocID,
			Rank:      hit.Rank,
			Score:     hit.Score,
			Snippet:   hit.Snippet,
			Title:     hit.Title,
		})
	}
	pageInfo := PageInfo{HasMore: result.PageInfo.HasMore}
	if result.PageInfo.NextCursor != "" {
		pageInfo.NextCursor = &result.PageInfo.NextCursor
	}
	return SearchResponse{Hits: hits, PageInfo: pageInfo}
}

func toCitations(citations []domain.Citation) []Citation {
	result := make([]Citation, 0, len(citations))
	for _, citation := range citations {
		item := Citation{
			ChunkId:   citation.ChunkID,
			DocId:     citation.DocID,
			LineEnd:   citation.LineEnd,
			LineStart: citation.LineStart,
			Path:      citation.Path,
		}
		if strings.TrimSpace(citation.Heading) != "" {
			item.Heading = &citation.Heading
		}
		result = append(result, item)
	}
	return result
}

func toGraphNeighborhood(result domain.GraphNeighborhood) GraphNeighborhoodResponse {
	nodes := make([]GraphNode, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		nodes = append(nodes, GraphNode{
			Citations: toCitations(node.Citations),
			Label:     node.Label,
			NodeId:    node.NodeID,
			Type:      GraphNodeType(node.Type),
		})
	}
	edges := make([]GraphEdge, 0, len(result.Edges))
	for _, edge := range result.Edges {
		edges = append(edges, GraphEdge{
			Citations:  toCitations(edge.Citations),
			EdgeId:     edge.EdgeID,
			FromNodeId: edge.FromNodeID,
			Kind:       GraphEdgeKind(edge.Kind),
			ToNodeId:   edge.ToNodeID,
		})
	}
	return GraphNeighborhoodResponse{Edges: edges, Nodes: nodes}
}

func toRecordEntity(entity domain.RecordEntity) RecordEntity {
	facts := make([]RecordFact, 0, len(entity.Facts))
	for _, fact := range entity.Facts {
		item := RecordFact{
			Key:   fact.Key,
			Value: fact.Value,
		}
		if fact.ObservedAt != nil {
			item.ObservedAt = fact.ObservedAt
		}
		facts = append(facts, item)
	}
	return RecordEntity{
		Citations:  toCitations(entity.Citations),
		EntityId:   entity.EntityID,
		EntityType: entity.EntityType,
		Facts:      facts,
		Name:       entity.Name,
		Summary:    entity.Summary,
		UpdatedAt:  entity.UpdatedAt,
	}
}

func toRecordsLookupResponse(result domain.RecordLookupResult) RecordsLookupResponse {
	entities := make([]RecordEntity, 0, len(result.Entities))
	for _, entity := range result.Entities {
		entities = append(entities, toRecordEntity(entity))
	}
	pageInfo := PageInfo{HasMore: result.PageInfo.HasMore}
	if result.PageInfo.NextCursor != "" {
		pageInfo.NextCursor = &result.PageInfo.NextCursor
	}
	return RecordsLookupResponse{
		Entities: entities,
		PageInfo: pageInfo,
	}
}

func toProvenanceEventsResponse(result domain.ProvenanceEventResult) ProvenanceEventsResponse {
	events := make([]ProvenanceEvent, 0, len(result.Events))
	for _, event := range result.Events {
		events = append(events, ProvenanceEvent{
			Details:    event.Details,
			EventId:    event.EventID,
			EventType:  event.EventType,
			OccurredAt: event.OccurredAt,
			RefId:      event.RefID,
			RefKind:    ProvenanceEventRefKind(event.RefKind),
			SourceRef:  event.SourceRef,
		})
	}
	pageInfo := PageInfo{HasMore: result.PageInfo.HasMore}
	if result.PageInfo.NextCursor != "" {
		pageInfo.NextCursor = &result.PageInfo.NextCursor
	}
	return ProvenanceEventsResponse{Events: events, PageInfo: pageInfo}
}

func toProjectionStatesResponse(result domain.ProjectionStateResult) ProjectionStatesResponse {
	projections := make([]ProjectionState, 0, len(result.Projections))
	for _, projection := range result.Projections {
		projections = append(projections, ProjectionState{
			Details:           projection.Details,
			Freshness:         ProjectionStateFreshness(projection.Freshness),
			Projection:        projection.Projection,
			ProjectionVersion: projection.ProjectionVersion,
			RefId:             projection.RefID,
			RefKind:           ProjectionStateRefKind(projection.RefKind),
			SourceRef:         projection.SourceRef,
			UpdatedAt:         projection.UpdatedAt,
		})
	}
	pageInfo := PageInfo{HasMore: result.PageInfo.HasMore}
	if result.PageInfo.NextCursor != "" {
		pageInfo.NextCursor = &result.PageInfo.NextCursor
	}
	return ProjectionStatesResponse{PageInfo: pageInfo, Projections: projections}
}

func toCapabilityModes(values []string) []CapabilitiesSearchModes {
	result := make([]CapabilitiesSearchModes, 0, len(values))
	for _, value := range values {
		result = append(result, CapabilitiesSearchModes(value))
	}
	return result
}

func toCapabilityExtensions(values []string) []CapabilitiesExtensions {
	result := make([]CapabilitiesExtensions, 0, len(values))
	for _, value := range values {
		result = append(result, CapabilitiesExtensions(value))
	}
	return result
}

func requestID() string {
	return fmt.Sprintf("req-%08x", rand.Uint32())
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func optionalInt(raw string) *int {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &value
}
