package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runner"
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

func TestSubcommandHelpShowsPromotedWorkflowActions(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		args []string
		want []string
	}{
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
				"body_facts",
				"agent_handoff",
				"mode defaults to create_or_update",
			},
		},
		{
			name: "retrieval",
			args: []string{"retrieval", "--help"},
			want: []string{"source_audit_report", "evidence_bundle_report", "duplicate_candidate_report", "memory_router_recall_report", "hybrid_retrieval_report", "agent_handoff", "Read-only"},
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
