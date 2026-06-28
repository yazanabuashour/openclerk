package runner_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRetrievalTaskSynthesisFreshnessProjectionAndProvenance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	source := createDocument(t, ctx, config, "sources/runner.md", "Runner Source", "# Runner Source\n\n## Summary\nInitial source guidance.\n")
	synthesis := createDocument(t, ctx, config, "synthesis/runner.md", "Runner Synthesis", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/runner.md
---
# Runner Synthesis

## Summary
Initial source guidance.

## Sources
- sources/runner.md

## Freshness
Checked source refs.
`)+"\n")

	time.Sleep(time.Millisecond)
	updated, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   source.DocID,
		Heading: "Summary",
		Content: "Updated source guidance.",
	})
	if err != nil {
		t.Fatalf("update source: %v", err)
	}
	if updated.Document == nil {
		t.Fatalf("update result = %+v", updated)
	}

	projections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      synthesis.DocID,
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("synthesis projection task: %v", err)
	}
	if projections.Projections == nil ||
		len(projections.Projections.Projections) != 1 ||
		projections.Projections.Projections[0].Freshness != "stale" ||
		projections.Projections.Projections[0].Details["stale_source_refs"] != "sources/runner.md" {
		t.Fatalf("synthesis projections result = %+v", projections)
	}

	sourceEvents, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "source",
			RefID:   source.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("source provenance task: %v", err)
	}
	if sourceEvents.Provenance == nil ||
		!runnerEventTypesInclude(sourceEvents.Provenance.Events, "source_created") ||
		!runnerEventTypesInclude(sourceEvents.Provenance.Events, "source_updated") {
		t.Fatalf("source provenance result = %+v", sourceEvents)
	}

	synthesisEvents, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "synthesis:" + synthesis.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("synthesis provenance task: %v", err)
	}
	if synthesisEvents.Provenance == nil || !runnerEventTypesInclude(synthesisEvents.Provenance.Events, "projection_invalidated") {
		t.Fatalf("synthesis provenance result = %+v", synthesisEvents)
	}
}

func TestRetrievalTaskAutonomyModesValidateAndGateWrites(t *testing.T) {
	t.Parallel()

	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	invalid, err := runner.RunRetrievalTask(context.Background(), config, runner.RetrievalTaskRequest{
		Action:   runner.RetrievalTaskActionValidate,
		Autonomy: runner.AutonomyModes{AudienceMode: "boardroom"},
	})
	if err != nil {
		t.Fatalf("invalid autonomy mode errored: %v", err)
	}
	if !invalid.Rejected || !strings.Contains(invalid.RejectionReason, "autonomy.audience_mode") {
		t.Fatalf("invalid autonomy rejection = %+v", invalid)
	}

	proposeOnly, err := runner.RunRetrievalTask(context.Background(), config, runner.RetrievalTaskRequest{
		Action:   runner.RetrievalTaskActionSourceAuditReport,
		Autonomy: runner.AutonomyModes{ApprovalMode: runner.ApprovalModeProposeOnly},
		SourceAudit: runner.SourceAuditReportOptions{
			Query:      "autonomy source audit repair",
			TargetPath: "synthesis/autonomy.md",
			Mode:       "repair_existing",
		},
	})
	if err != nil {
		t.Fatalf("propose-only source audit errored: %v", err)
	}
	if !proposeOnly.Rejected || !strings.Contains(proposeOnly.RejectionReason, "propose_only") {
		t.Fatalf("propose-only rejection = %+v", proposeOnly)
	}
}

func TestRetrievalTaskDuplicateCandidateReportIsReadOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	target := createDocument(t, ctx, config, "notes/duplicates/existing-renewal-note.md", "Existing Renewal Note", "# Existing Renewal Note\n\n## Summary\nRenewal packaging duplicate marker belongs to the existing account renewal note.\n")
	createDocument(t, ctx, config, "notes/duplicates/decoy-renewal-note.md", "Decoy Renewal Note", "# Decoy Renewal Note\n\n## Summary\nAdjacent renewal note without the exact duplicate marker.\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDuplicateCandidate,
		DuplicateCandidate: runner.DuplicateCandidateOptions{
			Query:      "renewal packaging duplicate marker account renewal",
			PathPrefix: "notes/duplicates/",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("duplicate candidate report: %v", err)
	}
	if result.Rejected || result.DuplicateCandidate == nil {
		t.Fatalf("duplicate candidate result = %+v", result)
	}
	report := result.DuplicateCandidate
	if report.LikelyTarget == nil ||
		report.LikelyTarget.DocID != target.DocID ||
		report.LikelyTarget.Path != target.Path ||
		report.DuplicateStatus != "likely_duplicate_found" ||
		report.WriteStatus != "read_only_no_document_created_or_updated" ||
		report.AgentHandoff == nil ||
		!containsString(report.EvidenceInspected, "search:renewal packaging duplicate marker account renewal") ||
		!containsString(report.EvidenceInspected, "list_documents:notes/duplicates/") ||
		!containsString(report.EvidenceInspected, "get_document:"+target.Path) ||
		!strings.Contains(report.ApprovalBoundary, "update the likely existing target") ||
		!strings.Contains(report.ValidationBoundaries, "read-only") {
		t.Fatalf("duplicate candidate report = %+v", report)
	}
	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "notes/duplicates/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list duplicates after report: %v", err)
	}
	if len(list.Documents) != 2 {
		t.Fatalf("duplicate report mutated documents: %+v", list.Documents)
	}
}

func TestRetrievalEvalCaptureAndReplayAreSanitizedAndLocal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "notes/eval/retrieval-target.md", "Retrieval Eval Target", "# Retrieval Eval Target\n\n## Summary\nRetrieval eval marker appears here. Raw vault body secret should not enter capture rows.\n")
	createDocument(t, ctx, config, "notes/eval/other.md", "Other Eval Note", "# Other Eval Note\n\n## Summary\nAdjacent retrieval eval content.\n")
	capturePath := filepath.Join(filepath.Dir(config.DatabasePath), "captures", "retrieval-eval.jsonl")

	capture, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionRetrievalEvalCapture,
		RetrievalEval: runner.RetrievalEvalOptions{
			Action:      runner.RetrievalTaskActionSearch,
			CapturePath: capturePath,
			Search: runner.SearchOptions{
				Text:       "Retrieval   eval marker",
				PathPrefix: "notes/eval/",
				Limit:      10,
			},
		},
	})
	if err != nil {
		t.Fatalf("capture retrieval eval: %v", err)
	}
	if capture.RetrievalEvalCapture == nil ||
		capture.RetrievalEvalCapture.WriteStatus != "local_eval_artifact_appended" ||
		capture.RetrievalEvalCapture.Case.Query != "Retrieval eval marker" ||
		len(capture.RetrievalEvalCapture.Case.Results) == 0 ||
		capture.RetrievalEvalCapture.AgentHandoff == nil {
		t.Fatalf("capture report = %+v", capture)
	}
	content, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("read capture path: %v", err)
	}
	if strings.Contains(string(content), "Raw vault body secret") ||
		strings.Contains(string(content), "Snippet") {
		t.Fatalf("capture row leaked raw body/snippet content: %s", string(content))
	}

	replay, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:          runner.RetrievalTaskActionRetrievalEvalReplay,
		RetrievalReplay: runner.RetrievalReplayOptions{CapturePath: capturePath, Limit: 10},
	})
	if err != nil {
		t.Fatalf("replay retrieval eval: %v", err)
	}
	if replay.RetrievalEvalReplay == nil ||
		replay.RetrievalEvalReplay.ComparedCases != 1 ||
		replay.RetrievalEvalReplay.AverageJaccard != 1 ||
		replay.RetrievalEvalReplay.Top1MatchRate != 1 ||
		replay.RetrievalEvalReplay.AgentHandoff == nil {
		t.Fatalf("replay report = %+v", replay)
	}

	paths, err := runclient.ResolvePaths(config)
	if err != nil {
		t.Fatalf("resolve paths: %v", err)
	}
	_, err = runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionRetrievalEvalCapture,
		RetrievalEval: runner.RetrievalEvalOptions{
			Action:      runner.RetrievalTaskActionSearch,
			CapturePath: filepath.Join(paths.VaultRoot, "capture.jsonl"),
			Search:      runner.SearchOptions{Text: "Retrieval eval marker", Limit: 10},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "must not be inside the vault root") {
		t.Fatalf("capture inside vault err = %v, want vault-root rejection", err)
	}
	_, err = runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionRetrievalEvalCapture,
		RetrievalEval: runner.RetrievalEvalOptions{
			Action:      runner.RetrievalTaskActionSearch,
			CapturePath: filepath.Join(t.TempDir(), "capture.jsonl"),
			Search:      runner.SearchOptions{Text: "Retrieval eval marker", Limit: 10},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "OpenClerk data directory") {
		t.Fatalf("capture outside data dir err = %v, want data-dir rejection", err)
	}
	proposeOnly, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:   runner.RetrievalTaskActionRetrievalEvalCapture,
		Autonomy: runner.AutonomyModes{ApprovalMode: runner.ApprovalModeProposeOnly},
		RetrievalEval: runner.RetrievalEvalOptions{
			Action: runner.RetrievalTaskActionSearch,
			Search: runner.SearchOptions{Text: "Retrieval eval marker", Limit: 10},
		},
	})
	if err != nil {
		t.Fatalf("propose-only capture: %v", err)
	}
	if !proposeOnly.Rejected || !strings.Contains(proposeOnly.RejectionReason, "propose_only") {
		t.Fatalf("propose-only capture result = %+v", proposeOnly)
	}
}

func TestRetrievalEvalCaptureExecutesStoredSanitizedQuery(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "notes/eval/long-query.md", "Long Query Target", "# Long Query Target\n\n## Summary\nOnly the tail marker should match this document.\n")
	capturePath := filepath.Join(filepath.Dir(config.DatabasePath), "captures", "long-query.jsonl")
	longQuery := strings.Repeat("prefix ", 60) + "tail marker"

	capture, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionRetrievalEvalCapture,
		RetrievalEval: runner.RetrievalEvalOptions{
			Action:      runner.RetrievalTaskActionSearch,
			CapturePath: capturePath,
			Search:      runner.SearchOptions{Text: longQuery, Limit: 10},
		},
	})
	if err != nil {
		t.Fatalf("capture long query: %v", err)
	}
	if capture.RetrievalEvalCapture == nil ||
		len([]rune(capture.RetrievalEvalCapture.Case.Query)) != 240 ||
		strings.Contains(capture.RetrievalEvalCapture.Case.Query, "tail marker") ||
		len(capture.RetrievalEvalCapture.Case.Results) != 0 {
		t.Fatalf("long-query capture = %+v", capture.RetrievalEvalCapture)
	}

	replay, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:          runner.RetrievalTaskActionRetrievalEvalReplay,
		RetrievalReplay: runner.RetrievalReplayOptions{CapturePath: capturePath, Limit: 10},
	})
	if err != nil {
		t.Fatalf("replay long query: %v", err)
	}
	if replay.RetrievalEvalReplay == nil ||
		replay.RetrievalEvalReplay.ComparedCases != 1 ||
		replay.RetrievalEvalReplay.AverageJaccard != 1 ||
		replay.RetrievalEvalReplay.Top1MatchRate != 1 {
		t.Fatalf("long-query replay = %+v", replay.RetrievalEvalReplay)
	}
}

func TestRetrievalEvalCaptureRefreshesVaultIndexBeforeWriting(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	document := createDocument(t, ctx, config, "notes/eval/fresh.md", "Fresh Eval", "# Fresh Eval\n\n## Summary\nOld indexed marker.\n")
	paths, err := runclient.ResolvePaths(config)
	if err != nil {
		t.Fatalf("resolve paths: %v", err)
	}
	if err := os.WriteFile(filepath.Join(paths.VaultRoot, filepath.FromSlash(document.Path)), []byte("# Fresh Eval\n\n## Summary\nFresh capture marker from disk.\n"), 0o600); err != nil {
		t.Fatalf("write fresh vault file: %v", err)
	}

	capture, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionRetrievalEvalCapture,
		RetrievalEval: runner.RetrievalEvalOptions{
			Action:      runner.RetrievalTaskActionSearch,
			CapturePath: filepath.Join(filepath.Dir(config.DatabasePath), "captures", "fresh.jsonl"),
			Search:      runner.SearchOptions{Text: "Fresh capture marker", Limit: 10},
		},
	})
	if err != nil {
		t.Fatalf("capture fresh eval: %v", err)
	}
	if capture.RetrievalEvalCapture == nil || len(capture.RetrievalEvalCapture.Case.Results) == 0 || capture.RetrievalEvalCapture.Case.Results[0].Path != document.Path {
		t.Fatalf("fresh capture result = %+v", capture.RetrievalEvalCapture)
	}
}

func TestSearchDiagnosticsReportShowsModeAndModulePosture(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "docs/search/diagnostics.md", "Search Diagnostics", "# Search Diagnostics\n\n## Summary\nSemantic recall citation quality diagnostic marker.\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearchDiagnostics,
		SearchDiagnostics: runner.SearchDiagnosticsOptions{
			Query:      "Semantic recall citation quality diagnostic marker",
			Intent:     "semantic recall",
			PathPrefix: "docs/search/",
			Provider:   "ollama",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("search diagnostics: %v", err)
	}
	report := result.SearchDiagnostics
	if report == nil ||
		report.RecommendedAction != runner.RetrievalTaskActionSearch ||
		!report.NoDefaultRankingChange ||
		report.LexicalSearch == nil ||
		len(report.LexicalSearch.Hits) == 0 ||
		len(report.ModePostures) < 2 ||
		!searchModulePostureContains(report.ModulePostures, "ollama", "not_installed") ||
		report.AgentHandoff == nil ||
		!strings.Contains(report.ValidationBoundaries, "does not execute semantic modules") {
		t.Fatalf("search diagnostics report = %+v", report)
	}
}

func TestMaintenanceReportPackagesExistingChecksWithoutRepair(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "notes/maintenance/related.md", "Maintenance Related", "# Maintenance Related\n\n## Summary\nRelated maintenance evidence.\n")
	createDocument(t, ctx, config, "notes/maintenance/source.md", "Maintenance Source", "# Maintenance Source\n\n## Summary\nMaintenance duplicate relationship marker supersedes and superseded by [Maintenance Related](related.md).\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionMaintenanceReport,
		Maintenance: runner.MaintenanceReportOptions{
			Query:      "Maintenance duplicate relationship marker",
			PathPrefix: "notes/maintenance/",
			Limit:      20,
		},
	})
	if err != nil {
		t.Fatalf("maintenance report: %v", err)
	}
	report := result.Maintenance
	if report == nil ||
		report.WriteStatus != "read_only_no_repair" ||
		report.Layout == nil ||
		report.Projections == nil ||
		report.RelationshipContext == nil ||
		report.DuplicateCandidate == nil ||
		report.GitLifecycle == nil ||
		report.AgentHandoff == nil ||
		!maintenanceFindingArea(report.Findings, "layout") ||
		!maintenanceFindingArea(report.Findings, "projection_freshness") ||
		!maintenanceFindingArea(report.Findings, "relationship_context") ||
		!maintenanceFindingStatus(report.Findings, "relationship_context", "attention") ||
		!maintenanceFindingArea(report.Findings, "duplicate_risk") ||
		!maintenanceFindingArea(report.Findings, "module_config") ||
		!maintenanceFindingArea(report.Findings, "git_lifecycle") ||
		!strings.Contains(report.ValidationBoundaries, "does not create cron jobs") {
		t.Fatalf("maintenance report = %+v", report)
	}
}

func TestRetrievalTaskDuplicateCandidateReportRejectsMissingQuery(t *testing.T) {
	t.Parallel()

	result, err := runner.RunRetrievalTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.RetrievalTaskRequest{
		Action:             runner.RetrievalTaskActionDuplicateCandidate,
		DuplicateCandidate: runner.DuplicateCandidateOptions{PathPrefix: "notes/"},
	})
	if err != nil {
		t.Fatalf("duplicate candidate reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "duplicate_candidate.query is required" {
		t.Fatalf("duplicate candidate rejection = %+v", result)
	}
}

func TestRetrievalTaskWorkflowGuideReportRoutesIntentWithoutStore(t *testing.T) {
	t.Parallel()

	result, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionWorkflowGuide,
		WorkflowGuide: runner.WorkflowGuideOptions{
			Intent: "Should I update an existing note or create a new duplicate?",
		},
	})
	if err != nil {
		t.Fatalf("workflow guide report: %v", err)
	}
	if result.Rejected || result.WorkflowGuide == nil {
		t.Fatalf("workflow guide result = %+v", result)
	}
	report := result.WorkflowGuide
	if report.RecommendedSurface != "duplicate_candidate_report" ||
		report.RunnerDomain != "retrieval" ||
		report.AgentHandoff == nil ||
		len(report.CandidateSurfaces) == 0 ||
		!strings.Contains(report.RequestShape, "duplicate_candidate_report") ||
		!strings.Contains(report.ValidationBoundaries, "read-only routing report") ||
		!strings.Contains(report.AuthorityLimits, "final answers") {
		t.Fatalf("workflow guide report = %+v", report)
	}
}

func TestRetrievalTaskWorkflowGuideReportRoutesExplicitOCRReview(t *testing.T) {
	t.Parallel()

	result, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionWorkflowGuide,
		WorkflowGuide: runner.WorkflowGuideOptions{
			Intent: "Plan OCR review for a scan-only PDF local file receipt.",
		},
	})
	if err != nil {
		t.Fatalf("workflow guide OCR report: %v", err)
	}
	if result.Rejected || result.WorkflowGuide == nil {
		t.Fatalf("workflow guide OCR result = %+v", result)
	}
	report := result.WorkflowGuide
	if report.RecommendedSurface != "artifact_candidate_plan" ||
		report.RunnerDomain != "document" ||
		!strings.Contains(report.RequestShape, "ocr_review") ||
		!strings.Contains(report.RequestShape, "<explicit-user-local-file>") ||
		containsString(report.DoNotUseFor, "OCR") ||
		!containsString(report.DoNotUseFor, "external OCR") {
		t.Fatalf("workflow guide OCR report = %+v", report)
	}
}

func TestRetrievalTaskWorkflowGuideReportRoutesPublicScannedPDFURL(t *testing.T) {
	t.Parallel()

	result, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionWorkflowGuide,
		WorkflowGuide: runner.WorkflowGuideOptions{
			Intent: "Plan intake for a public scanned PDF URL before writing it.",
		},
	})
	if err != nil {
		t.Fatalf("workflow guide public PDF report: %v", err)
	}
	if result.Rejected || result.WorkflowGuide == nil {
		t.Fatalf("workflow guide public PDF result = %+v", result)
	}
	report := result.WorkflowGuide
	if report.RecommendedSurface != "ingest_source_url inspect" ||
		report.RunnerDomain != "document" ||
		!strings.Contains(report.RequestShape, "ingest_source_url") ||
		!strings.Contains(report.RequestShape, `"mode":"inspect"`) ||
		strings.Contains(report.RequestShape, "local_path") {
		t.Fatalf("workflow guide public PDF report = %+v", report)
	}
}

func TestRetrievalTaskWorkflowGuideReportRejectsMissingIntent(t *testing.T) {
	t.Parallel()

	result, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionWorkflowGuide,
	})
	if err != nil {
		t.Fatalf("workflow guide reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "workflow_guide.intent is required" {
		t.Fatalf("workflow guide rejection = %+v", result)
	}
}

func TestRetrievalTaskWorkflowGuideReportPrioritizesSpecificDecisionSurfaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		intent string
		want   string
	}{
		{intent: "hybrid retrieval decision support", want: "hybrid_retrieval_report"},
		{intent: "structured data decision support", want: "structured_store_report"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.intent, func(t *testing.T) {
			t.Parallel()

			result, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
				Action:        runner.RetrievalTaskActionWorkflowGuide,
				WorkflowGuide: runner.WorkflowGuideOptions{Intent: tt.intent},
			})
			if err != nil {
				t.Fatalf("workflow guide report: %v", err)
			}
			if result.WorkflowGuide == nil || result.WorkflowGuide.RecommendedSurface != tt.want {
				t.Fatalf("workflow guide surface = %+v, want %q", result.WorkflowGuide, tt.want)
			}
		})
	}
}

func TestRetrievalTaskHybridRetrievalReportIsReadOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	source := createDocument(t, ctx, config, "sources/retrieval/hybrid-baseline.md", "Hybrid Baseline", "# Hybrid Baseline\n\n## Summary\nHybrid retrieval baseline marker preserves citation quality and canonical authority.\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionHybridRetrieval,
		HybridRetrieval: runner.HybridRetrievalOptions{
			Query:      "Hybrid retrieval baseline marker citation quality",
			PathPrefix: "sources/retrieval/",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("hybrid retrieval report: %v", err)
	}
	if result.Rejected || result.HybridRetrieval == nil {
		t.Fatalf("hybrid retrieval result = %+v", result)
	}
	report := result.HybridRetrieval
	if report.LexicalSearch == nil ||
		len(report.LexicalSearch.Hits) == 0 ||
		report.LexicalSearch.Hits[0].DocID != source.DocID ||
		report.LexicalSearch.Hits[0].Citations[0].Path != source.Path ||
		len(report.CandidateSurfaces) == 0 ||
		report.AgentHandoff == nil ||
		!containsString(report.EvidenceInspected, "search:Hybrid retrieval baseline marker citation quality") ||
		!strings.Contains(report.Recommendation, "keep lexical search as the default") ||
		!strings.Contains(report.ValidationBoundaries, "does not create embeddings") ||
		!strings.Contains(report.AuthorityLimits, "canonical markdown") {
		t.Fatalf("hybrid retrieval report = %+v", report)
	}
	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "sources/retrieval/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list after hybrid report: %v", err)
	}
	if len(list.Documents) != 1 {
		t.Fatalf("hybrid report mutated documents: %+v", list.Documents)
	}
}

func TestRetrievalTaskStructuredStoreReportIsReadOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "records/assets/structured-runner.md", "Structured Runner", strings.TrimSpace(`---
entity_type: tool
entity_name: Structured Runner
entity_id: structured-runner
---
# Structured Runner

## Summary
Structured runner evidence stays in canonical markdown and projects into schema-backed records.

## Facts
- status: active
`)+"\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionStructuredStore,
		StructuredStore: runner.StructuredStoreOptions{
			Domain:     "records",
			Query:      "Structured Runner",
			EntityType: "tool",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("structured store report: %v", err)
	}
	if result.Rejected || result.StructuredStore == nil {
		t.Fatalf("structured store result = %+v", result)
	}
	report := result.StructuredStore
	if report.Records == nil ||
		len(report.Records.Entities) != 1 ||
		report.Records.Entities[0].EntityID != "structured-runner" ||
		report.Projections == nil ||
		len(report.Projections.Projections) == 0 ||
		len(report.CandidateSurfaces) == 0 ||
		report.AgentHandoff == nil ||
		!containsString(report.EvidenceInspected, "domain:records") ||
		!containsString(report.EvidenceInspected, "records:1") ||
		!strings.Contains(report.Recommendation, "structured_store_report") ||
		!strings.Contains(report.ValidationBoundaries, "read-only") ||
		!strings.Contains(report.ValidationBoundaries, "does not create independent canonical tables") ||
		!strings.Contains(report.AuthorityLimits, "canonical markdown") {
		t.Fatalf("structured store report = %+v", report)
	}
	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "records/assets/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list after structured store report: %v", err)
	}
	if len(list.Documents) != 1 {
		t.Fatalf("structured store report mutated documents: %+v", list.Documents)
	}
}

func TestRetrievalTaskStructuredStoreReportRejectsMissingFilter(t *testing.T) {
	t.Parallel()

	result, err := runner.RunRetrievalTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.RetrievalTaskRequest{
		Action:          runner.RetrievalTaskActionStructuredStore,
		StructuredStore: runner.StructuredStoreOptions{Domain: "records"},
	})
	if err != nil {
		t.Fatalf("structured store reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "structured_store.query is required for records domain" {
		t.Fatalf("structured store rejection = %+v", result)
	}
}

func TestRetrievalTaskStructuredStoreReportRejectsUnsupportedDomainFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options runner.StructuredStoreOptions
		want    string
	}{
		{
			name:    "services scope",
			options: runner.StructuredStoreOptions{Domain: "services", Scope: "agentops"},
			want:    "structured_store services domain supports query, status, owner, and interface only",
		},
		{
			name:    "decisions interface",
			options: runner.StructuredStoreOptions{Domain: "decisions", Interface: "JSON runner"},
			want:    "structured_store decisions domain supports query, status, owner, and scope only",
		},
		{
			name:    "records owner",
			options: runner.StructuredStoreOptions{Domain: "records", Query: "structured runner", Owner: "platform"},
			want:    "structured_store records domain supports query and entity_type only",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := runner.RunRetrievalTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.RetrievalTaskRequest{
				Action:          runner.RetrievalTaskActionStructuredStore,
				StructuredStore: tt.options,
			})
			if err != nil {
				t.Fatalf("structured store unsupported filter reject: %v", err)
			}
			if !result.Rejected || result.RejectionReason != tt.want {
				t.Fatalf("structured store rejection = %+v, want %q", result, tt.want)
			}
		})
	}
}

func TestRetrievalTaskHybridRetrievalReportRejectsMissingQuery(t *testing.T) {
	t.Parallel()

	result, err := runner.RunRetrievalTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.RetrievalTaskRequest{
		Action:          runner.RetrievalTaskActionHybridRetrieval,
		HybridRetrieval: runner.HybridRetrievalOptions{PathPrefix: "sources/"},
	})
	if err != nil {
		t.Fatalf("hybrid retrieval reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "hybrid_retrieval.query is required" {
		t.Fatalf("hybrid retrieval rejection = %+v", result)
	}
}

func TestRetrievalTaskGraphContextReportPackagesRelationshipEvidence(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	source := createDocument(t, ctx, config, "notes/graph/context/index.md", "Graph Context Reference", strings.TrimSpace(`---
type: note
---
# Graph Context Reference

## Relationships
Graph context report marker requires [Routing](routing.md), supersedes [Freshness](freshness.md), is related to [Operations](operations.md), and operationalizes canonical markdown relationship text.
`)+"\n")
	createDocument(t, ctx, config, "notes/graph/context/routing.md", "Routing", "# Routing\n\n## Links\nRouting links back to [Graph Context Reference](index.md).\n")
	createDocument(t, ctx, config, "notes/graph/context/freshness.md", "Freshness", "# Freshness\n\n## Links\nFreshness links back to [Graph Context Reference](index.md).\n")
	createDocument(t, ctx, config, "notes/graph/context/operations.md", "Operations", "# Operations\n\n## Links\nOperations links back to [Graph Context Reference](index.md).\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraphContext,
		GraphContext: runner.GraphContextOptions{
			Path:  source.Path,
			Limit: 20,
		},
	})
	if err != nil {
		t.Fatalf("graph context report: %v", err)
	}
	if result.Rejected || result.GraphContext == nil {
		t.Fatalf("graph context result = %+v", result)
	}
	report := result.GraphContext
	if report.SourceDocument == nil ||
		report.SourceDocument.DocID != source.DocID ||
		report.SourceDocument.Path != source.Path ||
		report.SourceSelection != "path_exact_match" ||
		len(report.CanonicalRelationshipText) == 0 ||
		!strings.Contains(report.CanonicalRelationshipText[0].Text, "requires") ||
		report.CanonicalRelationshipText[0].Citation.Path != source.Path ||
		report.Links == nil ||
		!documentLinksContainRunnerPath(report.Links.Outgoing, "notes/graph/context/routing.md") ||
		!documentLinksContainRunnerPath(report.Links.Outgoing, "notes/graph/context/freshness.md") ||
		!documentLinksContainRunnerPath(report.Links.Outgoing, "notes/graph/context/operations.md") ||
		!documentLinksContainRunnerPath(report.Links.Incoming, "notes/graph/context/routing.md") ||
		report.Graph == nil ||
		len(report.Graph.Nodes) < 4 ||
		!graphContextEdgesHaveCitations(report.Graph.Edges) ||
		report.GraphProjection == nil ||
		len(report.GraphProjection.Projections) != 1 ||
		report.GraphProjection.Projections[0].Freshness != "fresh" ||
		report.Provenance == nil ||
		!runnerEventTypesInclude(report.Provenance.Events, "document_created") ||
		report.AgentHandoff == nil ||
		!graphContextContainsPrefix(report.ProvenanceRefs, "projection:graph:document:"+source.DocID+":fresh") ||
		!graphContextContainsPrefix(report.EvidenceInspected, "canonical_relationship_text:") ||
		!strings.Contains(report.ValidationBoundaries, "read-only") ||
		!strings.Contains(report.ValidationBoundaries, "no graph memory") ||
		!strings.Contains(report.AuthorityLimits, "canonical markdown") {
		t.Fatalf("graph context report = %+v", report)
	}

	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "notes/graph/context/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list after graph context report: %v", err)
	}
	if len(list.Documents) != 4 {
		t.Fatalf("graph context report mutated documents: %+v", list.Documents)
	}
}

func TestRetrievalTaskGraphContextReportLimitsLinksAndBacklinks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	source := createDocument(t, ctx, config, "notes/graph/limit/index.md", "Graph Context Limit", "# Graph Context Limit\n\n## Links\nSee [One](one.md), [Two](two.md), [Three](three.md), [Four](four.md), and [Five](five.md).\n")
	for _, note := range []struct {
		name  string
		title string
	}{
		{name: "one", title: "One"},
		{name: "two", title: "Two"},
		{name: "three", title: "Three"},
		{name: "four", title: "Four"},
		{name: "five", title: "Five"},
	} {
		createDocument(t, ctx, config, "notes/graph/limit/"+note.name+".md", note.title, "# "+note.title+"\n\n## Links\nBack to [Graph Context Limit](index.md).\n")
	}

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraphContext,
		GraphContext: runner.GraphContextOptions{
			Path:  source.Path,
			Limit: 2,
		},
	})
	if err != nil {
		t.Fatalf("graph context limited links: %v", err)
	}
	if result.GraphContext == nil || result.GraphContext.Links == nil {
		t.Fatalf("graph context result = %+v", result)
	}
	if len(result.GraphContext.Links.Outgoing) > 2 || len(result.GraphContext.Links.Incoming) > 2 {
		t.Fatalf("graph context links were not limited: outgoing=%d incoming=%d", len(result.GraphContext.Links.Outgoing), len(result.GraphContext.Links.Incoming))
	}
}

func TestRetrievalTaskGraphContextReportRejectsMissingSelector(t *testing.T) {
	t.Parallel()

	result, err := runner.RunRetrievalTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraphContext,
	})
	if err != nil {
		t.Fatalf("graph context reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "graph_context doc_id, path, or query is required" {
		t.Fatalf("graph context rejection = %+v", result)
	}
}

func TestRetrievalTaskGraphContextReportRejectsMultipleSelectors(t *testing.T) {
	t.Parallel()

	result, err := runner.RunRetrievalTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraphContext,
		GraphContext: runner.GraphContextOptions{
			Path:  "notes/graph/context/index.md",
			Query: "graph context",
		},
	})
	if err != nil {
		t.Fatalf("graph context multi-selector reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "graph_context accepts exactly one of doc_id, path, or query" {
		t.Fatalf("graph context rejection = %+v", result)
	}
}

func TestRetrievalTaskGraphRelationshipReportPackagesDeferredGraphEvidence(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	source := createDocument(t, ctx, config, "notes/graph/relationship/index.md", "Graph Relationship Reference", strings.TrimSpace(`---
type: note
---
# Graph Relationship Reference

## Relationships
Graph relationship report marker requires [Routing](routing.md), supersedes [Freshness](freshness.md), is related to [Operations](operations.md), and operationalizes canonical markdown relationship text.
`)+"\n")
	createDocument(t, ctx, config, "notes/graph/relationship/routing.md", "Routing", "# Routing\n\n## Links\nRouting links back to [Graph Relationship Reference](index.md).\n")
	createDocument(t, ctx, config, "notes/graph/relationship/freshness.md", "Freshness", "# Freshness\n\n## Links\nFreshness links back to [Graph Relationship Reference](index.md).\n")
	createDocument(t, ctx, config, "notes/graph/relationship/operations.md", "Operations", "# Operations\n\n## Links\nOperations links back to [Graph Relationship Reference](index.md).\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraphRelationship,
		GraphRelationship: runner.GraphRelationshipOptions{
			Path:  source.Path,
			Limit: 20,
		},
	})
	if err != nil {
		t.Fatalf("graph relationship report: %v", err)
	}
	if result.Rejected || result.GraphRelationship == nil {
		t.Fatalf("graph relationship result = %+v", result)
	}
	report := result.GraphRelationship
	if report.SourceDocument == nil ||
		report.SourceDocument.DocID != source.DocID ||
		report.SourceDocument.Path != source.Path ||
		report.SourceSelection != "path_exact_match" ||
		!graphRelationshipPathContains(report.RelationshipPaths, "outgoing", "notes/graph/relationship/routing.md") ||
		!graphRelationshipPathContains(report.RelationshipPaths, "incoming", "notes/graph/relationship/routing.md") ||
		!graphRelationshipEvidenceContains(report.DirectRelationships, "requires") ||
		!graphRelationshipEvidenceContains(report.DirectRelationships, "supersedes") ||
		!graphRelationshipEvidenceContains(report.DirectRelationships, "related_to") ||
		!graphRelationshipEvidenceContains(report.DirectRelationships, "operationalizes") ||
		len(report.DerivedRelationships) == 0 ||
		!graphRelationshipTypedCandidateContains(report.TypedRelationshipCandidates, "markdown_link") ||
		!graphRelationshipTypedCandidateContains(report.TypedRelationshipCandidates, "requires") ||
		!graphRelationshipAuditFinding(report.AuditFindings, "stale_graph_projection", "clear") ||
		!graphRelationshipAuditFinding(report.AuditFindings, "orphaned_graph_context", "clear") ||
		!graphRelationshipAuditFinding(report.AuditFindings, "contradictory_relationship_text", "clear_limited") ||
		report.GraphProjection == nil ||
		len(report.GraphProjection.Projections) != 1 ||
		report.GraphProjection.Projections[0].Freshness != "fresh" ||
		report.Provenance == nil ||
		!runnerEventTypesInclude(report.Provenance.Events, "document_created") ||
		report.AgentHandoff == nil ||
		!strings.Contains(report.AgentHandoff.AnswerSummary, "Candidate comparison") ||
		!strings.Contains(report.AgentHandoff.AnswerSummary, "promote graph_relationship_report") ||
		!strings.Contains(report.AgentHandoff.AnswerSummary, "no follow-up work is required") ||
		!containsString(report.AgentHandoff.Evidence, "decision=promote graph_relationship_report") ||
		!graphRelationshipCandidatesInclude(report.CandidateSurfaces, "current_primitives_plus_graph_context_report") ||
		!graphRelationshipCandidatesInclude(report.CandidateSurfaces, "split_specialized_reports") ||
		!strings.Contains(report.ValidationBoundaries, "read-only") ||
		!strings.Contains(report.ValidationBoundaries, "no graph memory") ||
		!strings.Contains(report.AuthorityLimits, "canonical markdown") ||
		!strings.Contains(report.AuthorityLimits, "typed relationship candidates") {
		t.Fatalf("graph relationship report = %+v", report)
	}

	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "notes/graph/relationship/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list after graph relationship report: %v", err)
	}
	if len(list.Documents) != 4 {
		t.Fatalf("graph relationship report mutated documents: %+v", list.Documents)
	}
}

func TestRetrievalTaskGraphRelationshipReportRejectsSelectors(t *testing.T) {
	t.Parallel()

	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	result, err := runner.RunRetrievalTask(context.Background(), config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraphRelationship,
	})
	if err != nil {
		t.Fatalf("graph relationship missing selector reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "graph_relationship doc_id, path, or query is required" {
		t.Fatalf("graph relationship rejection = %+v", result)
	}

	result, err = runner.RunRetrievalTask(context.Background(), config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraphRelationship,
		GraphRelationship: runner.GraphRelationshipOptions{
			Path:  "notes/graph/relationship/index.md",
			Query: "graph relationship",
		},
	})
	if err != nil {
		t.Fatalf("graph relationship multi-selector reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "graph_relationship accepts exactly one of doc_id, path, or query" {
		t.Fatalf("graph relationship rejection = %+v", result)
	}
}

func TestRetrievalTaskGraphRelationshipReportTreatsStructuralMentionsAsNonRelationships(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	source := createDocument(t, ctx, config, "notes/graph/relationship/orphan.md", "Orphan Graph Note", strings.TrimSpace(`---
type: note
---
# Orphan Graph Note

## Summary
This note has ordinary chunk text but no markdown relationship links.
`)+"\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraphRelationship,
		GraphRelationship: runner.GraphRelationshipOptions{
			Path:  source.Path,
			Limit: 20,
		},
	})
	if err != nil {
		t.Fatalf("graph relationship orphan report: %v", err)
	}
	if result.Rejected || result.GraphRelationship == nil {
		t.Fatalf("graph relationship orphan result = %+v", result)
	}
	report := result.GraphRelationship
	if len(report.DerivedRelationships) != 0 {
		t.Fatalf("structural graph mentions were reported as derived relationships: %+v", report.DerivedRelationships)
	}
	if !graphRelationshipAuditFinding(report.AuditFindings, "orphaned_graph_context", "attention") {
		t.Fatalf("orphan audit did not flag no relationship links: %+v", report.AuditFindings)
	}
	if report.AgentHandoff == nil || !containsString(report.AgentHandoff.Evidence, "source_citations=0") {
		t.Fatalf("orphan handoff should not claim source citations: %+v", report.AgentHandoff)
	}
}

func TestRetrievalTaskGraphRelationshipMaintenancePlanPackagesApprovalGatedWritePlan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	source := createDocument(t, ctx, config, "notes/graph/relationship/maintenance.md", "Graph Relationship Maintenance", strings.TrimSpace(`---
type: note
---
# Graph Relationship Maintenance

## Relationships
Graph relationship maintenance marker requires [Routing](routing.md), supersedes [Freshness](freshness.md), and is related to [Operations](operations.md).
`)+"\n")
	createDocument(t, ctx, config, "notes/graph/relationship/routing.md", "Routing", "# Routing\n\n## Links\nRouting links back to [Graph Relationship Maintenance](maintenance.md).\n")
	createDocument(t, ctx, config, "notes/graph/relationship/freshness.md", "Freshness", "# Freshness\n\n## Links\nFreshness links back to [Graph Relationship Maintenance](maintenance.md).\n")
	createDocument(t, ctx, config, "notes/graph/relationship/operations.md", "Operations", "# Operations\n\n## Links\nOperations links back to [Graph Relationship Maintenance](maintenance.md).\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraphRelationshipMaintenance,
		GraphRelationshipMaintenance: runner.GraphRelationshipMaintenanceOptions{
			Path:  source.Path,
			Limit: 20,
		},
	})
	if err != nil {
		t.Fatalf("graph relationship maintenance plan: %v", err)
	}
	if result.Rejected || result.GraphRelationshipMaintenance == nil {
		t.Fatalf("graph relationship maintenance result = %+v", result)
	}
	plan := result.GraphRelationshipMaintenance
	if plan.SourceDocument == nil ||
		plan.SourceDocument.DocID != source.DocID ||
		plan.SourceSelection != "path_exact_match" ||
		plan.WriteStatus != "planned_no_write" ||
		plan.CandidateSectionHeading != "Relationships" ||
		!strings.Contains(plan.CandidateSectionContent, "requires") ||
		!strings.Contains(plan.CandidateSectionContent, "Maintenance audit findings") ||
		!strings.Contains(plan.NextReplaceSectionRequest, `"action":"replace_section"`) ||
		!strings.Contains(plan.NextReplaceSectionRequest, `"heading":"Relationships"`) ||
		!strings.Contains(plan.NextAppendDocumentRequest, `"action":"append_document"`) ||
		!strings.Contains(plan.ApprovalBoundary, "not durable-write approval") ||
		!strings.Contains(plan.RollbackAuditPath, "git_lifecycle_report") ||
		!strings.Contains(plan.DuplicateHandling, "do not create duplicate") ||
		!graphRelationshipMaintenanceActionContains(plan.ProposedActions, "typed_relationship_annotation", "candidate_requires_approval") ||
		!graphRelationshipMaintenanceCandidatesInclude(plan.CandidateSurfaces, "current_primitives_plus_graph_relationship_report") ||
		!graphRelationshipMaintenanceCandidatesInclude(plan.CandidateSurfaces, "graph_relationship_maintenance_plan") ||
		!graphRelationshipMaintenanceCandidatesInclude(plan.CandidateSurfaces, "durable_semantic_graph_maintenance") ||
		plan.GraphProjection == nil ||
		plan.AgentHandoff == nil ||
		!strings.Contains(plan.ValidationBoundaries, "read-only maintenance plan") ||
		!strings.Contains(plan.AuthorityLimits, "candidate maintenance text") {
		t.Fatalf("graph relationship maintenance plan = %+v", plan)
	}

	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "notes/graph/relationship/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list after graph relationship maintenance plan: %v", err)
	}
	if len(list.Documents) != 4 {
		t.Fatalf("graph relationship maintenance plan mutated documents: %+v", list.Documents)
	}
}

func TestRetrievalTaskGraphRelationshipMaintenancePlanRejectsSelectors(t *testing.T) {
	t.Parallel()

	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	result, err := runner.RunRetrievalTask(context.Background(), config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraphRelationshipMaintenance,
	})
	if err != nil {
		t.Fatalf("graph relationship maintenance missing selector reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "graph_relationship_maintenance doc_id, path, or query is required" {
		t.Fatalf("graph relationship maintenance rejection = %+v", result)
	}

	result, err = runner.RunRetrievalTask(context.Background(), config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraphRelationshipMaintenance,
		GraphRelationshipMaintenance: runner.GraphRelationshipMaintenanceOptions{
			Path:  "notes/graph/relationship/index.md",
			Query: "graph relationship",
		},
	})
	if err != nil {
		t.Fatalf("graph relationship maintenance multi-selector reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "graph_relationship_maintenance accepts exactly one of doc_id, path, or query" {
		t.Fatalf("graph relationship maintenance rejection = %+v", result)
	}

	result, err = runner.RunRetrievalTask(context.Background(), config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraphRelationshipMaintenance,
		GraphRelationshipMaintenance: runner.GraphRelationshipMaintenanceOptions{
			PathPrefix: "notes/graph/",
		},
	})
	if err != nil {
		t.Fatalf("graph relationship maintenance path_prefix reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "graph_relationship_maintenance doc_id, path, or query is required" {
		t.Fatalf("graph relationship maintenance path_prefix rejection = %+v", result)
	}
}

func TestRetrievalTaskAuditContradictionsPlansAndRepairsExistingSynthesis(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	oldSource := createDocument(t, ctx, config, "sources/audit-runner-old.md", "Old audit runner source", strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/audit-runner-current.md
---
# Old audit runner source

## Summary
Older source-sensitive audit runner repair evidence said agents should prefer a legacy command-path workaround.
`)+"\n")
	currentSource := createDocument(t, ctx, config, "sources/audit-runner-current.md", "Current audit runner source", strings.TrimSpace(`---
type: source
status: active
supersedes: sources/audit-runner-old.md
---
# Current audit runner source

## Summary
Current source-sensitive audit runner repair evidence says use the installed openclerk JSON runner.
`)+"\n")
	synthesis := createDocument(t, ctx, config, "synthesis/audit-runner-routing.md", "Audit runner routing", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/audit-runner-current.md, sources/audit-runner-old.md
---
# Audit runner routing

## Summary
Stale audit claim: agents should prefer a legacy command-path workaround.

## Sources
- sources/audit-runner-current.md
- sources/audit-runner-old.md

## Freshness
Checked source refs.
`)+"\n")
	createDocument(t, ctx, config, "synthesis/audit-runner-routing-decoy.md", "Audit runner routing decoy", "# Audit runner routing decoy\n\n## Summary\nDecoy.\n")
	createDocument(t, ctx, config, "sources/audit-conflict-alpha.md", "Audit conflict alpha", strings.TrimSpace(`---
type: source
status: active
---
# Audit conflict alpha

## Summary
Source sensitive audit conflict runner retention is seven days.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-conflict-bravo.md", "Audit conflict bravo", strings.TrimSpace(`---
type: source
status: active
---
# Audit conflict bravo

## Summary
Source sensitive audit conflict runner retention is thirty days.
`)+"\n")

	time.Sleep(time.Millisecond)
	_, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   currentSource.DocID,
		Heading: "Summary",
		Content: "Current source-sensitive audit runner repair evidence says use the installed openclerk JSON runner for audit repairs.",
	})
	if err != nil {
		t.Fatalf("update current source: %v", err)
	}

	plan, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:         "source-sensitive audit runner repair evidence",
			TargetPath:    "synthesis/audit-runner-routing.md",
			Mode:          "plan_only",
			ConflictQuery: "source sensitive audit conflict runner retention",
			Limit:         10,
		},
	})
	if err != nil {
		t.Fatalf("audit plan: %v", err)
	}
	if plan.Audit == nil ||
		plan.Audit.RepairStatus != "planned" ||
		plan.Audit.RepairApplied ||
		plan.Audit.SelectedTargetPath != "synthesis/audit-runner-routing.md" ||
		!containsString(plan.Audit.CandidateSynthesisPaths, "synthesis/audit-runner-routing-decoy.md") ||
		!containsString(plan.Audit.CurrentSourcePaths, currentSource.Path) ||
		!containsString(plan.Audit.SupersededSourcePaths, oldSource.Path) ||
		len(plan.Audit.ProjectionFreshnessBefore) == 0 ||
		len(plan.Audit.ProjectionFreshnessAfter) == 0 ||
		len(plan.Audit.UnresolvedConflictGroups) != 1 ||
		plan.Audit.UnresolvedConflictGroups[0].Status != "unresolved" {
		t.Fatalf("audit plan result = %+v", plan.Audit)
	}
	if !auditInspectedPath(plan.Audit.ProvenanceInspected, "sources/audit-conflict-alpha.md") ||
		!auditInspectedPath(plan.Audit.ProvenanceInspected, "sources/audit-conflict-bravo.md") {
		t.Fatalf("audit plan did not inspect conflict provenance: %+v", plan.Audit.ProvenanceInspected)
	}
	unchanged, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  synthesis.DocID,
	})
	if err != nil {
		t.Fatalf("get unchanged synthesis: %v", err)
	}
	if !strings.Contains(unchanged.Document.Body, "legacy command-path workaround") {
		t.Fatalf("plan_only changed synthesis body = %q", unchanged.Document.Body)
	}

	sourceAudit, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSourceAuditReport,
		SourceAudit: runner.SourceAuditReportOptions{
			Query:         "source-sensitive audit runner repair evidence",
			TargetPath:    "synthesis/audit-runner-routing.md",
			Mode:          "explain",
			ConflictQuery: "source sensitive audit conflict runner retention",
			Limit:         10,
		},
	})
	if err != nil {
		t.Fatalf("source audit explain: %v", err)
	}
	if sourceAudit.SourceAudit == nil ||
		sourceAudit.SourceAudit.Mode != "explain" ||
		sourceAudit.SourceAudit.RepairStatus != "planned" ||
		sourceAudit.SourceAudit.RepairApplied ||
		sourceAudit.SourceAudit.SelectedTargetPath != "synthesis/audit-runner-routing.md" ||
		sourceAudit.SourceAudit.AgentHandoff == nil ||
		!strings.Contains(sourceAudit.SourceAudit.AgentHandoff.AnswerSummary, "source_audit_report explain") ||
		!strings.Contains(sourceAudit.SourceAudit.AgentHandoff.FollowUpPrimitiveInspection, "not required") ||
		!strings.Contains(sourceAudit.SourceAudit.ValidationBoundaries, "explain mode is read-only") ||
		!strings.Contains(sourceAudit.SourceAudit.AuthorityLimits, "unresolved current-source conflicts") {
		t.Fatalf("source audit explain result = %+v", sourceAudit.SourceAudit)
	}
	stillUnchanged, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  synthesis.DocID,
	})
	if err != nil {
		t.Fatalf("get source audit unchanged synthesis: %v", err)
	}
	if !strings.Contains(stillUnchanged.Document.Body, "legacy command-path workaround") {
		t.Fatalf("source_audit explain changed synthesis body = %q", stillUnchanged.Document.Body)
	}

	repaired, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:         "source-sensitive audit runner repair evidence",
			TargetPath:    "synthesis/audit-runner-routing.md",
			Mode:          "repair_existing",
			ConflictQuery: "source sensitive audit conflict runner retention",
			Limit:         10,
		},
	})
	if err != nil {
		t.Fatalf("audit repair: %v", err)
	}
	if repaired.Audit == nil ||
		repaired.Audit.RepairStatus != "applied" ||
		!repaired.Audit.RepairApplied ||
		repaired.Audit.DuplicatePrevention != "existing_target_selected_no_duplicate_created" ||
		repaired.Audit.FailureClassification != "none" {
		t.Fatalf("audit repair result = %+v", repaired.Audit)
	}
	if len(repaired.Audit.ProjectionFreshnessAfter) == 0 || repaired.Audit.ProjectionFreshnessAfter[0].Freshness != "fresh" {
		t.Fatalf("projection after repair = %+v", repaired.Audit.ProjectionFreshnessAfter)
	}
	updated, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  synthesis.DocID,
	})
	if err != nil {
		t.Fatalf("get repaired synthesis: %v", err)
	}
	for _, want := range []string{
		"source_refs: sources/audit-runner-current.md, sources/audit-runner-old.md",
		"Current source summary: Current source-sensitive audit runner repair evidence says use the installed openclerk JSON runner for audit repairs.",
		"Current source: sources/audit-runner-current.md",
		"Superseded source: sources/audit-runner-old.md",
		"Audit repair did not choose unresolved claims beyond source classification.",
		"## Sources",
		"## Freshness",
	} {
		if !strings.Contains(updated.Document.Body, want) {
			t.Fatalf("repaired body missing %q:\n%s", want, updated.Document.Body)
		}
	}
}

func TestRetrievalTaskSourceAuditReportRepairsOnlyExistingSynthesis(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/source-audit-old.md", "Old source audit source", strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/source-audit-current.md
---
# Old source audit source

## Summary
Old source-sensitive audit workflow says use direct vault inspection.
`)+"\n")
	createDocument(t, ctx, config, "sources/source-audit-current.md", "Current source audit source", strings.TrimSpace(`---
type: source
status: active
supersedes: sources/source-audit-old.md
---
# Current source audit source

## Summary
Current source-sensitive audit workflow says use source_audit_report.
`)+"\n")
	synthesis := createDocument(t, ctx, config, "synthesis/source-audit.md", "Source audit synthesis", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/source-audit-current.md, sources/source-audit-old.md
---
# Source audit synthesis

## Summary
Stale audit workflow says use direct vault inspection.

## Sources
- sources/source-audit-current.md
- sources/source-audit-old.md

## Freshness
Checked source refs.
`)+"\n")

	repaired, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSourceAuditReport,
		SourceAudit: runner.SourceAuditReportOptions{
			Query:      "source-sensitive audit workflow",
			TargetPath: "synthesis/source-audit.md",
			Mode:       "repair_existing",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("source audit repair: %v", err)
	}
	if repaired.SourceAudit == nil ||
		repaired.SourceAudit.Mode != "repair_existing" ||
		!repaired.SourceAudit.RepairApplied ||
		repaired.SourceAudit.RepairStatus != "applied" ||
		repaired.SourceAudit.DuplicatePrevention != "existing_target_selected_no_duplicate_created" {
		t.Fatalf("source audit repair result = %+v", repaired.SourceAudit)
	}

	updated, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  synthesis.DocID,
	})
	if err != nil {
		t.Fatalf("get repaired source audit synthesis: %v", err)
	}
	if strings.Contains(updated.Document.Body, "direct vault inspection") ||
		!strings.Contains(updated.Document.Body, "Audit repair did not choose unresolved claims beyond source classification.") {
		t.Fatalf("source audit repaired body = %q", updated.Document.Body)
	}

	missing, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSourceAuditReport,
		SourceAudit: runner.SourceAuditReportOptions{
			Query:      "source-sensitive audit workflow",
			TargetPath: "synthesis/missing-source-audit.md",
			Mode:       "repair_existing",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("source audit missing target: %v", err)
	}
	if missing.SourceAudit == nil ||
		missing.SourceAudit.RepairApplied ||
		missing.SourceAudit.FailureClassification != "target_not_found" {
		t.Fatalf("source audit missing target result = %+v", missing.SourceAudit)
	}
}

func TestRetrievalTaskAuditContradictionsFindsTargetAfterFirstSynthesisPage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/audit-page-old.md", "Old audit page source", strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/audit-page-current.md
---
# Old audit page source

## Summary
Older source-sensitive audit page evidence said to use a legacy command path.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-page-current.md", "Current audit page source", strings.TrimSpace(`---
type: source
status: active
supersedes: sources/audit-page-old.md
---
# Current audit page source

## Summary
Current source-sensitive audit page evidence says use the installed openclerk JSON runner.
`)+"\n")
	for i := 0; i < 101; i++ {
		createDocument(t, ctx, config, fmt.Sprintf("synthesis/aa-audit-decoy-%03d.md", i), fmt.Sprintf("Audit decoy %03d", i), fmt.Sprintf("# Audit decoy %03d\n\n## Summary\nDecoy.\n", i))
	}
	target := createDocument(t, ctx, config, "synthesis/zz-audit-page-target.md", "Audit page target", strings.TrimSpace(`---
type: synthesis
status: active
source_refs: sources/audit-page-current.md, sources/audit-page-old.md
---
# Audit page target

## Summary
Stale audit page claim.

## Sources
- sources/audit-page-current.md
- sources/audit-page-old.md

## Freshness
Checked source refs.
`)+"\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:      "source-sensitive audit page evidence",
			TargetPath: target.Path,
			Mode:       "plan_only",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("audit paginated target: %v", err)
	}
	if result.Audit == nil ||
		result.Audit.SelectedTargetPath != target.Path ||
		result.Audit.RepairStatus != "planned" ||
		result.Audit.DuplicatePrevention != "existing_target_selected_no_duplicate_created" {
		t.Fatalf("audit paginated target result = %+v", result.Audit)
	}
}

func TestRetrievalTaskAuditContradictionsDoesNotReportMatchingCurrentSourcesAsConflict(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/audit-agree-old.md", "Old audit agree source", strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/audit-agree-current.md
---
# Old audit agree source

## Summary
Older source-sensitive audit agreement evidence said to use a legacy command path.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-agree-current.md", "Current audit agree source", strings.TrimSpace(`---
type: source
status: active
supersedes: sources/audit-agree-old.md
---
# Current audit agree source

## Summary
Current source-sensitive audit agreement evidence says use the installed openclerk JSON runner.
`)+"\n")
	createDocument(t, ctx, config, "synthesis/audit-agree-target.md", "Audit agree target", strings.TrimSpace(`---
type: synthesis
status: active
source_refs: sources/audit-agree-current.md, sources/audit-agree-old.md
---
# Audit agree target

## Summary
Stale audit agreement claim.

## Sources
- sources/audit-agree-current.md
- sources/audit-agree-old.md

## Freshness
Checked source refs.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-retention-alpha.md", "Audit retention alpha", strings.TrimSpace(`---
type: source
status: active
---
# Audit retention alpha

## Summary
Source sensitive audit matching retention is seven days.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-retention-bravo.md", "Audit retention bravo", strings.TrimSpace(`---
type: source
status: active
---
# Audit retention bravo

## Summary
Source sensitive audit matching retention is seven days.
`)+"\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:         "source-sensitive audit agreement evidence",
			TargetPath:    "synthesis/audit-agree-target.md",
			Mode:          "plan_only",
			ConflictQuery: "source sensitive audit matching retention",
			Limit:         10,
		},
	})
	if err != nil {
		t.Fatalf("audit matching conflict: %v", err)
	}
	if result.Audit == nil || len(result.Audit.UnresolvedConflictGroups) != 0 {
		t.Fatalf("matching current sources reported as conflict = %+v", result.Audit)
	}
	if !auditInspectedPath(result.Audit.ProvenanceInspected, "sources/audit-retention-alpha.md") ||
		!auditInspectedPath(result.Audit.ProvenanceInspected, "sources/audit-retention-bravo.md") {
		t.Fatalf("audit matching conflict did not inspect provenance: %+v", result.Audit.ProvenanceInspected)
	}
}

func TestRetrievalTaskAuditContradictionsValidation(t *testing.T) {
	t.Parallel()

	missing, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			TargetPath: "synthesis/audit-runner-routing.md",
		},
	})
	if err != nil {
		t.Fatalf("missing audit query validation: %v", err)
	}
	if !missing.Rejected || missing.RejectionReason != "audit.query is required" {
		t.Fatalf("missing result = %+v", missing)
	}

	invalidMode, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:      "source-sensitive audit runner repair evidence",
			TargetPath: "synthesis/audit-runner-routing.md",
			Mode:       "create_new",
		},
	})
	if err != nil {
		t.Fatalf("invalid audit mode validation: %v", err)
	}
	if !invalidMode.Rejected || invalidMode.RejectionReason != "audit.mode must be plan_only or repair_existing" {
		t.Fatalf("invalid mode result = %+v", invalidMode)
	}
}

func TestRetrievalTaskSearchLinksRecordsAndProvenance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	architecture := createDocument(t, ctx, config, "notes/architecture/knowledge-plane.md", "Knowledge plane", "# Knowledge plane\n\n## Summary\nCanonical architecture note.\n")
	roadmap := createDocument(t, ctx, config, "notes/projects/roadmap.md", "Roadmap", "# Roadmap\n\n## Summary\nSee the [knowledge plane](../architecture/knowledge-plane.md).\n")
	createDocument(t, ctx, config, "records/assets/transmission-solenoid.md", "Transmission solenoid", "---\nentity_type: part\nentity_name: Transmission solenoid\nentity_id: transmission-solenoid\n---\n# Transmission solenoid\n\n## Facts\n- sku: SOL-1\n")
	createDocument(t, ctx, config, "records/services/openclerk-runner.md", "OpenClerk runner", "---\nservice_id: openclerk-runner\nservice_name: OpenClerk runner\nservice_status: active\nservice_owner: runner\nservice_interface: JSON runner\n---\n# OpenClerk runner\n\n## Summary\nProduction service for routine knowledge tasks.\n\n## Facts\n- tier: production\n")
	createDocument(t, ctx, config, "docs/architecture/runner-old-decision.md", "Old runner decision", "---\ndecision_id: adr-runner-old\ndecision_title: Old runner path\ndecision_status: superseded\ndecision_scope: agentops\ndecision_owner: platform\ndecision_date: 2026-04-20\nsuperseded_by: adr-runner-current\nsource_refs: sources/runner-old.md\n---\n# Old runner path\n\n## Summary\nOld decision used a retired runner path.\n")
	createDocument(t, ctx, config, "sources/runner-old.md", "Old runner source", "# Old runner source\n\n## Summary\nRetired runner path source.\n")
	createDocument(t, ctx, config, "notes/architecture/runner-current-decision.md", "Current runner decision", "---\ndecision_id: adr-runner-current\ndecision_title: Use JSON runner\ndecision_status: accepted\ndecision_scope: agentops\ndecision_owner: platform\ndecision_date: 2026-04-22\nsupersedes: adr-runner-old\nsource_refs: sources/runner-current.md\n---\n# Use JSON runner\n\n## Summary\nAccepted decision uses the JSON runner for routine AgentOps work.\n")
	createDocument(t, ctx, config, "sources/runner-current.md", "Current runner source", "# Current runner source\n\n## Summary\nCurrent runner source.\n")

	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  "roadmap",
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("search task: %v", err)
	}
	if search.Search == nil || len(search.Search.Hits) == 0 || len(search.Search.Hits[0].Citations) == 0 {
		t.Fatalf("search result = %+v", search)
	}

	links, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDocumentLinks,
		DocID:  roadmap.DocID,
	})
	if err != nil {
		t.Fatalf("links task: %v", err)
	}
	if links.Links == nil || len(links.Links.Outgoing) != 1 || links.Links.Outgoing[0].DocID != architecture.DocID {
		t.Fatalf("links result = %+v", links)
	}

	graph, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraph,
		DocID:  roadmap.DocID,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("graph task: %v", err)
	}
	if graph.Graph == nil || len(graph.Graph.Nodes) == 0 || len(graph.Graph.Edges) == 0 {
		t.Fatalf("graph result = %+v", graph)
	}

	records, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:  runner.RetrievalTaskActionRecordsLookup,
		Records: runner.RecordLookupOptions{Text: "solenoid", Limit: 10},
	})
	if err != nil {
		t.Fatalf("records task: %v", err)
	}
	if records.Records == nil || len(records.Records.Entities) != 1 {
		t.Fatalf("records result = %+v", records)
	}

	entity, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:   runner.RetrievalTaskActionRecordEntity,
		EntityID: records.Records.Entities[0].EntityID,
	})
	if err != nil {
		t.Fatalf("record entity task: %v", err)
	}
	if entity.Entity == nil || entity.Entity.EntityID != "transmission-solenoid" {
		t.Fatalf("entity result = %+v", entity)
	}

	services, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionServicesLookup,
		Services: runner.ServiceLookupOptions{
			Text:      "OpenClerk runner",
			Interface: "JSON runner",
			Limit:     10,
		},
	})
	if err != nil {
		t.Fatalf("services task: %v", err)
	}
	if services.Services == nil || len(services.Services.Services) != 1 || services.Services.Services[0].ServiceID != "openclerk-runner" {
		t.Fatalf("services result = %+v", services)
	}

	service, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:    runner.RetrievalTaskActionServiceRecord,
		ServiceID: "openclerk-runner",
	})
	if err != nil {
		t.Fatalf("service record task: %v", err)
	}
	if service.Service == nil ||
		service.Service.Status != "active" ||
		service.Service.Owner != "runner" ||
		service.Service.Interface != "JSON runner" ||
		len(service.Service.Citations) == 0 {
		t.Fatalf("service result = %+v", service)
	}

	decisions, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDecisionsLookup,
		Decisions: runner.DecisionLookupOptions{
			Text:   "JSON runner",
			Status: "accepted",
			Scope:  "agentops",
			Owner:  "platform",
			Limit:  10,
		},
	})
	if err != nil {
		t.Fatalf("decisions task: %v", err)
	}
	if decisions.Decisions == nil ||
		len(decisions.Decisions.Decisions) != 1 ||
		decisions.Decisions.Decisions[0].DecisionID != "adr-runner-current" ||
		len(decisions.Decisions.Decisions[0].Citations) == 0 {
		t.Fatalf("decisions result = %+v", decisions)
	}

	decision, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionDecisionRecord,
		DecisionID: "adr-runner-old",
	})
	if err != nil {
		t.Fatalf("decision record task: %v", err)
	}
	if decision.Decision == nil ||
		decision.Decision.Status != "superseded" ||
		len(decision.Decision.SupersededBy) != 1 ||
		decision.Decision.SupersededBy[0] != "adr-runner-current" ||
		len(decision.Decision.Citations) == 0 {
		t.Fatalf("decision result = %+v", decision)
	}

	provenance, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   roadmap.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("provenance task: %v", err)
	}
	if provenance.Provenance == nil || len(provenance.Provenance.Events) == 0 {
		t.Fatalf("provenance result = %+v", provenance)
	}

	projections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "graph",
			RefKind:    "document",
			RefID:      roadmap.DocID,
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("projection task: %v", err)
	}
	if projections.Projections == nil || len(projections.Projections.Projections) != 1 {
		t.Fatalf("projection result = %+v", projections)
	}

	serviceProjections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "services",
			RefKind:    "service",
			RefID:      "openclerk-runner",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("service projection task: %v", err)
	}
	if serviceProjections.Projections == nil ||
		len(serviceProjections.Projections.Projections) != 1 ||
		serviceProjections.Projections.Projections[0].Freshness != "fresh" {
		t.Fatalf("service projections result = %+v", serviceProjections)
	}

	decisionProjections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      "adr-runner-old",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("decision projection task: %v", err)
	}
	if decisionProjections.Projections == nil ||
		len(decisionProjections.Projections.Projections) != 1 ||
		decisionProjections.Projections.Projections[0].Freshness != "stale" ||
		decisionProjections.Projections.Projections[0].Details["superseded_by"] != "adr-runner-current" {
		t.Fatalf("decision projections result = %+v", decisionProjections)
	}

	evidence, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionEvidenceBundle,
		EvidenceBundle: runner.EvidenceBundleOptions{
			Query:      "JSON runner",
			EntityID:   "transmission-solenoid",
			DecisionID: "adr-runner-current",
			RefKind:    "document",
			RefID:      roadmap.DocID,
			Projection: "graph",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("evidence bundle task: %v", err)
	}
	if evidence.EvidenceBundle == nil ||
		evidence.EvidenceBundle.Search == nil ||
		evidence.EvidenceBundle.Records == nil ||
		evidence.EvidenceBundle.Entity == nil ||
		evidence.EvidenceBundle.Decisions == nil ||
		evidence.EvidenceBundle.Decision == nil ||
		evidence.EvidenceBundle.Provenance == nil ||
		evidence.EvidenceBundle.Projections == nil ||
		evidence.EvidenceBundle.AgentHandoff == nil ||
		!strings.Contains(evidence.EvidenceBundle.AgentHandoff.AnswerSummary, "evidence_bundle_report returned") ||
		!strings.Contains(evidence.EvidenceBundle.AgentHandoff.FollowUpPrimitiveInspection, "not required") ||
		len(evidence.EvidenceBundle.Citations) == 0 ||
		!strings.Contains(evidence.EvidenceBundle.ValidationBoundaries, "read-only") ||
		!strings.Contains(evidence.EvidenceBundle.AuthorityLimits, "does not create a new authority source") {
		t.Fatalf("evidence bundle result = %+v", evidence.EvidenceBundle)
	}

	entityOnlyEvidence, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionEvidenceBundle,
		EvidenceBundle: runner.EvidenceBundleOptions{
			EntityID: "transmission-solenoid",
			Limit:    10,
		},
	})
	if err != nil {
		t.Fatalf("entity-only evidence bundle task: %v", err)
	}
	if entityOnlyEvidence.EvidenceBundle == nil ||
		entityOnlyEvidence.EvidenceBundle.Records != nil ||
		entityOnlyEvidence.EvidenceBundle.Entity == nil ||
		entityOnlyEvidence.EvidenceBundle.AgentHandoff == nil ||
		!strings.Contains(entityOnlyEvidence.EvidenceBundle.AgentHandoff.AnswerSummary, "1 records") ||
		!strings.Contains(strings.Join(entityOnlyEvidence.EvidenceBundle.AgentHandoff.Evidence, "\n"), "records=1") ||
		!strings.Contains(strings.Join(entityOnlyEvidence.EvidenceBundle.AgentHandoff.Evidence, "\n"), "decisions=0") {
		t.Fatalf("entity-only evidence bundle handoff = %+v", entityOnlyEvidence.EvidenceBundle)
	}

	decisionOnlyEvidence, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionEvidenceBundle,
		EvidenceBundle: runner.EvidenceBundleOptions{
			DecisionID: "adr-runner-current",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("decision-only evidence bundle task: %v", err)
	}
	if decisionOnlyEvidence.EvidenceBundle == nil ||
		decisionOnlyEvidence.EvidenceBundle.Decisions != nil ||
		decisionOnlyEvidence.EvidenceBundle.Decision == nil ||
		decisionOnlyEvidence.EvidenceBundle.AgentHandoff == nil ||
		!strings.Contains(decisionOnlyEvidence.EvidenceBundle.AgentHandoff.AnswerSummary, "1 decisions") ||
		!strings.Contains(strings.Join(decisionOnlyEvidence.EvidenceBundle.AgentHandoff.Evidence, "\n"), "records=0") ||
		!strings.Contains(strings.Join(decisionOnlyEvidence.EvidenceBundle.AgentHandoff.Evidence, "\n"), "decisions=1") {
		t.Fatalf("decision-only evidence bundle handoff = %+v", decisionOnlyEvidence.EvidenceBundle)
	}

	afterEvidenceList, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 100},
	})
	if err != nil {
		t.Fatalf("list after evidence bundle: %v", err)
	}
	if len(afterEvidenceList.Documents) != 8 {
		t.Fatalf("evidence bundle changed document count: %+v", afterEvidenceList.Documents)
	}
}

func TestRetrievalTaskMemoryRouterRecallReport(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	seedMemoryRouterRecallDocs(t, ctx, config)

	before, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 50},
	})
	if err != nil {
		t.Fatalf("list before: %v", err)
	}

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionMemoryRouterRecall,
		MemoryRouterRecall: runner.MemoryRouterRecallOptions{
			Query: "memory router temporal recall session promotion feedback weighting routing canonical docs",
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("memory/router recall report: %v", err)
	}
	if result.Rejected || result.MemoryRouterRecall == nil {
		t.Fatalf("memory/router recall result = %+v", result)
	}
	report := result.MemoryRouterRecall
	for _, value := range []string{
		report.QuerySummary,
		report.TemporalStatus,
		report.StaleSessionStatus,
		report.FeedbackWeighting,
		report.RoutingRationale,
		report.SynthesisFreshness,
		report.ValidationBoundaries,
		report.AuthorityLimits,
	} {
		if strings.TrimSpace(value) == "" {
			t.Fatalf("empty report field in %+v", report)
		}
	}
	for _, want := range []string{
		"notes/memory-router/session-observation.md",
		"notes/memory-router/temporal-policy.md",
		"notes/memory-router/feedback-weighting.md",
		"notes/memory-router/routing-policy.md",
		"synthesis/memory-router-reference.md",
	} {
		if !containsString(report.CanonicalEvidenceRefs, want) {
			t.Fatalf("canonical refs %v missing %q", report.CanonicalEvidenceRefs, want)
		}
	}
	if len(report.ProvenanceRefs) == 0 || !strings.HasPrefix(report.ProvenanceRefs[0], "document:") {
		t.Fatalf("provenance refs = %+v", report.ProvenanceRefs)
	}
	if report.AgentHandoff == nil ||
		!strings.Contains(report.AgentHandoff.AnswerSummary, "memory-router policy evidence report") ||
		!containsString(report.AgentHandoff.Evidence, "notes/memory-router/session-observation.md") ||
		!strings.Contains(report.AgentHandoff.FollowUpPrimitiveInspection, "ordinary vault fact recall use retrieval search") {
		t.Fatalf("agent handoff = %+v", report.AgentHandoff)
	}
	if !strings.Contains(report.SynthesisFreshness, "fresh synthesis projection") {
		t.Fatalf("synthesis freshness = %q", report.SynthesisFreshness)
	}
	if strings.Contains(report.ValidationBoundaries, "missing evidence") {
		t.Fatalf("unexpected missing evidence in validation boundaries: %q", report.ValidationBoundaries)
	}

	after, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 50},
	})
	if err != nil {
		t.Fatalf("list after: %v", err)
	}
	if len(after.Documents) != len(before.Documents) {
		t.Fatalf("memory/router recall report mutated document count: before=%d after=%d", len(before.Documents), len(after.Documents))
	}
}

func TestRetrievalTaskMemoryRouterRecallReportRejectsProjectionWithoutCanonicalSourceRefs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "notes/memory-router/session-observation.md", "Memory Router Session Observation", "# Memory Router Session Observation\n\n## Summary\nSession observation evidence.\n")
	createDocument(t, ctx, config, "notes/memory-router/temporal-policy.md", "Temporal Recall Policy", "# Temporal Recall Policy\n\n## Summary\nTemporal evidence.\n")
	createDocument(t, ctx, config, "notes/memory-router/feedback-weighting.md", "Feedback Weighting", "# Feedback Weighting\n\n## Summary\nFeedback evidence.\n")
	createDocument(t, ctx, config, "notes/memory-router/routing-policy.md", "Routing Policy", "# Routing Policy\n\n## Summary\nRouting evidence.\n")
	createDocument(t, ctx, config, "synthesis/memory-router-reference.md", "Memory Router Reference", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: notes/memory-router/session-observation.md
---
# Memory Router Reference

## Summary
Memory router reference without all canonical source refs.

## Sources
- notes/memory-router/session-observation.md

## Freshness
Fresh projection is not enough without the expected source refs.
`)+"\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionMemoryRouterRecall,
		MemoryRouterRecall: runner.MemoryRouterRecallOptions{
			Query: "memory router temporal recall session promotion feedback weighting routing canonical docs",
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("memory/router recall report with incomplete source refs: %v", err)
	}
	if result.Rejected || result.MemoryRouterRecall == nil {
		t.Fatalf("incomplete source refs result = %+v", result)
	}
	report := result.MemoryRouterRecall
	if strings.Contains(report.SynthesisFreshness, "fresh synthesis projection") ||
		!strings.Contains(report.ValidationBoundaries, "missing evidence") ||
		!strings.Contains(report.ValidationBoundaries, "projection:synthesis/memory-router-reference.md") {
		t.Fatalf("report accepted incomplete source refs: %+v", report)
	}
}

func TestRetrievalTaskMemoryRouterRecallReportPagesUntilFixedEvidence(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	seedMemoryRouterRecallDocs(t, ctx, config)
	for i := range 12 {
		createDocument(
			t,
			ctx,
			config,
			fmt.Sprintf("notes/memory-router/000-extra-%02d.md", i),
			fmt.Sprintf("Memory Router Extra %02d", i),
			fmt.Sprintf("# Memory Router Extra %02d\n\n## Summary\nDistractor memory/router note.\n", i),
		)
	}
	for i := range 24 {
		createDocument(
			t,
			ctx,
			config,
			fmt.Sprintf("synthesis/000-extra-%02d.md", i),
			fmt.Sprintf("Synthesis Extra %02d", i),
			fmt.Sprintf("# Synthesis Extra %02d\n\n## Summary\nDistractor synthesis.\n", i),
		)
	}

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionMemoryRouterRecall,
		MemoryRouterRecall: runner.MemoryRouterRecallOptions{
			Query: "memory router temporal recall session promotion feedback weighting routing canonical docs",
			Limit: 4,
		},
	})
	if err != nil {
		t.Fatalf("memory/router recall report with paged evidence: %v", err)
	}
	if result.Rejected || result.MemoryRouterRecall == nil {
		t.Fatalf("paged evidence result = %+v", result)
	}
	report := result.MemoryRouterRecall
	for _, want := range []string{
		"notes/memory-router/session-observation.md",
		"notes/memory-router/temporal-policy.md",
		"notes/memory-router/feedback-weighting.md",
		"notes/memory-router/routing-policy.md",
		"synthesis/memory-router-reference.md",
	} {
		if !containsString(report.CanonicalEvidenceRefs, want) {
			t.Fatalf("canonical refs %v missing %q", report.CanonicalEvidenceRefs, want)
		}
	}
	if strings.Contains(report.ValidationBoundaries, "missing evidence") {
		t.Fatalf("paged lookup reported missing evidence: %q", report.ValidationBoundaries)
	}
}

func TestRetrievalTaskMemoryRouterRecallReportReportsMissingEvidence(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionMemoryRouterRecall,
		MemoryRouterRecall: runner.MemoryRouterRecallOptions{
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("memory/router recall report with missing evidence: %v", err)
	}
	if result.Rejected || result.MemoryRouterRecall == nil {
		t.Fatalf("missing evidence result = %+v", result)
	}
	report := result.MemoryRouterRecall
	if !containsString(report.CanonicalEvidenceRefs, "missing:notes/memory-router/session-observation.md") {
		t.Fatalf("missing evidence refs = %+v", report.CanonicalEvidenceRefs)
	}
	if !strings.Contains(report.ValidationBoundaries, "missing evidence") {
		t.Fatalf("validation boundaries = %q", report.ValidationBoundaries)
	}
}

func TestRetrievalTaskMemoryRouterRecallReportRejectsNegativeLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionMemoryRouterRecall,
		MemoryRouterRecall: runner.MemoryRouterRecallOptions{
			Limit: -3,
		},
	})
	if err != nil {
		t.Fatalf("negative memory/router recall limit: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "limit must be greater than or equal to 0" {
		t.Fatalf("negative limit result = %+v", result)
	}
}

func seedMemoryRouterRecallDocs(t *testing.T, ctx context.Context, config runclient.Config) {
	t.Helper()
	createDocument(t, ctx, config, "notes/memory-router/session-observation.md", "Memory Router Session Observation", strings.TrimSpace(`---
type: source
status: active
observed_at: 2026-04-22
---
# Memory Router Session Observation

## Summary
Session observation: a user asked whether memory routing should promote recall. Useful session material must be promoted only by writing canonical markdown with source refs.

## Feedback
Positive feedback weight 0.8 is advisory only and cannot hide stale canonical evidence.
`)+"\n")
	createDocument(t, ctx, config, "notes/memory-router/temporal-policy.md", "Temporal Recall Policy", "# Temporal Recall Policy\n\n## Summary\nCurrent canonical docs outrank stale session observations for memory/router recall.\n")
	createDocument(t, ctx, config, "notes/memory-router/feedback-weighting.md", "Feedback Weighting", "# Feedback Weighting\n\n## Summary\nFeedback weighting is advisory only and cannot hide stale canonical evidence.\n")
	createDocument(t, ctx, config, "notes/memory-router/routing-policy.md", "Routing Policy", "# Routing Policy\n\n## Summary\nRouting rationale uses existing AgentOps document and retrieval actions; no autonomous router API is authority.\n")
	createDocument(t, ctx, config, "synthesis/memory-router-reference.md", "Memory Router Reference", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, notes/memory-router/routing-policy.md
---
# Memory Router Reference

## Summary
Temporal status: current canonical docs outrank stale session observations.
Session promotion path: durable canonical markdown with source refs.
Feedback weighting: advisory only.
Routing choice: existing AgentOps document and retrieval actions.
Decision: implement read-only memory/router recall report.

## Sources
- notes/memory-router/session-observation.md
- notes/memory-router/temporal-policy.md
- notes/memory-router/feedback-weighting.md
- notes/memory-router/routing-policy.md

## Freshness
Fresh synthesis projection expected for current source refs.
`)+"\n")
}

func TestRetrievalTaskTypedRecordValidation(t *testing.T) {
	t.Parallel()

	missing, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionServiceRecord,
	})
	if err != nil {
		t.Fatalf("missing service id validation: %v", err)
	}
	if !missing.Rejected || missing.RejectionReason != "service_id is required" {
		t.Fatalf("missing result = %+v", missing)
	}

	negative, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action:   runner.RetrievalTaskActionServicesLookup,
		Services: runner.ServiceLookupOptions{Limit: -1},
	})
	if err != nil {
		t.Fatalf("negative service limit validation: %v", err)
	}
	if !negative.Rejected || negative.RejectionReason != "limit must be greater than or equal to 0" {
		t.Fatalf("negative result = %+v", negative)
	}

	missingDecision, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDecisionRecord,
	})
	if err != nil {
		t.Fatalf("missing decision id validation: %v", err)
	}
	if !missingDecision.Rejected || missingDecision.RejectionReason != "decision_id is required" {
		t.Fatalf("missing decision result = %+v", missingDecision)
	}

	negativeDecision, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action:    runner.RetrievalTaskActionDecisionsLookup,
		Decisions: runner.DecisionLookupOptions{Limit: -1},
	})
	if err != nil {
		t.Fatalf("negative decision limit validation: %v", err)
	}
	if !negativeDecision.Rejected || negativeDecision.RejectionReason != "limit must be greater than or equal to 0" {
		t.Fatalf("negative decision result = %+v", negativeDecision)
	}
}

func TestRetrievalTaskSearchTagFilter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "notes/tagging/account-renewal.md", "Account Renewal", strings.TrimSpace(`---
tag: account-renewal
---
# Account Renewal

## Summary
Renewal tag search evidence belongs to account renewal.
`)+"\n")
	createDocument(t, ctx, config, "notes/tagging/security-renewal.md", "Security Renewal", strings.TrimSpace(`---
tag: renewal
---
# Security Renewal

## Summary
Renewal tag search evidence belongs to security renewal.
`)+"\n")
	createDocument(t, ctx, config, "notes/tagging/ops-review.md", "Ops Review", strings.TrimSpace(`---
tag: ops-review
---
# Ops Review

## Summary
Near duplicate tag evidence belongs to singular ops review.
`)+"\n")
	createDocument(t, ctx, config, "notes/tagging/ops-reviews.md", "Ops Reviews", strings.TrimSpace(`---
tag: ops-reviews
---
# Ops Reviews

## Summary
Near duplicate tag evidence belongs to plural ops reviews.
`)+"\n")
	createDocument(t, ctx, config, "archive/tagging/account-renewal.md", "Archived Account Renewal", strings.TrimSpace(`---
tag: account-renewal
---
# Archived Account Renewal

## Summary
Archived renewal tag search evidence must be excluded by path prefix.
`)+"\n")

	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  "Renewal tag search evidence",
			Tag:   " account-renewal ",
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("search tag: %v", err)
	}
	if search.Search == nil || !searchResultContainsPath(search.Search.Hits, "notes/tagging/account-renewal.md") {
		t.Fatalf("search tag result = %+v", search.Search)
	}
	if searchResultContainsPath(search.Search.Hits, "notes/tagging/security-renewal.md") {
		t.Fatalf("search tag included wrong tag result: %+v", search.Search.Hits)
	}

	nearDuplicate, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  "Near duplicate tag evidence",
			Tag:   "ops-review",
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("near duplicate tag search: %v", err)
	}
	if nearDuplicate.Search == nil ||
		!searchResultContainsPath(nearDuplicate.Search.Hits, "notes/tagging/ops-review.md") ||
		searchResultContainsPath(nearDuplicate.Search.Hits, "notes/tagging/ops-reviews.md") {
		t.Fatalf("near duplicate tag result = %+v", nearDuplicate.Search)
	}

	scoped, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       "renewal tag search evidence",
			PathPrefix: "notes/tagging/",
			Tag:        "account-renewal",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("scoped tag search: %v", err)
	}
	if scoped.Search == nil ||
		!searchResultContainsPath(scoped.Search.Hits, "notes/tagging/account-renewal.md") ||
		searchResultContainsPath(scoped.Search.Hits, "archive/tagging/account-renewal.md") {
		t.Fatalf("scoped tag result = %+v", scoped.Search)
	}

	mixed, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          "renewal",
			Tag:           "account-renewal",
			MetadataKey:   "tag",
			MetadataValue: "account-renewal",
		},
	})
	if err != nil {
		t.Fatalf("mixed tag search validation: %v", err)
	}
	if !mixed.Rejected || mixed.RejectionReason != "search.tag cannot be combined with metadata_key or metadata_value" {
		t.Fatalf("mixed tag result = %+v", mixed)
	}

	empty, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text: "renewal",
			Tag:  " ",
		},
	})
	if err != nil {
		t.Fatalf("empty tag search validation: %v", err)
	}
	if !empty.Rejected || empty.RejectionReason != "search.tag must be non-empty" {
		t.Fatalf("empty tag result = %+v", empty)
	}
}

func searchResultContainsPath(hits []runner.SearchHit, path string) bool {
	for _, hit := range hits {
		for _, citation := range hit.Citations {
			if citation.Path == path {
				return true
			}
		}
	}
	return false
}

func TestRetrievalSearchZeroHitLexicalFallbackPreservesFiltersAndCitations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "docs/architecture/semantic-target.md", "Semantic Target", strings.TrimSpace(`---
tag: semantic-recall
---
# Semantic Target

## Summary
Hybrid retrieval architecture evidence says vector ranking should preserve citation quality.
`)+"\n")
	createDocument(t, ctx, config, "docs/architecture/semantic-duplicate.md", "Semantic Duplicate", strings.TrimSpace(`---
tag: semantic-recall
---
# Semantic Duplicate

## Summary
Hybrid retrieval architecture evidence duplicate chunk.

## Details
Hybrid retrieval architecture evidence duplicate chunk.
`)+"\n")
	createDocument(t, ctx, config, "archive/semantic-target.md", "Archived Semantic Target", strings.TrimSpace(`---
tag: semantic-recall
---
# Archived Semantic Target

## Summary
Hybrid retrieval architecture evidence archived copy.
`)+"\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       "architecture vector citations",
			PathPrefix: "docs/architecture/",
			Tag:        "semantic-recall",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("fallback search: %v", err)
	}
	if result.Search == nil ||
		!searchResultContainsPath(result.Search.Hits, "docs/architecture/semantic-target.md") ||
		searchResultContainsPath(result.Search.Hits, "archive/semantic-target.md") {
		t.Fatalf("fallback search result = %+v", result.Search)
	}
	seen := map[string]bool{}
	for _, hit := range result.Search.Hits {
		if len(hit.Citations) == 0 || hit.Citations[0].LineStart == 0 || hit.Citations[0].Path == "" {
			t.Fatalf("hit missing citation: %+v", hit)
		}
		if seen[hit.Citations[0].Path] {
			t.Fatalf("fallback did not collapse duplicate document hits: %+v", result.Search.Hits)
		}
		seen[hit.Citations[0].Path] = true
	}
}

func TestRetrievalSearchZeroHitFallbackDoesNotReplaceLaterEmptyFTSPage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "docs/architecture/exact.md", "Exact Match", "# Exact Match\n\n## Summary\nRed blue exact phrase.\n")
	createDocument(t, ctx, config, "docs/architecture/token-decoy.md", "Token Decoy", "# Token Decoy\n\n## Summary\nRed only token overlap.\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:   "red blue",
			Limit:  10,
			Cursor: base64.RawURLEncoding.EncodeToString([]byte("1")),
		},
	})
	if err != nil {
		t.Fatalf("cursor search: %v", err)
	}
	if result.Search == nil || len(result.Search.Hits) != 0 {
		t.Fatalf("later empty FTS page should not be replaced by fallback hits: %+v", result.Search)
	}
}

func TestRetrievalTaskSourceDiscoveryReportPackagesRepresentativeSources(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/discovery-alpha.md", "Discovery Alpha", "# Discovery Alpha\n\n## Summary\nRepresentative source discovery alpha evidence.\n")
	createDocument(t, ctx, config, "sources/discovery-beta.md", "Discovery Beta", "# Discovery Beta\n\n## Summary\nRepresentative source discovery beta evidence.\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSourceDiscovery,
		SourceDiscovery: runner.SourceDiscoveryOptions{
			Query: "representative source discovery",
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("source discovery report task: %v", err)
	}
	report := result.SourceDiscovery
	if report == nil ||
		report.SearchHitCount == 0 ||
		report.RepresentativeCount < 2 ||
		report.CitationCount == 0 ||
		len(report.SourceCategories) == 0 ||
		report.AgentHandoff == nil ||
		!strings.Contains(report.AgentHandoff.AnswerSummary, "source_discovery_report inspected") ||
		!strings.Contains(report.AgentHandoff.FollowUpPrimitiveInspection, "not required") ||
		!strings.Contains(report.ValidationBoundaries, "read-only") ||
		!strings.Contains(report.AuthorityLimits, "canonical markdown sources") {
		t.Fatalf("source discovery report = %+v", report)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal source discovery report: %v", err)
	}
	for _, privateToken := range []string{"Discovery Alpha", "discovery-alpha.md", "Representative source discovery alpha evidence"} {
		if strings.Contains(string(encoded), privateToken) {
			t.Fatalf("source discovery report leaked private token %q: %s", privateToken, encoded)
		}
	}
}

func TestRetrievalTaskDecisionLookupReportHandlesDecisionLikeRecords(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "records/decisions/lookup-current.md", "Lookup Current", strings.TrimSpace(`---
decision_id: adr-lookup-current
decision_title: Lookup Current
decision_status: accepted
decision_scope: routine-ux
decision_owner: platform
---
# Lookup Current

## Summary
Decision-like lookup should find formal decisions and package freshness.
`)+"\n")
	createDocument(t, ctx, config, "records/tools/lookup-record.md", "Lookup Record", strings.TrimSpace(`---
entity_id: lookup-record
entity_type: tool
entity_name: Lookup Record
---
# Lookup Record

## Summary
Decision-like lookup can also inspect promoted records.
`)+"\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDecisionLookup,
		DecisionLookup: runner.DecisionLookupReportOptions{
			Query: "Decision-like lookup",
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("decision lookup report task: %v", err)
	}
	report := result.DecisionLookup
	if report == nil ||
		report.Decisions == nil ||
		report.Decision == nil ||
		report.Records == nil ||
		report.Search == nil ||
		report.Provenance == nil ||
		report.Projections == nil ||
		report.AgentHandoff == nil ||
		report.LookupStatus != "formal_decision_record_found" ||
		!strings.Contains(report.AgentHandoff.AnswerSummary, "decision_lookup_report returned") ||
		!strings.Contains(report.AgentHandoff.FollowUpPrimitiveInspection, "not required") ||
		!strings.Contains(report.ValidationBoundaries, "read-only") ||
		!strings.Contains(report.AuthorityLimits, "degrades gracefully") {
		t.Fatalf("decision lookup report = %+v", report)
	}
}
