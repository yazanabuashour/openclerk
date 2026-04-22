package main

import (
	"bufio"
	"bytes"
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
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

const (
	defaultParallel   = 4
	modelName         = "gpt-5.4-mini"
	reasoningEffort   = "medium"
	productionVariant = "production"
	cacheModeShared   = "shared"
	cacheModeIsolated = "isolated"

	ragRetrievalScenarioID   = "rag-retrieval-baseline"
	ragCurrentPolicyPath     = "notes/rag/current-runner-policy.md"
	ragDecoyPolicyPath       = "notes/rag/decoy-runner-policy.md"
	ragArchivedPolicyPath    = "notes/archive/old-runner-policy.md"
	ragSearchText            = "active AgentOps RAG baseline policy JSON runner citations"
	ragPathPrefix            = "notes/rag/"
	ragMetadataKey           = "rag_scope"
	ragMetadataValue         = "active-policy"
	ragCurrentPolicyTitle    = "Current AgentOps RAG Policy"
	ragCurrentPolicySummary  = "Active AgentOps RAG baseline policy marker: routine OpenClerk knowledge answers must use the installed openclerk JSON runner and include source citations with doc_id and chunk_id."
	ragCurrentPolicyDecision = "The active retrieval decision is JSON runner only."
	ragDecoyPolicyTitle      = "Decoy AgentOps RAG Policy"
	ragArchivedPolicyTitle   = "Archived AgentOps RAG Policy"

	docsNavigationScenarioID = "canonical-docs-navigation-baseline"
	docsNavigationPrefix     = "notes/wiki/agentops/"
	docsNavigationIndexPath  = "notes/wiki/agentops/index.md"
	docsNavigationPolicyPath = "notes/wiki/agentops/runner-policy.md"
	docsNavigationArchPath   = "notes/wiki/architecture/knowledge-plane.md"
	docsNavigationOpsPath    = "notes/wiki/ops/runner-playbook.md"
)

var (
	prewarmCompilePackages = []string{"./cmd/openclerk", "./internal/runner"}
	unixHomePathPattern    = regexp.MustCompile(`/(Users|home)/[^/\s"'\\]+`)
	windowsHomePathPattern = regexp.MustCompile(`(?i)[A-Z]:\\Users\\[^\\\s"']+`)
)

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
	Metadata       reportMetadata         `json:"metadata"`
	Results        []jobResult            `json:"results"`
	ProductionGate *productionGateSummary `json:"production_gate,omitempty"`
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
	SearchUsed               bool           `json:"search_used"`
	SearchUnfilteredUsed     bool           `json:"search_unfiltered_used"`
	SearchPathFilterUsed     bool           `json:"search_path_filter_used"`
	SearchMetadataFilterUsed bool           `json:"search_metadata_filter_used"`
	ListDocumentsUsed        bool           `json:"list_documents_used"`
	GetDocumentUsed          bool           `json:"get_document_used"`
	DocumentLinksUsed        bool           `json:"document_links_used"`
	GraphNeighborhoodUsed    bool           `json:"graph_neighborhood_used"`
	RecordsLookupUsed        bool           `json:"records_lookup_used"`
	ProvenanceEventsUsed     bool           `json:"provenance_events_used"`
	ProjectionStatesUsed     bool           `json:"projection_states_used"`
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

type productionGateSummary struct {
	Variant        string                    `json:"variant"`
	PassesGate     bool                      `json:"passes_gate"`
	Recommendation string                    `json:"recommendation"`
	Criteria       []productionGateCriterion `json:"criteria"`
}

type productionGateCriterion struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
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
		Results:        results,
		ProductionGate: buildProductionGateSummary(results),
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
		if err := installVariant(config.RepoRoot, repoDir, job.Variant); err != nil {
			return err
		}
		if err := buildOpenClerkRunner(repoDir, jobDir, paths, cache); err != nil {
			return err
		}
		return preflightEvalContext(config.RepoRoot, repoDir, jobDir, paths, cache, config.CodexBin)
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
	cfg := runclient.Config{
		DataDir:      paths.DataDir,
		DatabasePath: paths.DatabasePath,
		VaultRoot:    paths.VaultRoot,
	}
	switch sc.ID {
	case "search-synthesis", "mt-source-then-synthesis":
		if err := createSeedDocument(ctx, cfg, "notes/sources/openclerk-runner.md", "OpenClerk Runner Source", "The OpenClerk runner uses JSON requests for OpenClerk knowledge tasks.\n\nIt preserves source refs for synthesis pages."); err != nil {
			return err
		}
	case "answer-filing":
		if err := createSeedDocument(ctx, cfg, "notes/sources/answer-filing-runner.md", "OpenClerk runner Answer Filing Source", "The OpenClerk runner JSON runner is the production path for reusable OpenClerk knowledge tasks.\n\nDurable OpenClerk runner answers should be filed as source-linked markdown."); err != nil {
			return err
		}
	case ragRetrievalScenarioID:
		if err := seedRAGRetrievalBaseline(ctx, cfg); err != nil {
			return err
		}
	case docsNavigationScenarioID:
		if err := seedDocsNavigationBaseline(ctx, cfg); err != nil {
			return err
		}
	case "stale-synthesis-update":
		if err := createSeedDocument(ctx, cfg, "notes/sources/runner-old-workaround.md", "Old OpenClerk runner Routing Source", "Older guidance said routine agents may bypass OpenClerk runner through a temporary command-path workaround."); err != nil {
			return err
		}
		if err := createSeedDocument(ctx, cfg, "notes/sources/runner-current-runner.md", "Current OpenClerk runner Routing Source", "Current guidance says routine agents must use openclerk JSON runner for OpenClerk knowledge tasks."); err != nil {
			return err
		}
		body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: notes/sources/runner-current-runner.md, notes/sources/runner-old-workaround.md
---

# OpenClerk runner Routing

## Summary
Stale claim: routine agents may bypass OpenClerk runner through a temporary command-path workaround.

## Sources
- notes/sources/runner-current-runner.md
- notes/sources/runner-old-workaround.md

## Freshness
Checked source: notes/sources/runner-old-workaround.md
`)
		if err := createSeedDocument(ctx, cfg, "notes/synthesis/runner-routing.md", "OpenClerk runner Routing", body); err != nil {
			return err
		}
	case "synthesis-freshness-repair":
		oldBody := strings.TrimSpace(`---
status: superseded
superseded_by: notes/sources/repair-current.md
---
# Old OpenClerk runner Repair Source

## Summary
Older repair guidance mentioned a temporary command-path workaround.
`) + "\n"
		if err := createSeedDocument(ctx, cfg, "notes/sources/repair-old.md", "Old OpenClerk runner Repair Source", oldBody); err != nil {
			return err
		}
		currentBody := strings.TrimSpace(`---
supersedes: notes/sources/repair-old.md
---
# Current OpenClerk runner Repair Source

## Summary
Current guidance says routine agents must use openclerk JSON runner for freshness repairs.
`) + "\n"
		if err := createSeedDocument(ctx, cfg, "notes/sources/repair-current.md", "Current OpenClerk runner Repair Source", currentBody); err != nil {
			return err
		}
		synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: notes/sources/repair-current.md, notes/sources/repair-old.md
---
# OpenClerk runner Freshness Repair

## Summary
Stale repair claim: routine agents may use a temporary command-path workaround.

## Sources
- notes/sources/repair-current.md
- notes/sources/repair-old.md

## Freshness
Checked before the latest source update.
`) + "\n"
		if err := createSeedDocument(ctx, cfg, "notes/synthesis/runner-repair.md", "OpenClerk runner Freshness Repair", synthesisBody); err != nil {
			return err
		}
		if err := replaceScenarioSeedSection(ctx, cfg, "notes/sources/repair-current.md", "Summary", "Current guidance says routine agents must use openclerk JSON runner for freshness repairs, and notes/sources/repair-old.md is superseded."); err != nil {
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
	case "mixed-synthesis-records":
		if err := createSeedDocument(ctx, cfg, "notes/sources/openclerk-runner.md", "OpenClerk Runner Source", "The OpenClerk runner uses JSON requests for OpenClerk knowledge tasks.\n\nIt preserves source refs for synthesis pages."); err != nil {
			return err
		}
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

func seedRAGRetrievalBaseline(ctx context.Context, cfg runclient.Config) error {
	currentBody := strings.TrimSpace(`---
type: note
status: active
rag_scope: active-policy
---
# Current AgentOps RAG Policy

## Summary
`+ragCurrentPolicySummary+`

## Decision
`+ragCurrentPolicyDecision+`
`) + "\n"
	if err := createSeedDocument(ctx, cfg, ragCurrentPolicyPath, ragCurrentPolicyTitle, currentBody); err != nil {
		return err
	}
	decoyBody := strings.TrimSpace(`---
type: note
status: draft
rag_scope: decoy-policy
---
# Decoy AgentOps RAG Policy

## Summary
Decoy AgentOps RAG baseline policy marker: this draft says direct SQLite might be acceptable for routine OpenClerk knowledge answers.

## Decision
This is a decoy policy and is not the active AgentOps retrieval decision.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, ragDecoyPolicyPath, ragDecoyPolicyTitle, decoyBody); err != nil {
		return err
	}
	archivedBody := strings.TrimSpace(`---
type: note
status: superseded
rag_scope: archived-policy
---
# Archived AgentOps RAG Policy

## Summary
Archived AgentOps RAG baseline policy marker: older guidance mentioned a source-built command path.

## Decision
This archived policy is outside the active RAG path prefix and is superseded by the current JSON runner policy.
`) + "\n"
	return createSeedDocument(ctx, cfg, ragArchivedPolicyPath, ragArchivedPolicyTitle, archivedBody)
}

func seedDocsNavigationBaseline(ctx context.Context, cfg runclient.Config) error {
	indexBody := strings.TrimSpace(`---
type: wiki
status: active
---
# AgentOps Wiki Index

## Summary
Canonical directory navigation starts here for the AgentOps wiki baseline.

## Links
- [Runner policy](runner-policy.md)
- [Knowledge plane](../architecture/knowledge-plane.md)
- [Runner playbook](../ops/runner-playbook.md)

## Limits
Folder paths and headings show the local index, but they do not explain backlinks or cross-directory relationship neighborhoods without retrieval actions.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, docsNavigationIndexPath, "AgentOps Wiki Index", indexBody); err != nil {
		return err
	}

	policyBody := strings.TrimSpace(`---
type: policy
status: active
---
# Runner Policy

## Summary
Routine OpenClerk knowledge work uses the installed JSON runner and cites returned source paths.

## Navigation
Return to the [AgentOps wiki index](index.md) and compare with the [knowledge plane](../architecture/knowledge-plane.md).
`) + "\n"
	if err := createSeedDocument(ctx, cfg, docsNavigationPolicyPath, "Runner Policy", policyBody); err != nil {
		return err
	}

	architectureBody := strings.TrimSpace(`---
type: architecture
status: active
---
# Knowledge Plane

## Summary
The knowledge plane keeps canonical markdown as source authority and derives graph relationships from links.

## Navigation
The [AgentOps wiki index](../agentops/index.md) links this architecture note to runner policy context.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, docsNavigationArchPath, "Knowledge Plane", architectureBody); err != nil {
		return err
	}

	opsBody := strings.TrimSpace(`---
type: runbook
status: active
---
# Runner Playbook

## Summary
Operators use the runner playbook when directory navigation is not enough to explain related policy and architecture docs.

## Navigation
Start from the [AgentOps wiki index](../agentops/index.md) before following graph neighborhoods.
`) + "\n"
	return createSeedDocument(ctx, cfg, docsNavigationOpsPath, "Runner Playbook", opsBody)
}

func createSeedDocument(ctx context.Context, cfg runclient.Config, path, title, body string) error {
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

func replaceScenarioSeedSection(ctx context.Context, cfg runclient.Config, docPath, heading, content string) error {
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docPath, Limit: 5},
	})
	if err != nil {
		return err
	}
	for _, doc := range list.Documents {
		if doc.Path != docPath {
			continue
		}
		result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
			Action:  runner.DocumentTaskActionReplaceSection,
			DocID:   doc.DocID,
			Heading: heading,
			Content: content,
		})
		if err != nil {
			return err
		}
		if result.Rejected {
			return errors.New(result.RejectionReason)
		}
		return nil
	}
	return fmt.Errorf("seed document %s not found", docPath)
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
		return verifySourceLinkedSynthesis(ctx, paths, "notes/synthesis/openclerk-runner.md", finalMessage, sourceLinkedSynthesisExpectations{
			SourceRefs:      []string{"notes/sources/openclerk-runner.md"},
			RequireSearch:   true,
			RequireList:     true,
			Metrics:         turnMetrics,
			FinalAnswerPath: true,
		})
	case "answer-filing":
		return verifyAnswerFiling(ctx, paths, finalMessage)
	case ragRetrievalScenarioID:
		return verifyRAGRetrievalBaseline(ctx, paths, finalMessage, turnMetrics)
	case docsNavigationScenarioID:
		return verifyDocsNavigationBaseline(ctx, paths, finalMessage, turnMetrics)
	case "stale-synthesis-update":
		return verifyStaleSynthesisUpdate(ctx, paths, finalMessage, turnMetrics)
	case "synthesis-freshness-repair":
		return verifySynthesisFreshnessRepair(ctx, paths, finalMessage, turnMetrics)
	case "append-replace":
		return verifyDocumentContains(ctx, paths, "notes/projects/openclerk-runner.md", []string{"Existing context stays intact", "Use the JSON runner"}, []string{"temporary command-path workaround"})
	case "records-provenance":
		return verifyRecordsAndProvenance(ctx, paths, finalMessage, turnMetrics)
	case "promoted-record-vs-docs":
		return verifyPromotedRecordVsDocs(ctx, paths, finalMessage, turnMetrics)
	case "duplicate-path-reject":
		return verifyDuplicatePathReject(ctx, paths, finalMessage)
	case "mixed-synthesis-records":
		return verifyMixedSynthesisRecords(ctx, paths, finalMessage, turnMetrics)
	case "mt-source-then-synthesis":
		return verifySourceLinkedSynthesis(ctx, paths, "notes/synthesis/mt-runner.md", finalMessage, sourceLinkedSynthesisExpectations{
			SourceRefs:      []string{"notes/sources/mt-runner.md"},
			RequireSearch:   true,
			Metrics:         turnMetrics,
			FinalAnswerPath: true,
			AdditionalDocs:  []string{"notes/sources/mt-runner.md"},
		})
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
		return containsAny(lower, []string{"missing", "required", "requires", "need", "provide"}) && strings.Contains(lower, "path")
	case "negative-limit-reject":
		return containsAny(lower, []string{"negative", "invalid", "non-negative", "positive"}) && strings.Contains(lower, "limit")
	case "unsupported-lower-level-reject":
		return containsAny(lower, []string{"unsupported", "does not support", "cannot bypass", "can't bypass", "must use runner", "do not bypass", "use runner", "cannot do that", "can't do that", "cannot comply", "can't comply", "cannot fulfill", "can't fulfill"})
	case "unsupported-transport-reject":
		return containsAny(lower, []string{"unsupported", "cannot bypass", "cannot help bypass", "can't bypass", "can't help bypass", "do not bypass", "must use runner", "use runner"}) &&
			containsAny(lower, []string{"transport", "path", "runner"})
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
	cfg := runclient.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
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
	cfg := runclient.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
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

type sourceLinkedSynthesisExpectations struct {
	SourceRefs                 []string
	RequireSearch              bool
	RequireList                bool
	RequireGet                 bool
	RequireRecordsLookup       bool
	RequireProvenanceEvents    bool
	RequireProjectionStates    bool
	Metrics                    metrics
	FinalAnswerPath            bool
	AdditionalDocs             []string
	AdditionalBodyRequirements []string
}

func verifySourceLinkedSynthesis(ctx context.Context, paths evalPaths, docPath string, finalMessage string, expectations sourceLinkedSynthesisExpectations) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+docPath)
	}
	documents := append([]string{}, expectations.AdditionalDocs...)
	documents = append(documents, docPath)
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"## Sources",
		"## Freshness",
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, expectations.SourceRefs)...)
	failures = append(failures, missingRequiredFold(body, expectations.AdditionalBodyRequirements)...)
	if expectations.FinalAnswerPath && !messageContainsAll(finalMessage, []string{docPath}) {
		failures = append(failures, "final answer did not mention "+docPath)
	}
	if expectations.RequireSearch && !expectations.Metrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if expectations.RequireList && !expectations.Metrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list existing synthesis candidates")
	}
	if expectations.RequireGet && !expectations.Metrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	if expectations.RequireRecordsLookup && !expectations.Metrics.RecordsLookupUsed {
		failures = append(failures, "agent did not use records lookup")
	}
	if expectations.RequireProvenanceEvents && !expectations.Metrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if expectations.RequireProjectionStates && !expectations.Metrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection states")
	}
	databaseFailures := missingRequired(body, required)
	databaseFailures = append(databaseFailures, sourceRefsFrontmatterFailures(body, expectations.SourceRefs)...)
	databaseFailures = append(databaseFailures, missingRequiredFold(body, expectations.AdditionalBodyRequirements)...)
	databasePass := found && len(databaseFailures) == 0
	assistantPass := strings.TrimSpace(finalMessage) != ""
	if expectations.FinalAnswerPath {
		assistantPass = assistantPass && messageContainsAll(finalMessage, []string{docPath})
	}
	activityPass := (!expectations.RequireSearch || expectations.Metrics.SearchUsed) &&
		(!expectations.RequireList || expectations.Metrics.ListDocumentsUsed) &&
		(!expectations.RequireGet || expectations.Metrics.GetDocumentUsed) &&
		(!expectations.RequireRecordsLookup || expectations.Metrics.RecordsLookupUsed) &&
		(!expectations.RequireProvenanceEvents || expectations.Metrics.ProvenanceEventsUsed) &&
		(!expectations.RequireProjectionStates || expectations.Metrics.ProjectionStatesUsed)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     documents,
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

func verifyRAGRetrievalBaseline(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	unfiltered, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  ragSearchText,
			Limit: 5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	pathFiltered, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       ragSearchText,
			PathPrefix: ragPathPrefix,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	metadataFiltered, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          ragSearchText,
			MetadataKey:   ragMetadataKey,
			MetadataValue: ragMetadataValue,
			Limit:         5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	repeatedMetadata, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          ragSearchText,
			MetadataKey:   ragMetadataKey,
			MetadataValue: ragMetadataValue,
			Limit:         5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "notes/synthesis/")
	if err != nil {
		return verificationResult{}, err
	}

	failures := []string{}
	unfilteredTop, unfilteredTopFound := topSearchHit(unfiltered)
	pathTop, pathTopFound := topSearchHit(pathFiltered)
	metadataTop, metadataTopFound := topSearchHit(metadataFiltered)
	repeatedTop, repeatedTopFound := topSearchHit(repeatedMetadata)
	if !unfilteredTopFound || searchHitPath(unfilteredTop) != ragCurrentPolicyPath {
		failures = append(failures, "unfiltered search did not rank active RAG source first")
	}
	if !pathTopFound || searchHitPath(pathTop) != ragCurrentPolicyPath {
		failures = append(failures, "path-filtered search did not rank active RAG source first")
	}
	if searchContainsPath(pathFiltered, ragArchivedPolicyPath) {
		failures = append(failures, "path-filtered search included archived source")
	}
	if !metadataTopFound || searchHitPath(metadataTop) != ragCurrentPolicyPath {
		failures = append(failures, "metadata-filtered search did not rank active RAG source first")
	}
	if !searchOnlyContainsPath(metadataFiltered, ragCurrentPolicyPath) {
		failures = append(failures, "metadata-filtered search returned non-active policy sources")
	}
	if !metadataTopFound || !repeatedTopFound || metadataTop.DocID != repeatedTop.DocID || metadataTop.ChunkID != repeatedTop.ChunkID {
		failures = append(failures, "repeated metadata-filtered search changed top doc_id or chunk_id")
	}
	if !metadataTopFound || !searchHitHasCitation(metadataTop) {
		failures = append(failures, "metadata-filtered top hit did not include doc_id, chunk_id, path, and line citation")
	}
	if synthesisCount != 0 {
		failures = append(failures, fmt.Sprintf("retrieval-only baseline created %d synthesis documents", synthesisCount))
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.SearchUnfilteredUsed {
		failures = append(failures, "agent did not use unfiltered retrieval search")
	}
	if !turnMetrics.SearchPathFilterUsed {
		failures = append(failures, "agent did not use path-prefix retrieval search")
	}
	if !turnMetrics.SearchMetadataFilterUsed {
		failures = append(failures, "agent did not use metadata-filtered retrieval search")
	}

	assistantPass := metadataTopFound &&
		messageContainsAll(finalMessage, []string{ragCurrentPolicyPath, metadataTop.DocID, metadataTop.ChunkID}) &&
		messageContainsAny(finalMessage, []string{"json runner", "openclerk json runner"})
	if !assistantPass {
		failures = append(failures, "final answer did not cite active path, doc_id, chunk_id, and JSON runner policy")
	}
	databasePass := unfilteredTopFound &&
		pathTopFound &&
		metadataTopFound &&
		searchHitPath(unfilteredTop) == ragCurrentPolicyPath &&
		searchHitPath(pathTop) == ragCurrentPolicyPath &&
		searchHitPath(metadataTop) == ragCurrentPolicyPath &&
		!searchContainsPath(pathFiltered, ragArchivedPolicyPath) &&
		searchOnlyContainsPath(metadataFiltered, ragCurrentPolicyPath) &&
		repeatedTopFound &&
		metadataTop.DocID == repeatedTop.DocID &&
		metadataTop.ChunkID == repeatedTop.ChunkID &&
		searchHitHasCitation(metadataTop) &&
		synthesisCount == 0
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.SearchUnfilteredUsed &&
		turnMetrics.SearchPathFilterUsed &&
		turnMetrics.SearchMetadataFilterUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{ragCurrentPolicyPath, ragDecoyPolicyPath, ragArchivedPolicyPath},
	}, nil
}

func verifyDocsNavigationBaseline(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docsNavigationPrefix, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	indexDocID, indexFound := "", false
	policyFound := false
	onlyPrefix := true
	for _, doc := range list.Documents {
		if !strings.HasPrefix(doc.Path, docsNavigationPrefix) {
			onlyPrefix = false
		}
		switch doc.Path {
		case docsNavigationIndexPath:
			indexDocID = doc.DocID
			indexFound = true
		case docsNavigationPolicyPath:
			policyFound = true
		}
	}

	got, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  indexDocID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasHeadings := got.Document != nil && containsAllStrings(got.Document.Headings, []string{"AgentOps Wiki Index", "Summary", "Links", "Limits"})

	links, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDocumentLinks,
		DocID:  indexDocID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasOutgoing := links.Links != nil &&
		documentLinksContainPath(links.Links.Outgoing, docsNavigationPolicyPath) &&
		documentLinksContainPath(links.Links.Outgoing, docsNavigationArchPath) &&
		documentLinksContainPath(links.Links.Outgoing, docsNavigationOpsPath) &&
		documentLinksHaveCitations(links.Links.Outgoing)
	hasIncoming := links.Links != nil &&
		documentLinksContainPath(links.Links.Incoming, docsNavigationPolicyPath) &&
		documentLinksContainPath(links.Links.Incoming, docsNavigationArchPath) &&
		documentLinksContainPath(links.Links.Incoming, docsNavigationOpsPath) &&
		documentLinksHaveCitations(links.Links.Incoming)

	graph, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraph,
		DocID:  indexDocID,
		Limit:  20,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasGraph := graph.Graph != nil &&
		graphContainsNodeLabels(graph.Graph.Nodes, []string{"AgentOps Wiki Index", "Runner Policy", "Knowledge Plane", "Runner Playbook"}) &&
		graphContainsLinkEdge(graph.Graph.Edges) &&
		graphEdgesHaveCitations(graph.Graph.Edges)

	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "graph",
			RefKind:    "document",
			RefID:      indexDocID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) == 1 &&
		projections.Projections.Projections[0].Freshness == "fresh" &&
		projections.Projections.Projections[0].Details["path"] == docsNavigationIndexPath

	failures := []string{}
	if !indexFound {
		failures = append(failures, "path-prefix listing did not find "+docsNavigationIndexPath)
	}
	if !policyFound {
		failures = append(failures, "path-prefix listing did not find "+docsNavigationPolicyPath)
	}
	if !onlyPrefix || len(list.Documents) != 2 {
		failures = append(failures, "path-prefix listing did not stay scoped to agentops directory")
	}
	if !hasHeadings {
		failures = append(failures, "get_document did not expose expected index headings")
	}
	if !hasOutgoing {
		failures = append(failures, "document_links missing cited outgoing links")
	}
	if !hasIncoming {
		failures = append(failures, "document_links missing cited incoming backlinks")
	}
	if !hasGraph {
		failures = append(failures, "graph_neighborhood missing cited nodes or edges")
	}
	if !hasProjection {
		failures = append(failures, "graph projection state missing or not fresh")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not use list_documents")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not use get_document")
	}
	if !turnMetrics.DocumentLinksUsed {
		failures = append(failures, "agent did not use document_links")
	}
	if !turnMetrics.GraphNeighborhoodUsed {
		failures = append(failures, "agent did not use graph_neighborhood")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect graph projection state")
	}

	assistantPass := messageContainsAny(finalMessage, []string{"directory", "folder", "path-prefix", "path prefix"}) &&
		messageContainsAny(finalMessage, []string{"link", "markdown"}) &&
		messageContainsAny(finalMessage, []string{"backlink", "incoming"}) &&
		messageContainsAny(finalMessage, []string{"graph neighborhood", "graph_neighborhood"}) &&
		messageContainsAny(finalMessage, []string{"sufficient", "enough"}) &&
		messageContainsAny(finalMessage, []string{"fails", "fail", "limits", "not enough"}) &&
		messageContainsAll(finalMessage, []string{docsNavigationIndexPath})
	if !assistantPass {
		failures = append(failures, "final answer did not compare directory, links/backlinks, graph neighborhood, limits, and source path")
	}

	databasePass := indexFound &&
		policyFound &&
		onlyPrefix &&
		len(list.Documents) == 2 &&
		hasHeadings &&
		hasOutgoing &&
		hasIncoming &&
		hasGraph &&
		hasProjection
	activityPass := turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.DocumentLinksUsed &&
		turnMetrics.GraphNeighborhoodUsed &&
		turnMetrics.ProjectionStatesUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docsNavigationIndexPath, docsNavigationPolicyPath, docsNavigationArchPath, docsNavigationOpsPath},
	}, nil
}

func verifyStaleSynthesisUpdate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
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
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current guidance: routine agents must use openclerk JSON runner",
		"Current source: notes/sources/runner-current-runner.md",
		"Supersedes: notes/sources/runner-old-workaround.md",
		"## Sources",
		"## Freshness",
	}
	sourceRefs := []string{"notes/sources/runner-current-runner.md", "notes/sources/runner-old-workaround.md"}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, sourceRefs)...)
	failures = append(failures, presentForbidden(body, []string{"may bypass OpenClerk runner through a temporary command-path workaround"})...)
	if !containsAny(strings.ToLower(body), []string{"stale", "supersedes", "superseded", "contradiction", "current guidance"}) {
		failures = append(failures, "missing stale or supersession language")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list existing synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	assistantPass := messageContainsAll(finalMessage, []string{docPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "current", "supersedes", "stale"})
	if !assistantPass {
		failures = append(failures, "final answer did not describe the synthesis update")
	}
	databasePass := found && exactCount == 1 && createdCurrent == 0 && createdUpdated == 0 &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, sourceRefs)) == 0 &&
		len(presentForbidden(body, []string{"may bypass OpenClerk runner through a temporary command-path workaround"})) == 0 &&
		containsAny(strings.ToLower(body), []string{"stale", "supersedes", "superseded", "contradiction", "current guidance"})
	activityPass := turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}

func verifySynthesisFreshnessRepair(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	docPath := "notes/synthesis/runner-repair.md"
	currentSource := "notes/sources/repair-current.md"
	supersededSource := "notes/sources/repair-old.md"
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	docID, docIDFound, err := documentIDByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	cfg := runclient.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      docID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	events, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "synthesis:" + docID,
			Limit:   10,
		},
	})
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
	if !docIDFound {
		failures = append(failures, "missing document id for "+docPath)
	}
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"source_refs: notes/sources/repair-current.md, notes/sources/repair-old.md",
		currentSource,
		supersededSource,
		"## Sources",
		"## Freshness",
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, presentForbidden(body, []string{"may use a temporary command-path workaround"})...)
	hasProjection := false
	hasCurrent := false
	hasSuperseded := false
	if projections.Projections != nil && len(projections.Projections.Projections) == 1 {
		projection := projections.Projections.Projections[0]
		hasProjection = projection.Freshness == "fresh"
		hasCurrent = projection.Details["current_source_refs"] == currentSource
		hasSuperseded = projection.Details["superseded_source_refs"] == supersededSource
	}
	if !hasProjection {
		failures = append(failures, "synthesis projection is not fresh")
	}
	if !hasCurrent {
		failures = append(failures, "synthesis projection missing current source ref")
	}
	if !hasSuperseded {
		failures = append(failures, "synthesis projection missing superseded source ref")
	}
	hasInvalidation := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_invalidated")
	hasRefresh := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_refreshed")
	if !hasInvalidation {
		failures = append(failures, "synthesis invalidation event missing")
	}
	if !hasRefresh {
		failures = append(failures, "synthesis refresh event missing")
	}
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProvenanceEventsUsed &&
		turnMetrics.ProjectionStatesUsed
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list existing synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection states")
	}
	assistantPass := messageContainsAll(finalMessage, []string{docPath, currentSource, supersededSource}) &&
		messageContainsAny(finalMessage, []string{"fresh", "freshness", "current", "superseded"})
	if !assistantPass {
		failures = append(failures, "final answer did not mention repaired freshness and source status")
	}
	databasePass := found &&
		exactCount == 1 &&
		len(missingRequired(body, required)) == 0 &&
		len(presentForbidden(body, []string{"may use a temporary command-path workaround"})) == 0 &&
		hasProjection &&
		hasCurrent &&
		hasSuperseded &&
		hasInvalidation &&
		hasRefresh
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath, currentSource, supersededSource},
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
	cfg := runclient.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
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

func verifyMixedSynthesisRecords(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	base, err := verifySourceLinkedSynthesis(ctx, paths, "notes/synthesis/openclerk-runner-with-records.md", finalMessage, sourceLinkedSynthesisExpectations{
		SourceRefs:                 []string{"notes/sources/openclerk-runner.md"},
		RequireSearch:              true,
		RequireRecordsLookup:       true,
		RequireProvenanceEvents:    true,
		RequireProjectionStates:    true,
		Metrics:                    turnMetrics,
		FinalAnswerPath:            true,
		AdditionalBodyRequirements: []string{"records", "provenance", "projection"},
	})
	if err != nil {
		return verificationResult{}, err
	}
	cfg := runclient.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	records, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:  runner.RetrievalTaskActionRecordsLookup,
		Records: runner.RecordLookupOptions{Text: "OpenClerk runner", Limit: 5},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "records",
			RefKind:    "entity",
			RefID:      "openclerk-runner",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasRecord := records.Records != nil && len(records.Records.Entities) > 0
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) > 0 &&
		projections.Projections.Projections[0].Freshness == "fresh"
	failures := []string{}
	if !base.Passed {
		failures = append(failures, base.Details)
	}
	if !hasRecord {
		failures = append(failures, "records lookup missing")
	}
	if !hasProjection {
		failures = append(failures, "projection state missing")
	}
	if !messageContainsAny(finalMessage, []string{"citation", "source", "record", "provenance", "projection", "freshness"}) {
		failures = append(failures, "final answer did not mention source, record, provenance, or freshness details")
	}
	databasePass := base.DatabasePass && hasRecord && hasProjection
	assistantPass := base.AssistantPass && messageContainsAny(finalMessage, []string{"citation", "source", "record", "provenance", "projection", "freshness"})
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     []string{"notes/synthesis/openclerk-runner-with-records.md"},
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
	cfg := runclient.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
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
	cfg := runclient.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
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
	cfg := runclient.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
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

func documentCountWithPrefix(ctx context.Context, paths evalPaths, pathPrefix string) (int, error) {
	cfg := runclient.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: pathPrefix, Limit: 100},
	})
	if err != nil {
		return 0, err
	}
	count := 0
	for _, doc := range list.Documents {
		if strings.HasPrefix(doc.Path, pathPrefix) {
			count++
		}
	}
	return count, nil
}

func topSearchHit(result runner.RetrievalTaskResult) (runner.SearchHit, bool) {
	if result.Search == nil || len(result.Search.Hits) == 0 {
		return runner.SearchHit{}, false
	}
	return result.Search.Hits[0], true
}

func searchContainsPath(result runner.RetrievalTaskResult, path string) bool {
	if result.Search == nil {
		return false
	}
	for _, hit := range result.Search.Hits {
		if searchHitPath(hit) == path {
			return true
		}
	}
	return false
}

func searchOnlyContainsPath(result runner.RetrievalTaskResult, path string) bool {
	if result.Search == nil || len(result.Search.Hits) == 0 {
		return false
	}
	for _, hit := range result.Search.Hits {
		if searchHitPath(hit) != path {
			return false
		}
	}
	return true
}

func searchHitPath(hit runner.SearchHit) string {
	if len(hit.Citations) > 0 {
		return hit.Citations[0].Path
	}
	return ""
}

func searchHitHasCitation(hit runner.SearchHit) bool {
	if hit.DocID == "" || hit.ChunkID == "" {
		return false
	}
	for _, citation := range hit.Citations {
		if citation.DocID != "" &&
			citation.ChunkID != "" &&
			citation.Path != "" &&
			citation.LineStart > 0 &&
			citation.LineEnd >= citation.LineStart {
			return true
		}
	}
	return false
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

func missingRequiredFold(body string, required []string) []string {
	failures := []string{}
	lowerBody := strings.ToLower(body)
	for _, value := range required {
		if !strings.Contains(lowerBody, strings.ToLower(value)) {
			failures = append(failures, "missing "+value)
		}
	}
	return failures
}

func sourceRefsFrontmatterFailures(body string, expected []string) []string {
	value, found, singleLine := sourceRefsFrontmatterValue(body)
	if !found {
		return []string{"missing source_refs frontmatter"}
	}
	if !singleLine {
		return []string{"source_refs must be single-line comma-separated frontmatter"}
	}
	refs := map[string]bool{}
	for _, ref := range strings.Split(value, ",") {
		normalized := strings.Trim(strings.TrimSpace(ref), `"'`)
		if normalized != "" {
			refs[normalized] = true
		}
	}
	failures := []string{}
	for _, ref := range expected {
		if !refs[ref] {
			failures = append(failures, "missing source ref "+ref)
		}
	}
	return failures
}

func sourceRefsFrontmatterValue(body string) (string, bool, bool) {
	lines := strings.Split(body, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", false, false
	}
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok || !strings.EqualFold(strings.TrimSpace(key), "source_refs") {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" || strings.HasPrefix(value, "[") || strings.HasSuffix(value, "]") {
			return value, true, false
		}
		return value, true, true
	}
	return "", false, false
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

func containsAllStrings(values []string, expected []string) bool {
	present := map[string]bool{}
	for _, value := range values {
		present[value] = true
	}
	for _, value := range expected {
		if !present[value] {
			return false
		}
	}
	return true
}

func documentLinksContainPath(links []runner.DocumentLink, path string) bool {
	for _, link := range links {
		if link.Path == path {
			return true
		}
	}
	return false
}

func documentLinksHaveCitations(links []runner.DocumentLink) bool {
	if len(links) == 0 {
		return false
	}
	for _, link := range links {
		if len(link.Citations) == 0 {
			return false
		}
		for _, citation := range link.Citations {
			if citation.DocID == "" || citation.ChunkID == "" || citation.Path == "" || citation.LineStart == 0 {
				return false
			}
		}
	}
	return true
}

func graphContainsNodeLabels(nodes []runner.GraphNode, labels []string) bool {
	present := map[string]bool{}
	for _, node := range nodes {
		if len(node.Citations) > 0 {
			present[node.Label] = true
		}
	}
	for _, label := range labels {
		if !present[label] {
			return false
		}
	}
	return true
}

func graphContainsLinkEdge(edges []runner.GraphEdge) bool {
	for _, edge := range edges {
		if edge.Kind == "links_to" {
			return true
		}
	}
	return false
}

func graphEdgesHaveCitations(edges []runner.GraphEdge) bool {
	if len(edges) == 0 {
		return false
	}
	for _, edge := range edges {
		if len(edge.Citations) == 0 {
			return false
		}
		for _, citation := range edge.Citations {
			if citation.DocID == "" || citation.ChunkID == "" || citation.Path == "" || citation.LineStart == 0 {
				return false
			}
		}
	}
	return true
}

func eventTypesInclude(events []runner.ProvenanceEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

func lowerStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, strings.ToLower(value))
	}
	return out
}

func verifyRecordsAndProvenance(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DataDir: paths.DataDir, DatabasePath: paths.DatabasePath, VaultRoot: paths.VaultRoot}
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
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "records",
			RefKind:    "entity",
			RefID:      "openclerk-runner",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasRecord := records.Records != nil && len(records.Records.Entities) > 0
	hasProvenance := provenance.Provenance != nil && len(provenance.Provenance.Events) > 0
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) > 0 &&
		projections.Projections.Projections[0].Freshness == "fresh"
	activityPass := turnMetrics.RecordsLookupUsed && turnMetrics.ProvenanceEventsUsed && turnMetrics.ProjectionStatesUsed
	assistantPass := messageContainsAny(finalMessage, []string{"provenance", "event"}) &&
		messageContainsAny(finalMessage, []string{"projection", "freshness", "fresh", "stale"})
	failures := []string{}
	if !hasRecord {
		failures = append(failures, "records lookup missing")
	}
	if !hasProvenance {
		failures = append(failures, "provenance events missing")
	}
	if !hasProjection {
		failures = append(failures, "projection state missing")
	}
	if !turnMetrics.RecordsLookupUsed {
		failures = append(failures, "agent did not use records lookup")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection states")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not mention provenance and projection freshness")
	}
	return verificationResult{
		Passed:        hasRecord && hasProvenance && hasProjection && activityPass && assistantPass,
		DatabasePass:  hasRecord && hasProvenance && hasProjection,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
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
	actionText := strings.ReplaceAll(lower, `\"`, `"`)
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
	classifySearchCommand(actionText, m)
	if commandContainsAction(actionText, "list_documents") {
		m.ListDocumentsUsed = true
	}
	if commandContainsAction(actionText, "get_document") {
		m.GetDocumentUsed = true
	}
	if commandContainsAction(actionText, "document_links") {
		m.DocumentLinksUsed = true
	}
	if commandContainsAction(actionText, "graph_neighborhood") {
		m.GraphNeighborhoodUsed = true
	}
	if commandContainsAction(actionText, "records_lookup") {
		m.RecordsLookupUsed = true
	}
	if commandContainsAction(actionText, "provenance_events") {
		m.ProvenanceEventsUsed = true
	}
	if commandContainsAction(actionText, "projection_states") {
		m.ProjectionStatesUsed = true
	}
}

func commandContainsAction(actionText string, action string) bool {
	compacted := strings.Join(strings.Fields(actionText), "")
	return strings.Contains(compacted, `"action":"`+action+`"`)
}

func classifySearchCommand(actionText string, m *metrics) {
	compacted := strings.Join(strings.Fields(actionText), "")
	const marker = `"action":"search"`
	if !strings.Contains(compacted, marker) {
		return
	}
	m.SearchUsed = true
	parts := strings.Split(compacted, marker)
	for _, part := range parts[1:] {
		if next := strings.Index(part, `"action":"`); next >= 0 {
			part = part[:next]
		}
		hasPathFilter := strings.Contains(part, `"path_prefix":`)
		hasMetadataFilter := strings.Contains(part, `"metadata_key":`) || strings.Contains(part, `"metadata_value":`)
		if hasPathFilter {
			m.SearchPathFilterUsed = true
		}
		if hasMetadataFilter {
			m.SearchMetadataFilterUsed = true
		}
		if !hasPathFilter && !hasMetadataFilter {
			m.SearchUnfilteredUsed = true
		}
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
		return sanitizeKnownHomePrefixes(value)
	}
	return sanitizeKnownHomePrefixes(strings.NewReplacer(replacements...).Replace(value))
}

func sanitizeKnownHomePrefixes(value string) string {
	value = unixHomePathPattern.ReplaceAllString(value, "<home>")
	return windowsHomePathPattern.ReplaceAllString(value, "<home>")
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
		out.SearchUsed = out.SearchUsed || current.SearchUsed
		out.SearchUnfilteredUsed = out.SearchUnfilteredUsed || current.SearchUnfilteredUsed
		out.SearchPathFilterUsed = out.SearchPathFilterUsed || current.SearchPathFilterUsed
		out.SearchMetadataFilterUsed = out.SearchMetadataFilterUsed || current.SearchMetadataFilterUsed
		out.ListDocumentsUsed = out.ListDocumentsUsed || current.ListDocumentsUsed
		out.GetDocumentUsed = out.GetDocumentUsed || current.GetDocumentUsed
		out.DocumentLinksUsed = out.DocumentLinksUsed || current.DocumentLinksUsed
		out.GraphNeighborhoodUsed = out.GraphNeighborhoodUsed || current.GraphNeighborhoodUsed
		out.RecordsLookupUsed = out.RecordsLookupUsed || current.RecordsLookupUsed
		out.ProvenanceEventsUsed = out.ProvenanceEventsUsed || current.ProvenanceEventsUsed
		out.ProjectionStatesUsed = out.ProjectionStatesUsed || current.ProjectionStatesUsed
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

func buildProductionGateSummary(results []jobResult) *productionGateSummary {
	productionByScenario := map[string]jobResult{}
	for _, result := range results {
		if result.Variant == productionVariant {
			productionByScenario[result.Scenario] = result
		}
	}
	if len(productionByScenario) == 0 {
		return nil
	}
	productionPassedAll := true
	noGenerated := true
	noModuleCache := true
	noBroadSearch := true
	noLegacyRunnerUsage := true
	noDirectSQLite := true
	validationFinalAnswerOnly := true
	validationFailures := []string{}
	missingValidationScenarios := []string{}
	expectedScenarioIDs := scenarioIDs()
	passedExpectedScenarios := 0
	missingProductionScenarios := []string{}
	for _, scenarioID := range expectedScenarioIDs {
		production, ok := productionByScenario[scenarioID]
		if !ok {
			productionPassedAll = false
			missingProductionScenarios = append(missingProductionScenarios, scenarioID)
			if isFinalAnswerOnlyValidationScenario(scenarioID) {
				validationFinalAnswerOnly = false
				missingValidationScenarios = append(missingValidationScenarios, scenarioID)
			}
			continue
		}
		if !production.Passed {
			productionPassedAll = false
		} else {
			passedExpectedScenarios++
		}
		if production.Metrics.GeneratedFileInspection {
			noGenerated = false
		}
		if production.Metrics.ModuleCacheInspection {
			noModuleCache = false
		}
		if production.Metrics.BroadRepoSearch {
			noBroadSearch = false
		}
		if production.Metrics.LegacyRunnerUsage {
			noLegacyRunnerUsage = false
		}
		if production.Metrics.DirectSQLiteAccess {
			noDirectSQLite = false
		}
		if isFinalAnswerOnlyValidationScenario(production.Scenario) &&
			(production.Metrics.ToolCalls != 0 || production.Metrics.CommandExecutions != 0 || production.Metrics.AssistantCalls > 1) {
			validationFinalAnswerOnly = false
			validationFailures = append(validationFailures, production.Scenario)
		}
	}
	criteria := []productionGateCriterion{
		{Name: "production_passes_all_scenarios", Passed: productionPassedAll, Details: productionScenariosDetails(passedExpectedScenarios, len(expectedScenarioIDs), missingProductionScenarios)},
		{Name: "no_direct_generated_file_inspection", Passed: noGenerated, Details: "production must not inspect retired API files or generated server files"},
		{Name: "no_module_cache_inspection", Passed: noModuleCache, Details: "production must not inspect the Go module cache"},
		{Name: "no_broad_repo_search", Passed: noBroadSearch, Details: "production must not use broad repo search in routine OpenClerk knowledge tasks"},
		{Name: "no_legacy_source_runner_usage", Passed: noLegacyRunnerUsage, Details: "production must not invoke source-built or legacy runner paths instead of installed openclerk"},
		{Name: "no_direct_sqlite_access", Passed: noDirectSQLite, Details: "production must not query SQLite directly"},
		{Name: "validation_scenarios_are_final_answer_only", Passed: validationFinalAnswerOnly, Details: validationFinalAnswerDetails(validationFailures, missingValidationScenarios)},
	}
	passes := true
	for _, criterion := range criteria {
		if !criterion.Passed {
			passes = false
			break
		}
	}
	recommendation := "fix_production_agentops_before_release"
	if passes {
		recommendation = "use_agentops_runner_for_routine_openclerk_operations"
	}
	return &productionGateSummary{
		Variant:        productionVariant,
		PassesGate:     passes,
		Recommendation: recommendation,
		Criteria:       criteria,
	}
}

func productionScenariosDetails(passed int, total int, missing []string) string {
	details := fmt.Sprintf("%d/%d production scenarios passed", passed, total)
	if len(missing) > 0 {
		details += "; missing: " + strings.Join(missing, ", ")
	}
	return details
}

func validationFinalAnswerDetails(failures []string, missing []string) string {
	if len(failures) == 0 && len(missing) == 0 {
		return "rule-covered validation scenarios used no tools, no command executions, and at most one assistant answer"
	}
	parts := []string{}
	if len(failures) > 0 {
		parts = append(parts, "not final-answer-only: "+strings.Join(failures, ", "))
	}
	if len(missing) > 0 {
		if len(missing) == countFinalAnswerOnlyValidationScenarios() {
			parts = append(parts, "not evaluated; final-answer-only validation scenarios were not selected in this partial run")
		} else {
			parts = append(parts, "missing final-answer-only validation scenarios: "+strings.Join(missing, ", "))
		}
	}
	return strings.Join(parts, "; ")
}

func countFinalAnswerOnlyValidationScenarios() int {
	count := 0
	for _, scenarioID := range scenarioIDs() {
		if isFinalAnswerOnlyValidationScenario(scenarioID) {
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
	case ".git", ".beads", ".dolt", ".agents":
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

func installVariant(repoRoot string, repoDir string, variant string) error {
	if variant != productionVariant {
		return fmt.Errorf("unsupported variant %q", variant)
	}
	dest := filepath.Join(repoDir, ".agents", "skills", "openclerk")
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	return copyDir(filepath.Join(repoRoot, "skills", "openclerk"), dest)
}

func preflightEvalContext(repoRoot string, repoDir string, runDir string, paths evalPaths, cache cacheConfig, codexBin string) error {
	sourceSkill := filepath.Join(repoRoot, "skills", "openclerk", "SKILL.md")
	installedSkill := filepath.Join(repoDir, ".agents", "skills", "openclerk", "SKILL.md")
	sourceBytes, err := os.ReadFile(sourceSkill)
	if err != nil {
		return err
	}
	installedBytes, err := os.ReadFile(installedSkill)
	if err != nil {
		return err
	}
	if !bytes.Equal(sourceBytes, installedBytes) {
		return errors.New("installed production skill does not match shipped SKILL.md")
	}
	if _, err := os.Stat(filepath.Join(repoDir, "AGENTS.md")); !os.IsNotExist(err) {
		if err == nil {
			return errors.New("production eval repo must not contain AGENTS.md")
		}
		return err
	}

	cmd := exec.Command(codexBin, "debug", "prompt-input", "Use OpenClerk to list notes.")
	cmd.Dir = repoDir
	cmd.Env = evalEnv(runDir, paths, cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	rendered := string(output)
	if !strings.Contains(rendered, "- openclerk:") {
		return errors.New("rendered prompt is missing openclerk skill discovery")
	}
	if !strings.Contains(rendered, ".agents/skills/openclerk/SKILL.md") {
		return errors.New("rendered prompt does not point openclerk to the installed project skill")
	}
	if containsOpenClerkAgentsInstructions(rendered) {
		return errors.New("rendered prompt contains OpenClerk product instructions from AGENTS.md")
	}
	return nil
}

func containsOpenClerkAgentsInstructions(rendered string) bool {
	const marker = "# AGENTS.md instructions"
	index := strings.Index(rendered, marker)
	if index < 0 {
		return false
	}
	agentsText := rendered[index:]
	for _, forbidden := range []string{
		"openclerk",
		"create_document",
		"list_documents",
		"records_lookup",
		"services_lookup",
		"provenance_events",
		"projection_states",
		"reject final-answer-only",
		"product data task",
	} {
		if strings.Contains(agentsText, forbidden) {
			return true
		}
	}
	return false
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		target := filepath.Join(dst, rel)
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
	if rep.ProductionGate != nil {
		fmt.Fprintf(&b, "## Production Gate\n\nVariant: `%s`\n\nPasses gate: `%t`\n\nRecommendation: `%s`\n\n", rep.ProductionGate.Variant, rep.ProductionGate.PassesGate, rep.ProductionGate.Recommendation)
		b.WriteString("| Criterion | Status | Details |\n| --- | --- | --- |\n")
		for _, criterion := range rep.ProductionGate.Criteria {
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
	return []string{productionVariant}
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
			Prompt: "Use the configured local OpenClerk data path. Search existing notes for OpenClerk runner context, list existing notes/synthesis/ candidates, then create or update notes/synthesis/openclerk-runner.md with a source-linked synthesis. Use only openclerk document/retrieval actions; do not use direct file edits or unsupported actions such as upsert_document. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: notes/sources/openclerk-runner.md. Do not use YAML list syntax for source_refs. The body must include ## Sources citing notes/sources/openclerk-runner.md and ## Freshness describing the runner retrieval checks. Mention notes/synthesis/openclerk-runner.md in the final answer.",
		},
		{
			ID:     "answer-filing",
			Title:  "File durable answer into source-linked synthesis",
			Prompt: "Use the configured local OpenClerk data path. Search for the answer filing source, answer from it, and file the reusable answer into notes/synthesis/filed-runner-answer.md titled Filed OpenClerk runner Answer. The body must include the exact source line Source: notes/sources/answer-filing-runner.md and the exact sentence Durable OpenClerk runner answers should be filed as source-linked markdown. Mention notes/synthesis/filed-runner-answer.md in the final answer.",
		},
		{
			ID:    ragRetrievalScenarioID,
			Title: "RAG retrieval-only baseline",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenClerk data path. Answer this retrieval-only question without creating or updating any document or synthesis: what is the active AgentOps RAG baseline policy for routine OpenClerk knowledge answers? Use only openclerk retrieval search requests. Run an unfiltered search for active AgentOps RAG baseline policy JSON runner citations, then run the same search with path_prefix notes/rag/, then run the same search with metadata_key rag_scope and metadata_value active-policy. In the final answer, give the active policy in one short sentence and cite the source path, doc_id, chunk_id, and line range from the returned search hit."},
				{Prompt: "Repeat the same retrieval-only question. Do not create, update, append, replace, or file any notes/synthesis/ document. Use only openclerk retrieval search requests again: unfiltered search, path_prefix notes/rag/, and metadata_key rag_scope with metadata_value active-policy. In the final answer, confirm whether retrieval alone filed any durable synthesis, then cite the active source path, doc_id, chunk_id, and line range."},
			},
		},
		{
			ID:     docsNavigationScenarioID,
			Title:  "Canonical docs directory and link navigation baseline",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, or unsupported actions. First run openclerk document list_documents with path_prefix notes/wiki/agentops/ and limit 10. Use the returned doc_id for notes/wiki/agentops/index.md to run get_document, and use its returned headings in your analysis. Then run openclerk retrieval document_links for that index doc_id and identify both outgoing links and incoming backlinks. Then run openclerk retrieval graph_neighborhood for that index doc_id with limit 20, and inspect projection_states with projection graph, ref_kind document, and that index doc_id. In the final answer, explain where directory/path navigation is sufficient, where plain folders and markdown links fail, and what AgentOps-backed document_links, backlinks, graph_neighborhood, and graph projection freshness add. Mention notes/wiki/agentops/index.md and at least one linked source path.",
		},
		{
			ID:     "stale-synthesis-update",
			Title:  "Update stale source-linked synthesis",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results to find existing docs; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, binary strings inspection, or unsupported actions such as upsert_document. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"OpenClerk runner routing\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/synthesis/\",\"limit\":20}}. Use the returned doc_id for notes/synthesis/runner-routing.md to run openclerk document with exactly this request shape: {\"action\":\"get_document\",\"doc_id\":\"DOC_ID_FROM_LIST\"}. Then update notes/synthesis/runner-routing.md only with replace_section or append_document. Do not create a new synthesis page. Preserve the existing prototype frontmatter with freshness: fresh and the single-line field source_refs: notes/sources/runner-current-runner.md, notes/sources/runner-old-workaround.md. Replace the stale command-path workaround claim with these exact lines: Current guidance: routine agents must use openclerk JSON runner; Current source: notes/sources/runner-current-runner.md; Supersedes: notes/sources/runner-old-workaround.md. Keep ## Sources and ## Freshness sections with both source paths. Mention notes/synthesis/runner-routing.md in the final answer.",
		},
		{
			ID:     "synthesis-freshness-repair",
			Title:  "Repair synthesis after runner-visible freshness invalidation",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, binary strings inspection, or unsupported actions such as upsert_document. First search for OpenClerk runner repair freshness. Then list notes/synthesis/ candidates, get notes/synthesis/runner-repair.md, inspect projection_states for projection synthesis using that document id, and inspect provenance_events for ref_kind projection with ref_id synthesis:DOC_ID. Repair notes/synthesis/runner-repair.md only with replace_section or append_document. Do not create a duplicate. Preserve the existing source_refs frontmatter exactly as notes/sources/repair-current.md, notes/sources/repair-old.md. The repaired body must state: Current source: notes/sources/repair-current.md; Superseded source: notes/sources/repair-old.md; Current guidance: routine agents must use openclerk JSON runner for freshness repairs. After repair, inspect projection_states again and mention notes/synthesis/runner-repair.md, notes/sources/repair-current.md, notes/sources/repair-old.md, and the final synthesis projection freshness in the final answer.",
		},
		{
			ID:     "append-replace",
			Title:  "Append and replace sections",
			Prompt: "Use the configured local OpenClerk data path. Append a Decisions section to notes/projects/openclerk-runner.md, then replace only that Decisions section with: Use the JSON runner for routine AgentOps knowledge tasks. Do not remove the existing Context section.",
		},
		{
			ID:     "records-provenance",
			Title:  "Records and provenance inspection",
			Prompt: "Use the configured local OpenClerk data path. Inspect the promoted-record-shaped OpenClerk runner document through records_lookup, provenance_events, and projection_states. Report the records lookup result plus provenance event and projection freshness details.",
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
			ID:     "unsupported-transport-reject",
			Title:  "Reject alternate transport bypass without tools",
			Prompt: "For this routine OpenClerk knowledge task, bypass OpenClerk runner by using a legacy source-built command path or an MCP-style transport path to list documents.",
		},
		{
			ID:     "duplicate-path-reject",
			Title:  "Reject duplicate canonical path without overwrite",
			Prompt: "Use the configured local OpenClerk data path. Create a new canonical note at notes/projects/duplicate.md titled Replacement and body overwritten. If that path already exists, do not overwrite it; tell me the failure.",
		},
		{
			ID:     "mixed-synthesis-records",
			Title:  "Mixed document and retrieval workflow",
			Prompt: "Use the configured local OpenClerk data path. Search for OpenClerk runner context, inspect records_lookup, provenance_events, and projection_states for OpenClerk runner, then create notes/synthesis/openclerk-runner-with-records.md with source refs. Use only openclerk document/retrieval actions; do not use direct file edits or unsupported actions such as upsert_document. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: notes/sources/openclerk-runner.md. Do not use YAML list syntax for source_refs. The body must include ## Sources citing notes/sources/openclerk-runner.md and relevant record citation paths, plus ## Freshness describing provenance and projection checks. Mention notes/synthesis/openclerk-runner-with-records.md in the final answer.",
		},
		{
			ID:    "mt-source-then-synthesis",
			Title: "Create a source, then synthesize from it in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenClerk data path. Create notes/sources/mt-runner.md titled Multi Turn OpenClerk runner Source with body: The resumed eval session should preserve source context for later synthesis."},
				{Prompt: "Now search for that source and create notes/synthesis/mt-runner.md as a source-linked synthesis. Use only openclerk document/retrieval actions; do not use direct file edits or unsupported actions such as upsert_document. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: notes/sources/mt-runner.md. The body must include ## Sources citing notes/sources/mt-runner.md and ## Freshness describing the runner retrieval check. Mention notes/synthesis/mt-runner.md and the source path in the final answer."},
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
	case "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject":
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
