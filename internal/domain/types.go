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
	Tag           string
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
	Tag           string
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

type SourceURLInput struct {
	URL           string
	PathHint      string
	AssetPathHint string
	Title         string
	Mode          string
	SourceType    string
	Limit         int
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
	DocID               string
	SourcePath          string
	SourceURL           string
	SourceType          string
	AssetPath           string
	DerivedPath         string
	Citations           []Citation
	SHA256              string
	SizeBytes           int64
	MIMEType            string
	PageCount           int
	CapturedAt          time.Time
	PDFMetadata         SourcePDFMetadata
	UpdateStatus        string
	NormalizedSourceURL string
	SourceDocID         string
	PreviousSHA256      string
	NewSHA256           string
	Changed             bool
	DuplicateStatus     string
	StaleDependents     []SourceStaleDependent
	ProjectionRefs      []SourceProjectionRef
	ProvenanceRefs      []SourceProvenanceRef
	SynthesisRepaired   bool
	NoRepairWarning     string
}

type SourceURLInspection struct {
	SourceURL   string
	SourceType  string
	Title       string
	TextPreview string
	SHA256      string
	SizeBytes   int64
	MIMEType    string
	PageCount   int
	PDFMetadata SourcePDFMetadata
	Links       []SourceURLInspectionLink
}

type SourceURLInspectionLink struct {
	URL  string
	Text string
}

type SourceStaleDependent struct {
	Path            string
	DocID           string
	Projection      string
	Freshness       string
	StaleSourceRefs []string
}

type SourceProjectionRef struct {
	Projection string
	RefKind    string
	RefID      string
	Freshness  string
	SourceRef  string
}

type SourceProvenanceRef struct {
	EventID   string
	EventType string
	RefKind   string
	RefID     string
	SourceRef string
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

type AppendDocumentInput struct {
	Content string
}

type ReplaceSectionInput struct {
	Heading            string
	Content            string
	IncludeHeading     bool
	IncludeSubsections *bool
	DryRun             bool
}

type ReplaceDocumentInput struct {
	Path             string
	Title            string
	Body             string
	Metadata         map[string]string
	AllowDocIDChange bool
	DryRun           bool
}

type MoveDocumentInput struct {
	DocID         string
	Path          string
	TargetPath    string
	UpdateLinks   bool
	UpdateIndexes bool
}

type DocumentMovePlan struct {
	DocID                string
	SourcePath           string
	TargetPath           string
	Title                string
	FrontmatterID        string
	StableIDStatus       string
	DuplicateRisk        string
	ExistingTarget       *DocumentSummary
	OutgoingLinks        []DocumentLink
	IncomingLinks        []DocumentLink
	LinkUpdates          []DocumentLinkUpdate
	IndexUpdates         []DocumentIndexUpdate
	ProjectionRefresh    []DocumentProjectionRefresh
	ValidationWarnings   []string
	WriteStatus          string
	ApprovalBoundary     string
	ValidationBoundaries string
}

type DocumentMoveResult struct {
	Plan                DocumentMovePlan
	Document            Document
	LinkUpdatesApplied  []DocumentLinkUpdate
	IndexUpdatesApplied []DocumentIndexUpdate
	ProvenanceRefs      []string
	ProjectionFreshness []ProjectionState
	WriteStatus         string
}

type DocumentLinkUpdate struct {
	DocID          string
	Path           string
	OldTarget      string
	NewTarget      string
	Occurrences    int
	IndexCandidate bool
}

type DocumentIndexUpdate struct {
	Path   string
	Status string
	Reason string
}

type DocumentProjectionRefresh struct {
	Projection string
	RefKind    string
	RefID      string
	Status     string
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

type DecisionLookupInput struct {
	Text   string
	Status string
	Scope  string
	Owner  string
	Limit  int
	Cursor string
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

type DecisionLookupResult struct {
	Decisions []DecisionRecord
	PageInfo  PageInfo
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
	InspectSourceURL(context.Context, SourceURLInput) (SourceURLInspection, error)
	IngestSourceURL(context.Context, SourceURLInput) (SourceIngestionResult, error)
	IngestVideoURL(context.Context, VideoURLInput) (VideoIngestionResult, error)
	GetDocument(context.Context, string) (Document, error)
	GetDocumentByPath(context.Context, string) (Document, error)
	GetDocumentLinks(context.Context, string, ...int) (DocumentLinks, error)
	AppendDocument(context.Context, string, AppendDocumentInput) (Document, error)
	ReplaceDocumentSection(context.Context, string, ReplaceSectionInput) (Document, error)
	ReplaceDocument(context.Context, string, ReplaceDocumentInput) (Document, error)
	PlanMoveDocument(context.Context, MoveDocumentInput) (DocumentMovePlan, error)
	MoveDocument(context.Context, MoveDocumentInput) (DocumentMoveResult, error)
	GetChunk(context.Context, string) (Chunk, error)
	GraphNeighborhood(context.Context, GraphNeighborhoodInput) (GraphNeighborhood, error)
	RecordsLookup(context.Context, RecordLookupInput) (RecordLookupResult, error)
	GetRecordEntity(context.Context, string) (RecordEntity, error)
	ServicesLookup(context.Context, ServiceLookupInput) (ServiceLookupResult, error)
	GetServiceRecord(context.Context, string) (ServiceRecord, error)
	DecisionsLookup(context.Context, DecisionLookupInput) (DecisionLookupResult, error)
	GetDecisionRecord(context.Context, string) (DecisionRecord, error)
	ListProvenanceEvents(context.Context, ProvenanceEventQuery) (ProvenanceEventResult, error)
	ListProjectionStates(context.Context, ProjectionStateQuery) (ProjectionStateResult, error)
	Close() error
}
