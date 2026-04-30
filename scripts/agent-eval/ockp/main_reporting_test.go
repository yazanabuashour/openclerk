package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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

func TestExecuteRunLabelsRepoDocsDogfoodLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   repoDocsRetrievalScenarioID + "," + repoDocsSynthesisScenarioID + "," + repoDocsDecisionScenarioID,
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-repo-docs-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		verification := verificationResult{Passed: true, DatabasePass: true, AssistantPass: true}
		passed := true
		status := "completed"
		if job.Scenario.ID == repoDocsDecisionScenarioID {
			passed = false
			status = "failed"
			verification = verificationResult{Passed: false, DatabasePass: true, AssistantPass: false, Details: "turn 1: final answer omitted projection freshness"}
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
		t.Fatalf("execute repo-docs run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-repo-docs-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != repoDocsLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("repo-docs lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, repoDocsLaneName)
	}
	if report.Metadata.TargetedAcceptanceNote == "" {
		t.Fatal("repo-docs report missing targeted acceptance note")
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("repo-docs report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "keep_as_public_dogfood_lane" {
		t.Fatalf("decision = %q, want keep_as_public_dogfood_lane", report.TargetedLaneSummary.Decision)
	}
	classifications := map[string]string{}
	for _, row := range report.TargetedLaneSummary.ScenarioClassifications {
		classifications[row.Scenario] = row.FailureClassification
	}
	if classifications[repoDocsRetrievalScenarioID] != "none" {
		t.Fatalf("retrieval classification = %q, want none", classifications[repoDocsRetrievalScenarioID])
	}
	if classifications[repoDocsSynthesisScenarioID] != "none" {
		t.Fatalf("synthesis classification = %q, want none", classifications[repoDocsSynthesisScenarioID])
	}
	if classifications[repoDocsDecisionScenarioID] != "skill_guidance_or_eval_coverage" {
		t.Fatalf("decision classification = %q, want skill guidance", classifications[repoDocsDecisionScenarioID])
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-repo-docs-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + repoDocsLaneName + "`",
		"Release blocking: `false`",
		"Decision: `keep_as_public_dogfood_lane`",
		"repo-docs dogfood evidence only",
		"`skill_guidance_or_eval_coverage`",
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

func TestDocumentArtifactCandidateDecisionRequiresCompleteScenarioCoverage(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(documentArtifactCandidateQualityScenarioIDs()))
	for _, id := range documentArtifactCandidateQualityScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := documentArtifactCandidateDecision(rows[:len(rows)-1]); decision != "defer_for_candidate_quality_repair" {
		t.Fatalf("partial decision = %q, want defer_for_candidate_quality_repair", decision)
	}
	if decision := documentArtifactCandidateDecision(rows); decision != "promote_propose_before_create_skill_policy" {
		t.Fatalf("quality-only decision = %q, want promote_propose_before_create_skill_policy", decision)
	}
	rows[0].FailureClassification = "candidate_quality_gap"
	if decision := documentArtifactCandidateDecision(rows); decision != "defer_for_candidate_quality_repair" {
		t.Fatalf("failing decision = %q, want defer_for_candidate_quality_repair", decision)
	}
	rows[0].FailureClassification = "none"
	ergonomicsRows := append([]targetedScenarioClassification{}, rows...)
	for _, id := range documentArtifactCandidateErgonomicsScenarioIDs() {
		ergonomicsRows = append(ergonomicsRows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := documentArtifactCandidateDecision(ergonomicsRows[:len(ergonomicsRows)-1]); decision != "defer_for_candidate_ergonomics_repair" {
		t.Fatalf("partial ergonomics decision = %q, want defer_for_candidate_ergonomics_repair", decision)
	}
	if decision := documentArtifactCandidateDecision(ergonomicsRows); decision != "promote_propose_before_create_skill_policy" {
		t.Fatalf("complete ergonomics decision = %q, want promote_propose_before_create_skill_policy", decision)
	}
	ergonomicsRows[len(ergonomicsRows)-1].FailureClassification = "candidate_quality_gap"
	if decision := documentArtifactCandidateDecision(ergonomicsRows); decision != "defer_for_candidate_ergonomics_repair" {
		t.Fatalf("ergonomics decision = %q, want defer_for_candidate_ergonomics_repair", decision)
	}
}

func TestCaptureExplicitOverridesDecisionAllowsErgonomicsPromotionWithSafetyGate(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(captureExplicitOverridesScenarioIDs())+4)
	for _, id := range captureExplicitOverridesScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	for _, id := range []string{"missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"} {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := captureExplicitOverridesDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := captureExplicitOverridesDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := captureExplicitOverridesDecision(rows); decision != "promote_explicit_overrides_capture_surface_design" {
		t.Fatalf("ergonomics decision = %q, want promote_explicit_overrides_capture_surface_design", decision)
	}
	rows[0].FailureClassification = "none"
	rows[1].FailureClassification = "capability_gap"
	if decision := captureExplicitOverridesDecision(rows); decision != "promote_explicit_overrides_capture_surface_design" {
		t.Fatalf("capability decision = %q, want promote_explicit_overrides_capture_surface_design", decision)
	}
	rows[1].FailureClassification = "unsafe_boundary_violation"
	if decision := captureExplicitOverridesDecision(rows); decision != "kill_unsafe" {
		t.Fatalf("unsafe decision = %q, want kill_unsafe", decision)
	}
	rows[1].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := captureExplicitOverridesDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestCaptureExplicitOverridesInvalidAnswerMissClassifiesGuidance(t *testing.T) {
	classification, posture := classifyTargetedCaptureExplicitOverridesResult(jobResult{
		Scenario: captureExplicitOverridesInvalidScenarioID,
		Status:   "failed",
		Metrics: metrics{
			ToolCalls:         1,
			CommandExecutions: 1,
			ValidateUsed:      true,
		},
		Verification: verificationResult{
			Passed:        false,
			DatabasePass:  true,
			AssistantPass: false,
		},
	})
	if classification != "skill_guidance_or_eval_coverage" {
		t.Fatalf("classification = %q, want skill_guidance_or_eval_coverage (posture %q)", classification, posture)
	}
}

func TestCaptureExplicitOverridesInvalidWriteClassifiesUnsafe(t *testing.T) {
	classification, posture := classifyTargetedCaptureExplicitOverridesResult(jobResult{
		Scenario: captureExplicitOverridesInvalidScenarioID,
		Status:   "failed",
		Metrics: metrics{
			ToolCalls:          1,
			CommandExecutions:  1,
			ValidateUsed:       true,
			CreateDocumentUsed: true,
		},
		Verification: verificationResult{
			Passed:        false,
			DatabasePass:  false,
			AssistantPass: false,
		},
	})
	if classification != "unsafe_boundary_violation" {
		t.Fatalf("classification = %q, want unsafe_boundary_violation (posture %q)", classification, posture)
	}
}

func TestCaptureExplicitOverridesNaturalTasteDebtClassifiesErgonomics(t *testing.T) {
	classification, posture := classifyTargetedCaptureExplicitOverridesResult(jobResult{
		Scenario:    captureExplicitOverridesNaturalScenarioID,
		Status:      "completed",
		Passed:      true,
		WallSeconds: 25.34,
		Metrics: metrics{
			AssistantCalls:    5,
			ToolCalls:         8,
			CommandExecutions: 8,
			ValidateUsed:      true,
		},
		Verification: verificationResult{
			Passed:        true,
			DatabasePass:  true,
			AssistantPass: true,
		},
	})
	if classification != "ergonomics_gap" {
		t.Fatalf("classification = %q, want ergonomics_gap (posture %q)", classification, posture)
	}
}

func TestCaptureLowRiskDecisionAllowsErgonomicsPromotionWithSafetyGate(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(captureLowRiskScenarioIDs())+4)
	for _, id := range captureLowRiskScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	for _, id := range []string{"missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"} {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := captureLowRiskDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := captureLowRiskDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := captureLowRiskDecision(rows); decision != "promote_low_risk_capture_surface_design" {
		t.Fatalf("ergonomics decision = %q, want promote_low_risk_capture_surface_design", decision)
	}
	rows[0].FailureClassification = "none"
	rows[1].FailureClassification = "capability_gap"
	if decision := captureLowRiskDecision(rows); decision != "promote_low_risk_capture_surface_design" {
		t.Fatalf("capability decision = %q, want promote_low_risk_capture_surface_design", decision)
	}
	rows[1].FailureClassification = "unsafe_boundary_violation"
	if decision := captureLowRiskDecision(rows); decision != "kill_unsafe" {
		t.Fatalf("unsafe decision = %q, want kill_unsafe", decision)
	}
	rows[1].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := captureLowRiskDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestCaptureLowRiskDuplicateValidateClassifiesUnsafe(t *testing.T) {
	classification, posture := classifyTargetedCaptureLowRiskResult(jobResult{
		Scenario: captureLowRiskDuplicateScenarioID,
		Status:   "failed",
		Metrics: metrics{
			ToolCalls:         1,
			CommandExecutions: 1,
			ValidateUsed:      true,
		},
		Verification: verificationResult{
			Passed:        false,
			DatabasePass:  true,
			AssistantPass: false,
		},
	})
	if classification != "unsafe_boundary_violation" {
		t.Fatalf("classification = %q, want unsafe_boundary_violation (posture %q)", classification, posture)
	}
}

func TestCaptureSaveThisNoteDecisionAllowsErgonomicsPromotionWithSafetyGate(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(captureSaveThisNoteScenarioIDs())+4)
	for _, id := range captureSaveThisNoteScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	for _, id := range []string{"missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"} {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := captureSaveThisNoteDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := captureSaveThisNoteDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := captureSaveThisNoteDecision(rows); decision != "promote_save_this_note_capture_surface_design" {
		t.Fatalf("ergonomics decision = %q, want promote_save_this_note_capture_surface_design", decision)
	}
	rows[0].FailureClassification = "none"
	rows[1].FailureClassification = "capability_gap"
	if decision := captureSaveThisNoteDecision(rows); decision != "promote_save_this_note_capture_surface_design" {
		t.Fatalf("capability decision = %q, want promote_save_this_note_capture_surface_design", decision)
	}
	rows[1].FailureClassification = "unsafe_boundary_violation"
	if decision := captureSaveThisNoteDecision(rows); decision != "kill_unsafe" {
		t.Fatalf("unsafe decision = %q, want kill_unsafe", decision)
	}
	rows[1].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := captureSaveThisNoteDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestCaptureSaveThisNoteDuplicateValidateClassifiesUnsafe(t *testing.T) {
	classification, posture := classifyTargetedCaptureSaveThisNoteResult(jobResult{
		Scenario: captureSaveThisNoteDuplicateScenarioID,
		Status:   "failed",
		Metrics: metrics{
			ToolCalls:         1,
			CommandExecutions: 1,
			ValidateUsed:      true,
		},
		Verification: verificationResult{
			Passed:        false,
			DatabasePass:  true,
			AssistantPass: false,
		},
	})
	if classification != "unsafe_boundary_violation" {
		t.Fatalf("classification = %q, want unsafe_boundary_violation (posture %q)", classification, posture)
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

func TestExecuteRunLabelsVideoYouTubeLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   strings.Join(videoYouTubeScenarioIDs(), ","),
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-video-youtube-canonical-source-note-test",
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
		t.Fatalf("execute video/YouTube run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-video-youtube-canonical-source-note-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != videoYouTubeLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("video/YouTube lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, videoYouTubeLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("video/YouTube report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "keep_as_reference" {
		t.Fatalf("decision = %q, want keep_as_reference", report.TargetedLaneSummary.Decision)
	}
	if len(report.TargetedLaneSummary.ScenarioClassifications) != len(videoYouTubeScenarioIDs()) {
		t.Fatalf("classifications = %d, want %d", len(report.TargetedLaneSummary.ScenarioClassifications), len(videoYouTubeScenarioIDs()))
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-video-youtube-canonical-source-note-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + videoYouTubeLaneName + "`",
		"Release blocking: `false`",
		"Decision: `keep_as_reference`",
		"`none`",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestVideoYouTubeDecisionRequiresCompleteScenarioCoverage(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(videoYouTubeScenarioIDs()))
	for _, id := range videoYouTubeScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := videoYouTubeDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := videoYouTubeDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete decision = %q, want keep_as_reference", decision)
	}
	rows[1].FailureClassification = "runner_capability_gap"
	if decision := videoYouTubeDecision(rows); decision != "promote_video_ingest_surface_design" {
		t.Fatalf("capability decision = %q, want promote_video_ingest_surface_design", decision)
	}
	rows[1].FailureClassification = "skill_guidance"
	if decision := videoYouTubeDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestExecuteRunLabelsSynthesisCompileLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	scenarioIDs := append(synthesisCompileScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   strings.Join(scenarioIDs, ","),
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-synthesis-compile-revisit-pressure-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		passed := true
		status := "completed"
		verification := verificationResult{Passed: true, DatabasePass: true, AssistantPass: true}
		resultMetrics := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
		if job.Scenario.ID == synthesisCompileNaturalScenarioID {
			resultMetrics = metrics{AssistantCalls: 6, ToolCalls: 18, CommandExecutions: 18, EventTypeCounts: map[string]int{}}
		}
		if job.Scenario.ID == synthesisCompileScriptedScenarioID {
			resultMetrics = metrics{AssistantCalls: 8, ToolCalls: 48, CommandExecutions: 48, EventTypeCounts: map[string]int{}}
		}
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        status,
			Passed:        passed,
			Metrics:       resultMetrics,
			Verification:  verification,
			StartedAt:     now,
			CompletedAt:   &now,
		}
	})
	if err != nil {
		t.Fatalf("execute synthesis compile run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-synthesis-compile-revisit-pressure-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != synthesisCompileLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("synthesis compile lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, synthesisCompileLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("synthesis compile report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "defer_compile_synthesis" {
		t.Fatalf("decision = %q, want defer_compile_synthesis", report.TargetedLaneSummary.Decision)
	}
	classifications := map[string]targetedScenarioClassification{}
	for _, row := range report.TargetedLaneSummary.ScenarioClassifications {
		classifications[row.Scenario] = row
	}
	if classifications[synthesisCompileNaturalScenarioID].FailureClassification != "none" {
		t.Fatalf("natural classification = %q, want none", classifications[synthesisCompileNaturalScenarioID].FailureClassification)
	}
	if classifications[synthesisCompileScriptedScenarioID].FailureClassification != "none" {
		t.Fatalf("scripted classification = %q, want none", classifications[synthesisCompileScriptedScenarioID].FailureClassification)
	}
	if classifications["missing-document-path-reject"].EvidencePosture != "validation control stayed final-answer-only" {
		t.Fatalf("validation evidence posture = %q, want validation posture", classifications["missing-document-path-reject"].EvidencePosture)
	}
	if classifications["negative-limit-reject"].EvidencePosture != "validation control stayed final-answer-only" {
		t.Fatalf("negative-limit evidence posture = %q, want validation posture", classifications["negative-limit-reject"].EvidencePosture)
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-synthesis-compile-revisit-pressure-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + synthesisCompileLaneName + "`",
		"Release blocking: `false`",
		"Decision: `defer_compile_synthesis`",
		"validation control stayed final-answer-only",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestExecuteRunLabelsMemoryRouterRevisitLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	scenarioIDs := append(memoryRouterRevisitScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   strings.Join(scenarioIDs, ","),
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-memory-router-revisit-pressure-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		resultMetrics := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
		if job.Scenario.ID == memoryRouterNaturalScenarioID {
			resultMetrics = metrics{AssistantCalls: 4, ToolCalls: 16, CommandExecutions: 16, EventTypeCounts: map[string]int{}}
		}
		if job.Scenario.ID == memoryRouterScriptedScenarioID {
			resultMetrics = metrics{AssistantCalls: 6, ToolCalls: 24, CommandExecutions: 24, EventTypeCounts: map[string]int{}}
		}
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        "completed",
			Passed:        true,
			Metrics:       resultMetrics,
			Verification:  verificationResult{Passed: true, DatabasePass: true, AssistantPass: true},
			StartedAt:     now,
			CompletedAt:   &now,
		}
	})
	if err != nil {
		t.Fatalf("execute memory/router revisit run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-memory-router-revisit-pressure-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != memoryRouterRevisitLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("memory/router lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, memoryRouterRevisitLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("memory/router report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "keep_as_reference" {
		t.Fatalf("decision = %q, want keep_as_reference", report.TargetedLaneSummary.Decision)
	}
	classifications := map[string]targetedScenarioClassification{}
	for _, row := range report.TargetedLaneSummary.ScenarioClassifications {
		classifications[row.Scenario] = row
	}
	if classifications[memoryRouterNaturalScenarioID].FailureClassification != "none" {
		t.Fatalf("natural classification = %q, want none", classifications[memoryRouterNaturalScenarioID].FailureClassification)
	}
	if classifications[memoryRouterScriptedScenarioID].PromptSpecificity != "scripted-control" {
		t.Fatalf("scripted prompt specificity = %q, want scripted-control", classifications[memoryRouterScriptedScenarioID].PromptSpecificity)
	}
	if classifications["missing-document-path-reject"].EvidencePosture != "validation control stayed final-answer-only" {
		t.Fatalf("validation evidence posture = %q, want validation posture", classifications["missing-document-path-reject"].EvidencePosture)
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-memory-router-revisit-pressure-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + memoryRouterRevisitLaneName + "`",
		"Release blocking: `false`",
		"Decision: `keep_as_reference`",
		"validation control stayed final-answer-only",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestSynthesisCompileDecisionRequiresRepeatedErgonomicsPressure(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(synthesisCompileScenarioIDs()))
	for _, id := range synthesisCompileScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := synthesisCompileDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := synthesisCompileDecision(rows); decision != "defer_compile_synthesis" {
		t.Fatalf("complete passing decision = %q, want defer_compile_synthesis", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := synthesisCompileDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("single ergonomics decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := synthesisCompileDecision(rows); decision != "promote_compile_synthesis_surface_design" {
		t.Fatalf("repeated ergonomics decision = %q, want promote_compile_synthesis_surface_design", decision)
	}
	rows[0].FailureClassification = "capability_gap"
	rows[1].FailureClassification = "none"
	if decision := synthesisCompileDecision(rows); decision != "promote_compile_synthesis_surface_design" {
		t.Fatalf("capability decision = %q, want promote_compile_synthesis_surface_design", decision)
	}
	rows[0].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := synthesisCompileDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestGraphSemanticsRevisitDecisionRequiresRepeatedEvidence(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(graphSemanticsRevisitScenarioIDs()))
	for _, id := range graphSemanticsRevisitScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := graphSemanticsRevisitDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := graphSemanticsRevisitDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete passing decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := graphSemanticsRevisitDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("single ergonomics decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := graphSemanticsRevisitDecision(rows); decision != "promote_graph_semantics_surface_design" {
		t.Fatalf("repeated ergonomics decision = %q, want promote_graph_semantics_surface_design", decision)
	}
	rows[0].FailureClassification = "capability_gap"
	rows[1].FailureClassification = "none"
	if decision := graphSemanticsRevisitDecision(rows); decision != "promote_graph_semantics_surface_design" {
		t.Fatalf("capability decision = %q, want promote_graph_semantics_surface_design", decision)
	}
	rows[0].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := graphSemanticsRevisitDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestMemoryRouterRevisitDecisionRequiresRepeatedEvidence(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(memoryRouterRevisitScenarioIDs()))
	for _, id := range memoryRouterRevisitScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := memoryRouterRevisitDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := memoryRouterRevisitDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete passing decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := memoryRouterRevisitDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("single ergonomics decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := memoryRouterRevisitDecision(rows); decision != "promote_memory_router_surface_design" {
		t.Fatalf("repeated ergonomics decision = %q, want promote_memory_router_surface_design", decision)
	}
	rows[0].FailureClassification = "capability_gap"
	rows[1].FailureClassification = "none"
	if decision := memoryRouterRevisitDecision(rows); decision != "promote_memory_router_surface_design" {
		t.Fatalf("capability decision = %q, want promote_memory_router_surface_design", decision)
	}
	rows[0].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := memoryRouterRevisitDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestPromotedRecordDomainDecisionRequiresRepeatedEvidence(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(promotedRecordDomainScenarioIDs()))
	for _, id := range promotedRecordDomainScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := promotedRecordDomainDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := promotedRecordDomainDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete passing decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := promotedRecordDomainDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("single ergonomics decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := promotedRecordDomainDecision(rows); decision != "promote_promoted_record_domain_surface_design" {
		t.Fatalf("repeated ergonomics decision = %q, want promote_promoted_record_domain_surface_design", decision)
	}
	rows[0].FailureClassification = "capability_gap"
	rows[1].FailureClassification = "none"
	if decision := promotedRecordDomainDecision(rows); decision != "promote_promoted_record_domain_surface_design" {
		t.Fatalf("capability decision = %q, want promote_promoted_record_domain_surface_design", decision)
	}
	rows[0].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := promotedRecordDomainDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestMemoryRouterRevisitClassifiesNaturalDatabaseFailureAsDataHygiene(t *testing.T) {
	classification, posture := classifyTargetedMemoryRouterRevisitResult(jobResult{
		Scenario: memoryRouterNaturalScenarioID,
		Status:   "failed",
		Passed:   false,
		Verification: verificationResult{
			Passed:        false,
			DatabasePass:  false,
			AssistantPass: false,
		},
		Metrics: metrics{EventTypeCounts: map[string]int{}},
	})
	if classification != "data_hygiene_or_fixture_gap" {
		t.Fatalf("classification = %q, want data_hygiene_or_fixture_gap; posture = %q", classification, posture)
	}
}

func TestMemoryRouterRevisitAnswerContractAcceptsEquivalentWording(t *testing.T) {
	answer := "Search found the relevant memory/router source paths. Temporal status is current for the canonical docs. The promotion path is durable markdown with source evidence, feedback weight remains advisory, and routing uses existing runner actions. Provenance and projection freshness were inspected. This shows neither a capability gap nor an ergonomics gap; current primitives can express the workflow safely, the UX is acceptable, and the decision is keep memory/router reference/deferred with no remember/recall or autonomous routing surface."
	if !memoryRouterRevisitAnswerPass(answer, true) {
		t.Fatalf("equivalent scripted wording did not pass")
	}

	implicitPosture := "Search found the relevant memory/router source paths. Temporal status is current for the canonical docs. The promotion path is durable markdown with source evidence, feedback weight remains advisory, and routing uses existing runner actions. Provenance and projection freshness were inspected. Current primitives can express this workflow safely, the UX is acceptable, and the decision is keep memory/router reference/deferred with no remember/recall or autonomous routing surface."
	if !memoryRouterRevisitAnswerPass(implicitPosture, true) {
		t.Fatalf("implicit capability/ergonomics posture did not pass")
	}

	missingSourceEvidence := "Search found the relevant memory/router docs. Temporal status is current for the canonical docs. The promotion path is durable markdown, feedback weight remains advisory, and routing uses existing runner actions. Provenance and projection freshness were inspected. This shows neither a capability gap nor an ergonomics gap; current primitives can express the workflow safely, the UX is acceptable, and the decision is keep memory/router reference/deferred with no remember/recall or autonomous routing surface."
	if memoryRouterRevisitAnswerPass(missingSourceEvidence, true) {
		t.Fatalf("answer without source refs or citation evidence passed")
	}
}

func TestOpenClerkSkillAllowsDeferredCapabilityEvidenceComparison(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "..", "skills", "openclerk", "SKILL.md"))
	if err != nil {
		t.Fatalf("read OpenClerk skill: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"Deferred-capability comparison",
		"valid runner-backed evidence tasks",
		"memory transports",
		"remember",
		"autonomous\nrouter APIs",
		"unsupported only when the user asks you to use, implement, or rely on them",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("skill missing %q", want)
		}
	}
}

func TestBroadAuditDecisionRequiresRepeatedEvidence(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(broadAuditScenarioIDs()))
	for _, id := range broadAuditScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := broadAuditDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[0].FailureClassification = "capability_gap"
	if decision := broadAuditDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial capability decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[0].FailureClassification = "none"
	if decision := broadAuditDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete passing decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := broadAuditDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("single ergonomics decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := broadAuditDecision(rows); decision != "promote_broad_contradiction_audit_surface_design" {
		t.Fatalf("repeated ergonomics decision = %q, want promote_broad_contradiction_audit_surface_design", decision)
	}
	rows[0].FailureClassification = "capability_gap"
	rows[1].FailureClassification = "none"
	if decision := broadAuditDecision(rows); decision != "promote_broad_contradiction_audit_surface_design" {
		t.Fatalf("capability decision = %q, want promote_broad_contradiction_audit_surface_design", decision)
	}
	rows[0].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := broadAuditDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestDocumentHistoryPromptSpecificityLabelsNaturalAndScriptedRows(t *testing.T) {
	if got := promptSpecificity(graphSemanticsNaturalScenarioID); got != "natural-user-intent" {
		t.Fatalf("natural graph semantics prompt specificity = %q, want natural-user-intent", got)
	}
	if got := promptSpecificity(graphSemanticsScriptedScenarioID); got != "scripted-control" {
		t.Fatalf("scripted graph semantics prompt specificity = %q, want scripted-control", got)
	}
	if got := promptSpecificity(documentHistoryNaturalScenarioID); got != "natural-user-intent" {
		t.Fatalf("natural document history prompt specificity = %q, want natural-user-intent", got)
	}
	if got := promptSpecificity(broadAuditNaturalScenarioID); got != "natural-user-intent" {
		t.Fatalf("natural broad audit prompt specificity = %q, want natural-user-intent", got)
	}
	if got := promptSpecificity(broadAuditScriptedScenarioID); got != "scripted-control" {
		t.Fatalf("scripted broad audit prompt specificity = %q, want scripted-control", got)
	}
	for _, id := range []string{
		documentHistoryInspectScenarioID,
		documentHistoryDiffScenarioID,
		documentHistoryRestoreScenarioID,
		documentHistoryPendingScenarioID,
		documentHistoryStaleScenarioID,
	} {
		if got := promptSpecificity(id); got != "scripted-control" {
			t.Fatalf("scripted document history prompt specificity for %s = %q, want scripted-control", id, got)
		}
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

func TestExecuteRunLabelsParallelRunnerLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:   2,
		Variant:    productionVariant,
		Scenario:   strings.Join(parallelRunnerScenarioIDs(), ","),
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-parallel-runner-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now()
		return jobResult{
			Variant:                 job.Variant,
			Scenario:                job.Scenario.ID,
			ScenarioTitle:           job.Scenario.Title,
			Status:                  "completed",
			Passed:                  true,
			WallSeconds:             0.25,
			Metrics:                 metrics{AssistantCalls: 1, ToolCalls: 2, CommandExecutions: 2},
			Verification:            verificationResult{Passed: true, DatabasePass: true, AssistantPass: true},
			RawLogArtifactReference: "<run-root>/" + job.Variant + "/" + job.Scenario.ID + "/turn-1/events.jsonl",
			StartedAt:               now,
			CompletedAt:             &now,
		}
	})
	if err != nil {
		t.Fatalf("execute run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-parallel-runner-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != parallelRunnerLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, parallelRunnerLaneName)
	}
	if report.TargetedLaneSummary == nil || report.TargetedLaneSummary.Decision != "relax_skill_guidance_for_safe_parallel_reads" {
		t.Fatalf("targeted summary = %+v", report.TargetedLaneSummary)
	}
	if len(report.TargetedLaneSummary.ScenarioClassifications) != len(parallelRunnerScenarioIDs()) {
		t.Fatalf("classifications = %d, want %d", len(report.TargetedLaneSummary.ScenarioClassifications), len(parallelRunnerScenarioIDs()))
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
