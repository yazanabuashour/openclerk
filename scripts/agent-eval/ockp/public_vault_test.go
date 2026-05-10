package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsePublicVaultConfigDefaults(t *testing.T) {
	config, err := parsePublicVaultConfig([]string{"kubernetes-docs", "--run-root", "run", "--parallel", "2"}, io.Discard)
	if err != nil {
		t.Fatalf("parse public vault config: %v", err)
	}
	if config.Mode != publicVaultModeKubernetesDocs ||
		config.Parallel != 2 ||
		config.PublicRepoURL != defaultPublicVaultRepoURL ||
		config.PublicRepoRef != defaultPublicVaultRepoRef ||
		config.PublicSubdir != defaultPublicVaultSubdir ||
		config.TaskManifestPath != defaultPublicVaultManifest ||
		config.ReportName != "ockp-public-vault-kubernetes-docs" {
		t.Fatalf("config = %+v", config)
	}
	_, err = parsePublicVaultConfig([]string{"unknown"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "unsupported public-vault mode") {
		t.Fatalf("unsupported mode error = %v", err)
	}
	_, err = parsePublicVaultConfig([]string{"kubernetes-docs", "--subdir", "../docs"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "--subdir") {
		t.Fatalf("bad subdir error = %v", err)
	}
	_, err = parsePublicVaultConfig([]string{"kubernetes-docs", "--subdir", "content/en/docs/../../.."}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "--subdir") {
		t.Fatalf("embedded traversal subdir error = %v", err)
	}
}

func TestValidatePublicVaultTaskManifest(t *testing.T) {
	manifest := validPublicVaultTaskManifestForTest()
	if err := validatePublicVaultTaskManifest(manifest); err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}
	manifest.Tasks[0].Prompt = ""
	if err := validatePublicVaultTaskManifest(manifest); err == nil || !strings.Contains(err.Error(), "prompt") {
		t.Fatalf("empty prompt error = %v", err)
	}
	manifest = validPublicVaultTaskManifestForTest()
	manifest.Tasks[0].Class = "unknown"
	if err := validatePublicVaultTaskManifest(manifest); err == nil || !strings.Contains(err.Error(), "unsupported class") {
		t.Fatalf("unsupported class error = %v", err)
	}
	manifest = validPublicVaultTaskManifestForTest()
	manifest.Tasks = manifest.Tasks[:7]
	if err := validatePublicVaultTaskManifest(manifest); err == nil || !strings.Contains(err.Error(), "exactly 8") {
		t.Fatalf("missing row error = %v", err)
	}
}

func TestMaterializePublicVaultCorpusCopiesMarkdownToSourcesPrefix(t *testing.T) {
	ctx := context.Background()
	sourceRoot := seedPublicVaultSourceForTest(t)
	config := publicVaultConfig{
		RunRoot:       t.TempDir(),
		PublicRepoURL: sourceRoot,
		PublicRepoRef: "local-test",
		PublicSubdir:  "content/en/docs",
	}
	corpus, err := materializePublicVaultCorpus(ctx, config)
	if err != nil {
		t.Fatalf("materialize public vault corpus: %v", err)
	}
	if corpus.MarkdownFiles != 3 || corpus.VaultPrefix != "sources/kubernetes/website/content/en/docs" {
		t.Fatalf("corpus = %+v", corpus)
	}
	if _, err := os.Stat(filepath.Join(config.RunRoot, "public-vault-copy", "sources", "kubernetes", "website", "content", "en", "docs", "concepts", "workloads", "controllers", "deployment.md")); err != nil {
		t.Fatalf("materialized source missing: %v", err)
	}
}

func TestExecutePublicVaultWritesPublicReports(t *testing.T) {
	runRoot := t.TempDir()
	reportDir := t.TempDir()
	sourceRoot := seedPublicVaultSourceForTest(t)
	manifestPath := filepath.Join(t.TempDir(), "tasks.json")
	writePublicVaultManifestForTest(t, manifestPath, validPublicVaultTaskManifestForTest())
	config := publicVaultConfig{
		Mode:             publicVaultModeKubernetesDocs,
		Parallel:         2,
		RunRoot:          runRoot,
		ReportDir:        reportDir,
		ReportName:       "test-public-vault",
		CodexBin:         "codex",
		RepoRoot:         ".",
		CacheMode:        cacheModeIsolated,
		PublicRepoURL:    sourceRoot,
		PublicRepoRef:    "local-test",
		PublicSubdir:     "content/en/docs",
		TaskManifestPath: manifestPath,
	}
	var stdout bytes.Buffer
	if err := executePublicVault(context.Background(), config, &stdout, fakePublicVaultRunner); err != nil {
		t.Fatalf("execute public vault: %v", err)
	}
	jsonPath := filepath.Join(reportDir, "test-public-vault.json")
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read JSON: %v", err)
	}
	var report publicVaultReport
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if !report.Summary.PassesGate || report.Summary.RowsCompleted != 8 || report.Summary.SafetyFailures != 0 || report.Summary.UXDebtRows != 0 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	if report.Corpus.RepoURL != "<local-public-repo>" || strings.Contains(string(content), sourceRoot) {
		t.Fatalf("report leaked local repo path: corpus=%+v", report.Corpus)
	}
	markdown := string(readReportForTest(t, filepath.Join(reportDir, "test-public-vault.md")))
	for _, want := range []string{"Public Kubernetes Docs Vault Trial", "sources/kubernetes/website/content/en/docs", "Passes gate: `true`"} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown missing %q:\n%s", want, markdown)
		}
	}
	if strings.Contains(markdown, runRoot) || strings.Contains(markdown, "events.jsonl") {
		t.Fatalf("markdown leaked local refs:\n%s", markdown)
	}
	if !strings.Contains(stdout.String(), "test-public-vault.json") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestWritePublicVaultMarkdownRejectsLocalPaths(t *testing.T) {
	report := publicVaultReport{
		Metadata: publicVaultReportMetadata{Lane: "public-vault-kubernetes-docs-trial"},
		Corpus: publicVaultCorpus{
			RepoURL:     defaultPublicVaultRepoURL,
			RepoRef:     defaultPublicVaultRepoRef,
			Subdir:      defaultPublicVaultSubdir,
			VaultPrefix: "/tmp/local-copy",
		},
		Summary: publicVaultReportSummary{PassesGate: true},
	}
	err := writePublicVaultMarkdownReport(filepath.Join(t.TempDir(), "report.md"), report)
	if err == nil || !strings.Contains(err.Error(), "machine-local path") {
		t.Fatalf("expected local path rejection, got %v", err)
	}
}

func TestPublicVaultMetricsDoNotTreatKnowledgeTextAsBypass(t *testing.T) {
	m := emptyMetrics()
	classifyCommand(`printf '%s' '{"action":"search","search":{"text":"Ingress exposes HTTP routes and Service can select Pods","limit":10}}' | openclerk retrieval`, &m)
	if m.ManualHTTPFetch || m.DirectSQLiteAccess {
		t.Fatalf("knowledge text was classified as bypass: %+v", m)
	}
	classifyCommand(`sqlite3 openclerk.sqlite 'select * from documents'`, &m)
	if !m.DirectSQLiteAccess {
		t.Fatalf("sqlite command was not classified as bypass: %+v", m)
	}
	m = emptyMetrics()
	classifyCommand(`http GET https://example.com`, &m)
	if !m.ManualHTTPFetch {
		t.Fatalf("http command was not classified as manual fetch: %+v", m)
	}
}

func fakePublicVaultRunner(_ context.Context, _ publicVaultConfig, job publicVaultJob, _ cacheConfig, _ publicVaultCorpus) publicVaultJobResult {
	m := metrics{
		AssistantCalls:    1,
		ToolCalls:         1,
		CommandExecutions: 1,
		EventTypeCounts:   map[string]int{},
	}
	switch job.Task.Class {
	case "source_discovery":
		m.SourceDiscoveryReportUsed = true
	case "cited_search_answer", "cross_source_comparison", "rbac_navigation":
		m.SearchUsed = true
	case "synthesis_create_update":
		m.CompileSynthesisUsed = true
	case "provenance_freshness":
		m.EvidenceBundleReportUsed = true
	case "decision_like_lookup":
		m.DecisionLookupReportUsed = true
	case "stale_duplicate_detection":
		m.SearchUsed = true
		m.ProjectionStatesUsed = true
	}
	return publicVaultJobResult{
		Index:        job.Index,
		Class:        job.Task.Class,
		Status:       "completed",
		WallSeconds:  1,
		Metrics:      m,
		Verification: verificationResult{Passed: true, DatabasePass: true, AssistantPass: true},
		RawLogRef:    "<run-root>/task/events.jsonl",
	}
}

func validPublicVaultTaskManifestForTest() publicVaultTaskManifest {
	classes := []string{
		"source_discovery",
		"cited_search_answer",
		"synthesis_create_update",
		"provenance_freshness",
		"decision_like_lookup",
		"stale_duplicate_detection",
		"cross_source_comparison",
		"rbac_navigation",
	}
	tasks := make([]publicVaultTask, 0, len(classes))
	for _, class := range classes {
		tasks = append(tasks, publicVaultTask{
			Class:                 class,
			Prompt:                "public prompt for " + class,
			ExpectedRunnerActions: []string{expectedPublicActionForTest(class)},
			PublicEvidenceRefs:    []string{"sources/kubernetes/website/content/en/docs/example.md"},
		})
	}
	return publicVaultTaskManifest{SchemaVersion: publicVaultTaskSchemaVersion, Tasks: tasks}
}

func expectedPublicActionForTest(class string) string {
	switch class {
	case "source_discovery":
		return "source_discovery_report"
	case "synthesis_create_update":
		return "compile_synthesis"
	case "provenance_freshness":
		return "evidence_bundle_report"
	case "decision_like_lookup":
		return "decision_lookup_report"
	case "stale_duplicate_detection":
		return "projection_states"
	default:
		return "search"
	}
}

func writePublicVaultManifestForTest(t *testing.T, path string, manifest publicVaultTaskManifest) {
	t.Helper()
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

func seedPublicVaultSourceForTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	docs := map[string]string{
		"content/en/docs/concepts/workloads/controllers/deployment.md": "# Deployment\n\nDeployment rollout public docs.\n",
		"content/en/docs/concepts/services-networking/service.md":      "# Service\n\nService exposure public docs.\n",
		"content/en/docs/reference/access-authn-authz/rbac.md":         "# RBAC\n\nRBAC public docs.\n",
	}
	for rel, body := range docs {
		target := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(target, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	return root
}
