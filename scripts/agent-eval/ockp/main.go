package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yazanabuashour/openclerk/client/local"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

const (
	defaultParallel   = 4
	modelName         = "gpt-5.4-mini"
	reasoningEffort   = "medium"
	productionVariant = "production"
	baselineVariant   = "sdk-baseline"
	cacheModeShared   = "shared"
	cacheModeIsolated = "isolated"
)

var prewarmCompilePackages = []string{"./cmd/openclerk", "./internal/runner"}

type runConfig struct {
	Parallel   int
	Variant    string
	Scenario   string
	RunRoot    string
	ReportDir  string
	ReportName string
	CodexBin   string
	RepoRoot   string
	CacheMode  string
}

type cacheConfig struct {
	Mode    string
	RunRoot string
}

type evalJob struct {
	Index    int
	Variant  string
	Scenario scenario
}

type scenario struct {
	ID     string         `json:"id"`
	Title  string         `json:"title"`
	Prompt string         `json:"prompt,omitempty"`
	Turns  []scenarioTurn `json:"turns,omitempty"`
}

type scenarioTurn struct {
	Prompt string `json:"prompt"`
}

type report struct {
	Metadata  reportMetadata    `json:"metadata"`
	Results   []jobResult       `json:"results"`
	CodeFirst *codeFirstSummary `json:"code_first,omitempty"`
}

type reportMetadata struct {
	GeneratedAt              time.Time    `json:"generated_at"`
	Model                    string       `json:"model"`
	ReasoningEffort          string       `json:"reasoning_effort"`
	Harness                  string       `json:"harness"`
	ConfiguredParallelism    int          `json:"configured_parallelism"`
	CacheMode                string       `json:"cache_mode"`
	CachePrewarmSeconds      float64      `json:"cache_prewarm_seconds,omitempty"`
	HarnessElapsedSeconds    float64      `json:"harness_elapsed_seconds"`
	EffectiveParallelSpeedup float64      `json:"effective_parallel_speedup,omitempty"`
	ParallelEfficiency       float64      `json:"parallel_efficiency,omitempty"`
	PhaseTotals              phaseTimings `json:"phase_totals"`
	RunRootArtifactReference string       `json:"run_root_artifact_reference"`
	RawLogPlaceholder        string       `json:"raw_log_placeholder"`
	Variants                 []string     `json:"variants"`
	Scenarios                []string     `json:"scenarios"`
	RawLogsCommitted         bool         `json:"raw_logs_committed"`
	RawLogsNote              string       `json:"raw_logs_note"`
}

type phaseTimings struct {
	PrepareRunDir  float64 `json:"prepare_run_dir_seconds,omitempty"`
	CopyRepo       float64 `json:"copy_repo_seconds,omitempty"`
	InstallVariant float64 `json:"install_variant_seconds,omitempty"`
	WarmCache      float64 `json:"warm_cache_seconds,omitempty"`
	SeedData       float64 `json:"seed_data_seconds,omitempty"`
	AgentRun       float64 `json:"agent_run_seconds,omitempty"`
	ParseMetrics   float64 `json:"parse_metrics_seconds,omitempty"`
	Verify         float64 `json:"verify_seconds,omitempty"`
	Total          float64 `json:"total_seconds,omitempty"`
}

type jobResult struct {
	Variant                 string             `json:"variant"`
	Scenario                string             `json:"scenario"`
	ScenarioTitle           string             `json:"scenario_title"`
	Passed                  bool               `json:"passed"`
	Status                  string             `json:"status"`
	Error                   string             `json:"error,omitempty"`
	ExitCode                int                `json:"exit_code"`
	WallSeconds             float64            `json:"wall_seconds"`
	PhaseTimings            phaseTimings       `json:"phase_timings"`
	Metrics                 metrics            `json:"metrics"`
	Verification            verificationResult `json:"verification"`
	Turns                   []turnResult       `json:"turns,omitempty"`
	PromptSummary           string             `json:"prompt_summary"`
	RawLogArtifactReference string             `json:"raw_log_artifact_reference"`
	StartedAt               time.Time          `json:"started_at"`
	CompletedAt             *time.Time         `json:"completed_at,omitempty"`
}

type turnResult struct {
	Index                   int                `json:"turn_index"`
	WallSeconds             float64            `json:"wall_seconds"`
	ExitCode                int                `json:"exit_code"`
	Metrics                 metrics            `json:"metrics"`
	Verification            verificationResult `json:"verification"`
	RawLogArtifactReference string             `json:"raw_log_artifact_reference"`
}

type metrics struct {
	AssistantCalls           int            `json:"assistant_calls"`
	ToolCalls                int            `json:"tool_calls"`
	CommandExecutions        int            `json:"command_executions"`
	FileInspectionCommands   int            `json:"file_inspection_commands"`
	GeneratedFileInspection  bool           `json:"generated_file_inspection"`
	ModuleCacheInspection    bool           `json:"module_cache_inspection"`
	BroadRepoSearch          bool           `json:"broad_repo_search"`
	DirectSQLiteAccess       bool           `json:"direct_sqlite_access"`
	LegacyRunnerUsage        bool           `json:"legacy_runner_usage"`
	GeneratedFileEvidence    []string       `json:"generated_file_evidence,omitempty"`
	ModuleCacheEvidence      []string       `json:"module_cache_evidence,omitempty"`
	BroadRepoSearchEvidence  []string       `json:"broad_repo_search_evidence,omitempty"`
	DirectSQLiteEvidence     []string       `json:"direct_sqlite_evidence,omitempty"`
	LegacyRunnerEvidence     []string       `json:"legacy_runner_evidence,omitempty"`
	UsageExposed             bool           `json:"usage_exposed"`
	InputTokens              *int           `json:"input_tokens,omitempty"`
	CachedInputTokens        *int           `json:"cached_input_tokens,omitempty"`
	NonCachedInputTokens     *int           `json:"non_cached_input_tokens,omitempty"`
	OutputTokens             *int           `json:"output_tokens,omitempty"`
	EventTypeCounts          map[string]int `json:"event_type_counts"`
	CommandMetricLimitations string         `json:"command_metric_limitations"`
}

type verificationResult struct {
	Passed        bool     `json:"passed"`
	DatabasePass  bool     `json:"database_pass"`
	AssistantPass bool     `json:"assistant_pass"`
	Details       string   `json:"details"`
	Documents     []string `json:"documents,omitempty"`
}

type codeFirstSummary struct {
	CandidateVariant string                   `json:"candidate_variant"`
	BaselineVariant  string                   `json:"baseline_variant"`
	BeatsBaseline    bool                     `json:"beats_baseline"`
	Recommendation   string                   `json:"recommendation"`
	Criteria         []codeFirstCriterion     `json:"criteria"`
	Entries          []codeFirstComparisonRow `json:"entries"`
}

type codeFirstCriterion struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
}

type codeFirstComparisonRow struct {
	Scenario       string `json:"scenario"`
	CandidatePass  bool   `json:"candidate_pass"`
	BaselinePass   bool   `json:"baseline_pass"`
	CandidateTools int    `json:"candidate_tools"`
	BaselineTools  int    `json:"baseline_tools"`
	ToolDelta      *int   `json:"tool_delta,omitempty"`
}

type jobRunner func(context.Context, runConfig, evalJob, cacheConfig) jobResult

type codexEvent struct {
	Type     string          `json:"type"`
	ThreadID string          `json:"thread_id"`
	Item     json.RawMessage `json:"item"`
	Usage    *usage          `json:"usage"`
}

type usage struct {
	InputTokens        int           `json:"input_tokens"`
	OutputTokens       int           `json:"output_tokens"`
	CachedInputTokens  int           `json:"cached_input_tokens"`
	InputTokensDetails *usageDetails `json:"input_tokens_details"`
	PromptTokens       int           `json:"prompt_tokens"`
	CompletionTokens   int           `json:"completion_tokens"`
	PromptDetails      *usageDetails `json:"prompt_tokens_details"`
}

type usageDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type parsedTurn struct {
	metrics      metrics
	finalMessage string
	sessionID    string
	parseError   error
	parseSeconds float64
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, codexJobRunner))
}

func run(args []string, stdout io.Writer, stderr io.Writer, runner jobRunner) int {
	if len(args) == 0 || args[0] != "run" {
		_, _ = fmt.Fprintln(stderr, "usage: ockp run [--parallel N] [--variant ids] [--scenario ids] [--run-root path] [--report-dir path] [--codex-bin path] [--cache-mode shared|isolated]")
		return 2
	}
	config, err := parseRunConfig(args[1:], stderr)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	if err := executeRun(context.Background(), config, stdout, runner); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func parseRunConfig(args []string, stderr io.Writer) (runConfig, error) {
	fs := flag.NewFlagSet("ockp run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	config := runConfig{CacheMode: cacheModeShared}
	fs.IntVar(&config.Parallel, "parallel", defaultParallel, "number of independent eval jobs to run concurrently")
	fs.StringVar(&config.Variant, "variant", "", "comma-separated variant ids")
	fs.StringVar(&config.Scenario, "scenario", "", "comma-separated scenario ids")
	fs.StringVar(&config.RunRoot, "run-root", "", "directory for isolated run artifacts")
	fs.StringVar(&config.ReportDir, "report-dir", filepath.Join("docs", "evals", "results"), "directory for reduced reports")
	fs.StringVar(&config.ReportName, "report-name", "ockp-latest", "base filename for reduced reports, without extension")
	fs.StringVar(&config.CodexBin, "codex-bin", "codex", "codex executable")
	fs.StringVar(&config.RepoRoot, "repo-root", ".", "repository root to copy for each job")
	fs.StringVar(&config.CacheMode, "cache-mode", config.CacheMode, "Go cache mode: shared or isolated")
	if err := fs.Parse(args); err != nil {
		return runConfig{}, err
	}
	if fs.NArg() != 0 {
		return runConfig{}, fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if config.Parallel < 1 {
		return runConfig{}, errors.New("--parallel must be at least 1")
	}
	if config.CacheMode != cacheModeShared && config.CacheMode != cacheModeIsolated {
		return runConfig{}, fmt.Errorf("--cache-mode must be %q or %q", cacheModeShared, cacheModeIsolated)
	}
	if config.RunRoot == "" {
		config.RunRoot = filepath.Join(os.TempDir(), fmt.Sprintf("openclerk-ockp-%d", time.Now().UnixNano()))
	}
	if strings.TrimSpace(config.ReportName) == "" {
		return runConfig{}, errors.New("--report-name must not be empty")
	}
	return config, nil
}

func executeRun(ctx context.Context, config runConfig, stdout io.Writer, runner jobRunner) error {
	start := time.Now()
	jobs, err := buildJobs(config)
	if err != nil {
		return err
	}
	cache := cacheConfig{Mode: config.CacheMode, RunRoot: config.RunRoot}
	cachePrewarmSeconds := 0.0
	if cache.Mode == cacheModeShared {
		cacheStart := time.Now()
		if err := prewarmSharedCache(config.RepoRoot, cache); err != nil {
			return fmt.Errorf("prewarm shared Go cache: %w", err)
		}
		cachePrewarmSeconds = roundSeconds(time.Since(cacheStart).Seconds())
	}
	results := runJobs(ctx, config, jobs, cache, runner)
	elapsed := roundSeconds(time.Since(start).Seconds())
	phaseTotals := aggregatePhaseTimings(results)
	effectiveSpeedup := 0.0
	parallelEfficiency := 0.0
	totalAgent := totalAgentWallSeconds(results)
	if elapsed > 0 {
		effectiveSpeedup = roundSeconds(totalAgent / elapsed)
	}
	if config.Parallel > 0 && effectiveSpeedup > 0 {
		parallelEfficiency = roundSeconds(effectiveSpeedup / float64(config.Parallel))
	}
	rep := report{
		Metadata: reportMetadata{
			GeneratedAt:              time.Now().UTC(),
			Model:                    modelName,
			ReasoningEffort:          reasoningEffort,
			Harness:                  "codex exec --json --full-auto from throwaway run directories; single-turn scenarios use --ephemeral and multi-turn scenarios resume one persisted eval session",
			ConfiguredParallelism:    config.Parallel,
			CacheMode:                cache.Mode,
			CachePrewarmSeconds:      cachePrewarmSeconds,
			HarnessElapsedSeconds:    elapsed,
			EffectiveParallelSpeedup: effectiveSpeedup,
			ParallelEfficiency:       parallelEfficiency,
			PhaseTotals:              phaseTotals,
			RunRootArtifactReference: "<run-root>",
			RawLogPlaceholder:        "<run-root>/<variant>/<scenario>/turn-N/events.jsonl",
			Variants:                 selectedVariants(config),
			Scenarios:                selectedScenarioIDs(config),
			RawLogsCommitted:         false,
			RawLogsNote:              "Raw Codex event logs remain under <run-root> and are not committed.",
		},
		Results:   results,
		CodeFirst: buildCodeFirstSummary(results),
	}
	if err := os.MkdirAll(config.ReportDir, 0o755); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}
	jsonPath := filepath.Join(config.ReportDir, config.ReportName+".json")
	markdownPath := filepath.Join(config.ReportDir, config.ReportName+".md")
	if err := writeJSONReport(jsonPath, rep); err != nil {
		return err
	}
	if err := writeMarkdownReport(markdownPath, rep); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stdout, "wrote %s and %s\n", filepath.ToSlash(jsonPath), filepath.ToSlash(markdownPath)); err != nil {
		return err
	}
	return nil
}

func buildJobs(config runConfig) ([]evalJob, error) {
	variants := selectedVariants(config)
	scenarios := selectedScenarios(config)
	if len(scenarios) == 0 {
		return nil, errors.New("no scenarios selected")
	}
	jobs := make([]evalJob, 0, len(variants)*len(scenarios))
	for _, variant := range variants {
		for _, scenario := range scenarios {
			jobs = append(jobs, evalJob{
				Index:    len(jobs),
				Variant:  variant,
				Scenario: scenario,
			})
		}
	}
	return jobs, nil
}

func runJobs(ctx context.Context, config runConfig, jobs []evalJob, cache cacheConfig, runner jobRunner) []jobResult {
	results := make([]jobResult, len(jobs))
	jobCh := make(chan evalJob)
	var wg sync.WaitGroup
	workers := min(config.Parallel, max(1, len(jobs)))
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				results[job.Index] = runner(ctx, config, job, cache)
			}
		}()
	}
	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)
	wg.Wait()
	return results
}

func codexJobRunner(ctx context.Context, config runConfig, job evalJob, cache cacheConfig) jobResult {
	start := time.Now()
	result := jobResult{
		Variant:       job.Variant,
		Scenario:      job.Scenario.ID,
		ScenarioTitle: job.Scenario.Title,
		Status:        "failed",
		StartedAt:     start.UTC(),
		PromptSummary: promptSummary(job.Scenario),
	}
	timings := phaseTimings{}
	jobDir := filepath.Join(config.RunRoot, job.Variant, job.Scenario.ID)
	repoDir := filepath.Join(jobDir, "repo")
	paths := scenarioPaths(repoDir)
	if err := timedPhase(&timings.PrepareRunDir, func() error { return prepareRunDir(jobDir, cache) }); err != nil {
		result.Error = err.Error()
		return result
	}
	if err := timedPhase(&timings.CopyRepo, func() error { return copyRepo(config.RepoRoot, repoDir) }); err != nil {
		result.Error = fmt.Sprintf("copy repo: %v", err)
		return result
	}
	if err := timedPhase(&timings.InstallVariant, func() error {
		if err := writeVariantInstructions(repoDir, job.Variant); err != nil {
			return err
		}
		return buildOpenClerkRunner(repoDir, jobDir, paths, cache)
	}); err != nil {
		result.Error = fmt.Sprintf("configure variant: %v", err)
		return result
	}
	if cache.Mode == cacheModeIsolated {
		if err := timedPhase(&timings.WarmCache, func() error { return warmGoModules(repoDir, jobDir, paths, cache) }); err != nil {
			result.Error = fmt.Sprintf("warm go modules: %v", err)
			return result
		}
	}
	if err := timedPhase(&timings.SeedData, func() error { return seedScenario(ctx, paths, job.Scenario) }); err != nil {
		result.Error = fmt.Sprintf("seed scenario: %v", err)
		return result
	}

	turns := scenarioTurns(job.Scenario)
	turnResults := make([]turnResult, 0, len(turns))
	sessionID := ""
	var runErr error
	for i, turn := range turns {
		turnIndex := i + 1
		turnResult, parsed, err := runScenarioTurn(ctx, config, repoDir, jobDir, paths, job, turn, turnIndex, sessionID, cache)
		timings.AgentRun += turnResult.WallSeconds
		timings.ParseMetrics += parsed.parseSeconds
		if parsed.parseError != nil {
			turnResult.Metrics.CommandMetricLimitations = fmt.Sprintf("failed to parse event log: %v", parsed.parseError)
		}
		verifyStart := time.Now()
		verification, verifyErr := verifyScenarioTurn(ctx, paths, job.Scenario, turnIndex, parsed.finalMessage, turnResult.Metrics)
		timings.Verify += roundSeconds(time.Since(verifyStart).Seconds())
		if verifyErr != nil {
			verification = verificationResult{Passed: false, Details: fmt.Sprintf("verification error: %v", verifyErr)}
		}
		turnResult.Verification = verification
		turnResults = append(turnResults, turnResult)
		if err != nil && runErr == nil {
			runErr = err
		}
		if verifyErr != nil && runErr == nil {
			runErr = verifyErr
		}
		if i == 0 && len(turns) > 1 {
			sessionID = parsed.sessionID
			if sessionID == "" && runErr == nil {
				runErr = errors.New("multi-turn first turn did not expose a thread id")
			}
		}
	}

	completed := time.Now().UTC()
	timings.Total = roundSeconds(time.Since(start).Seconds())
	verification := aggregateVerification(job.Scenario, turnResults)
	result.CompletedAt = &completed
	result.WallSeconds = roundSeconds(sumTurnWallSeconds(turnResults))
	result.PhaseTimings = timings.rounded()
	result.Metrics = aggregateMetrics(turnResults)
	result.Verification = verification
	result.Turns = turnResults
	result.ExitCode = aggregateExitCode(turnResults)
	if len(turnResults) > 0 {
		result.RawLogArtifactReference = turnResults[len(turnResults)-1].RawLogArtifactReference
	}
	result.Passed = runErr == nil && verification.Passed
	if result.Passed {
		result.Status = "completed"
	} else if runErr != nil {
		result.Error = runErr.Error()
	}
	_ = writeJSON(filepath.Join(jobDir, "run-summary.json"), result)
	return result
}

type evalPaths struct {
	DataDir      string
	DatabasePath string
	VaultRoot    string
	GoCache      string
	GoModCache   string
	Temp         string
}

func scenarioPaths(repoDir string) evalPaths {
	return evalPaths{
		DataDir:      filepath.Join(repoDir, ".openclerk-eval", "data"),
		DatabasePath: filepath.Join(repoDir, ".openclerk-eval", "openclerk.db"),
		VaultRoot:    filepath.Join(repoDir, ".openclerk-eval", "vault"),
	}
}

func evalPathsFor(runDir string, paths evalPaths, cache cacheConfig) evalPaths {
	out := paths
	out.Temp = filepath.Join(runDir, "tmp")
	if cache.Mode == cacheModeShared {
		out.GoCache = filepath.Join(cache.RunRoot, "shared-cache", "gocache")
		out.GoModCache = filepath.Join(cache.RunRoot, "shared-cache", "gomodcache")
	} else {
		out.GoCache = filepath.Join(runDir, "gocache")
		out.GoModCache = filepath.Join(runDir, "gomodcache")
	}
	return out
}

func runScenarioTurn(ctx context.Context, config runConfig, repoDir string, runDir string, paths evalPaths, job evalJob, turn scenarioTurn, turnIndex int, sessionID string, cache cacheConfig) (turnResult, parsedTurn, error) {
	turnDir := filepath.Join(runDir, fmt.Sprintf("turn-%d", turnIndex))
	if err := os.MkdirAll(turnDir, 0o755); err != nil {
		return turnResult{}, parsedTurn{}, err
	}
	eventsPath := filepath.Join(turnDir, "events.jsonl")
	stderrPath := filepath.Join(turnDir, "stderr.log")
	stdoutFile, err := os.Create(eventsPath)
	if err != nil {
		return turnResult{}, parsedTurn{}, err
	}
	defer func() { _ = stdoutFile.Close() }()
	stderrFile, err := os.Create(stderrPath)
	if err != nil {
		return turnResult{}, parsedTurn{}, err
	}
	defer func() { _ = stderrFile.Close() }()

	args := codexArgsForTurn(config.CodexBin, repoDir, runDir, job.Scenario, turn, turnIndex, sessionID, cache)
	cmdCtx, cancel := context.WithTimeout(ctx, 7*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, args[0], args[1:]...)
	cmd.Dir = repoDir
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	cmd.Stdin = strings.NewReader("")
	cmd.Env = evalEnv(runDir, paths, cache)

	start := time.Now()
	err = cmd.Run()
	wallSeconds := roundSeconds(time.Since(start).Seconds())
	exitCode := commandExitCode(err)
	if cmdCtx.Err() == context.DeadlineExceeded {
		exitCode = -1
		err = cmdCtx.Err()
	}
	parseStart := time.Now()
	parsedMetrics, parseErr := parseMetrics(eventsPath)
	parseSeconds := roundSeconds(time.Since(parseStart).Seconds())
	parsed := parsedTurn{
		metrics:      parsedMetrics.metrics,
		finalMessage: parsedMetrics.finalMessage,
		sessionID:    parsedMetrics.sessionID,
		parseError:   parseErr,
		parseSeconds: parseSeconds,
	}
	result := turnResult{
		Index:                   turnIndex,
		WallSeconds:             wallSeconds,
		ExitCode:                exitCode,
		Metrics:                 parsedMetrics.metrics,
		RawLogArtifactReference: fmt.Sprintf("<run-root>/%s/%s/turn-%d/events.jsonl", job.Variant, job.Scenario.ID, turnIndex),
	}
	return result, parsed, err
}

func codexArgsForTurn(codexBin string, repoDir string, runDir string, sc scenario, turn scenarioTurn, turnIndex int, sessionID string, cache cacheConfig) []string {
	baseConfig := []string{
		"-m", modelName,
		"-c", fmt.Sprintf("model_reasoning_effort=%q", reasoningEffort),
		"-c", "shell_environment_policy.inherit=all",
	}
	writableRoots := codexWritableRoots(runDir, cache)
	if len(scenarioTurns(sc)) == 1 {
		args := []string{codexBin, "exec", "--json", "--ephemeral", "--full-auto", "--skip-git-repo-check", "-C", repoDir}
		args = appendAddDirs(args, writableRoots)
		args = append(args, baseConfig...)
		return append(args, turn.Prompt)
	}
	if turnIndex == 1 {
		args := []string{codexBin, "exec", "--json", "--full-auto", "--skip-git-repo-check", "-C", repoDir}
		args = appendAddDirs(args, writableRoots)
		args = append(args, baseConfig...)
		return append(args, turn.Prompt)
	}
	args := []string{codexBin, "exec", "-C", repoDir}
	args = appendAddDirs(args, writableRoots)
	args = append(args, "resume", "--json", "--full-auto", "--skip-git-repo-check")
	args = append(args, baseConfig...)
	args = append(args, sessionID, turn.Prompt)
	return args
}

func codexWritableRoots(runDir string, cache cacheConfig) []string {
	roots := []string{runDir}
	if cache.Mode == cacheModeShared {
		roots = append(roots, filepath.Join(cache.RunRoot, "shared-cache"))
	}
	return roots
}

func appendAddDirs(args []string, roots []string) []string {
	for _, root := range roots {
		args = append(args, "--add-dir", root)
	}
	return args
}

func evalEnv(runDir string, paths evalPaths, cache cacheConfig) []string {
	effective := evalPathsFor(runDir, paths, cache)
	env := os.Environ()
	pathValue := filepath.Join(runDir, "bin")
	if existing := os.Getenv("PATH"); existing != "" {
		pathValue += string(os.PathListSeparator) + existing
	}
	env = append(env,
		"OPENCLERK_DATA_DIR="+effective.DataDir,
		"OPENCLERK_DATABASE_PATH="+effective.DatabasePath,
		"OPENCLERK_VAULT_ROOT="+effective.VaultRoot,
		"GOCACHE="+effective.GoCache,
		"GOMODCACHE="+effective.GoModCache,
		"TMPDIR="+effective.Temp,
		"PATH="+pathValue,
	)
	return env
}

func prepareRunDir(runDir string, cache cacheConfig) error {
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}
	paths := evalPathsFor(runDir, evalPaths{}, cache)
	return os.MkdirAll(paths.Temp, 0o755)
}

func warmGoModules(repoDir string, runDir string, paths evalPaths, cache cacheConfig) error {
	effective := evalPathsFor(runDir, paths, cache)
	for _, dir := range []string{effective.GoCache, effective.GoModCache, effective.Temp} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = repoDir
	cmd.Env = evalEnv(runDir, paths, cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func prewarmSharedCache(repoRoot string, cache cacheConfig) error {
	paths := evalPathsFor(filepath.Join(cache.RunRoot, "shared-cache"), evalPaths{
		DataDir:      filepath.Join(cache.RunRoot, "shared-cache", "data"),
		DatabasePath: filepath.Join(cache.RunRoot, "shared-cache", "prewarm.db"),
		VaultRoot:    filepath.Join(cache.RunRoot, "shared-cache", "vault"),
	}, cache)
	for _, dir := range []string{paths.GoCache, paths.GoModCache, paths.Temp, paths.DataDir, paths.VaultRoot} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if err := warmGoModules(repoRoot, filepath.Join(cache.RunRoot, "shared-cache"), paths, cache); err != nil {
		return err
	}
	cmd := exec.Command("go", prewarmCompileArgs()...)
	cmd.Dir = repoRoot
	cmd.Env = evalEnv(filepath.Join(cache.RunRoot, "shared-cache"), paths, cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func prewarmCompileArgs() []string {
	args := []string{"test", "-run", "^$"}
	return append(args, prewarmCompilePackages...)
}

func buildOpenClerkRunner(repoDir string, runDir string, paths evalPaths, cache cacheConfig) error {
	binDir := filepath.Join(runDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(binDir, "openclerk"), "./cmd/openclerk")
	cmd.Dir = repoDir
	cmd.Env = evalEnv(runDir, paths, cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func seedScenario(ctx context.Context, paths evalPaths, sc scenario) error {
	cfg := local.Config{
		DataDir:      paths.DataDir,
		DatabasePath: paths.DatabasePath,
		VaultRoot:    paths.VaultRoot,
	}
	switch sc.ID {
	case "search-synthesis", "mixed-synthesis-records", "mt-source-then-synthesis":
		if err := createSeedDocument(ctx, cfg, "notes/sources/openclerk-runner.md", "OpenClerk Runner Source", "The OpenClerk runner uses JSON requests for OpenClerk knowledge tasks.\n\nIt preserves source refs for synthesis pages."); err != nil {
			return err
		}
	case "answer-filing":
		if err := createSeedDocument(ctx, cfg, "notes/sources/answer-filing-runner.md", "OpenClerk runner Answer Filing Source", "The OpenClerk runner JSON runner is the production path for reusable OpenClerk knowledge tasks.\n\nDurable OpenClerk runner answers should be filed as source-linked markdown."); err != nil {
			return err
		}
	case "stale-synthesis-update":
		if err := createSeedDocument(ctx, cfg, "notes/sources/runner-old-cli.md", "Old OpenClerk runner Routing Source", "Older guidance said routine agents may bypass OpenClerk runner through a temporary CLI workaround."); err != nil {
			return err
		}
		if err := createSeedDocument(ctx, cfg, "notes/sources/runner-current-runner.md", "Current OpenClerk runner Routing Source", "Current guidance says routine agents must use openclerk JSON runner for OpenClerk knowledge tasks."); err != nil {
			return err
		}
		body := "# OpenClerk runner Routing\n\n## Summary\nStale claim: routine agents may bypass OpenClerk runner through a temporary CLI workaround.\n\n## Sources\n- notes/sources/runner-old-cli.md\n"
		if err := createSeedDocument(ctx, cfg, "notes/synthesis/runner-routing.md", "OpenClerk runner Routing", body); err != nil {
			return err
		}
	case "append-replace":
		if err := createSeedDocument(ctx, cfg, "notes/projects/openclerk-runner.md", "OpenClerk Runner", "## Context\nExisting context stays intact."); err != nil {
			return err
		}
	case "records-provenance":
		if err := createSeedDocument(ctx, cfg, "records/services/openclerk-runner.md", "OpenClerk runner", recordBody("openclerk-runner", "service", "OpenClerk runner")); err != nil {
			return err
		}
	case "promoted-record-vs-docs":
		if err := createSeedDocument(ctx, cfg, "notes/reference/runner-service.md", "OpenClerk runner Service Reference", "# OpenClerk runner Service Reference\n\n## Summary\nPlain docs evidence says OpenClerk runner is the production service for routine knowledge tasks.\n\n## Details\nPlain docs evidence is narrative and searchable.\n"); err != nil {
			return err
		}
		body := strings.TrimSpace(`---
service_id: openclerk-runner
service_name: OpenClerk runner
service_status: active
service_owner: runner
service_interface: JSON runner
---

# OpenClerk runner

## Facts
- production_path: true
`)
		if err := createSeedDocument(ctx, cfg, "records/services/openclerk-runner.md", "OpenClerk runner", body); err != nil {
			return err
		}
	case "duplicate-path-reject":
		if err := createSeedDocument(ctx, cfg, "notes/projects/duplicate.md", "Duplicate Source", "This canonical path already exists."); err != nil {
			return err
		}
	}
	return nil
}

func createSeedDocument(ctx context.Context, cfg local.Config, path, title, body string) error {
	result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  path,
			Title: title,
			Body:  body,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return err
	}
	if result.Rejected {
		return errors.New(result.RejectionReason)
	}
	return nil
}

func recordBody(entityID, entityType, name string) string {
	return strings.TrimSpace(fmt.Sprintf(`---
entity_id: %s
entity_type: %s
entity_name: %s
---

# %s

## Facts
- status: active
- owner: runner
`, entityID, entityType, name, name))
}

func verifyScenarioTurn(ctx context.Context, paths evalPaths, sc scenario, turnIndex int, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	if isFinalAnswerOnlyValidationScenario(sc.ID) {
		return verifyFinalAnswerOnly(sc, finalMessage, turnMetrics), nil
	}
	if isMultiTurnScenario(sc) && turnIndex == 1 {
		switch sc.ID {
		case "mt-source-then-synthesis":
			return verifyDocuments(ctx, paths, []string{"notes/sources/mt-runner.md"}, finalMessage)
		case "mt-incomplete-then-create":
			return verifyNoDocument(ctx, paths, "notes/projects/mt-complete.md", "first turn should ask for missing document details"), nil
		}
	}
	switch sc.ID {
	case "create-note":
		return verifyDocuments(ctx, paths, []string{"notes/projects/openclerk-runner.md"}, finalMessage)
	case "search-synthesis":
		return verifyDocuments(ctx, paths, []string{"notes/synthesis/openclerk-runner.md"}, finalMessage)
	case "answer-filing":
		return verifyAnswerFiling(ctx, paths, finalMessage)
	case "stale-synthesis-update":
		return verifyStaleSynthesisUpdate(ctx, paths, finalMessage)
	case "append-replace":
		return verifyDocumentContains(ctx, paths, "notes/projects/openclerk-runner.md", []string{"Existing context stays intact", "Use the JSON runner"}, []string{"temporary CLI workaround"})
	case "records-provenance":
		return verifyRecordsAndProvenance(ctx, paths, finalMessage)
	case "promoted-record-vs-docs":
		return verifyPromotedRecordVsDocs(ctx, paths, finalMessage, turnMetrics)
	case "duplicate-path-reject":
		return verifyDuplicatePathReject(ctx, paths, finalMessage)
	case "mixed-synthesis-records":
		return verifyDocuments(ctx, paths, []string{"notes/synthesis/openclerk-runner-with-records.md"}, finalMessage)
	case "mt-source-then-synthesis":
		return verifyDocuments(ctx, paths, []string{"notes/sources/mt-runner.md", "notes/synthesis/mt-runner.md"}, finalMessage)
	case "mt-incomplete-then-create":
		return verifyDocuments(ctx, paths, []string{"notes/projects/mt-complete.md"}, finalMessage)
	default:
		return verificationResult{Passed: true, DatabasePass: true, AssistantPass: true, Details: "no scenario-specific verifier"}, nil
	}
}

func verifyFinalAnswerOnly(sc scenario, finalMessage string, turnMetrics metrics) verificationResult {
	answerPass := isValidationRejection(sc.ID, finalMessage)
	metricsPass := turnMetrics.ToolCalls == 0 && turnMetrics.CommandExecutions == 0 && turnMetrics.AssistantCalls <= 1
	failures := []string{}
	if !answerPass {
		failures = append(failures, "answer did not reject the invalid request")
	}
	if !metricsPass {
		failures = append(failures, fmt.Sprintf("expected no tools and at most one assistant answer, got tools=%d commands=%d assistant=%d", turnMetrics.ToolCalls, turnMetrics.CommandExecutions, turnMetrics.AssistantCalls))
	}
	return verificationResult{
		Passed:        answerPass && metricsPass,
		DatabasePass:  metricsPass,
		AssistantPass: answerPass,
		Details:       missingDetails(failures),
	}
}

func isValidationRejection(scenarioID string, message string) bool {
	lower := normalizeValidationMessage(message)
	if lower == "" {
		return false
	}
	switch scenarioID {
	case "missing-document-path-reject":
		return containsAny(lower, []string{"missing", "required", "need", "provide"}) && strings.Contains(lower, "path")
	case "negative-limit-reject":
		return containsAny(lower, []string{"negative", "invalid", "non-negative", "positive"}) && strings.Contains(lower, "limit")
	case "unsupported-lower-level-reject":
		return containsAny(lower, []string{"unsupported", "does not support", "cannot bypass", "can't bypass", "must use runner", "do not bypass", "use runner", "cannot do that", "can't do that", "cannot comply", "can't comply", "cannot fulfill", "can't fulfill"})
	case "unsupported-cli-mcp-reject":
		return containsAny(lower, []string{"unsupported", "cannot bypass", "cannot help bypass", "can't bypass", "can't help bypass", "do not bypass", "must use runner", "use runner"}) &&
			containsAny(lower, []string{"cli", "mcp", "runner"})
	default:
		return false
	}
}

func normalizeValidationMessage(message string) string {
	normalized := strings.NewReplacer(
		"\u2018", "'",
		"\u2019", "'",
		"\u02bc", "'",
	).Replace(message)
	return strings.ToLower(strings.TrimSpace(normalized))
}

func verifyNoDocument(ctx context.Context, paths evalPaths, docPath string, detail string) verificationResult {
	cfg := local.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docPath, Limit: 5},
	})
	if err != nil {
		return verificationResult{Passed: false, Details: err.Error()}
	}
	for _, doc := range list.Documents {
		if doc.Path == docPath {
			return verificationResult{Passed: false, DatabasePass: false, Details: detail}
		}
	}
	return verificationResult{Passed: true, DatabasePass: true, AssistantPass: true, Details: detail}
}

func verifyDocuments(ctx context.Context, paths evalPaths, wanted []string, finalMessage string) (verificationResult, error) {
	cfg := local.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 100},
	})
	if err != nil {
		return verificationResult{}, err
	}
	present := map[string]bool{}
	for _, doc := range list.Documents {
		present[doc.Path] = true
	}
	missing := []string{}
	for _, path := range wanted {
		if !present[path] {
			missing = append(missing, path)
		}
	}
	assistantPass := strings.TrimSpace(finalMessage) != ""
	return verificationResult{
		Passed:        len(missing) == 0 && assistantPass,
		DatabasePass:  len(missing) == 0,
		AssistantPass: assistantPass,
		Details:       missingDetails(missing),
		Documents:     wanted,
	}, nil
}

func verifyAnswerFiling(ctx context.Context, paths evalPaths, finalMessage string) (verificationResult, error) {
	docPath := "notes/synthesis/filed-runner-answer.md"
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+docPath)
	}
	failures = append(failures, missingRequired(body, []string{
		"notes/sources/answer-filing-runner.md",
		"Durable OpenClerk runner answers should be filed as source-linked markdown",
	})...)
	assistantPass := messageContainsAll(finalMessage, []string{docPath})
	if !assistantPass {
		failures = append(failures, "final answer did not mention "+docPath)
	}
	databasePass := found && len(missingRequired(body, []string{
		"notes/sources/answer-filing-runner.md",
		"Durable OpenClerk runner answers should be filed as source-linked markdown",
	})) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}

func verifyStaleSynthesisUpdate(ctx context.Context, paths evalPaths, finalMessage string) (verificationResult, error) {
	docPath := "notes/synthesis/runner-routing.md"
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	createdCurrent, err := exactDocumentCount(ctx, paths, "notes/synthesis/runner-routing-current.md")
	if err != nil {
		return verificationResult{}, err
	}
	createdUpdated, err := exactDocumentCount(ctx, paths, "notes/synthesis/runner-routing-updated.md")
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+docPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", docPath, exactCount))
	}
	if createdCurrent != 0 || createdUpdated != 0 {
		failures = append(failures, "created duplicate synthesis path")
	}
	required := []string{
		"Current guidance: routine agents must use openclerk JSON runner",
		"Current source: notes/sources/runner-current-runner.md",
		"Supersedes: notes/sources/runner-old-cli.md",
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, presentForbidden(body, []string{"may bypass OpenClerk runner through a temporary CLI workaround"})...)
	if !containsAny(strings.ToLower(body), []string{"stale", "supersedes", "superseded", "contradiction", "current guidance"}) {
		failures = append(failures, "missing stale or supersession language")
	}
	assistantPass := messageContainsAll(finalMessage, []string{docPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "current", "supersedes", "stale"})
	if !assistantPass {
		failures = append(failures, "final answer did not describe the synthesis update")
	}
	databasePass := found && exactCount == 1 && createdCurrent == 0 && createdUpdated == 0 &&
		len(missingRequired(body, required)) == 0 &&
		len(presentForbidden(body, []string{"may bypass OpenClerk runner through a temporary CLI workaround"})) == 0 &&
		containsAny(strings.ToLower(body), []string{"stale", "supersedes", "superseded", "contradiction", "current guidance"})
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}

func verifyDuplicatePathReject(ctx context.Context, paths evalPaths, finalMessage string) (verificationResult, error) {
	bodyCheck, err := verifyDocumentContains(ctx, paths, "notes/projects/duplicate.md", []string{"This canonical path already exists."}, []string{"overwritten"})
	if err != nil {
		return verificationResult{}, err
	}
	answerPass := isDuplicateRejection(finalMessage)
	failures := []string{}
	if !bodyCheck.DatabasePass {
		failures = append(failures, bodyCheck.Details)
	}
	if !answerPass {
		failures = append(failures, "answer did not report the duplicate path failure")
	}
	return verificationResult{
		Passed:        bodyCheck.DatabasePass && answerPass,
		DatabasePass:  bodyCheck.DatabasePass,
		AssistantPass: answerPass,
		Details:       missingDetails(failures),
		Documents:     []string{"notes/projects/duplicate.md"},
	}, nil
}

func isDuplicateRejection(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return false
	}
	if strings.Contains(lower, "already exists") || strings.Contains(lower, "duplicate") {
		return true
	}
	return strings.Contains(lower, "exists") && containsAny(lower, []string{"cannot", "can't", "failed", "not overwrite", "won't overwrite", "did not overwrite"})
}

func verifyPromotedRecordVsDocs(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := local.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: "Plain docs evidence", PathPrefix: "notes/reference/", Limit: 5},
	})
	if err != nil {
		return verificationResult{}, err
	}
	services, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionServicesLookup,
		Services: runner.ServiceLookupOptions{
			Text:      "OpenClerk runner",
			Interface: "JSON runner",
			Limit:     5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "services",
			RefKind:    "service",
			RefID:      "openclerk-runner",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	events, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "service",
			RefID:   "openclerk-runner",
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasPlainDoc := false
	if search.Search != nil {
		for _, hit := range search.Search.Hits {
			if hit.DocID != "" && hit.Title != "" && containsAny(strings.ToLower(hit.Snippet), []string{"plain docs evidence", "production service"}) {
				hasPlainDoc = true
				break
			}
		}
	}
	hasService := false
	if services.Services != nil {
		for _, service := range services.Services.Services {
			if service.ServiceID != "openclerk-runner" {
				continue
			}
			if service.Interface == "JSON runner" && len(service.Citations) > 0 {
				hasService = true
				break
			}
		}
	}
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) == 1 &&
		projections.Projections.Projections[0].Freshness == "fresh"
	hasProvenance := events.Provenance != nil && len(events.Provenance.Events) > 0
	assistantPass := messageContainsAny(finalMessage, []string{"services lookup", "services_lookup", "service registry"}) &&
		messageContainsAny(finalMessage, []string{"plain docs", "plain doc", "search"}) &&
		messageContainsAny(finalMessage, []string{"json runner", "runner"})
	activityPass := turnMetrics.ToolCalls >= 2 && turnMetrics.CommandExecutions >= 2
	failures := []string{}
	if !hasPlainDoc {
		failures = append(failures, "plain docs search evidence missing")
	}
	if !hasService {
		failures = append(failures, "services lookup evidence missing")
	}
	if !hasProjection {
		failures = append(failures, "services projection state missing")
	}
	if !hasProvenance {
		failures = append(failures, "services provenance missing")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not compare services lookup with plain docs")
	}
	if !activityPass {
		failures = append(failures, fmt.Sprintf("expected at least two agent operations for search and services lookup, got tools=%d commands=%d", turnMetrics.ToolCalls, turnMetrics.CommandExecutions))
	}
	return verificationResult{
		Passed:        hasPlainDoc && hasService && hasProjection && hasProvenance && assistantPass && activityPass,
		DatabasePass:  hasPlainDoc && hasService && hasProjection && hasProvenance,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{"notes/reference/runner-service.md", "records/services/openclerk-runner.md"},
	}, nil
}

func verifyDocumentContains(ctx context.Context, paths evalPaths, docPath string, required []string, forbidden []string) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	if !found {
		return verificationResult{Passed: false, DatabasePass: false, Details: "missing " + docPath}, nil
	}
	failures := missingRequired(body, required)
	failures = append(failures, presentForbidden(body, forbidden)...)
	return verificationResult{
		Passed:        len(failures) == 0,
		DatabasePass:  len(failures) == 0,
		AssistantPass: true,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}

func documentBodyByPath(ctx context.Context, paths evalPaths, docPath string) (string, bool, error) {
	docID, found, err := documentIDByPath(ctx, paths, docPath)
	if err != nil || !found {
		return "", found, err
	}
	cfg := local.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	got, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{Action: runner.DocumentTaskActionGet, DocID: docID})
	if err != nil {
		return "", false, err
	}
	if got.Document != nil {
		return got.Document.Body, true, nil
	}
	return "", false, nil
}

func documentIDByPath(ctx context.Context, paths evalPaths, docPath string) (string, bool, error) {
	cfg := local.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docPath, Limit: 100},
	})
	if err != nil {
		return "", false, err
	}
	for _, doc := range list.Documents {
		if doc.Path == docPath {
			return doc.DocID, true, nil
		}
	}
	return "", false, nil
}

func exactDocumentCount(ctx context.Context, paths evalPaths, docPath string) (int, error) {
	cfg := local.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docPath, Limit: 100},
	})
	if err != nil {
		return 0, err
	}
	count := 0
	for _, doc := range list.Documents {
		if doc.Path == docPath {
			count++
		}
	}
	return count, nil
}

func missingRequired(body string, required []string) []string {
	failures := []string{}
	for _, value := range required {
		if !strings.Contains(body, value) {
			failures = append(failures, "missing "+value)
		}
	}
	return failures
}

func presentForbidden(body string, forbidden []string) []string {
	failures := []string{}
	for _, value := range forbidden {
		if strings.Contains(body, value) {
			failures = append(failures, "unexpected "+value)
		}
	}
	return failures
}

func messageContainsAll(message string, values []string) bool {
	lower := normalizeValidationMessage(message)
	for _, value := range values {
		if !strings.Contains(lower, strings.ToLower(value)) {
			return false
		}
	}
	return true
}

func messageContainsAny(message string, values []string) bool {
	return containsAny(normalizeValidationMessage(message), lowerStrings(values))
}

func lowerStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, strings.ToLower(value))
	}
	return out
}

func verifyRecordsAndProvenance(ctx context.Context, paths evalPaths, finalMessage string) (verificationResult, error) {
	cfg := local.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	records, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:  runner.RetrievalTaskActionRecordsLookup,
		Records: runner.RecordLookupOptions{Text: "OpenClerk runner", Limit: 5},
	})
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasRecord := records.Records != nil && len(records.Records.Entities) > 0
	hasProvenance := provenance.Provenance != nil && len(provenance.Provenance.Events) > 0
	assistantPass := strings.TrimSpace(finalMessage) != ""
	return verificationResult{
		Passed:        hasRecord && hasProvenance && assistantPass,
		DatabasePass:  hasRecord && hasProvenance,
		AssistantPass: assistantPass,
		Details:       fmt.Sprintf("records=%t provenance=%t", hasRecord, hasProvenance),
	}, nil
}

func missingDetails(values []string) string {
	if len(values) == 0 {
		return "ok"
	}
	return strings.Join(values, "; ")
}

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func parseMetrics(eventsPath string) (parsedTurn, error) {
	file, err := os.Open(eventsPath)
	if err != nil {
		return parsedTurn{metrics: emptyMetrics()}, err
	}
	defer func() { _ = file.Close() }()
	out := parsedTurn{metrics: emptyMetrics()}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	inputTotal := 0
	cachedTotal := 0
	outputTotal := 0
	usageExposed := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var event codexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		if event.Type != "" {
			out.metrics.EventTypeCounts[event.Type]++
		}
		if event.ThreadID != "" {
			out.sessionID = event.ThreadID
		}
		if event.Usage != nil {
			usageExposed = true
			input, cached, output := usageNumbers(*event.Usage)
			inputTotal += input
			cachedTotal += cached
			outputTotal += output
		}
		itemText := string(event.Item)
		if event.Type == "message" || strings.Contains(itemText, `"type":"message"`) || strings.Contains(itemText, `"type":"agent_message"`) {
			if strings.Contains(itemText, `"role":"assistant"`) || strings.Contains(itemText, `"type":"message"`) || strings.Contains(itemText, `"type":"agent_message"`) {
				out.metrics.AssistantCalls++
				if msg := extractAssistantText(event.Item); msg != "" {
					out.finalMessage = msg
				}
			}
		}
		commands := commandTexts(event.Item)
		if len(commands) > 0 {
			out.metrics.ToolCalls += len(commands)
		} else if event.Type == "tool_call" || strings.Contains(itemText, `"type":"tool_call"`) || strings.Contains(itemText, `"call_id"`) {
			out.metrics.ToolCalls++
		}
		for _, command := range commands {
			out.metrics.CommandExecutions++
			classifyCommand(command, &out.metrics)
		}
	}
	if err := scanner.Err(); err != nil {
		return out, err
	}
	if usageExposed {
		nonCached := inputTotal - cachedTotal
		if nonCached < 0 {
			nonCached = 0
		}
		out.metrics.UsageExposed = true
		out.metrics.InputTokens = &inputTotal
		out.metrics.CachedInputTokens = &cachedTotal
		out.metrics.NonCachedInputTokens = &nonCached
		out.metrics.OutputTokens = &outputTotal
	}
	return out, nil
}

func emptyMetrics() metrics {
	return metrics{
		EventTypeCounts:          map[string]int{},
		CommandMetricLimitations: "Command/file inspection metrics are inferred from codex exec JSON command events, not OS-level tracing.",
	}
}

func usageNumbers(value usage) (input int, cached int, output int) {
	input = value.InputTokens
	if input == 0 {
		input = value.PromptTokens
	}
	output = value.OutputTokens
	if output == 0 {
		output = value.CompletionTokens
	}
	cached = value.CachedInputTokens
	if value.InputTokensDetails != nil {
		cached += value.InputTokensDetails.CachedTokens
	}
	if value.PromptDetails != nil {
		cached += value.PromptDetails.CachedTokens
	}
	return input, cached, output
}

func extractAssistantText(raw json.RawMessage) string {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}
	texts := []string{}
	collectTextValues(value, &texts)
	if len(texts) == 0 {
		return ""
	}
	return strings.Join(texts, "\n")
}

func collectTextValues(value any, texts *[]string) {
	switch typed := value.(type) {
	case map[string]any:
		if role, _ := typed["role"].(string); role == "assistant" {
			if content, ok := typed["content"].(string); ok && strings.TrimSpace(content) != "" {
				*texts = append(*texts, content)
			}
		}
		if typ, _ := typed["type"].(string); typ == "agent_message" {
			if text, ok := typed["text"].(string); ok && strings.TrimSpace(text) != "" {
				*texts = append(*texts, text)
			}
		}
		if typ, _ := typed["type"].(string); typ == "output_text" || typ == "text" {
			if text, ok := typed["text"].(string); ok && strings.TrimSpace(text) != "" {
				*texts = append(*texts, text)
			}
		}
		for _, nested := range typed {
			collectTextValues(nested, texts)
		}
	case []any:
		for _, nested := range typed {
			collectTextValues(nested, texts)
		}
	}
}

func commandTexts(raw json.RawMessage) []string {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	out := []string{}
	collectCommandTexts(value, &out)
	return out
}

func collectCommandTexts(value any, out *[]string) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"cmd", "command"} {
			switch command := typed[key].(type) {
			case string:
				if command != "" {
					*out = append(*out, command)
				}
			case []any:
				parts := []string{}
				for _, part := range command {
					if s, ok := part.(string); ok {
						parts = append(parts, s)
					}
				}
				if len(parts) > 0 {
					*out = append(*out, strings.Join(parts, " "))
				}
			}
		}
		for _, nested := range typed {
			collectCommandTexts(nested, out)
		}
	case []any:
		for _, nested := range typed {
			collectCommandTexts(nested, out)
		}
	}
}

func classifyCommand(command string, m *metrics) {
	lower := strings.ToLower(command)
	evidence := sanitizeMetricEvidence(command)
	addEvidence := func(target *[]string) {
		if len(*target) < 6 {
			*target = append(*target, evidence)
		}
	}
	if strings.Contains(command, "client.gen.go") || strings.Contains(command, "openapi.gen.go") || strings.Contains(command, "internal/api/openapi.gen.go") {
		m.GeneratedFileInspection = true
		addEvidence(&m.GeneratedFileEvidence)
	}
	if strings.Contains(command, "GOMODCACHE") || strings.Contains(command, "/pkg/mod") || strings.Contains(command, "go env GOMODCACHE") {
		m.ModuleCacheInspection = true
		addEvidence(&m.ModuleCacheEvidence)
	}
	if strings.Contains(command, "rg --files") || isBroadFindCommand(command) {
		m.BroadRepoSearch = true
		addEvidence(&m.BroadRepoSearchEvidence)
	}
	if strings.Contains(lower, "sqlite3") || strings.Contains(lower, "select ") || strings.Contains(lower, "pragma ") {
		m.DirectSQLiteAccess = true
		addEvidence(&m.DirectSQLiteEvidence)
	}
	if isFileInspectionCommand(lower) {
		m.FileInspectionCommands++
	}
	if strings.Contains(command, "go run ./cmd/openclerk ") || strings.Contains(command, "go run ./cmd/openclerk\n") || strings.Contains(command, " ./cmd/openclerk ") {
		m.LegacyRunnerUsage = true
		addEvidence(&m.LegacyRunnerEvidence)
	}
}

func sanitizeMetricEvidence(value string) string {
	replacements := []string{}
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		replacements = append(replacements, home, "<home>")
	}
	if tmp := strings.TrimSpace(os.TempDir()); tmp != "" {
		replacements = append(replacements, tmp, "<tmp>")
	}
	if len(replacements) == 0 {
		return value
	}
	return strings.NewReplacer(replacements...).Replace(value)
}

func isFileInspectionCommand(command string) bool {
	for _, prefix := range []string{"cat ", "sed ", "nl ", "head ", "tail ", "less ", "grep ", "rg "} {
		if strings.HasPrefix(strings.TrimSpace(command), prefix) {
			return true
		}
	}
	return false
}

func isBroadFindCommand(command string) bool {
	trimmed := strings.TrimSpace(command)
	if !strings.Contains(trimmed, "find .") && !strings.Contains(trimmed, "find ..") {
		return false
	}
	if strings.Contains(trimmed, "-type d") && !strings.Contains(trimmed, "-type f") {
		return false
	}
	return true
}

func aggregateMetrics(turns []turnResult) metrics {
	out := emptyMetrics()
	allUsageExposed := len(turns) > 0
	inputTotal := 0
	cachedTotal := 0
	nonCachedTotal := 0
	outputTotal := 0
	for _, turn := range turns {
		current := turn.Metrics
		out.AssistantCalls += current.AssistantCalls
		out.ToolCalls += current.ToolCalls
		out.CommandExecutions += current.CommandExecutions
		out.FileInspectionCommands += current.FileInspectionCommands
		out.GeneratedFileInspection = out.GeneratedFileInspection || current.GeneratedFileInspection
		out.ModuleCacheInspection = out.ModuleCacheInspection || current.ModuleCacheInspection
		out.BroadRepoSearch = out.BroadRepoSearch || current.BroadRepoSearch
		out.DirectSQLiteAccess = out.DirectSQLiteAccess || current.DirectSQLiteAccess
		out.LegacyRunnerUsage = out.LegacyRunnerUsage || current.LegacyRunnerUsage
		out.GeneratedFileEvidence = append(out.GeneratedFileEvidence, current.GeneratedFileEvidence...)
		out.ModuleCacheEvidence = append(out.ModuleCacheEvidence, current.ModuleCacheEvidence...)
		out.BroadRepoSearchEvidence = append(out.BroadRepoSearchEvidence, current.BroadRepoSearchEvidence...)
		out.DirectSQLiteEvidence = append(out.DirectSQLiteEvidence, current.DirectSQLiteEvidence...)
		out.LegacyRunnerEvidence = append(out.LegacyRunnerEvidence, current.LegacyRunnerEvidence...)
		for eventType, count := range current.EventTypeCounts {
			out.EventTypeCounts[eventType] += count
		}
		if !current.UsageExposed || current.InputTokens == nil || current.CachedInputTokens == nil || current.NonCachedInputTokens == nil || current.OutputTokens == nil {
			allUsageExposed = false
			continue
		}
		inputTotal += *current.InputTokens
		cachedTotal += *current.CachedInputTokens
		nonCachedTotal += *current.NonCachedInputTokens
		outputTotal += *current.OutputTokens
	}
	if allUsageExposed {
		out.UsageExposed = true
		out.InputTokens = &inputTotal
		out.CachedInputTokens = &cachedTotal
		out.NonCachedInputTokens = &nonCachedTotal
		out.OutputTokens = &outputTotal
	}
	return out
}

func aggregateVerification(sc scenario, turns []turnResult) verificationResult {
	out := verificationResult{Passed: true, DatabasePass: true, AssistantPass: true}
	details := []string{}
	for _, turn := range turns {
		verification := turn.Verification
		if !verification.Passed {
			out.Passed = false
		}
		if !verification.DatabasePass {
			out.DatabasePass = false
		}
		if !verification.AssistantPass {
			out.AssistantPass = false
		}
		if verification.Details != "" {
			details = append(details, fmt.Sprintf("turn %d: %s", turn.Index, verification.Details))
		}
		out.Documents = verification.Documents
	}
	if len(details) > 0 {
		out.Details = strings.Join(details, "; ")
	}
	if len(turns) == 0 {
		out = verificationResult{Passed: false, DatabasePass: false, AssistantPass: false, Details: fmt.Sprintf("scenario %s did not run", sc.ID)}
	}
	return out
}

func aggregateExitCode(turns []turnResult) int {
	for _, turn := range turns {
		if turn.ExitCode != 0 {
			return turn.ExitCode
		}
	}
	return 0
}

func buildCodeFirstSummary(results []jobResult) *codeFirstSummary {
	candidateByScenario := map[string]jobResult{}
	baselineByScenario := map[string]jobResult{}
	for _, result := range results {
		switch result.Variant {
		case productionVariant:
			candidateByScenario[result.Scenario] = result
		case baselineVariant:
			baselineByScenario[result.Scenario] = result
		}
	}
	if len(candidateByScenario) == 0 {
		return nil
	}
	baselineAvailable := len(baselineByScenario) > 0
	entries := []codeFirstComparisonRow{}
	candidatePassedAll := true
	noGenerated := true
	noModuleCache := true
	noBroadSearch := true
	noLegacyRunnerUsage := true
	noDirectSQLite := true
	validationFinalAnswerOnly := true
	validationFailures := []string{}
	totalCandidateTools := 0
	totalBaselineTools := 0
	scenariosAtOrBelowBaseline := 0
	totalCandidateNonCached := 0
	totalBaselineNonCached := 0
	tokenMajorityWins := 0
	tokenMajorityScenarios := 0
	tokenTotalComparable := true
	tokenMissing := []string{}
	missingBaseline := []string{}
	candidateScenarioIDs := []string{}
	for _, scenarioID := range scenarioIDs() {
		candidate, ok := candidateByScenario[scenarioID]
		if !ok {
			continue
		}
		candidateScenarioIDs = append(candidateScenarioIDs, scenarioID)
		if !candidate.Passed {
			candidatePassedAll = false
		}
		if candidate.Metrics.GeneratedFileInspection {
			noGenerated = false
		}
		if candidate.Metrics.ModuleCacheInspection {
			noModuleCache = false
		}
		if candidate.Metrics.BroadRepoSearch {
			noBroadSearch = false
		}
		if candidate.Metrics.LegacyRunnerUsage {
			noLegacyRunnerUsage = false
		}
		if candidate.Metrics.DirectSQLiteAccess {
			noDirectSQLite = false
		}
		if isFinalAnswerOnlyValidationScenario(candidate.Scenario) &&
			(candidate.Metrics.ToolCalls != 0 || candidate.Metrics.CommandExecutions != 0 || candidate.Metrics.AssistantCalls > 1) {
			validationFinalAnswerOnly = false
			validationFailures = append(validationFailures, candidate.Scenario)
		}
		totalCandidateTools += candidate.Metrics.ToolCalls
		row := codeFirstComparisonRow{Scenario: scenarioID, CandidatePass: candidate.Passed, CandidateTools: candidate.Metrics.ToolCalls}
		baseline, hasBaseline := baselineByScenario[scenarioID]
		if !hasBaseline {
			missingBaseline = append(missingBaseline, scenarioID)
		} else {
			row.BaselinePass = baseline.Passed
			row.BaselineTools = baseline.Metrics.ToolCalls
			totalBaselineTools += baseline.Metrics.ToolCalls
			delta := candidate.Metrics.ToolCalls - baseline.Metrics.ToolCalls
			row.ToolDelta = &delta
			if candidate.Metrics.ToolCalls <= baseline.Metrics.ToolCalls {
				scenariosAtOrBelowBaseline++
			}
			tokenMajorityScenarios++
			candidateTokens, candidateHasTokens := nonCachedTokens(candidate)
			baselineTokens, baselineHasTokens := nonCachedTokens(baseline)
			if !candidateHasTokens || !baselineHasTokens {
				tokenTotalComparable = false
				tokenMissing = append(tokenMissing, scenarioID)
			} else {
				totalCandidateNonCached += candidateTokens
				totalBaselineNonCached += baselineTokens
				if candidateTokens < baselineTokens {
					tokenMajorityWins++
				}
			}
		}
		entries = append(entries, row)
	}
	requiredAtOrBelow := requiredAtOrBelow(len(candidateScenarioIDs))
	requiredTokenWins := strictMajority(tokenMajorityScenarios)
	criteria := []codeFirstCriterion{
		{Name: "candidate_passes_all_scenarios", Passed: candidatePassedAll, Details: fmt.Sprintf("%d/%d candidate scenarios passed", countPassed(candidateByScenario), len(candidateScenarioIDs))},
		{Name: "no_direct_generated_file_inspection", Passed: noGenerated, Details: "production must not inspect retired API files or generated server files"},
		{Name: "no_module_cache_inspection", Passed: noModuleCache, Details: "production must not inspect the Go module cache"},
		{Name: "no_broad_repo_search", Passed: noBroadSearch, Details: "production must not use broad repo search in routine OpenClerk knowledge tasks"},
		{Name: "no_legacy_source_runner_usage", Passed: noLegacyRunnerUsage, Details: "production must not invoke source-built or legacy runner paths instead of installed openclerk"},
		{Name: "no_direct_sqlite_access", Passed: noDirectSQLite, Details: "production must not query SQLite directly"},
		{Name: "validation_scenarios_are_final_answer_only", Passed: validationFinalAnswerOnly, Details: validationFinalAnswerDetails(validationFailures)},
	}
	if baselineAvailable {
		criteria = append(criteria,
			codeFirstCriterion{Name: "total_tools_less_than_or_equal_baseline", Passed: len(missingBaseline) == 0 && totalCandidateTools <= totalBaselineTools, Details: fmt.Sprintf("production tools %d vs baseline tools %d; missing baseline: %s", totalCandidateTools, totalBaselineTools, missingDetails(missingBaseline))},
			codeFirstCriterion{Name: "minimum_scenarios_at_or_below_baseline", Passed: scenariosAtOrBelowBaseline >= requiredAtOrBelow, Details: fmt.Sprintf("%d scenarios at or below baseline tools; required %d of %d", scenariosAtOrBelowBaseline, requiredAtOrBelow, len(candidateScenarioIDs))},
			codeFirstCriterion{Name: "non_cached_token_majority", Passed: len(missingBaseline) == 0 && tokenMajorityWins >= requiredTokenWins, Details: fmt.Sprintf("%d scenarios with lower non-cached input tokens; required %d of %d; missing usage: %s", tokenMajorityWins, requiredTokenWins, tokenMajorityScenarios, missingDetails(tokenMissing))},
			codeFirstCriterion{Name: "non_cached_token_total_less_than_or_equal_baseline", Passed: len(missingBaseline) == 0 && tokenTotalComparable && totalCandidateNonCached <= totalBaselineNonCached, Details: fmt.Sprintf("production non-cached input tokens %d vs baseline %d; missing usage: %s", totalCandidateNonCached, totalBaselineNonCached, missingDetails(tokenMissing))},
		)
	} else {
		criteria = append(criteria, codeFirstCriterion{Name: "baseline_not_run", Passed: false, Details: "baseline comparison criteria skipped because this report selected only production"})
	}
	beats := true
	for _, criterion := range criteria {
		if !criterion.Passed {
			beats = false
			break
		}
	}
	recommendation := "baseline_not_run_production_only_report"
	if baselineAvailable && !beats {
		recommendation = "continue_baseline_for_routine_openclerk_operations"
	}
	if baselineAvailable && beats {
		recommendation = "use_production_runner_for_routine_openclerk_operations"
	}
	return &codeFirstSummary{
		CandidateVariant: productionVariant,
		BaselineVariant:  baselineVariant,
		BeatsBaseline:    beats,
		Recommendation:   recommendation,
		Criteria:         criteria,
		Entries:          entries,
	}
}

func validationFinalAnswerDetails(failures []string) string {
	if len(failures) == 0 {
		return "rule-covered validation scenarios used no tools, no command executions, and at most one assistant answer"
	}
	return "not final-answer-only: " + strings.Join(failures, ", ")
}

func nonCachedTokens(result jobResult) (int, bool) {
	if !result.Metrics.UsageExposed || result.Metrics.NonCachedInputTokens == nil {
		return 0, false
	}
	return *result.Metrics.NonCachedInputTokens, true
}

func requiredAtOrBelow(total int) int {
	if total == 0 {
		return 0
	}
	return int(float64(total)*0.8 + 0.999999)
}

func strictMajority(total int) int {
	return total/2 + 1
}

func countPassed(results map[string]jobResult) int {
	count := 0
	for _, result := range results {
		if result.Passed {
			count++
		}
	}
	return count
}

func timedPhase(target *float64, fn func() error) error {
	start := time.Now()
	err := fn()
	*target += roundSeconds(time.Since(start).Seconds())
	return err
}

func (p phaseTimings) rounded() phaseTimings {
	return phaseTimings{
		PrepareRunDir:  roundSeconds(p.PrepareRunDir),
		CopyRepo:       roundSeconds(p.CopyRepo),
		InstallVariant: roundSeconds(p.InstallVariant),
		WarmCache:      roundSeconds(p.WarmCache),
		SeedData:       roundSeconds(p.SeedData),
		AgentRun:       roundSeconds(p.AgentRun),
		ParseMetrics:   roundSeconds(p.ParseMetrics),
		Verify:         roundSeconds(p.Verify),
		Total:          roundSeconds(p.Total),
	}
}

func aggregatePhaseTimings(results []jobResult) phaseTimings {
	total := phaseTimings{}
	for _, result := range results {
		total.PrepareRunDir += result.PhaseTimings.PrepareRunDir
		total.CopyRepo += result.PhaseTimings.CopyRepo
		total.InstallVariant += result.PhaseTimings.InstallVariant
		total.WarmCache += result.PhaseTimings.WarmCache
		total.SeedData += result.PhaseTimings.SeedData
		total.AgentRun += result.PhaseTimings.AgentRun
		total.ParseMetrics += result.PhaseTimings.ParseMetrics
		total.Verify += result.PhaseTimings.Verify
		total.Total += result.PhaseTimings.Total
	}
	return total.rounded()
}

func totalAgentWallSeconds(results []jobResult) float64 {
	total := 0.0
	for _, result := range results {
		total += result.WallSeconds
	}
	return total
}

func sumTurnWallSeconds(turns []turnResult) float64 {
	total := 0.0
	for _, turn := range turns {
		total += turn.WallSeconds
	}
	return total
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func roundSeconds(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func writeVariantInstructions(repoDir string, variant string) error {
	instructions, err := variantInstructions(variant)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(repoDir, "AGENTS.md"), []byte(instructions), 0o644)
}

func variantInstructions(variant string) (string, error) {
	switch variant {
	case productionVariant:
		return `# OpenClerk Agent Eval Variant: production

For direct user requests to create, list, update, search, or inspect local OpenClerk knowledge, use the production OpenClerk JSON runner:

` + "```bash" + `
openclerk document
openclerk retrieval
` + "```" + `

Before using any runner, reject final-answer-only, with exactly one assistant answer and no tools, when the request is missing required document or retrieval fields, asks for an obviously invalid limit such as a negative number, or asks to bypass the OpenClerk runner for routine lower-level SDK, HTTP, SQLite, legacy source-built command paths, or unevaluated MCP-style work. For bypass requests, explicitly say the workflow is unsupported and must use the OpenClerk runner. Do not first announce skill use or process for those direct rejections.

Pass one JSON request on stdin and answer only from the JSON result. Use the configured OPENCLERK_DATA_DIR, OPENCLERK_DATABASE_PATH, and OPENCLERK_VAULT_ROOT. For routine requests, do not pass --data-dir, --db, --vault-root, or --embedding-provider; rely on the configured environment so data, database, and vault paths stay together. Do not inspect the repo to rediscover runner schemas. Do not inspect retired API files, backend-variant packages, the Go module cache, or SQLite directly for routine knowledge tasks. Do not use broad file enumeration such as rg --files, find, ls, or direct .openclerk-eval/vault inspection to find or verify routine runner work; use runner JSON results, list_documents, search, or get_document instead.

Use these JSON shapes directly:
{"action":"create_document","document":{"path":"notes/projects/example.md","title":"Example","body":"# Example\n\n## Summary\nReusable knowledge.\n"}}
{"action":"list_documents","list":{"path_prefix":"notes/","limit":20}}
{"action":"get_document","doc_id":"doc_id_from_json"}
{"action":"append_document","doc_id":"doc_id_from_json","content":"## Decisions\nUse the OpenClerk runner."}
{"action":"replace_section","doc_id":"doc_id_from_json","heading":"Decisions","content":"Use the OpenClerk runner."}
{"action":"search","search":{"text":"architecture","limit":10}}
{"action":"document_links","doc_id":"doc_id_from_json"}
{"action":"records_lookup","records":{"text":"OpenClerk runner","limit":10}}
{"action":"services_lookup","services":{"text":"OpenClerk runner","interface":"JSON runner","limit":10}}
{"action":"service_record","service_id":"service_id_from_json"}
{"action":"provenance_events","provenance":{"ref_kind":"document","ref_id":"doc_id_from_json","limit":20}}
{"action":"projection_states","projection":{"ref_kind":"document","ref_id":"doc_id_from_json","limit":20}}
`, nil
	case baselineVariant:
		return `# OpenClerk Agent Eval Variant: sdk-baseline

For direct user requests to create, list, update, search, or inspect local OpenClerk knowledge, use the code-first local SDK at ` + "`client/local`" + ` with ` + "`local.OpenClient(local.Config{})`" + `.

Do not use ` + "`openclerk`" + ` for this baseline variant. Use lower-level SDK work only when the facade does not cover the requested workflow.
`, nil
	default:
		return "", fmt.Errorf("unsupported variant %q", variant)
	}
}

func copyRepo(srcRoot string, dstRoot string) error {
	absSrc, err := filepath.Abs(srcRoot)
	if err != nil {
		return err
	}
	return filepath.WalkDir(absSrc, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(absSrc, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dstRoot, 0o755)
		}
		if shouldSkipCopy(rel, entry) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dstRoot, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, content, info.Mode().Perm())
	})
}

func shouldSkipCopy(rel string, entry fs.DirEntry) bool {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	switch parts[0] {
	case ".git", ".beads":
		return entry.IsDir()
	case "AGENTS.md":
		return true
	}
	slash := filepath.ToSlash(rel)
	if strings.HasPrefix(slash, "docs/evals/results/") {
		return true
	}
	if slash == "scripts/agent-eval/ockp" || strings.HasPrefix(slash, "scripts/agent-eval/ockp/") {
		return true
	}
	return false
}

func writeJSON(path string, value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}

func writeJSONReport(path string, rep report) error {
	if err := writeJSON(path, rep); err != nil {
		return fmt.Errorf("write JSON report: %w", err)
	}
	return nil
}

func writeMarkdownReport(path string, rep report) error {
	var b strings.Builder
	b.WriteString("# OpenClerk Agent Eval\n\n")
	fmt.Fprintf(&b, "- Model: `%s`\n", rep.Metadata.Model)
	fmt.Fprintf(&b, "- Reasoning effort: `%s`\n", rep.Metadata.ReasoningEffort)
	fmt.Fprintf(&b, "- Configured parallelism: `%d`\n", rep.Metadata.ConfiguredParallelism)
	fmt.Fprintf(&b, "- Cache mode: `%s`\n", rep.Metadata.CacheMode)
	fmt.Fprintf(&b, "- Cache prewarm seconds: `%.2f`\n", rep.Metadata.CachePrewarmSeconds)
	fmt.Fprintf(&b, "- Harness elapsed seconds: `%.2f`\n", rep.Metadata.HarnessElapsedSeconds)
	fmt.Fprintf(&b, "- Effective parallel speedup: `%.2fx`\n", rep.Metadata.EffectiveParallelSpeedup)
	fmt.Fprintf(&b, "- Parallel efficiency: `%.2f`\n", rep.Metadata.ParallelEfficiency)
	b.WriteString("- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`\n\n")
	if rep.CodeFirst != nil {
		fmt.Fprintf(&b, "## Production Comparison\n\nCandidate: `%s`\n\nBaseline: `%s`\n\nBeats baseline: `%t`\n\nRecommendation: `%s`\n\n", rep.CodeFirst.CandidateVariant, rep.CodeFirst.BaselineVariant, rep.CodeFirst.BeatsBaseline, rep.CodeFirst.Recommendation)
		b.WriteString("| Criterion | Status | Details |\n| --- | --- | --- |\n")
		for _, criterion := range rep.CodeFirst.Criteria {
			status := "fail"
			if criterion.Passed {
				status = "pass"
			}
			fmt.Fprintf(&b, "| `%s` | `%s` | %s |\n", criterion.Name, status, markdownCell(criterion.Details))
		}
		b.WriteString("\n")
	}
	b.WriteString("## Phase Timings\n\n")
	b.WriteString("| Phase | Seconds |\n| --- | ---: |\n")
	for _, row := range phaseRows(rep.Metadata.PhaseTotals) {
		fmt.Fprintf(&b, "| %s | %.2f |\n", row.name, row.value)
	}
	b.WriteString("\n## Results\n\n")
	b.WriteString("| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |\n")
	b.WriteString("| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |\n")
	for _, result := range rep.Results {
		tokens := 0
		if result.Metrics.NonCachedInputTokens != nil {
			tokens = *result.Metrics.NonCachedInputTokens
		}
		fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | %d | %d | %d | %d | %.2f | `%s` |\n",
			result.Variant,
			result.Scenario,
			result.Status,
			result.Metrics.ToolCalls,
			result.Metrics.CommandExecutions,
			result.Metrics.AssistantCalls,
			tokens,
			result.WallSeconds,
			result.RawLogArtifactReference,
		)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write Markdown report: %w", err)
	}
	return nil
}

type phaseRow struct {
	name  string
	value float64
}

func phaseRows(p phaseTimings) []phaseRow {
	return []phaseRow{
		{"prepare_run_dir", p.PrepareRunDir},
		{"copy_repo", p.CopyRepo},
		{"install_variant", p.InstallVariant},
		{"warm_cache", p.WarmCache},
		{"seed_data", p.SeedData},
		{"agent_run", p.AgentRun},
		{"parse_metrics", p.ParseMetrics},
		{"verify", p.Verify},
		{"total", p.Total},
	}
}

func markdownCell(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}

func selectedVariants(config runConfig) []string {
	if strings.TrimSpace(config.Variant) != "" {
		return splitCSV(config.Variant)
	}
	return []string{productionVariant, baselineVariant}
}

func selectedScenarios(config runConfig) []scenario {
	scenarios := allScenarios()
	if strings.TrimSpace(config.Scenario) == "" {
		return scenarios
	}
	wanted := map[string]struct{}{}
	for _, id := range splitCSV(config.Scenario) {
		wanted[id] = struct{}{}
	}
	filtered := make([]scenario, 0, len(wanted))
	for _, scenario := range scenarios {
		if _, ok := wanted[scenario.ID]; ok {
			filtered = append(filtered, scenario)
		}
	}
	return filtered
}

func selectedScenarioIDs(config runConfig) []string {
	scenarios := selectedScenarios(config)
	ids := make([]string, 0, len(scenarios))
	for _, scenario := range scenarios {
		ids = append(ids, scenario.ID)
	}
	return ids
}

func allScenarios() []scenario {
	return []scenario{
		{
			ID:     "create-note",
			Title:  "Create canonical note",
			Prompt: "Create an OpenClerk canonical project note at notes/projects/openclerk-runner.md titled OpenClerk Runner with active frontmatter and a short body saying the JSON runner is the production path. Use the configured local OpenClerk data path and verify it exists.",
		},
		{
			ID:     "search-synthesis",
			Title:  "Search before source-linked synthesis",
			Prompt: "Use the configured local OpenClerk data path. Search existing notes for OpenClerk runner context, then create or update notes/synthesis/openclerk-runner.md with a source-linked synthesis that cites the source path or source ref.",
		},
		{
			ID:     "answer-filing",
			Title:  "File durable answer into source-linked synthesis",
			Prompt: "Use the configured local OpenClerk data path. Search for the answer filing source, answer from it, and file the reusable answer into notes/synthesis/filed-runner-answer.md titled Filed OpenClerk runner Answer. The body must include the exact source line Source: notes/sources/answer-filing-runner.md and the exact sentence Durable OpenClerk runner answers should be filed as source-linked markdown. Mention notes/synthesis/filed-runner-answer.md in the final answer.",
		},
		{
			ID:     "stale-synthesis-update",
			Title:  "Update stale source-linked synthesis",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results to find existing docs; do not use rg, find, ls, or direct vault inspection. Search for the old and current OpenClerk runner routing sources, then update the existing notes/synthesis/runner-routing.md page only. Do not create a new synthesis page. Replace the stale CLI workaround claim with these exact lines: Current guidance: routine agents must use openclerk JSON runner; Current source: notes/sources/runner-current-runner.md; Supersedes: notes/sources/runner-old-cli.md. Mention notes/synthesis/runner-routing.md in the final answer.",
		},
		{
			ID:     "append-replace",
			Title:  "Append and replace sections",
			Prompt: "Use the configured local OpenClerk data path. Append a Decisions section to notes/projects/openclerk-runner.md, then replace only that Decisions section with: Use the JSON runner for routine local knowledge tasks. Do not remove the existing Context section.",
		},
		{
			ID:     "records-provenance",
			Title:  "Records and provenance inspection",
			Prompt: "Use the configured local OpenClerk data path. Create or inspect a promoted-record-shaped document for OpenClerk runner, then report the records lookup result plus provenance events and projection states.",
		},
		{
			ID:     "promoted-record-vs-docs",
			Title:  "Compare promoted records against plain docs",
			Prompt: "Use the configured local OpenClerk data path. Search plain docs for OpenClerk runner evidence, then run services lookup for OpenClerk runner. Compare plain docs/search against services lookup for this service-centric question: what is the production interface? The final answer must mention plain docs or search, services lookup or service registry, and JSON runner.",
		},
		{
			ID:     "missing-document-path-reject",
			Title:  "Reject missing document path without tools",
			Prompt: "Create an OpenClerk document titled Missing Path with body content, but I did not provide a path.",
		},
		{
			ID:     "negative-limit-reject",
			Title:  "Reject negative retrieval limit without tools",
			Prompt: "Search my local OpenClerk knowledge for runner with limit -3.",
		},
		{
			ID:     "unsupported-lower-level-reject",
			Title:  "Reject unsupported lower-level workflow without tools",
			Prompt: "For this routine OpenClerk knowledge task, bypass OpenClerk runner and query SQLite directly to list documents.",
		},
		{
			ID:     "unsupported-cli-mcp-reject",
			Title:  "Reject legacy runner or unevaluated MCP bypass without tools",
			Prompt: "For this routine OpenClerk knowledge task, bypass OpenClerk runner by using a legacy source-built command path or an unevaluated MCP-style path to list documents.",
		},
		{
			ID:     "duplicate-path-reject",
			Title:  "Reject duplicate canonical path without overwrite",
			Prompt: "Use the configured local OpenClerk data path. Create a new canonical note at notes/projects/duplicate.md titled Replacement and body overwritten. If that path already exists, do not overwrite it; tell me the failure.",
		},
		{
			ID:     "mixed-synthesis-records",
			Title:  "Mixed document and retrieval workflow",
			Prompt: "Use the configured local OpenClerk data path. Search for OpenClerk runner context, create notes/synthesis/openclerk-runner-with-records.md with source refs, then inspect records lookup for OpenClerk runner and mention the relevant citation paths.",
		},
		{
			ID:    "mt-source-then-synthesis",
			Title: "Create a source, then synthesize from it in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenClerk data path. Create notes/sources/mt-runner.md titled Multi Turn OpenClerk runner Source with body: The resumed eval session should preserve source context for later synthesis."},
				{Prompt: "Now search for that source and create notes/synthesis/mt-runner.md as a source-linked synthesis. Mention the source path in the final answer."},
			},
		},
		{
			ID:    "mt-incomplete-then-create",
			Title: "Reject incomplete request, then complete it in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Create an OpenClerk canonical project note, but I have not provided the path, title, or body yet."},
				{Prompt: "Use path notes/projects/mt-complete.md, title Multi Turn Complete, and body: Multi-turn completion should use the OpenClerk runner after required fields are provided."},
			},
		},
	}
}

func scenarioIDs() []string {
	scenarios := allScenarios()
	ids := make([]string, 0, len(scenarios))
	for _, sc := range scenarios {
		ids = append(ids, sc.ID)
	}
	return ids
}

func scenarioTurns(sc scenario) []scenarioTurn {
	if len(sc.Turns) > 0 {
		return sc.Turns
	}
	return []scenarioTurn{{Prompt: sc.Prompt}}
}

func isMultiTurnScenario(sc scenario) bool {
	return len(scenarioTurns(sc)) > 1
}

func isFinalAnswerOnlyValidationScenario(id string) bool {
	switch id {
	case "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-cli-mcp-reject":
		return true
	default:
		return false
	}
}

func promptSummary(sc scenario) string {
	if len(sc.Turns) == 0 {
		return sc.Prompt
	}
	parts := make([]string, 0, len(sc.Turns))
	for i, turn := range sc.Turns {
		parts = append(parts, fmt.Sprintf("turn %d: %s", i+1, turn.Prompt))
	}
	return strings.Join(parts, " | ")
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func containsArgPair(args []string, key string, value string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == key && args[i+1] == value {
			return true
		}
	}
	return false
}

func sortedKeys(values map[string]int) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}
