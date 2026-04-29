package runclient

import "time"

type DocumentInput struct {
	Path  string
	Title string
	Body  string
}

type SourceURLInput struct {
	URL           string
	PathHint      string
	AssetPathHint string
	Title         string
	Mode          string
}

type VideoURLInput struct {
	URL           string
	PathHint      string
	AssetPathHint string
	Title         string
	Mode          string
	Transcript    VideoTranscriptInput
}

type VideoTranscriptInput struct {
	Text       string
	Policy     string
	Origin     string
	Language   string
	CapturedAt string
	Tool       string
	Model      string
	SHA256     string
}

type SourcePDFMetadata struct {
	Title         string
	Author        string
	PublishedDate string
}

type SourceIngestionResult struct {
	DocID       string
	SourcePath  string
	AssetPath   string
	DerivedPath string
	Citations   []Citation
	SHA256      string
	SizeBytes   int64
	MIMEType    string
	PageCount   int
	CapturedAt  time.Time
	PDFMetadata SourcePDFMetadata
}

type VideoIngestionResult struct {
	DocID                    string
	SourcePath               string
	SourceURL                string
	AssetPath                string
	Citations                []Citation
	TranscriptSHA256         string
	PreviousTranscriptSHA256 string
	NewTranscriptSHA256      string
	CapturedAt               time.Time
	TranscriptPolicy         string
	TranscriptOrigin         string
	Language                 string
	Tool                     string
	Model                    string
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

type ServiceLookupOptions struct {
	Text      string
	Status    string
	Owner     string
	Interface string
	Limit     int
	Cursor    string
}

type DecisionLookupOptions struct {
	Text   string
	Status string
	Scope  string
	Owner  string
	Limit  int
	Cursor string
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

type DecisionRecord struct {
	DecisionID   string
	Title        string
	Status       string
	Scope        string
	Owner        string
	Date         string
	Summary      string
	Supersedes   []string
	SupersededBy []string
	SourceRefs   []string
	Citations    []Citation
	UpdatedAt    time.Time
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

type ServiceLookupResult struct {
	Services []ServiceRecord
	PageInfo PageInfo
}

type DecisionLookupResult struct {
	Decisions []DecisionRecord
	PageInfo  PageInfo
}

type ProvenanceEventList struct {
	Events   []ProvenanceEvent
	PageInfo PageInfo
}

type ProjectionStateList struct {
	Projections []ProjectionState
	PageInfo    PageInfo
}
