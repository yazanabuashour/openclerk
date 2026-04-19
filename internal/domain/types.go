package domain

import (
	"context"
	"time"
)

type BackendKind string

const (
	BackendOpenClerk BackendKind = "openclerk"
)

type Capabilities struct {
	Backend     BackendKind
	AuthMode    string
	SearchModes []string
	Extensions  []string
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

type Chunk struct {
	ChunkID   string
	DocID     string
	Path      string
	Heading   string
	Content   string
	LineStart int
	LineEnd   int
}

type Citation struct {
	DocID     string
	ChunkID   string
	Path      string
	Heading   string
	LineStart int
	LineEnd   int
}

type PageInfo struct {
	NextCursor string
	HasMore    bool
}

type SearchQuery struct {
	Text          string
	Limit         int
	Cursor        string
	PathPrefix    string
	MetadataKey   string
	MetadataValue string
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

type DocumentListQuery struct {
	PathPrefix    string
	MetadataKey   string
	MetadataValue string
	Limit         int
	Cursor        string
}

type DocumentListResult struct {
	Documents []DocumentSummary
	PageInfo  PageInfo
}

type CreateDocumentInput struct {
	Path  string
	Title string
	Body  string
}

type AppendDocumentInput struct {
	Content string
}

type ReplaceSectionInput struct {
	Heading string
	Content string
}

type GraphNeighborhoodInput struct {
	DocID   string
	ChunkID string
	NodeID  string
	Limit   int
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

type RecordLookupInput struct {
	Text       string
	EntityType string
	Limit      int
	Cursor     string
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

type RecordLookupResult struct {
	Entities []RecordEntity
	PageInfo PageInfo
}

type ServiceLookupInput struct {
	Text      string
	Status    string
	Owner     string
	Interface string
	Limit     int
	Cursor    string
}

type ServiceFact struct {
	Key        string
	Value      string
	ObservedAt *time.Time
}

type ServiceRecord struct {
	ServiceID string
	Name      string
	Status    string
	Owner     string
	Interface string
	Summary   string
	Facts     []ServiceFact
	Citations []Citation
	UpdatedAt time.Time
}

type ServiceLookupResult struct {
	Services []ServiceRecord
	PageInfo PageInfo
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

type ProvenanceEventQuery struct {
	RefKind   string
	RefID     string
	SourceRef string
	Limit     int
	Cursor    string
}

type ProvenanceEventResult struct {
	Events   []ProvenanceEvent
	PageInfo PageInfo
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

type ProjectionStateQuery struct {
	Projection string
	RefKind    string
	RefID      string
	Limit      int
	Cursor     string
}

type ProjectionStateResult struct {
	Projections []ProjectionState
	PageInfo    PageInfo
}

type Store interface {
	Capabilities(context.Context) (Capabilities, error)
	Search(context.Context, SearchQuery) (SearchResult, error)
	ListDocuments(context.Context, DocumentListQuery) (DocumentListResult, error)
	CreateDocument(context.Context, CreateDocumentInput) (Document, error)
	GetDocument(context.Context, string) (Document, error)
	GetDocumentLinks(context.Context, string) (DocumentLinks, error)
	AppendDocument(context.Context, string, AppendDocumentInput) (Document, error)
	ReplaceDocumentSection(context.Context, string, ReplaceSectionInput) (Document, error)
	GetChunk(context.Context, string) (Chunk, error)
	GraphNeighborhood(context.Context, GraphNeighborhoodInput) (GraphNeighborhood, error)
	RecordsLookup(context.Context, RecordLookupInput) (RecordLookupResult, error)
	GetRecordEntity(context.Context, string) (RecordEntity, error)
	ServicesLookup(context.Context, ServiceLookupInput) (ServiceLookupResult, error)
	GetServiceRecord(context.Context, string) (ServiceRecord, error)
	ListProvenanceEvents(context.Context, ProvenanceEventQuery) (ProvenanceEventResult, error)
	ListProjectionStates(context.Context, ProjectionStateQuery) (ProjectionStateResult, error)
	Close() error
}
