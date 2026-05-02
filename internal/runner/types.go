// Package runner executes task-shaped OpenClerk JSON requests.
package runner

import (
	"bytes"
	"encoding/json"
	"time"
)

const (
	DocumentTaskActionValidate        = "validate"
	DocumentTaskActionCreate          = "create_document"
	DocumentTaskActionIngestSourceURL = "ingest_source_url"
	DocumentTaskActionIngestVideoURL  = "ingest_video_url"
	DocumentTaskActionList            = "list_documents"
	DocumentTaskActionGet             = "get_document"
	DocumentTaskActionAppend          = "append_document"
	DocumentTaskActionReplaceSection  = "replace_section"
	DocumentTaskActionResolvePaths    = "resolve_paths"
	DocumentTaskActionInspectLayout   = "inspect_layout"

	RetrievalTaskActionValidate            = "validate"
	RetrievalTaskActionSearch              = "search"
	RetrievalTaskActionDocumentLinks       = "document_links"
	RetrievalTaskActionGraph               = "graph_neighborhood"
	RetrievalTaskActionRecordsLookup       = "records_lookup"
	RetrievalTaskActionRecordEntity        = "record_entity"
	RetrievalTaskActionServicesLookup      = "services_lookup"
	RetrievalTaskActionServiceRecord       = "service_record"
	RetrievalTaskActionDecisionsLookup     = "decisions_lookup"
	RetrievalTaskActionDecisionRecord      = "decision_record"
	RetrievalTaskActionProvenanceEvents    = "provenance_events"
	RetrievalTaskActionProjectionStates    = "projection_states"
	RetrievalTaskActionAuditContradictions = "audit_contradictions"
	RetrievalTaskActionMemoryRouterRecall  = "memory_router_recall_report"
)

type DocumentTaskRequest struct {
	Action   string              `json:"action"`
	Document DocumentInput       `json:"document,omitempty"`
	Source   SourceURLInput      `json:"source,omitempty"`
	Video    VideoURLInput       `json:"video,omitempty"`
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

type SourceURLInput struct {
	URL           string `json:"url"`
	PathHint      string `json:"path_hint"`
	AssetPathHint string `json:"asset_path_hint"`
	Title         string `json:"title,omitempty"`
	Mode          string `json:"mode,omitempty"`
	SourceType    string `json:"source_type,omitempty"`
}

type VideoURLInput struct {
	URL           string               `json:"url"`
	PathHint      string               `json:"path_hint"`
	AssetPathHint string               `json:"asset_path_hint,omitempty"`
	Title         string               `json:"title,omitempty"`
	Mode          string               `json:"mode,omitempty"`
	Transcript    VideoTranscriptInput `json:"transcript,omitempty"`
}

type VideoTranscriptInput struct {
	Text       string `json:"text,omitempty"`
	Policy     string `json:"policy,omitempty"`
	Origin     string `json:"origin,omitempty"`
	Language   string `json:"language,omitempty"`
	CapturedAt string `json:"captured_at,omitempty"`
	Tool       string `json:"tool,omitempty"`
	Model      string `json:"model,omitempty"`
	SHA256     string `json:"sha256,omitempty"`
}

type SourcePDFMetadata struct {
	Title         string `json:"title,omitempty"`
	Author        string `json:"author,omitempty"`
	PublishedDate string `json:"published_date,omitempty"`
}

type SourceIngestionResult struct {
	DocID               string                  `json:"doc_id"`
	SourcePath          string                  `json:"source_path"`
	SourceURL           string                  `json:"source_url"`
	SourceType          string                  `json:"source_type"`
	AssetPath           string                  `json:"asset_path,omitempty"`
	DerivedPath         string                  `json:"derived_path"`
	Citations           []Citation              `json:"citations,omitempty"`
	SHA256              string                  `json:"sha256"`
	SizeBytes           int64                   `json:"size_bytes"`
	MIMEType            string                  `json:"mime_type"`
	PageCount           int                     `json:"page_count"`
	CapturedAt          time.Time               `json:"captured_at"`
	PDFMetadata         SourcePDFMetadata       `json:"pdf_metadata,omitempty"`
	UpdateStatus        string                  `json:"update_status,omitempty"`
	NormalizedSourceURL string                  `json:"normalized_source_url,omitempty"`
	SourceDocID         string                  `json:"source_doc_id,omitempty"`
	PreviousSHA256      string                  `json:"previous_sha256,omitempty"`
	NewSHA256           string                  `json:"new_sha256,omitempty"`
	Changed             *bool                   `json:"changed,omitempty"`
	DuplicateStatus     string                  `json:"duplicate_status,omitempty"`
	StaleDependents     *[]SourceStaleDependent `json:"stale_dependents,omitempty"`
	ProjectionRefs      *[]SourceProjectionRef  `json:"projection_refs,omitempty"`
	ProvenanceRefs      *[]SourceProvenanceRef  `json:"provenance_refs,omitempty"`
	SynthesisRepaired   *bool                   `json:"synthesis_repaired,omitempty"`
	NoRepairWarning     string                  `json:"no_repair_warning,omitempty"`
}

type SourceStaleDependent struct {
	Path            string   `json:"path"`
	DocID           string   `json:"doc_id"`
	Projection      string   `json:"projection"`
	Freshness       string   `json:"freshness"`
	StaleSourceRefs []string `json:"stale_source_refs,omitempty"`
}

type SourceProjectionRef struct {
	Projection string `json:"projection"`
	RefKind    string `json:"ref_kind"`
	RefID      string `json:"ref_id"`
	Freshness  string `json:"freshness"`
	SourceRef  string `json:"source_ref,omitempty"`
}

type SourceProvenanceRef struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	RefKind   string `json:"ref_kind"`
	RefID     string `json:"ref_id"`
	SourceRef string `json:"source_ref,omitempty"`
}

type VideoIngestionResult struct {
	DocID                    string     `json:"doc_id"`
	SourcePath               string     `json:"source_path"`
	SourceURL                string     `json:"source_url"`
	AssetPath                string     `json:"asset_path,omitempty"`
	Citations                []Citation `json:"citations,omitempty"`
	TranscriptSHA256         string     `json:"transcript_sha256"`
	PreviousTranscriptSHA256 string     `json:"previous_transcript_sha256,omitempty"`
	NewTranscriptSHA256      string     `json:"new_transcript_sha256,omitempty"`
	CapturedAt               time.Time  `json:"captured_at"`
	TranscriptPolicy         string     `json:"transcript_policy"`
	TranscriptOrigin         string     `json:"transcript_origin"`
	Language                 string     `json:"language,omitempty"`
	Tool                     string     `json:"tool,omitempty"`
	Model                    string     `json:"model,omitempty"`
}

type DocumentListOptions struct {
	PathPrefix    string `json:"path_prefix,omitempty"`
	MetadataKey   string `json:"metadata_key,omitempty"`
	MetadataValue string `json:"metadata_value,omitempty"`
	Tag           string `json:"tag,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	Cursor        string `json:"cursor,omitempty"`

	tagProvided bool
}

type DocumentTaskResult struct {
	Rejected        bool                   `json:"rejected"`
	RejectionReason string                 `json:"rejection_reason,omitempty"`
	Document        *Document              `json:"document,omitempty"`
	Ingestion       *SourceIngestionResult `json:"ingestion,omitempty"`
	VideoIngestion  *VideoIngestionResult  `json:"video_ingestion,omitempty"`
	Documents       []DocumentSummary      `json:"documents,omitempty"`
	Paths           *Paths                 `json:"paths,omitempty"`
	Layout          *KnowledgeLayout       `json:"layout,omitempty"`
	PageInfo        PageInfo               `json:"page_info,omitempty"`
	Summary         string                 `json:"summary"`
}

type RetrievalTaskRequest struct {
	Action             string                     `json:"action"`
	Search             SearchOptions              `json:"search,omitempty"`
	DocID              string                     `json:"doc_id,omitempty"`
	ChunkID            string                     `json:"chunk_id,omitempty"`
	NodeID             string                     `json:"node_id,omitempty"`
	EntityID           string                     `json:"entity_id,omitempty"`
	ServiceID          string                     `json:"service_id,omitempty"`
	DecisionID         string                     `json:"decision_id,omitempty"`
	Records            RecordLookupOptions        `json:"records,omitempty"`
	Services           ServiceLookupOptions       `json:"services,omitempty"`
	Decisions          DecisionLookupOptions      `json:"decisions,omitempty"`
	Provenance         ProvenanceEventOptions     `json:"provenance,omitempty"`
	Projection         ProjectionStateOptions     `json:"projection,omitempty"`
	Audit              AuditContradictionsOptions `json:"audit,omitempty"`
	MemoryRouterRecall MemoryRouterRecallOptions  `json:"memory_router_recall,omitempty"`
	Limit              int                        `json:"limit,omitempty"`
}

type SearchOptions struct {
	Text          string `json:"text,omitempty"`
	PathPrefix    string `json:"path_prefix,omitempty"`
	MetadataKey   string `json:"metadata_key,omitempty"`
	MetadataValue string `json:"metadata_value,omitempty"`
	Tag           string `json:"tag,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	Cursor        string `json:"cursor,omitempty"`

	tagProvided bool
}

func (options *DocumentListOptions) UnmarshalJSON(data []byte) error {
	type documentListOptionsAlias DocumentListOptions
	var decoded struct {
		documentListOptionsAlias
		Tag *string `json:"tag"`
	}
	if err := decodeStrictJSON(data, &decoded); err != nil {
		return err
	}
	*options = DocumentListOptions(decoded.documentListOptionsAlias)
	if decoded.Tag != nil {
		options.Tag = *decoded.Tag
		options.tagProvided = true
	}
	return nil
}

func (options *SearchOptions) UnmarshalJSON(data []byte) error {
	type searchOptionsAlias SearchOptions
	var decoded struct {
		searchOptionsAlias
		Tag *string `json:"tag"`
	}
	if err := decodeStrictJSON(data, &decoded); err != nil {
		return err
	}
	*options = SearchOptions(decoded.searchOptionsAlias)
	if decoded.Tag != nil {
		options.Tag = *decoded.Tag
		options.tagProvided = true
	}
	return nil
}

func decodeStrictJSON(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}

type RecordLookupOptions struct {
	Text       string `json:"text,omitempty"`
	EntityType string `json:"entity_type,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
}

type ServiceLookupOptions struct {
	Text      string `json:"text,omitempty"`
	Status    string `json:"status,omitempty"`
	Owner     string `json:"owner,omitempty"`
	Interface string `json:"interface,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
}

type DecisionLookupOptions struct {
	Text   string `json:"text,omitempty"`
	Status string `json:"status,omitempty"`
	Scope  string `json:"scope,omitempty"`
	Owner  string `json:"owner,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
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

type AuditContradictionsOptions struct {
	Query         string `json:"query,omitempty"`
	TargetPath    string `json:"target_path,omitempty"`
	Mode          string `json:"mode,omitempty"`
	ConflictQuery string `json:"conflict_query,omitempty"`
	Limit         int    `json:"limit,omitempty"`
}

type MemoryRouterRecallOptions struct {
	Query string `json:"query,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

type RetrievalTaskResult struct {
	Rejected           bool                       `json:"rejected"`
	RejectionReason    string                     `json:"rejection_reason,omitempty"`
	Search             *SearchResult              `json:"search,omitempty"`
	Links              *DocumentLinks             `json:"links,omitempty"`
	Graph              *GraphNeighborhood         `json:"graph,omitempty"`
	Records            *RecordLookupResult        `json:"records,omitempty"`
	Entity             *RecordEntity              `json:"entity,omitempty"`
	Services           *ServiceLookupResult       `json:"services,omitempty"`
	Service            *ServiceRecord             `json:"service,omitempty"`
	Decisions          *DecisionLookupResult      `json:"decisions,omitempty"`
	Decision           *DecisionRecord            `json:"decision,omitempty"`
	Provenance         *ProvenanceEventList       `json:"provenance,omitempty"`
	Projections        *ProjectionStateList       `json:"projections,omitempty"`
	Audit              *AuditContradictionsResult `json:"audit,omitempty"`
	MemoryRouterRecall *MemoryRouterRecallReport  `json:"memory_router_recall,omitempty"`
	Summary            string                     `json:"summary"`
}

type MemoryRouterRecallReport struct {
	QuerySummary          string   `json:"query_summary"`
	TemporalStatus        string   `json:"temporal_status"`
	CanonicalEvidenceRefs []string `json:"canonical_evidence_refs"`
	StaleSessionStatus    string   `json:"stale_session_status"`
	FeedbackWeighting     string   `json:"feedback_weighting"`
	RoutingRationale      string   `json:"routing_rationale"`
	ProvenanceRefs        []string `json:"provenance_refs"`
	SynthesisFreshness    string   `json:"synthesis_freshness"`
	ValidationBoundaries  string   `json:"validation_boundaries"`
	AuthorityLimits       string   `json:"authority_limits"`
}

type AuditContradictionsResult struct {
	Query                     string                      `json:"query"`
	TargetPath                string                      `json:"target_path"`
	Mode                      string                      `json:"mode"`
	SelectedTargetPath        string                      `json:"selected_target_path,omitempty"`
	CandidateSynthesisPaths   []string                    `json:"candidate_synthesis_paths,omitempty"`
	SourcePaths               []string                    `json:"source_paths,omitempty"`
	Citations                 []Citation                  `json:"citations,omitempty"`
	CurrentSourcePaths        []string                    `json:"current_source_paths,omitempty"`
	SupersededSourcePaths     []string                    `json:"superseded_source_paths,omitempty"`
	ProvenanceInspected       []AuditProvenanceInspection `json:"provenance_inspected,omitempty"`
	ProjectionFreshnessBefore []ProjectionState           `json:"projection_freshness_before,omitempty"`
	ProjectionFreshnessAfter  []ProjectionState           `json:"projection_freshness_after,omitempty"`
	RepairStatus              string                      `json:"repair_status"`
	RepairApplied             bool                        `json:"repair_applied"`
	DuplicatePrevention       string                      `json:"duplicate_prevention"`
	UnresolvedConflictGroups  []AuditConflictGroup        `json:"unresolved_conflict_groups,omitempty"`
	FailureClassification     string                      `json:"failure_classification"`
}

type AuditProvenanceInspection struct {
	RefKind    string            `json:"ref_kind"`
	RefID      string            `json:"ref_id"`
	SourcePath string            `json:"source_path,omitempty"`
	EventIDs   []string          `json:"event_ids,omitempty"`
	EventTypes []string          `json:"event_types,omitempty"`
	Details    map[string]string `json:"details,omitempty"`
}

type AuditConflictGroup struct {
	Query       string   `json:"query"`
	SourcePaths []string `json:"source_paths,omitempty"`
	Claims      []string `json:"claims,omitempty"`
	Status      string   `json:"status"`
	Reason      string   `json:"reason"`
}

type Paths struct {
	DatabasePath string `json:"database_path"`
	VaultRoot    string `json:"vault_root"`
}

type KnowledgeLayout struct {
	Valid                  bool                   `json:"valid"`
	Mode                   string                 `json:"mode"`
	ConfigArtifactRequired bool                   `json:"config_artifact_required"`
	ConfigArtifact         string                 `json:"config_artifact"`
	Paths                  Paths                  `json:"paths"`
	ConventionalPaths      []LayoutPathConvention `json:"conventional_paths,omitempty"`
	DocumentKinds          []LayoutDocumentKind   `json:"document_kinds,omitempty"`
	Checks                 []KnowledgeLayoutCheck `json:"checks,omitempty"`
}

type LayoutPathConvention struct {
	Name        string `json:"name"`
	PathPrefix  string `json:"path_prefix"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type LayoutDocumentKind struct {
	Kind        string   `json:"kind"`
	Description string   `json:"description"`
	Selectors   []string `json:"selectors,omitempty"`
	Required    []string `json:"required,omitempty"`
}

type KnowledgeLayoutCheck struct {
	ID      string            `json:"id"`
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Path    string            `json:"path,omitempty"`
	DocID   string            `json:"doc_id,omitempty"`
	Details map[string]string `json:"details,omitempty"`
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

type ServiceFact struct {
	Key        string     `json:"key"`
	Value      string     `json:"value"`
	ObservedAt *time.Time `json:"observed_at,omitempty"`
}

type ServiceRecord struct {
	ServiceID string        `json:"service_id"`
	Name      string        `json:"name"`
	Status    string        `json:"status,omitempty"`
	Owner     string        `json:"owner,omitempty"`
	Interface string        `json:"interface,omitempty"`
	Summary   string        `json:"summary"`
	Facts     []ServiceFact `json:"facts,omitempty"`
	Citations []Citation    `json:"citations,omitempty"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type ServiceLookupResult struct {
	Services []ServiceRecord `json:"services,omitempty"`
	PageInfo PageInfo        `json:"page_info,omitempty"`
}

type DecisionRecord struct {
	DecisionID   string     `json:"decision_id"`
	Title        string     `json:"title"`
	Status       string     `json:"status,omitempty"`
	Scope        string     `json:"scope,omitempty"`
	Owner        string     `json:"owner,omitempty"`
	Date         string     `json:"date,omitempty"`
	Summary      string     `json:"summary"`
	Supersedes   []string   `json:"supersedes,omitempty"`
	SupersededBy []string   `json:"superseded_by,omitempty"`
	SourceRefs   []string   `json:"source_refs,omitempty"`
	Citations    []Citation `json:"citations,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type DecisionLookupResult struct {
	Decisions []DecisionRecord `json:"decision_records,omitempty"`
	PageInfo  PageInfo         `json:"page_info,omitempty"`
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
