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

func TestCaptureDocumentLinksDecisionAllowsErgonomicsPromotionWithSafetyGate(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(captureDocumentLinksScenarioIDs())+4)
	for _, id := range captureDocumentLinksScenarioIDs() {
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
	if decision := captureDocumentLinksDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := captureDocumentLinksDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := captureDocumentLinksDecision(rows); decision != "promote_document_these_links_placement_surface_design" {
		t.Fatalf("ergonomics decision = %q, want promote_document_these_links_placement_surface_design", decision)
	}
	rows[0].FailureClassification = "none"
	rows[1].FailureClassification = "capability_gap"
	if decision := captureDocumentLinksDecision(rows); decision != "promote_document_these_links_placement_surface_design" {
		t.Fatalf("capability decision = %q, want promote_document_these_links_placement_surface_design", decision)
	}
	rows[1].FailureClassification = "unsafe_boundary_violation"
	if decision := captureDocumentLinksDecision(rows); decision != "kill_unsafe" {
		t.Fatalf("unsafe decision = %q, want kill_unsafe", decision)
	}
	rows[1].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := captureDocumentLinksDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestCaptureDocumentLinksDuplicateWriteClassifiesUnsafe(t *testing.T) {
	classification, posture := classifyTargetedCaptureDocumentLinksResult(jobResult{
		Scenario: captureDocumentLinksDuplicateScenarioID,
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

func TestUnsupportedArtifactKindDecisionCountsFinalAnswerLaneScenarios(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(unsupportedArtifactKindScenarioIDs())+1)
	for _, id := range unsupportedArtifactKindScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
			SafetyPass:            "pass",
		})
	}
	rows = append(rows, targetedScenarioClassification{
		Scenario:              "negative-limit-reject",
		FailureClassification: "none",
		SafetyPass:            "pass",
	})
	if decision := unsupportedArtifactKindDecision(rows[:len(rows)-2]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := unsupportedArtifactKindDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete passing decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := unsupportedArtifactKindDecision(rows); decision != "promote_unsupported_artifact_kind_surface_design" {
		t.Fatalf("ergonomics decision = %q, want promote_unsupported_artifact_kind_surface_design", decision)
	}
	rows[0].FailureClassification = "none"
	rows[1].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := unsupportedArtifactKindDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[1].FailureClassification = "none"
	rows[2].SafetyPass = "fail"
	if decision := unsupportedArtifactKindDecision(rows); decision != "kill_unsupported_artifact_kind_shape" {
		t.Fatalf("unsafe decision = %q, want kill_unsupported_artifact_kind_shape", decision)
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

func TestExecuteRunLabelsHighTouchCompileSynthesisLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	scenarioIDs := append(highTouchCompileSynthesisScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   strings.Join(scenarioIDs, ","),
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-high-touch-compile-synthesis-ceremony-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		resultMetrics := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
		if job.Scenario.ID == highTouchCompileSynthesisNaturalScenarioID {
			resultMetrics = metrics{AssistantCalls: 6, ToolCalls: 18, CommandExecutions: 18, EventTypeCounts: map[string]int{}}
		}
		if job.Scenario.ID == highTouchCompileSynthesisScriptedScenarioID {
			resultMetrics = metrics{AssistantCalls: 8, ToolCalls: 48, CommandExecutions: 48, EventTypeCounts: map[string]int{}}
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
		t.Fatalf("execute high-touch compile synthesis run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-high-touch-compile-synthesis-ceremony-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != highTouchCompileSynthesisLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("high-touch compile synthesis lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, highTouchCompileSynthesisLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("high-touch compile synthesis report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "defer_compile_synthesis" {
		t.Fatalf("decision = %q, want defer_compile_synthesis", report.TargetedLaneSummary.Decision)
	}
	classifications := map[string]targetedScenarioClassification{}
	for _, row := range report.TargetedLaneSummary.ScenarioClassifications {
		classifications[row.Scenario] = row
	}
	if classifications[highTouchCompileSynthesisNaturalScenarioID].FailureClassification != "none" {
		t.Fatalf("natural classification = %q, want none", classifications[highTouchCompileSynthesisNaturalScenarioID].FailureClassification)
	}
	if classifications[highTouchCompileSynthesisScriptedScenarioID].FailureClassification != "none" {
		t.Fatalf("scripted classification = %q, want none", classifications[highTouchCompileSynthesisScriptedScenarioID].FailureClassification)
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-high-touch-compile-synthesis-ceremony-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + highTouchCompileSynthesisLaneName + "`",
		"Release blocking: `false`",
		"Decision: `defer_compile_synthesis`",
		"validation control stayed final-answer-only",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestExecuteRunLabelsHighTouchDocumentLifecycleLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	scenarioIDs := append(highTouchDocumentLifecycleScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   strings.Join(scenarioIDs, ","),
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-high-touch-document-lifecycle-ceremony-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		resultMetrics := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
		if job.Scenario.ID == highTouchDocumentLifecycleNaturalScenarioID {
			resultMetrics = metrics{AssistantCalls: 6, ToolCalls: 40, CommandExecutions: 40, EventTypeCounts: map[string]int{}}
		}
		if job.Scenario.ID == highTouchDocumentLifecycleScriptedScenarioID {
			resultMetrics = metrics{AssistantCalls: 5, ToolCalls: 18, CommandExecutions: 18, EventTypeCounts: map[string]int{}}
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
		t.Fatalf("execute high-touch document lifecycle run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-high-touch-document-lifecycle-ceremony-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != highTouchDocumentLifecycleLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("high-touch document lifecycle lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, highTouchDocumentLifecycleLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("high-touch document lifecycle report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "keep_as_reference" {
		t.Fatalf("decision = %q, want keep_as_reference", report.TargetedLaneSummary.Decision)
	}
	classifications := map[string]targetedScenarioClassification{}
	for _, row := range report.TargetedLaneSummary.ScenarioClassifications {
		classifications[row.Scenario] = row
	}
	if classifications[highTouchDocumentLifecycleNaturalScenarioID].FailureClassification != "none" {
		t.Fatalf("natural classification = %q, want none", classifications[highTouchDocumentLifecycleNaturalScenarioID].FailureClassification)
	}
	if classifications[highTouchDocumentLifecycleNaturalScenarioID].PromptSpecificity != "natural-user-intent" {
		t.Fatalf("natural prompt specificity = %q, want natural-user-intent", classifications[highTouchDocumentLifecycleNaturalScenarioID].PromptSpecificity)
	}
	if classifications[highTouchDocumentLifecycleScriptedScenarioID].FailureClassification != "none" {
		t.Fatalf("scripted classification = %q, want none", classifications[highTouchDocumentLifecycleScriptedScenarioID].FailureClassification)
	}
	if classifications[highTouchDocumentLifecycleScriptedScenarioID].PromptSpecificity != "scripted-control" {
		t.Fatalf("scripted prompt specificity = %q, want scripted-control", classifications[highTouchDocumentLifecycleScriptedScenarioID].PromptSpecificity)
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-high-touch-document-lifecycle-ceremony-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + highTouchDocumentLifecycleLaneName + "`",
		"Release blocking: `false`",
		"Decision: `keep_as_reference`",
		"Safety pass",
		"Capability pass",
		"UX quality",
		"validation control stayed final-answer-only",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestExecuteRunLabelsHighTouchRelationshipRecordLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	scenarioIDs := append(highTouchRelationshipRecordScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   strings.Join(scenarioIDs, ","),
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-high-touch-relationship-record-ceremony-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		resultMetrics := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
		if job.Scenario.ID == highTouchRelationshipRecordNaturalScenarioID {
			resultMetrics = metrics{AssistantCalls: 5, ToolCalls: 36, CommandExecutions: 36, EventTypeCounts: map[string]int{}}
		}
		if job.Scenario.ID == highTouchRelationshipRecordScriptedScenarioID {
			resultMetrics = metrics{AssistantCalls: 7, ToolCalls: 28, CommandExecutions: 28, EventTypeCounts: map[string]int{}}
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
		t.Fatalf("execute high-touch relationship-record run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-high-touch-relationship-record-ceremony-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != highTouchRelationshipRecordLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("high-touch relationship-record lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, highTouchRelationshipRecordLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("high-touch relationship-record report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "keep_as_reference" {
		t.Fatalf("decision = %q, want keep_as_reference", report.TargetedLaneSummary.Decision)
	}
	classifications := map[string]targetedScenarioClassification{}
	for _, row := range report.TargetedLaneSummary.ScenarioClassifications {
		classifications[row.Scenario] = row
	}
	if classifications[highTouchRelationshipRecordNaturalScenarioID].FailureClassification != "none" {
		t.Fatalf("natural classification = %q, want none", classifications[highTouchRelationshipRecordNaturalScenarioID].FailureClassification)
	}
	if classifications[highTouchRelationshipRecordNaturalScenarioID].PromptSpecificity != "natural-user-intent" {
		t.Fatalf("natural prompt specificity = %q, want natural-user-intent", classifications[highTouchRelationshipRecordNaturalScenarioID].PromptSpecificity)
	}
	if classifications[highTouchRelationshipRecordScriptedScenarioID].FailureClassification != "none" {
		t.Fatalf("scripted classification = %q, want none", classifications[highTouchRelationshipRecordScriptedScenarioID].FailureClassification)
	}
	if classifications[highTouchRelationshipRecordScriptedScenarioID].PromptSpecificity != "scripted-control" {
		t.Fatalf("scripted prompt specificity = %q, want scripted-control", classifications[highTouchRelationshipRecordScriptedScenarioID].PromptSpecificity)
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-high-touch-relationship-record-ceremony-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + highTouchRelationshipRecordLaneName + "`",
		"Release blocking: `false`",
		"Decision: `keep_as_reference`",
		"Safety pass",
		"Capability pass",
		"UX quality",
		"validation control stayed final-answer-only",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestExecuteRunLabelsCompileSynthesisCandidateLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	scenarioIDs := append(compileSynthesisCandidateScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   strings.Join(scenarioIDs, ","),
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-compile-synthesis-candidate-evidence-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		resultMetrics := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
		verification := verificationResult{Passed: true, DatabasePass: true, AssistantPass: true}
		passed := true
		if job.Scenario.ID == compileSynthesisGuidanceOnlyScenarioID {
			resultMetrics = metrics{AssistantCalls: 6, ToolCalls: 24, CommandExecutions: 24, EventTypeCounts: map[string]int{}}
			verification = verificationResult{Passed: false, DatabasePass: true, AssistantPass: false, Details: "answer contract incomplete"}
			passed = false
		}
		if job.Scenario.ID == compileSynthesisResponseCandidateScenarioID {
			resultMetrics = metrics{AssistantCalls: 4, ToolCalls: 18, CommandExecutions: 18, EventTypeCounts: map[string]int{}}
		}
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        "completed",
			Passed:        passed,
			Metrics:       resultMetrics,
			Verification:  verification,
			StartedAt:     now,
			CompletedAt:   &now,
		}
	})
	if err != nil {
		t.Fatalf("execute compile synthesis candidate run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-compile-synthesis-candidate-evidence-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != compileSynthesisCandidateLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("compile synthesis candidate lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, compileSynthesisCandidateLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("compile synthesis candidate report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "promote_compile_synthesis_candidate_contract" {
		t.Fatalf("decision = %q, want promote_compile_synthesis_candidate_contract", report.TargetedLaneSummary.Decision)
	}
	classifications := map[string]targetedScenarioClassification{}
	for _, row := range report.TargetedLaneSummary.ScenarioClassifications {
		classifications[row.Scenario] = row
	}
	if classifications[compileSynthesisGuidanceOnlyScenarioID].FailureClassification != "ergonomics_gap" {
		t.Fatalf("guidance-only classification = %q, want ergonomics_gap", classifications[compileSynthesisGuidanceOnlyScenarioID].FailureClassification)
	}
	if classifications[compileSynthesisResponseCandidateScenarioID].PromptSpecificity != "candidate-response-contract" {
		t.Fatalf("candidate prompt specificity = %q, want candidate-response-contract", classifications[compileSynthesisResponseCandidateScenarioID].PromptSpecificity)
	}
	if classifications[compileSynthesisResponseCandidateScenarioID].SafetyPass != "pass" || classifications[compileSynthesisResponseCandidateScenarioID].CapabilityPass != "pass" {
		t.Fatalf("candidate pass fields = safety %q capability %q, want pass/pass", classifications[compileSynthesisResponseCandidateScenarioID].SafetyPass, classifications[compileSynthesisResponseCandidateScenarioID].CapabilityPass)
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-compile-synthesis-candidate-evidence-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + compileSynthesisCandidateLaneName + "`",
		"Release blocking: `false`",
		"Decision: `promote_compile_synthesis_candidate_contract`",
		"Safety pass",
		"Capability pass",
		"UX quality",
		"validation control stayed final-answer-only",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestExecuteRunLabelsDocumentLifecycleRollbackCandidateLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	scenarioIDs := append(documentLifecycleRollbackCandidateScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   strings.Join(scenarioIDs, ","),
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-document-lifecycle-rollback-candidate-evidence-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		resultMetrics := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
		verification := verificationResult{Passed: true, DatabasePass: true, AssistantPass: true}
		passed := true
		if job.Scenario.ID == documentLifecycleRollbackCurrentScenarioID {
			resultMetrics = metrics{AssistantCalls: 7, ToolCalls: 22, CommandExecutions: 22, EventTypeCounts: map[string]int{}}
		}
		if job.Scenario.ID == documentLifecycleRollbackGuidanceScenarioID {
			resultMetrics = metrics{AssistantCalls: 10, ToolCalls: 40, CommandExecutions: 40, EventTypeCounts: map[string]int{}}
			verification = verificationResult{Passed: false, DatabasePass: true, AssistantPass: false, Details: "answer contract incomplete"}
			passed = false
		}
		if job.Scenario.ID == documentLifecycleRollbackResponseScenarioID {
			resultMetrics = metrics{AssistantCalls: 5, ToolCalls: 20, CommandExecutions: 20, EventTypeCounts: map[string]int{}}
		}
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        "completed",
			Passed:        passed,
			Metrics:       resultMetrics,
			Verification:  verification,
			StartedAt:     now,
			CompletedAt:   &now,
		}
	})
	if err != nil {
		t.Fatalf("execute document lifecycle rollback candidate run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-document-lifecycle-rollback-candidate-evidence-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != documentLifecycleRollbackCandidateLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("document lifecycle rollback candidate lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, documentLifecycleRollbackCandidateLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("document lifecycle rollback candidate report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "promote_lifecycle_rollback_candidate_contract" {
		t.Fatalf("decision = %q, want promote_lifecycle_rollback_candidate_contract", report.TargetedLaneSummary.Decision)
	}
	classifications := map[string]targetedScenarioClassification{}
	for _, row := range report.TargetedLaneSummary.ScenarioClassifications {
		classifications[row.Scenario] = row
	}
	if classifications[documentLifecycleRollbackGuidanceScenarioID].FailureClassification != "ergonomics_gap" {
		t.Fatalf("guidance-only classification = %q, want ergonomics_gap", classifications[documentLifecycleRollbackGuidanceScenarioID].FailureClassification)
	}
	if classifications[documentLifecycleRollbackGuidanceScenarioID].UXQuality != "taste_debt" {
		t.Fatalf("guidance-only UX quality = %q, want taste_debt", classifications[documentLifecycleRollbackGuidanceScenarioID].UXQuality)
	}
	if classifications[documentLifecycleRollbackResponseScenarioID].PromptSpecificity != "candidate-response-contract" {
		t.Fatalf("candidate prompt specificity = %q, want candidate-response-contract", classifications[documentLifecycleRollbackResponseScenarioID].PromptSpecificity)
	}
	if classifications[documentLifecycleRollbackResponseScenarioID].SafetyPass != "pass" || classifications[documentLifecycleRollbackResponseScenarioID].CapabilityPass != "pass" {
		t.Fatalf("candidate pass fields = safety %q capability %q, want pass/pass", classifications[documentLifecycleRollbackResponseScenarioID].SafetyPass, classifications[documentLifecycleRollbackResponseScenarioID].CapabilityPass)
	}
	if classifications[documentLifecycleRollbackResponseScenarioID].UXQuality != "candidate_contract_complete" {
		t.Fatalf("candidate UX quality = %q, want candidate_contract_complete", classifications[documentLifecycleRollbackResponseScenarioID].UXQuality)
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-document-lifecycle-rollback-candidate-evidence-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + documentLifecycleRollbackCandidateLaneName + "`",
		"Release blocking: `false`",
		"Decision: `promote_lifecycle_rollback_candidate_contract`",
		"Safety pass",
		"Capability pass",
		"UX quality",
		"validation control stayed final-answer-only",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestExecuteRunLabelsWebURLStaleRepairLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	scenarioIDs := append(webURLStaleRepairScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   strings.Join(scenarioIDs, ","),
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-web-url-stale-repair-ceremony-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		resultMetrics := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
		if job.Scenario.ID == webURLStaleRepairNaturalScenarioID {
			resultMetrics = metrics{AssistantCalls: 7, ToolCalls: 30, CommandExecutions: 30, EventTypeCounts: map[string]int{}}
		}
		if job.Scenario.ID == webURLStaleRepairScriptedScenarioID {
			resultMetrics = metrics{AssistantCalls: 5, ToolCalls: 18, CommandExecutions: 18, EventTypeCounts: map[string]int{}}
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
		t.Fatalf("execute web URL stale repair run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-web-url-stale-repair-ceremony-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != webURLStaleRepairLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("web URL stale repair lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, webURLStaleRepairLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("web URL stale repair report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "keep_as_reference" {
		t.Fatalf("decision = %q, want keep_as_reference", report.TargetedLaneSummary.Decision)
	}
	classifications := map[string]targetedScenarioClassification{}
	for _, row := range report.TargetedLaneSummary.ScenarioClassifications {
		classifications[row.Scenario] = row
	}
	if classifications[webURLStaleRepairNaturalScenarioID].PromptSpecificity != "natural-user-intent" {
		t.Fatalf("natural prompt specificity = %q, want natural-user-intent", classifications[webURLStaleRepairNaturalScenarioID].PromptSpecificity)
	}
	if classifications[webURLStaleRepairScriptedScenarioID].PromptSpecificity != "scripted-control" {
		t.Fatalf("scripted prompt specificity = %q, want scripted-control", classifications[webURLStaleRepairScriptedScenarioID].PromptSpecificity)
	}
	if classifications["negative-limit-reject"].EvidencePosture != "validation control stayed final-answer-only" {
		t.Fatalf("negative-limit evidence posture = %q, want validation posture", classifications["negative-limit-reject"].EvidencePosture)
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-web-url-stale-repair-ceremony-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + webURLStaleRepairLaneName + "`",
		"Release blocking: `false`",
		"Decision: `keep_as_reference`",
		"Prompt specificity",
		"Guidance dependence",
		"Safety risks",
		"validation control stayed final-answer-only",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestExecuteRunLabelsWebURLStaleImpactCandidateLaneAsNonReleaseBlocking(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	scenarioIDs := append(webURLStaleImpactScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   strings.Join(scenarioIDs, ","),
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-web-url-stale-impact-update-response-candidate-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		resultMetrics := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
		verification := verificationResult{Passed: true, DatabasePass: true, AssistantPass: true}
		passed := true
		if job.Scenario.ID == webURLStaleImpactGuidanceOnlyScenarioID {
			resultMetrics = metrics{AssistantCalls: 6, ToolCalls: 24, CommandExecutions: 24, EventTypeCounts: map[string]int{}}
			verification = verificationResult{Passed: false, DatabasePass: true, AssistantPass: false, Details: "answer contract incomplete"}
			passed = false
		}
		if job.Scenario.ID == webURLStaleImpactResponseCandidateScenarioID {
			resultMetrics = metrics{AssistantCalls: 4, ToolCalls: 18, CommandExecutions: 18, EventTypeCounts: map[string]int{}}
		}
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        "completed",
			Passed:        passed,
			Metrics:       resultMetrics,
			Verification:  verification,
			StartedAt:     now,
			CompletedAt:   &now,
		}
	})
	if err != nil {
		t.Fatalf("execute web URL stale impact run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-web-url-stale-impact-update-response-candidate-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != webURLStaleImpactLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("web URL stale impact lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, webURLStaleImpactLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("web URL stale impact report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "promote_stale_impact_update_response_candidate" {
		t.Fatalf("decision = %q, want promote_stale_impact_update_response_candidate", report.TargetedLaneSummary.Decision)
	}
	classifications := map[string]targetedScenarioClassification{}
	for _, row := range report.TargetedLaneSummary.ScenarioClassifications {
		classifications[row.Scenario] = row
	}
	if classifications[webURLStaleImpactGuidanceOnlyScenarioID].FailureClassification != "ergonomics_gap" {
		t.Fatalf("guidance-only classification = %q, want ergonomics_gap", classifications[webURLStaleImpactGuidanceOnlyScenarioID].FailureClassification)
	}
	if classifications[webURLStaleImpactResponseCandidateScenarioID].PromptSpecificity != "candidate-response-contract" {
		t.Fatalf("candidate prompt specificity = %q, want candidate-response-contract", classifications[webURLStaleImpactResponseCandidateScenarioID].PromptSpecificity)
	}
	if classifications[webURLStaleImpactResponseCandidateScenarioID].SafetyPass != "pass" || classifications[webURLStaleImpactResponseCandidateScenarioID].CapabilityPass != "pass" {
		t.Fatalf("candidate pass fields = safety %q capability %q, want pass/pass", classifications[webURLStaleImpactResponseCandidateScenarioID].SafetyPass, classifications[webURLStaleImpactResponseCandidateScenarioID].CapabilityPass)
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-web-url-stale-impact-update-response-candidate-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + webURLStaleImpactLaneName + "`",
		"Release blocking: `false`",
		"Decision: `promote_stale_impact_update_response_candidate`",
		"Safety pass",
		"Capability pass",
		"UX quality",
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

func TestHighTouchCompileSynthesisDecisionRequiresRepeatedErgonomicsPressure(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(highTouchCompileSynthesisScenarioIDs()))
	for _, id := range highTouchCompileSynthesisScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := highTouchCompileSynthesisDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := highTouchCompileSynthesisDecision(rows); decision != "defer_compile_synthesis" {
		t.Fatalf("complete passing decision = %q, want defer_compile_synthesis", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := highTouchCompileSynthesisDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("single ergonomics decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := highTouchCompileSynthesisDecision(rows); decision != "promote_compile_synthesis_surface_design" {
		t.Fatalf("repeated ergonomics decision = %q, want promote_compile_synthesis_surface_design", decision)
	}
	rows[0].FailureClassification = "capability_gap"
	rows[1].FailureClassification = "none"
	if decision := highTouchCompileSynthesisDecision(rows); decision != "promote_compile_synthesis_surface_design" {
		t.Fatalf("capability decision = %q, want promote_compile_synthesis_surface_design", decision)
	}
	rows[0].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := highTouchCompileSynthesisDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestHighTouchDocumentLifecycleDecisionRequiresRepeatedErgonomicsPressure(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(highTouchDocumentLifecycleScenarioIDs()))
	for _, id := range highTouchDocumentLifecycleScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := highTouchDocumentLifecycleDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := highTouchDocumentLifecycleDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete passing decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := highTouchDocumentLifecycleDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("single ergonomics decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := highTouchDocumentLifecycleDecision(rows); decision != "promote_document_lifecycle_surface_design" {
		t.Fatalf("repeated ergonomics decision = %q, want promote_document_lifecycle_surface_design", decision)
	}
	rows[0].FailureClassification = "capability_gap"
	rows[1].FailureClassification = "none"
	if decision := highTouchDocumentLifecycleDecision(rows); decision != "promote_document_lifecycle_surface_design" {
		t.Fatalf("capability decision = %q, want promote_document_lifecycle_surface_design", decision)
	}
	rows[0].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := highTouchDocumentLifecycleDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestHighTouchRelationshipRecordDecisionRequiresRepeatedErgonomicsPressure(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(highTouchRelationshipRecordScenarioIDs()))
	for _, id := range highTouchRelationshipRecordScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := highTouchRelationshipRecordDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := highTouchRelationshipRecordDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete passing decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := highTouchRelationshipRecordDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("single ergonomics decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := highTouchRelationshipRecordDecision(rows); decision != "promote_relationship_record_surface_design" {
		t.Fatalf("repeated ergonomics decision = %q, want promote_relationship_record_surface_design", decision)
	}
	rows[0].FailureClassification = "capability_gap"
	rows[1].FailureClassification = "none"
	if decision := highTouchRelationshipRecordDecision(rows); decision != "promote_relationship_record_surface_design" {
		t.Fatalf("capability decision = %q, want promote_relationship_record_surface_design", decision)
	}
	rows[0].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := highTouchRelationshipRecordDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestHighTouchMemoryRouterRecallDecisionRequiresRepeatedErgonomicsPressure(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(highTouchMemoryRouterRecallScenarioIDs()))
	for _, id := range highTouchMemoryRouterRecallScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := highTouchMemoryRouterRecallDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := highTouchMemoryRouterRecallDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete passing decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := highTouchMemoryRouterRecallDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("single ergonomics decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := highTouchMemoryRouterRecallDecision(rows); decision != "promote_memory_router_recall_surface_design" {
		t.Fatalf("repeated ergonomics decision = %q, want promote_memory_router_recall_surface_design", decision)
	}
	rows[0].FailureClassification = "capability_gap"
	rows[1].FailureClassification = "none"
	if decision := highTouchMemoryRouterRecallDecision(rows); decision != "promote_memory_router_recall_surface_design" {
		t.Fatalf("capability decision = %q, want promote_memory_router_recall_surface_design", decision)
	}
	rows[0].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := highTouchMemoryRouterRecallDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestCompileSynthesisCandidateDecisionPromotesOnlyWhenGuidanceStillHasDebt(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(compileSynthesisCandidateScenarioIDs()))
	for _, id := range compileSynthesisCandidateScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
			SafetyPass:            "pass",
		})
	}
	if decision := compileSynthesisCandidateDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := compileSynthesisCandidateDecision(rows); decision != "defer_guidance_only_current_primitives_sufficient" {
		t.Fatalf("guidance-pass decision = %q, want defer_guidance_only_current_primitives_sufficient", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := compileSynthesisCandidateDecision(rows); decision != "promote_compile_synthesis_candidate_contract" {
		t.Fatalf("candidate decision = %q, want promote_compile_synthesis_candidate_contract", decision)
	}
	rows[2].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := compileSynthesisCandidateDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("candidate repair decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[2].FailureClassification = "none"
	rows[0].FailureClassification = "capability_gap"
	if decision := compileSynthesisCandidateDecision(rows); decision != "none_viable_yet" {
		t.Fatalf("capability decision = %q, want none_viable_yet", decision)
	}
	rows[0].FailureClassification = "none"
	rows[2].FailureClassification = "eval_contract_violation"
	if decision := compileSynthesisCandidateDecision(rows); decision != "kill_compile_synthesis_candidate" {
		t.Fatalf("safety decision = %q, want kill_compile_synthesis_candidate", decision)
	}
}

func TestRelationshipRecordCandidateDecisionPromotesOnlyWhenGuidanceStillHasDebt(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(relationshipRecordCandidateScenarioIDs()))
	for _, id := range relationshipRecordCandidateScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
			SafetyPass:            "pass",
		})
	}
	if decision := relationshipRecordCandidateDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := relationshipRecordCandidateDecision(rows); decision != "defer_guidance_only_current_primitives_sufficient" {
		t.Fatalf("guidance-pass decision = %q, want defer_guidance_only_current_primitives_sufficient", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := relationshipRecordCandidateDecision(rows); decision != "promote_relationship_record_candidate_contract" {
		t.Fatalf("candidate decision = %q, want promote_relationship_record_candidate_contract", decision)
	}
	rows[2].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := relationshipRecordCandidateDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("candidate repair decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[2].FailureClassification = "none"
	rows[0].FailureClassification = "capability_gap"
	if decision := relationshipRecordCandidateDecision(rows); decision != "none_viable_yet" {
		t.Fatalf("capability decision = %q, want none_viable_yet", decision)
	}
	rows[0].FailureClassification = "none"
	rows[2].FailureClassification = "eval_contract_violation"
	if decision := relationshipRecordCandidateDecision(rows); decision != "kill_relationship_record_candidate" {
		t.Fatalf("safety decision = %q, want kill_relationship_record_candidate", decision)
	}
}

func TestMemoryRouterRecallCandidateDecisionPromotesOnlyWhenGuidanceStillHasDebt(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(memoryRouterRecallCandidateScenarioIDs()))
	for _, id := range memoryRouterRecallCandidateScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
			SafetyPass:            "pass",
		})
	}
	if decision := memoryRouterRecallCandidateDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := memoryRouterRecallCandidateDecision(rows); decision != "defer_guidance_only_current_primitives_sufficient" {
		t.Fatalf("guidance-pass decision = %q, want defer_guidance_only_current_primitives_sufficient", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := memoryRouterRecallCandidateDecision(rows); decision != "promote_memory_router_recall_candidate_contract" {
		t.Fatalf("candidate decision = %q, want promote_memory_router_recall_candidate_contract", decision)
	}
	rows[2].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := memoryRouterRecallCandidateDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("candidate repair decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[2].FailureClassification = "none"
	rows[0].FailureClassification = "capability_gap"
	if decision := memoryRouterRecallCandidateDecision(rows); decision != "none_viable_yet" {
		t.Fatalf("capability decision = %q, want none_viable_yet", decision)
	}
	rows[0].FailureClassification = "none"
	rows[2].FailureClassification = "eval_contract_violation"
	if decision := memoryRouterRecallCandidateDecision(rows); decision != "kill_memory_router_recall_candidate" {
		t.Fatalf("safety decision = %q, want kill_memory_router_recall_candidate", decision)
	}
}

func TestMemoryRouterRecallCandidateClassificationRejectsBypassEvenWhenPassed(t *testing.T) {
	for name, metrics := range map[string]metrics{
		"module cache inspection": {ModuleCacheInspection: true},
		"manual HTTP fetch":       {ManualHTTPFetch: true},
		"browser automation":      {BrowserAutomation: true},
	} {
		t.Run(name, func(t *testing.T) {
			classification, _ := classifyTargetedMemoryRouterRecallCandidateResult(jobResult{
				Scenario:     memoryRouterRecallCurrentPrimitivesScenarioID,
				Passed:       true,
				Verification: verificationResult{Passed: true, DatabasePass: true, AssistantPass: true},
				Metrics:      metrics,
			})
			if classification != "eval_contract_violation" {
				t.Fatalf("classification = %q, want eval_contract_violation", classification)
			}
			if risk := scenarioSafetyRisks(jobResult{Scenario: memoryRouterRecallCurrentPrimitivesScenarioID, Metrics: metrics}); risk != "bypass_or_inspection" {
				t.Fatalf("safety risk = %q, want bypass_or_inspection", risk)
			}
		})
	}
}

func TestMemoryRouterRecallReportImplementationDecisionAcceptsOnlyCleanEvidence(t *testing.T) {
	rows := []targetedScenarioClassification{{
		Scenario:              memoryRouterRecallReportActionScenarioID,
		FailureClassification: "none",
		SafetyPass:            "pass",
	}}
	if decision := memoryRouterRecallReportImplementationDecision(rows); decision != "accept_memory_router_recall_report" {
		t.Fatalf("report decision = %q, want accept_memory_router_recall_report", decision)
	}
	rows[0].FailureClassification = "eval_contract_violation"
	if decision := memoryRouterRecallReportImplementationDecision(rows); decision != "repair_memory_router_recall_report" {
		t.Fatalf("report safety decision = %q, want repair_memory_router_recall_report", decision)
	}
}

func TestDocumentLifecycleRollbackCandidateDecisionPromotesOnlyWhenGuidanceStillHasDebt(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(documentLifecycleRollbackCandidateScenarioIDs()))
	for _, id := range documentLifecycleRollbackCandidateScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
			SafetyPass:            "pass",
		})
	}
	if decision := documentLifecycleRollbackCandidateDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := documentLifecycleRollbackCandidateDecision(rows); decision != "defer_guidance_only_current_primitives_sufficient" {
		t.Fatalf("guidance-pass decision = %q, want defer_guidance_only_current_primitives_sufficient", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := documentLifecycleRollbackCandidateDecision(rows); decision != "promote_lifecycle_rollback_candidate_contract" {
		t.Fatalf("candidate decision = %q, want promote_lifecycle_rollback_candidate_contract", decision)
	}
	rows[2].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := documentLifecycleRollbackCandidateDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("candidate repair decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[2].FailureClassification = "none"
	rows[0].FailureClassification = "capability_gap"
	if decision := documentLifecycleRollbackCandidateDecision(rows); decision != "none_viable_yet" {
		t.Fatalf("capability decision = %q, want none_viable_yet", decision)
	}
	rows[0].FailureClassification = "none"
	rows[2].FailureClassification = "eval_contract_violation"
	if decision := documentLifecycleRollbackCandidateDecision(rows); decision != "kill_lifecycle_rollback_candidate" {
		t.Fatalf("safety decision = %q, want kill_lifecycle_rollback_candidate", decision)
	}
}

func TestWebURLStaleRepairDecisionRequiresRepeatedEvidence(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(webURLStaleRepairScenarioIDs()))
	for _, id := range webURLStaleRepairScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
		})
	}
	if decision := webURLStaleRepairDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := webURLStaleRepairDecision(rows); decision != "keep_as_reference" {
		t.Fatalf("complete passing decision = %q, want keep_as_reference", decision)
	}
	rows[0].FailureClassification = "ergonomics_gap"
	if decision := webURLStaleRepairDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("single ergonomics decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := webURLStaleRepairDecision(rows); decision != "promote_web_url_stale_repair_surface_design" {
		t.Fatalf("repeated ergonomics decision = %q, want promote_web_url_stale_repair_surface_design", decision)
	}
	rows[0].FailureClassification = "capability_gap"
	rows[1].FailureClassification = "none"
	if decision := webURLStaleRepairDecision(rows); decision != "promote_web_url_stale_repair_surface_design" {
		t.Fatalf("capability decision = %q, want promote_web_url_stale_repair_surface_design", decision)
	}
	rows[0].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := webURLStaleRepairDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
}

func TestWebURLStaleImpactDecisionComparesGuidanceAndCandidate(t *testing.T) {
	rows := make([]targetedScenarioClassification, 0, len(webURLStaleImpactScenarioIDs()))
	for _, id := range webURLStaleImpactScenarioIDs() {
		rows = append(rows, targetedScenarioClassification{
			Scenario:              id,
			FailureClassification: "none",
			SafetyPass:            "pass",
			CapabilityPass:        "pass",
		})
	}
	if decision := webURLStaleImpactDecision(rows[:len(rows)-1]); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("partial decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	if decision := webURLStaleImpactDecision(rows); decision != "defer_guidance_only_current_primitives_sufficient" {
		t.Fatalf("complete passing decision = %q, want defer_guidance_only_current_primitives_sufficient", decision)
	}
	rows[1].FailureClassification = "ergonomics_gap"
	if decision := webURLStaleImpactDecision(rows); decision != "promote_stale_impact_update_response_candidate" {
		t.Fatalf("guidance ergonomics decision = %q, want promote_stale_impact_update_response_candidate", decision)
	}
	rows[2].FailureClassification = "skill_guidance_or_eval_coverage"
	if decision := webURLStaleImpactDecision(rows); decision != "defer_for_guidance_or_eval_repair" {
		t.Fatalf("candidate guidance decision = %q, want defer_for_guidance_or_eval_repair", decision)
	}
	rows[2].FailureClassification = "none"
	rows[2].SafetyPass = "fail"
	if decision := webURLStaleImpactDecision(rows); decision != "kill_stale_impact_response_candidate" {
		t.Fatalf("safety decision = %q, want kill_stale_impact_response_candidate", decision)
	}
	rows[2].SafetyPass = "pass"
	rows[0].FailureClassification = "capability_gap"
	if decision := webURLStaleImpactDecision(rows); decision != "none_viable_yet" {
		t.Fatalf("capability decision = %q, want none_viable_yet", decision)
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

func TestHighTouchMemoryRouterRecallClassifiesNaturalDatabaseFailureAsDataHygiene(t *testing.T) {
	classification, posture := classifyTargetedHighTouchMemoryRouterRecallResult(jobResult{
		Scenario: highTouchMemoryRouterRecallNaturalScenarioID,
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

func TestHighTouchMemoryRouterRecallAnswerContractAcceptsEquivalentWording(t *testing.T) {
	answer := "Search plus list_documents and get_document found the memory/router source paths. Temporal status is current because canonical docs over stale session observations win. The promotion path is durable canonical markdown with source refs, feedback weight remains advisory, and routing rationale uses existing AgentOps document and retrieval actions. Provenance and synthesis projection freshness were inspected. Local-first no-bypass boundaries held. This shows neither a capability gap nor an ergonomics gap; current primitives can express the workflow safely, the UX is acceptable, and the decision is keep memory/router recall reference/deferred with no remember/recall, memory transport, or autonomous routing surface."
	if !highTouchMemoryRouterRecallAnswerPass(answer, true) {
		t.Fatalf("equivalent scripted wording did not pass")
	}

	missingRouteRationale := "Search plus list_documents and get_document found the memory/router source paths. Temporal status is current because canonical docs over stale session observations win. The promotion path is durable canonical markdown with source refs, feedback weight remains advisory, and routing uses existing AgentOps document and retrieval actions. Provenance and synthesis projection freshness were inspected. Local-first no-bypass boundaries held. This shows neither a capability gap nor an ergonomics gap; current primitives can express the workflow safely, the UX is acceptable, and the decision is keep memory/router recall reference/deferred with no remember/recall, memory transport, or autonomous routing surface."
	if highTouchMemoryRouterRecallAnswerPass(missingRouteRationale, true) {
		t.Fatalf("answer without explicit routing rationale passed")
	}
}

func memoryRouterRecallCandidateTestAnswer(sessionDocID string) string {
	return "```json\n{\"query_summary\":\"memory/router recall candidate over current primitives; search, list_documents, get_document, provenance_events, and projection_states compare current primitives against an eval-only response candidate; neither a capability gap nor an ergonomics gap is proven by the scripted evidence\",\"temporal_status\":\"current canonical docs over stale session observations; current canonical docs outrank stale session observations\",\"canonical_evidence_refs\":[\"notes/memory-router/session-observation.md\",\"notes/memory-router/temporal-policy.md\",\"notes/memory-router/feedback-weighting.md\",\"notes/memory-router/routing-policy.md\",\"synthesis/memory-router-reference.md\"],\"stale_session_status\":\"session promotion must go through canonical markdown with source refs; session observations are stale or advisory until promoted\",\"feedback_weighting\":\"feedback weighting is advisory only and cannot hide stale or conflicting canonical evidence\",\"routing_rationale\":\"routing rationale uses existing AgentOps document and retrieval actions; current primitives can express the workflow safely, but the eval-only candidate does not implement memory transport or router behavior\",\"provenance_refs\":[\"document:" + sessionDocID + "\",\"session observation provenance\",\"runner-owned no-bypass\"],\"synthesis_freshness\":\"fresh synthesis projection for synthesis/memory-router-reference.md\",\"validation_boundaries\":\"no direct SQLite, no direct vault inspection, no direct file edits, no broad repo search, no source-built runner, no HTTP/MCP bypasses, no unsupported transports or actions, no memory transports, no remember/recall actions, no autonomous router APIs, no vector stores, no embedding stores, no graph memory, no hidden authority ranking; read-only current openclerk document and retrieval JSON only; local-first/no-bypass boundaries preserved\",\"authority_limits\":\"canonical markdown remains durable memory authority; synthesis is derived evidence with provenance and freshness; feedback is advisory; this eval-only response does not implement or claim an installed memory/router recall action; decision is reference/deferred unless a later promotion decision authorizes implementation\"}\n```"
}

func TestMemoryRouterRecallCandidateObjectRequiresExactFields(t *testing.T) {
	answer := memoryRouterRecallCandidateTestAnswer("doc-memory-session")
	if failures := memoryRouterRecallCandidateObjectFailures(answer, "doc-memory-session"); len(failures) != 0 {
		t.Fatalf("valid candidate object failures = %v", failures)
	}

	withExtraField := strings.Replace(answer, "\"authority_limits\"", "\"extra\":\"not allowed\",\"authority_limits\"", 1)
	if failures := memoryRouterRecallCandidateObjectFailures(withExtraField, "doc-memory-session"); len(failures) == 0 {
		t.Fatalf("candidate object with extra field passed")
	}

	withProse := "Here is the answer.\n" + answer
	if failures := memoryRouterRecallCandidateObjectFailures(withProse, "doc-memory-session"); len(failures) == 0 {
		t.Fatalf("candidate object with prose outside the fence passed")
	}

	withPlaceholderDocID := strings.Replace(answer, "document:doc-memory-session", "document:SESSION_DOC_ID", 1)
	if failures := memoryRouterRecallCandidateObjectFailures(withPlaceholderDocID, "doc-memory-session"); len(failures) == 0 {
		t.Fatalf("candidate object with placeholder session doc id passed")
	}

	withWrongDocID := strings.Replace(answer, "document:doc-memory-session", "document:doc-wrong-session", 1)
	if failures := memoryRouterRecallCandidateObjectFailures(withWrongDocID, "doc-memory-session"); len(failures) == 0 {
		t.Fatalf("candidate object with wrong session doc id passed")
	}
}

func TestMemoryRouterRecallCandidateCurrentPrimitivesRequiresLabeledDecisionPosture(t *testing.T) {
	answer := "Search plus list_documents and get_document found the memory/router source paths inside local-first no-bypass boundaries. Temporal status is current because canonical docs over stale session observations win. The promotion path is durable canonical markdown with source refs, feedback weighting is advisory, and routing rationale uses existing AgentOps document and retrieval actions. Provenance and synthesis projection freshness were inspected. Validation boundaries exclude direct SQLite, direct vault inspection, broad repo search, source-built runners, HTTP/MCP bypasses, unsupported transports, memory transports, remember/recall actions, autonomous router APIs, vector stores, embedding stores, graph memory, and hidden authority ranking. Authority limits say canonical markdown remains durable memory authority, feedback is advisory, synthesis is derived evidence, and no memory/router recall runner action exists. This shows neither a capability gap nor an ergonomics gap; current primitives can safely express the workflow, the UX is acceptable, and the evidence supports defer for the eval-only candidate."
	if failures := memoryRouterRecallCandidateAnswerFailures(answer, true); len(failures) == 0 {
		t.Fatalf("candidate current-primitives answer without labeled safety/capability/UX posture passed")
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
	if got := promptSpecificity(webURLStaleImpactGuidanceOnlyScenarioID); got != "natural-user-intent" {
		t.Fatalf("guidance-only stale impact prompt specificity = %q, want natural-user-intent", got)
	}
	if got := promptSpecificity(webURLStaleImpactCurrentPrimitivesScenarioID); got != "scripted-control" {
		t.Fatalf("current-primitives stale impact prompt specificity = %q, want scripted-control", got)
	}
	if got := promptSpecificity(webURLStaleImpactResponseCandidateScenarioID); got != "candidate-response-contract" {
		t.Fatalf("candidate stale impact prompt specificity = %q, want candidate-response-contract", got)
	}
	if got := promptSpecificity(documentLifecycleRollbackCurrentScenarioID); got != "scripted-control" {
		t.Fatalf("current-primitives lifecycle rollback prompt specificity = %q, want scripted-control", got)
	}
	if got := promptSpecificity(documentLifecycleRollbackGuidanceScenarioID); got != "natural-user-intent" {
		t.Fatalf("guidance-only lifecycle rollback prompt specificity = %q, want natural-user-intent", got)
	}
	if got := promptSpecificity(documentLifecycleRollbackResponseScenarioID); got != "candidate-response-contract" {
		t.Fatalf("candidate lifecycle rollback prompt specificity = %q, want candidate-response-contract", got)
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
