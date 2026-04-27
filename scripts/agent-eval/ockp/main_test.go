package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func TestParseRunConfigDefaultsParallelAndSharedCache(t *testing.T) {
	config, err := parseRunConfig(nil, &strings.Builder{})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if config.Parallel != defaultParallel {
		t.Fatalf("parallel = %d, want %d", config.Parallel, defaultParallel)
	}
	if config.CacheMode != cacheModeShared {
		t.Fatalf("cache mode = %q, want %q", config.CacheMode, cacheModeShared)
	}
}

func TestParseRunConfigRejectsInvalidParallelAndCacheMode(t *testing.T) {
	if _, err := parseRunConfig([]string{"--parallel", "0"}, &strings.Builder{}); err == nil {
		t.Fatal("expected invalid parallel error")
	}
	if _, err := parseRunConfig([]string{"--cache-mode", "bad"}, &strings.Builder{}); err == nil || !strings.Contains(err.Error(), "--cache-mode") {
		t.Fatalf("cache-mode error = %v, want validation error", err)
	}
}

func TestRunJobsPreservesDeterministicOrder(t *testing.T) {
	jobs := []evalJob{
		{Index: 0, Variant: "production", Scenario: scenario{ID: "first", Title: "First"}},
		{Index: 1, Variant: "production", Scenario: scenario{ID: "second", Title: "Second"}},
		{Index: 2, Variant: "variant-smoke", Scenario: scenario{ID: "third", Title: "Third"}},
	}
	results := runJobs(context.Background(), runConfig{Parallel: 3}, jobs, cacheConfig{Mode: cacheModeIsolated}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		if job.Index == 0 {
			time.Sleep(30 * time.Millisecond)
		}
		return jobResult{
			Variant:  job.Variant,
			Scenario: job.Scenario.ID,
			Status:   "completed",
		}
	})
	for i, result := range results {
		if result.Scenario != jobs[i].Scenario.ID {
			t.Fatalf("result %d scenario = %q, want %q", i, result.Scenario, jobs[i].Scenario.ID)
		}
	}
}

func TestRunJobsPreservesErrorIdentity(t *testing.T) {
	jobs := []evalJob{
		{Index: 0, Variant: "production", Scenario: scenario{ID: "ok", Title: "OK"}},
		{Index: 1, Variant: "variant-smoke", Scenario: scenario{ID: "bad", Title: "Bad"}},
	}
	results := runJobs(context.Background(), runConfig{Parallel: 2}, jobs, cacheConfig{Mode: cacheModeIsolated}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		result := jobResult{Variant: job.Variant, Scenario: job.Scenario.ID}
		if job.Scenario.ID == "bad" {
			result.Status = "failed"
			result.Error = "boom"
			return result
		}
		result.Status = "completed"
		return result
	})
	if results[1].Variant != "variant-smoke" || results[1].Scenario != "bad" || results[1].Error != "boom" {
		t.Fatalf("error result = %+v", results[1])
	}
}

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

func TestExecuteRunLabelsPopulatedVaultLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   populatedHeterogeneousScenarioID + "," + populatedFreshnessConflictScenarioID + "," + populatedSynthesisUpdateScenarioID,
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-populated-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		passed := true
		status := "completed"
		verification := verificationResult{Passed: true, DatabasePass: true, AssistantPass: true}
		if job.Scenario.ID == populatedHeterogeneousScenarioID {
			passed = false
			status = "failed"
			verification = verificationResult{
				Passed:        false,
				DatabasePass:  true,
				AssistantPass: false,
				Details:       "turn 1: final answer repeated polluted decoy claims",
			}
		}
		if job.Scenario.ID == populatedSynthesisUpdateScenarioID {
			passed = false
			status = "failed"
			verification = verificationResult{Passed: true, DatabasePass: true, AssistantPass: true}
		}
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        status,
			Passed:        passed,
			Metrics:       metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}},
			Verification:  verification,
			StartedAt:     now,
			CompletedAt:   &now,
		}
	})
	if err != nil {
		t.Fatalf("execute populated run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-populated-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != populatedLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("populated lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, populatedLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("populated report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "keep_as_reference" {
		t.Fatalf("decision = %q, want keep_as_reference", report.TargetedLaneSummary.Decision)
	}
	if !containsAllStrings(report.TargetedLaneSummary.PublicSurface, []string{"openclerk document", "openclerk retrieval"}) {
		t.Fatalf("public surface = %+v", report.TargetedLaneSummary.PublicSurface)
	}
	classifications := map[string]string{}
	for _, row := range report.TargetedLaneSummary.ScenarioClassifications {
		classifications[row.Scenario] = row.FailureClassification
	}
	if classifications[populatedHeterogeneousScenarioID] != "skill_guidance_or_eval_coverage" {
		t.Fatalf("heterogeneous classification = %q, want skill guidance", classifications[populatedHeterogeneousScenarioID])
	}
	if classifications[populatedFreshnessConflictScenarioID] != "none" {
		t.Fatalf("passing classification = %q, want none", classifications[populatedFreshnessConflictScenarioID])
	}
	if classifications[populatedSynthesisUpdateScenarioID] != "runner_execution_failure" {
		t.Fatalf("execution failure classification = %q, want runner_execution_failure", classifications[populatedSynthesisUpdateScenarioID])
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-populated-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + populatedLaneName + "`",
		"Release blocking: `false`",
		"## Targeted Lane Summary",
		"Decision: `keep_as_reference`",
		"Public surface: `openclerk document`, `openclerk retrieval`",
		"no promoted runner action, schema, migration, storage API, product behavior, or public OpenClerk interface",
		"`skill_guidance_or_eval_coverage`",
		"`runner_execution_failure`",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestExecuteRunLabelsAgentChosenPathLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   agentChosenExplicitScenarioID + "," + agentChosenMissingFieldsScenarioID + "," + agentChosenPathProposalScenarioID + "," + agentChosenAutonomousScenarioID,
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-agent-chosen-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		verification := verificationResult{Passed: true, DatabasePass: true, AssistantPass: true}
		passed := true
		status := "completed"
		if job.Scenario.ID == agentChosenAutonomousScenarioID {
			passed = false
			status = "failed"
			verification = verificationResult{Passed: false, DatabasePass: true, AssistantPass: false, Details: "turn 1: final answer omitted chosen path"}
		}
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        status,
			Passed:        passed,
			Metrics:       metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}},
			Verification:  verification,
			StartedAt:     now,
			CompletedAt:   &now,
		}
	})
	if err != nil {
		t.Fatalf("execute agent-chosen run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-agent-chosen-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != agentChosenPathLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("agent-chosen lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, agentChosenPathLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("agent-chosen report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "keep_as_reference" {
		t.Fatalf("decision = %q, want keep_as_reference", report.TargetedLaneSummary.Decision)
	}
	classifications := map[string]string{}
	for _, row := range report.TargetedLaneSummary.ScenarioClassifications {
		classifications[row.Scenario] = row.FailureClassification
	}
	if classifications[agentChosenPathProposalScenarioID] != "none" {
		t.Fatalf("proposal classification = %q, want none", classifications[agentChosenPathProposalScenarioID])
	}
	if classifications[agentChosenExplicitScenarioID] != "none" {
		t.Fatalf("explicit-fields classification = %q, want none", classifications[agentChosenExplicitScenarioID])
	}
	if classifications[agentChosenMissingFieldsScenarioID] != "none" {
		t.Fatalf("missing-fields classification = %q, want none", classifications[agentChosenMissingFieldsScenarioID])
	}
	if classifications[agentChosenAutonomousScenarioID] != "skill_guidance_or_eval_coverage" {
		t.Fatalf("autonomous classification = %q, want skill guidance", classifications[agentChosenAutonomousScenarioID])
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-agent-chosen-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + agentChosenPathLaneName + "`",
		"Release blocking: `false`",
		"Decision: `keep_as_reference`",
		"no promoted runner action, schema, migration, storage API, product behavior, public OpenClerk interface, or change to missing-path clarification",
		"`skill_guidance_or_eval_coverage`",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestExecuteRunLabelsPathTitleAutonomyPressureLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   pathTitleURLOnlyScenarioID + "," + pathTitleArtifactMissingHintsScenarioID + "," + pathTitleDuplicateRiskScenarioID,
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-path-title-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        "completed",
			Passed:        true,
			Metrics:       metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}},
			Verification:  verificationResult{Passed: true, DatabasePass: true, AssistantPass: true},
			StartedAt:     now,
			CompletedAt:   &now,
		}
	})
	if err != nil {
		t.Fatalf("execute path-title run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-path-title-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != pathTitleAutonomyLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("path-title lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, pathTitleAutonomyLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("path-title report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "evaluate_for_oc_iat" {
		t.Fatalf("decision = %q, want evaluate_for_oc_iat", report.TargetedLaneSummary.Decision)
	}
	if len(report.TargetedLaneSummary.ScenarioClassifications) != 3 {
		t.Fatalf("classifications = %d, want 3", len(report.TargetedLaneSummary.ScenarioClassifications))
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-path-title-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + pathTitleAutonomyLaneName + "`",
		"Release blocking: `false`",
		"Decision: `evaluate_for_oc_iat`",
		"no promoted runner action, schema, migration, skill behavior, storage API, product behavior, or public OpenClerk interface from this eval",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestExecuteRunLabelsDocumentThisIntakePressureLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   documentThisMissingFieldsScenarioID + "," + documentThisExplicitCreateScenarioID + "," + documentThisDuplicateCandidateScenarioID,
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-document-this-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        "completed",
			Passed:        true,
			Metrics:       metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}},
			Verification:  verificationResult{Passed: true, DatabasePass: true, AssistantPass: true},
			StartedAt:     now,
			CompletedAt:   &now,
		}
	})
	if err != nil {
		t.Fatalf("execute document-this run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-document-this-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != documentThisLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("document-this lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, documentThisLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("document-this report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "evaluate_for_oc_99z" {
		t.Fatalf("decision = %q, want evaluate_for_oc_99z", report.TargetedLaneSummary.Decision)
	}
	if len(report.TargetedLaneSummary.ScenarioClassifications) != 3 {
		t.Fatalf("classifications = %d, want 3", len(report.TargetedLaneSummary.ScenarioClassifications))
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-document-this-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + documentThisLaneName + "`",
		"Release blocking: `false`",
		"Decision: `evaluate_for_oc_99z`",
		"no promoted runner action, schema, migration, skill behavior, storage API, product behavior, or public OpenClerk interface from this eval",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestExecuteRunDefersPartialDocumentArtifactCandidateLane(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   candidateNoteFromPastedContentScenarioID + "," + candidateDuplicateRiskAsksScenarioID + "," + candidateLowConfidenceAsksScenarioID,
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-document-artifact-candidate-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        "completed",
			Passed:        true,
			Metrics:       metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}},
			Verification:  verificationResult{Passed: true, DatabasePass: true, AssistantPass: true},
			StartedAt:     now,
			CompletedAt:   &now,
		}
	})
	if err != nil {
		t.Fatalf("execute document artifact candidate run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-document-artifact-candidate-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != documentArtifactCandidateLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("document artifact candidate lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, documentArtifactCandidateLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("document artifact candidate report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "defer_for_candidate_quality_repair" {
		t.Fatalf("decision = %q, want defer_for_candidate_quality_repair", report.TargetedLaneSummary.Decision)
	}
	if len(report.TargetedLaneSummary.ScenarioClassifications) != 3 {
		t.Fatalf("classifications = %d, want 3", len(report.TargetedLaneSummary.ScenarioClassifications))
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-document-artifact-candidate-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + documentArtifactCandidateLaneName + "`",
		"Release blocking: `false`",
		"Decision: `defer_for_candidate_quality_repair`",
		"no promoted skill policy yet; repair candidate quality gaps before any propose-before-create skill behavior change",
		"`none`",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestDocumentArtifactCandidateDecisionRequiresCompleteScenarioCoverage(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(documentArtifactCandidateScenarioIDs()))
	for _, id := range documentArtifactCandidateScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := documentArtifactCandidateDecision(rows[:len(rows)-1]); decision != "defer_for_candidate_quality_repair" {
		t.Fatalf("partial decision = %q, want defer_for_candidate_quality_repair", decision)
	}
	if decision := documentArtifactCandidateDecision(rows); decision != "promote_propose_before_create_skill_policy" {
		t.Fatalf("complete decision = %q, want promote_propose_before_create_skill_policy", decision)
	}
	rows[0].FailureClassification = "candidate_quality_gap"
	if decision := documentArtifactCandidateDecision(rows); decision != "defer_for_candidate_quality_repair" {
		t.Fatalf("failing decision = %q, want defer_for_candidate_quality_repair", decision)
	}
}

func TestExecuteRunLabelsArtifactIngestionLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   strings.Join(artifactIngestionScenarioIDs(), ","),
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-heterogeneous-artifact-ingestion-pressure-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        "completed",
			Passed:        true,
			Metrics:       metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}},
			Verification:  verificationResult{Passed: true, DatabasePass: true, AssistantPass: true},
			StartedAt:     now,
			CompletedAt:   &now,
		}
	})
	if err != nil {
		t.Fatalf("execute artifact ingestion run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-heterogeneous-artifact-ingestion-pressure-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != artifactIngestionLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("artifact ingestion lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, artifactIngestionLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("artifact ingestion report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "keep_as_reference" {
		t.Fatalf("decision = %q, want keep_as_reference", report.TargetedLaneSummary.Decision)
	}
	if len(report.TargetedLaneSummary.ScenarioClassifications) != len(artifactIngestionScenarioIDs()) {
		t.Fatalf("classifications = %d, want %d", len(report.TargetedLaneSummary.ScenarioClassifications), len(artifactIngestionScenarioIDs()))
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-heterogeneous-artifact-ingestion-pressure-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + artifactIngestionLaneName + "`",
		"Release blocking: `false`",
		"Decision: `keep_as_reference`",
		"no promoted runner action, parser, schema, storage migration, direct create behavior, or public API change",
		"`none`",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestArtifactIngestionDecisionRequiresCompleteScenarioCoverage(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(artifactIngestionScenarioIDs()))
	for _, id := range artifactIngestionScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := artifactIngestionDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := artifactIngestionDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "runner_capability_gap"
	if decision := artifactIngestionDecision(rows); decision != "defer_for_artifact_runner_surface_design" {
		t.Fatalf("gap decision = %q, want defer_for_artifact_runner_surface_design", decision)
	}
	rows[0].FailureClassification = "skill_guidance"
	if decision := artifactIngestionDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestVerifyArtifactTranscriptRequiresTranscriptPathFilter(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: artifactTranscriptScenarioID}); err != nil {
		t.Fatalf("seed artifact transcript scenario: %v", err)
	}
	metrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		SearchPathFilterUsed: true,
		SearchPathPrefixes:   []string{"transcripts/"},
		EventTypeCounts:      map[string]int{},
	}
	answer := artifactTranscriptPath + " doc_id shows canonical markdown transcript evidence."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: artifactTranscriptScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify transcript: %v", err)
	}
	if !result.Passed {
		t.Fatalf("transcript verification failed: %+v", result)
	}

	missingPathFilter := metrics
	missingPathFilter.SearchPathPrefixes = nil
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: artifactTranscriptScenarioID}, 1, answer, missingPathFilter)
	if err != nil {
		t.Fatalf("verify transcript without path filter: %v", err)
	}
	if result.Passed {
		t.Fatalf("transcript verification passed without transcripts/ path filter: %+v", result)
	}
}

func TestVerifyArtifactInvoiceReceiptRequiresBothMetadataFilters(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: artifactInvoiceReceiptScenarioID}); err != nil {
		t.Fatalf("seed artifact invoice/receipt scenario: %v", err)
	}
	metrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		SearchMetadataFilterUsed: true,
		SearchMetadataFilters:    []string{"artifact_kind=invoice", "artifact_kind=receipt"},
		EventTypeCounts:          map[string]int{},
	}
	answer := artifactInvoicePath + " and " + artifactReceiptPath + " doc_id cite USD 1250.00, approval above USD 500, and USD 86.40."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: artifactInvoiceReceiptScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify invoice/receipt: %v", err)
	}
	if !result.Passed {
		t.Fatalf("invoice/receipt verification failed: %+v", result)
	}

	onlyInvoiceFilter := metrics
	onlyInvoiceFilter.SearchMetadataFilters = []string{"artifact_kind=invoice"}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: artifactInvoiceReceiptScenarioID}, 1, answer, onlyInvoiceFilter)
	if err != nil {
		t.Fatalf("verify invoice/receipt without receipt filter: %v", err)
	}
	if result.Passed {
		t.Fatalf("invoice/receipt verification passed without receipt metadata filter: %+v", result)
	}
}

func TestCandidateHeadingScenarioDoesNotLeakExpectedPath(t *testing.T) {
	sc := requireScenarioByID(t, candidateTitleAndPathFromHeadingScenarioID)
	if strings.Contains(sc.Prompt, candidateHeadingPath) {
		t.Fatalf("heading-derived candidate scenario leaked expected path %q:\n%s", candidateHeadingPath, sc.Prompt)
	}
	for _, want := range []string{
		"Choose a candidate path from the heading under notes/candidates/",
		"title from the heading",
		"Run openclerk document only with action validate",
		"Do not create the document.",
	} {
		if !strings.Contains(sc.Prompt, want) {
			t.Fatalf("heading-derived candidate scenario missing %q:\n%s", want, sc.Prompt)
		}
	}
}

func TestVerifyPathTitleURLOnlyRequiresStoredTitle(t *testing.T) {
	ctx := context.Background()
	metrics := metrics{
		AssistantCalls:    1,
		ToolCalls:         1,
		CommandExecutions: 1,
		EventTypeCounts:   map[string]int{},
	}
	wrongTitleBody := strings.TrimSpace(`---
type: source
---
# Wrong Stored Title

Path/title policy: autonomy pressure create then report.

Sources:
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/prompt-guidance
`) + "\n"
	body := strings.Replace(wrongTitleBody, "# Wrong Stored Title", "# OpenAI Harness and Prompt Guidance", 1)
	finalAnswer := "Created " + pathTitleURLOnlyPath + " titled " + pathTitleURLOnlyTitle + "."

	wrongTitlePaths := scenarioPaths(t.TempDir())
	wrongTitleCfg := runclient.Config{DatabasePath: wrongTitlePaths.DatabasePath}
	if err := createSeedDocument(ctx, wrongTitleCfg, pathTitleURLOnlyPath, "Wrong Stored Title", wrongTitleBody); err != nil {
		t.Fatalf("create wrong-title path/title source: %v", err)
	}
	result, err := verifyScenarioTurn(ctx, wrongTitlePaths, scenario{ID: pathTitleURLOnlyScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify wrong-title path/title source: %v", err)
	}
	if result.Passed || result.DatabasePass {
		t.Fatalf("path/title source with wrong stored title passed: %+v", result)
	}
	if !strings.Contains(result.Details, "expected stored title") {
		t.Fatalf("wrong-title failure details = %q", result.Details)
	}

	correctTitlePaths := scenarioPaths(t.TempDir())
	correctTitleCfg := runclient.Config{DatabasePath: correctTitlePaths.DatabasePath}
	if err := createSeedDocument(ctx, correctTitleCfg, pathTitleURLOnlyPath, pathTitleURLOnlyTitle, body); err != nil {
		t.Fatalf("create correct-title path/title source: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, correctTitlePaths, scenario{ID: pathTitleURLOnlyScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify correct-title path/title source: %v", err)
	}
	if !result.Passed {
		t.Fatalf("path/title source with correct stored title failed: %+v", result)
	}
}

func TestVerifyPathTitleDuplicateRiskRejectsAnyExtraPathTitleSource(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: pathTitleDuplicateRiskScenarioID}); err != nil {
		t.Fatalf("seed duplicate-risk scenario: %v", err)
	}
	metrics := metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	}
	finalAnswer := "Duplicate risk found at " + pathTitleDuplicateExistingPath + "; no new duplicate source was created."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: pathTitleDuplicateRiskScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify duplicate-risk baseline: %v", err)
	}
	if !result.Passed {
		t.Fatalf("duplicate-risk baseline failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "sources/path-title/alternate-openai-harness.md", "Alternate OpenAI Harness", "# Alternate\n"); err != nil {
		t.Fatalf("create alternate duplicate source: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: pathTitleDuplicateRiskScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify duplicate-risk alternate duplicate: %v", err)
	}
	if result.Passed || result.DatabasePass {
		t.Fatalf("duplicate-risk passed with alternate duplicate source: %+v", result)
	}
	if !strings.Contains(result.Details, "expected only the seeded path-title source document") {
		t.Fatalf("alternate duplicate failure details = %q", result.Details)
	}
}

func TestCodexArgsForSingleAndResumedTurns(t *testing.T) {
	cache := cacheConfig{Mode: cacheModeShared, RunRoot: "run-root"}
	single := scenario{ID: "single", Prompt: "single prompt"}
	singleArgs := codexArgsForTurn("codex", "run-root/production/single/repo", "run-root/production/single", single, scenarioTurn{Prompt: "single prompt"}, 1, "", cache)
	if !containsArgPair(singleArgs, "--add-dir", filepath.Join("run-root", "shared-cache")) {
		t.Fatalf("single args missing shared cache writable root: %v", singleArgs)
	}
	if !containsValue(singleArgs, "--ephemeral") {
		t.Fatalf("single args must use --ephemeral: %v", singleArgs)
	}
	if !containsValue(singleArgs, "--ignore-user-config") {
		t.Fatalf("single args missing --ignore-user-config: %v", singleArgs)
	}

	multi := scenario{ID: "multi", Turns: []scenarioTurn{{Prompt: "first"}, {Prompt: "second"}}}
	firstArgs := codexArgsForTurn("codex", "run-root/production/multi/repo", "run-root/production/multi", multi, scenarioTurn{Prompt: "first"}, 1, "", cache)
	if !containsValue(firstArgs, "--ignore-user-config") {
		t.Fatalf("first multi-turn args missing --ignore-user-config: %v", firstArgs)
	}
	if containsValue(firstArgs, "--ephemeral") {
		t.Fatalf("first multi-turn args must persist the session: %v", firstArgs)
	}
	resumeArgs := codexArgsForTurn("codex", "run-root/production/multi/repo", "run-root/production/multi", multi, scenarioTurn{Prompt: "second"}, 2, "session-123", cache)
	if containsValue(resumeArgs, "--ephemeral") {
		t.Fatalf("resume args must not use --ephemeral: %v", resumeArgs)
	}
	if !containsValue(resumeArgs, "--ignore-user-config") {
		t.Fatalf("resume args missing --ignore-user-config: %v", resumeArgs)
	}
	if !containsValue(resumeArgs, "resume") || !containsValue(resumeArgs, "session-123") {
		t.Fatalf("resume args must persist the multi-turn session: %v", resumeArgs)
	}
}

func TestEvalEnvSharedAndIsolatedCache(t *testing.T) {
	paths := scenarioPaths(filepath.Join("run-root", "production", "create", "repo"))
	shared := strings.Join(evalEnv(filepath.Join("run-root", "production", "create"), paths, cacheConfig{Mode: cacheModeShared, RunRoot: "run-root"}), "\n")
	for _, want := range []string{
		"OPENCLERK_DATABASE_PATH=" + filepath.Join("run-root", "production", "create", "repo", ".openclerk-eval", "openclerk.db"),
		"GOCACHE=" + filepath.Join("run-root", "shared-cache", "gocache"),
		"GOMODCACHE=" + filepath.Join("run-root", "shared-cache", "gomodcache"),
	} {
		if !strings.Contains(shared, want) {
			t.Fatalf("shared env missing %q in %s", want, shared)
		}
	}
	isolated := strings.Join(evalEnv(filepath.Join("run-root", "production", "create"), paths, cacheConfig{Mode: cacheModeIsolated, RunRoot: "run-root"}), "\n")
	if !strings.Contains(isolated, "GOCACHE="+filepath.Join("run-root", "production", "create", "gocache")) {
		t.Fatalf("isolated env = %s", isolated)
	}
}

func TestCopyRepoSkipsEvalContextContamination(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "copy")
	files := map[string]string{
		"README.md":                                      "kept",
		"AGENTS.md":                                      "root instructions",
		".agents/skills/openclerk/SKILL.md":              "stale skill",
		".dolt/config":                                   "dolt",
		"docs/evals/results/previous.md":                 "report",
		"scripts/agent-eval/ockp/main.go":                "harness",
		"scripts/agent-eval/ockp/nested/fixture.txt":     "harness fixture",
		"scripts/agent-eval/other-harness/kept-file.txt": "other harness",
	}
	for path, content := range files {
		target := filepath.Join(src, path)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	if err := copyRepo(src, dst); err != nil {
		t.Fatalf("copy repo: %v", err)
	}
	for _, want := range []string{"README.md", "scripts/agent-eval/other-harness/kept-file.txt"} {
		if _, err := os.Stat(filepath.Join(dst, want)); err != nil {
			t.Fatalf("expected copied %s: %v", want, err)
		}
	}
	for _, skipped := range []string{
		"AGENTS.md",
		".agents/skills/openclerk/SKILL.md",
		".dolt/config",
		"docs/evals/results/previous.md",
		"scripts/agent-eval/ockp/main.go",
		"scripts/agent-eval/ockp/nested/fixture.txt",
	} {
		if _, err := os.Stat(filepath.Join(dst, skipped)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be skipped, stat err=%v", skipped, err)
		}
	}
}

func TestInstallVariantInstallsExactSkillAndNoAgentsFile(t *testing.T) {
	repoRoot := t.TempDir()
	repoDir := t.TempDir()
	sourceSkill := []byte("# OpenClerk\n\nUse the runner.\n")
	skillDir := filepath.Join(repoRoot, "skills", "openclerk")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), sourceSkill, 0o644); err != nil {
		t.Fatalf("write source skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill_payload_test.go"), []byte("package ignored\n"), 0o644); err != nil {
		t.Fatalf("write source test helper: %v", err)
	}

	if err := installVariant(repoRoot, repoDir, productionVariant); err != nil {
		t.Fatalf("install production variant: %v", err)
	}
	installedSkill, err := os.ReadFile(filepath.Join(repoDir, ".agents", "skills", "openclerk", "SKILL.md"))
	if err != nil {
		t.Fatalf("read installed skill: %v", err)
	}
	if !bytes.Equal(installedSkill, sourceSkill) {
		t.Fatalf("installed skill bytes = %q, want %q", installedSkill, sourceSkill)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("installVariant must not create AGENTS.md, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(repoDir, ".agents", "skills", "openclerk", "skill_payload_test.go")); !os.IsNotExist(err) {
		t.Fatalf("installVariant must not install Go test payload, stat err=%v", err)
	}
}

func TestVariantSelectionProductionOnly(t *testing.T) {
	variants := selectedVariants(runConfig{})
	if len(variants) != 1 || variants[0] != productionVariant {
		t.Fatalf("default variants = %v, want [%s]", variants, productionVariant)
	}

	repoRoot := t.TempDir()
	repoDir := t.TempDir()
	if err := installVariant(repoRoot, repoDir, "unknown"); err == nil {
		t.Fatal("expected unknown variant error")
	}
}

func TestEvalEnvOverridesHostOpenClerkPaths(t *testing.T) {
	t.Setenv("CODEX_HOME", "/host/.codex")
	t.Setenv("OPENCLERK_DATA_DIR", "/host/data")
	t.Setenv("OPENCLERK_DATABASE_PATH", "/host/openclerk.db")
	t.Setenv("OPENCLERK_VAULT_ROOT", "/host/vault")

	runDir := filepath.Join(t.TempDir(), "run")
	paths := evalPaths{
		DatabasePath: filepath.Join(runDir, "openclerk.db"),
	}
	env := evalEnv(runDir, paths, cacheConfig{})
	got := map[string]string{}
	for _, entry := range env {
		key, value, found := strings.Cut(entry, "=")
		if found {
			got[key] = value
		}
	}
	if got["CODEX_HOME"] != filepath.Join(runDir, "codex-home") {
		t.Fatalf("CODEX_HOME = %q, want %q", got["CODEX_HOME"], filepath.Join(runDir, "codex-home"))
	}
	if got["ZDOTDIR"] != filepath.Join(runDir, "zdotdir") {
		t.Fatalf("ZDOTDIR = %q, want %q", got["ZDOTDIR"], filepath.Join(runDir, "zdotdir"))
	}
	if _, ok := got["OPENCLERK_DATA_DIR"]; ok {
		t.Fatalf("OPENCLERK_DATA_DIR should not be exported: %v", got)
	}
	if got["OPENCLERK_DATABASE_PATH"] != paths.DatabasePath {
		t.Fatalf("OPENCLERK_DATABASE_PATH = %q, want %q", got["OPENCLERK_DATABASE_PATH"], paths.DatabasePath)
	}
	if _, ok := got["OPENCLERK_VAULT_ROOT"]; ok {
		t.Fatalf("OPENCLERK_VAULT_ROOT should not be exported: %v", got)
	}
}

func TestPromptInputPreflightFlagsOpenClerkAgentsInstructions(t *testing.T) {
	clean := "### Project Skills\n- OpenClerk: Use OpenClerk for local-first knowledge-plane tasks through the installed openclerk JSON runner. Bootstrap no-tools rule for routine OpenClerk requests - if required fields are missing, if creating or updating a document but document path, title, or body is missing, this description is complete; respond with exactly one no-tools assistant answer that names the missing fields and asks the user to provide them. Do not open this skill file, run commands, use tools, call the runner, or inspect files for those validation cases. If a numeric limit is negative such as limit -3, or if the user asks to bypass the runner through SQLite, HTTP, MCP, legacy or source-built paths, or unsupported transports, reject final-answer-only without opening this skill file, running commands, or using tools. For valid create or update requests, use only openclerk document or openclerk retrieval JSON results; never use rg --files, find, ls, direct vault inspection, direct file edits, or repo search to verify routine work. (file: /tmp/repo/.agents/skills/openclerk/SKILL.md)\n"
	if !containsOpenClerkSkillDiscovery(clean) {
		t.Fatalf("clean skill discovery is missing the OpenClerk skill marker: %s", clean)
	}
	if containsOpenClerkAgentsInstructions(clean) {
		t.Fatalf("clean skill discovery was flagged: %s", clean)
	}
	if !containsOpenClerkBootstrapRejectionGuidance(clean) {
		t.Fatalf("clean skill discovery is missing bootstrap rejection guidance: %s", clean)
	}

	missingBootstrap := "### Project Skills\n- openclerk: Use OpenClerk. (file: /tmp/repo/.agents/skills/openclerk/SKILL.md)\n"
	if containsOpenClerkBootstrapRejectionGuidance(missingBootstrap) {
		t.Fatalf("incomplete skill discovery passed bootstrap guidance check: %s", missingBootstrap)
	}

	contaminated := "# AGENTS.md instructions for /tmp/repo\n\nUse `openclerk document` with create_document JSON action names.\n"
	if !containsOpenClerkAgentsInstructions(contaminated) {
		t.Fatalf("contaminated AGENTS block was not flagged: %s", contaminated)
	}
}

func TestPrepareRunDirCreatesAuthOnlyCodexHome(t *testing.T) {
	hostCodexHome := filepath.Join(t.TempDir(), "host-codex-home")
	if err := os.MkdirAll(filepath.Join(hostCodexHome, "sessions"), 0o755); err != nil {
		t.Fatalf("mkdir host codex home: %v", err)
	}
	for name, content := range map[string]string{
		"auth.json":          `{"token":"test"}`,
		"config.toml":        "model = \"gpt-5.4-mini\"\n",
		"installation_id":    "test-installation\n",
		"sessions/old.jsonl": "old session\n",
	} {
		if err := os.WriteFile(filepath.Join(hostCodexHome, name), []byte(content), 0o600); err != nil {
			t.Fatalf("write host codex home %s: %v", name, err)
		}
	}
	t.Setenv("CODEX_HOME", hostCodexHome)

	runDir := filepath.Join(t.TempDir(), "run")
	if err := prepareRunDir(runDir, cacheConfig{}); err != nil {
		t.Fatalf("prepare run dir: %v", err)
	}
	for _, want := range []string{
		filepath.Join(runDir, "codex-home"),
		filepath.Join(runDir, "zdotdir"),
		filepath.Join(runDir, "tmp"),
	} {
		if info, err := os.Stat(want); err != nil || !info.IsDir() {
			t.Fatalf("expected directory %s, stat err=%v", want, err)
		}
	}
	got, err := os.ReadFile(filepath.Join(runDir, "codex-home", "auth.json"))
	if err != nil {
		t.Fatalf("read seeded auth: %v", err)
	}
	if string(got) != `{"token":"test"}` {
		t.Fatalf("seeded auth = %q, want copied auth", got)
	}
	homeInfo, err := os.Stat(filepath.Join(runDir, "codex-home"))
	if err != nil {
		t.Fatalf("stat eval codex home: %v", err)
	}
	if homeInfo.Mode().Perm()&0o077 != 0 {
		t.Fatalf("eval codex home permissions = %v, want no group/other access", homeInfo.Mode().Perm())
	}
	for _, unwanted := range []string{"config.toml", "installation_id", filepath.Join("sessions", "old.jsonl")} {
		if _, err := os.Stat(filepath.Join(runDir, "codex-home", unwanted)); !os.IsNotExist(err) {
			t.Fatalf("unexpected copied %s: stat error = %v", unwanted, err)
		}
	}
}

func TestSetupEvalCodexHomeRequiresAuth(t *testing.T) {
	err := setupEvalCodexHomeFromSource(filepath.Join(t.TempDir(), "codex-home"), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "run codex login") {
		t.Fatalf("setupEvalCodexHomeFromSource() error = %v, want login guidance", err)
	}
}

func TestGoListDoesNotExposePublicLocalClientPackage(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..", "..")
	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list ./...: %v\n%s", err, string(output))
	}
	forbiddenPackage := "github.com/yazanabuashour/openclerk/" + "client/" + "local"
	if strings.Contains(string(output), forbiddenPackage) {
		t.Fatalf("go list exposes removed package %q:\n%s", forbiddenPackage, string(output))
	}
}

func TestRepositoryDoesNotDocumentRemovedPublicClientInterface(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..", "..")
	forbidden := []string{
		"client/" + "local",
		"local." + "Open" + "Client",
		"Open" + "Client",
		"sdk-" + "baseline",
		"Local Go " + "S" + "DK",
		"direct-local Go " + "package",
		"examples/openclerk-" + "client",
		"examples/openclerk-" + "query",
	}
	if err := filepath.WalkDir(repoRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".beads", ".openclerk-eval":
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(path) {
		case ".go", ".md", ".json", ".yaml", ".yml", ".toml", ".txt":
		default:
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(content)
		displayPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			displayPath = path
		}
		for _, stale := range forbidden {
			if strings.Contains(text, stale) {
				t.Fatalf("%s contains removed public client interface text %q", displayPath, stale)
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("scan repo: %v", err)
	}
}

func TestParseMetricsFromCodexJSONLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	log := strings.Join([]string{
		`{"type":"thread.started","thread_id":"session-123"}`,
		`{"type":"item.completed","item":{"type":"agent_message","text":"done"},"usage":{"input_tokens":100,"cached_input_tokens":30,"output_tokens":12}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files /Users/example/.codex"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files /home/runner/.codex"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files C:\\Users\\runner\\.codex"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"search\",\"search\":{\"text\":\"runner\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"search\",\"search\":{\"text\":\"runner\",\"path_prefix\":\"notes/rag/\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"search\",\"search\":{\"text\":\"runner\",\"metadata_key\":\"rag_scope\",\"metadata_value\":\"active-policy\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\"}}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"get_document\",\"doc_id\":\"doc_1\"}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"inspect_layout\"}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\n  \"document_links\",\"doc_id\":\"doc_1\"}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\n  \"graph_neighborhood\",\"doc_id\":\"doc_1\"}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"records_lookup\",\"records\":{\"text\":\"runner\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"decisions_lookup\",\"decisions\":{\"text\":\"runner\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"decision_record\",\"decision_id\":\"adr-runner\"}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"provenance_events\",\"provenance\":{\"ref_kind\":\"document\",\"ref_id\":\"doc_alpha\",\"limit\":10}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"projection_states\",\"projection\":{\"limit\":10}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"/bin/zsh -lc \"printf '%s' '{\\\"action\\\":\\\"search\\\",\\\"search\\\":{\\\"text\\\":\\\"runner\\\"}}' | openclerk retrieval\""}}`,
		`not json`,
	}, "\n")
	if err := os.WriteFile(path, []byte(log), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}
	parsed, err := parseMetrics(path)
	if err != nil {
		t.Fatalf("parse metrics: %v", err)
	}
	if parsed.sessionID != "session-123" || parsed.finalMessage != "done" {
		t.Fatalf("parsed = %+v", parsed)
	}
	if parsed.metrics.ToolCalls != 19 || parsed.metrics.CommandExecutions != 19 || parsed.metrics.AssistantCalls != 1 {
		t.Fatalf("metrics = %+v", parsed.metrics)
	}
	if !parsed.metrics.BroadRepoSearch {
		t.Fatalf("expected broad repo search metric")
	}
	forbiddenEvidencePaths := []string{"/Users/example", "/home/runner", `C:\Users\runner`}
	for _, evidence := range parsed.metrics.BroadRepoSearchEvidence {
		for _, forbidden := range forbiddenEvidencePaths {
			if strings.Contains(evidence, forbidden) {
				t.Fatalf("evidence was not sanitized: %v", parsed.metrics.BroadRepoSearchEvidence)
			}
		}
	}
	if parsed.metrics.NonCachedInputTokens == nil || *parsed.metrics.NonCachedInputTokens != 70 || parsed.metrics.OutputTokens == nil || *parsed.metrics.OutputTokens != 12 {
		t.Fatalf("token metrics = %+v", parsed.metrics)
	}
	if !provenanceEventRefIDsInclude(parsed.metrics.ProvenanceEventRefIDs, "doc_alpha") {
		t.Fatalf("expected provenance event ref id in %+v", parsed.metrics)
	}
	if !decisionRecordIDsInclude(parsed.metrics.DecisionRecordIDs, "adr-runner") {
		t.Fatalf("expected decision record id in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.ListDocumentPathPrefixes, []string{"synthesis/"}) {
		t.Fatalf("expected list document path prefix in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.SearchPathPrefixes, []string{"notes/rag/"}) {
		t.Fatalf("expected search path prefix in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.SearchMetadataFilters, []string{"rag_scope=active-policy"}) {
		t.Fatalf("expected search metadata filter in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.GetDocumentDocIDs, []string{"doc_1"}) {
		t.Fatalf("expected get document doc id in %+v", parsed.metrics)
	}
	for name, used := range map[string]bool{
		"search":                 parsed.metrics.SearchUsed,
		"search_unfiltered":      parsed.metrics.SearchUnfilteredUsed,
		"search_path_filter":     parsed.metrics.SearchPathFilterUsed,
		"search_metadata_filter": parsed.metrics.SearchMetadataFilterUsed,
		"list_documents":         parsed.metrics.ListDocumentsUsed,
		"get_document":           parsed.metrics.GetDocumentUsed,
		"inspect_layout":         parsed.metrics.InspectLayoutUsed,
		"document_links":         parsed.metrics.DocumentLinksUsed,
		"graph_neighborhood":     parsed.metrics.GraphNeighborhoodUsed,
		"records_lookup":         parsed.metrics.RecordsLookupUsed,
		"decisions_lookup":       parsed.metrics.DecisionsLookupUsed,
		"decision_record":        parsed.metrics.DecisionRecordUsed,
		"provenance_events":      parsed.metrics.ProvenanceEventsUsed,
		"projection_states":      parsed.metrics.ProjectionStatesUsed,
	} {
		if !used {
			t.Fatalf("expected %s action metric in %+v", name, parsed.metrics)
		}
	}
}

func TestAggregateMetricsRequiresAllTurnsExposeUsage(t *testing.T) {
	input := 100
	cached := 10
	nonCached := 90
	output := 20
	aggregated := aggregateMetrics([]turnResult{
		{Metrics: metrics{UsageExposed: true, InputTokens: &input, CachedInputTokens: &cached, NonCachedInputTokens: &nonCached, OutputTokens: &output, EventTypeCounts: map[string]int{"message": 1}}},
		{Metrics: metrics{EventTypeCounts: map[string]int{"tool_call": 1}}},
	})
	if aggregated.UsageExposed {
		t.Fatalf("usage should not be exposed unless all turns expose usage: %+v", aggregated)
	}
}

func TestFinalAnswerOnlyAndProductionGates(t *testing.T) {
	prodTokens := 80
	results := []jobResult{}
	for _, scenarioID := range scenarioIDs() {
		tools := 2
		commands := 2
		if isFinalAnswerOnlyValidationScenario(scenarioID) {
			tools = 0
			commands = 0
		}
		results = append(results, comparisonResult(productionVariant, scenarioID, true, tools, commands, 1, prodTokens))
	}
	summary := buildProductionGateSummary(results)
	if summary == nil {
		t.Fatal("missing summary")
	}
	criteria := map[string]bool{}
	for _, criterion := range summary.Criteria {
		criteria[criterion.Name] = criterion.Passed
	}
	for _, name := range []string{"production_passes_all_scenarios", "validation_scenarios_are_final_answer_only", "no_direct_sqlite_access"} {
		if !criteria[name] {
			t.Fatalf("%s failed in %+v", name, summary.Criteria)
		}
	}
}

func TestCreateNoteScenarioForbidsBroadInspection(t *testing.T) {
	prompt := ""
	for _, sc := range allScenarios() {
		if sc.ID == "create-note" {
			prompt = sc.Prompt
			break
		}
	}
	if prompt == "" {
		t.Fatal("create-note scenario missing")
	}
	for _, want := range []string{
		"Use only OpenClerk runner document JSON results",
		"do not use rg",
		"find",
		"ls",
		"repo search",
		"direct vault inspection",
		"direct file edits",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("create-note prompt missing %q: %s", want, prompt)
		}
	}
}

func TestVerifyFinalAnswerOnlyRequiresRejectionAndNoTools(t *testing.T) {
	sc := scenario{ID: "negative-limit-reject"}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	if result := verifyFinalAnswerOnly(sc, "Done.", noTools); result.Passed {
		t.Fatalf("non-rejection answer passed: %+v", result)
	}
	withTool := metrics{AssistantCalls: 1, ToolCalls: 1, CommandExecutions: 1, EventTypeCounts: map[string]int{}}
	if result := verifyFinalAnswerOnly(sc, "The limit is invalid because it cannot be negative.", withTool); result.Passed {
		t.Fatalf("tool-using rejection passed: %+v", result)
	}
	if result := verifyFinalAnswerOnly(sc, "The limit is invalid because it cannot be negative.", noTools); !result.Passed {
		t.Fatalf("valid direct rejection failed: %+v", result)
	}

	unsupported := scenario{ID: "unsupported-lower-level-reject"}
	if result := verifyFinalAnswerOnly(unsupported, "I can\u2019t bypass OpenClerk runner or query SQLite directly for this task.", noTools); !result.Passed {
		t.Fatalf("curly apostrophe rejection failed: %+v", result)
	}
	if result := verifyFinalAnswerOnly(unsupported, "I can't do that for routine OpenClerk knowledge tasks.", noTools); !result.Passed {
		t.Fatalf("plain refusal rejection failed: %+v", result)
	}
	if result := verifyFinalAnswerOnly(unsupported, "I can't query SQLite directly for routine OpenClerk knowledge tasks.", noTools); !result.Passed {
		t.Fatalf("direct SQLite rejection failed: %+v", result)
	}

	unsupportedTransport := scenario{ID: "unsupported-transport-reject"}
	if result := verifyFinalAnswerOnly(unsupportedTransport, "I can't bypass OpenClerk runner through a legacy command path or MCP-style transport path for this routine task.", noTools); !result.Passed {
		t.Fatalf("unsupported transport bypass rejection failed: %+v", result)
	}

	missingPath := scenario{ID: "missing-document-path-reject"}
	message := "I can't create the document yet because document.path is missing. Provide the target path and I can continue with the OpenClerk runner."
	if result := verifyFinalAnswerOnly(missingPath, message, noTools); !result.Passed {
		t.Fatalf("requires path rejection failed: %+v", result)
	}

	missingPolicyFields := scenario{ID: agentChosenMissingFieldsScenarioID}
	message = "I can't create the document yet because path, title, and type are missing. Provide the missing path, title, and document type and I can continue."
	if result := verifyFinalAnswerOnly(missingPolicyFields, message, noTools); !result.Passed {
		t.Fatalf("path/title/type clarification failed: %+v", result)
	}

	missingArtifactHints := scenario{ID: pathTitleArtifactMissingHintsScenarioID}
	message = "I can't ingest the source yet because source.path_hint and source.asset_path_hint are missing. Provide source.path_hint and source.asset_path_hint and I can continue."
	if result := verifyFinalAnswerOnly(missingArtifactHints, message, noTools); !result.Passed {
		t.Fatalf("artifact hint clarification failed: %+v", result)
	}

	documentThisMissingHints := scenario{ID: documentThisSourceURLMissingHintsScenarioID}
	message = "I can't ingest the source yet because source.path_hint and source.asset_path_hint are missing. Provide source.path_hint and source.asset_path_hint and I can continue."
	if result := verifyFinalAnswerOnly(documentThisMissingHints, message, noTools); !result.Passed {
		t.Fatalf("document-this source hint clarification failed: %+v", result)
	}
}

func TestVerifyMissingFieldClarificationRequiresNoToolsAndFieldNames(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "mt-incomplete-then-create"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", "I can't create the note yet because path, title, and body are missing. Provide them and I can continue.", noTools, []string{"path", "title", "body"})
	if err != nil {
		t.Fatalf("verify clarification: %v", err)
	}
	if !result.Passed {
		t.Fatalf("valid clarification failed: %+v", result)
	}

	withTools := metrics{AssistantCalls: 1, ToolCalls: 1, CommandExecutions: 1, EventTypeCounts: map[string]int{}}
	result, err = verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", "I can't create the note yet because path, title, and body are missing. Provide them and I can continue.", withTools, []string{"path", "title", "body"})
	if err != nil {
		t.Fatalf("verify clarification with tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("tool-using clarification passed: %+v", result)
	}

	result, err = verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", "I need more information.", noTools, []string{"path", "title", "body"})
	if err != nil {
		t.Fatalf("verify incomplete clarification: %v", err)
	}
	if result.Passed {
		t.Fatalf("missing-fields clarification passed without naming fields: %+v", result)
	}

	result, err = verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", "I need to use the HTTP transport for path, title, and body.", noTools, []string{"path", "title", "body"})
	if err != nil {
		t.Fatalf("verify non-clarifying message: %v", err)
	}
	if result.Passed {
		t.Fatalf("non-clarifying message passed: %+v", result)
	}
}

func TestScenarioIDsIncludeADRProofObligations(t *testing.T) {
	ids := map[string]bool{}
	for _, id := range scenarioIDs() {
		ids[id] = true
	}
	for _, want := range []string{"answer-filing", ragRetrievalScenarioID, docsNavigationScenarioID, graphSemanticsScenarioID, memoryRouterScenarioID, configuredLayoutScenarioID, invalidLayoutScenarioID, sourceURLUpdateDuplicateScenarioID, sourceURLUpdateSameSHAScenarioID, sourceURLUpdateChangedScenarioID, sourceURLUpdateConflictScenarioID, synthesisCandidatePressureScenarioID, synthesisSourceSetPressureScenarioID, decisionRecordVsDocsScenarioID, decisionSupersessionScenarioID, sourceAuditRepairScenarioID, sourceAuditConflictScenarioID, documentHistoryInspectScenarioID, documentHistoryDiffScenarioID, documentHistoryRestoreScenarioID, documentHistoryPendingScenarioID, documentHistoryStaleScenarioID, populatedHeterogeneousScenarioID, populatedFreshnessConflictScenarioID, populatedSynthesisUpdateScenarioID, agentChosenExplicitScenarioID, agentChosenMissingFieldsScenarioID, agentChosenPathProposalScenarioID, agentChosenAutonomousScenarioID, agentChosenSynthesisScenarioID, agentChosenAmbiguousScenarioID, agentChosenUserPathScenarioID, pathTitleURLOnlyScenarioID, pathTitleArtifactMissingHintsScenarioID, pathTitleMultiSourceDuplicateScenarioID, pathTitleExplicitOverridesScenarioID, pathTitleDuplicateRiskScenarioID, pathTitleMetadataAuthorityScenarioID, documentThisMissingFieldsScenarioID, documentThisExplicitCreateScenarioID, documentThisSourceURLMissingHintsScenarioID, documentThisExplicitOverridesScenarioID, documentThisDuplicateCandidateScenarioID, documentThisExistingUpdateScenarioID, documentThisSynthesisFreshnessScenarioID, candidateNoteFromPastedContentScenarioID, candidateTitleAndPathFromHeadingScenarioID, candidateMixedSourceSummaryScenarioID, candidateExplicitOverridesWinScenarioID, candidateDuplicateRiskAsksScenarioID, candidateLowConfidenceAsksScenarioID, candidateBodyFaithfulnessScenarioID, artifactPDFSourceURLScenarioID, artifactPDFNaturalIntentScenarioID, artifactTranscriptScenarioID, artifactInvoiceReceiptScenarioID, artifactMixedSynthesisScenarioID, artifactSourceMissingHintsScenarioID, artifactUnsupportedVideoScenarioID, artifactBypassScenarioID, mtSynthesisDriftPressureScenarioID, "stale-synthesis-update", "promoted-record-vs-docs", "unsupported-transport-reject"} {
		if !ids[want] {
			t.Fatalf("scenarioIDs missing %q in %v", want, scenarioIDs())
		}
	}
}

func TestDefaultScenarioSelectionExcludesPopulatedTargetedLane(t *testing.T) {
	defaultIDs := map[string]bool{}
	for _, scenario := range selectedScenarios(runConfig{}) {
		defaultIDs[scenario.ID] = true
	}
	for _, id := range []string{populatedHeterogeneousScenarioID, populatedFreshnessConflictScenarioID, populatedSynthesisUpdateScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted populated scenario %q", id)
		}
	}
	for _, id := range []string{documentHistoryInspectScenarioID, documentHistoryDiffScenarioID, documentHistoryRestoreScenarioID, documentHistoryPendingScenarioID, documentHistoryStaleScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted document history scenario %q", id)
		}
	}
	for _, id := range []string{agentChosenExplicitScenarioID, agentChosenMissingFieldsScenarioID, agentChosenPathProposalScenarioID, agentChosenAutonomousScenarioID, agentChosenSynthesisScenarioID, agentChosenAmbiguousScenarioID, agentChosenUserPathScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted agent-chosen path scenario %q", id)
		}
	}
	for _, id := range []string{pathTitleURLOnlyScenarioID, pathTitleArtifactMissingHintsScenarioID, pathTitleMultiSourceDuplicateScenarioID, pathTitleExplicitOverridesScenarioID, pathTitleDuplicateRiskScenarioID, pathTitleMetadataAuthorityScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted path-title scenario %q", id)
		}
	}
	for _, id := range []string{sourceURLUpdateDuplicateScenarioID, sourceURLUpdateSameSHAScenarioID, sourceURLUpdateChangedScenarioID, sourceURLUpdateConflictScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted source URL update scenario %q", id)
		}
	}
	for _, id := range []string{documentThisMissingFieldsScenarioID, documentThisExplicitCreateScenarioID, documentThisSourceURLMissingHintsScenarioID, documentThisExplicitOverridesScenarioID, documentThisDuplicateCandidateScenarioID, documentThisExistingUpdateScenarioID, documentThisSynthesisFreshnessScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted document-this scenario %q", id)
		}
	}
	for _, id := range []string{candidateNoteFromPastedContentScenarioID, candidateTitleAndPathFromHeadingScenarioID, candidateMixedSourceSummaryScenarioID, candidateExplicitOverridesWinScenarioID, candidateDuplicateRiskAsksScenarioID, candidateLowConfidenceAsksScenarioID, candidateBodyFaithfulnessScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted document artifact candidate scenario %q", id)
		}
	}
	for _, id := range artifactIngestionScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted artifact ingestion scenario %q", id)
		}
	}
	selected := selectedScenarioIDs(runConfig{Scenario: populatedHeterogeneousScenarioID + "," + populatedFreshnessConflictScenarioID + "," + populatedSynthesisUpdateScenarioID})
	lane, releaseBlocking := reportLane(selected)
	if lane != populatedLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, populatedLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: documentHistoryInspectScenarioID + "," + documentHistoryDiffScenarioID + "," + documentHistoryRestoreScenarioID + "," + documentHistoryPendingScenarioID + "," + documentHistoryStaleScenarioID + ",missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject"})
	lane, releaseBlocking = reportLane(selected)
	if lane != documentHistoryLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, documentHistoryLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: agentChosenExplicitScenarioID + "," + agentChosenMissingFieldsScenarioID + "," + agentChosenPathProposalScenarioID + "," + agentChosenAutonomousScenarioID + "," + agentChosenSynthesisScenarioID + "," + agentChosenAmbiguousScenarioID + "," + agentChosenUserPathScenarioID + ",missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject"})
	lane, releaseBlocking = reportLane(selected)
	if lane != agentChosenPathLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, agentChosenPathLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: pathTitleURLOnlyScenarioID + "," + pathTitleArtifactMissingHintsScenarioID + "," + pathTitleMultiSourceDuplicateScenarioID + "," + pathTitleExplicitOverridesScenarioID + "," + pathTitleDuplicateRiskScenarioID + "," + pathTitleMetadataAuthorityScenarioID})
	lane, releaseBlocking = reportLane(selected)
	if lane != pathTitleAutonomyLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, pathTitleAutonomyLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: sourceURLUpdateDuplicateScenarioID + "," + sourceURLUpdateSameSHAScenarioID + "," + sourceURLUpdateChangedScenarioID + "," + sourceURLUpdateConflictScenarioID})
	lane, releaseBlocking = reportLane(selected)
	if lane != sourceURLUpdateLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, sourceURLUpdateLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: documentThisMissingFieldsScenarioID + "," + documentThisExplicitCreateScenarioID + "," + documentThisSourceURLMissingHintsScenarioID + "," + documentThisExplicitOverridesScenarioID + "," + documentThisDuplicateCandidateScenarioID + "," + documentThisExistingUpdateScenarioID + "," + documentThisSynthesisFreshnessScenarioID})
	lane, releaseBlocking = reportLane(selected)
	if lane != documentThisLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, documentThisLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: candidateNoteFromPastedContentScenarioID + "," + candidateTitleAndPathFromHeadingScenarioID + "," + candidateMixedSourceSummaryScenarioID + "," + candidateExplicitOverridesWinScenarioID + "," + candidateDuplicateRiskAsksScenarioID + "," + candidateLowConfidenceAsksScenarioID + "," + candidateBodyFaithfulnessScenarioID})
	lane, releaseBlocking = reportLane(selected)
	if lane != documentArtifactCandidateLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, documentArtifactCandidateLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(artifactIngestionScenarioIDs(), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != artifactIngestionLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, artifactIngestionLaneName)
	}
}

func TestSourceURLUpdateFixturePromptRendering(t *testing.T) {
	fixtures := startSourceURLUpdateFixtures(sourceURLUpdateChangedScenarioID)
	if fixtures == nil {
		t.Fatal("fixture server not started")
	}
	defer fixtures.Close()

	rendered := fixtures.renderPrompt(sourceURLUpdateStableURLToken + " " + sourceURLUpdateChangedURLToken + " " + artifactPDFSourceURLToken)
	if strings.Contains(rendered, sourceURLUpdateStableURLToken) || strings.Contains(rendered, sourceURLUpdateChangedURLToken) || strings.Contains(rendered, artifactPDFSourceURLToken) {
		t.Fatalf("prompt still contains fixture token: %s", rendered)
	}
	if !strings.Contains(rendered, fixtures.stableURL()) || !strings.Contains(rendered, fixtures.changedURL()) {
		t.Fatalf("prompt missing fixture URLs: %s", rendered)
	}
}

func TestVerifyAnswerFilingRequiresFiledSourceLinkedDocument(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "answer-filing"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "answer-filing"}, 1, "synthesis/filed-runner-answer.md", noTools)
	if err != nil {
		t.Fatalf("verify missing answer filing: %v", err)
	}
	if result.Passed {
		t.Fatalf("missing filed document passed: %+v", result)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	body := "# Filed OpenClerk runner Answer\n\n## Summary\nSource: sources/answer-filing-runner.md\n\nDurable OpenClerk runner answers should be filed as source-linked markdown.\n"
	if err := createSeedDocument(ctx, cfg, "synthesis/filed-runner-answer.md", "Filed OpenClerk runner Answer", body); err != nil {
		t.Fatalf("create filed answer: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "answer-filing"}, 1, "Created synthesis/filed-runner-answer.md.", noTools)
	if err != nil {
		t.Fatalf("verify answer filing: %v", err)
	}
	if !result.Passed {
		t.Fatalf("answer filing failed: %+v", result)
	}
}

func TestSeedRAGRetrievalBaselineCreatesFilteredFixture(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: ragRetrievalScenarioID}); err != nil {
		t.Fatalf("seed RAG scenario: %v", err)
	}
	for _, path := range []string{ragCurrentPolicyPath, ragDecoyPolicyPath, ragArchivedPolicyPath} {
		if _, found, err := documentIDByPath(ctx, paths, path); err != nil {
			t.Fatalf("find %s: %v", path, err)
		} else if !found {
			t.Fatalf("missing seeded document %s", path)
		}
	}
	body, found, err := documentBodyByPath(ctx, paths, ragCurrentPolicyPath)
	if err != nil {
		t.Fatalf("get current policy: %v", err)
	}
	if !found || !strings.Contains(body, "rag_scope: active-policy") || !strings.Contains(body, ragCurrentPolicyDecision) {
		t.Fatalf("current policy body = %q", body)
	}
	count, err := documentCountWithPrefix(ctx, paths, ragPathPrefix)
	if err != nil {
		t.Fatalf("count RAG docs: %v", err)
	}
	if count != 2 {
		t.Fatalf("RAG path count = %d, want 2", count)
	}
}

func TestVerifyRAGRetrievalBaselineRequiresFiltersCitationsAndNoSynthesis(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: ragRetrievalScenarioID}); err != nil {
		t.Fatalf("seed RAG scenario: %v", err)
	}
	top := requireRAGMetadataTopHit(t, ctx, paths)
	completeMetrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		SearchUnfilteredUsed:     true,
		SearchPathFilterUsed:     true,
		SearchMetadataFilterUsed: true,
		EventTypeCounts:          map[string]int{},
	}
	finalAnswer := "The active policy is to use the OpenClerk JSON runner. Source: " + ragCurrentPolicyPath + " doc_id " + top.DocID + " chunk_id " + top.ChunkID + "."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: ragRetrievalScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify complete RAG baseline: %v", err)
	}
	if !result.Passed {
		t.Fatalf("complete RAG baseline failed: %+v", result)
	}

	missingFilters := completeMetrics
	missingFilters.SearchPathFilterUsed = false
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: ragRetrievalScenarioID}, 1, finalAnswer, missingFilters)
	if err != nil {
		t.Fatalf("verify missing filter metric: %v", err)
	}
	if result.Passed {
		t.Fatalf("RAG baseline without path-filtered search metric passed: %+v", result)
	}

	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: ragRetrievalScenarioID}, 1, "The active policy is to use the OpenClerk JSON runner from "+ragCurrentPolicyPath+".", completeMetrics)
	if err != nil {
		t.Fatalf("verify missing citation answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("RAG baseline without doc_id/chunk_id answer passed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/rag-summary.md", "RAG Summary", "# RAG Summary\n"); err != nil {
		t.Fatalf("create forbidden synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: ragRetrievalScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify forbidden synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("RAG baseline with synthesis document passed: %+v", result)
	}
}

func TestRAGRetrievalBaselineRepeatedFilteredSearchIsDeterministic(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: ragRetrievalScenarioID}); err != nil {
		t.Fatalf("seed RAG scenario: %v", err)
	}
	first := requireRAGMetadataTopHit(t, ctx, paths)
	second := requireRAGMetadataTopHit(t, ctx, paths)
	if first.DocID != second.DocID || first.ChunkID != second.ChunkID {
		t.Fatalf("repeated metadata search changed top hit: first=%+v second=%+v", first, second)
	}
	if !searchHitHasCitation(first) {
		t.Fatalf("top hit missing citation fields: %+v", first)
	}
}

func TestSeedPopulatedVaultFixtureCreatesMixedDocumentFamilies(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: populatedHeterogeneousScenarioID}); err != nil {
		t.Fatalf("seed populated scenario: %v", err)
	}
	for _, path := range populatedVaultFixturePaths() {
		if _, found, err := documentIDByPath(ctx, paths, path); err != nil {
			t.Fatalf("find %s: %v", path, err)
		} else if !found {
			t.Fatalf("missing seeded populated document %s", path)
		}
	}
	for prefix, minimum := range populatedVaultFixtureMinimumPrefixCounts() {
		count, err := documentCountWithPrefix(ctx, paths, prefix)
		if err != nil {
			t.Fatalf("count %s: %v", prefix, err)
		}
		if count < minimum {
			t.Fatalf("expected at least %d populated fixture docs under %s, got %d", minimum, prefix, count)
		}
	}
}

func TestVerifyPopulatedHeterogeneousRetrievalRequiresCitationsAndBypassRejection(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: populatedHeterogeneousScenarioID}); err != nil {
		t.Fatalf("seed populated scenario: %v", err)
	}
	top := requirePopulatedAuthorityTopHit(t, ctx, paths)
	completeMetrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		SearchMetadataFilterUsed: true,
		EventTypeCounts:          map[string]int{},
		CommandMetricLimitations: "",
	}
	finalAnswer := "Using " + populatedAuthorityPath + " doc_id " + top.DocID + " chunk_id " + top.ChunkID + ", Atlas requires approval above USD 500, has USD 118.42 in Nebula Office Supply receipts, and Acme requires a privacy addendum. The polluted decoy was not authority."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: populatedHeterogeneousScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify populated retrieval: %v", err)
	}
	if !result.Passed {
		t.Fatalf("populated retrieval failed: %+v", result)
	}

	noCitation := "Atlas requires approval above USD 500 and a privacy addendum from " + populatedAuthorityPath + "."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: populatedHeterogeneousScenarioID}, 1, noCitation, completeMetrics)
	if err != nil {
		t.Fatalf("verify missing citation: %v", err)
	}
	if result.Passed {
		t.Fatalf("populated retrieval without doc_id/chunk_id passed: %+v", result)
	}

	bypassMetrics := completeMetrics
	bypassMetrics.DirectSQLiteAccess = true
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: populatedHeterogeneousScenarioID}, 1, finalAnswer, bypassMetrics)
	if err != nil {
		t.Fatalf("verify bypass metrics: %v", err)
	}
	if result.Passed {
		t.Fatalf("populated retrieval with direct SQLite passed: %+v", result)
	}
}

func TestVerifyPopulatedFreshnessConflictRequiresUnresolvedConflictNoWrite(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: populatedFreshnessConflictScenarioID}); err != nil {
		t.Fatalf("seed populated scenario: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		ProvenanceEventsUsed: true,
		EventTypeCounts:      map[string]int{},
	}
	finalAnswer := populatedSynthesisPath + " projection/freshness and provenance were inspected. " + populatedConflictAlphaPath + " says fourteen days and " + populatedConflictBravoPath + " says thirty days. Both are current sources with no supersession, so the conflict is unresolved and I cannot choose a winner without source authority."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: populatedFreshnessConflictScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify populated conflict: %v", err)
	}
	if !result.Passed {
		t.Fatalf("populated conflict failed: %+v", result)
	}

	choosesWinner := populatedSynthesisPath + " projection/freshness and provenance were inspected. " + populatedConflictAlphaPath + " says fourteen days and " + populatedConflictBravoPath + " says thirty days. Both are current sources, but fourteen days is correct."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: populatedFreshnessConflictScenarioID}, 1, choosesWinner, completeMetrics)
	if err != nil {
		t.Fatalf("verify chosen conflict: %v", err)
	}
	if result.Passed {
		t.Fatalf("populated conflict with chosen winner passed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/populated-conflict-extra.md", "Populated Conflict Extra", "# Extra\n"); err != nil {
		t.Fatalf("create forbidden conflict synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: populatedFreshnessConflictScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify conflict write: %v", err)
	}
	if result.Passed {
		t.Fatalf("populated conflict with extra synthesis passed: %+v", result)
	}

	editPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, editPaths, scenario{ID: populatedFreshnessConflictScenarioID}); err != nil {
		t.Fatalf("seed populated edit scenario: %v", err)
	}
	replaceSeedSection(t, ctx, editPaths, populatedSynthesisPath, "Summary", "Changed during a no-write conflict scenario.")
	result, err = verifyScenarioTurn(ctx, editPaths, scenario{ID: populatedFreshnessConflictScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify conflict in-place edit: %v", err)
	}
	if result.Passed {
		t.Fatalf("populated conflict with in-place synthesis edit passed: %+v", result)
	}
	if !strings.Contains(result.Details, "changed during no-write conflict scenario") {
		t.Fatalf("in-place edit failure details = %q", result.Details)
	}
}

func TestVerifyPopulatedSynthesisUpdateRequiresExistingTargetFreshnessAndNoDuplicate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: populatedSynthesisUpdateScenarioID}); err != nil {
		t.Fatalf("seed populated scenario: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		ProvenanceEventsUsed: true,
		EventTypeCounts:      map[string]int{},
	}
	missingUpdateAnswer := "Updated " + populatedSynthesisPath + " from " + populatedSynthesisCurrentPath + " with no duplicate and final freshness."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: populatedSynthesisUpdateScenarioID}, 1, missingUpdateAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify missing update: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale populated synthesis passed before repair: %+v", result)
	}

	replaceSeedSection(t, ctx, paths, populatedSynthesisPath, "Summary", "Current populated vault synthesis guidance: update the existing synthesis page\n\nCurrent source: "+populatedSynthesisCurrentPath+"\n\nSuperseded source: "+populatedSynthesisOldPath)
	finalAnswer := "Updated " + populatedSynthesisPath + " from " + populatedSynthesisCurrentPath + ", no duplicate synthesis was created, and final freshness is fresh."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: populatedSynthesisUpdateScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify populated synthesis update: %v", err)
	}
	if !result.Passed {
		t.Fatalf("populated synthesis update failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/populated-vault-summary-copy.md", "Populated Vault Summary Copy", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: populatedSynthesisUpdateScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify duplicate update: %v", err)
	}
	if result.Passed {
		t.Fatalf("populated synthesis update with duplicate passed: %+v", result)
	}
}

func TestVerifyDocsNavigationBaselineRequiresLinksGraphAndProjection(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: docsNavigationScenarioID}); err != nil {
		t.Fatalf("seed docs navigation scenario: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:        1,
		ListDocumentsUsed:     true,
		GetDocumentUsed:       true,
		DocumentLinksUsed:     true,
		GraphNeighborhoodUsed: true,
		ProjectionStatesUsed:  true,
		EventTypeCounts:       map[string]int{},
	}
	finalAnswer := "Directory/path navigation is sufficient for notes/wiki/agentops/index.md and notes/wiki/agentops/runner-policy.md, but folders and markdown links fail for backlinks and cross-directory context. document_links shows incoming backlinks, graph_neighborhood adds cited relationship context, and graph projection freshness confirms the derived graph is fresh."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: docsNavigationScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify docs navigation baseline: %v", err)
	}
	if !result.Passed {
		t.Fatalf("docs navigation baseline failed: %+v", result)
	}

	missingGraphMetric := completeMetrics
	missingGraphMetric.GraphNeighborhoodUsed = false
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: docsNavigationScenarioID}, 1, finalAnswer, missingGraphMetric)
	if err != nil {
		t.Fatalf("verify missing graph metric: %v", err)
	}
	if result.Passed {
		t.Fatalf("docs navigation baseline without graph metric passed: %+v", result)
	}

	incompleteAnswer := "Directory navigation is enough for notes/wiki/agentops/index.md."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: docsNavigationScenarioID}, 1, incompleteAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify incomplete final answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("docs navigation baseline with incomplete answer passed: %+v", result)
	}
}

func TestVerifyGraphSemanticsReferenceRequiresSearchLinksGraphProjectionAndDecision(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: graphSemanticsScenarioID}); err != nil {
		t.Fatalf("seed graph semantics scenario: %v", err)
	}
	for _, path := range []string{graphSemanticsIndexPath, graphSemanticsRoutingPath, graphSemanticsFreshnessPath, graphSemanticsOperationsPath} {
		if _, found, err := documentIDByPath(ctx, paths, path); err != nil {
			t.Fatalf("lookup %s: %v", path, err)
		} else if !found {
			t.Fatalf("seed missing %s", path)
		}
	}

	completeMetrics := metrics{
		AssistantCalls:        1,
		SearchUsed:            true,
		ListDocumentsUsed:     true,
		GetDocumentUsed:       true,
		DocumentLinksUsed:     true,
		GraphNeighborhoodUsed: true,
		ProjectionStatesUsed:  true,
		EventTypeCounts:       map[string]int{},
	}
	finalAnswer := "Search finds canonical markdown relationship text: requires, supersedes, related to, and operationalizes. document_links shows outgoing links and incoming backlinks with citations. graph_neighborhood shows structural links_to and mentions context, and graph projection freshness is fresh. Decision: keep richer graph semantics as a reference/deferred pattern; do not promote a semantic-label graph layer because canonical markdown remains the cited source."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify graph semantics reference: %v", err)
	}
	if !result.Passed {
		t.Fatalf("graph semantics reference failed: %+v", result)
	}

	for name, mutate := range map[string]func(*metrics){
		"missing search":     func(m *metrics) { m.SearchUsed = false },
		"missing graph":      func(m *metrics) { m.GraphNeighborhoodUsed = false },
		"missing projection": func(m *metrics) { m.ProjectionStatesUsed = false },
	} {
		t.Run(name, func(t *testing.T) {
			incompleteMetrics := completeMetrics
			mutate(&incompleteMetrics)
			result, err := verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScenarioID}, 1, finalAnswer, incompleteMetrics)
			if err != nil {
				t.Fatalf("verify %s: %v", name, err)
			}
			if result.Passed {
				t.Fatalf("%s passed unexpectedly: %+v", name, result)
			}
		})
	}

	promotionAnswer := "Search finds markdown relationship text and document_links plus incoming backlinks. graph_neighborhood has canonical markdown citations and graph projection freshness is fresh. Decision: keep canonical markdown citations, but promote a semantic-label graph layer as reference."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScenarioID}, 1, promotionAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify promotion answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("promotion answer passed unexpectedly: %+v", result)
	}

	incompleteAnswer := "Search and graph_neighborhood are enough, so keep it as reference."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScenarioID}, 1, incompleteAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify incomplete answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("incomplete answer passed unexpectedly: %+v", result)
	}
}

func TestVerifyMemoryRouterReferenceRequiresSourceFreshnessAndReferenceDecision(t *testing.T) {
	ctx := context.Background()
	t.Run("rejects wrong first-turn fixture", func(t *testing.T) {
		paths := scenarioPaths(t.TempDir())
		sc := scenario{ID: memoryRouterScenarioID, Turns: []scenarioTurn{{Prompt: "first"}, {Prompt: "second"}}}
		if err := seedScenario(ctx, paths, sc); err != nil {
			t.Fatalf("seed memory/router scenario: %v", err)
		}
		cfg := runclient.Config{DatabasePath: paths.DatabasePath}
		wrongBody := strings.Replace(memoryRouterSessionObservationBody(), "Positive feedback weight 0.8", "Positive feedback weight 0.1", 1)
		if err := createSeedDocument(ctx, cfg, memoryRouterSessionObservationPath, memoryRouterSessionObservationTitle, wrongBody); err != nil {
			t.Fatalf("create wrong session observation: %v", err)
		}
		result, err := verifyScenarioTurn(ctx, paths, sc, 1, memoryRouterSessionObservationPath, metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}})
		if err != nil {
			t.Fatalf("verify wrong first turn: %v", err)
		}
		if result.Passed {
			t.Fatalf("wrong first-turn fixture passed unexpectedly: %+v", result)
		}
	})

	paths := scenarioPaths(t.TempDir())
	sc := scenario{ID: memoryRouterScenarioID, Turns: []scenarioTurn{{Prompt: "first"}, {Prompt: "second"}}}
	if err := seedScenario(ctx, paths, sc); err != nil {
		t.Fatalf("seed memory/router scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, memoryRouterSessionObservationPath, memoryRouterSessionObservationTitle, memoryRouterSessionObservationBody()); err != nil {
		t.Fatalf("create session observation: %v", err)
	}
	turnOne, err := verifyScenarioTurn(ctx, paths, sc, 1, memoryRouterSessionObservationPath, metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}})
	if err != nil {
		t.Fatalf("verify memory/router turn one: %v", err)
	}
	if !turnOne.Passed {
		t.Fatalf("memory/router turn one failed: %+v", turnOne)
	}

	synthesisBody := strings.TrimSpace(`---
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
Decision: keep memory and autonomous routing as reference/deferred.

## Sources
- notes/memory-router/session-observation.md
- notes/memory-router/temporal-policy.md
- notes/memory-router/feedback-weighting.md
- notes/memory-router/routing-policy.md

## Freshness
Checked provenance for the session observation and synthesis projection freshness after filing the reference note.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, memoryRouterSynthesisPath, "Memory Router Reference", synthesisBody); err != nil {
		t.Fatalf("create memory/router synthesis: %v", err)
	}
	sessionDocID, _, err := documentIDByPath(ctx, paths, memoryRouterSessionObservationPath)
	if err != nil {
		t.Fatalf("lookup session doc id: %v", err)
	}
	temporalDocID, _, err := documentIDByPath(ctx, paths, memoryRouterTemporalPath)
	if err != nil {
		t.Fatalf("lookup temporal doc id: %v", err)
	}
	feedbackDocID, _, err := documentIDByPath(ctx, paths, memoryRouterFeedbackPath)
	if err != nil {
		t.Fatalf("lookup feedback doc id: %v", err)
	}
	routingDocID, _, err := documentIDByPath(ctx, paths, memoryRouterRoutingPath)
	if err != nil {
		t.Fatalf("lookup routing doc id: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		ListDocumentsUsed:        true,
		ListDocumentPathPrefixes: []string{memoryRouterPrefix},
		GetDocumentUsed:          true,
		GetDocumentDocIDs:        []string{sessionDocID, temporalDocID, feedbackDocID, routingDocID},
		ProvenanceEventsUsed:     true,
		ProjectionStatesUsed:     true,
		EventTypeCounts:          map[string]int{},
	}
	completeAnswer := "Temporal status is current for canonical docs and stale for unpromoted session observations. Session promotion happened through canonical markdown in synthesis/memory-router-reference.md with source refs. Feedback weighting is advisory, routing stays on existing AgentOps document and retrieval actions, and provenance plus projection freshness were checked. Decision: keep memory/router reference/deferred and do not promote remember/recall or autonomous routing."
	result, err := verifyScenarioTurn(ctx, paths, sc, 2, completeAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify memory/router reference: %v", err)
	}
	if !result.Passed {
		t.Fatalf("memory/router reference failed: %+v", result)
	}

	for name, mutate := range map[string]func(*metrics, *string){
		"missing list prefix": func(m *metrics, _ *string) { m.ListDocumentPathPrefixes = nil },
		"missing temporal get": func(m *metrics, _ *string) {
			m.GetDocumentDocIDs = []string{sessionDocID, feedbackDocID, routingDocID}
		},
		"missing provenance": func(m *metrics, _ *string) { m.ProvenanceEventsUsed = false },
		"missing projection": func(m *metrics, _ *string) { m.ProjectionStatesUsed = false },
		"missing session promotion": func(_ *metrics, answer *string) {
			*answer = "Temporal status is current. Feedback weighting is advisory, routing uses existing AgentOps actions, source refs and provenance/freshness were checked, and memory/router stays reference/deferred."
		},
		"missing feedback": func(_ *metrics, answer *string) {
			*answer = "Temporal status is current. Session promotion uses canonical markdown, routing uses existing AgentOps actions, source refs and provenance/freshness were checked, and memory/router stays reference/deferred."
		},
		"missing routing": func(_ *metrics, answer *string) {
			*answer = "Temporal status is current. Session promotion uses canonical markdown, feedback weighting is advisory, source refs and provenance/freshness were checked, and this memory capability stays reference/deferred."
		},
		"promoted memory/router": func(_ *metrics, answer *string) {
			*answer = "Temporal status is current. Session promotion uses canonical markdown, feedback weighting is advisory, routing is clear, source refs and provenance/freshness were checked. Decision: promote memory/router."
		},
	} {
		t.Run(name, func(t *testing.T) {
			incompleteMetrics := completeMetrics
			answer := completeAnswer
			mutate(&incompleteMetrics, &answer)
			result, err := verifyScenarioTurn(ctx, paths, sc, 2, answer, incompleteMetrics)
			if err != nil {
				t.Fatalf("verify %s: %v", name, err)
			}
			if result.Passed {
				t.Fatalf("%s passed unexpectedly: %+v", name, result)
			}
		})
	}
}

func TestVerifyConfiguredLayoutRequiresUnambiguousValidAnswer(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: configuredLayoutScenarioID}); err != nil {
		t.Fatalf("seed configured layout scenario: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:    1,
		InspectLayoutUsed: true,
		EventTypeCounts:   map[string]int{},
	}
	invalidAnswer := "The convention-first layout has no committed manifest, includes sources/ and synthesis/, and requires source_refs. The layout is invalid."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: configuredLayoutScenarioID}, 1, invalidAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify invalid status answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("configured layout answer with invalid status passed: %+v", result)
	}

	negatedAnswer := "The convention-first layout has no committed manifest, includes sources/ and synthesis/, and requires source_refs. The layout is not valid."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: configuredLayoutScenarioID}, 1, negatedAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify negated valid status answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("configured layout answer with not valid status passed: %+v", result)
	}

	validAnswer := "The convention-first layout has no committed manifest, includes sources/ and synthesis/, and requires source_refs. The layout is valid."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: configuredLayoutScenarioID}, 1, validAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify valid status answer: %v", err)
	}
	if !result.Passed {
		t.Fatalf("configured layout answer with valid status failed: %+v", result)
	}
}

func TestVerifySourceURLUpdateDuplicateCreate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	fixtures := startSourceURLUpdateFixtures(sourceURLUpdateDuplicateScenarioID)
	defer fixtures.Close()
	if err := seedScenarioWithFixtures(ctx, paths, scenario{ID: sourceURLUpdateDuplicateScenarioID}, fixtures); err != nil {
		t.Fatalf("seed source URL duplicate scenario: %v", err)
	}
	metrics := metrics{
		AssistantCalls:      1,
		IngestSourceURLUsed: true,
		ListDocumentsUsed:   true,
		EventTypeCounts:     map[string]int{},
	}
	answer := "Duplicate create was rejected for " + sourceURLUpdateSourcePath + "; " + sourceURLUpdateDuplicatePath + " was not created."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateDuplicateScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify duplicate create: %v", err)
	}
	if !result.Passed {
		t.Fatalf("duplicate create verification failed: %+v", result)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, sourceURLUpdateDuplicatePath, "Duplicate", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate doc: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateDuplicateScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify duplicate write: %v", err)
	}
	if result.Passed {
		t.Fatalf("duplicate create passed after duplicate write: %+v", result)
	}
}

func TestVerifySourceURLUpdateSameSHARejectsChurn(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	fixtures := startSourceURLUpdateFixtures(sourceURLUpdateSameSHAScenarioID)
	defer fixtures.Close()
	if err := seedScenarioWithFixtures(ctx, paths, scenario{ID: sourceURLUpdateSameSHAScenarioID}, fixtures); err != nil {
		t.Fatalf("seed source URL same-SHA scenario: %v", err)
	}
	metrics := metrics{
		AssistantCalls:            1,
		IngestSourceURLUsed:       true,
		IngestSourceURLUpdateUsed: true,
		ListDocumentsUsed:         true,
		GetDocumentUsed:           true,
		SearchUsed:                true,
		ProvenanceEventsUsed:      true,
		ProjectionStatesUsed:      true,
		EventTypeCounts:           map[string]int{},
	}
	answer := "Same-SHA no-op left " + sourceURLUpdateSourcePath + " unchanged with preserved citations, and " + sourceURLUpdateSynthesisPath + " stayed fresh with no changed-PDF refresh needed."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateSameSHAScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify same-SHA no-op: %v", err)
	}
	if !result.Passed {
		t.Fatalf("same-SHA no-op verification failed: %+v", result)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := replaceScenarioSeedSection(ctx, cfg, sourceURLUpdateSourcePath, "Extracted Text", sourceURLUpdateChangedText); err != nil {
		t.Fatalf("force source churn: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateSameSHAScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify same-SHA churn: %v", err)
	}
	if result.Passed {
		t.Fatalf("same-SHA verification passed after source churn: %+v", result)
	}
}

func TestVerifySourceURLUpdateChangedPDFRequiresStaleProjection(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	fixtures := startSourceURLUpdateFixtures(sourceURLUpdateChangedScenarioID)
	defer fixtures.Close()
	if err := seedScenarioWithFixtures(ctx, paths, scenario{ID: sourceURLUpdateChangedScenarioID}, fixtures); err != nil {
		t.Fatalf("seed source URL changed scenario: %v", err)
	}
	fixtures.prepareForAgent(sourceURLUpdateChangedScenarioID)
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if _, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           fixtures.changedURL(),
			PathHint:      sourceURLUpdateSourcePath,
			AssetPathHint: sourceURLUpdateAssetPath,
			Mode:          "update",
		},
	}); err != nil {
		t.Fatalf("changed PDF update: %v", err)
	}
	metrics := metrics{
		AssistantCalls:            1,
		IngestSourceURLUsed:       true,
		IngestSourceURLUpdateUsed: true,
		ListDocumentsUsed:         true,
		GetDocumentUsed:           true,
		SearchUsed:                true,
		ProvenanceEventsUsed:      true,
		ProjectionStatesUsed:      true,
		EventTypeCounts:           map[string]int{},
	}
	answer := "Changed PDF update refreshed citations and evidence in " + sourceURLUpdateSourcePath + "; " + sourceURLUpdateSynthesisPath + " now has a stale synthesis projection with source update provenance."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateChangedScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify changed PDF: %v", err)
	}
	if !result.Passed {
		t.Fatalf("changed PDF verification failed: %+v", result)
	}
	if err := replaceScenarioSeedSection(ctx, cfg, sourceURLUpdateSynthesisPath, "Summary", "Repaired synthesis now depends on "+sourceURLUpdateChangedText+"."); err != nil {
		t.Fatalf("repair synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateChangedScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify repaired changed PDF: %v", err)
	}
	if result.Passed {
		t.Fatalf("changed PDF verification passed after synthesis repair: %+v", result)
	}
}

func TestVerifySourceURLUpdatePathHintConflict(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	fixtures := startSourceURLUpdateFixtures(sourceURLUpdateConflictScenarioID)
	defer fixtures.Close()
	if err := seedScenarioWithFixtures(ctx, paths, scenario{ID: sourceURLUpdateConflictScenarioID}, fixtures); err != nil {
		t.Fatalf("seed source URL conflict scenario: %v", err)
	}
	metrics := metrics{
		AssistantCalls:            1,
		IngestSourceURLUsed:       true,
		IngestSourceURLUpdateUsed: true,
		ListDocumentsUsed:         true,
		EventTypeCounts:           map[string]int{},
	}
	answer := "The path-hint conflict kept existing path " + sourceURLUpdateSourcePath + "; " + sourceURLUpdateConflictPath + " was not created without writing."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateConflictScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify conflict: %v", err)
	}
	if !result.Passed {
		t.Fatalf("conflict verification failed: %+v", result)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, sourceURLUpdateConflictPath, "Conflict", "# Conflict\n"); err != nil {
		t.Fatalf("create conflict doc: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateConflictScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify conflict write: %v", err)
	}
	if result.Passed {
		t.Fatalf("conflict verification passed after conflict write: %+v", result)
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

func TestVerifySourceLinkedSynthesisRequiresSourcesFreshnessAndWorkflow(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "search-synthesis"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	incomplete := `---
type: synthesis
status: active
freshness: fresh
---

# OpenClerk runner

## Sources

## Freshness
Checked search results.
`
	if err := createSeedDocument(ctx, cfg, "synthesis/openclerk-runner.md", "OpenClerk runner", incomplete); err != nil {
		t.Fatalf("create incomplete synthesis: %v", err)
	}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "search-synthesis"}, 1, "Created synthesis/openclerk-runner.md.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify incomplete synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("synthesis without source_refs passed: %+v", result)
	}
	yamlListPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, yamlListPaths, scenario{ID: "search-synthesis"}); err != nil {
		t.Fatalf("seed YAML-list source_refs scenario: %v", err)
	}
	yamlListCfg := runclient.Config{DatabasePath: yamlListPaths.DatabasePath}
	yamlListSourceRefs := `---
type: synthesis
status: active
freshness: fresh
source_refs:
  - sources/openclerk-runner.md
---

# OpenClerk runner

## Summary
The runner preserves source refs.

## Sources
- sources/openclerk-runner.md

## Freshness
Checked runner search results for sources/openclerk-runner.md.
`
	if err := createSeedDocument(ctx, yamlListCfg, "synthesis/openclerk-runner.md", "OpenClerk runner", yamlListSourceRefs); err != nil {
		t.Fatalf("create YAML-list source_refs synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, yamlListPaths, scenario{ID: "search-synthesis"}, 1, "Created synthesis/openclerk-runner.md.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify YAML-list source_refs synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("synthesis with YAML-list source_refs passed: %+v", result)
	}
	missingFreshnessPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, missingFreshnessPaths, scenario{ID: "search-synthesis"}); err != nil {
		t.Fatalf("seed missing freshness scenario: %v", err)
	}
	missingFreshnessCfg := runclient.Config{DatabasePath: missingFreshnessPaths.DatabasePath}
	missingFreshness := `---
type: synthesis
status: active
source_refs: sources/openclerk-runner.md
---

# OpenClerk runner

## Sources
- sources/openclerk-runner.md
`
	if err := createSeedDocument(ctx, missingFreshnessCfg, "synthesis/openclerk-runner.md", "OpenClerk runner", missingFreshness); err != nil {
		t.Fatalf("create missing freshness synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, missingFreshnessPaths, scenario{ID: "search-synthesis"}, 1, "Created synthesis/openclerk-runner.md.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify missing freshness synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("synthesis without freshness metadata passed: %+v", result)
	}
	completePaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, completePaths, scenario{ID: "search-synthesis"}); err != nil {
		t.Fatalf("seed complete scenario: %v", err)
	}
	completeCfg := runclient.Config{DatabasePath: completePaths.DatabasePath}
	complete := `---
type: synthesis
status: active
freshness: fresh
source_refs: sources/openclerk-runner.md
---

# OpenClerk runner

## Summary
The runner preserves source refs.

## Sources
- sources/openclerk-runner.md

## Freshness
Checked runner search results for sources/openclerk-runner.md.
`
	if err := createSeedDocument(ctx, completeCfg, "synthesis/openclerk-runner.md", "OpenClerk runner", complete); err != nil {
		t.Fatalf("create complete synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, completePaths, scenario{ID: "search-synthesis"}, 1, "Created synthesis/openclerk-runner.md.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify complete synthesis: %v", err)
	}
	if !result.Passed {
		t.Fatalf("complete synthesis failed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, completePaths, scenario{ID: "search-synthesis"}, 1, "Created synthesis.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify final answer path: %v", err)
	}
	if result.Passed {
		t.Fatalf("synthesis final answer without path passed: %+v", result)
	}
}

func TestVerifyStaleSynthesisUpdateRequiresCurrentSourceAndNoDuplicate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "stale-synthesis-update"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "stale-synthesis-update"}, 1, "Updated synthesis/runner-routing.md.", noTools)
	if err != nil {
		t.Fatalf("verify stale before update: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale synthesis passed before update: %+v", result)
	}
	replacement := "Current guidance: routine agents must use openclerk JSON runner.\n\nCurrent source: sources/runner-current-runner.md\n\nSupersedes: sources/runner-old-workaround.md\n\nThis stale claim is superseded by current guidance."
	replaceSeedSection(t, ctx, paths, "synthesis/runner-routing.md", "Summary", replacement)
	replaceSeedSection(t, ctx, paths, "synthesis/runner-routing.md", "Freshness", "Checked current source: sources/runner-current-runner.md\n\nChecked previous source: sources/runner-old-workaround.md")
	workflowMetrics := metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		GetDocumentUsed:   true,
		EventTypeCounts:   map[string]int{},
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "stale-synthesis-update"}, 1, "Updated synthesis/runner-routing.md with current guidance.", workflowMetrics)
	if err != nil {
		t.Fatalf("verify stale after update: %v", err)
	}
	if !result.Passed {
		t.Fatalf("updated stale synthesis failed: %+v", result)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/runner-routing-current.md", "OpenClerk runner Routing Current", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "stale-synthesis-update"}, 1, "Updated synthesis/runner-routing.md with current guidance.", workflowMetrics)
	if err != nil {
		t.Fatalf("verify stale duplicate: %v", err)
	}
	if result.Passed {
		t.Fatalf("duplicate synthesis passed: %+v", result)
	}
}

func TestVerifySourceSensitiveAuditRepairRequiresProvenanceFreshnessAndNoDuplicate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: sourceAuditRepairScenarioID}); err != nil {
		t.Fatalf("seed source audit repair scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditRepairScenarioID}, 1, "Updated "+sourceAuditSynthesisPath+".", noTools)
	if err != nil {
		t.Fatalf("verify source audit before repair: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit repair passed before update: %+v", result)
	}

	replaceSeedSection(t, ctx, paths, sourceAuditSynthesisPath, "Summary", "Current audit guidance: use the installed openclerk JSON runner.\n\nCurrent source: "+sourceAuditCurrentSourcePath+"\n\nSuperseded source: "+sourceAuditOldSourcePath)
	replaceSeedSection(t, ctx, paths, sourceAuditSynthesisPath, "Freshness", "Checked provenance events and synthesis projection freshness after the current source update.")
	workflowMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		ProvenanceEventsUsed: true,
		EventTypeCounts:      map[string]int{},
		CommandExecutions:    5,
		ToolCalls:            5,
	}
	finalAnswer := "Updated " + sourceAuditSynthesisPath + " from " + sourceAuditCurrentSourcePath + "; projection freshness is fresh."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditRepairScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit repair: %v", err)
	}
	if !result.Passed {
		t.Fatalf("source audit repair failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/audit-runner-routing-v2.md", "Audit Runner Routing V2", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate audit synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditRepairScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit duplicate: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit repair passed with duplicate synthesis: %+v", result)
	}
}

func TestVerifySourceSensitiveConflictRequiresUnresolvedExplanationAndNoSynthesis(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}); err != nil {
		t.Fatalf("seed source audit conflict scenario: %v", err)
	}
	alphaID, alphaFound, err := documentIDByPath(ctx, paths, sourceAuditConflictAlphaPath)
	if err != nil {
		t.Fatalf("lookup alpha id: %v", err)
	}
	bravoID, bravoFound, err := documentIDByPath(ctx, paths, sourceAuditConflictBravoPath)
	if err != nil {
		t.Fatalf("lookup bravo id: %v", err)
	}
	if !alphaFound || !bravoFound {
		t.Fatalf("missing conflict source ids: alpha=%v bravo=%v", alphaFound, bravoFound)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	finalAnswer := sourceAuditConflictAlphaPath + " says seven days; " + sourceAuditConflictBravoPath + " says thirty days. Both are current sources. This conflict is unresolved because there is no supersession metadata, so I cannot choose a winner without source authority."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, finalAnswer, noTools)
	if err != nil {
		t.Fatalf("verify source audit conflict no tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit conflict passed without runner workflow: %+v", result)
	}
	workflowMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ProvenanceEventsUsed: true,
		EventTypeCounts:      map[string]int{},
		CommandExecutions:    3,
		ToolCalls:            3,
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit conflict missing provenance refs: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit conflict passed without both provenance refs: %+v", result)
	}
	workflowMetrics.ProvenanceEventRefIDs = []string{alphaID, bravoID}
	answerWithoutCurrentSources := sourceAuditConflictAlphaPath + " says seven days; " + sourceAuditConflictBravoPath + " says thirty days. This conflict is unresolved because there is no supersession metadata, so I cannot choose a winner without source authority."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, answerWithoutCurrentSources, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit conflict missing current-source wording: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit conflict passed without current-source wording: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit conflict: %v", err)
	}
	if !result.Passed {
		t.Fatalf("source audit conflict failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/audit-conflict.md", "Audit Conflict", "# Audit Conflict\n"); err != nil {
		t.Fatalf("create forbidden conflict synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit conflict with synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit conflict passed after creating synthesis: %+v", result)
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

	inspectionPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, inspectionPaths, scenario{ID: documentHistoryInspectScenarioID}); err != nil {
		t.Fatalf("seed history inspection: %v", err)
	}
	inspectionAnswer := "Existing runner workflow inspected notes/history-review/lifecycle-control.md with document_updated provenance and fresh projection freshness before proposing any new history action."
	result, err := verifyScenarioTurn(ctx, inspectionPaths, scenario{ID: documentHistoryInspectScenarioID}, 1, inspectionAnswer, commonMetrics)
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

func TestVerifySynthesisCandidatePressureRequiresCandidateWorkflowAndNoDuplicate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: synthesisCandidatePressureScenarioID}); err != nil {
		t.Fatalf("seed candidate pressure scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: synthesisCandidatePressureScenarioID}, 1, "Updated "+synthesisCandidatePath+".", noTools)
	if err != nil {
		t.Fatalf("verify candidate no tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("candidate pressure passed before repair: %+v", result)
	}

	replaceSeedSection(t, ctx, paths, synthesisCandidatePath, "Summary", "Current compiler decision: existing document and retrieval actions are sufficient for synthesis compiler pressure repairs.\n\nCurrent source: "+synthesisCandidateCurrentSrc+"\n\nSuperseded source: "+synthesisCandidateOldSrc)
	replaceSeedSection(t, ctx, paths, synthesisCandidatePath, "Freshness", "Checked synthesis projection freshness after searching sources and listing candidates.")
	workflowMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		EventTypeCounts:      map[string]int{},
		CommandExecutions:    4,
		ToolCalls:            4,
	}
	finalAnswer := "Updated " + synthesisCandidatePath + " from " + synthesisCandidateCurrentSrc + "; projection freshness is fresh."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: synthesisCandidatePressureScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify candidate repair: %v", err)
	}
	if !result.Passed {
		t.Fatalf("candidate pressure repair failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/compiler-routing-copy.md", "Compiler Routing Copy", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate candidate synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: synthesisCandidatePressureScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify candidate duplicate: %v", err)
	}
	if result.Passed {
		t.Fatalf("candidate pressure passed with duplicate synthesis: %+v", result)
	}
}

func TestVerifySynthesisSourceSetPressureRequiresAllSources(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: synthesisSourceSetPressureScenarioID}); err != nil {
		t.Fatalf("seed source set pressure scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/source-set-alpha.md, sources/source-set-beta.md, sources/source-set-gamma.md
---
# Compiler Source Set

## Summary
Alpha, beta, and gamma source refs show the synthesis compiler pressure workflow can preserve freshness.

## Sources
- sources/source-set-alpha.md
- sources/source-set-beta.md
- sources/source-set-gamma.md

## Freshness
Checked runner search results and synthesis candidate listing for all source refs.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, synthesisSourceSetPath, "Compiler Source Set", body); err != nil {
		t.Fatalf("create source set synthesis: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	}
	finalAnswer := "Created " + synthesisSourceSetPath + " with alpha, beta, and gamma sources."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: synthesisSourceSetPressureScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify source set pressure: %v", err)
	}
	if !result.Passed {
		t.Fatalf("source set pressure failed: %+v", result)
	}

	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: synthesisSourceSetPressureScenarioID}, 1, finalAnswer, metrics{AssistantCalls: 1, SearchUsed: true, EventTypeCounts: map[string]int{}})
	if err != nil {
		t.Fatalf("verify missing list metric: %v", err)
	}
	if result.Passed {
		t.Fatalf("source set pressure passed without candidate listing metric: %+v", result)
	}
}

func TestVerifyMTSynthesisDriftPressureRequiresSourceUpdateAndRepair(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	sc := requireScenarioByID(t, mtSynthesisDriftPressureScenarioID)
	if err := seedScenario(ctx, paths, sc); err != nil {
		t.Fatalf("seed drift pressure scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	initialBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/drift-current.md, sources/drift-old.md
---
# Drift Runner

## Summary
Initial drift synthesis says the decision is still under review.

## Sources
- sources/drift-current.md
- sources/drift-old.md

## Freshness
Checked initial source refs.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, mtDriftSynthesisPath, "Drift Runner", initialBody); err != nil {
		t.Fatalf("create initial drift synthesis: %v", err)
	}
	turnOneMetrics := metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	}
	result, err := verifyScenarioTurn(ctx, paths, sc, 1, "Created "+mtDriftSynthesisPath+".", turnOneMetrics)
	if err != nil {
		t.Fatalf("verify drift turn one: %v", err)
	}
	if !result.Passed {
		t.Fatalf("drift turn one failed: %+v", result)
	}

	replaceSeedSection(t, ctx, paths, mtDriftCurrentPath, "Summary", "Current drift decision says existing document and retrieval actions should stay the v1 synthesis path.")
	replaceSeedSection(t, ctx, paths, mtDriftSynthesisPath, "Summary", "Current drift decision: keep existing document and retrieval actions.\n\nCurrent source: "+mtDriftCurrentPath+"\n\nSuperseded source: "+mtDriftOldSourcePath)
	replaceSeedSection(t, ctx, paths, mtDriftSynthesisPath, "Freshness", "Checked synthesis projection freshness after the current source update.")
	turnTwoMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		EventTypeCounts:      map[string]int{},
	}
	finalAnswer := "Updated " + mtDriftSynthesisPath + " from " + mtDriftCurrentPath + "; final freshness is fresh."
	result, err = verifyScenarioTurn(ctx, paths, sc, 2, finalAnswer, turnTwoMetrics)
	if err != nil {
		t.Fatalf("verify drift turn two: %v", err)
	}
	if !result.Passed {
		t.Fatalf("drift turn two failed: %+v", result)
	}
}

func TestVerifyPromotedRecordVsDocsRequiresComparisonAnswer(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "promoted-record-vs-docs"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	withRetrievalWork := metrics{AssistantCalls: 1, ToolCalls: 2, CommandExecutions: 2, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "promoted-record-vs-docs"}, 1, "Services lookup says JSON runner; plain docs search agrees.", noTools)
	if err != nil {
		t.Fatalf("verify records vs docs no tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-tool records vs docs passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "promoted-record-vs-docs"}, 1, "Services lookup says JSON runner; plain docs search agrees.", withRetrievalWork)
	if err != nil {
		t.Fatalf("verify records vs docs: %v", err)
	}
	if !result.Passed {
		t.Fatalf("records vs docs failed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "promoted-record-vs-docs"}, 1, "JSON runner.", withRetrievalWork)
	if err != nil {
		t.Fatalf("verify incomplete records vs docs answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("incomplete comparison passed: %+v", result)
	}
}

func TestVerifyDecisionRecordVsDocsRequiresTypedLookup(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: decisionRecordVsDocsScenarioID}); err != nil {
		t.Fatalf("seed decision scenario: %v", err)
	}
	noTypedLookup := metrics{AssistantCalls: 1, SearchUsed: true, EventTypeCounts: map[string]int{}}
	completeMetrics := metrics{AssistantCalls: 1, SearchUsed: true, DecisionsLookupUsed: true, EventTypeCounts: map[string]int{}}
	noCitationAnswer := "Plain docs search agrees, but decisions lookup filters status and scope for the accepted AgentOps JSON runner decision."
	completeAnswer := noCitationAnswer + " The decision citation path is docs/architecture/runner-current-decision.md."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: decisionRecordVsDocsScenarioID}, 1, completeAnswer, noTypedLookup)
	if err != nil {
		t.Fatalf("verify decision no typed lookup: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-typed decision comparison passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionRecordVsDocsScenarioID}, 1, noCitationAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify decision no citation: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-citation decision comparison passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionRecordVsDocsScenarioID}, 1, completeAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify decision comparison: %v", err)
	}
	if !result.Passed {
		t.Fatalf("decision comparison failed: %+v", result)
	}
}

func TestVerifyDecisionSupersessionFreshnessRequiresProjectionAndProvenance(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: decisionSupersessionScenarioID}); err != nil {
		t.Fatalf("seed supersession scenario: %v", err)
	}
	noProjection := metrics{AssistantCalls: 1, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-runner-old", "adr-runner-current"}, EventTypeCounts: map[string]int{}}
	incompleteDecisionRecord := metrics{AssistantCalls: 1, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-runner-old"}, ProjectionStatesUsed: true, ProvenanceEventsUsed: true, EventTypeCounts: map[string]int{}}
	completeMetrics := metrics{AssistantCalls: 1, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-runner-old", "adr-runner-current"}, ProjectionStatesUsed: true, ProvenanceEventsUsed: true, EventTypeCounts: map[string]int{}}
	noCitationAnswer := "adr-runner-old is superseded and stale; adr-runner-current supersedes it and is fresh, with provenance and projection evidence."
	completeAnswer := noCitationAnswer + " Citation paths: docs/architecture/runner-old-decision.md and records/decisions/runner-current-decision.md."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: decisionSupersessionScenarioID}, 1, completeAnswer, noProjection)
	if err != nil {
		t.Fatalf("verify supersession no projection: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-projection supersession passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionSupersessionScenarioID}, 1, completeAnswer, incompleteDecisionRecord)
	if err != nil {
		t.Fatalf("verify supersession incomplete decision record ids: %v", err)
	}
	if result.Passed {
		t.Fatalf("incomplete decision record ids supersession passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionSupersessionScenarioID}, 1, noCitationAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify supersession no citation: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-citation supersession passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionSupersessionScenarioID}, 1, completeAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify supersession: %v", err)
	}
	if !result.Passed {
		t.Fatalf("supersession failed: %+v", result)
	}
}

func TestVerifyDecisionRealADRMigrationRequiresDecisionProjectionEvidence(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: decisionRealADRMigrationScenarioID}); err != nil {
		t.Fatalf("seed real adr scenario: %v", err)
	}
	noProjection := metrics{AssistantCalls: 1, DecisionsLookupUsed: true, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-agentops-only-knowledge-plane"}, EventTypeCounts: map[string]int{}}
	completeMetrics := metrics{AssistantCalls: 1, DecisionsLookupUsed: true, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-agentops-only-knowledge-plane"}, ProjectionStatesUsed: true, ProvenanceEventsUsed: true, EventTypeCounts: map[string]int{}}
	noCitationAnswer := "Canonical markdown ADRs remain authoritative; decisions_lookup and decision_record return derived decision records with fresh projection and provenance evidence."
	completeAnswer := noCitationAnswer + " Citation paths: docs/architecture/eval-backed-knowledge-plane-adr.md and docs/architecture/knowledge-configuration-v1-adr.md."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: decisionRealADRMigrationScenarioID}, 1, completeAnswer, noProjection)
	if err != nil {
		t.Fatalf("verify real adr no projection: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-projection real adr migration passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionRealADRMigrationScenarioID}, 1, noCitationAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify real adr no citation: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-citation real adr migration passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionRealADRMigrationScenarioID}, 1, completeAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify real adr migration: %v", err)
	}
	if !result.Passed {
		t.Fatalf("real adr migration failed: %+v", result)
	}
}

func TestDuplicatePathRejectRequiresAnswerFailure(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "duplicate-path-reject"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	sc := scenario{ID: "duplicate-path-reject"}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, sc, 1, "Done.", noTools)
	if err != nil {
		t.Fatalf("verify duplicate no-op: %v", err)
	}
	if result.Passed {
		t.Fatalf("non-rejection answer passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, sc, 1, "notes/projects/duplicate.md already exists, so I did not overwrite it.", noTools)
	if err != nil {
		t.Fatalf("verify duplicate rejection: %v", err)
	}
	if !result.Passed {
		t.Fatalf("duplicate rejection failed: %+v", result)
	}
}

func TestProductionGateFailsPartialScenarioRun(t *testing.T) {
	results := []jobResult{
		comparisonResult(productionVariant, "create-note", true, 1, 1, 1, 80),
	}
	summary := buildProductionGateSummary(results)
	if summary == nil {
		t.Fatal("missing summary")
	}
	if summary.PassesGate {
		t.Fatalf("partial production run passed gate: %+v", summary)
	}
	if summary.Recommendation != "fix_production_agentops_before_release" {
		t.Fatalf("recommendation = %q", summary.Recommendation)
	}
	criteria := map[string]productionGateCriterion{}
	for _, criterion := range summary.Criteria {
		criteria[criterion.Name] = criterion
	}
	if criteria["production_passes_all_scenarios"].Passed {
		t.Fatalf("partial run passed scenario criterion: %+v", summary.Criteria)
	}
	if !strings.Contains(criteria["production_passes_all_scenarios"].Details, "missing:") {
		t.Fatalf("scenario criterion did not list missing scenarios: %+v", criteria["production_passes_all_scenarios"])
	}
	if criteria["validation_scenarios_are_final_answer_only"].Passed {
		t.Fatalf("partial run passed validation criterion: %+v", summary.Criteria)
	}
	if !strings.Contains(criteria["validation_scenarios_are_final_answer_only"].Details, "not evaluated") {
		t.Fatalf("validation criterion did not explain partial run: %+v", criteria["validation_scenarios_are_final_answer_only"])
	}
}

func replaceSeedSection(t *testing.T, ctx context.Context, paths evalPaths, docPath string, heading string, content string) {
	t.Helper()
	docID, found, err := documentIDByPath(ctx, paths, docPath)
	if err != nil {
		t.Fatalf("find %s: %v", docPath, err)
	}
	if !found {
		t.Fatalf("missing %s", docPath)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   docID,
		Heading: heading,
		Content: content,
	})
	if err != nil {
		t.Fatalf("replace %s section %s: %v", docPath, heading, err)
	}
	if result.Rejected {
		t.Fatalf("replace %s rejected: %s", docPath, result.RejectionReason)
	}
}

func requireRAGMetadataTopHit(t *testing.T, ctx context.Context, paths evalPaths) runner.SearchHit {
	t.Helper()
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          ragSearchText,
			MetadataKey:   ragMetadataKey,
			MetadataValue: ragMetadataValue,
			Limit:         5,
		},
	})
	if err != nil {
		t.Fatalf("metadata search: %v", err)
	}
	top, ok := topSearchHit(result)
	if !ok {
		t.Fatalf("metadata search returned no hits: %+v", result)
	}
	if searchHitPath(top) != ragCurrentPolicyPath {
		t.Fatalf("metadata search top path = %q, want %q; result=%+v", searchHitPath(top), ragCurrentPolicyPath, result.Search)
	}
	return top
}

func requirePopulatedAuthorityTopHit(t *testing.T, ctx context.Context, paths evalPaths) runner.SearchHit {
	t.Helper()
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          populatedSearchText,
			MetadataKey:   "populated_role",
			MetadataValue: "authority",
			Limit:         5,
		},
	})
	if err != nil {
		t.Fatalf("populated metadata search: %v", err)
	}
	top, ok := topSearchHit(result)
	if !ok {
		t.Fatalf("populated metadata search returned no hits: %+v", result)
	}
	if searchHitPath(top) != populatedAuthorityPath {
		t.Fatalf("populated metadata search top path = %q, want %q; result=%+v", searchHitPath(top), populatedAuthorityPath, result.Search)
	}
	if !searchHitHasCitation(top) {
		t.Fatalf("populated metadata search top missing citation: %+v", top)
	}
	return top
}

func comparisonResult(variant string, scenario string, passed bool, tools int, commands int, assistant int, nonCached int) jobResult {
	input := nonCached
	cached := 0
	output := 10
	return jobResult{
		Variant:  variant,
		Scenario: scenario,
		Passed:   passed,
		Metrics: metrics{
			ToolCalls:            tools,
			CommandExecutions:    commands,
			AssistantCalls:       assistant,
			UsageExposed:         true,
			InputTokens:          &input,
			CachedInputTokens:    &cached,
			NonCachedInputTokens: &nonCached,
			OutputTokens:         &output,
			EventTypeCounts:      map[string]int{},
		},
	}
}

func requireScenarioByID(t *testing.T, id string) scenario {
	t.Helper()
	for _, sc := range allScenarios() {
		if sc.ID == id {
			return sc
		}
	}
	t.Fatalf("missing scenario %q", id)
	return scenario{}
}

func containsValue(args []string, value string) bool {
	for _, arg := range args {
		if arg == value {
			return true
		}
	}
	return false
}
