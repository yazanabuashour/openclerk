package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/yazanabuashour/openclerk/internal/chronicler"
	"github.com/yazanabuashour/openclerk/internal/runner"
	_ "modernc.org/sqlite"
)

func TestRunnerVersion(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{{"--version"}, {"version"}} {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := run(args, strings.NewReader(""), &stdout, &stderr)
		if code != 0 {
			t.Fatalf("run %v exit = %d stderr=%s", args, code, stderr.String())
		}
		if got := strings.TrimSpace(stdout.String()); !strings.HasPrefix(got, "openclerk ") {
			t.Fatalf("version output = %q, want openclerk prefix", got)
		}
	}
}

func TestCapabilitiesManifestShowsBuildingBlocks(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"capabilities"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("capabilities exit = %d stderr=%s", code, stderr.String())
	}
	var result capabilitiesResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode capabilities: %v\n%s", err, stdout.String())
	}
	if result.SchemaVersion != "openclerk-capabilities.v1" ||
		result.NorthStar != "https://mitchellh.com/writing/building-block-economy" {
		t.Fatalf("capabilities identity = %+v", result)
	}
	if !containsString(result.Principles, "Expose high-quality, well-documented runner primitives that agents can compose.") ||
		!containsString(result.Boundaries, "Default retrieval search remains lexical; semantic ranking is explicit and module-gated.") {
		t.Fatalf("capabilities missing building-block principles or boundaries: %+v", result)
	}
	if !hasCapabilityAction(result, "document", "compile_synthesis") ||
		!hasCapabilityAction(result, "config", "configure_profile") ||
		!hasCapabilityAction(result, "retrieval", "audit_contradictions") ||
		!hasCapabilityAction(result, "retrieval", "source_discovery_report") ||
		!hasCapabilityAction(result, "retrieval", "decision_lookup_report") ||
		!hasCapabilityAction(result, "retrieval", "graph_context_report") ||
		!hasCapabilityAction(result, "retrieval", "graph_relationship_report") ||
		!hasCapabilityAction(result, "retrieval", "graph_relationship_maintenance_plan") ||
		!hasCapabilityAction(result, "retrieval", "semantic_search") ||
		!hasCapabilityAction(result, "retrieval", "retrieval_eval_capture") ||
		!hasCapabilityAction(result, "retrieval", "retrieval_eval_replay") ||
		!hasCapabilityAction(result, "retrieval", "search_diagnostics_report") ||
		!hasCapabilityAction(result, "retrieval", "maintenance_report") ||
		!hasCapabilityAction(result, "module", "install_module") ||
		!hasCapabilityAction(result, "clerk", "run --once") ||
		!hasCapabilityAction(result, "clerk", "session_record_report") ||
		!hasCapabilityAction(result, "clerk", "inbox_scan") ||
		!hasCapabilityAction(result, "clerk", "context_pack") {
		t.Fatalf("capabilities missing expected document/retrieval/module actions: %+v", result.Domains)
	}
	if !hasCapabilityExtension(result, "ollama-embeddings", "modules/ollama-embeddings/module.json") ||
		!hasCapabilityExtension(result, "gemini-embeddings", "modules/gemini-embeddings/module.json") ||
		!hasCapabilityExtension(result, "tesseract-ocr", "modules/tesseract-ocr/module.json") {
		t.Fatalf("capabilities missing expected module extension points: %+v", result.ExtensionPoints)
	}
	if capabilitiesJSON := stdout.String(); !strings.Contains(capabilitiesJSON, "storage") ||
		!strings.Contains(capabilitiesJSON, "git lifecycle") ||
		!strings.Contains(capabilitiesJSON, "module") {
		t.Fatalf("capabilities missing config introspection summary: %s", capabilitiesJSON)
	}
}

func TestSubcommandHelpShowsPromotedWorkflowActions(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "inspect",
			args: []string{"inspect", "--help"},
			want: []string{"openclerk-inspect.v1", "Read-only", "no init", "storage", "vault", "derived knowledge layers", "next safe runner requests"},
		},
		{
			name: "config",
			args: []string{"config", "--help"},
			want: []string{"inspect_config", "configure_profile", "clear_profile", "configure_vault_ignore_paths", "vault_ignore_paths", "approval_mode", "storage", "vault_root", "openclerk init --vault-root", "git_lifecycle", "checkpoint_persistence", "openclerk module configure_module"},
		},
		{
			name: "document",
			args: []string{"document", "--help"},
			want: []string{
				"Primitive request shapes:",
				"ingest_source_url",
				"asset_path_hint",
				"placement plan",
				"ingest_video_url",
				"transcript",
				"compile_synthesis",
				"validation_synthesis_report",
				"web_search_plan",
				"artifact_candidate_plan",
				"plan_move_document",
				"move_document",
				"rename_document",
				"promote_candidate",
				"replace_document",
				"preimage_sha256",
				"refuse existing targets",
				"plan_path_cleanup",
				"autonomous_trusted",
				"ocr_review",
				"verified local OCR module",
				"git_lifecycle_report",
				"--git-checkpoints",
				"body_facts",
				"agent_handoff",
				"mode defaults to create_or_update",
			},
		},
		{
			name: "retrieval",
			args: []string{"retrieval", "--help"},
			want: []string{"document_links", "graph_neighborhood", "canonical markdown remains relationship authority", "records_lookup", "services_lookup", "decisions_lookup", "canonical markdown with citations and freshness", "source_discovery_report", "source_audit_report", "evidence_bundle_report", "decision_lookup_report", "duplicate_candidate_report", "workflow_guide_report", "memory_router_recall_report", "ordinary vault fact recall", "structured_store_report", "hybrid_retrieval_report", "graph_context_report", "graph_relationship_report", "graph_relationship_maintenance_plan", "semantic_search", "retrieval_eval_capture", "retrieval_eval_replay", "search_diagnostics_report", "maintenance_report", "no default ranking change", "agent_handoff", "Read-only"},
		},
		{
			name: "module",
			args: []string{"module", "--help"},
			want: []string{"install_module", "configure_module", "remove_module", "list_modules", "runtime_config:GEMINI_API_KEY", "redacted"},
		},
		{
			name: "clerk",
			args: []string{"clerk", "--help"},
			want: []string{"Chronicler Lite", "session-to-repo-knowledge", "autonomous/dreaming/always-on Chronicler is shelved", "openclerk clerk run --once", "openclerk clerk session_record_report", "openclerk clerk inbox_scan", "openclerk clerk context_pack", "planned_no_write=true", "writes_performed=0", "no durable vault writes"},
		},
		{
			name: "clerk run",
			args: []string{"clerk", "run", "--help"},
			want: []string{"openclerk-clerk.v1", "--inbox-path", "--task", "--query", "blockers", "deferred"},
		},
		{
			name: "clerk session_record_report",
			args: []string{"clerk", "session_record_report", "--help"},
			want: []string{"openclerk-clerk.v1", "--inbox-path", "--task", "Output action: session_record_report", "planned_no_write", "writes_performed"},
		},
		{
			name: "clerk inbox_scan",
			args: []string{"clerk", "inbox_scan", "--help"},
			want: []string{"openclerk-clerk.v1", "--inbox-path", "Output action: inbox_scan", "planned_no_write", "writes_performed"},
		},
		{
			name: "clerk context_pack",
			args: []string{"clerk", "context_pack", "--help"},
			want: []string{"openclerk-clerk.v1", "--task", "--query", "--path-prefix", "Output action: context_pack", "context_packs"},
		},
		{
			name: "demo",
			args: []string{"demo", "--help"},
			want: []string{"openclerk demo <init|ask>", "stale projection freshness", "compile_synthesis repair request", "Knowledge pack templates", "codebase-decisions"},
		},
		{
			name: "demo init",
			args: []string{"demo", "init", "--help"},
			want: []string{"openclerk demo init", "--template", "empty or already marked", "agent-session-to-docs"},
		},
		{
			name: "demo ask",
			args: []string{"demo", "ask", "--help"},
			want: []string{"openclerk demo ask", "stale projection evidence"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := run(tt.args, strings.NewReader(""), &stdout, &stderr)
			if code != 0 {
				t.Fatalf("run %v exit = %d stderr=%s", tt.args, code, stderr.String())
			}
			for _, want := range tt.want {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("help output missing %q:\n%s", want, stdout.String())
				}
			}
		})
	}
}

func TestInspectMissingStorageIsReadOnlyJSON(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	var result inspectEnvelope
	code, stderr := runJSON(t, []string{"inspect", "--db", dbPath}, "", &result)
	if code != 0 {
		t.Fatalf("inspect exit = %d stderr=%s", code, stderr)
	}
	if result.SchemaVersion != inspectSchemaVersion ||
		result.Action != "inspect" ||
		!result.Result.ReadOnly ||
		result.Result.WritesPerformed != 0 ||
		result.Result.Storage.Status != "missing" ||
		result.Result.Storage.DatabaseBound ||
		result.Result.Storage.DatabasePathKind != "flag_override" ||
		result.Result.Vault.Status != "unbound" ||
		result.Result.Knowledge.DerivedLayers.SearchIndex != "unavailable" ||
		result.Result.Modules.Status != "unknown" ||
		len(result.Result.Blockers) == 0 ||
		!inspectHasNextAction(result, "Bind a vault") {
		t.Fatalf("inspect missing result = %+v", result)
	}
	output := mustMarshalInspect(t, result)
	for _, field := range []string{"runner", "storage", "vault", "knowledge", "modules", "git", "recommended_next_actions", "blockers"} {
		if !strings.Contains(output, `"`+field+`"`) {
			t.Fatalf("inspect output missing top-level field %s: %+v", field, result)
		}
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); !os.IsNotExist(err) {
		t.Fatalf("inspect created storage directory: %v", err)
	}
}

func TestInspectUninitializedDatabaseDoesNotBindVault(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dbPath := filepath.Join(root, "data", "openclerk.sqlite")
	writeUnboundSQLiteDatabase(t, dbPath)
	vaultPath := filepath.Join(filepath.Dir(dbPath), "vault")

	var result inspectEnvelope
	code, stderr := runJSON(t, []string{"inspect", "--db", dbPath}, "", &result)
	if code != 0 {
		t.Fatalf("inspect exit = %d stderr=%s", code, stderr)
	}
	if result.Result.Storage.Status != "uninitialized" ||
		!result.Result.Storage.DatabaseBound ||
		result.Result.Vault.Status != "unbound" ||
		result.Result.Vault.VaultRootKind != "unconfigured" ||
		result.Result.WritesPerformed != 0 ||
		len(result.Result.Storage.Blockers) == 0 {
		t.Fatalf("uninitialized inspect result = %+v", result)
	}
	if _, err := os.Stat(vaultPath); !os.IsNotExist(err) {
		t.Fatalf("inspect bound or created vault root: %v", err)
	}
}

func TestInspectBoundEmptyVaultJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dbPath := filepath.Join(root, "data", "openclerk.sqlite")
	vaultRoot := filepath.Join(root, "vault")
	var initResult struct{}
	code, stderr := runJSON(t, []string{"init", "--db", dbPath, "--vault-root", vaultRoot}, "", &initResult)
	if code != 0 {
		t.Fatalf("init exit = %d stderr=%s", code, stderr)
	}
	for _, path := range []string{dbPath + "-shm", dbPath + "-wal"} {
		_ = os.Remove(path)
	}

	var result inspectEnvelope
	code, stderr = runJSON(t, []string{"inspect", "--db", dbPath}, "", &result)
	if code != 0 {
		t.Fatalf("inspect exit = %d stderr=%s", code, stderr)
	}
	if result.Result.Storage.Status != "ready" ||
		result.Result.Vault.Status != "ready" ||
		result.Result.Vault.Documents.KnownCount != 0 ||
		result.Result.Vault.Documents.ChunkCount != 0 ||
		result.Result.Modules.Status != "none_installed" ||
		result.Result.Knowledge.CanonicalMarkdownAuthority != true ||
		result.Result.Knowledge.DerivedLayers.SearchIndex != "ready" ||
		result.Result.Git.Status != "not_git" ||
		result.Result.WritesPerformed != 0 ||
		!inspectHasNextAction(result, "Get task context") {
		t.Fatalf("bound empty inspect result = %+v", result)
	}
	for _, path := range []string{dbPath + "-shm", dbPath + "-wal"} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("inspect created sqlite sidecar %s: %v", filepath.Base(path), err)
		}
	}
}

func TestInspectPopulatedVaultReportsCountsAndStaleSynthesis(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	sourceRequest := `{"action":"create_document","document":{"path":"sources/inspect-source.md","title":"Inspect Source","body":"# Inspect Source\n\n## Summary\nInitial inspect evidence.\n"}}`
	var source runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, sourceRequest, &source)
	if code != 0 {
		t.Fatalf("create source exit = %d stderr=%s", code, stderr)
	}
	synthesisRequest := `{"action":"create_document","document":{"path":"synthesis/inspect-summary.md","title":"Inspect Summary","body":"---\ntype: synthesis\nstatus: active\nfreshness: fresh\nsource_refs: sources/inspect-source.md\n---\n# Inspect Summary\n\n## Summary\nInitial inspect evidence.\n\n## Sources\n- sources/inspect-source.md\n\n## Freshness\nChecked source refs.\n"}}`
	var synthesis runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, synthesisRequest, &synthesis)
	if code != 0 {
		t.Fatalf("create synthesis exit = %d stderr=%s", code, stderr)
	}
	updateRequest := `{"action":"replace_section","doc_id":"` + source.Document.DocID + `","heading":"Summary","content":"Updated inspect evidence."}`
	var update runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, updateRequest, &update)
	if code != 0 {
		t.Fatalf("update source exit = %d stderr=%s", code, stderr)
	}

	var result inspectEnvelope
	code, stderr = runJSON(t, []string{"inspect", "--db", dbPath}, "", &result)
	if code != 0 {
		t.Fatalf("inspect exit = %d stderr=%s", code, stderr)
	}
	if result.Result.Storage.Status != "ready" ||
		result.Result.Vault.Status != "ready" ||
		result.Result.Vault.Documents.KnownCount != 2 ||
		result.Result.Vault.Documents.ChunkCount == 0 ||
		result.Result.Knowledge.DerivedLayers.SearchIndex != "ready" ||
		result.Result.Knowledge.DerivedLayers.Synthesis != "stale" ||
		result.Result.Knowledge.Synthesis.StaleCount != 1 ||
		result.Result.Knowledge.Synthesis.MissingSourceRefCount != 0 ||
		result.Result.Knowledge.DuplicateRisk.Status != "checked" ||
		result.Result.Knowledge.DuplicateRisk.CandidateCount != 0 ||
		result.Result.Modules.Status != "none_installed" ||
		!inspectHasNextAction(result, "Review stale synthesis") {
		t.Fatalf("populated inspect result = %+v", result)
	}
}

func TestInspectGitWorktreeDoesNotRefreshIndex(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dbPath := filepath.Join(root, "data", "openclerk.sqlite")
	vaultRoot := filepath.Join(root, "vault")
	var initResult struct{}
	code, stderr := runJSON(t, []string{"init", "--db", dbPath, "--vault-root", vaultRoot}, "", &initResult)
	if code != 0 {
		t.Fatalf("init exit = %d stderr=%s", code, stderr)
	}
	runGitForTest(t, vaultRoot, "init", "-q")
	runGitForTest(t, vaultRoot, "config", "user.email", "test@example.com")
	runGitForTest(t, vaultRoot, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(vaultRoot, "note.md"), []byte("# Note\n"), 0o644); err != nil {
		t.Fatalf("write git note: %v", err)
	}
	runGitForTest(t, vaultRoot, "add", "note.md")
	runGitForTest(t, vaultRoot, "commit", "-q", "-m", "init")
	indexPath := filepath.Join(vaultRoot, ".git", "index")
	before := fileModTime(t, indexPath)

	var result inspectEnvelope
	code, stderr = runJSON(t, []string{"inspect", "--db", dbPath}, "", &result)
	if code != 0 {
		t.Fatalf("inspect exit = %d stderr=%s", code, stderr)
	}
	if result.Result.Git.Status != "clean" {
		t.Fatalf("git inspect result = %+v", result.Result.Git)
	}
	if after := fileModTime(t, indexPath); !after.Equal(before) {
		t.Fatalf("inspect refreshed git index: before=%s after=%s", before, after)
	}
}

func TestInspectMissingVaultRootReturnsBlockerWithoutCreatingIt(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dbPath := filepath.Join(root, "data", "openclerk.sqlite")
	vaultRoot := filepath.Join(root, "vault")
	var initResult struct{}
	code, stderr := runJSON(t, []string{"init", "--db", dbPath, "--vault-root", vaultRoot}, "", &initResult)
	if code != 0 {
		t.Fatalf("init exit = %d stderr=%s", code, stderr)
	}
	if err := os.RemoveAll(vaultRoot); err != nil {
		t.Fatalf("remove vault root: %v", err)
	}

	var result inspectEnvelope
	code, stderr = runJSON(t, []string{"inspect", "--db", dbPath}, "", &result)
	if code != 0 {
		t.Fatalf("inspect exit = %d stderr=%s", code, stderr)
	}
	if result.Result.Storage.Status != "ready" ||
		result.Result.Vault.Status != "missing" ||
		len(result.Result.Vault.Blockers) == 0 ||
		result.Result.Vault.Blockers[0].Code != "vault_missing" ||
		result.Result.WritesPerformed != 0 {
		t.Fatalf("missing vault inspect result = %+v", result)
	}
	if _, err := os.Stat(vaultRoot); !os.IsNotExist(err) {
		t.Fatalf("inspect recreated missing vault root: %v", err)
	}
}

func TestDemoInitRejectsUnmarkedNonEmptyRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	vaultPath := filepath.Join(root, "vault", "real.md")
	if err := os.MkdirAll(filepath.Dir(vaultPath), 0o755); err != nil {
		t.Fatalf("mkdir existing vault: %v", err)
	}
	if err := os.WriteFile(vaultPath, []byte("# Real vault\n"), 0o644); err != nil {
		t.Fatalf("write existing vault: %v", err)
	}

	code, stderr := runJSON(t, []string{"demo", "init", "--root", root}, "", nil)
	if code != 1 || !strings.Contains(stderr, "not empty and is not marked") {
		t.Fatalf("demo init exit=%d stderr=%s", code, stderr)
	}
	if _, err := os.Stat(vaultPath); err != nil {
		t.Fatalf("existing vault was not preserved: %v", err)
	}
}

func TestDemoInitAndAskJSON(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "openclerk-demo")
	var initResult demoResult
	code, stderr := runJSON(t, []string{"demo", "init", "--root", root}, "", &initResult)
	if code != 0 {
		t.Fatalf("demo init exit = %d stderr=%s", code, stderr)
	}
	if initResult.SchemaVersion != "openclerk-demo.v1" ||
		initResult.Action != "init" ||
		initResult.SourcePath != "sources/plan.md" ||
		initResult.SynthesisPath != "synthesis/account-memory.md" ||
		initResult.RepairRequest == nil ||
		initResult.RepairRequest.Action != runner.DocumentTaskActionCompileSynthesis {
		t.Fatalf("demo init result = %+v", initResult)
	}
	if len(initResult.ProjectionFreshness) != 1 ||
		initResult.ProjectionFreshness[0].Freshness != "stale" ||
		initResult.ProjectionFreshness[0].StaleSourceRefs != "sources/plan.md" {
		t.Fatalf("demo init projection = %+v", initResult.ProjectionFreshness)
	}
	if _, err := os.Stat(filepath.Join(root, demoMarkerFile)); err != nil {
		t.Fatalf("demo marker missing: %v", err)
	}

	var askResult demoResult
	code, stderr = runJSON(t, []string{"demo", "ask", "--root", root, "what changed and what is stale?"}, "", &askResult)
	if code != 0 {
		t.Fatalf("demo ask exit = %d stderr=%s", code, stderr)
	}
	if askResult.Action != "ask" ||
		!strings.Contains(askResult.Summary, "synthesis/account-memory.md") ||
		!strings.Contains(askResult.Summary, "sources/plan.md") ||
		askResult.RepairRequest == nil ||
		len(askResult.RepairRequest.Synthesis.SourceRefs) != 1 ||
		askResult.RepairRequest.Synthesis.SourceRefs[0] != "sources/plan.md" {
		t.Fatalf("demo ask result = %+v", askResult)
	}
}

func TestDemoInitKnowledgePackTemplates(t *testing.T) {
	t.Parallel()

	for _, template := range availableDemoTemplates() {
		template := template
		t.Run(template, func(t *testing.T) {
			t.Parallel()

			root := filepath.Join(t.TempDir(), "openclerk-demo")
			var initResult demoResult
			code, stderr := runJSON(t, []string{"demo", "init", "--root", root, "--template", template}, "", &initResult)
			if code != 0 {
				t.Fatalf("demo init template exit = %d stderr=%s", code, stderr)
			}
			if initResult.SchemaVersion != "openclerk-demo.v1" ||
				initResult.Action != "init" ||
				initResult.Template != template ||
				initResult.TemplateDocumentCount == 0 ||
				len(initResult.Next) == 0 {
				t.Fatalf("template init result = %+v", initResult)
			}
			if _, err := os.Stat(filepath.Join(root, demoMarkerFile)); err != nil {
				t.Fatalf("demo marker missing: %v", err)
			}

			var inspect inspectEnvelope
			code, stderr = runJSON(t, []string{"inspect", "--db", initResult.DatabasePath}, "", &inspect)
			if code != 0 {
				t.Fatalf("inspect template exit = %d stderr=%s", code, stderr)
			}
			if inspect.Result.Storage.Status != "ready" ||
				inspect.Result.Vault.Status != "ready" ||
				inspect.Result.Vault.Documents.KnownCount != initResult.TemplateDocumentCount {
				t.Fatalf("inspect template result = %+v", inspect)
			}

			var search runner.RetrievalTaskResult
			code, stderr = runJSON(t, []string{"retrieval", "--db", initResult.DatabasePath}, `{"action":"search","search":{"text":"Summary","limit":5}}`, &search)
			if code != 0 {
				t.Fatalf("search template exit = %d stderr=%s", code, stderr)
			}
			if search.Search == nil || len(search.Search.Hits) == 0 {
				t.Fatalf("search template result = %+v", search)
			}

			if template == "stale-runbook" {
				var document runner.DocumentTaskResult
				code, stderr = runJSON(t, []string{"document", "--db", initResult.DatabasePath}, `{"action":"get_document","path":"synthesis/deploy-runbook.md"}`, &document)
				if code != 0 {
					t.Fatalf("get stale runbook exit = %d stderr=%s", code, stderr)
				}
				if document.Document == nil {
					t.Fatalf("get stale runbook result = %+v", document)
				}
				var projections runner.RetrievalTaskResult
				code, stderr = runJSON(t, []string{"retrieval", "--db", initResult.DatabasePath}, fmt.Sprintf(`{"action":"projection_states","projection":{"projection":"synthesis","ref_kind":"document","ref_id":%q,"limit":10}}`, document.Document.DocID), &projections)
				if code != 0 {
					t.Fatalf("projection stale runbook exit = %d stderr=%s", code, stderr)
				}
				if projections.Projections == nil ||
					len(projections.Projections.Projections) != 1 ||
					projections.Projections.Projections[0].Freshness != "stale" ||
					projections.Projections.Projections[0].Details["stale_source_refs"] != "sources/deploy-source.md" {
					t.Fatalf("stale runbook projections = %+v", projections)
				}
			}
		})
	}
}

func TestDemoInitUnknownTemplateListsAvailableTemplates(t *testing.T) {
	t.Parallel()

	var result demoResult
	code, stderr := runJSON(t, []string{"demo", "init", "--root", filepath.Join(t.TempDir(), "demo"), "--template", "missing-template"}, "", &result)
	if code != 2 ||
		!strings.Contains(stderr, "unknown demo template") ||
		!strings.Contains(stderr, "codebase-decisions") ||
		!strings.Contains(stderr, "agent-session-to-docs") {
		t.Fatalf("unknown template exit=%d stderr=%s", code, stderr)
	}
}

func TestDemoInitTemplateDoesNotTouchConfiguredVault(t *testing.T) {
	configRoot := t.TempDir()
	configuredDB := filepath.Join(configRoot, "data", "openclerk.sqlite")
	configuredVault := filepath.Join(configRoot, "real-vault")
	if err := os.MkdirAll(configuredVault, 0o755); err != nil {
		t.Fatalf("create configured vault: %v", err)
	}
	realPath := filepath.Join(configuredVault, "real.md")
	if err := os.WriteFile(realPath, []byte("# Real Vault\n"), 0o644); err != nil {
		t.Fatalf("write configured vault: %v", err)
	}
	var initConfigured struct{}
	code, stderr := runJSON(t, []string{"init", "--db", configuredDB, "--vault-root", configuredVault}, "", &initConfigured)
	if code != 0 {
		t.Fatalf("init configured vault exit=%d stderr=%s", code, stderr)
	}
	t.Setenv("OPENCLERK_DATABASE_PATH", configuredDB)

	demoRoot := filepath.Join(t.TempDir(), "demo")
	var initResult demoResult
	code, stderr = runJSON(t, []string{"demo", "init", "--root", demoRoot, "--template", "research-vault"}, "", &initResult)
	if code != 0 {
		t.Fatalf("demo template exit=%d stderr=%s", code, stderr)
	}
	if initResult.VaultRoot == configuredVault {
		t.Fatalf("demo reused configured vault: %+v", initResult)
	}
	if data, err := os.ReadFile(realPath); err != nil || string(data) != "# Real Vault\n" {
		t.Fatalf("configured vault changed: data=%q err=%v", string(data), err)
	}
}

func TestClerkInboxScanJSON(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	createRequest := `{"action":"create_document","document":{"path":"notes/cli-inbox-seed.md","title":"CLI Inbox Seed","body":"# CLI Inbox Seed\n\nExisting Core storage seed.\n"}}`
	var createResult runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, createRequest, &createResult)
	if code != 0 {
		t.Fatalf("create document exit = %d stderr=%s", code, stderr)
	}
	if createResult.Document == nil {
		t.Fatalf("create document result = %+v", createResult)
	}
	inboxPath := filepath.Join(t.TempDir(), "candidate.md")
	if err := os.WriteFile(inboxPath, []byte("# CLI Inbox Candidate\n\nCLI inbox scan marker.\n"), 0o644); err != nil {
		t.Fatalf("write inbox: %v", err)
	}

	var result chronicler.RunEnvelope
	code, stderr = runJSON(t, []string{"clerk", "inbox_scan", "--db", dbPath, "--inbox-path", inboxPath, "--limit", "5"}, "", &result)
	if code != 0 {
		t.Fatalf("clerk inbox_scan exit = %d stderr=%s", code, stderr)
	}
	if result.SchemaVersion != chronicler.SchemaVersion ||
		result.Action != chronicler.ActionInboxScan ||
		result.Result.Mode != chronicler.ActionInboxScan ||
		!result.Result.PlannedNoWrite ||
		result.Result.WritesPerformed != 0 ||
		len(result.Result.InboxCandidates) != 1 ||
		len(result.Result.ContextPacks) != 0 {
		t.Fatalf("clerk inbox_scan result = %+v", result)
	}
	if result.Result.InboxCandidates[0].WriteStatus != "planned_no_write" {
		t.Fatalf("candidate = %+v", result.Result.InboxCandidates[0])
	}
}

func TestClerkContextPackJSON(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	createRequest := `{"action":"create_document","document":{"path":"docs/cli/context-pack.md","title":"CLI Context Pack","body":"# CLI Context Pack\n\nCLI context pack marker evidence.\n"}}`
	var createResult runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, createRequest, &createResult)
	if code != 0 {
		t.Fatalf("create document exit = %d stderr=%s", code, stderr)
	}
	if createResult.Document == nil {
		t.Fatalf("create document result = %+v", createResult)
	}

	var result chronicler.RunEnvelope
	code, stderr = runJSON(t, []string{"clerk", "context_pack", "--db", dbPath, "--task", "CLI context pack marker", "--limit", "5"}, "", &result)
	if code != 0 {
		t.Fatalf("clerk context_pack exit = %d stderr=%s", code, stderr)
	}
	if result.SchemaVersion != chronicler.SchemaVersion ||
		result.Action != chronicler.ActionContextPack ||
		result.Result.Mode != chronicler.ActionContextPack ||
		!result.Result.PlannedNoWrite ||
		result.Result.WritesPerformed != 0 ||
		len(result.Result.InboxCandidates) != 0 ||
		len(result.Result.ContextPacks) != 1 {
		t.Fatalf("clerk context_pack result = %+v", result)
	}
	if len(result.Result.ContextPacks[0].MustRead) == 0 ||
		result.Result.ContextPacks[0].WriteStatus != "read_only_no_write" {
		t.Fatalf("context pack = %+v", result.Result.ContextPacks[0])
	}
}

func TestClerkSessionRecordReportJSON(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	createRequest := `{"action":"create_document","document":{"path":"docs/cli/session-record.md","title":"CLI Session Record","body":"# CLI Session Record\n\nCLI session record marker evidence.\n"}}`
	var createResult runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, createRequest, &createResult)
	if code != 0 {
		t.Fatalf("create document exit = %d stderr=%s", code, stderr)
	}
	if createResult.Document == nil {
		t.Fatalf("create document result = %+v", createResult)
	}
	inboxPath := filepath.Join(t.TempDir(), "session.md")
	if err := os.WriteFile(inboxPath, []byte("# CLI Session\n\nCLI session record marker candidate.\n"), 0o644); err != nil {
		t.Fatalf("write session: %v", err)
	}

	var result chronicler.RunEnvelope
	code, stderr = runJSON(t, []string{"clerk", "session_record_report", "--db", dbPath, "--inbox-path", inboxPath, "--task", "CLI session record marker", "--limit", "5"}, "", &result)
	if code != 0 {
		t.Fatalf("clerk session_record_report exit = %d stderr=%s", code, stderr)
	}
	if result.SchemaVersion != chronicler.SchemaVersion ||
		result.Action != chronicler.ActionSessionRecordReport ||
		result.Result.Mode != chronicler.ActionSessionRecordReport ||
		!result.Result.PlannedNoWrite ||
		result.Result.WritesPerformed != 0 ||
		len(result.Result.InboxCandidates) != 1 ||
		len(result.Result.ContextPacks) != 1 {
		t.Fatalf("clerk session_record_report result = %+v", result)
	}
	if result.Result.InboxCandidates[0].NextCreateRequest == "" ||
		result.Result.InboxCandidates[0].WriteStatus != "planned_no_write" ||
		result.Result.ContextPacks[0].WriteStatus != "read_only_no_write" {
		t.Fatalf("clerk session_record_report detail = %+v", result.Result)
	}
}

func TestClerkRunOnceEmptyJSON(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	var result chronicler.RunEnvelope
	code, stderr := runJSON(t, []string{"clerk", "run", "--once", "--db", dbPath}, "", &result)
	if code != 0 {
		t.Fatalf("clerk run exit = %d stderr=%s", code, stderr)
	}
	if result.SchemaVersion != chronicler.SchemaVersion ||
		result.Action != chronicler.ActionRun ||
		result.Result.Mode != "once" ||
		!result.Result.PlannedNoWrite ||
		result.Result.WritesPerformed != 0 ||
		len(result.Result.InboxCandidates) != 0 ||
		len(result.Result.ContextPacks) != 0 ||
		len(result.Result.Blockers) != 0 {
		t.Fatalf("clerk result = %+v", result)
	}
}

func TestRunnerConfigProfileJSONPersistsAcrossInvocations(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	configureRequest := `{"action":"configure_profile","autonomy":{"approval_mode":"approve_write"},"profile":{"approval_mode":"propose_only","drafting_mode":"require_explicit_fields","write_target_mode":"existing_only","citation_mode":"strict","privacy_mode":"private_summary_only","audience_mode":"executive_summary"}}`
	var configured runner.ConfigTaskResult
	code, stderr := runJSON(t, []string{"config", "--db", dbPath}, configureRequest, &configured)
	if code != 0 {
		t.Fatalf("configure profile exit = %d stderr=%s", code, stderr)
	}
	if configured.Rejected ||
		configured.Profile.ApprovalMode != runner.ApprovalModeProposeOnly ||
		configured.Profile.DraftingMode != runner.DraftingModeRequireExplicitFields ||
		configured.Profile.WriteTargetMode != runner.WriteTargetModeExistingOnly ||
		configured.Profile.CitationMode != runner.CitationModeStrict ||
		configured.Profile.PrivacyMode != runner.PrivacyModePrivateSummaryOnly ||
		configured.Profile.AudienceMode != runner.AudienceModeExecutiveSummary {
		t.Fatalf("configured profile = %+v", configured)
	}

	var inspected runner.ConfigTaskResult
	code, stderr = runJSON(t, []string{"config", "--db", dbPath}, `{"action":"inspect_config"}`, &inspected)
	if code != 0 {
		t.Fatalf("inspect profile exit = %d stderr=%s", code, stderr)
	}
	if inspected.Profile != configured.Profile {
		t.Fatalf("inspected profile = %+v, want %+v", inspected.Profile, configured.Profile)
	}
	if inspected.Storage == nil ||
		inspected.Storage.DatabasePath != dbPath ||
		inspected.Storage.DatabaseSource != "flag" ||
		inspected.GitLifecycle == nil ||
		inspected.GitLifecycle.CheckpointPersistence != "unsupported" {
		t.Fatalf("inspected config summary = %+v", inspected)
	}

	var invalid runner.ConfigTaskResult
	code, stderr = runJSON(t, []string{"config", "--db", dbPath}, `{"action":"configure_profile","autonomy":{"approval_mode":"approve_write"},"profile":{"privacy_mode":"public_everything"}}`, &invalid)
	if code != 0 {
		t.Fatalf("invalid profile exit = %d stderr=%s", code, stderr)
	}
	if !invalid.Rejected || !strings.Contains(invalid.RejectionReason, "profile.privacy_mode") {
		t.Fatalf("invalid profile = %+v", invalid)
	}

	var cleared runner.ConfigTaskResult
	code, stderr = runJSON(t, []string{"config", "--db", dbPath}, `{"action":"clear_profile","autonomy":{"approval_mode":"approve_write"}}`, &cleared)
	if code != 0 {
		t.Fatalf("clear profile exit = %d stderr=%s", code, stderr)
	}
	if cleared.Rejected || cleared.Profile.ApprovalMode != runner.ApprovalModeApproveWrite {
		t.Fatalf("cleared profile = %+v", cleared)
	}
}

func TestRunnerConfigVaultIgnoreJSONPersistsAcrossInvocations(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	configureRequest := `{"action":"configure_vault_ignore_paths","autonomy":{"approval_mode":"approve_write"},"vault_ignore_paths":["scratch/"]}`
	var configured runner.ConfigTaskResult
	code, stderr := runJSON(t, []string{"config", "--db", dbPath}, configureRequest, &configured)
	if code != 0 {
		t.Fatalf("configure vault ignores exit = %d stderr=%s", code, stderr)
	}
	if configured.Storage == nil || !containsString(configured.Storage.CustomVaultIgnorePaths, "scratch") {
		t.Fatalf("configured storage = %+v", configured.Storage)
	}

	var inspected runner.ConfigTaskResult
	code, stderr = runJSON(t, []string{"config", "--db", dbPath}, `{"action":"inspect_config"}`, &inspected)
	if code != 0 {
		t.Fatalf("inspect config exit = %d stderr=%s", code, stderr)
	}
	if inspected.Storage == nil {
		t.Fatalf("storage = nil")
	}
	for _, path := range []string{".git", ".stversions", ".openclerk", ".backups", "scratch"} {
		if !containsString(inspected.Storage.VaultIgnorePaths, path) {
			t.Fatalf("vault ignore paths = %+v, missing %s", inspected.Storage.VaultIgnorePaths, path)
		}
	}
	if !containsString(inspected.Storage.CustomVaultIgnorePaths, "scratch") {
		t.Fatalf("custom vault ignore paths = %+v, missing scratch", inspected.Storage.CustomVaultIgnorePaths)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func hasCapabilityAction(result capabilitiesResult, domainName string, action string) bool {
	for _, domain := range result.Domains {
		if domain.Name != domainName {
			continue
		}
		for _, candidate := range append(domain.Primitive, domain.Workflow...) {
			if candidate.Action == action {
				return true
			}
		}
	}
	return false
}

func hasCapabilityExtension(result capabilitiesResult, name string, manifestPath string) bool {
	for _, extension := range result.ExtensionPoints {
		if extension.Name == name && extension.ManifestPath == manifestPath {
			return true
		}
	}
	return false
}

func TestRunnerModuleInstallConfigureListRemoveJSON(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	manifestPath := writeCLISemanticModuleManifest(t, t.TempDir(), "gemini")
	commandPath := writeCLIExecutable(t, "semantic-retrieval-adapter")
	installRequest := `{"action":"install_module","module":{"provider":"gemini","manifest_path":"` + filepath.ToSlash(manifestPath) + `","command":"` + filepath.ToSlash(commandPath) + `","provider_config":{"embedding_model":"gemini-embedding-001","gemini_api_base":"https://generativelanguage.googleapis.com/v1beta","api_key":"do-not-store"}}}`
	var installResult moduleTaskResult
	code, stderr := runJSON(t, []string{"module", "--db", dbPath}, installRequest, &installResult)
	if code != 0 {
		t.Fatalf("install exit = %d stderr=%s", code, stderr)
	}
	if installResult.Rejected ||
		installResult.Module == nil ||
		installResult.Module.Provider != "gemini" ||
		!installResult.Module.Enabled ||
		installResult.Module.ProviderConfig["credential_ref"] != "runtime_config:GEMINI_API_KEY" ||
		installResult.Module.ProviderConfig["api_key"] != "" {
		t.Fatalf("install result = %+v", installResult)
	}

	configureRequest := `{"action":"configure_module","config":{"provider":"gemini","enabled":false,"provider_config":{"embedding_output_dimensions":"3072"}}}`
	var configureResult moduleTaskResult
	code, stderr = runJSON(t, []string{"module", "--db", dbPath}, configureRequest, &configureResult)
	if code != 0 {
		t.Fatalf("configure exit = %d stderr=%s", code, stderr)
	}
	if configureResult.Module == nil ||
		configureResult.Module.Enabled ||
		configureResult.Module.ProviderConfig["embedding_output_dimensions"] != "3072" ||
		configureResult.Module.ProviderConfig["credential_ref"] != "runtime_config:GEMINI_API_KEY" {
		t.Fatalf("configure result = %+v", configureResult)
	}

	var inspectResult runner.ConfigTaskResult
	code, stderr = runJSON(t, []string{"config", "--db", dbPath}, `{"action":"inspect_config"}`, &inspectResult)
	if code != 0 {
		t.Fatalf("inspect config exit = %d stderr=%s", code, stderr)
	}
	if len(inspectResult.Modules) != 1 ||
		inspectResult.Modules[0].Provider != "gemini" ||
		inspectResult.Modules[0].Kind != "embedding_provider" ||
		inspectResult.Modules[0].Enabled ||
		inspectResult.Modules[0].RedactionStatus != "redacted" {
		t.Fatalf("inspect module summaries = %+v", inspectResult.Modules)
	}

	var listResult moduleTaskResult
	code, stderr = runJSON(t, []string{"module", "--db", dbPath}, `{"action":"list_modules"}`, &listResult)
	if code != 0 {
		t.Fatalf("list exit = %d stderr=%s", code, stderr)
	}
	if len(listResult.Modules) != 1 || listResult.Modules[0].Provider != "gemini" || listResult.Modules[0].ProviderConfig["api_key"] != "" {
		t.Fatalf("list result = %+v", listResult)
	}

	var removeResult moduleTaskResult
	code, stderr = runJSON(t, []string{"module", "--db", dbPath}, `{"action":"remove_module","provider":"gemini"}`, &removeResult)
	if code != 0 {
		t.Fatalf("remove exit = %d stderr=%s", code, stderr)
	}
	if removeResult.Module == nil || removeResult.Module.Enabled {
		t.Fatalf("remove result = %+v", removeResult)
	}
	var emptyList moduleTaskResult
	code, stderr = runJSON(t, []string{"module", "--db", dbPath}, `{"action":"list_modules"}`, &emptyList)
	if code != 0 {
		t.Fatalf("empty list exit = %d stderr=%s", code, stderr)
	}
	if len(emptyList.Modules) != 0 {
		t.Fatalf("empty list result = %+v", emptyList)
	}
}

func TestRunnerModuleInstallTesseractOCRJSON(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	manifestPath := writeCLIOCRModuleManifest(t, t.TempDir())
	tesseractPath := writeCLIExecutable(t, "tesseract")
	ocrmypdfPath := writeCLIExecutable(t, "ocrmypdf")
	installRequest := `{"action":"install_module","module":{"kind":"ocr_provider","provider":"tesseract","manifest_path":"` + filepath.ToSlash(manifestPath) + `","command":"` + filepath.ToSlash(tesseractPath) + `","provider_config":{"ocrmypdf_command":"` + filepath.ToSlash(ocrmypdfPath) + `","language":"eng"}}}`
	var installResult moduleTaskResult
	code, stderr := runJSON(t, []string{"module", "--db", dbPath}, installRequest, &installResult)
	if code != 0 {
		t.Fatalf("install OCR exit = %d stderr=%s", code, stderr)
	}
	if installResult.Rejected ||
		installResult.Module == nil ||
		installResult.Module.Kind != "ocr_provider" ||
		installResult.Module.Provider != "tesseract" ||
		installResult.Module.ProviderConfig["language"] != "eng" {
		t.Fatalf("install OCR result = %+v", installResult)
	}

	var listResult moduleTaskResult
	code, stderr = runJSON(t, []string{"module", "--db", dbPath}, `{"action":"list_modules"}`, &listResult)
	if code != 0 {
		t.Fatalf("list OCR exit = %d stderr=%s", code, stderr)
	}
	if len(listResult.Modules) != 1 || listResult.Modules[0].Kind != "ocr_provider" {
		t.Fatalf("list OCR result = %+v", listResult)
	}
}

func TestRunnerModuleInstallUsesManifestRootInput(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	releaseRoot := t.TempDir()
	manifestPath := writeCLISemanticModuleManifest(t, filepath.Join(releaseRoot, "modules", "ollama-embeddings"), "ollama")
	commandPath := writeCLIExecutable(t, "semantic-retrieval-adapter")
	manifestRel, err := filepath.Rel(releaseRoot, manifestPath)
	if err != nil {
		t.Fatalf("relative manifest path: %v", err)
	}
	installRequest := `{"action":"install_module","module":{"provider":"ollama","manifest_path":"` + filepath.ToSlash(manifestRel) + `","manifest_root":"` + filepath.ToSlash(releaseRoot) + `","command":"` + filepath.ToSlash(commandPath) + `","provider_config":{"embedding_model":"embeddinggemma","ollama_url":"http://localhost:11434"}}}`
	var installResult moduleTaskResult
	code, stderr := runJSON(t, []string{"module", "--db", dbPath}, installRequest, &installResult)
	if code != 0 {
		t.Fatalf("install exit = %d stderr=%s", code, stderr)
	}
	if installResult.Rejected ||
		installResult.Module == nil ||
		installResult.Module.Provider != "ollama" ||
		installResult.Module.ManifestPath != filepath.ToSlash(manifestRel) {
		t.Fatalf("install result = %+v", installResult)
	}

	var listResult moduleTaskResult
	code, stderr = runJSON(t, []string{"module", "--db", dbPath}, `{"action":"list_modules"}`, &listResult)
	if code != 0 {
		t.Fatalf("list exit = %d stderr=%s", code, stderr)
	}
	if len(listResult.Modules) != 1 || listResult.Modules[0].ManifestPath != filepath.ToSlash(manifestRel) {
		t.Fatalf("list result = %+v", listResult)
	}
}

func TestResolvedVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		linkerVersion string
		info          *debug.BuildInfo
		ok            bool
		want          string
	}{
		{
			name:          "linker version wins",
			linkerVersion: "v0.1.0",
			info:          &debug.BuildInfo{Main: debug.Module{Version: "v0.0.9"}},
			ok:            true,
			want:          "v0.1.0",
		},
		{
			name: "module version",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v0.1.0"}},
			ok:   true,
			want: "v0.1.0",
		},
		{
			name: "development fallback",
			info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:   true,
			want: "dev",
		},
		{
			name: "missing build info fallback",
			want: "dev",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := resolvedVersion(tt.linkerVersion, tt.info, tt.ok); got != tt.want {
				t.Fatalf("resolvedVersion = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunnerDocumentAndRetrievalJSONRoundTrip(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	createRequest := `{"action":"create_document","document":{"path":"notes/runner.md","title":"Runner","body":"# Runner\n\n## Summary\nOpenClerk runner note.\n"}}`
	var createResult runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, createRequest, &createResult)
	if code != 0 {
		t.Fatalf("create exit = %d stderr=%s", code, stderr)
	}
	if createResult.Document == nil || createResult.Document.DocID == "" {
		t.Fatalf("create result = %+v", createResult)
	}

	searchRequest := `{"action":"search","search":{"text":"runner","limit":10}}`
	var searchResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, searchRequest, &searchResult)
	if code != 0 {
		t.Fatalf("search exit = %d stderr=%s", code, stderr)
	}
	if searchResult.Search == nil || len(searchResult.Search.Hits) == 0 {
		t.Fatalf("search result = %+v", searchResult)
	}

	capturePath := filepath.Join(filepath.Dir(dbPath), "retrieval-eval.jsonl")
	captureRequestBytes, err := json.Marshal(runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionRetrievalEvalCapture,
		RetrievalEval: runner.RetrievalEvalOptions{
			Action:      runner.RetrievalTaskActionSearch,
			CapturePath: capturePath,
			Search:      runner.SearchOptions{Text: "runner", Limit: 10},
		},
	})
	if err != nil {
		t.Fatalf("marshal retrieval eval capture request: %v", err)
	}
	var captureResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, string(captureRequestBytes), &captureResult)
	if code != 0 {
		t.Fatalf("retrieval eval capture exit = %d stderr=%s", code, stderr)
	}
	if captureResult.RetrievalEvalCapture == nil ||
		captureResult.RetrievalEvalCapture.WriteStatus != "local_eval_artifact_appended" ||
		captureResult.RetrievalEvalCapture.AgentHandoff == nil {
		t.Fatalf("retrieval eval capture result = %+v", captureResult)
	}

	replayRequestBytes, err := json.Marshal(runner.RetrievalTaskRequest{
		Action:          runner.RetrievalTaskActionRetrievalEvalReplay,
		RetrievalReplay: runner.RetrievalReplayOptions{CapturePath: capturePath, Limit: 10},
	})
	if err != nil {
		t.Fatalf("marshal retrieval eval replay request: %v", err)
	}
	var replayResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, string(replayRequestBytes), &replayResult)
	if code != 0 {
		t.Fatalf("retrieval eval replay exit = %d stderr=%s", code, stderr)
	}
	if replayResult.RetrievalEvalReplay == nil ||
		replayResult.RetrievalEvalReplay.ComparedCases != 1 ||
		replayResult.RetrievalEvalReplay.AgentHandoff == nil {
		t.Fatalf("retrieval eval replay result = %+v", replayResult)
	}

	searchDiagnosticsRequest := `{"action":"search_diagnostics_report","search_diagnostics":{"query":"runner","intent":"semantic recall","limit":10,"provider":"ollama"}}`
	var searchDiagnosticsResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, searchDiagnosticsRequest, &searchDiagnosticsResult)
	if code != 0 {
		t.Fatalf("search diagnostics exit = %d stderr=%s", code, stderr)
	}
	if searchDiagnosticsResult.SearchDiagnostics == nil ||
		searchDiagnosticsResult.SearchDiagnostics.RecommendedAction == "" ||
		!searchDiagnosticsResult.SearchDiagnostics.NoDefaultRankingChange ||
		searchDiagnosticsResult.SearchDiagnostics.AgentHandoff == nil {
		t.Fatalf("search diagnostics result = %+v", searchDiagnosticsResult)
	}

	maintenanceRequest := `{"action":"maintenance_report","maintenance":{"query":"runner","path_prefix":"notes/","limit":20}}`
	var maintenanceResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, maintenanceRequest, &maintenanceResult)
	if code != 0 {
		t.Fatalf("maintenance report exit = %d stderr=%s", code, stderr)
	}
	if maintenanceResult.Maintenance == nil ||
		maintenanceResult.Maintenance.WriteStatus != "read_only_no_repair" ||
		maintenanceResult.Maintenance.AgentHandoff == nil {
		t.Fatalf("maintenance report result = %+v", maintenanceResult)
	}

	recallRequest := `{"action":"memory_router_recall_report","memory_router_recall":{"query":"memory router temporal recall session promotion feedback weighting routing canonical docs","limit":10}}`
	var recallResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, recallRequest, &recallResult)
	if code != 0 {
		t.Fatalf("memory/router recall report exit = %d stderr=%s", code, stderr)
	}
	if recallResult.MemoryRouterRecall == nil || !strings.Contains(recallResult.MemoryRouterRecall.ValidationBoundaries, "missing evidence") {
		t.Fatalf("memory/router recall report result = %+v", recallResult)
	}

	taggedRequest := `{"action":"create_document","document":{"path":"notes/tagged-runner.md","title":"Tagged Runner","body":"---\ntag: runner-tag\n---\n# Tagged Runner\n\n## Summary\nTagged runner evidence.\n"}}`
	var taggedCreate runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, taggedRequest, &taggedCreate)
	if code != 0 {
		t.Fatalf("create tagged exit = %d stderr=%s", code, stderr)
	}
	tagSearchRequest := `{"action":"search","search":{"text":"Tagged runner evidence","tag":"runner-tag","limit":10}}`
	var tagSearchResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, tagSearchRequest, &tagSearchResult)
	if code != 0 {
		t.Fatalf("tag search exit = %d stderr=%s", code, stderr)
	}
	if tagSearchResult.Search == nil || len(tagSearchResult.Search.Hits) != 1 || tagSearchResult.Search.Hits[0].Citations[0].Path != "notes/tagged-runner.md" {
		t.Fatalf("tag search result = %+v", tagSearchResult.Search)
	}
	tagListRequest := `{"action":"list_documents","list":{"path_prefix":"notes/","tag":"runner-tag","limit":20}}`
	var tagListResult runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, tagListRequest, &tagListResult)
	if code != 0 {
		t.Fatalf("tag list exit = %d stderr=%s", code, stderr)
	}
	if len(tagListResult.Documents) != 1 || tagListResult.Documents[0].Path != "notes/tagged-runner.md" {
		t.Fatalf("tag list result = %+v", tagListResult.Documents)
	}
	emptyTagRequest := `{"action":"search","search":{"text":"runner","tag":""}}`
	var emptyTagResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, emptyTagRequest, &emptyTagResult)
	if code != 0 {
		t.Fatalf("empty tag exit = %d stderr=%s", code, stderr)
	}
	if !emptyTagResult.Rejected || emptyTagResult.RejectionReason != "search.tag must be non-empty" {
		t.Fatalf("empty tag result = %+v", emptyTagResult)
	}

	serviceRequest := `{"action":"create_document","document":{"path":"records/services/openclerk-runner.md","title":"OpenClerk runner","body":"---\nservice_id: openclerk-runner\nservice_name: OpenClerk runner\nservice_status: active\nservice_owner: runner\nservice_interface: JSON runner\n---\n# OpenClerk runner\n\n## Summary\nProduction service.\n"}}`
	var serviceCreate runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, serviceRequest, &serviceCreate)
	if code != 0 {
		t.Fatalf("create service exit = %d stderr=%s", code, stderr)
	}

	servicesRequest := `{"action":"services_lookup","services":{"text":"OpenClerk runner","interface":"JSON runner","limit":10}}`
	var servicesResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, servicesRequest, &servicesResult)
	if code != 0 {
		t.Fatalf("services exit = %d stderr=%s", code, stderr)
	}
	if servicesResult.Services == nil || len(servicesResult.Services.Services) != 1 {
		t.Fatalf("services result = %+v", servicesResult)
	}

	serviceDetailRequest := `{"action":"service_record","service_id":"openclerk-runner"}`
	var serviceDetail runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, serviceDetailRequest, &serviceDetail)
	if code != 0 {
		t.Fatalf("service detail exit = %d stderr=%s", code, stderr)
	}
	if serviceDetail.Service == nil || serviceDetail.Service.Interface != "JSON runner" {
		t.Fatalf("service detail = %+v", serviceDetail)
	}

	structuredStoreRequest := `{"action":"structured_store_report","structured_store":{"domain":"services","query":"OpenClerk runner","interface":"JSON runner","limit":10}}`
	var structuredStoreResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, structuredStoreRequest, &structuredStoreResult)
	if code != 0 {
		t.Fatalf("structured store exit = %d stderr=%s", code, stderr)
	}
	if structuredStoreResult.StructuredStore == nil ||
		structuredStoreResult.StructuredStore.Services == nil ||
		structuredStoreResult.StructuredStore.Projections == nil ||
		structuredStoreResult.StructuredStore.AgentHandoff == nil {
		t.Fatalf("structured store result = %+v", structuredStoreResult)
	}

	workflowGuideRequest := `{"action":"workflow_guide_report","workflow_guide":{"intent":"Which surface should handle duplicate update versus new?"}}`
	var workflowGuideResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, workflowGuideRequest, &workflowGuideResult)
	if code != 0 {
		t.Fatalf("workflow guide exit = %d stderr=%s", code, stderr)
	}
	if workflowGuideResult.WorkflowGuide == nil ||
		workflowGuideResult.WorkflowGuide.RecommendedSurface != "duplicate_candidate_report" ||
		workflowGuideResult.WorkflowGuide.AgentHandoff == nil {
		t.Fatalf("workflow guide result = %+v", workflowGuideResult)
	}

	decisionRequest := `{"action":"create_document","document":{"path":"docs/architecture/runner-decision.md","title":"Runner decision","body":"---\ndecision_id: adr-runner\ndecision_title: Use JSON runner\ndecision_status: accepted\ndecision_scope: agentops\ndecision_owner: platform\ndecision_date: 2026-04-22\n---\n# Runner decision\n\n## Summary\nUse the JSON runner.\n"}}`
	var decisionCreate runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, decisionRequest, &decisionCreate)
	if code != 0 {
		t.Fatalf("create decision exit = %d stderr=%s", code, stderr)
	}

	decisionsRequest := `{"action":"decisions_lookup","decisions":{"text":"JSON runner","status":"accepted","scope":"agentops","owner":"platform","limit":10}}`
	var decisionsResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, decisionsRequest, &decisionsResult)
	if code != 0 {
		t.Fatalf("decisions exit = %d stderr=%s", code, stderr)
	}
	if decisionsResult.Decisions == nil || len(decisionsResult.Decisions.Decisions) != 1 {
		t.Fatalf("decisions result = %+v", decisionsResult)
	}

	decisionDetailRequest := `{"action":"decision_record","decision_id":"adr-runner"}`
	var decisionDetail runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, decisionDetailRequest, &decisionDetail)
	if code != 0 {
		t.Fatalf("decision detail exit = %d stderr=%s", code, stderr)
	}
	if decisionDetail.Decision == nil || decisionDetail.Decision.Status != "accepted" {
		t.Fatalf("decision detail = %+v", decisionDetail)
	}

	layoutRequest := `{"action":"inspect_layout"}`
	var layoutResult runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, layoutRequest, &layoutResult)
	if code != 0 {
		t.Fatalf("inspect layout exit = %d stderr=%s", code, stderr)
	}
	if layoutResult.Layout == nil || !layoutResult.Layout.Valid || layoutResult.Layout.Mode != "convention_first" {
		t.Fatalf("layout result = %+v", layoutResult)
	}

	compileRequestBytes, err := json.Marshal(runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:          "synthesis/runner-workflow.md",
			Title:         "Runner Workflow",
			SourceRefs:    []string{"sources/runner-current.md", "sources/runner-old.md"},
			BodyFacts:     []string{"Compiled runner workflow evidence."},
			FreshnessNote: "Checked current runner sources.",
			Mode:          "create_or_update",
		},
	})
	if err != nil {
		t.Fatalf("marshal compile request: %v", err)
	}
	var compileResult runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, string(compileRequestBytes), &compileResult)
	if code != 0 {
		t.Fatalf("compile synthesis exit = %d stderr=%s", code, stderr)
	}
	if compileResult.CompileSynthesis == nil ||
		compileResult.CompileSynthesis.WriteStatus != "created" ||
		compileResult.CompileSynthesis.AgentHandoff == nil ||
		compileResult.AgentHandoff == nil ||
		!strings.Contains(compileResult.CompileSynthesis.FinalAnswer, "duplicate_status=no_duplicate_created") ||
		!strings.Contains(compileResult.CompileSynthesis.AgentHandoff.AnswerSummary, "compile_synthesis created synthesis/runner-workflow.md") {
		t.Fatalf("compile synthesis result = %+v", compileResult)
	}

	sourceAuditRequest := `{"action":"source_audit_report","source_audit":{"query":"runner source","target_path":"synthesis/runner-workflow.md","mode":"explain","limit":10}}`
	var sourceAuditResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, sourceAuditRequest, &sourceAuditResult)
	if code != 0 {
		t.Fatalf("source audit exit = %d stderr=%s", code, stderr)
	}
	if sourceAuditResult.SourceAudit == nil ||
		sourceAuditResult.SourceAudit.Mode != "explain" ||
		sourceAuditResult.SourceAudit.RepairApplied ||
		sourceAuditResult.SourceAudit.AgentHandoff == nil {
		t.Fatalf("source audit result = %+v", sourceAuditResult)
	}

	evidenceRequest := `{"action":"evidence_bundle_report","evidence_bundle":{"query":"JSON runner","decision_id":"adr-runner","ref_kind":"document","ref_id":"` + decisionCreate.Document.DocID + `","projection":"decisions","limit":10}}`
	var evidenceResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, evidenceRequest, &evidenceResult)
	if code != 0 {
		t.Fatalf("evidence bundle exit = %d stderr=%s", code, stderr)
	}
	if evidenceResult.EvidenceBundle == nil ||
		evidenceResult.EvidenceBundle.Search == nil ||
		evidenceResult.EvidenceBundle.Decision == nil ||
		evidenceResult.EvidenceBundle.Provenance == nil ||
		evidenceResult.EvidenceBundle.Projections == nil ||
		evidenceResult.EvidenceBundle.AgentHandoff == nil {
		t.Fatalf("evidence bundle result = %+v", evidenceResult)
	}
}

func TestRunnerDocumentSourceURLUpdateStaleImpactJSON(t *testing.T) {
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "web", "runner-product.html")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir web fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte(`<!doctype html><html><head><title>Runner Web Title</title></head><body><h1>Runner Web Title</h1><p>Initial CLI runner evidence.</p></body></html>`), 0o644); err != nil {
		t.Fatalf("write web fixture: %v", err)
	}
	t.Setenv("OPENCLERK_ENABLE_EVAL_SOURCE_FIXTURES", "1")
	t.Setenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT", fixtureRoot)

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	sourceURL := "http://openclerk-eval.local/web/runner-product.html"
	createRequest := `{"action":"ingest_source_url","source":{"url":"` + sourceURL + `","path_hint":"sources/web/cli-runner-product.md"}}`
	var createResult runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, createRequest, &createResult)
	if code != 0 {
		t.Fatalf("create source exit = %d stderr=%s", code, stderr)
	}
	if createResult.Ingestion == nil || createResult.Ingestion.UpdateStatus != "" {
		t.Fatalf("create ingestion = %+v", createResult.Ingestion)
	}

	createSynthesisRequest := `{"action":"create_document","document":{"path":"synthesis/cli-web-runner.md","title":"CLI Web Runner Synthesis","body":"---\ntype: synthesis\nsource_refs: sources/web/cli-runner-product.md\n---\n# CLI Web Runner Synthesis\n\n## Summary\nInitial CLI runner evidence.\n"}}`
	var synthesisResult runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, createSynthesisRequest, &synthesisResult)
	if code != 0 {
		t.Fatalf("create synthesis exit = %d stderr=%s", code, stderr)
	}

	if err := os.WriteFile(fixturePath, []byte(`<!doctype html><html><head><title>Runner Web Title Updated</title></head><body><h1>Runner Web Title Updated</h1><p>Updated CLI runner evidence.</p></body></html>`), 0o644); err != nil {
		t.Fatalf("write updated web fixture: %v", err)
	}
	updateRequest := `{"action":"ingest_source_url","source":{"url":"` + sourceURL + `","path_hint":"sources/web/cli-runner-product.md","source_type":"web","mode":"update"}}`
	var stdout bytes.Buffer
	var stderrBuffer bytes.Buffer
	code = run([]string{"document", "--db", dbPath}, strings.NewReader(updateRequest), &stdout, &stderrBuffer)
	if code != 0 {
		t.Fatalf("update source exit = %d stderr=%s", code, stderrBuffer.String())
	}
	var updateResult runner.DocumentTaskResult
	if err := json.Unmarshal(stdout.Bytes(), &updateResult); err != nil {
		t.Fatalf("decode update stdout %q: %v", stdout.String(), err)
	}
	if updateResult.Ingestion == nil ||
		updateResult.Ingestion.UpdateStatus != "changed" ||
		updateResult.Ingestion.NormalizedSourceURL != sourceURL ||
		updateResult.Ingestion.SourceDocID != createResult.Ingestion.DocID ||
		updateResult.Ingestion.PreviousSHA256 != createResult.Ingestion.SHA256 ||
		updateResult.Ingestion.NewSHA256 == createResult.Ingestion.SHA256 ||
		updateResult.Ingestion.Changed == nil || !*updateResult.Ingestion.Changed ||
		updateResult.Ingestion.SynthesisRepaired == nil || *updateResult.Ingestion.SynthesisRepaired {
		t.Fatalf("update ingestion = %+v", updateResult.Ingestion)
	}
	for _, want := range []string{`"update_status":"changed"`, `"normalized_source_url":"` + sourceURL + `"`, `"source_doc_id":"`, `"previous_sha256":"`, `"new_sha256":"`, `"changed":true`, `"stale_dependents":[`, `"projection_refs":[`, `"provenance_refs":[`, `"synthesis_repaired":false`, `"no_repair_warning":"`} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("update stdout missing %s: %s", want, stdout.String())
		}
	}
}

func TestRunnerValidationRejectionDoesNotCreateDatabase(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	request := `{"action":"create_document","document":{"title":"Missing path","body":"# Missing path\n"}}`
	var result runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, request, &result)
	if code != 0 {
		t.Fatalf("exit = %d stderr=%s", code, stderr)
	}
	if !result.Rejected || result.RejectionReason == "" {
		t.Fatalf("result = %+v, want rejection", result)
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); !os.IsNotExist(err) {
		t.Fatalf("data dir exists after rejected request: %v", err)
	}
}

func TestRunnerRejectsInvalidCreateFrontmatter(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	request := `{"action":"create_document","document":{"path":"sources/uploaded-pdf.md","title":"Uploaded PDF","body":"---\ntype: source\nmodality: pdf\n---\n# Uploaded PDF\n\n## Summary\nExtracted note.\n"}}`
	var result runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, request, &result)
	if code != 0 {
		t.Fatalf("exit = %d stderr=%s", code, stderr)
	}
	if !result.Rejected || !strings.Contains(result.RejectionReason, "modality") || !strings.Contains(result.RejectionReason, "markdown") {
		t.Fatalf("result = %+v, want modality rejection", result)
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); !os.IsNotExist(err) {
		t.Fatalf("data dir exists after rejected request: %v", err)
	}
}

func TestRunnerErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		args   []string
		input  string
		want   int
		stderr string
	}{
		{name: "unknown command", args: []string{"unknown"}, input: `{}`, want: 2, stderr: "unknown openclerk command"},
		{name: "bad json", args: []string{"document"}, input: `{`, want: 1, stderr: "decode document request"},
		{name: "multiple json", args: []string{"document"}, input: `{} {}`, want: 1, stderr: "multiple JSON values"},
		{name: "unknown json field", args: []string{"document"}, input: `{"action":"validate","extra":true}`, want: 1, stderr: "unknown field"},
		{name: "unknown config json field", args: []string{"config"}, input: `{"action":"inspect_config","extra":true}`, want: 1, stderr: "unknown field"},
		{name: "unknown list json field", args: []string{"document"}, input: `{"action":"list_documents","list":{"path_prefix":"notes/","tga":"account-renewal"}}`, want: 1, stderr: "unknown field"},
		{name: "unknown search json field", args: []string{"retrieval"}, input: `{"action":"search","search":{"text":"renewal","tga":"account-renewal"}}`, want: 1, stderr: "unknown field"},
		{name: "unexpected arg", args: []string{"retrieval", "extra"}, input: `{}`, want: 2, stderr: "unexpected positional arguments"},
		{name: "retired embedding provider flag", args: []string{"document", "--embedding-provider", "local"}, input: `{}`, want: 2, stderr: "embedding-provider"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := run(tt.args, strings.NewReader(tt.input), &stdout, &stderr)
			if code != tt.want {
				t.Fatalf("exit = %d, want %d; stderr=%s", code, tt.want, stderr.String())
			}
			if !strings.Contains(stderr.String(), tt.stderr) {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tt.stderr)
			}
		})
	}
}

func TestRunnerRuntimeErrorExitsNonZero(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	request := `{"action":"get_document","doc_id":"missing"}`
	var result runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, request, &result)
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero")
	}
	if !strings.Contains(stderr, "run document task") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestRunnerConfigErrorIsActionable(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatalf("create db dir: %v", err)
	}
	if err := os.WriteFile(dbPath, []byte("not a sqlite database"), 0o644); err != nil {
		t.Fatalf("write corrupt db: %v", err)
	}

	var result runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, `{"action":"resolve_paths"}`, &result)
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero")
	}
	if !strings.Contains(stderr, dbPath) {
		t.Fatalf("stderr = %q, want database path %q", stderr, dbPath)
	}
	if !strings.Contains(stderr, "resolve_paths") && !strings.Contains(stderr, "inspect_layout") {
		t.Fatalf("stderr = %q, want diagnostic action hint", stderr)
	}
	if strings.Contains(stderr, "upsert runtime config") ||
		strings.Contains(stderr, "initialize runtime config") {
		t.Fatalf("stderr leaked raw runtime config message: %q", stderr)
	}
}

func TestRunnerDBFlag(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "custom", "openclerk.sqlite")
	request := `{"action":"resolve_paths"}`
	var result runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, request, &result)
	if code != 0 {
		t.Fatalf("exit = %d stderr=%s", code, stderr)
	}
	if result.Paths == nil || result.Paths.DatabasePath != dbPath {
		t.Fatalf("paths = %+v, want db %q", result.Paths, dbPath)
	}
	if result.Paths.VaultRoot != filepath.Join(filepath.Dir(dbPath), "vault") {
		t.Fatalf("paths = %+v, want default sibling vault", result.Paths)
	}
}

func TestRunnerInitBindsVaultRoot(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "custom", "openclerk.sqlite")
	vaultRoot := filepath.Join(t.TempDir(), "wiki")
	var initResult struct {
		Paths runner.Paths `json:"paths"`
	}
	code, stderr := runJSON(t, []string{"init", "--db", dbPath, "--vault-root", vaultRoot}, "", &initResult)
	if code != 0 {
		t.Fatalf("init exit = %d stderr=%s", code, stderr)
	}
	if initResult.Paths.DatabasePath != dbPath || initResult.Paths.VaultRoot != vaultRoot {
		t.Fatalf("init paths = %+v", initResult.Paths)
	}

	var inspectResult runner.ConfigTaskResult
	code, stderr = runJSON(t, []string{"config", "--db", dbPath}, `{"action":"inspect_config"}`, &inspectResult)
	if code != 0 {
		t.Fatalf("inspect config exit = %d stderr=%s", code, stderr)
	}
	if inspectResult.Storage == nil || inspectResult.Storage.VaultRoot != vaultRoot {
		t.Fatalf("inspect storage = %+v, want vault %q", inspectResult.Storage, vaultRoot)
	}

	var rejectedStorageWrite runner.ConfigTaskResult
	code, stderr = runJSON(t, []string{"config", "--db", dbPath}, `{"action":"configure_storage"}`, &rejectedStorageWrite)
	if code != 0 {
		t.Fatalf("configure_storage exit = %d stderr=%s", code, stderr)
	}
	if !rejectedStorageWrite.Rejected || !strings.Contains(rejectedStorageWrite.RejectionReason, "unsupported config action") {
		t.Fatalf("configure_storage result = %+v", rejectedStorageWrite)
	}
	code, stderr = runJSON(t, []string{"config", "--db", dbPath}, `{"action":"configure_storage","storage":{"vault_root":"`+filepath.ToSlash(filepath.Join(t.TempDir(), "forbidden"))+`"}}`, nil)
	if code == 0 || !strings.Contains(stderr, "unknown field") {
		t.Fatalf("storage write shape exit = %d stderr=%s", code, stderr)
	}
	code, stderr = runJSON(t, []string{"config", "--db", dbPath}, `{"action":"inspect_config"}`, &inspectResult)
	if code != 0 {
		t.Fatalf("inspect after rejected storage write exit = %d stderr=%s", code, stderr)
	}
	if inspectResult.Storage == nil || inspectResult.Storage.VaultRoot != vaultRoot {
		t.Fatalf("inspect storage after rejected write = %+v, want vault %q", inspectResult.Storage, vaultRoot)
	}

	reboundVaultRoot := filepath.Join(t.TempDir(), "rebound-wiki")
	code, stderr = runJSON(t, []string{"init", "--db", dbPath, "--vault-root", reboundVaultRoot}, "", &initResult)
	if code != 0 {
		t.Fatalf("rebind init exit = %d stderr=%s", code, stderr)
	}
	if initResult.Paths.DatabasePath != dbPath || initResult.Paths.VaultRoot != reboundVaultRoot {
		t.Fatalf("rebind init paths = %+v", initResult.Paths)
	}

	request := `{"action":"resolve_paths"}`
	var result runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, request, &result)
	if code != 0 {
		t.Fatalf("resolve exit = %d stderr=%s", code, stderr)
	}
	if result.Paths == nil || result.Paths.VaultRoot != reboundVaultRoot {
		t.Fatalf("paths = %+v, want vault %q", result.Paths, reboundVaultRoot)
	}
	code, stderr = runJSON(t, []string{"config", "--db", dbPath}, `{"action":"inspect_config"}`, &inspectResult)
	if code != 0 {
		t.Fatalf("inspect rebound config exit = %d stderr=%s", code, stderr)
	}
	if inspectResult.Storage == nil || inspectResult.Storage.VaultRoot != reboundVaultRoot {
		t.Fatalf("inspect rebound storage = %+v, want vault %q", inspectResult.Storage, reboundVaultRoot)
	}
}

func runJSON(t *testing.T, args []string, input string, output any) (int, string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(args, strings.NewReader(input), &stdout, &stderr)
	if output != nil && stdout.Len() > 0 {
		if err := json.Unmarshal(stdout.Bytes(), output); err != nil {
			t.Fatalf("decode stdout %q: %v", stdout.String(), err)
		}
	}
	return code, stderr.String()
}

func inspectHasNextAction(result inspectEnvelope, label string) bool {
	for _, action := range result.Result.RecommendedNextActions {
		if action.Label == label {
			return true
		}
	}
	return false
}

func mustMarshalInspect(t *testing.T, result inspectEnvelope) string {
	t.Helper()
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal inspect result: %v", err)
	}
	return string(data)
}

func writeUnboundSQLiteDatabase(t *testing.T, dbPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatalf("create db dir: %v", err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()
	if _, err := db.Exec(`CREATE TABLE scratch (id TEXT PRIMARY KEY);`); err != nil {
		t.Fatalf("create scratch table: %v", err)
	}
}

func runGitForTest(t *testing.T, workdir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", workdir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
}

func fileModTime(t *testing.T, path string) time.Time {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return info.ModTime()
}

func writeCLISemanticModuleManifest(t *testing.T, dir string, provider string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create manifest dir: %v", err)
	}
	path := filepath.Join(dir, "module.json")
	manifest := map[string]any{
		"schema_version": "openclerk-module.v1",
		"module": map[string]any{
			"name":    provider + "-embeddings",
			"version": "0.1.0",
			"kind":    "embedding_provider",
		},
		"provides": []map[string]any{{
			"type": "command",
			"name": "semantic-retrieval-adapter search",
		}},
		"requires": map[string]any{
			"tools": []string{"semantic-retrieval-adapter"},
		},
		"authority": map[string]any{
			"default":        "read_only",
			"durable_writes": "forbidden",
			"forbidden":      []string{"write_documents", "change_openclerk_search_default"},
		},
		"release": map[string]any{
			"status": "supported_optional_module",
		},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}

func writeCLIOCRModuleManifest(t *testing.T, dir string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create OCR manifest dir: %v", err)
	}
	path := filepath.Join(dir, "module.json")
	manifest := map[string]any{
		"schema_version": "openclerk-module.v1",
		"module": map[string]any{
			"name":    "tesseract-ocr",
			"version": "0.1.0",
			"kind":    "ocr_provider",
		},
		"provides": []map[string]any{{
			"type": "command",
			"name": "tesseract ocr",
		}, {
			"type": "command",
			"name": "ocrmypdf ocr",
		}},
		"requires": map[string]any{
			"tools": []string{"tesseract", "ocrmypdf"},
		},
		"authority": map[string]any{
			"default":        "read_only",
			"durable_writes": "forbidden",
			"forbidden":      []string{"write_documents", "hidden_cloud_egress"},
		},
		"release": map[string]any{
			"status": "supported_optional_module",
		},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal OCR manifest: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write OCR manifest: %v", err)
	}
	return path
}

func writeCLIExecutable(t *testing.T, name string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf '%s\\n' "+name+"\n"), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}
	return path
}
