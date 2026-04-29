package runclient

import (
	"context"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

// Client is the preferred internal runner client for the embedded OpenClerk runtime.
type Client struct {
	runtime *Runtime
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

func (c *Client) Capabilities(ctx context.Context) (domain.Capabilities, error) {
	store, err := c.store()
	if err != nil {
		return domain.Capabilities{}, err
	}
	return wrapResult(store.Capabilities(ctx))
}

func (c *Client) CreateDocument(ctx context.Context, input domain.CreateDocumentInput) (domain.Document, error) {
	store, err := c.store()
	if err != nil {
		return domain.Document{}, err
	}
	return wrapResult(store.CreateDocument(ctx, input))
}

func (c *Client) IngestSourceURL(ctx context.Context, input domain.SourceURLInput) (domain.SourceIngestionResult, error) {
	store, err := c.store()
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	return wrapResult(store.IngestSourceURL(ctx, input))
}

func (c *Client) IngestVideoURL(ctx context.Context, input domain.VideoURLInput) (domain.VideoIngestionResult, error) {
	store, err := c.store()
	if err != nil {
		return domain.VideoIngestionResult{}, err
	}
	return wrapResult(store.IngestVideoURL(ctx, input))
}

func (c *Client) GetDocument(ctx context.Context, docID string) (domain.Document, error) {
	store, err := c.store()
	if err != nil {
		return domain.Document{}, err
	}
	return wrapResult(store.GetDocument(ctx, docID))
}

func (c *Client) ListDocuments(ctx context.Context, query domain.DocumentListQuery) (domain.DocumentListResult, error) {
	store, err := c.store()
	if err != nil {
		return domain.DocumentListResult{}, err
	}
	return wrapResult(store.ListDocuments(ctx, query))
}

func (c *Client) Search(ctx context.Context, query domain.SearchQuery) (domain.SearchResult, error) {
	store, err := c.store()
	if err != nil {
		return domain.SearchResult{}, err
	}
	return wrapResult(store.Search(ctx, query))
}

func (c *Client) AppendDocument(ctx context.Context, docID string, input domain.AppendDocumentInput) (domain.Document, error) {
	store, err := c.store()
	if err != nil {
		return domain.Document{}, err
	}
	return wrapResult(store.AppendDocument(ctx, docID, input))
}

func (c *Client) ReplaceSection(ctx context.Context, docID string, input domain.ReplaceSectionInput) (domain.Document, error) {
	store, err := c.store()
	if err != nil {
		return domain.Document{}, err
	}
	return wrapResult(store.ReplaceDocumentSection(ctx, docID, input))
}

func (c *Client) GetDocumentLinks(ctx context.Context, docID string) (domain.DocumentLinks, error) {
	store, err := c.store()
	if err != nil {
		return domain.DocumentLinks{}, err
	}
	return wrapResult(store.GetDocumentLinks(ctx, docID))
}

func (c *Client) GraphNeighborhood(ctx context.Context, input domain.GraphNeighborhoodInput) (domain.GraphNeighborhood, error) {
	store, err := c.store()
	if err != nil {
		return domain.GraphNeighborhood{}, err
	}
	return wrapResult(store.GraphNeighborhood(ctx, input))
}

func (c *Client) LookupRecords(ctx context.Context, input domain.RecordLookupInput) (domain.RecordLookupResult, error) {
	store, err := c.store()
	if err != nil {
		return domain.RecordLookupResult{}, err
	}
	return wrapResult(store.RecordsLookup(ctx, input))
}

func (c *Client) GetRecordEntity(ctx context.Context, entityID string) (domain.RecordEntity, error) {
	store, err := c.store()
	if err != nil {
		return domain.RecordEntity{}, err
	}
	return wrapResult(store.GetRecordEntity(ctx, entityID))
}

func (c *Client) LookupServices(ctx context.Context, input domain.ServiceLookupInput) (domain.ServiceLookupResult, error) {
	store, err := c.store()
	if err != nil {
		return domain.ServiceLookupResult{}, err
	}
	return wrapResult(store.ServicesLookup(ctx, input))
}

func (c *Client) GetServiceRecord(ctx context.Context, serviceID string) (domain.ServiceRecord, error) {
	store, err := c.store()
	if err != nil {
		return domain.ServiceRecord{}, err
	}
	return wrapResult(store.GetServiceRecord(ctx, serviceID))
}

func (c *Client) LookupDecisions(ctx context.Context, input domain.DecisionLookupInput) (domain.DecisionLookupResult, error) {
	store, err := c.store()
	if err != nil {
		return domain.DecisionLookupResult{}, err
	}
	return wrapResult(store.DecisionsLookup(ctx, input))
}

func (c *Client) GetDecisionRecord(ctx context.Context, decisionID string) (domain.DecisionRecord, error) {
	store, err := c.store()
	if err != nil {
		return domain.DecisionRecord{}, err
	}
	return wrapResult(store.GetDecisionRecord(ctx, decisionID))
}

func (c *Client) ListProvenanceEvents(ctx context.Context, query domain.ProvenanceEventQuery) (domain.ProvenanceEventResult, error) {
	store, err := c.store()
	if err != nil {
		return domain.ProvenanceEventResult{}, err
	}
	return wrapResult(store.ListProvenanceEvents(ctx, query))
}

func (c *Client) ListProjectionStates(ctx context.Context, query domain.ProjectionStateQuery) (domain.ProjectionStateResult, error) {
	store, err := c.store()
	if err != nil {
		return domain.ProjectionStateResult{}, err
	}
	return wrapResult(store.ListProjectionStates(ctx, query))
}

func (c *Client) store() (domain.Store, error) {
	if c == nil || c.runtime == nil || c.runtime.store == nil {
		return nil, &Error{
			Code:    "invalid_client",
			Message: "local OpenClerk client is required",
			Status:  400,
		}
	}
	return c.runtime.store, nil
}

func wrapResult[T any](value T, err error) (T, error) {
	if err != nil {
		return value, wrapError(err)
	}
	return value, nil
}
