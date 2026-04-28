package main

import (
	"context"
	"encoding/json"
	"time"
)

type runConfig struct {
	Parallel   int
	Variant    string
	Scenario   string
	RunRoot    string
	ReportDir  string
	ReportName string
	CodexBin   string
	RepoRoot   string
	CacheMode  string
}
type cacheConfig struct {
	Mode    string
	RunRoot string
}
type evalJob struct {
	Index    int
	Variant  string
	Scenario scenario
}
type scenario struct {
	ID     string         `json:"id"`
	Title  string         `json:"title"`
	Prompt string         `json:"prompt,omitempty"`
	Turns  []scenarioTurn `json:"turns,omitempty"`
}
type scenarioTurn struct {
	Prompt string `json:"prompt"`
}
type report struct {
	Metadata            reportMetadata         `json:"metadata"`
	Results             []jobResult            `json:"results"`
	ProductionGate      *productionGateSummary `json:"production_gate,omitempty"`
	TargetedLaneSummary *targetedLaneSummary   `json:"targeted_lane_summary,omitempty"`
}
type reportMetadata struct {
	GeneratedAt              time.Time    `json:"generated_at"`
	Model                    string       `json:"model"`
	ReasoningEffort          string       `json:"reasoning_effort"`
	Harness                  string       `json:"harness"`
	ConfiguredParallelism    int          `json:"configured_parallelism"`
	CacheMode                string       `json:"cache_mode"`
	CachePrewarmSeconds      float64      `json:"cache_prewarm_seconds,omitempty"`
	HarnessElapsedSeconds    float64      `json:"harness_elapsed_seconds"`
	EffectiveParallelSpeedup float64      `json:"effective_parallel_speedup,omitempty"`
	ParallelEfficiency       float64      `json:"parallel_efficiency,omitempty"`
	PhaseTotals              phaseTimings `json:"phase_totals"`
	RunRootArtifactReference string       `json:"run_root_artifact_reference"`
	RawLogPlaceholder        string       `json:"raw_log_placeholder"`
	Variants                 []string     `json:"variants"`
	Scenarios                []string     `json:"scenarios"`
	Lane                     string       `json:"lane"`
	ReleaseBlocking          bool         `json:"release_blocking"`
	TargetedAcceptanceNote   string       `json:"targeted_acceptance_note,omitempty"`
	RawLogsCommitted         bool         `json:"raw_logs_committed"`
	RawLogsNote              string       `json:"raw_logs_note"`
}
type phaseTimings struct {
	PrepareRunDir  float64 `json:"prepare_run_dir_seconds,omitempty"`
	CopyRepo       float64 `json:"copy_repo_seconds,omitempty"`
	InstallVariant float64 `json:"install_variant_seconds,omitempty"`
	WarmCache      float64 `json:"warm_cache_seconds,omitempty"`
	SeedData       float64 `json:"seed_data_seconds,omitempty"`
	AgentRun       float64 `json:"agent_run_seconds,omitempty"`
	ParseMetrics   float64 `json:"parse_metrics_seconds,omitempty"`
	Verify         float64 `json:"verify_seconds,omitempty"`
	Total          float64 `json:"total_seconds,omitempty"`
}
type jobResult struct {
	Variant                 string             `json:"variant"`
	Scenario                string             `json:"scenario"`
	ScenarioTitle           string             `json:"scenario_title"`
	Passed                  bool               `json:"passed"`
	Status                  string             `json:"status"`
	Error                   string             `json:"error,omitempty"`
	ExitCode                int                `json:"exit_code"`
	WallSeconds             float64            `json:"wall_seconds"`
	PhaseTimings            phaseTimings       `json:"phase_timings"`
	Metrics                 metrics            `json:"metrics"`
	FixturePreflight        *fixturePreflight  `json:"fixture_preflight,omitempty"`
	Verification            verificationResult `json:"verification"`
	Turns                   []turnResult       `json:"turns,omitempty"`
	PromptSummary           string             `json:"prompt_summary"`
	RawLogArtifactReference string             `json:"raw_log_artifact_reference"`
	StartedAt               time.Time          `json:"started_at"`
	CompletedAt             *time.Time         `json:"completed_at,omitempty"`
}
type turnResult struct {
	Index                   int                `json:"turn_index"`
	WallSeconds             float64            `json:"wall_seconds"`
	ExitCode                int                `json:"exit_code"`
	Metrics                 metrics            `json:"metrics"`
	Verification            verificationResult `json:"verification"`
	RawLogArtifactReference string             `json:"raw_log_artifact_reference"`
}
type metrics struct {
	AssistantCalls            int            `json:"assistant_calls"`
	ToolCalls                 int            `json:"tool_calls"`
	CommandExecutions         int            `json:"command_executions"`
	FileInspectionCommands    int            `json:"file_inspection_commands"`
	GeneratedFileInspection   bool           `json:"generated_file_inspection"`
	ModuleCacheInspection     bool           `json:"module_cache_inspection"`
	BroadRepoSearch           bool           `json:"broad_repo_search"`
	DirectSQLiteAccess        bool           `json:"direct_sqlite_access"`
	LegacyRunnerUsage         bool           `json:"legacy_runner_usage"`
	SearchUsed                bool           `json:"search_used"`
	SearchUnfilteredUsed      bool           `json:"search_unfiltered_used"`
	SearchPathFilterUsed      bool           `json:"search_path_filter_used"`
	SearchPathPrefixes        []string       `json:"search_path_prefixes,omitempty"`
	SearchMetadataFilterUsed  bool           `json:"search_metadata_filter_used"`
	SearchMetadataFilters     []string       `json:"search_metadata_filters,omitempty"`
	IngestSourceURLUsed       bool           `json:"ingest_source_url_used"`
	IngestSourceURLUpdateUsed bool           `json:"ingest_source_url_update_used"`
	IngestVideoURLUsed        bool           `json:"ingest_video_url_used"`
	IngestVideoURLUpdateUsed  bool           `json:"ingest_video_url_update_used"`
	SourcePDFDownloadFailure  bool           `json:"source_pdf_download_failure"`
	ValidateUsed              bool           `json:"validate_used"`
	CreateDocumentUsed        bool           `json:"create_document_used"`
	ReplaceSectionUsed        bool           `json:"replace_section_used"`
	AppendDocumentUsed        bool           `json:"append_document_used"`
	ListDocumentsUsed         bool           `json:"list_documents_used"`
	ListDocumentPathPrefixes  []string       `json:"list_document_path_prefixes,omitempty"`
	GetDocumentUsed           bool           `json:"get_document_used"`
	GetDocumentDocIDs         []string       `json:"get_document_doc_ids,omitempty"`
	InspectLayoutUsed         bool           `json:"inspect_layout_used"`
	DocumentLinksUsed         bool           `json:"document_links_used"`
	GraphNeighborhoodUsed     bool           `json:"graph_neighborhood_used"`
	RecordsLookupUsed         bool           `json:"records_lookup_used"`
	DecisionsLookupUsed       bool           `json:"decisions_lookup_used"`
	DecisionRecordUsed        bool           `json:"decision_record_used"`
	DecisionRecordIDs         []string       `json:"decision_record_ids,omitempty"`
	ProvenanceEventsUsed      bool           `json:"provenance_events_used"`
	ProvenanceEventRefIDs     []string       `json:"provenance_event_ref_ids,omitempty"`
	ProjectionStatesUsed      bool           `json:"projection_states_used"`
	GeneratedFileEvidence     []string       `json:"generated_file_evidence,omitempty"`
	ModuleCacheEvidence       []string       `json:"module_cache_evidence,omitempty"`
	BroadRepoSearchEvidence   []string       `json:"broad_repo_search_evidence,omitempty"`
	DirectSQLiteEvidence      []string       `json:"direct_sqlite_evidence,omitempty"`
	LegacyRunnerEvidence      []string       `json:"legacy_runner_evidence,omitempty"`
	UsageExposed              bool           `json:"usage_exposed"`
	InputTokens               *int           `json:"input_tokens,omitempty"`
	CachedInputTokens         *int           `json:"cached_input_tokens,omitempty"`
	NonCachedInputTokens      *int           `json:"non_cached_input_tokens,omitempty"`
	OutputTokens              *int           `json:"output_tokens,omitempty"`
	EventTypeCounts           map[string]int `json:"event_type_counts"`
	CommandMetricLimitations  string         `json:"command_metric_limitations"`
}
type verificationResult struct {
	Passed        bool     `json:"passed"`
	DatabasePass  bool     `json:"database_pass"`
	AssistantPass bool     `json:"assistant_pass"`
	Details       string   `json:"details"`
	Documents     []string `json:"documents,omitempty"`
}
type fixturePreflight struct {
	Name       string   `json:"name"`
	Passed     bool     `json:"passed"`
	Details    string   `json:"details,omitempty"`
	Documents  []string `json:"documents,omitempty"`
	SourcePath string   `json:"source_path,omitempty"`
	AssetPath  string   `json:"asset_path,omitempty"`
}
type productionGateSummary struct {
	Variant        string                    `json:"variant"`
	PassesGate     bool                      `json:"passes_gate"`
	Recommendation string                    `json:"recommendation"`
	Criteria       []productionGateCriterion `json:"criteria"`
}
type productionGateCriterion struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
}
type targetedLaneSummary struct {
	Lane                    string                           `json:"lane"`
	Decision                string                           `json:"decision"`
	PublicSurface           []string                         `json:"public_surface"`
	Promotion               string                           `json:"promotion"`
	ReleaseBlocking         bool                             `json:"release_blocking"`
	ScenarioClassifications []targetedScenarioClassification `json:"scenario_classifications"`
}
type targetedScenarioClassification struct {
	Variant               string  `json:"variant"`
	Scenario              string  `json:"scenario"`
	Status                string  `json:"status"`
	FailureClassification string  `json:"failure_classification"`
	EvidencePosture       string  `json:"evidence_posture"`
	ToolCalls             int     `json:"tool_calls"`
	CommandExecutions     int     `json:"command_executions"`
	AssistantCalls        int     `json:"assistant_calls"`
	WallSeconds           float64 `json:"wall_seconds"`
	PromptSpecificity     string  `json:"prompt_specificity,omitempty"`
	UX                    string  `json:"ux,omitempty"`
	Brittleness           string  `json:"brittleness,omitempty"`
	Retries               int     `json:"retries"`
	StepCount             int     `json:"step_count"`
	Latency               string  `json:"latency,omitempty"`
	GuidanceDependence    string  `json:"guidance_dependence,omitempty"`
	SafetyRisks           string  `json:"safety_risks,omitempty"`
	FixturePreflight      string  `json:"fixture_preflight,omitempty"`
}
type jobRunner func(context.Context, runConfig, evalJob, cacheConfig) jobResult
type codexEvent struct {
	Type     string          `json:"type"`
	ThreadID string          `json:"thread_id"`
	Item     json.RawMessage `json:"item"`
	Usage    *usage          `json:"usage"`
}
type usage struct {
	InputTokens        int           `json:"input_tokens"`
	OutputTokens       int           `json:"output_tokens"`
	CachedInputTokens  int           `json:"cached_input_tokens"`
	InputTokensDetails *usageDetails `json:"input_tokens_details"`
	PromptTokens       int           `json:"prompt_tokens"`
	CompletionTokens   int           `json:"completion_tokens"`
	PromptDetails      *usageDetails `json:"prompt_tokens_details"`
}
type usageDetails struct {
	CachedTokens int `json:"cached_tokens"`
}
type parsedTurn struct {
	metrics      metrics
	finalMessage string
	sessionID    string
	parseError   error
	parseSeconds float64
}
