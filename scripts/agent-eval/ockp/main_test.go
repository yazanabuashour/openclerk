package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yazanabuashour/openclerk/client/local"
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
		{Index: 2, Variant: "sdk-baseline", Scenario: scenario{ID: "third", Title: "Third"}},
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
		{Index: 1, Variant: "sdk-baseline", Scenario: scenario{ID: "bad", Title: "Bad"}},
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
	if results[1].Variant != "sdk-baseline" || results[1].Scenario != "bad" || results[1].Error != "boom" {
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
	for _, want := range []string{"Configured parallelism", "Cache mode", "Phase Timings", "<run-root>/production/create-note/turn-1/events.jsonl"} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
	if !strings.Contains(output.String(), "ockp-test.json") {
		t.Fatalf("stdout = %q", output.String())
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

	multi := scenario{ID: "multi", Turns: []scenarioTurn{{Prompt: "first"}, {Prompt: "second"}}}
	resumeArgs := codexArgsForTurn("codex", "run-root/production/multi/repo", "run-root/production/multi", multi, scenarioTurn{Prompt: "second"}, 2, "session-123", cache)
	if containsValue(resumeArgs, "--ephemeral") {
		t.Fatalf("resume args must not use --ephemeral: %v", resumeArgs)
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

func TestVariantInstructionsDistinguishProductionAndSDKBaseline(t *testing.T) {
	production, err := variantInstructions("production")
	if err != nil {
		t.Fatalf("production instructions: %v", err)
	}
	if !strings.Contains(production, "openclerk") || strings.Contains(production, "local.OpenClient") {
		t.Fatalf("production instructions = %s", production)
	}
	if !strings.Contains(production, "reject final-answer-only") {
		t.Fatalf("production instructions missing direct rejection guidance = %s", production)
	}

	baseline, err := variantInstructions("sdk-baseline")
	if err != nil {
		t.Fatalf("baseline instructions: %v", err)
	}
	if !strings.Contains(baseline, "local.OpenClient") || !strings.Contains(baseline, "Do not use `openclerk`") {
		t.Fatalf("baseline instructions = %s", baseline)
	}

	if _, err := variantInstructions("unknown"); err == nil {
		t.Fatal("expected unknown variant error")
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
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files /Users/y/.codex"}}`,
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
	if parsed.metrics.ToolCalls != 3 || parsed.metrics.CommandExecutions != 3 || parsed.metrics.AssistantCalls != 1 {
		t.Fatalf("metrics = %+v", parsed.metrics)
	}
	if !parsed.metrics.BroadRepoSearch {
		t.Fatalf("expected broad repo search metric")
	}
	for _, evidence := range parsed.metrics.BroadRepoSearchEvidence {
		if strings.Contains(evidence, "/Users/y") {
			t.Fatalf("evidence was not sanitized: %v", parsed.metrics.BroadRepoSearchEvidence)
		}
	}
	if parsed.metrics.NonCachedInputTokens == nil || *parsed.metrics.NonCachedInputTokens != 70 || parsed.metrics.OutputTokens == nil || *parsed.metrics.OutputTokens != 12 {
		t.Fatalf("token metrics = %+v", parsed.metrics)
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

func TestFinalAnswerOnlyAndTokenComparisonGates(t *testing.T) {
	prodTokens := 80
	baseTokens := 120
	results := []jobResult{
		comparisonResult(productionVariant, "missing-document-path-reject", true, 0, 0, 1, prodTokens),
		comparisonResult(baselineVariant, "missing-document-path-reject", true, 2, 2, 1, baseTokens),
		comparisonResult(productionVariant, "create-note", true, 2, 2, 1, prodTokens),
		comparisonResult(baselineVariant, "create-note", true, 3, 3, 1, baseTokens),
	}
	summary := buildCodeFirstSummary(results)
	if summary == nil {
		t.Fatal("missing summary")
	}
	criteria := map[string]bool{}
	for _, criterion := range summary.Criteria {
		criteria[criterion.Name] = criterion.Passed
	}
	for _, name := range []string{"validation_scenarios_are_final_answer_only", "non_cached_token_majority", "non_cached_token_total_less_than_or_equal_baseline"} {
		if !criteria[name] {
			t.Fatalf("%s failed in %+v", name, summary.Criteria)
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

	cliMCP := scenario{ID: "unsupported-cli-mcp-reject"}
	if result := verifyFinalAnswerOnly(cliMCP, "I can't bypass OpenClerk runner through the CLI or MCP for this routine task.", noTools); !result.Passed {
		t.Fatalf("CLI/MCP bypass rejection failed: %+v", result)
	}
}

func TestScenarioIDsIncludeADRProofObligations(t *testing.T) {
	ids := map[string]bool{}
	for _, id := range scenarioIDs() {
		ids[id] = true
	}
	for _, want := range []string{"answer-filing", "stale-synthesis-update", "promoted-record-vs-docs", "unsupported-cli-mcp-reject"} {
		if !ids[want] {
			t.Fatalf("scenarioIDs missing %q in %v", want, scenarioIDs())
		}
	}
}

func TestVerifyAnswerFilingRequiresFiledSourceLinkedDocument(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "answer-filing"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "answer-filing"}, 1, "notes/synthesis/filed-runner-answer.md", noTools)
	if err != nil {
		t.Fatalf("verify missing answer filing: %v", err)
	}
	if result.Passed {
		t.Fatalf("missing filed document passed: %+v", result)
	}
	cfg := local.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	body := "# Filed OpenClerk runner Answer\n\n## Summary\nSource: notes/sources/answer-filing-runner.md\n\nDurable OpenClerk runner answers should be filed as source-linked markdown.\n"
	if err := createSeedDocument(ctx, cfg, "notes/synthesis/filed-runner-answer.md", "Filed OpenClerk runner Answer", body); err != nil {
		t.Fatalf("create filed answer: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "answer-filing"}, 1, "Created notes/synthesis/filed-runner-answer.md.", noTools)
	if err != nil {
		t.Fatalf("verify answer filing: %v", err)
	}
	if !result.Passed {
		t.Fatalf("answer filing failed: %+v", result)
	}
}

func TestVerifyStaleSynthesisUpdateRequiresCurrentSourceAndNoDuplicate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "stale-synthesis-update"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "stale-synthesis-update"}, 1, "Updated notes/synthesis/runner-routing.md.", noTools)
	if err != nil {
		t.Fatalf("verify stale before update: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale synthesis passed before update: %+v", result)
	}
	replacement := "Current guidance: routine agents must use openclerk JSON runner.\n\nCurrent source: notes/sources/runner-current-runner.md\n\nSupersedes: notes/sources/runner-old-cli.md\n\nThis stale claim is superseded by current guidance."
	replaceSeedSection(t, ctx, paths, "notes/synthesis/runner-routing.md", "Summary", replacement)
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "stale-synthesis-update"}, 1, "Updated notes/synthesis/runner-routing.md with current guidance.", noTools)
	if err != nil {
		t.Fatalf("verify stale after update: %v", err)
	}
	if !result.Passed {
		t.Fatalf("updated stale synthesis failed: %+v", result)
	}
	cfg := local.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	if err := createSeedDocument(ctx, cfg, "notes/synthesis/runner-routing-current.md", "OpenClerk runner Routing Current", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "stale-synthesis-update"}, 1, "Updated notes/synthesis/runner-routing.md with current guidance.", noTools)
	if err != nil {
		t.Fatalf("verify stale duplicate: %v", err)
	}
	if result.Passed {
		t.Fatalf("duplicate synthesis passed: %+v", result)
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

func TestProductionOnlySummaryDoesNotBeatMissingBaseline(t *testing.T) {
	results := []jobResult{
		comparisonResult(productionVariant, "create-note", true, 1, 1, 1, 80),
	}
	summary := buildCodeFirstSummary(results)
	if summary == nil {
		t.Fatal("missing summary")
	}
	if summary.BeatsBaseline {
		t.Fatalf("production-only report should not beat missing baseline: %+v", summary)
	}
	if summary.Recommendation != "baseline_not_run_production_only_report" {
		t.Fatalf("recommendation = %q", summary.Recommendation)
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
	cfg := local.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
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

func containsValue(args []string, value string) bool {
	for _, arg := range args {
		if arg == value {
			return true
		}
	}
	return false
}
