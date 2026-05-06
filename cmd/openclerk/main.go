package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

var version string

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
	case "module":
		return runModule(args[1:], stdin, stdout, stderr)
	case "document":
		return runDocument(args[1:], stdin, stdout, stderr)
	case "retrieval":
		return runRetrieval(args[1:], stdin, stdout, stderr)
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
				Name:    "document",
				Command: "openclerk document",
				Posture: "strict JSON runner for validation, source intake, document mutation, placement planning, and document-side workflow blocks",
				Primitive: []capabilityAction{
					{Action: "validate", Purpose: "validate a candidate document without writing", Posture: "read_only"},
					{Action: "create_document", Purpose: "create an approved vault-relative markdown document", Posture: "durable_write_requires_approval"},
					{Action: "ingest_source_url", Purpose: "plan, create, or update public web/PDF source notes through the runner", Posture: "plan_read_only_or_approved_write"},
					{Action: "ingest_video_url", Purpose: "create or update video source notes from supplied transcripts", Posture: "approved_write_with_user_supplied_transcript"},
					{Action: "list_documents", Purpose: "list runner-visible documents", Posture: "read_only"},
					{Action: "get_document", Purpose: "read one runner-visible document by doc_id", Posture: "read_only"},
					{Action: "append_document", Purpose: "append approved content to an existing document", Posture: "durable_write_requires_approval"},
					{Action: "replace_section", Purpose: "replace an approved markdown section", Posture: "durable_write_requires_approval"},
					{Action: "resolve_paths", Purpose: "show configured database and vault paths", Posture: "read_only_diagnostic"},
					{Action: "inspect_layout", Purpose: "inspect configured OpenClerk layout", Posture: "read_only_diagnostic"},
				},
				Workflow: []capabilityAction{
					{Action: "compile_synthesis", Purpose: "create or update source-linked synthesis with freshness and source refs", Posture: "durable_write_requires_approval", Handoff: "compile_synthesis.agent_handoff"},
					{Action: "web_search_plan", Purpose: "plan harness-supplied public search-result intake before fetch/write", Posture: "read_only", Handoff: "web_search_plan.agent_handoff"},
					{Action: "artifact_candidate_plan", Purpose: "plan artifact path, title, body preview, tags, fields, duplicates, and create/ingest handoff", Posture: "read_only", Handoff: "artifact_candidate_plan.agent_handoff"},
					{Action: "git_lifecycle_report", Purpose: "report local storage status/history or create explicit local checkpoints", Posture: "read_only_or_explicit_checkpoint"},
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
					{Action: "source_audit_report", Purpose: "explain source-sensitive gaps or repair an existing synthesis target", Posture: "read_only_or_existing_target_repair", Handoff: "source_audit.agent_handoff"},
					{Action: "evidence_bundle_report", Purpose: "package citations, provenance, projection freshness, and authority limits", Posture: "read_only", Handoff: "evidence_bundle.agent_handoff"},
					{Action: "duplicate_candidate_report", Purpose: "choose update-versus-new evidence before durable writes", Posture: "read_only", Handoff: "duplicate_candidate.agent_handoff"},
					{Action: "workflow_guide_report", Purpose: "select the natural runner surface for an intent", Posture: "read_only", Handoff: "workflow_guide.agent_handoff"},
					{Action: "memory_router_recall_report", Purpose: "package routine memory/router recall evidence without memory transports", Posture: "read_only", Handoff: "memory_router_recall.agent_handoff"},
					{Action: "structured_store_report", Purpose: "review structured-data and canonical-store evidence", Posture: "read_only", Handoff: "structured_store.agent_handoff"},
					{Action: "hybrid_retrieval_report", Purpose: "review lexical baseline and hybrid/vector candidate boundaries", Posture: "read_only", Handoff: "hybrid_retrieval.agent_handoff"},
					{Action: "semantic_search", Purpose: "run explicit citation-bearing semantic search through a verified provider module", Posture: "module_gated_read_only", Handoff: "semantic_search.agent_handoff", Requires: "installed enabled embedding provider module"},
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
		},
		ExtensionPoints: []capabilityExtension{
			{Name: "ollama-embeddings", Kind: "embedding_provider", ManifestPath: "modules/ollama-embeddings/module.json", SkillPath: "modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md", Posture: "local_first_optional_module"},
			{Name: "gemini-embeddings", Kind: "embedding_provider", ManifestPath: "modules/gemini-embeddings/module.json", SkillPath: "modules/gemini-embeddings/skill/gemini-embeddings/SKILL.md", Posture: "explicit_remote_provider_opt_in"},
			{Name: "tesseract-ocr", Kind: "ocr_provider", ManifestPath: "modules/tesseract-ocr/module.json", SkillPath: "modules/tesseract-ocr/skill/tesseract-ocr/SKILL.md", Posture: "local_ocr_review_optional_module"},
		},
		AgentHandoff: capabilityAgentHandoff{
			AnswerSummary: "Use OpenClerk as a narrow mainline runner plus composable document, retrieval, and module building blocks.",
			SelectionGuidance: []string{
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
	return runclient.Config{DatabasePath: *databasePath, GitCheckpoints: *gitCheckpoints}, true
}

func wantsSubcommandHelp(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "help", "-h", "--help":
			return true
		}
	}
	return false
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
	_, _ = fmt.Fprintln(stderr, "usage: openclerk <version|capabilities|init|module|document|retrieval> [--db path]")
	_, _ = fmt.Fprintln(stderr, "       openclerk init [--db path] [--vault-root path]")
	_, _ = fmt.Fprintln(stderr, "       openclerk capabilities")
	_, _ = fmt.Fprintln(stderr, "       openclerk module --help")
	_, _ = fmt.Fprintln(stderr, "       openclerk document --help")
	_, _ = fmt.Fprintln(stderr, "       openclerk retrieval --help")
	_, _ = fmt.Fprintln(stderr, "document/retrieval read strict JSON from stdin and use configured paths by default; pass --db only for an explicit dataset.")
	_, _ = fmt.Fprintln(stderr, "promoted workflow actions: compile_synthesis, ingest_source_url plan, web_search_plan, artifact_candidate_plan, git_lifecycle_report, source_audit_report, evidence_bundle_report, duplicate_candidate_report, workflow_guide_report, memory_router_recall_report, structured_store_report, hybrid_retrieval_report, semantic_search")
}

func moduleUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk module [--db path] < request.json")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Reads one strict JSON object from stdin and writes one JSON result.")
	_, _ = fmt.Fprintln(w, "Manages optional embedding and OCR modules through redacted runtime_config state.")
	_, _ = fmt.Fprintln(w, "Semantic modules always use semantic-retrieval-adapter on PATH; command_args are unsupported.")
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
	_, _ = fmt.Fprintln(w, "Uses configured paths by default; pass --db only for an explicit dataset.")
	_, _ = fmt.Fprintln(w, "Primitive request shapes:")
	_, _ = fmt.Fprintln(w, `  validate/create_document: {"action":"validate","document":{"path":"notes/example.md","title":"Example","body":"# Example\n\nBody."}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url PDF create: {"action":"ingest_source_url","source":{"url":"https://example.test/source.pdf","path_hint":"sources/example.md","asset_path_hint":"assets/sources/example.pdf","title":"Optional title"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url web create: {"action":"ingest_source_url","source":{"url":"https://example.test/page.html","path_hint":"sources/web/example.md","source_type":"web","title":"Optional title"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url update: {"action":"ingest_source_url","source":{"url":"https://example.test/page.html","mode":"update","source_type":"web"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url placement plan: {"action":"ingest_source_url","source":{"url":"https://example.test/page.html","mode":"plan","source_type":"web","title":"Optional title"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_video_url create: {"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=demo","path_hint":"sources/video-youtube/demo.md","transcript":{"text":"Supplied transcript text.","policy":"supplied","origin":"user_supplied_transcript"}}}`)
	_, _ = fmt.Fprintln(w, `  ingest_video_url update: {"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=demo","mode":"update","transcript":{"text":"Updated supplied transcript text.","policy":"supplied","origin":"user_supplied_transcript"}}}`)
	_, _ = fmt.Fprintln(w, `  list/get/edit: {"action":"list_documents","list":{"path_prefix":"notes/","limit":20}} | {"action":"get_document","doc_id":"doc_id_from_json"} | {"action":"replace_section","doc_id":"doc_id_from_json","heading":"Summary","content":"Updated summary."}`)
	_, _ = fmt.Fprintln(w, `  diagnostics: {"action":"resolve_paths"} | {"action":"inspect_layout"}`)
	_, _ = fmt.Fprintln(w, "Promoted workflow action:")
	_, _ = fmt.Fprintln(w, `  compile_synthesis: {"action":"compile_synthesis","synthesis":{"path":"synthesis/example.md","title":"Example","source_refs":["sources/a.md"],"body":"...","body_facts":["..."],"freshness_note":"...","mode":"create_or_update"}}`)
	_, _ = fmt.Fprintln(w, "  Requires path, title, non-empty source_refs, and either body or body_facts. mode defaults to create_or_update.")
	_, _ = fmt.Fprintln(w, "  Returns compile_synthesis.agent_handoff with final-answer evidence; use primitives only after rejection or explicit drill-down.")
	_, _ = fmt.Fprintln(w, `  git_lifecycle_report status/history: {"action":"git_lifecycle_report","git_lifecycle":{"mode":"status","paths":["synthesis/example.md"],"limit":10}}`)
	_, _ = fmt.Fprintln(w, `  git_lifecycle_report checkpoint: {"action":"git_lifecycle_report","git_lifecycle":{"mode":"checkpoint","paths":["synthesis/example.md"],"message":"openclerk: update synthesis example"}}`)
	_, _ = fmt.Fprintln(w, "  Status/history are read-only. Checkpoint requires --git-checkpoints or OPENCLERK_GIT_CHECKPOINTS=1, never pushes, switches branches, restores, or emits raw diffs.")
	_, _ = fmt.Fprintln(w, `  web_search_plan: {"action":"web_search_plan","web_search":{"query":"example source","results":[{"url":"https://example.test/page.html","title":"Example","snippet":"Search snippet"}],"limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Plans harness-supplied public URL candidates with duplicate and placement hints; approved fetch/write still uses ingest_source_url.")
	_, _ = fmt.Fprintln(w, `  artifact_candidate_plan: {"action":"artifact_candidate_plan","artifact":{"content":"# Receipt\n\nTotal paid: 42 USD","artifact_kind":"receipt","tags":["finance"],"fields":{"owner":"ap"},"limit":5}}`)
	_, _ = fmt.Fprintln(w, `  artifact_candidate_plan local file: {"action":"artifact_candidate_plan","artifact":{"local_path":"<explicit-user-local-file>","artifact_kind":"receipt","limit":5}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Plans artifact path, title, body preview, tags, metadata fields, duplicate evidence, confidence, and approved create/ingest handoff; explicit local_path inspection is limited to UTF-8 text, markdown, or text-bearing PDF; no OCR, opaque parsing, fetch, or write.")
}

func retrievalUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk retrieval [--db path] < request.json")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Reads one strict JSON object from stdin and writes one JSON result.")
	_, _ = fmt.Fprintln(w, "Uses configured paths by default; pass --db only for an explicit dataset.")
	_, _ = fmt.Fprintln(w, "Promoted workflow actions:")
	_, _ = fmt.Fprintln(w, `  source_audit_report: {"action":"source_audit_report","source_audit":{"query":"...","target_path":"synthesis/example.md","mode":"explain","conflict_query":"...","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Default mode is explain; repair_existing may update only an existing synthesis target.")
	_, _ = fmt.Fprintln(w, `  evidence_bundle_report: {"action":"evidence_bundle_report","evidence_bundle":{"query":"...","entity_id":"...","decision_id":"...","ref_kind":"document","ref_id":"...","projection":"records","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns evidence_bundle.agent_handoff with citations, provenance, projection freshness, validation boundaries, and authority limits.")
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
	_, _ = fmt.Fprintln(w, `  semantic_search: {"action":"semantic_search","semantic_search":{"query":"semantic recall citation quality","path_prefix":"docs/","limit":10,"provider":"ollama","embedding_model":"embeddinggemma"}}`)
	_, _ = fmt.Fprintln(w, "  Explicit module-gated mode. Routes through an installed verified Ollama or Gemini module, returns citation-bearing semantic_search hits with cache/provider status, and leaves default search lexical.")
}
