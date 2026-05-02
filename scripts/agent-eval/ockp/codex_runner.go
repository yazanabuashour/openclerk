package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

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
	fixtures := startSourceURLUpdateFixtures(job.Scenario.ID)
	if fixtures != nil {
		defer fixtures.Close()
	}
	if err := timedPhase(&timings.PrepareRunDir, func() error { return prepareRunDir(jobDir, cache) }); err != nil {
		result.Error = err.Error()
		return result
	}
	if fixtures != nil {
		if err := fixtures.prepareFiles(jobDir); err != nil {
			result.Error = fmt.Sprintf("prepare fixture files: %v", err)
			return result
		}
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
	if fixtures != nil && isArtifactPDFScenario(job.Scenario.ID) {
		preflight := runArtifactPDFFixturePreflight(ctx, jobDir, paths, cache, fixtures)
		result.FixturePreflight = &preflight
	}
	if cache.Mode == cacheModeIsolated {
		if err := timedPhase(&timings.WarmCache, func() error { return warmGoModules(repoDir, jobDir, paths, cache) }); err != nil {
			result.Error = fmt.Sprintf("warm go modules: %v", err)
			return result
		}
	}
	if err := timedPhase(&timings.SeedData, func() error { return seedScenarioWithFixtures(ctx, paths, job.Scenario, fixtures) }); err != nil {
		result.Error = fmt.Sprintf("seed scenario: %v", err)
		return result
	}
	if fixtures != nil {
		if err := fixtures.prepareForAgent(jobDir, job.Scenario.ID); err != nil {
			result.Error = fmt.Sprintf("prepare source URL fixture state: %v", err)
			return result
		}
		if err := prepareSourceURLUpdateAgentState(ctx, paths, job.Scenario, fixtures); err != nil {
			result.Error = fmt.Sprintf("prepare source URL update state: %v", err)
			return result
		}
	}

	turns := scenarioTurns(job.Scenario)
	turnResults := make([]turnResult, 0, len(turns))
	sessionID := ""
	var runErr error
	for i, turn := range turns {
		turnIndex := i + 1
		turnResult, parsed, err := runScenarioTurn(ctx, config, repoDir, jobDir, paths, job, turn, turnIndex, sessionID, cache, fixtures)
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
	DatabasePath string
	GoCache      string
	GoModCache   string
	CodexHome    string
	ZDotDir      string
	Temp         string
}

func scenarioPaths(repoDir string) evalPaths {
	return evalPaths{
		DatabasePath: filepath.Join(repoDir, ".openclerk-eval", "openclerk.db"),
	}
}
func evalPathsFor(runDir string, paths evalPaths, cache cacheConfig) evalPaths {
	out := paths
	out.CodexHome = filepath.Join(runDir, "codex-home")
	out.ZDotDir = filepath.Join(runDir, "zdotdir")
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
func runScenarioTurn(ctx context.Context, config runConfig, repoDir string, runDir string, paths evalPaths, job evalJob, turn scenarioTurn, turnIndex int, sessionID string, cache cacheConfig, fixtures *sourceURLUpdateFixtures) (turnResult, parsedTurn, error) {
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

	if fixtures != nil {
		turn.Prompt = fixtures.renderPrompt(turn.Prompt)
	}
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
		args := []string{codexBin, "exec", "--json", "--ephemeral", "--full-auto", "--skip-git-repo-check", "--ignore-user-config", "-C", repoDir}
		args = appendAddDirs(args, writableRoots)
		args = append(args, baseConfig...)
		return append(args, turn.Prompt)
	}
	if turnIndex == 1 {
		args := []string{codexBin, "exec", "--json", "--full-auto", "--skip-git-repo-check", "--ignore-user-config", "-C", repoDir}
		args = appendAddDirs(args, writableRoots)
		args = append(args, baseConfig...)
		return append(args, turn.Prompt)
	}
	args := []string{codexBin, "exec", "-C", repoDir}
	args = appendAddDirs(args, writableRoots)
	args = append(args, "resume", "--json", "--full-auto", "--skip-git-repo-check", "--ignore-user-config")
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
	env := filteredEnv(os.Environ(),
		"CODEX_HOME",
		"OPENCLERK_DATA_DIR",
		"OPENCLERK_DATABASE_PATH",
		evalSourceFixtureRootEnv,
		"OPENCLERK_VAULT_ROOT",
		"GOCACHE",
		"GOMODCACHE",
		"TMPDIR",
		"PATH",
		"ZDOTDIR",
	)
	pathValue := filepath.Join(runDir, "bin")
	if existing := os.Getenv("PATH"); existing != "" {
		pathValue += string(os.PathListSeparator) + existing
	}
	env = append(env,
		"CODEX_HOME="+effective.CodexHome,
		"ZDOTDIR="+effective.ZDotDir,
		"OPENCLERK_DATABASE_PATH="+effective.DatabasePath,
		evalSourceFixtureRootEnv+"="+evalSourceFixtureRoot(runDir),
		"GOCACHE="+effective.GoCache,
		"GOMODCACHE="+effective.GoModCache,
		"TMPDIR="+effective.Temp,
		"PATH="+pathValue,
	)
	return env
}
func filteredEnv(env []string, keys ...string) []string {
	if len(keys) == 0 {
		return append([]string{}, env...)
	}
	blocked := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		blocked[key] = struct{}{}
	}
	filtered := make([]string, 0, len(env))
	for _, entry := range env {
		key, _, found := strings.Cut(entry, "=")
		if found {
			if _, blockedKey := blocked[key]; blockedKey {
				continue
			}
		}
		filtered = append(filtered, entry)
	}
	return filtered
}
func prepareRunDir(runDir string, cache cacheConfig) error {
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}
	paths := evalPathsFor(runDir, evalPaths{}, cache)
	for _, dir := range []string{paths.ZDotDir, paths.Temp} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if err := setupEvalCodexHome(paths.CodexHome); err != nil {
		return err
	}
	return nil
}
func setupEvalCodexHome(dst string) error {
	srcRoot, err := sourceCodexHome()
	if err != nil {
		return err
	}
	return setupEvalCodexHomeFromSource(dst, srcRoot)
}
func sourceCodexHome() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("CODEX_HOME")); configured != "" {
		return configured, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex"), nil
}
func setupEvalCodexHomeFromSource(dst string, sourceHome string) error {
	authBytes, err := os.ReadFile(filepath.Join(sourceHome, "auth.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("missing Codex auth at %s; run codex login before running evals", filepath.Join(sourceHome, "auth.json"))
		}
		return err
	}
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, "auth.json"), authBytes, 0o600)
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
		DatabasePath: filepath.Join(cache.RunRoot, "shared-cache", "prewarm.db"),
	}, cache)
	for _, dir := range []string{paths.GoCache, paths.GoModCache, paths.Temp} {
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
	projectSkill := filepath.Join(repoDir, ".agents", "skills", "openclerk", "SKILL.md")
	effectivePaths := evalPathsFor(runDir, paths, cache)
	codexHomeSkill := filepath.Join(effectivePaths.CodexHome, "skills", "openclerk", "SKILL.md")
	sourceBytes, err := os.ReadFile(sourceSkill)
	if err != nil {
		return err
	}
	projectBytes, err := os.ReadFile(projectSkill)
	if err != nil {
		return err
	}
	if !bytes.Equal(sourceBytes, projectBytes) {
		return errors.New("installed project production skill does not match shipped SKILL.md")
	}
	if err := copyDir(filepath.Dir(sourceSkill), filepath.Dir(codexHomeSkill)); err != nil {
		return fmt.Errorf("install eval CODEX_HOME openclerk skill: %w", err)
	}
	codexHomeBytes, err := os.ReadFile(codexHomeSkill)
	if err != nil {
		return err
	}
	if !bytes.Equal(sourceBytes, codexHomeBytes) {
		return errors.New("installed CODEX_HOME production skill does not match shipped SKILL.md")
	}
	if _, err := os.Stat(filepath.Join(repoDir, "AGENTS.md")); !os.IsNotExist(err) {
		if err == nil {
			return errors.New("production eval repo must not contain AGENTS.md")
		}
		return err
	}

	cmd := exec.Command(codexBin, "debug", "prompt-input", "Use OpenClerk for a routine local-first knowledge-plane task. Search my local OpenClerk knowledge for runner with limit -3.")
	cmd.Dir = repoDir
	cmd.Env = evalEnv(runDir, paths, cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	rendered := string(output)
	if !containsOpenClerkSkillDiscovery(rendered) {
		return errors.New("rendered prompt is missing openclerk skill discovery")
	}
	if !strings.Contains(rendered, "codex-home/skills/openclerk/SKILL.md") &&
		!strings.Contains(rendered, ".agents/skills/openclerk/SKILL.md") {
		return errors.New("rendered prompt does not point openclerk to an installed eval skill")
	}
	if missing := missingOpenClerkBootstrapRejectionGuidance(rendered); len(missing) != 0 {
		return fmt.Errorf("rendered prompt is missing openclerk bootstrap rejection guidance: %s", strings.Join(missing, ", "))
	}
	if containsOpenClerkAgentsInstructions(rendered) {
		return errors.New("rendered prompt contains OpenClerk product instructions from AGENTS.md")
	}
	return nil
}
func containsOpenClerkSkillDiscovery(rendered string) bool {
	return strings.Contains(rendered, "- OpenClerk:") || strings.Contains(rendered, "- openclerk:")
}
func containsOpenClerkBootstrapRejectionGuidance(rendered string) bool {
	return len(missingOpenClerkBootstrapRejectionGuidance(rendered)) == 0
}
func missingOpenClerkBootstrapRejectionGuidance(rendered string) []string {
	lower := strings.ToLower(rendered)
	checks := map[string]bool{
		openClerkBootstrapRejectionText: strings.Contains(lower, openClerkBootstrapRejectionText),
		"required":                      strings.Contains(lower, "required"),
		"missing":                       strings.Contains(lower, "missing"),
		"document":                      strings.Contains(lower, "document"),
		"faithful":                      strings.Contains(lower, "faithful"),
		"explicit user content":         strings.Contains(lower, "explicit user content"),
		"source":                        strings.Contains(lower, "source"),
		"negative or limit -3":          strings.Contains(lower, "negative") || strings.Contains(lower, "limit -3"),
		"limit":                         strings.Contains(lower, "limit"),
		"sqlite":                        strings.Contains(lower, "sqlite"),
		"http":                          strings.Contains(lower, "http"),
		"mcp":                           strings.Contains(lower, "mcp"),
		"source-built":                  strings.Contains(lower, "source-built"),
		"unsupported transport":         strings.Contains(lower, "unsupported transport"),
		"bypass":                        strings.Contains(lower, "bypass"),
		"without tools or no-tools":     strings.Contains(lower, "without tools") || strings.Contains(lower, "no-tools"),
		"openclerk document":            strings.Contains(lower, "openclerk document"),
		"openclerk retrieval":           strings.Contains(lower, "openclerk retrieval"),
	}
	missing := []string{}
	for name, ok := range checks {
		if !ok {
			missing = append(missing, name)
		}
	}
	return missing
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
		"decisions_lookup",
		"decision_record",
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
