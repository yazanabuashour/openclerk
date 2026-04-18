package main

import (
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
	"strings"
	"sync"
	"time"
)

const defaultParallel = 4

type runConfig struct {
	Parallel  int
	Variant   string
	Scenario  string
	RunRoot   string
	ReportDir string
	CodexBin  string
	RepoRoot  string
}

type evalJob struct {
	Index    int
	Variant  string
	Scenario scenario
}

type scenario struct {
	ID     string
	Title  string
	Prompt string
}

type report struct {
	Metadata reportMetadata `json:"metadata"`
	Results  []jobResult    `json:"results"`
}

type reportMetadata struct {
	GeneratedAt              time.Time `json:"generated_at"`
	ConfiguredParallelism    int       `json:"configured_parallelism"`
	HarnessElapsedSeconds    float64   `json:"harness_elapsed_seconds"`
	RunRootArtifactReference string    `json:"run_root_artifact_reference"`
	RawLogPlaceholder        string    `json:"raw_log_placeholder"`
	Variants                 []string  `json:"variants"`
	Scenarios                []string  `json:"scenarios"`
}

type jobResult struct {
	Variant                 string     `json:"variant"`
	Scenario                string     `json:"scenario"`
	ScenarioTitle           string     `json:"scenario_title"`
	Status                  string     `json:"status"`
	Error                   string     `json:"error,omitempty"`
	ToolCalls               int        `json:"tool_calls"`
	AssistantCalls          int        `json:"assistant_calls"`
	WallSeconds             float64    `json:"wall_seconds"`
	NonCacheInputTokens     int        `json:"non_cache_input_tokens"`
	OutputTokens            int        `json:"output_tokens"`
	GeneratedFileInspection bool       `json:"generated_file_inspection"`
	ModuleCacheInspection   bool       `json:"module_cache_inspection"`
	BroadRepoSearch         bool       `json:"broad_repo_search"`
	DirectSQLiteAccess      bool       `json:"direct_sqlite_access"`
	RawLogArtifactReference string     `json:"raw_log_artifact_reference"`
	PromptSummary           string     `json:"prompt_summary"`
	StartedAt               time.Time  `json:"started_at"`
	CompletedAt             *time.Time `json:"completed_at,omitempty"`
}

type jobRunner func(context.Context, runConfig, evalJob) jobResult

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, codexJobRunner))
}

func run(args []string, stdout io.Writer, stderr io.Writer, runner jobRunner) int {
	if len(args) == 0 || args[0] != "run" {
		_, _ = fmt.Fprintln(stderr, "usage: ockp run [--parallel N] [--variant ids] [--scenario ids] [--run-root path] [--report-dir path] [--codex-bin path]")
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
	config := runConfig{}
	fs.IntVar(&config.Parallel, "parallel", defaultParallel, "number of independent eval jobs to run concurrently")
	fs.StringVar(&config.Variant, "variant", "", "comma-separated variant ids")
	fs.StringVar(&config.Scenario, "scenario", "", "comma-separated scenario ids")
	fs.StringVar(&config.RunRoot, "run-root", "", "directory for isolated run artifacts")
	fs.StringVar(&config.ReportDir, "report-dir", filepath.Join("docs", "evals", "results"), "directory for reduced reports")
	fs.StringVar(&config.CodexBin, "codex-bin", "codex", "codex executable")
	fs.StringVar(&config.RepoRoot, "repo-root", ".", "repository root to copy for each job")
	if err := fs.Parse(args); err != nil {
		return runConfig{}, err
	}
	if fs.NArg() != 0 {
		return runConfig{}, fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if config.Parallel < 1 {
		return runConfig{}, errors.New("--parallel must be at least 1")
	}
	if config.RunRoot == "" {
		config.RunRoot = filepath.Join(os.TempDir(), fmt.Sprintf("openclerk-ockp-%d", time.Now().UnixNano()))
	}
	return config, nil
}

func executeRun(ctx context.Context, config runConfig, stdout io.Writer, runner jobRunner) error {
	start := time.Now()
	jobs, err := buildJobs(config)
	if err != nil {
		return err
	}
	results := runJobs(ctx, config, jobs, runner)
	rep := report{
		Metadata: reportMetadata{
			GeneratedAt:              time.Now().UTC(),
			ConfiguredParallelism:    config.Parallel,
			HarnessElapsedSeconds:    time.Since(start).Seconds(),
			RunRootArtifactReference: "<run-root>",
			RawLogPlaceholder:        "<run-root>/<variant>/<scenario>/events.jsonl",
			Variants:                 selectedVariants(config),
			Scenarios:                selectedScenarioIDs(config),
		},
		Results: results,
	}
	if err := os.MkdirAll(config.ReportDir, 0o755); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}
	jsonPath := filepath.Join(config.ReportDir, "ockp-latest.json")
	markdownPath := filepath.Join(config.ReportDir, "ockp-latest.md")
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

func runJobs(ctx context.Context, config runConfig, jobs []evalJob, runner jobRunner) []jobResult {
	results := make([]jobResult, len(jobs))
	jobCh := make(chan evalJob)
	var wg sync.WaitGroup
	workers := min(config.Parallel, max(1, len(jobs)))
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				results[job.Index] = runner(ctx, config, job)
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

func codexJobRunner(ctx context.Context, config runConfig, job evalJob) jobResult {
	start := time.Now()
	result := jobResult{
		Variant:                 job.Variant,
		Scenario:                job.Scenario.ID,
		ScenarioTitle:           job.Scenario.Title,
		Status:                  "failed",
		StartedAt:               start.UTC(),
		RawLogArtifactReference: fmt.Sprintf("<run-root>/%s/%s/events.jsonl", job.Variant, job.Scenario.ID),
		PromptSummary:           job.Scenario.Prompt,
	}
	jobDir := filepath.Join(config.RunRoot, job.Variant, job.Scenario.ID)
	repoDir := filepath.Join(jobDir, "repo")
	rawLogPath := filepath.Join(jobDir, "events.jsonl")
	if err := os.MkdirAll(jobDir, 0o755); err != nil {
		result.Error = err.Error()
		return result
	}
	if err := copyRepo(config.RepoRoot, repoDir); err != nil {
		result.Error = fmt.Sprintf("copy repo: %v", err)
		return result
	}
	if err := writeVariantInstructions(repoDir, job.Variant); err != nil {
		result.Error = fmt.Sprintf("configure variant: %v", err)
		return result
	}
	cmd := exec.CommandContext(ctx, config.CodexBin,
		"exec",
		"--json",
		"--ephemeral",
		"--full-auto",
		"--skip-git-repo-check",
		"--add-dir", jobDir,
		"-C", repoDir,
		job.Scenario.Prompt,
	)
	cmd.Env = append(os.Environ(),
		"OPENCLERK_DATA_DIR="+filepath.Join(jobDir, "data"),
		"GOMODCACHE="+filepath.Join(jobDir, "gomodcache"),
	)
	output, err := cmd.CombinedOutput()
	completed := time.Now().UTC()
	result.CompletedAt = &completed
	result.WallSeconds = time.Since(start).Seconds()
	if writeErr := os.WriteFile(rawLogPath, output, 0o644); writeErr != nil && err == nil {
		err = writeErr
	}
	result.ToolCalls = countLinesContaining(output, `"type":"tool_call"`)
	result.AssistantCalls = countLinesContaining(output, `"type":"message"`)
	tokens := extractTokenMetrics(output)
	result.NonCacheInputTokens = tokens.NonCacheInputTokens
	result.OutputTokens = tokens.OutputTokens
	result.GeneratedFileInspection = bytesContainAny(output, []string{"client.gen.go", "openapi.gen.go"})
	result.ModuleCacheInspection = bytesContainAny(output, []string{"GOMODCACHE", "/pkg/mod"})
	result.BroadRepoSearch = bytesContainAny(output, []string{"rg --files", "find ."})
	result.DirectSQLiteAccess = bytesContainAny(output, []string{"sqlite3", "SELECT "})
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Status = "completed"
	return result
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
	case "production":
		return `# OpenClerk Agent Eval Variant: production

For direct user requests to create, list, update, search, or inspect local OpenClerk knowledge, use the production AgentOps JSON runner:

` + "```bash" + `
go run ./cmd/openclerk-agentops document
go run ./cmd/openclerk-agentops retrieval
` + "```" + `

Pass one JSON request on stdin and answer only from the JSON result. Do not inspect generated clients, backend-variant packages, the Go module cache, or SQLite directly for routine knowledge tasks.
`, nil
	case "sdk-baseline":
		return `# OpenClerk Agent Eval Variant: sdk-baseline

For direct user requests to create, list, update, search, or inspect local OpenClerk knowledge, use the code-first local SDK at ` + "`client/local`" + ` with ` + "`local.OpenClient(local.Config{})`" + `.

Do not use ` + "`cmd/openclerk-agentops`" + ` for this baseline variant. Use generated OpenAPI clients only if the SDK facade does not cover the requested workflow.
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
	case ".git", ".beads", "docs", "scripts":
		return entry.IsDir()
	case "AGENTS.md":
		return true
	default:
		return false
	}
}

func writeJSONReport(path string, rep report) error {
	content, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON report: %w", err)
	}
	content = append(content, '\n')
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write JSON report: %w", err)
	}
	return nil
}

func writeMarkdownReport(path string, rep report) error {
	var b strings.Builder
	b.WriteString("# OpenClerk Agent Eval\n\n")
	fmt.Fprintf(&b, "- Configured parallelism: `%d`\n", rep.Metadata.ConfiguredParallelism)
	fmt.Fprintf(&b, "- Harness elapsed seconds: `%.3f`\n", rep.Metadata.HarnessElapsedSeconds)
	b.WriteString("- Raw logs: `<run-root>/<variant>/<scenario>/events.jsonl`\n\n")
	b.WriteString("| Variant | Scenario | Status | Tools | Assistant Calls | Wall Seconds | Raw Log |\n")
	b.WriteString("| --- | --- | --- | ---: | ---: | ---: | --- |\n")
	for _, result := range rep.Results {
		fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | %d | %d | %.3f | `%s` |\n",
			result.Variant,
			result.Scenario,
			result.Status,
			result.ToolCalls,
			result.AssistantCalls,
			result.WallSeconds,
			result.RawLogArtifactReference,
		)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write Markdown report: %w", err)
	}
	return nil
}

func selectedVariants(config runConfig) []string {
	if strings.TrimSpace(config.Variant) != "" {
		return splitCSV(config.Variant)
	}
	return []string{"production", "sdk-baseline"}
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
			Prompt: "Create an OpenClerk canonical project note at notes/projects/agentops-runner.md with active frontmatter and verify it exists.",
		},
		{
			ID:     "search-synthesis",
			Title:  "Search before source-linked synthesis",
			Prompt: "Search existing OpenClerk notes for agentops runner context, then create or update a source-linked synthesis page with source refs.",
		},
		{
			ID:     "append-replace",
			Title:  "Append and replace sections",
			Prompt: "Append a Decisions section to an OpenClerk note, then replace only that section without losing unrelated content.",
		},
		{
			ID:     "records-provenance",
			Title:  "Records and provenance inspection",
			Prompt: "Create a promoted-record-shaped OpenClerk document, look it up, and inspect provenance and projection freshness.",
		},
	}
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

func countLinesContaining(content []byte, needle string) int {
	count := 0
	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, needle) {
			count++
		}
	}
	return count
}

type tokenMetrics struct {
	NonCacheInputTokens int
	OutputTokens        int
}

func extractTokenMetrics(content []byte) tokenMetrics {
	var metrics tokenMetrics
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var event any
		decoder := json.NewDecoder(strings.NewReader(line))
		decoder.UseNumber()
		if err := decoder.Decode(&event); err != nil {
			continue
		}
		accumulateTokenMetrics(event, &metrics)
	}
	return metrics
}

func accumulateTokenMetrics(value any, metrics *tokenMetrics) {
	switch typed := value.(type) {
	case map[string]any:
		if applyUsageMetrics(typed, metrics) {
			return
		}
		for _, nested := range typed {
			accumulateTokenMetrics(nested, metrics)
		}
	case []any:
		for _, nested := range typed {
			accumulateTokenMetrics(nested, metrics)
		}
	}
}

func applyUsageMetrics(values map[string]any, metrics *tokenMetrics) bool {
	input := intFromAny(firstPresent(values, "input_tokens", "prompt_tokens"))
	output := intFromAny(firstPresent(values, "output_tokens", "completion_tokens"))
	cached := intFromAny(firstPresent(values, "cached_input_tokens", "cached_tokens"))
	cached += intFromAny(nestedValue(values, "input_tokens_details", "cached_tokens"))
	cached += intFromAny(nestedValue(values, "prompt_tokens_details", "cached_tokens"))
	if input == 0 && output == 0 && cached == 0 {
		return false
	}
	if input > cached {
		metrics.NonCacheInputTokens += input - cached
	}
	metrics.OutputTokens += output
	return true
}

func firstPresent(values map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return value
		}
	}
	return nil
}

func nestedValue(values map[string]any, key string, nestedKey string) any {
	nested, ok := values[key].(map[string]any)
	if !ok {
		return nil
	}
	return nested[nestedKey]
}

func intFromAny(value any) int {
	switch typed := value.(type) {
	case json.Number:
		number, err := typed.Int64()
		if err != nil {
			return 0
		}
		return int(number)
	case float64:
		return int(typed)
	case int:
		return typed
	default:
		return 0
	}
}

func bytesContainAny(content []byte, needles []string) bool {
	value := string(content)
	for _, needle := range needles {
		if strings.Contains(value, needle) {
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
