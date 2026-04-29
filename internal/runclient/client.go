package runclient

import (
	"context"
	"github.com/yazanabuashour/openclerk/internal/app"
	"github.com/yazanabuashour/openclerk/internal/domain"
)

// Client is the preferred internal runner client for the embedded OpenClerk runtime.
type Client struct {
	runtime *Runtime
}

type Capabilities struct {
	Backend     string
	AuthMode    string
	SearchModes []string
	Extensions  []string
}

type GraphNeighborhood struct {
	Nodes []GraphNode
	Edges []GraphEdge
}

// Open creates the primary embedded OpenClerk client without binding a local port.
func Open(cfg Config) (*Client, error) {
	runtime, err := newRuntime(domain.BackendOpenClerk, cfg)
	if err != nil {
		return nil, wrapError(err)
	}
	return &Client{runtime: runtime}, nil
}

// OpenForWrite creates an embedded client for serialized runner write actions.
// Mutating actions sync the affected document after writing, so startup skips a
// vault-wide sync while the process-wide write lock is held.
func OpenForWrite(cfg Config) (*Client, error) {
	runtime, err := newWriteRuntime(domain.BackendOpenClerk, cfg)
	if err != nil {
		return nil, wrapError(err)
	}
	return &Client{runtime: runtime}, nil
}

// OpenReadOnly creates an embedded OpenClerk client for runner actions that do
// not mutate vault files, document registry rows, provenance, or projections.
func OpenReadOnly(cfg Config) (*Client, error) {
	runtime, err := newReadOnlyRuntime(domain.BackendOpenClerk, cfg)
	if err != nil {
		return nil, wrapError(err)
	}
	return &Client{runtime: runtime}, nil
}

// Close releases the internal runtime.
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

func (c *Client) IngestSourceURL(ctx context.Context, input SourceURLInput) (SourceIngestionResult, error) {
	service, err := c.service()
	if err != nil {
		return SourceIngestionResult{}, err
	}
	ingestion, err := service.IngestSourceURL(ctx, domain.SourceURLInput(input))
	if err != nil {
		return SourceIngestionResult{}, wrapError(err)
	}
	return toSourceIngestionResult(ingestion), nil
}

func (c *Client) IngestVideoURL(ctx context.Context, input VideoURLInput) (VideoIngestionResult, error) {
	service, err := c.service()
	if err != nil {
		return VideoIngestionResult{}, err
	}
	ingestion, err := service.IngestVideoURL(ctx, domain.VideoURLInput{
		URL:           input.URL,
		PathHint:      input.PathHint,
		AssetPathHint: input.AssetPathHint,
		Title:         input.Title,
		Mode:          input.Mode,
		Transcript: domain.VideoTranscriptInput{
			Text:       input.Transcript.Text,
			Policy:     input.Transcript.Policy,
			Origin:     input.Transcript.Origin,
			Language:   input.Transcript.Language,
			CapturedAt: input.Transcript.CapturedAt,
			Tool:       input.Transcript.Tool,
			Model:      input.Transcript.Model,
			SHA256:     input.Transcript.SHA256,
		},
	})
	if err != nil {
		return VideoIngestionResult{}, wrapError(err)
	}
	return toVideoIngestionResult(ingestion), nil
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

func (c *Client) LookupServices(ctx context.Context, options ServiceLookupOptions) (ServiceLookupResult, error) {
	service, err := c.service()
	if err != nil {
		return ServiceLookupResult{}, err
	}
	result, err := service.ServicesLookup(ctx, domain.ServiceLookupInput(options))
	if err != nil {
		return ServiceLookupResult{}, wrapError(err)
	}
	return ServiceLookupResult{
		Services: toServiceRecords(result.Services),
		PageInfo: toPageInfo(result.PageInfo),
	}, nil
}

func (c *Client) GetServiceRecord(ctx context.Context, serviceID string) (ServiceRecord, error) {
	service, err := c.service()
	if err != nil {
		return ServiceRecord{}, err
	}
	serviceRecord, err := service.GetServiceRecord(ctx, serviceID)
	if err != nil {
		return ServiceRecord{}, wrapError(err)
	}
	return toServiceRecord(serviceRecord), nil
}

func (c *Client) LookupDecisions(ctx context.Context, options DecisionLookupOptions) (DecisionLookupResult, error) {
	service, err := c.service()
	if err != nil {
		return DecisionLookupResult{}, err
	}
	result, err := service.DecisionsLookup(ctx, domain.DecisionLookupInput(options))
	if err != nil {
		return DecisionLookupResult{}, wrapError(err)
	}
	return DecisionLookupResult{
		Decisions: toDecisionRecords(result.Decisions),
		PageInfo:  toPageInfo(result.PageInfo),
	}, nil
}

func (c *Client) GetDecisionRecord(ctx context.Context, decisionID string) (DecisionRecord, error) {
	service, err := c.service()
	if err != nil {
		return DecisionRecord{}, err
	}
	decision, err := service.GetDecisionRecord(ctx, decisionID)
	if err != nil {
		return DecisionRecord{}, wrapError(err)
	}
	return toDecisionRecord(decision), nil
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
