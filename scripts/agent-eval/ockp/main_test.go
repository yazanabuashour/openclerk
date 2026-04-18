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

func TestParseRunConfigDefaultsParallelToFour(t *testing.T) {
	config, err := parseRunConfig(nil, &strings.Builder{})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if config.Parallel != defaultParallel {
		t.Fatalf("parallel = %d, want %d", config.Parallel, defaultParallel)
	}
}

func TestParseRunConfigRejectsInvalidParallel(t *testing.T) {
	_, err := parseRunConfig([]string{"--parallel", "0"}, &strings.Builder{})
	if err == nil {
		t.Fatal("expected invalid parallel error")
	}
}

func TestRunJobsPreservesDeterministicOrder(t *testing.T) {
	jobs := []evalJob{
		{Index: 0, Variant: "production", Scenario: scenario{ID: "first", Title: "First"}},
		{Index: 1, Variant: "production", Scenario: scenario{ID: "second", Title: "Second"}},
		{Index: 2, Variant: "sdk-baseline", Scenario: scenario{ID: "third", Title: "Third"}},
	}
	results := runJobs(context.Background(), runConfig{Parallel: 3}, jobs, func(_ context.Context, _ runConfig, job evalJob) jobResult {
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
		{Index: 1, Variant: "sdk-baseline", Scenario: scenario{ID: "bad", Title: "Bad"}},
	}
	results := runJobs(context.Background(), runConfig{Parallel: 2}, jobs, func(_ context.Context, _ runConfig, job evalJob) jobResult {
		result := jobResult{Variant: job.Variant, Scenario: job.Scenario.ID}
		if job.Scenario.ID == "bad" {
			result.Status = "failed"
			result.Error = "boom"
			return result
		}
		result.Status = "completed"
		return result
	})
	if results[1].Variant != "sdk-baseline" || results[1].Scenario != "bad" || results[1].Error != "boom" {
		t.Fatalf("error result = %+v", results[1])
	}
}

func TestExecuteRunWritesParallelMetadataAndReports(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:  2,
		Variant:   "production",
		Scenario:  "create-note,append-replace",
		RunRoot:   filepath.Join(t.TempDir(), "run"),
		ReportDir: reportDir,
		RepoRoot:  ".",
		CodexBin:  "codex",
	}
	var output strings.Builder
	err := executeRun(context.Background(), config, &output, func(_ context.Context, _ runConfig, job evalJob) jobResult {
		now := time.Now().UTC()
		return jobResult{
			Variant:                 job.Variant,
			Scenario:                job.Scenario.ID,
			ScenarioTitle:           job.Scenario.Title,
			Status:                  "completed",
			ToolCalls:               job.Index + 1,
			AssistantCalls:          1,
			WallSeconds:             0.25,
			RawLogArtifactReference: "<run-root>/" + job.Variant + "/" + job.Scenario.ID + "/events.jsonl",
			StartedAt:               now,
			CompletedAt:             &now,
		}
	})
	if err != nil {
		t.Fatalf("execute run: %v", err)
	}
	jsonPath := filepath.Join(reportDir, "ockp-latest.json")
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.ConfiguredParallelism != 2 || report.Metadata.HarnessElapsedSeconds <= 0 {
		t.Fatalf("metadata = %+v", report.Metadata)
	}
	if len(report.Results) != 2 || report.Results[0].Scenario != "create-note" || report.Results[1].Scenario != "append-replace" {
		t.Fatalf("results = %+v", report.Results)
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-latest.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	if !strings.Contains(string(markdown), "Configured parallelism") || !strings.Contains(string(markdown), "<run-root>/production/create-note/events.jsonl") {
		t.Fatalf("markdown report = %s", string(markdown))
	}
	if !strings.Contains(output.String(), "ockp-latest.json") {
		t.Fatalf("stdout = %q", output.String())
	}
}

func TestVariantInstructionsDistinguishProductionAndSDKBaseline(t *testing.T) {
	production, err := variantInstructions("production")
	if err != nil {
		t.Fatalf("production instructions: %v", err)
	}
	if !strings.Contains(production, "cmd/openclerk-agentops") || strings.Contains(production, "local.OpenClient") {
		t.Fatalf("production instructions = %s", production)
	}

	baseline, err := variantInstructions("sdk-baseline")
	if err != nil {
		t.Fatalf("baseline instructions: %v", err)
	}
	if !strings.Contains(baseline, "local.OpenClient") || !strings.Contains(baseline, "Do not use `cmd/openclerk-agentops`") {
		t.Fatalf("baseline instructions = %s", baseline)
	}

	if _, err := variantInstructions("unknown"); err == nil {
		t.Fatal("expected unknown variant error")
	}
}

func TestExtractTokenMetricsFromCodexJSONLines(t *testing.T) {
	log := strings.Join([]string{
		`{"type":"message","usage":{"input_tokens":100,"cached_input_tokens":30,"output_tokens":12}}`,
		`{"type":"message","response":{"usage":{"prompt_tokens":50,"prompt_tokens_details":{"cached_tokens":10},"completion_tokens":7}}}`,
		`not json`,
	}, "\n")
	metrics := extractTokenMetrics([]byte(log))
	if metrics.NonCacheInputTokens != 110 || metrics.OutputTokens != 19 {
		t.Fatalf("metrics = %+v", metrics)
	}
}
