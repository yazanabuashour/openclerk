// Package agentops exposes task-shaped OpenClerk helpers for coding agents.
package agentops

import "time"

const (
	DocumentTaskActionValidate       = "validate"
	DocumentTaskActionCreate         = "create_document"
	DocumentTaskActionList           = "list_documents"
	DocumentTaskActionGet            = "get_document"
	DocumentTaskActionAppend         = "append_document"
	DocumentTaskActionReplaceSection = "replace_section"
	DocumentTaskActionResolvePaths   = "resolve_paths"

	RetrievalTaskActionValidate         = "validate"
	RetrievalTaskActionSearch           = "search"
	RetrievalTaskActionDocumentLinks    = "document_links"
	RetrievalTaskActionGraph            = "graph_neighborhood"
	RetrievalTaskActionRecordsLookup    = "records_lookup"
	RetrievalTaskActionRecordEntity     = "record_entity"
	RetrievalTaskActionProvenanceEvents = "provenance_events"
	RetrievalTaskActionProjectionStates = "projection_states"
)

type DocumentTaskRequest struct {
	Action   string              `json:"action"`
	Document DocumentInput       `json:"document,omitempty"`
	DocID    string              `json:"doc_id,omitempty"`
	Content  string              `json:"content,omitempty"`
	Heading  string              `json:"heading,omitempty"`
	List     DocumentListOptions `json:"list,omitempty"`
}

type DocumentInput struct {
	Path  string `json:"path"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type DocumentListOptions struct {
	PathPrefix    string `json:"path_prefix,omitempty"`
	MetadataKey   string `json:"metadata_key,omitempty"`
	MetadataValue string `json:"metadata_value,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	Cursor        string `json:"cursor,omitempty"`
}

type DocumentTaskResult struct {
	Rejected        bool              `json:"rejected"`
	RejectionReason string            `json:"rejection_reason,omitempty"`
	Document        *Document         `json:"document,omitempty"`
	Documents       []DocumentSummary `json:"documents,omitempty"`
	Paths           *Paths            `json:"paths,omitempty"`
	PageInfo        PageInfo          `json:"page_info,omitempty"`
	Summary         string            `json:"summary"`
}

type RetrievalTaskRequest struct {
	Action     string                 `json:"action"`
	Search     SearchOptions          `json:"search,omitempty"`
	DocID      string                 `json:"doc_id,omitempty"`
	ChunkID    string                 `json:"chunk_id,omitempty"`
	NodeID     string                 `json:"node_id,omitempty"`
	EntityID   string                 `json:"entity_id,omitempty"`
	Records    RecordLookupOptions    `json:"records,omitempty"`
	Provenance ProvenanceEventOptions `json:"provenance,omitempty"`
	Projection ProjectionStateOptions `json:"projection,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
}

type SearchOptions struct {
	Text          string `json:"text,omitempty"`
	PathPrefix    string `json:"path_prefix,omitempty"`
	MetadataKey   string `json:"metadata_key,omitempty"`
	MetadataValue string `json:"metadata_value,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	Cursor        string `json:"cursor,omitempty"`
}

type RecordLookupOptions struct {
	Text       string `json:"text,omitempty"`
	EntityType string `json:"entity_type,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
}

type ProvenanceEventOptions struct {
	RefKind   string `json:"ref_kind,omitempty"`
	RefID     string `json:"ref_id,omitempty"`
	SourceRef string `json:"source_ref,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
}

type ProjectionStateOptions struct {
	Projection string `json:"projection,omitempty"`
	RefKind    string `json:"ref_kind,omitempty"`
	RefID      string `json:"ref_id,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
}

type RetrievalTaskResult struct {
	Rejected        bool                 `json:"rejected"`
	RejectionReason string               `json:"rejection_reason,omitempty"`
	Search          *SearchResult        `json:"search,omitempty"`
	Links           *DocumentLinks       `json:"links,omitempty"`
	Graph           *GraphNeighborhood   `json:"graph,omitempty"`
	Records         *RecordLookupResult  `json:"records,omitempty"`
	Entity          *RecordEntity        `json:"entity,omitempty"`
	Provenance      *ProvenanceEventList `json:"provenance,omitempty"`
	Projections     *ProjectionStateList `json:"projections,omitempty"`
	Summary         string               `json:"summary"`
}

type Paths struct {
	DataDir      string `json:"data_dir"`
	DatabasePath string `json:"database_path"`
	VaultRoot    string `json:"vault_root"`
}

type Document struct {
	DocID     string            `json:"doc_id"`
	Path      string            `json:"path"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Headings  []string          `json:"headings,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type DocumentSummary struct {
	DocID     string            `json:"doc_id"`
	Path      string            `json:"path"`
	Title     string            `json:"title"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type Citation struct {
	DocID     string `json:"doc_id"`
	ChunkID   string `json:"chunk_id"`
	Path      string `json:"path"`
	Heading   string `json:"heading,omitempty"`
	LineStart int    `json:"line_start,omitempty"`
	LineEnd   int    `json:"line_end,omitempty"`
}

type SearchHit struct {
	Rank      int        `json:"rank"`
	Score     float64    `json:"score"`
	DocID     string     `json:"doc_id"`
	ChunkID   string     `json:"chunk_id"`
	Title     string     `json:"title"`
	Snippet   string     `json:"snippet"`
	Citations []Citation `json:"citations,omitempty"`
}

type SearchResult struct {
	Hits     []SearchHit `json:"hits,omitempty"`
	PageInfo PageInfo    `json:"page_info,omitempty"`
}

type PageInfo struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more,omitempty"`
}

type DocumentLink struct {
	DocID     string     `json:"doc_id"`
	Path      string     `json:"path"`
	Title     string     `json:"title"`
	Citations []Citation `json:"citations,omitempty"`
}

type DocumentLinks struct {
	DocID    string         `json:"doc_id"`
	Outgoing []DocumentLink `json:"outgoing,omitempty"`
	Incoming []DocumentLink `json:"incoming,omitempty"`
}

type GraphNode struct {
	NodeID    string     `json:"node_id"`
	Type      string     `json:"type"`
	Label     string     `json:"label"`
	Citations []Citation `json:"citations,omitempty"`
}

type GraphEdge struct {
	EdgeID     string     `json:"edge_id"`
	FromNodeID string     `json:"from_node_id"`
	ToNodeID   string     `json:"to_node_id"`
	Kind       string     `json:"kind"`
	Citations  []Citation `json:"citations,omitempty"`
}

type GraphNeighborhood struct {
	Nodes []GraphNode `json:"nodes,omitempty"`
	Edges []GraphEdge `json:"edges,omitempty"`
}

type RecordFact struct {
	Key        string     `json:"key"`
	Value      string     `json:"value"`
	ObservedAt *time.Time `json:"observed_at,omitempty"`
}

type RecordEntity struct {
	EntityID   string       `json:"entity_id"`
	EntityType string       `json:"entity_type"`
	Name       string       `json:"name"`
	Summary    string       `json:"summary"`
	Facts      []RecordFact `json:"facts,omitempty"`
	Citations  []Citation   `json:"citations,omitempty"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

type RecordLookupResult struct {
	Entities []RecordEntity `json:"entities,omitempty"`
	PageInfo PageInfo       `json:"page_info,omitempty"`
}

type ProvenanceEvent struct {
	EventID    string            `json:"event_id"`
	EventType  string            `json:"event_type"`
	RefKind    string            `json:"ref_kind"`
	RefID      string            `json:"ref_id"`
	SourceRef  string            `json:"source_ref,omitempty"`
	OccurredAt time.Time         `json:"occurred_at"`
	Details    map[string]string `json:"details,omitempty"`
}

type ProvenanceEventList struct {
	Events   []ProvenanceEvent `json:"events,omitempty"`
	PageInfo PageInfo          `json:"page_info,omitempty"`
}

type ProjectionState struct {
	Projection        string            `json:"projection"`
	RefKind           string            `json:"ref_kind"`
	RefID             string            `json:"ref_id"`
	SourceRef         string            `json:"source_ref,omitempty"`
	Freshness         string            `json:"freshness"`
	ProjectionVersion string            `json:"projection_version"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Details           map[string]string `json:"details,omitempty"`
}

type ProjectionStateList struct {
	Projections []ProjectionState `json:"projections,omitempty"`
	PageInfo    PageInfo          `json:"page_info,omitempty"`
}
