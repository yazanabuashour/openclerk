package local

import (
	"context"
	"errors"
	"time"

	"github.com/yazanabuashour/openclerk/internal/app"
	"github.com/yazanabuashour/openclerk/internal/domain"
)

// Client is the preferred code-first facade for the embedded OpenClerk runtime.
type Client struct {
	runtime *Runtime
}

// Error is the public error shape returned by the local SDK facade.
type Error struct {
	Code      string
	Message   string
	Status    int
	Retryable bool
	Details   map[string]any
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

type Capabilities struct {
	Backend     string
	AuthMode    string
	SearchModes []string
	Extensions  []string
}

type DocumentInput struct {
	Path  string
	Title string
	Body  string
}

type DocumentListOptions struct {
	PathPrefix    string
	MetadataKey   string
	MetadataValue string
	Limit         int
	Cursor        string
}

type SearchOptions struct {
	Text          string
	PathPrefix    string
	MetadataKey   string
	MetadataValue string
	Limit         int
	Cursor        string
}

type RecordLookupOptions struct {
	Text       string
	EntityType string
	Limit      int
	Cursor     string
}

type GraphNeighborhoodOptions struct {
	DocID   string
	ChunkID string
	NodeID  string
	Limit   int
}

type ProvenanceEventOptions struct {
	RefKind   string
	RefID     string
	SourceRef string
	Limit     int
	Cursor    string
}

type ProjectionStateOptions struct {
	Projection string
	RefKind    string
	RefID      string
	Limit      int
	Cursor     string
}

type Document struct {
	DocID     string
	Path      string
	Title     string
	Body      string
	Headings  []string
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DocumentSummary struct {
	DocID     string
	Path      string
	Title     string
	Metadata  map[string]string
	UpdatedAt time.Time
}

type Citation struct {
	DocID     string
	ChunkID   string
	Path      string
	Heading   string
	LineStart int
	LineEnd   int
}

type SearchHit struct {
	Rank      int
	Score     float64
	DocID     string
	ChunkID   string
	Title     string
	Snippet   string
	Citations []Citation
}

type SearchResult struct {
	Hits     []SearchHit
	PageInfo PageInfo
}

type PageInfo struct {
	NextCursor string
	HasMore    bool
}

type DocumentLink struct {
	DocID     string
	Path      string
	Title     string
	Citations []Citation
}

type DocumentLinks struct {
	DocID    string
	Outgoing []DocumentLink
	Incoming []DocumentLink
}

type GraphNode struct {
	NodeID    string
	Type      string
	Label     string
	Citations []Citation
}

type GraphEdge struct {
	EdgeID     string
	FromNodeID string
	ToNodeID   string
	Kind       string
	Citations  []Citation
}

type GraphNeighborhood struct {
	Nodes []GraphNode
	Edges []GraphEdge
}

type RecordFact struct {
	Key        string
	Value      string
	ObservedAt *time.Time
}

type RecordEntity struct {
	EntityID   string
	EntityType string
	Name       string
	Summary    string
	Facts      []RecordFact
	Citations  []Citation
	UpdatedAt  time.Time
}

type ProvenanceEvent struct {
	EventID    string
	EventType  string
	RefKind    string
	RefID      string
	SourceRef  string
	OccurredAt time.Time
	Details    map[string]string
}

type ProjectionState struct {
	Projection        string
	RefKind           string
	RefID             string
	SourceRef         string
	Freshness         string
	ProjectionVersion string
	UpdatedAt         time.Time
	Details           map[string]string
}

type DocumentList struct {
	Documents []DocumentSummary
	PageInfo  PageInfo
}

type RecordLookupResult struct {
	Entities []RecordEntity
	PageInfo PageInfo
}

type ProvenanceEventList struct {
	Events   []ProvenanceEvent
	PageInfo PageInfo
}

type ProjectionStateList struct {
	Projections []ProjectionState
	PageInfo    PageInfo
}

// OpenClient creates the primary embedded OpenClerk facade without binding a local port.
func OpenClient(cfg Config) (*Client, error) {
	runtime, err := newRuntime(domain.BackendOpenClerk, withDefaultEmbeddingProvider(cfg))
	if err != nil {
		return nil, wrapError(err)
	}
	return &Client{runtime: runtime}, nil
}

// Close releases the embedded runtime.
func (c *Client) Close() error {
	if c == nil || c.runtime == nil {
		return nil
	}
	return wrapError(c.runtime.Close())
}

// Paths returns the resolved storage locations for this client.
func (c *Client) Paths() Paths {
	if c == nil || c.runtime == nil {
		return Paths{}
	}
	return c.runtime.Paths()
}

func (c *Client) Capabilities(ctx context.Context) (Capabilities, error) {
	service, err := c.service()
	if err != nil {
		return Capabilities{}, err
	}
	capabilities, err := service.Capabilities(ctx)
	if err != nil {
		return Capabilities{}, wrapError(err)
	}
	return Capabilities{
		Backend:     string(capabilities.Backend),
		AuthMode:    capabilities.AuthMode,
		SearchModes: append([]string(nil), capabilities.SearchModes...),
		Extensions:  append([]string(nil), capabilities.Extensions...),
	}, nil
}

func (c *Client) CreateDocument(ctx context.Context, input DocumentInput) (Document, error) {
	service, err := c.service()
	if err != nil {
		return Document{}, err
	}
	document, err := service.CreateDocument(ctx, domain.CreateDocumentInput(input))
	if err != nil {
		return Document{}, wrapError(err)
	}
	return toDocument(document), nil
}

func (c *Client) GetDocument(ctx context.Context, docID string) (Document, error) {
	service, err := c.service()
	if err != nil {
		return Document{}, err
	}
	document, err := service.GetDocument(ctx, docID)
	if err != nil {
		return Document{}, wrapError(err)
	}
	return toDocument(document), nil
}

func (c *Client) ListDocuments(ctx context.Context, options DocumentListOptions) (DocumentList, error) {
	service, err := c.service()
	if err != nil {
		return DocumentList{}, err
	}
	result, err := service.ListDocuments(ctx, domain.DocumentListQuery(options))
	if err != nil {
		return DocumentList{}, wrapError(err)
	}
	return DocumentList{
		Documents: toDocumentSummaries(result.Documents),
		PageInfo:  toPageInfo(result.PageInfo),
	}, nil
}

func (c *Client) Search(ctx context.Context, options SearchOptions) (SearchResult, error) {
	service, err := c.service()
	if err != nil {
		return SearchResult{}, err
	}
	result, err := service.Search(ctx, domain.SearchQuery{
		Text:          options.Text,
		Limit:         options.Limit,
		Cursor:        options.Cursor,
		PathPrefix:    options.PathPrefix,
		MetadataKey:   options.MetadataKey,
		MetadataValue: options.MetadataValue,
	})
	if err != nil {
		return SearchResult{}, wrapError(err)
	}
	return toSearchResult(result), nil
}

func (c *Client) AppendDocument(ctx context.Context, docID string, content string) (Document, error) {
	service, err := c.service()
	if err != nil {
		return Document{}, err
	}
	document, err := service.AppendDocument(ctx, docID, domain.AppendDocumentInput{Content: content})
	if err != nil {
		return Document{}, wrapError(err)
	}
	return toDocument(document), nil
}

func (c *Client) ReplaceSection(ctx context.Context, docID string, heading string, content string) (Document, error) {
	service, err := c.service()
	if err != nil {
		return Document{}, err
	}
	document, err := service.ReplaceDocumentSection(ctx, docID, domain.ReplaceSectionInput{
		Heading: heading,
		Content: content,
	})
	if err != nil {
		return Document{}, wrapError(err)
	}
	return toDocument(document), nil
}

func (c *Client) GetDocumentLinks(ctx context.Context, docID string) (DocumentLinks, error) {
	service, err := c.service()
	if err != nil {
		return DocumentLinks{}, err
	}
	links, err := service.GetDocumentLinks(ctx, docID)
	if err != nil {
		return DocumentLinks{}, wrapError(err)
	}
	return toDocumentLinksResult(links), nil
}

func (c *Client) GraphNeighborhood(ctx context.Context, options GraphNeighborhoodOptions) (GraphNeighborhood, error) {
	service, err := c.service()
	if err != nil {
		return GraphNeighborhood{}, err
	}
	neighborhood, err := service.GraphNeighborhood(ctx, domain.GraphNeighborhoodInput(options))
	if err != nil {
		return GraphNeighborhood{}, wrapError(err)
	}
	return toGraphNeighborhood(neighborhood), nil
}

func (c *Client) LookupRecords(ctx context.Context, options RecordLookupOptions) (RecordLookupResult, error) {
	service, err := c.service()
	if err != nil {
		return RecordLookupResult{}, err
	}
	result, err := service.RecordsLookup(ctx, domain.RecordLookupInput(options))
	if err != nil {
		return RecordLookupResult{}, wrapError(err)
	}
	return RecordLookupResult{
		Entities: toRecordEntities(result.Entities),
		PageInfo: toPageInfo(result.PageInfo),
	}, nil
}

func (c *Client) GetRecordEntity(ctx context.Context, entityID string) (RecordEntity, error) {
	service, err := c.service()
	if err != nil {
		return RecordEntity{}, err
	}
	entity, err := service.GetRecordEntity(ctx, entityID)
	if err != nil {
		return RecordEntity{}, wrapError(err)
	}
	return toRecordEntity(entity), nil
}

func (c *Client) ListProvenanceEvents(ctx context.Context, options ProvenanceEventOptions) (ProvenanceEventList, error) {
	service, err := c.service()
	if err != nil {
		return ProvenanceEventList{}, err
	}
	result, err := service.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery(options))
	if err != nil {
		return ProvenanceEventList{}, wrapError(err)
	}
	return ProvenanceEventList{
		Events:   toProvenanceEvents(result.Events),
		PageInfo: toPageInfo(result.PageInfo),
	}, nil
}

func (c *Client) ListProjectionStates(ctx context.Context, options ProjectionStateOptions) (ProjectionStateList, error) {
	service, err := c.service()
	if err != nil {
		return ProjectionStateList{}, err
	}
	result, err := service.ListProjectionStates(ctx, domain.ProjectionStateQuery(options))
	if err != nil {
		return ProjectionStateList{}, wrapError(err)
	}
	return ProjectionStateList{
		Projections: toProjectionStates(result.Projections),
		PageInfo:    toPageInfo(result.PageInfo),
	}, nil
}

func (c *Client) service() (*app.Service, error) {
	if c == nil || c.runtime == nil || c.runtime.service == nil {
		return nil, &Error{
			Code:    "invalid_client",
			Message: "local OpenClerk client is required",
			Status:  400,
		}
	}
	return c.runtime.service, nil
}

func toDocument(document domain.Document) Document {
	return Document{
		DocID:     document.DocID,
		Path:      document.Path,
		Title:     document.Title,
		Body:      document.Body,
		Headings:  append([]string(nil), document.Headings...),
		Metadata:  cloneStringMap(document.Metadata),
		CreatedAt: document.CreatedAt,
		UpdatedAt: document.UpdatedAt,
	}
}

func toDocumentSummaries(documents []domain.DocumentSummary) []DocumentSummary {
	result := make([]DocumentSummary, 0, len(documents))
	for _, document := range documents {
		result = append(result, DocumentSummary{
			DocID:     document.DocID,
			Path:      document.Path,
			Title:     document.Title,
			Metadata:  cloneStringMap(document.Metadata),
			UpdatedAt: document.UpdatedAt,
		})
	}
	return result
}

func toPageInfo(pageInfo domain.PageInfo) PageInfo {
	return PageInfo{
		NextCursor: pageInfo.NextCursor,
		HasMore:    pageInfo.HasMore,
	}
}

func toSearchResult(result domain.SearchResult) SearchResult {
	hits := make([]SearchHit, 0, len(result.Hits))
	for _, hit := range result.Hits {
		hits = append(hits, SearchHit{
			Rank:      hit.Rank,
			Score:     hit.Score,
			DocID:     hit.DocID,
			ChunkID:   hit.ChunkID,
			Title:     hit.Title,
			Snippet:   hit.Snippet,
			Citations: toCitations(hit.Citations),
		})
	}
	return SearchResult{
		Hits:     hits,
		PageInfo: toPageInfo(result.PageInfo),
	}
}

func toCitations(citations []domain.Citation) []Citation {
	result := make([]Citation, 0, len(citations))
	for _, citation := range citations {
		result = append(result, Citation{
			DocID:     citation.DocID,
			ChunkID:   citation.ChunkID,
			Path:      citation.Path,
			Heading:   citation.Heading,
			LineStart: citation.LineStart,
			LineEnd:   citation.LineEnd,
		})
	}
	return result
}

func toDocumentLinksResult(links domain.DocumentLinks) DocumentLinks {
	return DocumentLinks{
		DocID:    links.DocID,
		Outgoing: toDocumentLinks(links.Outgoing),
		Incoming: toDocumentLinks(links.Incoming),
	}
}

func toDocumentLinks(links []domain.DocumentLink) []DocumentLink {
	result := make([]DocumentLink, 0, len(links))
	for _, link := range links {
		result = append(result, DocumentLink{
			DocID:     link.DocID,
			Path:      link.Path,
			Title:     link.Title,
			Citations: toCitations(link.Citations),
		})
	}
	return result
}

func toGraphNeighborhood(neighborhood domain.GraphNeighborhood) GraphNeighborhood {
	nodes := make([]GraphNode, 0, len(neighborhood.Nodes))
	for _, node := range neighborhood.Nodes {
		nodes = append(nodes, GraphNode{
			NodeID:    node.NodeID,
			Type:      node.Type,
			Label:     node.Label,
			Citations: toCitations(node.Citations),
		})
	}
	edges := make([]GraphEdge, 0, len(neighborhood.Edges))
	for _, edge := range neighborhood.Edges {
		edges = append(edges, GraphEdge{
			EdgeID:     edge.EdgeID,
			FromNodeID: edge.FromNodeID,
			ToNodeID:   edge.ToNodeID,
			Kind:       edge.Kind,
			Citations:  toCitations(edge.Citations),
		})
	}
	return GraphNeighborhood{
		Nodes: nodes,
		Edges: edges,
	}
}

func toRecordEntities(entities []domain.RecordEntity) []RecordEntity {
	result := make([]RecordEntity, 0, len(entities))
	for _, entity := range entities {
		result = append(result, toRecordEntity(entity))
	}
	return result
}

func toRecordEntity(entity domain.RecordEntity) RecordEntity {
	facts := make([]RecordFact, 0, len(entity.Facts))
	for _, fact := range entity.Facts {
		facts = append(facts, RecordFact{
			Key:        fact.Key,
			Value:      fact.Value,
			ObservedAt: fact.ObservedAt,
		})
	}
	return RecordEntity{
		EntityID:   entity.EntityID,
		EntityType: entity.EntityType,
		Name:       entity.Name,
		Summary:    entity.Summary,
		Facts:      facts,
		Citations:  toCitations(entity.Citations),
		UpdatedAt:  entity.UpdatedAt,
	}
}

func toProvenanceEvents(events []domain.ProvenanceEvent) []ProvenanceEvent {
	result := make([]ProvenanceEvent, 0, len(events))
	for _, event := range events {
		result = append(result, ProvenanceEvent{
			EventID:    event.EventID,
			EventType:  event.EventType,
			RefKind:    event.RefKind,
			RefID:      event.RefID,
			SourceRef:  event.SourceRef,
			OccurredAt: event.OccurredAt,
			Details:    cloneStringMap(event.Details),
		})
	}
	return result
}

func toProjectionStates(projections []domain.ProjectionState) []ProjectionState {
	result := make([]ProjectionState, 0, len(projections))
	for _, projection := range projections {
		result = append(result, ProjectionState{
			Projection:        projection.Projection,
			RefKind:           projection.RefKind,
			RefID:             projection.RefID,
			SourceRef:         projection.SourceRef,
			Freshness:         projection.Freshness,
			ProjectionVersion: projection.ProjectionVersion,
			UpdatedAt:         projection.UpdatedAt,
			Details:           cloneStringMap(projection.Details),
		})
	}
	return result
}

func wrapError(err error) error {
	if err == nil {
		return nil
	}
	var localErr *Error
	if errors.As(err, &localErr) {
		return localErr
	}
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		return &Error{
			Code:      domainErr.Code,
			Message:   domainErr.Message,
			Status:    domainErr.Status,
			Retryable: domainErr.Retryable,
			Details:   cloneDetails(domainErr.Details),
		}
	}
	return &Error{
		Code:      "internal_error",
		Message:   err.Error(),
		Status:    500,
		Retryable: true,
	}
}

func cloneDetails(details map[string]any) map[string]any {
	if len(details) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(details))
	for key, value := range details {
		cloned[key] = value
	}
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
