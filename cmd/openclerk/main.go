package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/yazanabuashour/openclerk/internal/chronicler"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

var version string

const demoMarkerFile = ".openclerk-demo-root"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}

	switch args[0] {
	case "help", "-h", "--help":
		usage(stdout)
		return 0
	case "version", "--version":
		writeVersion(stdout)
		return 0
	case "capabilities":
		return runCapabilities(args[1:], stdout, stderr)
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "config":
		return runConfig(args[1:], stdin, stdout, stderr)
	case "module":
		return runModule(args[1:], stdin, stdout, stderr)
	case "document":
		return runDocument(args[1:], stdin, stdout, stderr)
	case "retrieval":
		return runRetrieval(args[1:], stdin, stdout, stderr)
	case "clerk":
		return runClerk(args[1:], stdout, stderr)
	case "demo":
		return runDemo(args[1:], stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "unknown openclerk command %q\n", args[0])
		usage(stderr)
		return 2
	}
}

type capabilitiesResult struct {
	SchemaVersion   string                 `json:"schema_version"`
	Product         string                 `json:"product"`
	Summary         string                 `json:"summary"`
	NorthStar       string                 `json:"north_star"`
	Principles      []string               `json:"principles"`
	Boundaries      []string               `json:"boundaries"`
	Domains         []capabilityDomain     `json:"domains"`
	ExtensionPoints []capabilityExtension  `json:"extension_points"`
	AgentHandoff    capabilityAgentHandoff `json:"agent_handoff"`
}

type capabilityDomain struct {
	Name      string             `json:"name"`
	Command   string             `json:"command"`
	Posture   string             `json:"posture"`
	Primitive []capabilityAction `json:"primitive_actions"`
	Workflow  []capabilityAction `json:"workflow_actions"`
}

type capabilityAction struct {
	Action   string `json:"action"`
	Purpose  string `json:"purpose"`
	Posture  string `json:"posture"`
	Handoff  string `json:"handoff,omitempty"`
	Requires string `json:"requires,omitempty"`
}

type capabilityExtension struct {
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	ManifestPath string `json:"manifest_path"`
	SkillPath    string `json:"skill_path"`
	Posture      string `json:"posture"`
}

type capabilityAgentHandoff struct {
	AnswerSummary        string   `json:"answer_summary"`
	SelectionGuidance    []string `json:"selection_guidance"`
	ValidationBoundaries []string `json:"validation_boundaries"`
	AuthorityLimits      []string `json:"authority_limits"`
}

func runCapabilities(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", args)
		_, _ = fmt.Fprintln(stderr, "usage: openclerk capabilities")
		return 2
	}
	if err := json.NewEncoder(stdout).Encode(buildCapabilitiesResult()); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode capabilities result: %v\n", err)
		return 1
	}
	return 0
}

func buildCapabilitiesResult() capabilitiesResult {
	return capabilitiesResult{
		SchemaVersion: "openclerk-capabilities.v1",
		Product:       "openclerk",
		Summary:       "Local-first knowledge-plane building blocks for agent assembly.",
		NorthStar:     "https://mitchellh.com/writing/building-block-economy",
		Principles: []string{
			"Expose high-quality, well-documented runner primitives that agents can compose.",
			"Promote repeated ceremonial workflows into compact runner actions with agent_handoff.",
			"Keep mainline behavior narrow, stable, local-first, citation-bearing, and approval-aware.",
			"Ship optional provider functionality as verified modules rather than hidden core defaults.",
		},
		Boundaries: []string{
			"Canonical markdown, citations, provenance, and projection freshness remain authority.",
			"Public read/fetch/inspect permission is separate from durable-write approval.",
			"No direct SQLite, raw vault inspection, unsupported transports, or source-built runner bypasses.",
			"Default retrieval search remains lexical; semantic ranking is explicit and module-gated.",
		},
		Domains: []capabilityDomain{
			{
				Name:    "config",
				Command: "openclerk config",
				Posture: "strict JSON runner for product configuration and persisted default profile preferences",
				Primitive: []capabilityAction{
					{Action: "inspect_config", Purpose: "inspect effective storage, profile, module, and git lifecycle configuration without exposing raw runtime_config keys", Posture: "read_only"},
					{Action: "configure_profile", Purpose: "persist default autonomy/profile preferences for document and retrieval requests", Posture: "configuration_write"},
					{Action: "clear_profile", Purpose: "clear persisted profile preferences and return to built-in defaults", Posture: "configuration_write"},
					{Action: "configure_vault_ignore_paths", Purpose: "replace additional vault-relative paths excluded from sync; pass an empty list to clear custom ignores", Posture: "configuration_write"},
				},
				Workflow: []capabilityAction{},
			},
			{
				Name:    "document",
				Command: "openclerk document",
				Posture: "strict JSON runner for validation, source intake, document mutation, placement planning, and document-side workflow blocks",
				Primitive: []capabilityAction{
					{Action: "validate", Purpose: "validate a candidate document without writing", Posture: "read_only"},
					{Action: "create_document", Purpose: "create an approved vault-relative markdown document", Posture: "durable_write_requires_approval"},
					{Action: "ingest_source_url", Purpose: "inspect, plan, create, or update public web, Markdown, or PDF source notes through the runner", Posture: "inspect_or_plan_read_only_or_approved_write"},
					{Action: "ingest_video_url", Purpose: "create or update video source notes from supplied transcripts", Posture: "approved_write_with_user_supplied_transcript"},
					{Action: "list_documents", Purpose: "list runner-visible documents", Posture: "read_only"},
					{Action: "get_document", Purpose: "read one runner-visible document by doc_id", Posture: "read_only"},
					{Action: "append_document", Purpose: "append approved content to an existing document", Posture: "durable_write_requires_approval"},
					{Action: "replace_section", Purpose: "replace an approved markdown section", Posture: "durable_write_requires_approval"},
					{Action: "plan_move_document", Purpose: "plan a safe vault-relative document move, rename, or candidate promotion before writing", Posture: "read_only"},
					{Action: "move_document", Purpose: "move an approved markdown document while preserving stable id and updating reported links", Posture: "durable_write_requires_approval_no_overwrite"},
					{Action: "rename_document", Purpose: "same-directory move convenience wrapper for path/title cleanup", Posture: "durable_write_requires_approval_no_overwrite"},
					{Action: "promote_candidate", Purpose: "promote a notes/candidates document into a canonical destination with duplicate checks", Posture: "durable_write_requires_approval_no_overwrite"},
					{Action: "plan_path_cleanup", Purpose: "propose title/taxonomy path cleanup candidates and optionally apply low-risk candidates under autonomous modes", Posture: "read_only_or_autonomous_apply_no_overwrite"},
					{Action: "resolve_paths", Purpose: "show configured database and vault paths", Posture: "read_only_diagnostic"},
					{Action: "inspect_layout", Purpose: "inspect configured OpenClerk layout", Posture: "read_only_diagnostic"},
				},
				Workflow: []capabilityAction{
					{Action: "compile_synthesis", Purpose: "create or update source-linked synthesis with freshness and source refs", Posture: "durable_write_requires_approval", Handoff: "compile_synthesis.agent_handoff"},
					{Action: "web_search_plan", Purpose: "plan harness-supplied public search-result intake before fetch/write", Posture: "read_only", Handoff: "web_search_plan.agent_handoff"},
					{Action: "artifact_candidate_plan", Purpose: "plan artifact path, title, body preview, tags, fields, duplicates, and create/ingest handoff", Posture: "read_only", Handoff: "artifact_candidate_plan.agent_handoff"},
					{Action: "git_lifecycle_report", Purpose: "report local storage status/history or create explicit local checkpoints", Posture: "read_only_or_explicit_checkpoint"},
					{Action: "validation_synthesis_report", Purpose: "create or update disposable validation synthesis with auditable source refs and freshness", Posture: "disposable_validation_write", Handoff: "validation_synthesis.agent_handoff"},
				},
			},
			{
				Name:    "retrieval",
				Command: "openclerk retrieval",
				Posture: "strict JSON runner for citation-bearing lookup, provenance, projections, and retrieval-side workflow blocks",
				Primitive: []capabilityAction{
					{Action: "validate", Purpose: "validate a retrieval request without querying storage", Posture: "read_only"},
					{Action: "search", Purpose: "lexical citation-bearing document search", Posture: "read_only"},
					{Action: "document_links", Purpose: "inspect markdown relationship links", Posture: "read_only"},
					{Action: "graph_neighborhood", Purpose: "inspect nearby relationship graph evidence", Posture: "read_only"},
					{Action: "records_lookup", Purpose: "lookup promoted record projections", Posture: "read_only"},
					{Action: "record_entity", Purpose: "read a promoted record entity projection", Posture: "read_only"},
					{Action: "services_lookup", Purpose: "lookup promoted service projections", Posture: "read_only"},
					{Action: "service_record", Purpose: "read a promoted service record projection", Posture: "read_only"},
					{Action: "decisions_lookup", Purpose: "lookup promoted decision projections", Posture: "read_only"},
					{Action: "decision_record", Purpose: "read a promoted decision record projection", Posture: "read_only"},
					{Action: "provenance_events", Purpose: "inspect derivation and write provenance", Posture: "read_only"},
					{Action: "projection_states", Purpose: "inspect projection freshness state", Posture: "read_only"},
				},
				Workflow: []capabilityAction{
					{Action: "audit_contradictions", Purpose: "plan or repair existing contradiction-audit synthesis with source authority visible", Posture: "read_only_or_existing_target_repair"},
					{Action: "source_discovery_report", Purpose: "find representative runner-visible sources and sanitized source-category summaries", Posture: "read_only", Handoff: "source_discovery.agent_handoff"},
					{Action: "source_audit_report", Purpose: "explain source-sensitive gaps or repair an existing synthesis target", Posture: "read_only_or_existing_target_repair", Handoff: "source_audit.agent_handoff"},
					{Action: "evidence_bundle_report", Purpose: "package citations, provenance, projection freshness, and authority limits", Posture: "read_only", Handoff: "evidence_bundle.agent_handoff"},
					{Action: "decision_lookup_report", Purpose: "lookup decision-like evidence across decisions, records, search, provenance, and projection freshness", Posture: "read_only", Handoff: "decision_lookup.agent_handoff"},
					{Action: "duplicate_candidate_report", Purpose: "choose update-versus-new evidence before durable writes", Posture: "read_only", Handoff: "duplicate_candidate.agent_handoff"},
					{Action: "workflow_guide_report", Purpose: "select the natural runner surface for an intent", Posture: "read_only", Handoff: "workflow_guide.agent_handoff"},
					{Action: "memory_router_recall_report", Purpose: "package routine memory/router recall evidence without memory transports", Posture: "read_only", Handoff: "memory_router_recall.agent_handoff"},
					{Action: "structured_store_report", Purpose: "review structured-data and canonical-store evidence", Posture: "read_only", Handoff: "structured_store.agent_handoff"},
					{Action: "hybrid_retrieval_report", Purpose: "review lexical baseline and hybrid/vector candidate boundaries", Posture: "read_only", Handoff: "hybrid_retrieval.agent_handoff"},
					{Action: "graph_context_report", Purpose: "package relationship graph context with canonical markdown authority and freshness", Posture: "read_only", Handoff: "graph_context.agent_handoff"},
					{Action: "graph_relationship_report", Purpose: "package relationship paths, direct-vs-derived evidence, typed candidates, and limited graph audits from canonical markdown authority", Posture: "read_only", Handoff: "graph_relationship.agent_handoff"},
					{Action: "graph_relationship_maintenance_plan", Purpose: "plan approval-gated canonical markdown relationship maintenance from relationship report evidence", Posture: "read_only", Handoff: "graph_relationship_maintenance.agent_handoff"},
					{Action: "semantic_search", Purpose: "run explicit citation-bearing semantic search through a verified provider module", Posture: "module_gated_read_only", Handoff: "semantic_search.agent_handoff", Requires: "installed enabled embedding provider module"},
					{Action: "retrieval_eval_capture", Purpose: "append an explicit local-only sanitized retrieval eval case for replay", Posture: "opt_in_local_eval_artifact", Handoff: "retrieval_eval_capture.agent_handoff"},
					{Action: "retrieval_eval_replay", Purpose: "replay sanitized retrieval eval cases and report Jaccard, top-1, and latency metrics", Posture: "read_only_local_eval_replay", Handoff: "retrieval_eval_replay.agent_handoff"},
					{Action: "search_diagnostics_report", Purpose: "recommend search versus explicit semantic_search with tuning and module posture visibility", Posture: "read_only", Handoff: "search_diagnostics.agent_handoff"},
					{Action: "maintenance_report", Purpose: "package layout, projection, relationship, duplicate, module, and git lifecycle maintenance posture without repair", Posture: "read_only", Handoff: "maintenance.agent_handoff"},
				},
			},
			{
				Name:    "module",
				Command: "openclerk module",
				Posture: "strict JSON runner for optional verified provider building blocks with redacted runtime configuration",
				Primitive: []capabilityAction{
					{Action: "install_module", Purpose: "verify and register a provider module manifest", Posture: "configuration_write"},
					{Action: "configure_module", Purpose: "enable, disable, or update redacted provider defaults", Posture: "configuration_write"},
					{Action: "remove_module", Purpose: "remove OpenClerk module registration without deleting unrelated provider state", Posture: "configuration_write"},
					{Action: "list_modules", Purpose: "list verified module/provider state", Posture: "read_only"},
				},
				Workflow: []capabilityAction{},
			},
			{
				Name:    "clerk",
				Command: "openclerk clerk",
				Posture: "Chronicler Lite over Core; read-only session-to-repo-knowledge planning with no durable writes",
				Primitive: []capabilityAction{
					{Action: "run --once", Purpose: "plan explicit local inbox candidates and task context packs without durable writes", Posture: "read_only_planned_no_write", Handoff: "openclerk-clerk.v1"},
					{Action: "inbox_scan", Purpose: "plan explicit local inbox candidates without durable writes", Posture: "read_only_planned_no_write", Handoff: "openclerk-clerk.v1"},
					{Action: "context_pack", Purpose: "package task context, must-read documents, decisions, and citations without durable writes", Posture: "read_only_planned_no_write", Handoff: "openclerk-clerk.v1"},
				},
				Workflow: []capabilityAction{},
			},
		},
		ExtensionPoints: []capabilityExtension{
			{Name: "ollama-embeddings", Kind: "embedding_provider", ManifestPath: "modules/ollama-embeddings/module.json", SkillPath: "modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md", Posture: "local_first_optional_module"},
			{Name: "gemini-embeddings", Kind: "embedding_provider", ManifestPath: "modules/gemini-embeddings/module.json", SkillPath: "modules/gemini-embeddings/skill/gemini-embeddings/SKILL.md", Posture: "explicit_remote_provider_opt_in"},
			{Name: "tesseract-ocr", Kind: "ocr_provider", ManifestPath: "modules/tesseract-ocr/module.json", SkillPath: "modules/tesseract-ocr/skill/tesseract-ocr/SKILL.md", Posture: "local_ocr_review_optional_module"},
		},
		AgentHandoff: capabilityAgentHandoff{
			AnswerSummary: "Use OpenClerk as a narrow mainline runner plus composable document, retrieval, and module building blocks.",
			SelectionGuidance: []string{
				"Use config for persisted profile preferences and module for provider settings.",
				"Start with promoted workflow actions when they match the user intent.",
				"Use primitives for explicit manual work, advanced inspection, or after a workflow action rejects.",
				"Use optional modules only after install/list verifies the module boundary.",
			},
			ValidationBoundaries: []string{
				"Capabilities are a static discovery manifest, not evidence from a user vault.",
				"Run the selected document, retrieval, or module action for task-specific results.",
			},
			AuthorityLimits: []string{
				"Do not answer source-sensitive user questions from the capabilities manifest alone.",
				"Use returned citations, provenance, projection freshness, validation boundaries, and authority limits from the selected runner action.",
			},
		},
	}
}

func runInit(args []string, stdout io.Writer, stderr io.Writer) int {
	config, vaultRoot, ok := parseInitConfig(args, stderr)
	if !ok {
		return 2
	}
	paths, err := runclient.InitializePaths(config, vaultRoot)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "initialize OpenClerk paths: %v\n", err)
		return 1
	}
	result := struct {
		Paths   runner.Paths `json:"paths"`
		Summary string       `json:"summary"`
	}{
		Paths: runner.Paths{
			DatabasePath: paths.DatabasePath,
			VaultRoot:    paths.VaultRoot,
		},
		Summary: "initialized OpenClerk paths",
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode init result: %v\n", err)
		return 1
	}
	return 0
}

func runConfig(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if wantsSubcommandHelp(args) {
		configUsage(stdout)
		return 0
	}
	config, ok := parseConfig("config", args, stderr)
	if !ok {
		return 2
	}
	var request runner.ConfigTaskRequest
	if err := decodeRequest(stdin, &request); err != nil {
		_, _ = fmt.Fprintf(stderr, "decode config request: %v\n", err)
		return 1
	}
	result, err := runner.RunConfigTask(context.Background(), config, request)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "run config task: %v\n", err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode config result: %v\n", err)
		return 1
	}
	return 0
}

type moduleTaskRequest struct {
	Action   string                                 `json:"action"`
	Module   runclient.SemanticModuleInstallInput   `json:"module,omitempty"`
	Config   runclient.SemanticModuleConfigureInput `json:"config,omitempty"`
	Provider string                                 `json:"provider,omitempty"`
}

type moduleTaskResult struct {
	Rejected        bool                             `json:"rejected"`
	RejectionReason string                           `json:"rejection_reason,omitempty"`
	Module          *runclient.SemanticModuleConfig  `json:"module,omitempty"`
	Modules         []runclient.SemanticModuleConfig `json:"modules,omitempty"`
	Summary         string                           `json:"summary"`
}

func runModule(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if wantsSubcommandHelp(args) {
		moduleUsage(stdout)
		return 0
	}
	config, ok := parseConfig("module", args, stderr)
	if !ok {
		return 2
	}
	var request moduleTaskRequest
	if err := decodeRequest(stdin, &request); err != nil {
		_, _ = fmt.Fprintf(stderr, "decode module request: %v\n", err)
		return 1
	}
	result, err := runModuleTask(context.Background(), config, request)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "run module task: %v\n", err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode module result: %v\n", err)
		return 1
	}
	return 0
}

func runModuleTask(ctx context.Context, config runclient.Config, request moduleTaskRequest) (moduleTaskResult, error) {
	action := request.Action
	if action == "" {
		action = "list_modules"
	}
	switch action {
	case "install_module":
		module, err := runclient.InstallSemanticModule(ctx, config, request.Module)
		if err != nil {
			return moduleRejected(err), nil
		}
		return moduleTaskResult{Module: &module, Summary: fmt.Sprintf("installed %s module", module.Kind)}, nil
	case "configure_module":
		module, err := runclient.ConfigureSemanticModule(ctx, config, request.Config)
		if err != nil {
			return moduleRejected(err), nil
		}
		return moduleTaskResult{Module: &module, Summary: fmt.Sprintf("configured %s module", module.Kind)}, nil
	case "remove_module":
		module, err := runclient.RemoveSemanticModule(ctx, config, request.Provider)
		if err != nil {
			return moduleRejected(err), nil
		}
		return moduleTaskResult{Module: &module, Summary: fmt.Sprintf("removed %s module", module.Kind)}, nil
	case "list_modules":
		modules, err := runclient.ListSemanticModules(ctx, config)
		if err != nil {
			return moduleTaskResult{}, err
		}
		return moduleTaskResult{Modules: modules, Summary: fmt.Sprintf("returned %d semantic modules", len(modules))}, nil
	default:
		return moduleTaskResult{Rejected: true, RejectionReason: fmt.Sprintf("unsupported module action %q", action), Summary: fmt.Sprintf("unsupported module action %q", action)}, nil
	}
}

func moduleRejected(err error) moduleTaskResult {
	return moduleTaskResult{Rejected: true, RejectionReason: err.Error(), Summary: err.Error()}
}

func runDocument(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if wantsSubcommandHelp(args) {
		documentUsage(stdout)
		return 0
	}
	config, ok := parseConfig("document", args, stderr)
	if !ok {
		return 2
	}
	var request runner.DocumentTaskRequest
	if err := decodeRequest(stdin, &request); err != nil {
		_, _ = fmt.Fprintf(stderr, "decode document request: %v\n", err)
		return 1
	}
	result, err := runner.RunDocumentTask(context.Background(), config, request)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "run document task: %v\n", err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode document result: %v\n", err)
		return 1
	}
	return 0
}

func runRetrieval(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if wantsSubcommandHelp(args) {
		retrievalUsage(stdout)
		return 0
	}
	config, ok := parseConfig("retrieval", args, stderr)
	if !ok {
		return 2
	}
	var request runner.RetrievalTaskRequest
	if err := decodeRequest(stdin, &request); err != nil {
		_, _ = fmt.Fprintf(stderr, "decode retrieval request: %v\n", err)
		return 1
	}
	result, err := runner.RunRetrievalTask(context.Background(), config, request)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "run retrieval task: %v\n", err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode retrieval result: %v\n", err)
		return 1
	}
	return 0
}

func runClerk(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || isHelpArg(args[0]) {
		clerkUsage(stdout)
		return 0
	}
	switch args[0] {
	case "run":
		return runClerkRun(args[1:], stdout, stderr)
	case "inbox_scan":
		return runClerkInboxScan(args[1:], stdout, stderr)
	case "context_pack":
		return runClerkContextPack(args[1:], stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "unknown openclerk clerk command %q\n", args[0])
		clerkUsage(stderr)
		return 2
	}
}

func runClerkRun(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 1 && isHelpArg(args[0]) {
		clerkRunUsage(stdout)
		return 0
	}
	config, request, once, ok := parseClerkRunConfig(args, stderr)
	if !ok {
		return 2
	}
	if !once {
		_, _ = fmt.Fprintln(stderr, "usage: openclerk clerk run --once [--db path] [--inbox-path path] [--task text] [--query text] [--path-prefix prefix] [--limit n]")
		_, _ = fmt.Fprintln(stderr, "Chronicler Lite only supports read-only --once runs; daemon/watch mode and autonomous background improvement are shelved.")
		return 2
	}
	result, err := chronicler.RunOnce(context.Background(), config, request)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "run clerk task: %v\n", err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode clerk result: %v\n", err)
		return 1
	}
	return 0
}

func runClerkInboxScan(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 1 && isHelpArg(args[0]) {
		clerkInboxScanUsage(stdout)
		return 0
	}
	config, request, ok := parseClerkInboxScanConfig(args, stderr)
	if !ok {
		return 2
	}
	result, err := chronicler.RunInboxScan(context.Background(), config, request)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "run clerk inbox scan: %v\n", err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode clerk inbox scan result: %v\n", err)
		return 1
	}
	return 0
}

func runClerkContextPack(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 1 && isHelpArg(args[0]) {
		clerkContextPackUsage(stdout)
		return 0
	}
	config, request, ok := parseClerkContextPackConfig(args, stderr)
	if !ok {
		return 2
	}
	result, err := chronicler.RunContextPack(context.Background(), config, request)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "run clerk context pack: %v\n", err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode clerk context pack result: %v\n", err)
		return 1
	}
	return 0
}

type demoResult struct {
	SchemaVersion       string             `json:"schema_version"`
	Action              string             `json:"action"`
	Question            string             `json:"question,omitempty"`
	Summary             string             `json:"summary"`
	DemoRoot            string             `json:"demo_root"`
	DatabasePath        string             `json:"database_path"`
	VaultRoot           string             `json:"vault_root"`
	SourcePath          string             `json:"source_path,omitempty"`
	SynthesisPath       string             `json:"synthesis_path,omitempty"`
	SynthesisDocID      string             `json:"synthesis_doc_id,omitempty"`
	ProjectionFreshness []demoProjection   `json:"projection_freshness,omitempty"`
	Citations           []demoCitation     `json:"citations,omitempty"`
	RepairRequest       *demoRepairRequest `json:"repair_request,omitempty"`
	Next                []string           `json:"next,omitempty"`
}

type demoProjection struct {
	Path            string `json:"path"`
	Freshness       string `json:"freshness"`
	FreshnessReason string `json:"freshness_reason,omitempty"`
	StaleSourceRefs string `json:"stale_source_refs,omitempty"`
}

type demoCitation struct {
	Path      string `json:"path"`
	Heading   string `json:"heading,omitempty"`
	LineStart int    `json:"line_start,omitempty"`
	LineEnd   int    `json:"line_end,omitempty"`
}

type demoRepairRequest struct {
	Action    string                       `json:"action"`
	Synthesis runner.CompileSynthesisInput `json:"synthesis"`
}

func runDemo(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || isHelpArg(args[0]) {
		demoUsage(stdout)
		return 0
	}
	switch args[0] {
	case "init":
		return runDemoInit(args[1:], stdout, stderr)
	case "ask":
		return runDemoAsk(args[1:], stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "unknown openclerk demo command %q\n", args[0])
		demoUsage(stderr)
		return 2
	}
}

func runDemoInit(args []string, stdout io.Writer, stderr io.Writer) int {
	if wantsSubcommandHelp(args) {
		demoInitUsage(stdout)
		return 0
	}
	root, _, ok := parseDemoConfig("init", args, stderr)
	if !ok {
		return 2
	}
	result, err := seedDemo(context.Background(), root)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "run demo init: %v\n", err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode demo result: %v\n", err)
		return 1
	}
	return 0
}

func runDemoAsk(args []string, stdout io.Writer, stderr io.Writer) int {
	if wantsSubcommandHelp(args) {
		demoAskUsage(stdout)
		return 0
	}
	root, remaining, ok := parseDemoConfig("ask", args, stderr)
	if !ok {
		return 2
	}
	question := strings.TrimSpace(strings.Join(remaining, " "))
	if question == "" {
		question = "what changed and what is stale?"
	}
	result, err := answerDemo(context.Background(), root, question)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "run demo ask: %v\n", err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode demo answer: %v\n", err)
		return 1
	}
	return 0
}

func parseDemoConfig(name string, args []string, stderr io.Writer) (string, []string, bool) {
	fs := flag.NewFlagSet("openclerk demo "+name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", defaultDemoRoot(), "isolated demo root")
	if err := fs.Parse(args); err != nil {
		return "", nil, false
	}
	if strings.TrimSpace(*root) == "" {
		_, _ = fmt.Fprintln(stderr, "demo root is required")
		return "", nil, false
	}
	absRoot, err := filepath.Abs(*root)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "resolve demo root: %v\n", err)
		return "", nil, false
	}
	return absRoot, fs.Args(), true
}

func defaultDemoRoot() string {
	base, err := os.UserCacheDir()
	if err != nil || strings.TrimSpace(base) == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "openclerk", "demo")
}

func seedDemo(ctx context.Context, root string) (demoResult, error) {
	paths, err := resetDemoPaths(root)
	if err != nil {
		return demoResult{}, err
	}
	config := runclient.Config{DatabasePath: paths.DatabasePath}
	if _, err := runclient.InitializePaths(config, paths.VaultRoot); err != nil {
		return demoResult{}, err
	}

	source, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  "sources/plan.md",
			Title: "Plan Source",
			Body:  "# Plan Source\n\n## Summary\nAcme plan includes 10 seats.\n",
		},
	})
	if err != nil {
		return demoResult{}, err
	}
	if source.Document == nil {
		return demoResult{}, fmt.Errorf("demo source was not created")
	}

	synthesis, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:          "synthesis/account-memory.md",
			Title:         "Account Memory",
			SourceRefs:    []string{"sources/plan.md"},
			BodyFacts:     []string{"Acme plan includes 10 seats."},
			FreshnessNote: "Checked current source evidence through the demo.",
			Mode:          "create_or_update",
		},
	})
	if err != nil {
		return demoResult{}, err
	}
	if synthesis.CompileSynthesis == nil {
		return demoResult{}, fmt.Errorf("demo synthesis was not created")
	}

	time.Sleep(time.Millisecond)
	if _, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   source.Document.DocID,
		Heading: "Summary",
		Content: "Acme plan now includes 25 seats. The old 10-seat synthesis is stale.",
	}); err != nil {
		return demoResult{}, err
	}

	answer, err := buildDemoAnswer(ctx, root, config, paths, "what changed and what is stale?", synthesis.CompileSynthesis.DocumentID)
	if err != nil {
		return demoResult{}, err
	}
	answer.Action = "init"
	answer.Summary = "seeded demo vault and detected stale synthesis"
	answer.Next = []string{`openclerk demo ask "what changed and what is stale?"`}
	return answer, nil
}

func resetDemoPaths(root string) (runclient.Paths, error) {
	root = filepath.Clean(root)
	paths := runclient.Paths{
		DatabasePath: filepath.Join(root, "openclerk.sqlite"),
		VaultRoot:    filepath.Join(root, "vault"),
	}
	if err := guardDemoRoot(root); err != nil {
		return runclient.Paths{}, err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return runclient.Paths{}, fmt.Errorf("create demo root: %w", err)
	}
	for _, path := range []string{
		paths.VaultRoot,
		paths.DatabasePath,
		paths.DatabasePath + "-shm",
		paths.DatabasePath + "-wal",
		paths.DatabasePath + ".runner-write.lock",
		paths.DatabasePath + ".runtime-config.lock",
	} {
		if err := os.RemoveAll(path); err != nil {
			return runclient.Paths{}, fmt.Errorf("reset demo path %q: %w", path, err)
		}
	}
	if err := os.MkdirAll(paths.VaultRoot, 0o755); err != nil {
		return runclient.Paths{}, fmt.Errorf("create demo vault: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, demoMarkerFile), []byte("openclerk demo root\n"), 0o644); err != nil {
		return runclient.Paths{}, fmt.Errorf("write demo marker: %w", err)
	}
	return paths, nil
}

func guardDemoRoot(root string) error {
	info, err := os.Stat(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect demo root: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("demo root must be a directory")
	}
	if _, err := os.Stat(filepath.Join(root, demoMarkerFile)); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect demo marker: %w", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("inspect demo root entries: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("demo root %q is not empty and is not marked as an OpenClerk demo root", root)
	}
	return nil
}

func answerDemo(ctx context.Context, root string, question string) (demoResult, error) {
	paths := runclient.Paths{
		DatabasePath: filepath.Join(root, "openclerk.sqlite"),
		VaultRoot:    filepath.Join(root, "vault"),
	}
	config := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "synthesis/", Limit: 10},
	})
	if err != nil {
		return demoResult{}, fmt.Errorf("read demo vault; run openclerk demo init first: %w", err)
	}
	if len(list.Documents) == 0 {
		return demoResult{}, fmt.Errorf("demo synthesis missing; run openclerk demo init first")
	}
	return buildDemoAnswer(ctx, root, config, paths, question, list.Documents[0].DocID)
}

func buildDemoAnswer(ctx context.Context, root string, config runclient.Config, paths runclient.Paths, question string, synthesisDocID string) (demoResult, error) {
	projections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      synthesisDocID,
			Limit:      10,
		},
	})
	if err != nil {
		return demoResult{}, err
	}
	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: "Acme plan seats stale", Limit: 5},
	})
	if err != nil {
		return demoResult{}, err
	}

	summary := "sources/plan.md changed after synthesis/account-memory.md; synthesis is stale and should be repaired through compile_synthesis."
	return demoResult{
		SchemaVersion:       "openclerk-demo.v1",
		Action:              "ask",
		Question:            question,
		Summary:             summary,
		DemoRoot:            root,
		DatabasePath:        paths.DatabasePath,
		VaultRoot:           paths.VaultRoot,
		SourcePath:          "sources/plan.md",
		SynthesisPath:       "synthesis/account-memory.md",
		SynthesisDocID:      synthesisDocID,
		ProjectionFreshness: demoProjections(projections.Projections),
		Citations:           demoCitations(search.Search),
		RepairRequest: &demoRepairRequest{
			Action: runner.DocumentTaskActionCompileSynthesis,
			Synthesis: runner.CompileSynthesisInput{
				Path:          "synthesis/account-memory.md",
				Title:         "Account Memory",
				SourceRefs:    []string{"sources/plan.md"},
				BodyFacts:     []string{"Acme plan now includes 25 seats."},
				FreshnessNote: "Repair stale synthesis after source change.",
				Mode:          "create_or_update",
			},
		},
	}, nil
}

func demoProjections(list *runner.ProjectionStateList) []demoProjection {
	if list == nil {
		return nil
	}
	result := make([]demoProjection, 0, len(list.Projections))
	for _, projection := range list.Projections {
		result = append(result, demoProjection{
			Path:            projection.Details["synthesis_path"],
			Freshness:       projection.Freshness,
			FreshnessReason: projection.Details["freshness_reason"],
			StaleSourceRefs: projection.Details["stale_source_refs"],
		})
	}
	return result
}

func demoCitations(search *runner.SearchResult) []demoCitation {
	if search == nil {
		return nil
	}
	result := []demoCitation{}
	for _, hit := range search.Hits {
		for _, citation := range hit.Citations {
			result = append(result, demoCitation{
				Path:      citation.Path,
				Heading:   citation.Heading,
				LineStart: citation.LineStart,
				LineEnd:   citation.LineEnd,
			})
			if len(result) == 5 {
				return result
			}
		}
	}
	return result
}

func parseConfig(name string, args []string, stderr io.Writer) (runclient.Config, bool) {
	fs := flag.NewFlagSet("openclerk "+name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	databasePath := fs.String("db", "", "OpenClerk SQLite database path")
	gitCheckpoints := fs.Bool("git-checkpoints", false, "enable explicit local git checkpoint writes")
	if err := fs.Parse(args); err != nil {
		return runclient.Config{}, false
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", fs.Args())
		return runclient.Config{}, false
	}
	return runclient.Config{
		DatabasePath:   *databasePath,
		GitCheckpoints: *gitCheckpoints,
	}, true
}

func wantsSubcommandHelp(args []string) bool {
	for _, arg := range args {
		if isHelpArg(arg) {
			return true
		}
	}
	return false
}

func isHelpArg(arg string) bool {
	switch arg {
	case "help", "-h", "--help":
		return true
	default:
		return false
	}
}

func parseInitConfig(args []string, stderr io.Writer) (runclient.Config, string, bool) {
	fs := flag.NewFlagSet("openclerk init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	databasePath := fs.String("db", "", "OpenClerk SQLite database path")
	vaultRoot := fs.String("vault-root", "", "OpenClerk vault root")
	if err := fs.Parse(args); err != nil {
		return runclient.Config{}, "", false
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", fs.Args())
		return runclient.Config{}, "", false
	}
	return runclient.Config{DatabasePath: *databasePath}, *vaultRoot, true
}

type stringListFlag []string

func (f *stringListFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *stringListFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func parseClerkRunConfig(args []string, stderr io.Writer) (runclient.Config, chronicler.RunRequest, bool, bool) {
	fs := flag.NewFlagSet("openclerk clerk run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	databasePath := fs.String("db", "", "OpenClerk SQLite database path")
	once := fs.Bool("once", false, "run one read-only Chronicler pass")
	task := fs.String("task", "", "user task text for context-pack generation")
	query := fs.String("query", "", "optional retrieval query for context-pack generation")
	pathPrefix := fs.String("path-prefix", "", "optional vault-relative path prefix for context-pack retrieval")
	limit := fs.Int("limit", 0, "optional result limit")
	var inboxPaths stringListFlag
	fs.Var(&inboxPaths, "inbox-path", "explicit local inbox file or directory; may be repeated")
	fs.Var(&inboxPaths, "inbox", "alias for --inbox-path")
	if err := fs.Parse(args); err != nil {
		return runclient.Config{}, chronicler.RunRequest{}, false, false
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", fs.Args())
		return runclient.Config{}, chronicler.RunRequest{}, false, false
	}
	return runclient.Config{DatabasePath: *databasePath}, chronicler.RunRequest{
		InboxPaths: append([]string(nil), inboxPaths...),
		Task:       *task,
		Query:      *query,
		PathPrefix: *pathPrefix,
		Limit:      *limit,
	}, *once, true
}

func parseClerkInboxScanConfig(args []string, stderr io.Writer) (runclient.Config, chronicler.RunRequest, bool) {
	fs := flag.NewFlagSet("openclerk clerk inbox_scan", flag.ContinueOnError)
	fs.SetOutput(stderr)
	databasePath := fs.String("db", "", "OpenClerk SQLite database path")
	limit := fs.Int("limit", 0, "optional result limit")
	var inboxPaths stringListFlag
	fs.Var(&inboxPaths, "inbox-path", "explicit local inbox file or directory; may be repeated")
	fs.Var(&inboxPaths, "inbox", "alias for --inbox-path")
	if err := fs.Parse(args); err != nil {
		return runclient.Config{}, chronicler.RunRequest{}, false
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", fs.Args())
		return runclient.Config{}, chronicler.RunRequest{}, false
	}
	return runclient.Config{DatabasePath: *databasePath}, chronicler.RunRequest{
		InboxPaths: append([]string(nil), inboxPaths...),
		Limit:      *limit,
	}, true
}

func parseClerkContextPackConfig(args []string, stderr io.Writer) (runclient.Config, chronicler.RunRequest, bool) {
	fs := flag.NewFlagSet("openclerk clerk context_pack", flag.ContinueOnError)
	fs.SetOutput(stderr)
	databasePath := fs.String("db", "", "OpenClerk SQLite database path")
	task := fs.String("task", "", "user task text for context-pack generation")
	query := fs.String("query", "", "optional retrieval query for context-pack generation")
	pathPrefix := fs.String("path-prefix", "", "optional vault-relative path prefix for context-pack retrieval")
	limit := fs.Int("limit", 0, "optional result limit")
	if err := fs.Parse(args); err != nil {
		return runclient.Config{}, chronicler.RunRequest{}, false
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", fs.Args())
		return runclient.Config{}, chronicler.RunRequest{}, false
	}
	return runclient.Config{DatabasePath: *databasePath}, chronicler.RunRequest{
		Task:       *task,
		Query:      *query,
		PathPrefix: *pathPrefix,
		Limit:      *limit,
	}, true
}

func decodeRequest[T any](stdin io.Reader, request *T) error {
	decoder := json.NewDecoder(stdin)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(request); err != nil {
		return err
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values are not supported")
		}
		return err
	}
	return nil
}

func writeVersion(w io.Writer) {
	info, ok := readBuildInfo()
	_, _ = fmt.Fprintf(w, "openclerk %s\n", resolvedVersion(version, info, ok))
}

func readBuildInfo() (*debug.BuildInfo, bool) {
	return debug.ReadBuildInfo()
}

func resolvedVersion(linkerVersion string, info *debug.BuildInfo, ok bool) string {
	if linkerVersion != "" {
		return linkerVersion
	}
	if ok && info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func usage(stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, "usage: openclerk <version|capabilities|init|config|module|document|retrieval|clerk|demo> [--db path]")
	_, _ = fmt.Fprintln(stderr, "       openclerk init [--db path] [--vault-root path]")
	_, _ = fmt.Fprintln(stderr, "       openclerk capabilities")
	_, _ = fmt.Fprintln(stderr, "       openclerk config --help")
	_, _ = fmt.Fprintln(stderr, "       openclerk module --help")
	_, _ = fmt.Fprintln(stderr, "       openclerk document --help")
	_, _ = fmt.Fprintln(stderr, "       openclerk retrieval --help")
	_, _ = fmt.Fprintln(stderr, "       openclerk clerk --help")
	_, _ = fmt.Fprintln(stderr, "       openclerk demo --help")
	_, _ = fmt.Fprintln(stderr, "document/retrieval read strict JSON from stdin and use configured paths by default; pass --db only for an explicit dataset.")
	_, _ = fmt.Fprintln(stderr, "promoted workflow actions: compile_synthesis, validation_synthesis_report, ingest_source_url inspect/plan, web_search_plan, artifact_candidate_plan, git_lifecycle_report, source_discovery_report, source_audit_report, evidence_bundle_report, decision_lookup_report, duplicate_candidate_report, workflow_guide_report, memory_router_recall_report, structured_store_report, hybrid_retrieval_report, graph_context_report, graph_relationship_report, graph_relationship_maintenance_plan, semantic_search, retrieval_eval_capture, retrieval_eval_replay, search_diagnostics_report, maintenance_report, clerk run --once, clerk inbox_scan, clerk context_pack")
}

func demoUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk demo <init|ask> [--root path]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Creates and queries an isolated demo vault/database under the demo root.")
	_, _ = fmt.Fprintln(w, `  openclerk demo init`)
	_, _ = fmt.Fprintln(w, `  openclerk demo ask "what changed and what is stale?"`)
	_, _ = fmt.Fprintln(w, "The demo seeds one source note, one synthesis page, updates the source, reports stale projection freshness, and returns a compile_synthesis repair request.")
}

func demoInitUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk demo init [--root path]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Seeds an isolated demo vault/database, then reports stale synthesis evidence.")
	_, _ = fmt.Fprintln(w, "The root must be empty or already marked as an OpenClerk demo root.")
}

func demoAskUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, `usage: openclerk demo ask [--root path] ["question"]`)
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Reads the isolated demo vault and returns stale projection evidence plus a compile_synthesis repair request.")
}

func configUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk config [--db path] [--git-checkpoints] < request.json")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Reads one strict JSON object from stdin and writes one JSON result.")
	_, _ = fmt.Fprintln(w, "Inspects effective storage/profile/module/git lifecycle config and manages persisted product/profile configuration.")
	_, _ = fmt.Fprintln(w, `  inspect: {"action":"inspect_config"}`)
	_, _ = fmt.Fprintln(w, `  configure profile: {"action":"configure_profile","profile":{"approval_mode":"approve_write","drafting_mode":"suggest_fields","write_target_mode":"create_or_update","citation_mode":"balanced","privacy_mode":"allow_paths","audience_mode":"technical"}}`)
	_, _ = fmt.Fprintln(w, `  clear profile: {"action":"clear_profile"}`)
	_, _ = fmt.Fprintln(w, `  configure vault ignores: {"action":"configure_vault_ignore_paths","vault_ignore_paths":["scratch/","private/drafts/"]}; pass [] to clear custom ignores`)
	_, _ = fmt.Fprintln(w, "inspect_config returns storage, profile, modules, and git_lifecycle summaries; checkpoint_persistence is unsupported by design.")
	_, _ = fmt.Fprintln(w, "Storage is read-only here; initialize or intentionally rebind vault_root with openclerk init --vault-root.")
	_, _ = fmt.Fprintln(w, "Request-level document/retrieval autonomy fields override persisted profile defaults field-by-field.")
	_, _ = fmt.Fprintln(w, "Provider/module settings remain under openclerk module configure_module.")
}

func moduleUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk module [--db path] < request.json")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Reads one strict JSON object from stdin and writes one JSON result.")
	_, _ = fmt.Fprintln(w, "Manages optional embedding and OCR modules through redacted runtime_config state.")
	_, _ = fmt.Fprintln(w, "Semantic modules resolve semantic-retrieval-adapter at install time, pin its digest, and reject command_args.")
	_, _ = fmt.Fprintln(w, `  install Ollama: {"action":"install_module","module":{"provider":"ollama","manifest_path":"modules/ollama-embeddings/module.json","command":"semantic-retrieval-adapter","provider_config":{"embedding_model":"embeddinggemma","ollama_url":"http://localhost:11434"}}}`)
	_, _ = fmt.Fprintln(w, `  install Gemini: {"action":"install_module","module":{"provider":"gemini","manifest_path":"modules/gemini-embeddings/module.json","command":"semantic-retrieval-adapter","provider_config":{"embedding_model":"gemini-embedding-001","gemini_api_base":"https://generativelanguage.googleapis.com/v1beta","embedding_output_dimensions":"3072"}}}`)
	_, _ = fmt.Fprintln(w, `  install Tesseract OCR: {"action":"install_module","module":{"kind":"ocr_provider","provider":"tesseract","manifest_path":"modules/tesseract-ocr/module.json","command":"tesseract","provider_config":{"ocrmypdf_command":"ocrmypdf","language":"eng"}}}`)
	_, _ = fmt.Fprintln(w, `  configure: {"action":"configure_module","config":{"provider":"ollama","enabled":true,"provider_config":{"embedding_model":"embeddinggemma"}}}`)
	_, _ = fmt.Fprintln(w, `  remove/list: {"action":"remove_module","provider":"ollama"} | {"action":"list_modules"}`)
	_, _ = fmt.Fprintln(w, "Gemini stores only redacted provider config and uses runtime_config:GEMINI_API_KEY from the configured database when explicitly selected.")
}

func documentUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk document [--db path] [--git-checkpoints] < request.json")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Reads one strict JSON object from stdin and writes one JSON result.")
	_, _ = fmt.Fprintln(w, "Uses configured paths by default; pass --db only for an explicit dataset. Configure additional vault-relative sync ignores through openclerk config.")
	_, _ = fmt.Fprintln(w, "Primitive request shapes:")
	_, _ = fmt.Fprintln(w, `  validate/create_document: {"action":"validate","document":{"path":"notes/example.md","title":"Example","body":"# Example\n\nBody."}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url PDF create: {"action":"ingest_source_url","source":{"url":"https://example.test/source.pdf","path_hint":"sources/example.md","asset_path_hint":"assets/sources/example.pdf","title":"Optional title"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url web create: {"action":"ingest_source_url","source":{"url":"https://example.test/page.html","path_hint":"sources/web/example.md","source_type":"web","title":"Optional title"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url markdown create: {"action":"ingest_source_url","source":{"url":"https://github.com/owner/repo/blob/main/README.md","path_hint":"sources/web/repo-readme.md","source_type":"web","title":"Optional title"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url update: {"action":"ingest_source_url","source":{"url":"https://example.test/page.html","mode":"update","source_type":"web"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url placement plan: {"action":"ingest_source_url","source":{"url":"https://example.test/page.html","mode":"plan","source_type":"web","title":"Optional title"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url inspect: {"action":"ingest_source_url","source":{"url":"https://github.com/owner/repo","mode":"inspect","source_type":"web","title":"Optional title","limit":8}}`)
	_, _ = fmt.Fprintln(w, `  ingest_video_url create: {"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=demo","path_hint":"sources/video-youtube/demo.md","transcript":{"text":"Supplied transcript text.","policy":"supplied","origin":"user_supplied_transcript"}}}`)
	_, _ = fmt.Fprintln(w, `  ingest_video_url update: {"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=demo","mode":"update","transcript":{"text":"Updated supplied transcript text.","policy":"supplied","origin":"user_supplied_transcript"}}}`)
	_, _ = fmt.Fprintln(w, `  list/get/edit: {"action":"list_documents","list":{"path_prefix":"notes/","limit":20}} | {"action":"get_document","doc_id":"doc_id_from_json"} | {"action":"replace_section","doc_id":"doc_id_from_json","heading":"Summary","content":"Updated summary."}`)
	_, _ = fmt.Fprintln(w, `  plan move: {"action":"plan_move_document","move":{"path":"technology/projects.md","target_path":"technology/project-ideas.md","update_indexes":true}}`)
	_, _ = fmt.Fprintln(w, `  move/rename/promote: {"action":"move_document","move":{"doc_id":"doc_id_from_json","target_path":"notes/projects/example.md"}} | {"action":"rename_document","move":{"path":"notes/projects/rough.md","target_path":"notes/projects/precise.md"}} | {"action":"promote_candidate","move":{"path":"notes/candidates/idea.md","target_path":"notes/projects/idea.md"}}`)
	_, _ = fmt.Fprintln(w, "  Move actions preserve stable id, update only reported markdown links/index links, record provenance, refresh projections, and refuse existing targets instead of overwriting.")
	_, _ = fmt.Fprintln(w, `  plan/apply path cleanup: {"action":"plan_path_cleanup","path_cleanup":{"path_prefix":"notes/candidates/","cleanup_kind":"candidate_promotion","target_prefix":"notes/projects/","limit":5}} | {"action":"plan_path_cleanup","autonomy":{"approval_mode":"autonomous_trusted"},"path_cleanup":{"doc_id":"doc_id_from_json","mode":"apply"}}`)
	_, _ = fmt.Fprintln(w, "  Path cleanup plan mode is read-only. Apply mode uses returned low-risk move candidates only and requires autonomous_trusted or autonomous_disposable.")
	_, _ = fmt.Fprintln(w, `  diagnostics: {"action":"resolve_paths"} | {"action":"inspect_layout"}`)
	_, _ = fmt.Fprintln(w, "Promoted workflow action:")
	_, _ = fmt.Fprintln(w, `  compile_synthesis: {"action":"compile_synthesis","synthesis":{"path":"synthesis/example.md","title":"Example","source_refs":["sources/a.md"],"body":"...","body_facts":["..."],"freshness_note":"...","mode":"create_or_update"}}`)
	_, _ = fmt.Fprintln(w, "  Requires path, title, non-empty source_refs, and either body or body_facts. mode defaults to create_or_update.")
	_, _ = fmt.Fprintln(w, "  Returns compile_synthesis.agent_handoff with final-answer evidence; use primitives only after rejection or explicit drill-down.")
	_, _ = fmt.Fprintln(w, `  validation_synthesis_report: {"action":"validation_synthesis_report","validation_synthesis":{"disposable_validation":true,"doc_id":"optional_doc_id","body_facts":["validated claim"],"freshness_note":"checked disposable source evidence"}}`)
	_, _ = fmt.Fprintln(w, "  Requires a routine UX disposable vault copy with the validation marker and returns validation_synthesis.agent_handoff; the live private vault is not the mutation target.")
	_, _ = fmt.Fprintln(w, `  git_lifecycle_report status/history: {"action":"git_lifecycle_report","git_lifecycle":{"mode":"status","paths":["synthesis/example.md"],"limit":10}}`)
	_, _ = fmt.Fprintln(w, `  git_lifecycle_report checkpoint: {"action":"git_lifecycle_report","git_lifecycle":{"mode":"checkpoint","paths":["synthesis/example.md"],"message":"openclerk: update synthesis example"}}`)
	_, _ = fmt.Fprintln(w, "  Status/history are read-only. Checkpoint requires --git-checkpoints or OPENCLERK_GIT_CHECKPOINTS=1, never pushes, switches branches, restores, or emits raw diffs.")
	_, _ = fmt.Fprintln(w, `  web_search_plan: {"action":"web_search_plan","web_search":{"query":"example source","results":[{"url":"https://example.test/page.html","title":"Example","snippet":"Search snippet"}],"limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Plans harness-supplied public URL candidates with duplicate and placement hints; approved fetch/write still uses ingest_source_url.")
	_, _ = fmt.Fprintln(w, `  artifact_candidate_plan: {"action":"artifact_candidate_plan","artifact":{"content":"# Receipt\n\nTotal paid: 42 USD","artifact_kind":"receipt","tags":["finance"],"fields":{"owner":"ap"},"limit":5}}`)
	_, _ = fmt.Fprintln(w, `  artifact_candidate_plan local file: {"action":"artifact_candidate_plan","artifact":{"local_path":"<explicit-user-local-file>","artifact_kind":"receipt","limit":5}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Plans artifact path, title, body preview, tags, metadata fields, duplicate evidence, confidence, and approved create/ingest handoff; explicit local_path inspection is limited to UTF-8 text, markdown, text-bearing PDF, or text_extraction=ocr_review through an installed verified local OCR module; no opaque parsing, fetch, or write.")
}

func retrievalUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk retrieval [--db path] < request.json")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Reads one strict JSON object from stdin and writes one JSON result.")
	_, _ = fmt.Fprintln(w, "Uses configured paths by default; pass --db only for an explicit dataset. Configure additional vault-relative sync ignores through openclerk config.")
	_, _ = fmt.Fprintln(w, "Primitive read-only actions:")
	_, _ = fmt.Fprintln(w, `  document_links: {"action":"document_links","doc_id":"doc_..."}`)
	_, _ = fmt.Fprintln(w, `  graph_neighborhood: {"action":"graph_neighborhood","doc_id":"doc_...","limit":10}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Inspects markdown links and nearby derived graph evidence; canonical markdown remains relationship authority.")
	_, _ = fmt.Fprintln(w, `  records_lookup: {"action":"records_lookup","records":{"text":"AgentOps Escalation Policy","entity_type":"policy","limit":10}}`)
	_, _ = fmt.Fprintln(w, `  services_lookup: {"action":"services_lookup","services":{"text":"OpenClerk runner","interface":"JSON runner","limit":10}}`)
	_, _ = fmt.Fprintln(w, `  decisions_lookup: {"action":"decisions_lookup","decisions":{"text":"knowledge configuration","status":"accepted","scope":"knowledge-configuration","limit":5}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Promoted record, service, and decision projections are derived from canonical markdown with citations and freshness.")
	_, _ = fmt.Fprintln(w, "Promoted workflow actions:")
	_, _ = fmt.Fprintln(w, `  source_discovery_report: {"action":"source_discovery_report","source_discovery":{"query":"representative source evidence","path_prefix":"sources/","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns source_discovery.agent_handoff with sanitized source-category summaries, representative counts, citations, validation boundaries, and authority limits.")
	_, _ = fmt.Fprintln(w, `  source_audit_report: {"action":"source_audit_report","source_audit":{"query":"...","target_path":"synthesis/example.md","mode":"explain","conflict_query":"...","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Default mode is explain; repair_existing may update only an existing synthesis target.")
	_, _ = fmt.Fprintln(w, `  evidence_bundle_report: {"action":"evidence_bundle_report","evidence_bundle":{"query":"...","entity_id":"...","decision_id":"...","ref_kind":"document","ref_id":"...","projection":"records","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns evidence_bundle.agent_handoff with citations, provenance, projection freshness, validation boundaries, and authority limits.")
	_, _ = fmt.Fprintln(w, `  decision_lookup_report: {"action":"decision_lookup_report","decision_lookup":{"query":"decision-like evidence","decision_id":"optional-decision-id","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns decision_lookup.agent_handoff with formal decisions, decision-like records/search evidence, provenance, projection freshness, validation boundaries, and authority limits.")
	_, _ = fmt.Fprintln(w, `  duplicate_candidate_report: {"action":"duplicate_candidate_report","duplicate_candidate":{"query":"renewal packaging notes","path_prefix":"notes/","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns duplicate_candidate.agent_handoff with likely target, evidence inspected, no-write status, and approval boundary.")
	_, _ = fmt.Fprintln(w, `  workflow_guide_report: {"action":"workflow_guide_report","workflow_guide":{"intent":"should I update an existing note or create a new one?"}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns workflow_guide.agent_handoff with runner-owned surface selection guidance; it does not inspect storage or replace the selected action result.")
	_, _ = fmt.Fprintln(w, `  memory_router_recall_report: {"action":"memory_router_recall_report","memory_router_recall":{"query":"memory router temporal recall session promotion feedback weighting routing canonical docs","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns memory_router_recall.agent_handoff with canonical refs, stale-session posture, feedback weighting, provenance, freshness, and authority limits.")
	_, _ = fmt.Fprintln(w, `  structured_store_report: {"action":"structured_store_report","structured_store":{"domain":"records","query":"structured canonical record evidence","entity_type":"tool","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns structured_store.agent_handoff with promoted record/service/decision projection evidence, candidate-store boundaries, freshness, and authority limits.")
	_, _ = fmt.Fprintln(w, `  hybrid_retrieval_report: {"action":"hybrid_retrieval_report","hybrid_retrieval":{"query":"semantic recall citation quality","path_prefix":"docs/","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns hybrid_retrieval.agent_handoff with lexical baseline evidence and candidate-surface boundaries; it does not create vectors or change default ranking.")
	_, _ = fmt.Fprintln(w, `  graph_context_report: {"action":"graph_context_report","graph_context":{"path":"notes/graph/semantics/index.md","limit":20}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns graph_context.agent_handoff with source identity, canonical markdown relationship text, links/backlinks, nearby graph evidence, graph freshness, provenance refs, validation boundaries, and authority limits; it does not add semantic-label graph truth or graph memory.")
	_, _ = fmt.Fprintln(w, `  graph_relationship_report: {"action":"graph_relationship_report","graph_relationship":{"path":"notes/graph/semantics/index.md","limit":20}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns graph_relationship.agent_handoff with relationship paths, direct-vs-derived evidence, typed candidates from cited markdown, limited stale/orphaned/contradiction audit findings, provenance refs, validation boundaries, and authority limits; it does not add semantic-label graph truth or graph memory.")
	_, _ = fmt.Fprintln(w, `  graph_relationship_maintenance_plan: {"action":"graph_relationship_maintenance_plan","graph_relationship_maintenance":{"path":"notes/graph/semantics/index.md","limit":20}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns graph_relationship_maintenance.agent_handoff with candidate section content, approval boundary, next replace/append requests, duplicate handling, rollback/audit path, failure modes, freshness, provenance refs, and planned_no_write status.")
	_, _ = fmt.Fprintln(w, `  semantic_search: {"action":"semantic_search","semantic_search":{"query":"semantic recall citation quality","path_prefix":"docs/","limit":10,"provider":"ollama","embedding_model":"embeddinggemma"}}`)
	_, _ = fmt.Fprintln(w, "  Explicit module-gated mode. Routes through an installed verified Ollama or Gemini module, returns citation-bearing semantic_search hits with cache/provider status, and leaves default search lexical.")
	_, _ = fmt.Fprintln(w, `  retrieval_eval_capture: {"action":"retrieval_eval_capture","retrieval_eval":{"action":"search","search":{"text":"dogfood query","path_prefix":"docs/","limit":10},"capture_path":"retrieval-eval-capture.jsonl"}}`)
	_, _ = fmt.Fprintln(w, `  retrieval_eval_replay: {"action":"retrieval_eval_replay","retrieval_replay":{"capture_path":"retrieval-eval-capture.jsonl","limit":100}}`)
	_, _ = fmt.Fprintln(w, "  Explicit local-only eval loop. Capture is off by default, stores sanitized queries/filters/result ids/provider status/latency only, and replay reports Jaccard/top-1/latency metrics without raw vault content or writes.")
	_, _ = fmt.Fprintln(w, `  search_diagnostics_report: {"action":"search_diagnostics_report","search_diagnostics":{"query":"semantic recall citation quality","intent":"semantic recall","path_prefix":"docs/","limit":10,"provider":"ollama"}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Recommends search versus explicit semantic_search with visible filters, module readiness, cost/latency posture, and no default ranking change.")
	_, _ = fmt.Fprintln(w, `  maintenance_report: {"action":"maintenance_report","maintenance":{"query":"renewal packaging notes","path_prefix":"notes/","limit":20}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Packages layout validity, projection freshness, relationship context, duplicate risk, module posture, and git lifecycle posture; it does not run repairs, cron, or background jobs.")
}

func clerkUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk clerk <run|inbox_scan|context_pack> [options]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Chronicler Lite records after-work evidence by planning session-to-repo-knowledge candidates over OpenClerk Core.")
	_, _ = fmt.Fprintln(w, "Core remains the canonical authority; autonomous/dreaming/always-on Chronicler is shelved.")
	_, _ = fmt.Fprintln(w, "Lite commands:")
	_, _ = fmt.Fprintln(w, "  openclerk clerk run --once [--db path] [--inbox-path path] [--task text] [--query text] [--path-prefix prefix] [--limit n]")
	_, _ = fmt.Fprintln(w, "  openclerk clerk inbox_scan [--db path] [--inbox-path path] [--limit n]")
	_, _ = fmt.Fprintln(w, "  openclerk clerk context_pack [--db path] [--task text] [--query text] [--path-prefix prefix] [--limit n]")
	_, _ = fmt.Fprintln(w, "Chronicler Lite is read-only: planned_no_write=true, writes_performed=0, no daemon/watch mode, no autonomous routing, no hidden memory, and no durable vault writes.")
	_, _ = fmt.Fprintln(w, "Planning that inspects Core evidence requires existing OpenClerk storage; Chronicler will not initialize SQLite.")
}

func clerkRunUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk clerk run --once [--db path] [--inbox-path path] [--task text] [--query text] [--path-prefix prefix] [--limit n]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Runs one read-only Chronicler Lite pass and writes one openclerk-clerk.v1 JSON report.")
	_, _ = fmt.Fprintln(w, "  --inbox-path may be a markdown/text file or an explicit non-recursive directory; it may be repeated.")
	_, _ = fmt.Fprintln(w, "  --task creates a context pack; --query overrides the retrieval query; --path-prefix narrows retrieval.")
	_, _ = fmt.Fprintln(w, "  --limit caps planner/retrieval results.")
	_, _ = fmt.Fprintln(w, "Requires existing OpenClerk storage when inbox or context planning would inspect Core evidence.")
	_, _ = fmt.Fprintln(w, "Output action: clerk_run. Result fields include planned_no_write, writes_performed, inbox_candidates, context_packs, duplicate_risks, pending_review, blockers, and deferred.")
}

func clerkInboxScanUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk clerk inbox_scan [--db path] [--inbox-path path] [--limit n]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Plans explicit local inbox candidates and writes one openclerk-clerk.v1 JSON report.")
	_, _ = fmt.Fprintln(w, "  --inbox-path may be a markdown/text file or an explicit non-recursive directory; it may be repeated.")
	_, _ = fmt.Fprintln(w, "  --limit caps planner results.")
	_, _ = fmt.Fprintln(w, "Requires existing OpenClerk storage; Chronicler will not initialize SQLite.")
	_, _ = fmt.Fprintln(w, "Output action: inbox_scan. Result fields include planned_no_write, writes_performed, inbox_candidates, duplicate_risks, pending_review, blockers, and deferred.")
}

func clerkContextPackUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk clerk context_pack [--db path] [--task text] [--query text] [--path-prefix prefix] [--limit n]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Builds one read-only task context pack and writes one openclerk-clerk.v1 JSON report.")
	_, _ = fmt.Fprintln(w, "  --task creates a context pack; --query overrides the retrieval query; one of --task or --query is required.")
	_, _ = fmt.Fprintln(w, "  --path-prefix narrows retrieval to a vault-relative prefix.")
	_, _ = fmt.Fprintln(w, "  --limit caps retrieval and decision results.")
	_, _ = fmt.Fprintln(w, "Requires existing OpenClerk storage; Chronicler will not initialize SQLite.")
	_, _ = fmt.Fprintln(w, "Output action: context_pack. Result fields include planned_no_write, writes_performed, context_packs, blockers, and deferred.")
}
