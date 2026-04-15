package domain

import (
	"context"
	"time"
)

type BackendKind string

const (
	BackendFTS     BackendKind = "fts"
	BackendHybrid  BackendKind = "hybrid"
	BackendGraph   BackendKind = "graph"
	BackendRecords BackendKind = "records"
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
	CreatedAt time.Time
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
	Text   string
	Limit  int
	Cursor string
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

type Store interface {
	Capabilities(context.Context) (Capabilities, error)
	Search(context.Context, SearchQuery) (SearchResult, error)
	CreateDocument(context.Context, CreateDocumentInput) (Document, error)
	GetDocument(context.Context, string) (Document, error)
	AppendDocument(context.Context, string, AppendDocumentInput) (Document, error)
	ReplaceDocumentSection(context.Context, string, ReplaceSectionInput) (Document, error)
	GetChunk(context.Context, string) (Chunk, error)
	GraphNeighborhood(context.Context, GraphNeighborhoodInput) (GraphNeighborhood, error)
	RecordsLookup(context.Context, RecordLookupInput) (RecordLookupResult, error)
	GetRecordEntity(context.Context, string) (RecordEntity, error)
	Close() error
}
