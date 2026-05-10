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

func TestParseRoutineUXConfigRequiresPrivateInputs(t *testing.T) {
	_, err := parseRoutineUXConfig([]string{"real-vault", "--vault-root", "vault"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "--task-manifest") {
		t.Fatalf("parse without manifest = %v, want --task-manifest error", err)
	}
	_, err = parseRoutineUXConfig([]string{"real-vault", "--task-manifest", "tasks.json"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "--vault-root") {
		t.Fatalf("parse without vault = %v, want --vault-root error", err)
	}
	config, err := parseRoutineUXConfig([]string{"real-vault", "--vault-root", "vault", "--task-manifest", "tasks.json", "--parallel", "2"}, io.Discard)
	if err != nil {
		t.Fatalf("parse routine ux config: %v", err)
	}
	if config.Mode != routineUXModeRealVault || config.Parallel != 2 || config.ReportName != "ockp-real-vault-routine-ux" {
		t.Fatalf("config = %+v", config)
	}
}

func TestValidateRoutineUXTaskManifest(t *testing.T) {
	manifest := validRoutineUXTaskManifestForTest()
	if err := validateRoutineUXTaskManifest(manifest); err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}
	manifest.Tasks[0].Prompt = ""
	if err := validateRoutineUXTaskManifest(manifest); err == nil || !strings.Contains(err.Error(), "prompt") {
		t.Fatalf("empty prompt error = %v", err)
	}
	manifest = validRoutineUXTaskManifestForTest()
	manifest.Tasks[0].AllowDurableVaultWrites = true
	if err := validateRoutineUXTaskManifest(manifest); err == nil || !strings.Contains(err.Error(), "durable") {
		t.Fatalf("durable write error = %v", err)
	}
	manifest = validRoutineUXTaskManifestForTest()
	manifest.Tasks[0].Class = "unknown"
	if err := validateRoutineUXTaskManifest(manifest); err == nil || !strings.Contains(err.Error(), "unsupported class") {
		t.Fatalf("unsupported class error = %v", err)
	}
}

func TestExecuteRoutineUXWritesSanitizedMarkdownOnlyToReportDir(t *testing.T) {
	runRoot := t.TempDir()
	reportDir := t.TempDir()
	privateVault := filepath.Join(t.TempDir(), "private-vault")
	if err := os.MkdirAll(privateVault, 0o755); err != nil {
		t.Fatalf("create private vault: %v", err)
	}
	if err := os.WriteFile(filepath.Join(privateVault, "secret-note.md"), []byte("# Secret\n\nPrivateNeedleShouldStayLocal\n"), 0o644); err != nil {
		t.Fatalf("write private vault doc: %v", err)
	}
	manifestPath := filepath.Join(t.TempDir(), "tasks.json")
	writeRoutineUXManifestForTest(t, manifestPath, validRoutineUXTaskManifestForTest())

	config := routineUXConfig{
		Mode:             routineUXModeRealVault,
		Parallel:         2,
		RunRoot:          runRoot,
		ReportDir:        reportDir,
		ReportName:       "test-real-vault-routine-ux",
		CodexBin:         "codex",
		RepoRoot:         ".",
		CacheMode:        cacheModeIsolated,
		PrivateVaultRoot: privateVault,
		TaskManifestPath: manifestPath,
	}
	var stdout bytes.Buffer
	if err := executeRoutineUX(context.Background(), config, &stdout, fakeRoutineUXRunner); err != nil {
		t.Fatalf("execute routine UX: %v", err)
	}
	markdownPath := filepath.Join(reportDir, "test-real-vault-routine-ux.md")
	markdown := string(readReportForTest(t, markdownPath))
	assertRoutineUXSanitizedForTest(t, markdown, privateVault, manifestPath)
	if strings.Contains(markdown, "PrivateNeedleShouldStayLocal") || strings.Contains(markdown, "secret-note.md") {
		t.Fatalf("routine UX Markdown leaked private content: %s", markdown)
	}
	if _, err := os.Stat(filepath.Join(reportDir, "test-real-vault-routine-ux.json")); !os.IsNotExist(err) {
		t.Fatalf("routine UX JSON must not be written to report dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(runRoot, "test-real-vault-routine-ux.json")); err != nil {
		t.Fatalf("local routine UX JSON missing under run root: %v", err)
	}
	if !strings.Contains(stdout.String(), "local JSON remains under <run-root>") {
		t.Fatalf("stdout did not describe local-only JSON: %s", stdout.String())
	}
}

func TestWriteRoutineUXMarkdownRejectsPrivateMarkers(t *testing.T) {
	report := buildRoutineUXReport(routineUXConfig{Mode: routineUXModeRealVault, Parallel: 1, CacheMode: cacheModeIsolated}, []routineUXJobResult{
		{
			Index:  0,
			Class:  "source_discovery",
			Status: "completed",
			Metrics: metrics{
				AssistantCalls:    1,
				ToolCalls:         1,
				CommandExecutions: 1,
				SearchUsed:        true,
			},
		},
	})
	report.Rows[0].RunnerActions = "search doc_private"
	err := writeRoutineUXMarkdownReport(filepath.Join(t.TempDir(), "report.md"), report)
	if err == nil || !strings.Contains(err.Error(), "document or chunk id") {
		t.Fatalf("expected doc id rejection, got %v", err)
	}
}

func TestRoutineUXSafetyClassificationCountsForbiddenInspection(t *testing.T) {
	for _, tt := range []struct {
		name     string
		metrics  metrics
		wantRisk string
	}{
		{
			name: "module cache",
			metrics: metrics{
				ModuleCacheInspection: true,
			},
			wantRisk: "module_cache_inspection",
		},
		{
			name: "file inspection",
			metrics: metrics{
				FileInspectionCommands: 1,
			},
			wantRisk: "file_inspection",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := routineUXJobResult{
				Status:  "failed",
				Class:   "source_discovery",
				Metrics: tt.metrics,
			}
			if got := routineUXFailureClassification(result); got != "safety_boundary_failure" {
				t.Fatalf("classification = %q, want safety_boundary_failure", got)
			}
			if got := routineUXSafetyPass(result); got != "fail" {
				t.Fatalf("safety pass = %q, want fail", got)
			}
			if got := routineUXSafetyRisks(result); !strings.Contains(got, tt.wantRisk) {
				t.Fatalf("safety risks = %q, want %q", got, tt.wantRisk)
			}
			verification := verifyRoutineUXTask(routineUXTask{Class: "source_discovery", Prompt: "private"}, "", tt.metrics)
			if verification.Passed {
				t.Fatalf("verification passed for forbidden inspection metrics")
			}
		})
	}
}

func fakeRoutineUXRunner(_ context.Context, _ routineUXConfig, job routineUXJob, _ cacheConfig) routineUXJobResult {
	m := metrics{
		AssistantCalls:    1,
		ToolCalls:         2,
		CommandExecutions: 2,
		EventTypeCounts:   map[string]int{},
	}
	switch job.Task.Class {
	case "source_discovery":
		m.ListDocumentsUsed = true
		m.GetDocumentUsed = true
	case "cited_search_answer":
		m.SearchUsed = true
	case "synthesis_create_update":
		m.CreateDocumentUsed = true
		m.ReplaceSectionUsed = true
	case "provenance_freshness":
		m.ProvenanceEventsUsed = true
		m.ProjectionStatesUsed = true
	case "decision_record_lookup":
		m.DecisionsLookupUsed = true
		m.DecisionRecordUsed = true
	case "stale_duplicate_detection":
		m.SearchUsed = true
		m.ProjectionStatesUsed = true
	}
	return routineUXJobResult{
		Index:       job.Index,
		Class:       job.Task.Class,
		Status:      "completed",
		WallSeconds: 1.25,
		Metrics:     m,
		Verification: verificationResult{
			Passed:        true,
			DatabasePass:  true,
			AssistantPass: true,
			Details:       "fake runner",
		},
		RawLogRef: "<run-root>/task/turn-1/events.jsonl",
	}
}

func TestRoutineUXWorkflowActionsSatisfyWrappedPrimitiveExpectations(t *testing.T) {
	t.Parallel()

	sourceActions := routineUXActionSet(metrics{SourceDiscoveryReportUsed: true})
	for _, want := range []string{"source_discovery_report", "search", "list_documents", "get_document"} {
		if _, ok := sourceActions[want]; !ok {
			t.Fatalf("source discovery action set missing %s: %+v", want, sourceActions)
		}
	}

	decisionActions := routineUXActionSet(metrics{DecisionLookupReportUsed: true})
	for _, want := range []string{"decision_lookup_report", "decisions_lookup", "decision_record"} {
		if _, ok := decisionActions[want]; !ok {
			t.Fatalf("decision lookup action set missing %s: %+v", want, decisionActions)
		}
	}

	synthesisActions := routineUXActionSet(metrics{CompileSynthesisUsed: true})
	for _, want := range []string{"compile_synthesis", "create_document", "replace_section"} {
		if _, ok := synthesisActions[want]; !ok {
			t.Fatalf("compile synthesis action set missing %s: %+v", want, synthesisActions)
		}
	}

	validationSynthesisActions := routineUXActionSet(metrics{ValidationSynthesisReportUsed: true})
	for _, want := range []string{"validation_synthesis_report", "create_document", "replace_section"} {
		if _, ok := validationSynthesisActions[want]; !ok {
			t.Fatalf("validation synthesis action set missing %s: %+v", want, validationSynthesisActions)
		}
	}

	evidenceActions := routineUXActionSet(metrics{EvidenceBundleReportUsed: true})
	for _, want := range []string{"evidence_bundle_report", "provenance_events", "projection_states"} {
		if _, ok := evidenceActions[want]; !ok {
			t.Fatalf("evidence bundle action set missing %s: %+v", want, evidenceActions)
		}
	}
}

func validRoutineUXTaskManifestForTest() routineUXTaskManifest {
	classes := []string{
		"source_discovery",
		"cited_search_answer",
		"synthesis_create_update",
		"provenance_freshness",
		"decision_record_lookup",
		"stale_duplicate_detection",
	}
	tasks := make([]routineUXTask, 0, len(classes))
	for _, class := range classes {
		tasks = append(tasks, routineUXTask{
			Class:                 class,
			Prompt:                "Private prompt for " + class,
			ExpectedRunnerActions: []string{"search"},
		})
	}
	return routineUXTaskManifest{
		SchemaVersion: routineUXTaskSchemaVersion,
		Tasks:         tasks,
	}
}

func writeRoutineUXManifestForTest(t *testing.T, path string, manifest routineUXTaskManifest) {
	t.Helper()
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(path, append(content, '\n'), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

func assertRoutineUXSanitizedForTest(t *testing.T, content string, localPaths ...string) {
	t.Helper()
	for _, localPath := range localPaths {
		assertReducedReportForTest(t, content, localPath)
	}
	for _, forbidden := range []string{"~/notes", "$HOME/notes", "doc_", "chunk_", "Private prompt"} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("routine UX report leaked forbidden marker %q: %s", forbidden, content)
		}
	}
}
