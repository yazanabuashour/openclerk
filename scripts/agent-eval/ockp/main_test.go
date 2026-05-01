package main

import (
	"context"
	"encoding/json"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExecuteRunWritesParallelCacheTimingAndReports(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:   2,
		Variant:    "production",
		Scenario:   "create-note,append-replace",
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	var output strings.Builder
	err := executeRun(context.Background(), config, &output, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		time.Sleep(10 * time.Millisecond)
		now := time.Now().UTC()
		input := 100 + job.Index
		cached := 20
		nonCached := input - cached
		outputTokens := 10
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        "completed",
			Passed:        true,
			WallSeconds:   0.25,
			PhaseTimings:  phaseTimings{AgentRun: 0.25, Total: 0.30},
			Metrics: metrics{
				AssistantCalls:       1,
				ToolCalls:            job.Index + 1,
				CommandExecutions:    job.Index + 1,
				UsageExposed:         true,
				InputTokens:          &input,
				CachedInputTokens:    &cached,
				NonCachedInputTokens: &nonCached,
				OutputTokens:         &outputTokens,
				EventTypeCounts:      map[string]int{"message": 1},
			},
			Verification:            verificationResult{Passed: true, DatabasePass: true, AssistantPass: true},
			RawLogArtifactReference: "<run-root>/" + job.Variant + "/" + job.Scenario.ID + "/turn-1/events.jsonl",
			StartedAt:               now,
			CompletedAt:             &now,
		}
	})
	if err != nil {
		t.Fatalf("execute run: %v", err)
	}
	jsonPath := filepath.Join(reportDir, "ockp-test.json")
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.ConfiguredParallelism != 2 || report.Metadata.CacheMode != cacheModeIsolated || report.Metadata.HarnessElapsedSeconds <= 0 {
		t.Fatalf("metadata = %+v", report.Metadata)
	}
	if report.Metadata.Lane != populatedDefaultLaneName || !report.Metadata.ReleaseBlocking {
		t.Fatalf("lane metadata = %q/%t, want %q/true", report.Metadata.Lane, report.Metadata.ReleaseBlocking, populatedDefaultLaneName)
	}
	if report.Metadata.PhaseTotals.AgentRun != 0.50 {
		t.Fatalf("phase totals = %+v", report.Metadata.PhaseTotals)
	}
	if len(report.Results) != 2 || report.Results[0].Scenario != "create-note" || report.Results[1].Scenario != "append-replace" {
		t.Fatalf("results = %+v", report.Results)
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{"Lane", "Release blocking", "Configured parallelism", "Cache mode", "Phase Timings", "<run-root>/production/create-note/turn-1/events.jsonl"} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
	if !strings.Contains(output.String(), "ockp-test.json") {
		t.Fatalf("stdout = %q", output.String())
	}
}

func TestVerifyRepoDocsRetrievalRequiresArchitecturePathPrefix(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	seedMinimalRepoDocsDogfood(t, ctx, cfg)

	answer := repoDocsAgentOpsADRPath + " describes the AgentOps installed openclerk runner surface with doc_id and chunk_id citation evidence."
	metrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		SearchPathFilterUsed: true,
		SearchPathPrefixes:   []string{"sources/"},
		EventTypeCounts:      map[string]int{},
	}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: repoDocsRetrievalScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify repo-docs retrieval with wrong prefix: %v", err)
	}
	if result.Passed {
		t.Fatalf("repo-docs retrieval passed without docs/architecture/ path_prefix: %+v", result)
	}

	metrics.SearchPathPrefixes = []string{"docs/architecture/"}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: repoDocsRetrievalScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify repo-docs retrieval with architecture prefix: %v", err)
	}
	if !result.Passed {
		t.Fatalf("repo-docs retrieval failed with docs/architecture/ path_prefix: %+v", result)
	}
}

func TestGraphContainsNodeLabelsRequiresEveryCitedLabel(t *testing.T) {
	nodes := []runner.GraphNode{
		{Label: "AgentOps Wiki Index", Citations: []runner.Citation{{DocID: "doc_1"}}},
		{Label: "Runner Policy", Citations: []runner.Citation{{DocID: "doc_2"}}},
		{Label: "Knowledge Plane", Citations: []runner.Citation{{DocID: "doc_3"}}},
	}
	if graphContainsNodeLabels(nodes, []string{"AgentOps Wiki Index", "Runner Policy", "Knowledge Plane", "Runner Playbook"}) {
		t.Fatal("graph labels passed without the runner playbook node")
	}
	nodes = append(nodes, runner.GraphNode{Label: "Runner Playbook", Citations: []runner.Citation{{DocID: "doc_4"}}})
	if !graphContainsNodeLabels(nodes, []string{"AgentOps Wiki Index", "Runner Policy", "Knowledge Plane", "Runner Playbook"}) {
		t.Fatal("graph labels failed with every required cited node")
	}
}

func TestVerifyDocumentHistoryReviewScenarios(t *testing.T) {
	ctx := context.Background()
	commonMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProvenanceEventsUsed: true,
		ProjectionStatesUsed: true,
		CommandExecutions:    5,
		ToolCalls:            5,
		EventTypeCounts:      map[string]int{},
	}

	naturalPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, naturalPaths, scenario{ID: documentHistoryNaturalScenarioID}); err != nil {
		t.Fatalf("seed natural lifecycle pressure: %v", err)
	}
	replaceSeedSection(t, ctx, naturalPaths, documentHistoryRestoreTargetPath, "Summary", "Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits.")
	naturalAnswer := "Restored notes/history-review/restore-target.md from sources/history-review/restore-authority.md after natural lifecycle review intent. The rollback preserved source evidence, provenance, projection freshness, and no raw private diffs."
	result, err := verifyScenarioTurn(ctx, naturalPaths, scenario{ID: documentHistoryNaturalScenarioID}, 1, naturalAnswer, commonMetrics)
	if err != nil {
		t.Fatalf("verify natural lifecycle pressure: %v", err)
	}
	if !result.Passed {
		t.Fatalf("natural lifecycle pressure failed: %+v", result)
	}

	inspectionPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, inspectionPaths, scenario{ID: documentHistoryInspectScenarioID}); err != nil {
		t.Fatalf("seed history inspection: %v", err)
	}
	inspectionAnswer := "Existing runner workflow inspected notes/history-review/lifecycle-control.md with document_updated provenance and fresh projection freshness before proposing any new history action."
	result, err = verifyScenarioTurn(ctx, inspectionPaths, scenario{ID: documentHistoryInspectScenarioID}, 1, inspectionAnswer, commonMetrics)
	if err != nil {
		t.Fatalf("verify history inspection: %v", err)
	}
	if !result.Passed {
		t.Fatalf("history inspection failed: %+v", result)
	}

	diffPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, diffPaths, scenario{ID: documentHistoryDiffScenarioID}); err != nil {
		t.Fatalf("seed diff review: %v", err)
	}
	diffAnswer := "Semantic summary only: sources/history-review/diff-previous.md said review was optional, while notes/history-review/diff-current.md says review is required. Citations/source refs are preserved, and raw private diffs are not included."
	commonMetrics.ListDocumentPathPrefixes = []string{"notes/history-review/"}
	result, err = verifyScenarioTurn(ctx, diffPaths, scenario{ID: documentHistoryDiffScenarioID}, 1, diffAnswer, commonMetrics)
	if err != nil {
		t.Fatalf("verify diff review: %v", err)
	}
	if !result.Passed {
		t.Fatalf("diff review failed: %+v", result)
	}
	noListMetrics := commonMetrics
	noListMetrics.ListDocumentsUsed = false
	result, err = verifyScenarioTurn(ctx, diffPaths, scenario{ID: documentHistoryDiffScenarioID}, 1, diffAnswer, noListMetrics)
	if err != nil {
		t.Fatalf("verify diff review without list: %v", err)
	}
	if result.Passed {
		t.Fatalf("diff review passed without list_documents: %+v", result)
	}
	missingPrefixMetrics := commonMetrics
	missingPrefixMetrics.ListDocumentPathPrefixes = nil
	result, err = verifyScenarioTurn(ctx, diffPaths, scenario{ID: documentHistoryDiffScenarioID}, 1, diffAnswer, missingPrefixMetrics)
	if err != nil {
		t.Fatalf("verify diff review without path prefix: %v", err)
	}
	if result.Passed {
		t.Fatalf("diff review passed without required path_prefix: %+v", result)
	}
	extraPrefixMetrics := commonMetrics
	extraPrefixMetrics.ListDocumentPathPrefixes = []string{"notes/history-review/", "sources/history-review/"}
	result, err = verifyScenarioTurn(ctx, diffPaths, scenario{ID: documentHistoryDiffScenarioID}, 1, diffAnswer, extraPrefixMetrics)
	if err != nil {
		t.Fatalf("verify diff review with extra path prefix: %v", err)
	}
	if result.Passed {
		t.Fatalf("diff review passed with extra path_prefix: %+v", result)
	}
	badPathMetrics := commonMetrics
	badPathMetrics.ListDocumentPathPrefixes = []string{".openclerk-eval/vault/notes/history-review/"}
	result, err = verifyScenarioTurn(ctx, diffPaths, scenario{ID: documentHistoryDiffScenarioID}, 1, diffAnswer, badPathMetrics)
	if err != nil {
		t.Fatalf("verify diff review with storage path: %v", err)
	}
	if result.Passed {
		t.Fatalf("diff review passed with storage path prefix: %+v", result)
	}
	for _, badPath := range []string{"/tmp/vault/notes/history-review/", `C:\Users\me\vault\notes\history-review\`, `notes\history-review\`} {
		badPathMetrics.ListDocumentPathPrefixes = []string{badPath}
		result, err = verifyScenarioTurn(ctx, diffPaths, scenario{ID: documentHistoryDiffScenarioID}, 1, diffAnswer, badPathMetrics)
		if err != nil {
			t.Fatalf("verify diff review with bad path %q: %v", badPath, err)
		}
		if result.Passed {
			t.Fatalf("diff review passed with bad path prefix %q: %+v", badPath, result)
		}
	}
	leakyAnswer := diffAnswer + " Storage path: .openclerk-eval/vault/notes/history-review/diff-current.md."
	result, err = verifyScenarioTurn(ctx, diffPaths, scenario{ID: documentHistoryDiffScenarioID}, 1, leakyAnswer, commonMetrics)
	if err != nil {
		t.Fatalf("verify diff review with leaked final path: %v", err)
	}
	if result.Passed {
		t.Fatalf("diff review passed with leaked final path: %+v", result)
	}

	restorePaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, restorePaths, scenario{ID: documentHistoryRestoreScenarioID}); err != nil {
		t.Fatalf("seed restore pressure: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, restorePaths, scenario{ID: documentHistoryRestoreScenarioID}, 1, "No restore yet.", commonMetrics)
	if err != nil {
		t.Fatalf("verify restore before update: %v", err)
	}
	if result.Passed {
		t.Fatalf("restore pressure passed before restore: %+v", result)
	}
	replaceSeedSection(t, ctx, restorePaths, documentHistoryRestoreTargetPath, "Summary", "Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits.")
	restoreAnswer := "Restored notes/history-review/restore-target.md from sources/history-review/restore-authority.md as a rollback with source evidence, provenance, projection freshness, and citations."
	result, err = verifyScenarioTurn(ctx, restorePaths, scenario{ID: documentHistoryRestoreScenarioID}, 1, restoreAnswer, commonMetrics)
	if err != nil {
		t.Fatalf("verify restore pressure: %v", err)
	}
	if !result.Passed {
		t.Fatalf("restore pressure failed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, restorePaths, scenario{ID: documentHistoryRestoreScenarioID}, 1, restoreAnswer, noListMetrics)
	if err != nil {
		t.Fatalf("verify restore pressure without list: %v", err)
	}
	if result.Passed {
		t.Fatalf("restore pressure passed without list_documents: %+v", result)
	}

	highTouchScriptedPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, highTouchScriptedPaths, scenario{ID: highTouchDocumentLifecycleScriptedScenarioID}); err != nil {
		t.Fatalf("seed high-touch scripted lifecycle pressure: %v", err)
	}
	replaceSeedSection(t, ctx, highTouchScriptedPaths, documentHistoryRestoreTargetPath, "Summary", "Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits.")
	targetID, targetFound, err := documentIDByPath(ctx, highTouchScriptedPaths, documentHistoryRestoreTargetPath)
	if err != nil {
		t.Fatalf("resolve restore target id: %v", err)
	}
	if !targetFound {
		t.Fatal("restore target id not found")
	}
	highTouchScriptedMetrics := commonMetrics
	highTouchScriptedMetrics.ListDocumentPathPrefixes = []string{documentHistoryDiffListPrefix}
	highTouchScriptedMetrics.GetDocumentDocIDs = []string{targetID}
	highTouchScriptedMetrics.DocumentActionEvents = []string{
		"search",
		"list_documents:" + documentHistoryDiffListPrefix,
		"get_document:" + targetID,
		"replace_section:" + targetID,
		"provenance_events:" + targetID,
		"projection_states:" + targetID,
	}
	result, err = verifyScenarioTurn(ctx, highTouchScriptedPaths, scenario{ID: highTouchDocumentLifecycleScriptedScenarioID}, 1, restoreAnswer, highTouchScriptedMetrics)
	if err != nil {
		t.Fatalf("verify high-touch scripted lifecycle pressure: %v", err)
	}
	if !result.Passed {
		t.Fatalf("high-touch scripted lifecycle pressure failed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, highTouchScriptedPaths, scenario{ID: documentLifecycleRollbackCurrentScenarioID}, 1, restoreAnswer, highTouchScriptedMetrics)
	if err != nil {
		t.Fatalf("verify lifecycle rollback current-primitives pressure: %v", err)
	}
	if !result.Passed {
		t.Fatalf("lifecycle rollback current-primitives pressure failed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, highTouchScriptedPaths, scenario{ID: documentLifecycleRollbackGuidanceScenarioID}, 1, restoreAnswer, highTouchScriptedMetrics)
	if err != nil {
		t.Fatalf("verify lifecycle rollback guidance-only pressure: %v", err)
	}
	if !result.Passed {
		t.Fatalf("lifecycle rollback guidance-only pressure failed: %+v", result)
	}
	sourceListMetrics := highTouchScriptedMetrics
	sourceListMetrics.ListDocumentPathPrefixes = []string{documentHistoryDiffListPrefix, "sources/history-review/"}
	for _, scenarioID := range []string{documentLifecycleRollbackCurrentScenarioID, documentLifecycleRollbackGuidanceScenarioID} {
		result, err = verifyScenarioTurn(ctx, highTouchScriptedPaths, scenario{ID: scenarioID}, 1, restoreAnswer, sourceListMetrics)
		if err != nil {
			t.Fatalf("verify lifecycle rollback source-prefix rejection for %s: %v", scenarioID, err)
		}
		if result.Passed {
			t.Fatalf("lifecycle rollback scenario %s passed with sources/history-review/ list prefix: %+v", scenarioID, result)
		}
	}
	lateGetMetrics := highTouchScriptedMetrics
	lateGetMetrics.DocumentActionEvents = []string{
		"search",
		"list_documents:" + documentHistoryDiffListPrefix,
		"replace_section:" + targetID,
		"get_document:" + targetID,
		"provenance_events:" + targetID,
		"projection_states:" + targetID,
	}
	result, err = verifyScenarioTurn(ctx, highTouchScriptedPaths, scenario{ID: highTouchDocumentLifecycleScriptedScenarioID}, 1, restoreAnswer, lateGetMetrics)
	if err != nil {
		t.Fatalf("verify high-touch scripted lifecycle pressure with late get: %v", err)
	}
	if result.Passed {
		t.Fatalf("high-touch scripted lifecycle pressure passed with get after replace: %+v", result)
	}

	candidateAnswer := "```json\n" +
		"{\n" +
		"  \"target_path\": \"notes/history-review/restore-target.md\",\n" +
		"  \"target_doc_id\": \"" + targetID + "\",\n" +
		"  \"source_refs\": [\"sources/history-review/restore-authority.md\"],\n" +
		"  \"source_evidence\": [\"sources/history-review/restore-authority.md requires runner-visible review before accepting source-sensitive durable edits\"],\n" +
		"  \"before_summary\": \"Unsafe accepted edit said source-sensitive lifecycle edits may bypass review.\",\n" +
		"  \"after_summary\": \"Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits.\",\n" +
		"  \"restore_reason\": \"Rollback restored the unsafe accepted lifecycle edit to source-backed policy.\",\n" +
		"  \"provenance_refs\": [\"document:" + targetID + "\", \"document_updated\", \"runner-owned no-bypass retrieval\"],\n" +
		"  \"projection_freshness\": \"fresh document projection for notes/history-review/restore-target.md\",\n" +
		"  \"write_status\": \"replace_section restored the target Summary\",\n" +
		"  \"privacy_boundaries\": \"privacy-safe semantic summary only; no raw private diff; no storage-root path\",\n" +
		"  \"validation_boundaries\": \"sqlite, vault, source-built, unsupported transport, broad repo search, and direct file edit paths are rejected\",\n" +
		"  \"authority_limits\": \"canonical markdown source only; eval-only candidate object does not implement a runner action\"\n" +
		"}\n" +
		"```"
	result, err = verifyScenarioTurn(ctx, highTouchScriptedPaths, scenario{ID: documentLifecycleRollbackResponseScenarioID}, 1, candidateAnswer, highTouchScriptedMetrics)
	if err != nil {
		t.Fatalf("verify lifecycle rollback response candidate: %v", err)
	}
	if !result.Passed {
		t.Fatalf("lifecycle rollback response candidate failed: %+v", result)
	}
	wrappedCandidateAnswer := "Candidate response:\n" + candidateAnswer
	result, err = verifyScenarioTurn(ctx, highTouchScriptedPaths, scenario{ID: documentLifecycleRollbackResponseScenarioID}, 1, wrappedCandidateAnswer, highTouchScriptedMetrics)
	if err != nil {
		t.Fatalf("verify lifecycle rollback response candidate prose wrapper: %v", err)
	}
	if result.Passed {
		t.Fatalf("lifecycle rollback response candidate passed with prose wrapper: %+v", result)
	}
	for name, badAnswer := range map[string]string{
		"rollback target inaccuracy": strings.Replace(candidateAnswer, documentHistoryRestoreTargetPath, "notes/history-review/wrong-target.md", 1),
		"missing provenance":         strings.Replace(candidateAnswer, "\"provenance_refs\": [\"document:"+targetID+"\", \"document_updated\", \"runner-owned no-bypass retrieval\"],", "\"provenance_refs\": [],", 1),
		"missing freshness":          strings.Replace(candidateAnswer, "fresh document projection for notes/history-review/restore-target.md", "projection unknown", 1),
		"missing privacy":            strings.Replace(candidateAnswer, "privacy-safe semantic summary only; no raw private diff; no storage-root path", "full raw private diff included", 1),
		"privacy leak allowed":       strings.Replace(candidateAnswer, "privacy-safe semantic summary only; no raw private diff; no storage-root path", "privacy-safe semantic summary only; raw private diff included; no storage-root path", 1),
		"missing bypass boundaries":  strings.Replace(candidateAnswer, "sqlite, vault, source-built, unsupported transport, broad repo search, and direct file edit paths are rejected", "lower-level access allowed", 1),
		"bypass boundaries allowed":  strings.Replace(candidateAnswer, "sqlite, vault, source-built, unsupported transport, broad repo search, and direct file edit paths are rejected", "direct SQLite, vault inspection, source-built runner, unsupported transport, broad repo search, and direct file edit paths are allowed", 1),
	} {
		result, err = verifyScenarioTurn(ctx, highTouchScriptedPaths, scenario{ID: documentLifecycleRollbackResponseScenarioID}, 1, badAnswer, highTouchScriptedMetrics)
		if err != nil {
			t.Fatalf("verify lifecycle rollback response candidate %s: %v", name, err)
		}
		if result.Passed {
			t.Fatalf("lifecycle rollback response candidate passed with %s: %+v", name, result)
		}
	}

	pendingPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, pendingPaths, scenario{ID: documentHistoryPendingScenarioID}); err != nil {
		t.Fatalf("seed pending review: %v", err)
	}
	cfg := runclient.Config{DatabasePath: pendingPaths.DatabasePath}
	proposal := strings.TrimSpace(`---
type: review
status: pending
---
# Pending History Review Change

## Summary
Review state: pending human review.

## Proposal
Proposed change: Auto-accept pending change only after operator approval.

Target document: notes/history-review/pending-target.md
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryPendingProposalPath, "Pending History Review Change", proposal); err != nil {
		t.Fatalf("create pending proposal: %v", err)
	}
	pendingAnswer := "reviews/history-review/pending-change.md records pending human/operator review for notes/history-review/pending-target.md. The accepted target did not change and did not become accepted knowledge."
	result, err = verifyScenarioTurn(ctx, pendingPaths, scenario{ID: documentHistoryPendingScenarioID}, 1, pendingAnswer, commonMetrics)
	if err != nil {
		t.Fatalf("verify pending review: %v", err)
	}
	if !result.Passed {
		t.Fatalf("pending review failed: %+v", result)
	}

	stalePaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, stalePaths, scenario{ID: documentHistoryStaleScenarioID}); err != nil {
		t.Fatalf("seed stale synthesis: %v", err)
	}
	staleAnswer := "synthesis/history-review-stale.md is stale after sources/history-review/stale-current.md was updated. Projection freshness and provenance invalidated evidence show the stale state; no repair was performed."
	result, err = verifyScenarioTurn(ctx, stalePaths, scenario{ID: documentHistoryStaleScenarioID}, 1, staleAnswer, commonMetrics)
	if err != nil {
		t.Fatalf("verify stale synthesis: %v", err)
	}
	if !result.Passed {
		t.Fatalf("stale synthesis failed: %+v", result)
	}
}
